package scanengine

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/nvd"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

type fixedScanner struct{ apps []scanner.InstalledApp }

func (f *fixedScanner) Scan(_ context.Context) ([]scanner.InstalledApp, error) { return f.apps, nil }

// TestIntegration_FirefoxKnownCVECaught — pipeline'ın uçtan uca doğru
// çalıştığını kanıtlar: scanner → matcher → nvd (fake) → store.
// Firefox 100.0 sahte CVE'sinin gerçekten yakalandığını doğrular.
func TestIntegration_FirefoxKnownCVECaught(t *testing.T) {
	firefoxFixture, err := os.ReadFile("../../testdata/nvd_responses/firefox.json")
	if err != nil {
		t.Fatalf("fixture okuma: %v", err)
	}
	emptyResp := []byte(`{"vulnerabilities":[]}`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Sorguda firefox geçiyorsa firefox fixture'ı, aksi halde boş.
		if strings.Contains(r.URL.RawQuery, "firefox") {
			_, _ = w.Write(firefoxFixture)
			return
		}
		_, _ = w.Write(emptyResp)
	}))
	defer srv.Close()

	st, err := store.OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	m, err := matcher.NewDefaultMatcher()
	if err != nil {
		t.Fatal(err)
	}
	nvdClient := nvd.NewClient().WithEndpoint(srv.URL)

	eng := New(
		&fixedScanner{apps: []scanner.InstalledApp{
			{Name: "Mozilla Firefox 100.0 (x64 en-US)", Version: "100.0", Publisher: "Mozilla"},
			{Name: "Brand New Tool", Version: "1.0", Publisher: "Acme"},
		}},
		m,
		nvdClient,
		st,
	)

	res, err := eng.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(res.Findings) != 1 {
		t.Fatalf("1 finding bekleniyor, alındı: %d", len(res.Findings))
	}
	if res.Findings[0].CVEID != "CVE-2022-FIREFOX" {
		t.Errorf("CVE = %s, want CVE-2022-FIREFOX", res.Findings[0].CVEID)
	}
	if res.Findings[0].FixedVersion != "100.5" {
		t.Errorf("FixedVersion = %s, want 100.5", res.Findings[0].FixedVersion)
	}
	if len(res.Unmatched) != 1 || res.Unmatched[0].Name != "Brand New Tool" {
		t.Errorf("unmatched beklenmedik: %+v", res.Unmatched)
	}

	// Persist edildi mi?
	_, findings, err := st.GetScan(context.Background(), res.ScanID)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 1 {
		t.Errorf("DB'de 1 finding bekleniyor, alındı: %d", len(findings))
	}
}
