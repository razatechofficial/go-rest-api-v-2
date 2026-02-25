// Package main is the entry point for the database migration tool.
// Run this before starting the API server to apply pending migrations.
//
// Usage:
//
//	go run cmd/migrate/main.go up        — apply all pending migrations
//	go run cmd/migrate/main.go down      — rollback last migration
//	go run cmd/migrate/main.go down-all  — rollback all (dev only)
//	go run cmd/migrate/main.go version   — show current version
//
// Or via Makefile:
//
//	make migrate-up
//	make migrate-down
//	make migrate-version
package main

import (
	"fmt"
	"os"

	"github.com/razatechofficial/go-rest-api-v-2/config"
	"github.com/razatechofficial/go-rest-api-v-2/internal/infrastructure/persistence/postgres"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

// migrationsPath is relative to the project root.
// When running via Makefile or go run from project root, this resolves correctly.
const migrationsPath = "migrations"

func main() {
	// ── 1. parse command ──────────────────────────────────────────
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	command := os.Args[1]

	// ── 2. load config ────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	// ── 3. init logger ────────────────────────────────────────────
	if err := logger.Init(cfg.Log.Level, cfg.Log.Format); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	// ── 4. create migrator ────────────────────────────────────────
	migrator, err := postgres.NewMigrator(cfg, migrationsPath)
	if err != nil {
		logger.Fatal("failed to create migrator", logger.Err(err))
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			logger.Error("error closing migrator", logger.Err(err))
		}
	}()

	// ── 5. execute command ────────────────────────────────────────
	switch command {
	case "up":
		if err := migrator.Up(); err != nil {
			logger.Fatal("migration up failed", logger.Err(err))
		}

	case "down":
		if err := migrator.Down(); err != nil {
			logger.Fatal("migration down failed", logger.Err(err))
		}

	case "down-all":
		// safety check — do not allow in production
		if cfg.App.IsProduction() {
			logger.Fatal("down-all is not allowed in production")
		}
		if err := migrator.DownAll(); err != nil {
			logger.Fatal("migration down-all failed", logger.Err(err))
		}

	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			logger.Fatal("failed to get migration version", logger.Err(err))
		}
		fmt.Printf("current migration version: %d (dirty: %v)\n", version, dirty)

	default:
		fmt.Printf("unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: migrate <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  up         apply all pending migrations")
	fmt.Println("  down       rollback last migration")
	fmt.Println("  down-all   rollback all migrations (development only)")
	fmt.Println("  version    show current migration version")
}
