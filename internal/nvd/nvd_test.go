package nvd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/vuln"
)

func TestClient_Query_VersionFiltering(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/nvd_responses/git.json")
	if err != nil {
		t.Fatalf("fixture okunamadı: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := NewClient().WithEndpoint(srv.URL)
	id := matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:2.3:a:git-scm:git"}

	// Git 2.50.0: CVE-2025-AFFECTED (< 2.51.0) etkiler, CVE-2020-OLD (< 2.40.0) etkilemez.
	vulns, err := c.Query(context.Background(), id, "2.50.0")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(vulns) != 1 {
		t.Fatalf("1 vuln bekleniyor (version filtresi), alındı: %d", len(vulns))
	}
	if vulns[0].ID != "CVE-2025-AFFECTED" {
		t.Errorf("ID = %s, want CVE-2025-AFFECTED", vulns[0].ID)
	}
	if vulns[0].Severity != vuln.SeverityHigh {
		t.Errorf("severity = %s, want high", vulns[0].Severity)
	}
	if len(vulns[0].FixedVersions) != 1 || vulns[0].FixedVersions[0] != "2.51.0" {
		t.Errorf("FixedVersions = %v, want [2.51.0]", vulns[0].FixedVersions)
	}
}

func TestClient_Query_OldVersionGetsBoth(t *testing.T) {
	fixture, _ := os.ReadFile("../../testdata/nvd_responses/git.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := NewClient().WithEndpoint(srv.URL)
	id := matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:2.3:a:git-scm:git"}

	// Git 2.30.0: her iki CVE de etkiler (< 2.40.0 ve < 2.51.0).
	vulns, err := c.Query(context.Background(), id, "2.30.0")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(vulns) != 2 {
		t.Errorf("2 vuln bekleniyor, alındı: %d", len(vulns))
	}
}

func TestClient_Query_PurlIgnored(t *testing.T) {
	c := NewClient()
	id := matcher.Identifier{Type: matcher.IdentifierTypePURL, Value: "pkg:npm/foo"}
	vulns, err := c.Query(context.Background(), id, "1.0")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if vulns != nil {
		t.Errorf("purl için nil bekleniyor, alındı: %v", vulns)
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"2.50.0", "2.51.0", -1},
		{"2.51.0", "2.50.0", 1},
		{"2.50.0", "2.50.0", 0},
		{"2.50", "2.50.0", 0},
		{"120.0.6099", "120.0.6100", -1},
	}
	for _, tc := range cases {
		if got := compareVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("compareVersions(%q,%q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
