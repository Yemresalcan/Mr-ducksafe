// Package scanengine scanner, matcher, osv ve store'u orkestre eder.
package scanengine

import (
	"time"

	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

// ScanResult tek bir taramanın özet sonucu.
type ScanResult struct {
	ScanID    int64
	StartedAt time.Time
	Finished  time.Time
	Apps      []scanner.InstalledApp
	Findings  []store.Finding
	Unmatched []scanner.InstalledApp
	Warnings  []ScanError
}

// Stage scan içindeki bir aşamayı belirtir.
type Stage string

const (
	StageScan    Stage = "scan"
	StageMatch   Stage = "match"
	StageQuery   Stage = "query"
	StagePersist Stage = "persist"
)

// ScanError scan akışında ortaya çıkan bir hatayı taşır.
type ScanError struct {
	Stage   Stage
	AppName string // boş olabilir
	Cause   error
	Fatal   bool
}

func (e ScanError) Error() string {
	if e.AppName == "" {
		return string(e.Stage) + ": " + e.Cause.Error()
	}
	return string(e.Stage) + " (" + e.AppName + "): " + e.Cause.Error()
}

// CountBySeverity result.Findings'i severity'ye göre sayar.
func (r *ScanResult) CountBySeverity() map[string]int {
	out := map[string]int{}
	for _, f := range r.Findings {
		out[f.Severity]++
	}
	return out
}
