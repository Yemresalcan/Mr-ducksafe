package server

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	scans, err := s.Store.ListScans(r.Context(), 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := map[string]any{"Title": "Dashboard"}
	if len(scans) > 0 {
		data["LastScan"] = scans[0]
		_, findings, _ := s.Store.GetScan(r.Context(), scans[0].ID)
		data["FindingsCount"] = len(findings)
	}
	s.render(w, "dashboard_content", data)
}

func (s *Server) handleScansList(w http.ResponseWriter, r *http.Request) {
	scans, err := s.Store.ListScans(r.Context(), 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.render(w, "scans_content", map[string]any{
		"Title": "Geçmiş",
		"Scans": scans,
	})
}

func (s *Server) handleScanDetail(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/scans/")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	scan, findings, err := s.Store.GetScan(r.Context(), id)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	s.render(w, "scan_detail_content", map[string]any{
		"Title":    "Scan #" + idStr,
		"Scan":     scan,
		"Findings": findings,
	})
}

func (s *Server) handleAPIScan(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !s.tryAcquireScan() {
		http.Error(w, "another scan is already running", http.StatusConflict)
		return
	}
	go s.runScanAsync(context.Background())
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
