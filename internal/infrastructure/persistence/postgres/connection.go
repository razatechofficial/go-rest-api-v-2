// Package postgres provides PostgreSQL connectivity using pgx.
// pgx is the fastest and most feature-complete PostgreSQL driver for Go.
// This package is the only place in the entire application that
// imports pgx directly — everything else uses the Pool wrapper.
package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/razatechofficial/go-rest-api-v-2/config"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

// Pool wraps pgxpool.Pool.
// All database access in the application goes through this type.
// Use Begin(ctx) to start a transaction; attach it with postgres.WithTx(ctx, tx)
// and pass that context to repository methods so they participate in the same transaction.
// See tx.go for QuerierFromContext and transaction usage.
//
// Repositories receive *Pool and use postgres.QuerierFromContext(ctx, r.db)
// so they work with or without a transaction in context.
type Pool struct {
	*pgxpool.Pool
}

// NewConnection creates a PostgreSQL connection pool from config.
// Called once at startup from app.go — not from modules.
// Fails fast — if DB is unreachable at startup, the app should not start.
func NewConnection(cfg *config.Config) (*Pool, error) {
	// use ConnectionString() helper from your DatabaseConfig
	connConfig, err := pgxpool.ParseConfig(cfg.Database.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("parsing connection string: %w", err)
	}

	// ── pool sizing ───────────────────────────────────────────────
	// all values come from config — never hardcoded
	// tunable per environment via YAML or env vars
	connConfig.MaxConns = cfg.Database.MaxOpenConns
	connConfig.MinConns = cfg.Database.MinIdleConns
	connConfig.MaxConnLifetime = cfg.Database.MaxConnLifetime
	connConfig.MaxConnIdleTime = cfg.Database.MaxConnIdleTime

	// ── health check ──────────────────────────────────────────────
	// pgx periodically pings idle connections to detect stale ones
	// 1 minute is more aggressive than default — good for production
	connConfig.HealthCheckPeriod = 1 * time.Minute

	// ── query performance ─────────────────────────────────────────
	// CacheDescribe caches query parameter types after first execution
	// avoids a round-trip to DB for type info on repeated queries
	connConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeCacheDescribe

	// ── connection hook ───────────────────────────────────────────
	// runs every time a new connection is established in the pool
	// application_name appears in pg_stat_activity — essential for
	// identifying your app in DB monitoring tools (pgAdmin, Datadog, etc)
	// Note: SET does not support bound parameters ($1), so we quote the value safely.
	connConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		appName := strings.ReplaceAll(cfg.App.Name, "'", "''")
		_, err := conn.Exec(ctx, fmt.Sprintf("SET application_name = '%s'", appName))
		return err
	}

	// ── create pool ───────────────────────────────────────────────
	pool, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	// ── verify connectivity ───────────────────────────────────────
	// ping with timeout — fail fast if DB unreachable
	// 5 seconds is enough for local and cloud DBs
	// if this fails, app.go returns error and main.go exits — correct
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close() // release resources before returning error
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	logger.Info("database connection pool established",
		logger.String("host", cfg.Database.Host),
		logger.Int("port", cfg.Database.Port),
		logger.String("database", cfg.Database.Database),
		logger.String("ssl_mode", cfg.Database.SSLMode),
		logger.Int32("max_conns", cfg.Database.MaxOpenConns),
		logger.Int32("min_conns", cfg.Database.MinIdleConns),
	)

	return &Pool{pool}, nil
}

// Health checks if the database is reachable.
// Called by the /ready endpoint in router.go.
// k8s uses this to decide if the pod should receive traffic.
func (p *Pool) Health(ctx context.Context) error {
	return p.Ping(ctx)
}

// Close gracefully shuts down the connection pool.
// Called from app.go Shutdown() — after HTTP server drains.
// Waits for all active queries to finish before closing.
func (p *Pool) Close() {
	logger.Info("closing database connection pool")
	p.Pool.Close()
	logger.Info("database connection pool closed")
}
