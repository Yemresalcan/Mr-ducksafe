package scanengine

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
	"github.com/yemresalcan/mr-ducksafe/internal/vuln"
)

const defaultConcurrency = 5

// Engine scan akışını orkestre eder.
type Engine struct {
	Scanner     scanner.Scanner
	Matcher     matcher.Matcher
	OSV         vuln.Client
	Store       store.Store
	Concurrency int
}

// New verilen bağımlılıklarla yeni Engine oluşturur.
func New(s scanner.Scanner, m matcher.Matcher, o vuln.Client, st store.Store) *Engine {
	return &Engine{
		Scanner:     s,
		Matcher:     m,
		OSV:         o,
		Store:       st,
		Concurrency: defaultConcurrency,
	}
}

type matchedApp struct {
	app scanner.InstalledApp
	id  matcher.Identifier
}

// Run tek bir scan döngüsünü yürütür ve sonucu döndürür.
// progress nil değilse her aşamada event yayınlar; receiver yoksa engellemez.
func (e *Engine) Run(ctx context.Context, progress chan<- Progress) (*ScanResult, error) {
	started := time.Now()
	host, _ := os.Hostname()

	scanID, err := e.Store.SaveScan(ctx, store.Scan{
		StartedAt: started,
		Host:      host,
		OS:        runtime.GOOS,
		Status:    store.StatusRunning,
	})
	if err != nil {
		return nil, err
	}

	result := &ScanResult{ScanID: scanID, StartedAt: started}

	apps, err := e.Scanner.Scan(ctx)
	if err != nil {
		result.Warnings = append(result.Warnings, ScanError{Stage: StageScan, Cause: err, Fatal: true})
		_ = e.Store.UpdateScanStatus(ctx, scanID, store.StatusFailed, time.Now())
		result.Finished = time.Now()
		return result, err
	}
	result.Apps = apps
	emit(progress, Progress{Stage: StageScan, Done: len(apps), Total: len(apps), Message: "scan complete"})

	var toQuery []matchedApp
	for _, app := range apps {
		id, conf, mErr := e.Matcher.Match(ctx, app)
		if mErr != nil {
			result.Warnings = append(result.Warnings, ScanError{Stage: StageMatch, AppName: app.Name, Cause: mErr})
			continue
		}
		if conf == matcher.ConfidenceNone {
			result.Unmatched = append(result.Unmatched, app)
			continue
		}
		toQuery = append(toQuery, matchedApp{app: app, id: id})
	}
	emit(progress, Progress{Stage: StageMatch, Done: len(toQuery), Total: len(apps), Message: "matching complete"})

	findings, queryWarnings := e.queryAll(ctx, scanID, toQuery, progress)
	result.Findings = findings
	result.Warnings = append(result.Warnings, queryWarnings...)

	if err := e.Store.AppendFindings(ctx, scanID, findings); err != nil {
		result.Warnings = append(result.Warnings, ScanError{Stage: StagePersist, Cause: err, Fatal: true})
		_ = e.Store.UpdateScanStatus(ctx, scanID, store.StatusFailed, time.Now())
		result.Finished = time.Now()
		return result, err
	}

	finished := time.Now()
	if err := e.Store.UpdateScanStatus(ctx, scanID, store.StatusSuccess, finished); err != nil {
		result.Warnings = append(result.Warnings, ScanError{Stage: StagePersist, Cause: err})
	}
	result.Finished = finished
	emit(progress, Progress{Stage: StagePersist, Done: len(findings), Total: len(findings), Message: "done"})

	return result, nil
}

func (e *Engine) queryAll(ctx context.Context, scanID int64, items []matchedApp, progress chan<- Progress) ([]store.Finding, []ScanError) {
	var (
		findings []store.Finding
		warnings []ScanError
		mu       sync.Mutex
	)

	conc := e.Concurrency
	if conc <= 0 {
		conc = defaultConcurrency
	}
	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup

	for i, item := range items {
		select {
		case <-ctx.Done():
			break
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func(idx int, m matchedApp) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					warnings = append(warnings, ScanError{Stage: StageQuery, AppName: m.app.Name})
					mu.Unlock()
				}
				<-sem
			}()

			vulns, err := e.OSV.Query(ctx, m.id, m.app.Version)
			mu.Lock()
			if err != nil {
				warnings = append(warnings, ScanError{Stage: StageQuery, AppName: m.app.Name, Cause: err})
			} else {
				for _, v := range vulns {
					findings = append(findings, store.Finding{
						ScanID:       scanID,
						AppName:      m.app.Name,
						AppVersion:   m.app.Version,
						CVEID:        v.ID,
						Severity:     string(v.Severity),
						FixedVersion: firstOr(v.FixedVersions, ""),
						Summary:      v.Summary,
						References:   v.References,
					})
				}
			}
			mu.Unlock()
			emit(progress, Progress{Stage: StageQuery, Done: idx + 1, Total: len(items)})
		}(i, item)
	}
	wg.Wait()
	return findings, warnings
}

func firstOr(s []string, def string) string {
	if len(s) > 0 {
		return s[0]
	}
	return def
}

func emit(ch chan<- Progress, p Progress) {
	if ch == nil {
		return
	}
	select {
	case ch <- p:
	default: // receiver hazır değilse drop et
	}
}
