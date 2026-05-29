package scanengine

import (
	"context"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
	"github.com/yemresalcan/mr-ducksafe/internal/vuln"
)

type fakeScanner struct{ apps []scanner.InstalledApp }

func (f *fakeScanner) Scan(_ context.Context) ([]scanner.InstalledApp, error) { return f.apps, nil }

type fakeMatcher struct{}

func (fakeMatcher) Match(_ context.Context, app scanner.InstalledApp) (matcher.Identifier, matcher.MatchConfidence, error) {
	if app.Name == "Known" {
		return matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:test:known"}, matcher.ConfidenceHigh, nil
	}
	return matcher.Identifier{}, matcher.ConfidenceNone, nil
}

type fakeOSV struct{ vulns []vuln.Vulnerability }

func (f *fakeOSV) Query(_ context.Context, _ matcher.Identifier, _ string) ([]vuln.Vulnerability, error) {
	return f.vulns, nil
}

func newMemoryStore(t *testing.T) store.Store {
	t.Helper()
	s, err := store.OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("store open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestEngine_Run_MatchesAndPersists(t *testing.T) {
	st := newMemoryStore(t)
	eng := New(
		&fakeScanner{apps: []scanner.InstalledApp{
			{Name: "Known", Version: "1.0"},
			{Name: "Unknown", Version: "2.0"},
		}},
		fakeMatcher{},
		&fakeOSV{vulns: []vuln.Vulnerability{
			{ID: "CVE-1", Severity: vuln.SeverityHigh, FixedVersions: []string{"1.1"}},
		}},
		st,
	)

	res, err := eng.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Apps) != 2 {
		t.Errorf("apps = %d, want 2", len(res.Apps))
	}
	if len(res.Unmatched) != 1 {
		t.Errorf("unmatched = %d, want 1", len(res.Unmatched))
	}
	if len(res.Findings) != 1 {
		t.Errorf("findings = %d, want 1", len(res.Findings))
	}
	if res.Findings[0].CVEID != "CVE-1" {
		t.Errorf("CVE = %s, want CVE-1", res.Findings[0].CVEID)
	}
}
