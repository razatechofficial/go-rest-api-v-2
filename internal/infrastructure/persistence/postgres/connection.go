// Package database provides PostgreSQL connectivity using pgx.
// pgx is the fastest and most feature-complete PostgreSQL driver for Go.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/razatechofficial/go-rest-api-v-2/config"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

// Pool wraps pgxpool.Pool with additional functionality.
// This is the main database handle used throughout the application.
type Pool struct {
	*pgxpool.Pool
}

// NewConnection creates a connection pool with optimal settings.
// The pool manages multiple connections for concurrent use.
func NewConnection(cfg config.DatabaseConfig) (*Pool, error) {
	// Parse connection string into pgx config
	connConfig, err := pgxpool.ParseConfig(cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("parse connection string: %w", err)
	}

	// Connection pool configuration
	// These settings prevent resource exhaustion and ensure performance

	// MaxConns limits total open connections
	// Too many = database overload, Too few = app waits for connections
	connConfig.MaxConns = cfg.MaxOpenConns

	// MinConns keeps idle connections ready
	// Helps handle traffic spikes without connection creation delay
	connConfig.MinConns = cfg.MinIdleConns

	// MaxConnLifetime prevents using stale connections
	// Important for long-running apps (connection reset by database)
	connConfig.MaxConnLifetime = cfg.MaxConnLifetime

	// MaxConnIdleTime closes unused connections to save resources
	connConfig.MaxConnIdleTime = cfg.MaxConnIdleTime

	// HealthCheckPeriod checks connection validity periodically
	connConfig.HealthCheckPeriod = 5 * time.Minute

	// Optimize prepared statement caching
	// pgx automatically prepares and caches statements for reuse
	connConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	// AfterConnect hook runs when new connection is established
	// Use for session-level settings
	connConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		// Set application name for database monitoring (pg_stat_activity)
		_, err := conn.Exec(ctx, "SET application_name = 'enterprise-api'")
		return err
	}

	// Create the pool
	pool, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// Verify connectivity with ping
	// This fails fast if database is unreachable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	logger.Info("Database connection pool established",
		logger.String("host", cfg.Host),
		logger.Int("port", cfg.Port),
		logger.Int32("max_conns", cfg.MaxOpenConns),
	)

	return &Pool{pool}, nil
}

// Close gracefully shuts down the connection pool.
// Call this on application shutdown to release resources.
func (p *Pool) Close() {
	logger.Info("Closing database connection pool")
	p.Pool.Close()
}
