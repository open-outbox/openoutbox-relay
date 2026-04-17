package relay

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"go.uber.org/zap"
)

// Server is an HTTP server that exposes administrative and observability endpoints.
// It serves as a diagnostic window into the relay's operation, providing
// metrics and health status.
type Server struct {
	storage   Storage
	publisher Publisher
	server    *http.Server
	logger    *zap.Logger
}

func registerHandlers(s *Server, mux *http.ServeMux) {
	mux.HandleFunc("/stats", s.handleStats)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"OK"}`))
	})
	mux.HandleFunc("/readyz", s.handleHealthz)
}

// NewServer creates an instrumented HTTP server. It uses otelhttp to
// automatically generate trace spans for every incoming request.
func NewServer(
	ctx context.Context,
	s Storage,
	p Publisher,
	addr string,
	logger *zap.Logger,
) *Server {

	mux := http.NewServeMux()
	handler := otelhttp.NewHandler(mux, "server-request",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)

	srv := &Server{
		storage:   s,
		publisher: p,
		logger:    logger,
		server: &http.Server{
			Addr:         addr,
			BaseContext:  func(net.Listener) context.Context { return ctx },
			ReadTimeout:  time.Second,
			WriteTimeout: 10 * time.Second,
			Handler:      handler,
		},
	}
	registerHandlers(srv, mux)

	return srv
}

// handleStats returns a JSON snapshot of the outbox table metrics,
// such as the number of pending events and the age of the oldest record.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.storage.GetStats(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		s.logger.Error("failed to encode stats", zap.Error(err))
	}
}

// handleHealthz responds with 200 OK if the relay is healthy and can
// communicate with its storage and publisher backends.
func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// 1. Check DB
	if err := s.storage.Ping(ctx); err != nil {
		s.logger.Error("db unreachable", zap.Error(err))
		http.Error(w, "storage error", 503)
		return
	}

	// 2. Check Publisher (Optional but recommended)
	if err := s.publisher.Ping(ctx); err != nil {
		s.logger.Error("broker unreachable", zap.Error(err))
		http.Error(w, "broker error", 503)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"OK"}`))
}

// Start runs the HTTP server. It blocks until the provided context is canceled
// or the underlying listener returns an error. When the context is canceled,
// it performs a graceful shutdown with a 5-second timeout.
func (s *Server) Start(ctx context.Context) (err error) {

	srvErr := make(chan error, 1)
	go func() {
		s.logger.Info("starting relay api", zap.String("addr", s.server.Addr))
		srvErr <- s.server.ListenAndServe()
	}()

	select {
	case err = <-srvErr:
		return err
	case <-ctx.Done():
		s.logger.Info("Shut down signal received, shudding down api server...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = s.server.Shutdown(shutdownCtx)
	s.logger.Info("Api server stopped")
	return err

}
