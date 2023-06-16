package main

import (
	"context"
	"net/http"
)

type OriginServer struct {
	httpServer *http.Server
}

func NewOriginServer(addr string) *OriginServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/status204", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "s-maxage=30")
		w.WriteHeader(http.StatusNoContent)
	})
	return &OriginServer{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}

func (s *OriginServer) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *OriginServer) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
