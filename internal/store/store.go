// Package store SQLite üzerinden scan + findings persistence sağlar.
package store

import (
	"context"
	"time"
)

// ScanStatus bir scan'in mevcut durumu.
type ScanStatus string

const (
	StatusRunning     ScanStatus = "running"
	StatusSuccess     ScanStatus = "success"
	StatusFailed      ScanStatus = "failed"
	StatusInterrupted ScanStatus = "interrupted"
)

// Scan bir tarama oturumunun meta verileri.
type Scan struct {
	ID         int64
	StartedAt  time.Time
	FinishedAt *time.Time
	Host       string
	OS         string
	Status     ScanStatus
}

// Finding tek bir CVE bulgusu.
type Finding struct {
	ID           int64
	ScanID       int64
	AppName      string
	AppVersion   string
	CVEID        string
	Severity     string
	FixedVersion string
	Summary      string
	References   []string
}

// Store SQLite tabanlı persistence arayüzü.
type Store interface {
	SaveScan(ctx context.Context, s Scan) (int64, error)
	AppendFindings(ctx context.Context, scanID int64, findings []Finding) error
	UpdateScanStatus(ctx context.Context, scanID int64, status ScanStatus, finishedAt time.Time) error
	ListScans(ctx context.Context, limit int) ([]Scan, error)
	GetScan(ctx context.Context, id int64) (Scan, []Finding, error)
	CountFindings(ctx context.Context, scanID int64) (int, error)
	PruneOlderThan(ctx context.Context, olderThan time.Duration) (int, error)
	Close() error
}
