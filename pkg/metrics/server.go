package metrics

import (
	"context"
	"errors"
	"go.uber.org/zap"
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
	logger  *zap.Logger
}

func (s *Server) ListenAndServe(addr string) error {
	s.logger.Info("metrics: listening on %s")

	s.srv = &http.Server{
		Addr:              addr,
		Handler:           s.handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := s.srv.ListenAndServe(); err != nil {
		return err
	}

	return nil
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

func NewServer(logger *zap.Logger) *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{handler: mux, logger: logger}
}
