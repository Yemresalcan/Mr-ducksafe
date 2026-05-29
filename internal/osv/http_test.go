package osv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/vuln"
)

func TestHTTPClient_Query_Success(t *testing.T) {
	fixture, err := os.ReadFile("../../testdata/osv_responses/firefox_100.json")
	if err != nil {
		t.Fatalf("fixture okunamadı: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	defer srv.Close()

	c := NewHTTPClient().WithEndpoint(srv.URL)
	id := matcher.Identifier{
		Type:      matcher.IdentifierTypeCPE,
		Value:     "cpe:2.3:a:mozilla:firefox",
		Ecosystem: "Mozilla",
	}
	vulns, err := c.Query(context.Background(), id, "100.0")
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(vulns) != 1 {
		t.Fatalf("1 vuln bekleniyor, alındı: %d", len(vulns))
	}
	if vulns[0].ID != "CVE-2022-26485" {
		t.Errorf("ID = %s, want CVE-2022-26485", vulns[0].ID)
	}
	if vulns[0].Severity != vuln.SeverityHigh {
		t.Errorf("severity = %s, want high (CVSS vector default)", vulns[0].Severity)
	}
	if len(vulns[0].FixedVersions) != 1 || vulns[0].FixedVersions[0] != "97.0.2" {
		t.Errorf("FixedVersions = %v, want [97.0.2]", vulns[0].FixedVersions)
	}
}

func TestHTTPClient_Query_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewHTTPClient().WithEndpoint(srv.URL)
	c.maxRetry = 2 // hızlı fail
	id := matcher.Identifier{Type: matcher.IdentifierTypeCPE, Value: "cpe:2.3:a:test:test"}

	_, err := c.Query(context.Background(), id, "1.0")
	if err == nil {
		t.Fatal("retry sonrası error bekleniyor")
	}
}
