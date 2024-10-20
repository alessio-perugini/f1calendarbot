package metrics

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	ErrNilServer = errors.New("error: nil metrics server")
)

type Server struct {
	srv     *http.Server
	handler http.Handler
}

func (s *Server) ListenAndServe(addr string) error {
	slog.Info("metrics: listening on %s")

	s.srv = &http.Server{
		Addr:              addr,
		Handler:           s.handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s.srv.ListenAndServe()
}

func (s *Server) Close() error {
	if s.srv == nil {
		return ErrNilServer
	}

	return s.srv.Close()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.srv == nil {
		return ErrNilServer
	}

	return s.srv.Shutdown(ctx)
}

func NewServer() *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{handler: mux}
}
