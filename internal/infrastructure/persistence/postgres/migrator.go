package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	// postgres driver for migrate
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// file source — reads .sql files from disk
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/razatechofficial/go-rest-api-v-2/config"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

// Migrator wraps golang-migrate with logging and error handling.
// Used by cmd/migrate/main.go — not by the API server directly.
// The API server never auto-migrates in production —
// migrations are always run as a separate step before deployment.
type Migrator struct {
	migrate *migrate.Migrate
	cfg     *config.Config
}

// NewMigrator creates a migrator pointing at the migrations directory.
// migrationsPath is the path to the folder containing .sql files.
// Typically "migrations" relative to the project root.
func NewMigrator(cfg *config.Config, migrationsPath string) (*Migrator, error) {
	// file:// prefix tells migrate to read from the local filesystem
	sourceURL := fmt.Sprintf("file://%s", migrationsPath)

	// use the same connection string as the main app
	m, err := migrate.New(sourceURL, cfg.Database.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("creating migrator: %w", err)
	}

	return &Migrator{
		migrate: m,
		cfg:     cfg,
	}, nil
}

// Up applies all pending migrations.
// Safe to run multiple times — already applied migrations are skipped.
// This is what you run before every deployment.
func (m *Migrator) Up() error {
	logger.Info("running migrations",
		logger.String("database", m.cfg.Database.Database),
		logger.String("host", m.cfg.Database.Host),
	)

	if err := m.migrate.Up(); err != nil {
		// ErrNoChange means all migrations are already applied — not an error
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("no pending migrations — database is up to date")
			return nil
		}
		return fmt.Errorf("applying migrations: %w", err)
	}

	version, _, _ := m.migrate.Version()
	logger.Info("migrations applied successfully",
		logger.Int("version", int(version)),
	)

	return nil
}

// Down rolls back the most recently applied migration.
// Use with caution in production — always have a backup first.
// Typically only used in development and CI.
func (m *Migrator) Down() error {
	logger.Info("rolling back last migration")

	if err := m.migrate.Steps(-1); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("no migrations to roll back")
			return nil
		}
		return fmt.Errorf("rolling back migration: %w", err)
	}

	version, _, _ := m.migrate.Version()
	logger.Info("migration rolled back",
		logger.Int("version", int(version)),
	)

	return nil
}

// DownAll rolls back ALL migrations to a clean state.
// DANGEROUS in production — drops all tables.
// Only use in development to reset the database completely.
func (m *Migrator) DownAll() error {
	logger.Info("rolling back ALL migrations — this will drop all tables")

	if err := m.migrate.Down(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Info("no migrations to roll back")
			return nil
		}
		return fmt.Errorf("rolling back all migrations: %w", err)
	}

	logger.Info("all migrations rolled back")
	return nil
}

// Version returns the current migration version applied to the database.
// Useful for debugging and health checks.
func (m *Migrator) Version() (uint, bool, error) {
	version, dirty, err := m.migrate.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			// no migrations applied yet — version 0
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("getting migration version: %w", err)
	}
	return version, dirty, nil
}

// Close releases the migrator's database connection.
// Always call this when done with the migrator.
func (m *Migrator) Close() error {
	sourceErr, dbErr := m.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("closing migration source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("closing migration db: %w", dbErr)
	}
	return nil
}
