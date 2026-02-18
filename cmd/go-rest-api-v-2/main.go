package main

import (
	"fmt"

	"github.com/razatechofficial/go-rest-api-v-2/config"
	database "github.com/razatechofficial/go-rest-api-v-2/internal/infrastructure/persistence/postgres"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

func main() {

	// Step 1: Load configuration
	// Configuration comes from YAML files and environment variables
	cfg, err := config.Load()
	if err != nil {
		// Use standard logger since our logger isn't initialized yet
		panic("Failed to load configuration: " + err.Error())
	}

	fmt.Printf("Configuration loaded successfully: %+v\n", cfg)

	// Step 2: Initialize logger FIRST (before any logging)
	// This creates the actual zap.Logger instance
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync() // Flush logs on exit

	logger.Info("Starting application",
		logger.String("name", cfg.App.Name),
		logger.String("version", cfg.App.Version),
		logger.String("environment", cfg.App.Environment),
	)

	// Step 3: Connect to database
	db, err := database.NewConnection(cfg.Database)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer db.Close()
}
