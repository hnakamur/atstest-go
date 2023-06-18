package atstest

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
)

type OriginServer struct {
	httpServer *http.Server
}

func NewOriginServer(addr string) *OriginServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := http.StatusOK
		if statusStr := r.FormValue("s"); statusStr != "" {
			var err error
			status, err = strconv.Atoi(statusStr)
			if err != nil {
				w.WriteHeader(http.StatusUnprocessableEntity)
				fmt.Fprintf(w, "invalid status query parameter: %s", statusStr)
				return
			}
		}

		body := ""
		if status >= http.StatusOK && status != http.StatusNoContent && status != http.StatusNotModified {
			body = fmt.Sprintf("This is a response for requestURI=%s\n", r.RequestURI)
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		}

		if sMaxAge := r.FormValue("s-maxage"); sMaxAge != "" {
			w.Header().Set("Cache-Control", "s-maxage="+sMaxAge)
		}
		w.WriteHeader(status)
		w.Write([]byte(body))
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
