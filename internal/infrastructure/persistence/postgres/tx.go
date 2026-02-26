// Package postgres: transaction support.
// Use Pool.Begin to start a transaction, then postgres.WithTx to attach it to context.
// Repositories use postgres.Querier(ctx, pool) so they automatically use the transaction
// when present — no need to pass Tx explicitly.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier is the interface used for all DB operations (pool or transaction).
// Both *Pool and *Tx implement it so repositories can use either transparently.
type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Ensure *Pool implements Querier (Pool embeds *pgxpool.Pool which has these methods).
var _ Querier = (*Pool)(nil)

// Tx wraps pgx.Tx and implements Querier so it can be used wherever a pool is used.
// Start with Pool.Begin(ctx); commit with Commit(ctx) or rollback with Rollback(ctx).
type Tx struct {
	pgx.Tx
}

// Exec forwards to the underlying transaction.
func (t *Tx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.Tx.Exec(ctx, sql, args...)
}

// Query forwards to the underlying transaction.
func (t *Tx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.Tx.Query(ctx, sql, args...)
}

// QueryRow forwards to the underlying transaction.
func (t *Tx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.Tx.QueryRow(ctx, sql, args...)
}

// Commit commits the transaction.
func (t *Tx) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

// Rollback aborts the transaction. Safe to call multiple times or after Commit.
func (t *Tx) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}

// Ensure *Tx implements Querier.
var _ Querier = (*Tx)(nil)

// ── context keys ─────────────────────────────────────────────────────────────

type txKey struct{}

// WithTx attaches a transaction to the context.
// Repositories that use Querier(ctx, pool) will use this tx when present.
func WithTx(ctx context.Context, tx *Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// Querier returns the Querier to use for this request.
// If the context has a transaction (from WithTx), that is returned; otherwise the pool.
// Use this in every repository method so the same code path works with or without a transaction.
func QuerierFromContext(ctx context.Context, pool *Pool) Querier {
	if tx, ok := ctx.Value(txKey{}).(*Tx); ok && tx != nil {
		return tx
	}
	return pool
}

// Begin starts a new transaction from the pool.
// Call WithTx(ctx, tx) and pass that context to repository methods; then Commit or Rollback.
//
// Example:
//
//	tx, err := pool.Begin(ctx)
//	if err != nil { return err }
//	defer tx.Rollback(ctx)
//	ctxTx := postgres.WithTx(ctx, tx)
//	if err := userRepo.Create(ctxTx, user); err != nil { return err }
//	if err := orderRepo.Create(ctxTx, order); err != nil { return err }
//	return tx.Commit(ctx)
func (p *Pool) Begin(ctx context.Context) (*Tx, error) {
	pgxTx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: pgxTx}, nil
}
