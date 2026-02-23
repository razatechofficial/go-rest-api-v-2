package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/razatechofficial/go-rest-api-v-2/config"
	"github.com/razatechofficial/go-rest-api-v-2/internal/app"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

func main() {

	//! ================================ CONFIGURATION LOADING ================================
	//* Step 1: Load configuration
	// Configuration comes from YAML files and environment variables
	cfg, err := config.Load()
	if err != nil {
		// Use standard logger since our logger isn't initialized yet
		panic("Failed to load configuration: " + err.Error())
	}

	fmt.Printf("Configuration loaded successfully: %+v\n", cfg)

	//! ================================ LOGGER INITIALIZATION ================================
	//* Step 2: Initialize logger FIRST (before any logging)
	// This creates the actual zap.Logger instance
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync() // Flush logs on exit

	//! ================================ APPLICATION BOOTSTRAP ================================
	//* 3. bootstrap application
	application, err := app.New(cfg)
	if err != nil {
		logger.Fatal("failed to initialize application",
			logger.Err(err),
		)
	}

	//! ================================ HTTP SERVER ================================
	//* 4. start HTTP server in goroutine
	go func() {
		if err := application.HTTP.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("http server error", logger.Err(err))
		}
	}()

	//* 5. block until SIGINT (Ctrl+C) or SIGTERM (k8s shutdown)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("shutdown signal received", logger.String("signal", sig.String()))

	//* 6. graceful shutdown — 10s to drain in-flight requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := application.HTTP.Shutdown(ctx); err != nil {
		logger.Fatal("forced shutdown", logger.Err(err))
	}
}
