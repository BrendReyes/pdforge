package server

import (
	"context"
	"net/http"
	"time"
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

	// Static assets and index — wired up in Phase 5.
	mux.HandleFunc("/", s.handleIndex)
}

// handleIndex is replaced in Phase 5 with the real embedded frontend.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head><meta charset="utf-8"><title>pdforge</title></head>
<body style="font-family:monospace;padding:2rem">
<h1>pdforge serve</h1><p>Web GUI coming in Phase 5.</p>
</body>
</html>`))
}
