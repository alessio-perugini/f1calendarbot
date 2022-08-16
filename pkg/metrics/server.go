package metrics

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

var (
	ErrNilServer = errors.New("error: nil metrics server")
)

type Server struct {
	srv     *http.Server
	handler http.Handler
}

func (s *Server) ListenAndServe(addr string) error {
	log.Info().Msgf("metrics: listening on %s", addr)

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

func NewServer() *Server {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return &Server{handler: mux}
}
