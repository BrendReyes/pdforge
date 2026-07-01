package server

import (
	"context"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/brendreyes/pdforge/web"
)

// Server owns the HTTP server, the SSE broker, and the result store.
type Server struct {
	http   *http.Server
	broker *broker
	store  *resultStore
}

// New creates a Server bound to addr.
func New(addr string) *Server {
	b := newBroker()
	st := newResultStore()

	srv := &Server{
		broker: b,
		store:  st,
	}

	mux := http.NewServeMux()
	srv.registerRoutes(mux)

	srv.http = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return srv
}

// Start begins listening and blocks until Shutdown is called or a fatal error occurs.
func (s *Server) Start() error {
	if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully stops the server with a 5-second deadline.
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.http.Shutdown(ctx)
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Embedded static assets under /static/
	staticFS, _ := fs.Sub(web.Static, "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Operation endpoints
	mux.HandleFunc("POST /api/optimize", s.handleOptimize)
	mux.HandleFunc("POST /api/merge", s.handleMerge)
	mux.HandleFunc("POST /api/convert", s.handleConvert)
	mux.HandleFunc("POST /api/rotate", s.handleRotate)
	mux.HandleFunc("POST /api/rmpage", s.handleRmpage)
	mux.HandleFunc("POST /api/split", s.handleSplit)

	// Progress and download
	mux.HandleFunc("GET /api/progress/{token}", s.broker.serveSSE)
	mux.HandleFunc("GET /api/download/{token}", s.handleDownload)

	// Index — serve the embedded frontend for all unmatched paths
	mux.HandleFunc("/", s.handleIndex)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	f, err := web.Static.Open("static/index.html")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, f)
}
