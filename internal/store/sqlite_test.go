package store

import (
	"context"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *SQLiteStore {
	t.Helper()
	s, err := OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestSQLiteStore_SaveAndGetScan(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.SaveScan(ctx, Scan{
		StartedAt: time.Now(),
		Host:      "test-host",
		OS:        "windows",
		Status:    StatusRunning,
	})
	if err != nil {
		t.Fatalf("SaveScan: %v", err)
	}
	if id == 0 {
		t.Fatal("ID 0 dönmemeli")
	}

	findings := []Finding{
		{
			AppName:      "Firefox",
			AppVersion:   "100.0",
			CVEID:        "CVE-2022-26485",
			Severity:     "high",
			FixedVersion: "97.0.2",
			Summary:      "Use-after-free",
			References:   []string{"https://example.com/advisory"},
		},
	}
	if err := s.AppendFindings(ctx, id, findings); err != nil {
		t.Fatalf("AppendFindings: %v", err)
	}

	if err := s.UpdateScanStatus(ctx, id, StatusSuccess, time.Now()); err != nil {
		t.Fatalf("UpdateScanStatus: %v", err)
	}

	sc, got, err := s.GetScan(ctx, id)
	if err != nil {
		t.Fatalf("GetScan: %v", err)
	}
	if sc.Status != StatusSuccess {
		t.Errorf("status = %s, want success", sc.Status)
	}
	if len(got) != 1 {
		t.Fatalf("1 finding bekleniyor, alındı: %d", len(got))
	}
	if got[0].CVEID != "CVE-2022-26485" {
		t.Errorf("CVE = %s, want CVE-2022-26485", got[0].CVEID)
	}
	if len(got[0].References) != 1 {
		t.Errorf("references = %v, want 1 url", got[0].References)
	}
}

func TestSQLiteStore_CleanupStaleRunning(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.SaveScan(ctx, Scan{
		StartedAt: time.Now().Add(-1 * time.Hour),
		Status:    StatusRunning,
	})
	if err != nil {
		t.Fatal(err)
	}

	n, err := s.CleanupStaleRunning(ctx, 10*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("cleanup count = %d, want 1", n)
	}

	sc, _, err := s.GetScan(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Status != StatusInterrupted {
		t.Errorf("status = %s, want interrupted", sc.Status)
	}
}

func TestSQLiteStore_ListScans(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		if _, err := s.SaveScan(ctx, Scan{StartedAt: time.Now(), Status: StatusSuccess}); err != nil {
			t.Fatal(err)
		}
	}
	scans, err := s.ListScans(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(scans) != 3 {
		t.Errorf("3 scan bekleniyor, alındı: %d", len(scans))
	}
}
