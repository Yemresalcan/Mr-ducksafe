package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yemresalcan/mr-ducksafe/internal/matcher"
	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/scanner"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
	"github.com/yemresalcan/mr-ducksafe/internal/vuln"
)

// Sahte bileşenler — gerçek Registry/NVD'ye gitmeden test için.
type stubScanner struct{}

func (stubScanner) Scan(_ context.Context) ([]scanner.InstalledApp, error) { return nil, nil }

type stubMatcher struct{}

func (stubMatcher) Match(_ context.Context, _ scanner.InstalledApp) (matcher.Identifier, matcher.MatchConfidence, error) {
	return matcher.Identifier{}, matcher.ConfidenceNone, nil
}

type stubVuln struct{}

func (stubVuln) Query(_ context.Context, _ matcher.Identifier, _ string) ([]vuln.Vulnerability, error) {
	return nil, nil
}

func newTestServer(t *testing.T) (*Server, store.Store) {
	t.Helper()
	st, err := store.OpenSQLite("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	eng := scanengine.New(stubScanner{}, stubMatcher{}, stubVuln{}, st)
	srv, err := New(eng, st)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return srv, st
}

func TestDashboard_ReturnsHTML(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Content-Type = %s, want text/html", rec.Header().Get("Content-Type"))
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), "mr-ducksafe") {
		t.Error("body 'mr-ducksafe' içermeli")
	}
}

func TestScansList_ReturnsHTML(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/scans", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestAPIScan_RejectsGET(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/scan", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want 405", rec.Code)
	}
}

func TestAPIScan_SecondCallReturns409(t *testing.T) {
	srv, _ := newTestServer(t)
	if !srv.tryAcquireScan() {
		t.Fatal("first acquire failed")
	}
	defer srv.releaseScan()

	req := httptest.NewRequest(http.MethodPost, "/api/scan", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", rec.Code)
	}
}

func TestStaticServed(t *testing.T) {
	srv, _ := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/static/style.css", nil)
	rec := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}
