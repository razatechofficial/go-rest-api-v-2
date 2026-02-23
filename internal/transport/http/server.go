package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/razatechofficial/go-rest-api-v-2/config"
	"github.com/razatechofficial/go-rest-api-v-2/internal/container"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

// Server wraps the HTTP server and gin engine.
// Separates lifecycle (start/stop) from routing concerns.
type Server struct {
	engine     *gin.Engine
	httpServer *http.Server
	cfg        *config.Config
	container  *container.Container // stored for probe handlers

}

// NewServer creates and configures the HTTP server.
// Does NOT start listening — call Start() for that.
func NewServer(cfg *config.Config) *Server {
	// set gin mode from config (debug/release/test)
	gin.SetMode(cfg.Server.Mode)

	// gin.New() — no default middleware
	// we attach everything explicitly so nothing is hidden
	engine := gin.New()

	srv := &Server{
		engine: engine,
		cfg:    cfg,
		httpServer: &http.Server{
			Addr:         cfg.Server.Address(), // uses your Address() helper
			Handler:      engine,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,

			// production hardening
			MaxHeaderBytes: 1 << 20, // 1MB max header size
		},
	}

	return srv
}

// Engine exposes gin engine so router.go can register routes and middleware.
// This is the only coupling point between server and router.
func (s *Server) Engine() *gin.Engine {
	return s.engine
}

// * Start begins accepting connections.
// Returns http.ErrServerClosed on clean shutdown — caller should handle this.
func (s *Server) Start() error {
	logger.Info("http server starting",
		logger.String("addr", s.httpServer.Addr),
		logger.String("env", s.cfg.App.Environment),
		logger.String("version", s.cfg.App.Version),
	)

	if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server failed: %w", err)
	}

	return nil
}

// * Shutdown gracefully drains in-flight requests.
// ctx controls how long we wait before forcing close.
// Call this on SIGINT/SIGTERM from main.go.
func (s *Server) Shutdown(ctx context.Context) error {
	logger.Info("http server shutting down gracefully")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("http server shutdown failed: %w", err)
	}

	logger.Info("http server stopped cleanly")
	return nil
}
