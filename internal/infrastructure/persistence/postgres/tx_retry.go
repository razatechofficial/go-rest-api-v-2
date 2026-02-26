// Package postgres: transaction with retry for retryable errors.
// Use RunWithRetry when a unit of work must be retried on serialization failure or deadlock.
package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

// PostgreSQL retryable error codes (retry the whole transaction).
const (
	CodeSerializationFailure = "40001" // serialization_failure
	CodeDeadlockDetected     = "40P01" // deadlock_detected
)

// RetryConfig configures retry behavior for RunWithRetry.
type RetryConfig struct {
	// MaxAttempts is the total number of attempts (first try + retries). Default 3.
	MaxAttempts int
	// InitialBackoff is the delay before the first retry. Doubles each retry. Default 50ms.
	InitialBackoff time.Duration
}

// DefaultRetryConfig returns a sensible default (3 attempts, 50ms initial backoff).
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:    3,
		InitialBackoff: 50 * time.Millisecond,
	}
}

// isRetryable returns true for errors that are safe to retry (whole transaction).
func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	switch pgErr.Code {
	case CodeSerializationFailure, CodeDeadlockDetected:
		return true
	default:
		return false
	}
}

// RunWithRetry runs fn inside a transaction and retries on retryable errors.
// fn receives a context with the transaction attached; all repo calls using
// QuerierFromContext(ctx, pool) will use that transaction.
//
// Retryable: 40001 (serialization_failure), 40P01 (deadlock_detected).
// On success, the transaction is committed. On non-retryable error or after
// max attempts, the transaction is rolled back and the error is returned.
//
// Usage:
//
//	err := postgres.RunWithRetry(ctx, pool, postgres.DefaultRetryConfig(), func(ctxTx context.Context) error {
//	    if err := userRepo.Create(ctxTx, user); err != nil { return err }
//	    if err := orderRepo.Create(ctxTx, order); err != nil { return err }
//	    return nil
//	})
func RunWithRetry(ctx context.Context, pool *Pool, cfg RetryConfig, fn func(ctxTx context.Context) error) error {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.InitialBackoff <= 0 {
		cfg.InitialBackoff = 50 * time.Millisecond
	}

	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}

		ctxTx := WithTx(ctx, tx)
		lastErr = fn(ctxTx)

		if lastErr == nil {
			if err := tx.Commit(ctx); err != nil {
				_ = tx.Rollback(ctx)
				return err
			}
			return nil
		}

		_ = tx.Rollback(ctx)

		if !isRetryable(lastErr) || attempt == cfg.MaxAttempts {
			return lastErr
		}

		logger.Warn("transaction retry",
			logger.Err(lastErr),
			logger.Int("attempt", attempt),
			logger.Int("max_attempts", cfg.MaxAttempts),
			logger.Duration("backoff", backoff),
		)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
			backoff *= 2
		}
	}

	return lastErr
}
