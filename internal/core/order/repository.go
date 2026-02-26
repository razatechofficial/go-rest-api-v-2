package order

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
	"github.com/razatechofficial/go-rest-api-v-2/internal/infrastructure/persistence/postgres"
	apperr "github.com/razatechofficial/go-rest-api-v-2/pkg/errors"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/pagination"
)

type repository struct {
	db *postgres.Pool
}

func NewRepository(db *postgres.Pool) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (id, user_id, status, total_amount_cents, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	q := postgres.QuerierFromContext(ctx, r.db)
	_, err := q.Exec(ctx, query,
		order.ID,
		order.UserID,
		order.Status,
		order.TotalAmountCents,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting order: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	query := `
		SELECT id, user_id, status, total_amount_cents, created_at, updated_at
		FROM   orders
		WHERE  id = $1 AND deleted_at IS NULL
	`
	o := &domain.Order{}
	q := postgres.QuerierFromContext(ctx, r.db)
	err := q.QueryRow(ctx, query, id).Scan(
		&o.ID,
		&o.UserID,
		&o.Status,
		&o.TotalAmountCents,
		&o.CreatedAt,
		&o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("finding order by id: %w", err)
	}
	return o, nil
}

func (r *repository) FindAll(ctx context.Context, query OrderListQuery) ([]*domain.Order, int, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	idx := 1

	if query.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", idx))
		args = append(args, query.UserID)
		idx++
	}
	if query.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", idx))
		args = append(args, query.Status)
		idx++
	}
	if query.CreatedFrom != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", idx))
		args = append(args, *query.CreatedFrom)
		idx++
	}
	if query.CreatedTo != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", idx))
		args = append(args, *query.CreatedTo)
		idx++
	}

	where := "WHERE " + strings.Join(conditions, " AND ")

	q := postgres.QuerierFromContext(ctx, r.db)
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders %s", where)
	if err := q.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting orders: %w", err)
	}

	sortColumn := map[string]string{
		"created_at": "created_at",
		"status":     "status",
		"total_amount_cents": "total_amount_cents",
	}
	column := "created_at"
	if col, ok := sortColumn[query.Pagination.SortBy]; ok {
		column = col
	}
	dir := "DESC"
	if query.Pagination.SortOrder == pagination.SortAsc {
		dir = "ASC"
	}

	dataQuery := fmt.Sprintf(`
		SELECT id, user_id, status, total_amount_cents, created_at, updated_at
		FROM   orders
		%s
		ORDER  BY %s %s
		LIMIT  $%d OFFSET $%d
	`, where, column, dir, idx, idx+1)
	args = append(args, query.Pagination.Limit, query.Pagination.Offset)

	rows, err := q.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying orders: %w", err)
	}
	defer rows.Close()

	orders := make([]*domain.Order, 0)
	for rows.Next() {
		o := &domain.Order{}
		if err := rows.Scan(
			&o.ID, &o.UserID, &o.Status, &o.TotalAmountCents, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating orders: %w", err)
	}
	return orders, total, nil
}

func (r *repository) Update(ctx context.Context, order *domain.Order) error {
	query := `
		UPDATE orders
		SET    status = $1, total_amount_cents = $2, updated_at = $3
		WHERE  id = $4 AND deleted_at IS NULL
	`
	q := postgres.QuerierFromContext(ctx, r.db)
	result, err := q.Exec(ctx, query,
		order.Status,
		order.TotalAmountCents,
		time.Now().UTC(),
		order.ID,
	)
	if err != nil {
		return fmt.Errorf("updating order: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id domain.OrderID) error {
	query := `
		UPDATE orders SET deleted_at = $1 WHERE id = $2 AND deleted_at IS NULL
	`
	q := postgres.QuerierFromContext(ctx, r.db)
	result, err := q.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("deleting order: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}
