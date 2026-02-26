package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/razatechofficial/go-rest-api-v-2/internal/container"
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/order"
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/user"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

// RegisterRoutes mounts global middleware and all module route groups.
// Stores the container on the server for use by probe handlers.
// Called once from app.go after NewServer().
func (s *Server) RegisterRoutes(c *container.Container) {
	// store container for readinessProbe DB health check
	s.container = c

	// ── global middleware ─────────────────────────────────────────
	// s.engine.Use(
	// 	middleware.RequestID(),
	// 	middleware.Logger(),
	// 	middleware.Recovery(),
	// 	middleware.CORS(s.cfg.App.IsProduction()),
	// 	middleware.SecureHeaders(s.cfg.App.IsProduction()),
	// 	middleware.RateLimit(),
	// )

	// ── system routes ─────────────────────────────────────────────
	s.engine.GET("/health", s.livenessProbe)
	s.engine.GET("/ready", s.readinessProbe)

	// ── api v1 ────────────────────────────────────────────────────
	v1 := s.engine.Group("/api/v1")
	{
		user.RegisterRoutes(v1, c.Handlers.User)
		order.RegisterRoutes(v1, c.Handlers.Order)
	}
}

// livenessProbe handles GET /health
// Returns 200 as long as the process is running.
// Does not check dependencies — just confirms the process is alive.
func (s *Server) livenessProbe(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"version": s.cfg.App.Version,
		"env":     s.cfg.App.Environment,
	})
}

// readinessProbe handles GET /ready
// Returns 200 only when all dependencies are healthy.
// Returns 503 if DB is unreachable — k8s stops routing traffic to this pod.
func (s *Server) readinessProbe(ctx *gin.Context) {
	if err := s.container.DB.Health(ctx.Request.Context()); err != nil {
		logger.Warn("readiness check failed — database unreachable",
			logger.Err(err),
		)
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unavailable",
			"reason": "database unreachable",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"version": s.cfg.App.Version,
	})
}
