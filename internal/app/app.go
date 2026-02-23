package app

import (
	"github.com/razatechofficial/go-rest-api-v-2/config"
	database "github.com/razatechofficial/go-rest-api-v-2/internal/infrastructure/persistence/postgres"

	transporthttp "github.com/razatechofficial/go-rest-api-v-2/internal/transport/http"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

type App struct {
	HTTP *transporthttp.Server
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
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer db.Close()

	// container — all DI wiring
	// c := container.New(cfg, db)

	//! ================================ TRANSPORT ================================
	//* Step 4: Create HTTP server
	httpServer := transporthttp.NewServer(cfg)
	// httpServer.RegisterRoutes(c)

	return &App{HTTP: httpServer}, nil
}
