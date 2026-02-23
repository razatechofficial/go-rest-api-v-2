package app

import (
	"context"

	"github.com/razatechofficial/go-rest-api-v-2/config"
	"github.com/razatechofficial/go-rest-api-v-2/internal/infrastructure/persistence/postgres"

	transporthttp "github.com/razatechofficial/go-rest-api-v-2/internal/transport/http"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

type App struct {
	HTTP *transporthttp.Server
	db   *postgres.Pool // kept so we can close it on shutdown
}

func New(cfg *config.Config) (*App, error) {
	// logger is already initialized in main.go before New() is called
	// just use it directly via your global logger package
	logger.Info("initializing application",
		logger.String("name", cfg.App.Name),
		logger.String("version", cfg.App.Version),
		logger.String("env", cfg.App.Environment),
	)

	//! ================================ INFRASTRUCTURE ================================
	//* Step 3: Connect to database
	db, err := postgres.NewConnection(cfg)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}

	// container — all DI wiring
	// c := container.New(cfg, db)

	//! ================================ TRANSPORT ================================
	//* Step 4: Create HTTP server
	httpServer := transporthttp.NewServer(cfg)
	// httpServer.RegisterRoutes(c)

	return &App{
		HTTP: httpServer,
		db:   db,
	}, nil
}

// Shutdown cleanly stops all resources in reverse startup order.
// HTTP server drains first — no new requests accepted.
// Then DB pool closes — all in-flight queries finish first.
func (a *App) Shutdown(ctx context.Context) error {
	logger.Info("shutting down application")

	// 1. stop HTTP first — no new requests accepted after this
	if err := a.HTTP.Shutdown(ctx); err != nil {
		// do not return here — still need to close DB
		// log the error and continue shutdown
		logger.Error("http server shutdown error", logger.Err(err))
	}

	// 2. close DB after HTTP is fully drained
	// at this point no in-flight HTTP handler can be making DB calls
	a.db.Close()

	logger.Info("application shutdown complete")
	return nil
}
