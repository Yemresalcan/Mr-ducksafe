// Package server mr-ducksafe lokal HTTP server'ını ve gömülü UI'yı sunar.
package server

import (
	"context"
	"html/template"
	"net/http"
	"sync"

	"github.com/yemresalcan/mr-ducksafe/internal/scanengine"
	"github.com/yemresalcan/mr-ducksafe/internal/store"
)

// Server HTTP handler'ları + bağımlılıkları taşır.
type Server struct {
	Engine    *scanengine.Engine
	Store     store.Store
	templates *template.Template

	mu         sync.Mutex
	activeScan int64 // tek aktif scan kuralı; -1 = rezerve, 0 = boş, >0 = scanID
}

// New verilen engine ve store ile Server oluşturur.
func New(engine *scanengine.Engine, st store.Store) (*Server, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}
	return &Server{Engine: engine, Store: st, templates: tmpl}, nil
}

// Handler net/http için mux döner.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/scans", s.handleScansList)
	mux.HandleFunc("/scans/", s.handleScanDetail)
	mux.HandleFunc("/api/scan", s.handleAPIScan)
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
	return mux
}

// render layout + bir content template'i çağırır.
// contentName karşılık gelen {{define "X_content"}} bloğunun adı.
func (s *Server) render(w http.ResponseWriter, contentName string, data map[string]any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl, err := s.templates.Clone()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// "content" adında bir alias yarat ki layout {{template "content" .}} çağrısı doğru sayfaya gitsin.
	if _, err := tmpl.New("content").Parse(`{{template "` + contentName + `" .}}`); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) tryAcquireScan() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeScan != 0 {
		return false
	}
	s.activeScan = -1 // rezerve
	return true
}

func (s *Server) setActiveScan(id int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeScan = id
}

func (s *Server) releaseScan() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activeScan = 0
}

func (s *Server) runScanAsync(ctx context.Context) {
	defer s.releaseScan()
	res, err := s.Engine.Run(ctx, nil)
	if err == nil && res != nil {
		s.setActiveScan(res.ScanID)
	}
}
