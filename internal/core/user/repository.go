package user

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

func (r *repository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, name, email, password, is_active, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Name,
		user.Email,
		user.Password,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting user: %w", err)
	}
	return nil
}

func (r *repository) FindByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	query := `
        SELECT id, name, email, password, is_active, created_at, updated_at
        FROM   users
        WHERE  id         = $1
          AND  deleted_at IS NULL
    `
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by id: %w", err)
	}
	return u, nil
}

func (r *repository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
        SELECT id, name, email, password, is_active, created_at, updated_at
        FROM   users
        WHERE  email      = $1
          AND  deleted_at IS NULL
    `
	u := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.IsActive,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperr.ErrNotFound
		}
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	return u, nil
}

func (r *repository) FindAll(ctx context.Context, query UserListQuery) ([]*domain.User, int, error) {
	conditions := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	idx := 1

	// user-specific filter conditions
	if query.Search != "" {
		conditions = append(conditions,
			fmt.Sprintf("(name ILIKE $%d OR email ILIKE $%d)", idx, idx+1),
		)
		pattern := "%" + query.Search + "%"
		args = append(args, pattern, pattern)
		idx += 2
	}

	if query.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", idx))
		args = append(args, *query.IsActive)
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

	// count total matching records
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", where)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}

	// whitelist sort columns — never interpolate user input directly
	sortColumn := map[string]string{
		"name":       "name",
		"email":      "email",
		"created_at": "created_at",
		"is_active":  "is_active",
	}

	column := "created_at"
	if col, ok := sortColumn[query.Pagination.SortBy]; ok {
		column = col
	}

	dir := "DESC"
	if query.Pagination.SortOrder == pagination.SortAsc {
		dir = "ASC"
	}

	// paginated data query
	dataQuery := fmt.Sprintf(`
		SELECT id, name, email, is_active, created_at, updated_at
		FROM   users
		%s
		ORDER  BY %s %s
		LIMIT  $%d OFFSET $%d
	`, where, column, dir, idx, idx+1)

	args = append(args, query.Pagination.Limit, query.Pagination.Offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Email,
			&u.IsActive, &u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating users: %w", err)
	}

	return users, total, nil
}

func (r *repository) Update(ctx context.Context, user *domain.User) error {
	query := `
        UPDATE users
        SET    name       = $1,
               email      = $2,
               updated_at = $3
        WHERE  id         = $4
          AND  deleted_at IS NULL
    `
	result, err := r.db.Exec(ctx, query,
		user.Name,
		user.Email,
		time.Now().UTC(),
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *repository) Delete(ctx context.Context, id domain.UserID) error {
	query := `
        UPDATE users
        SET    deleted_at = $1
        WHERE  id         = $2
          AND  deleted_at IS NULL
    `
	result, err := r.db.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return apperr.ErrNotFound
	}
	return nil
}

func (r *repository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1 FROM users
            WHERE  email      = $1
              AND  deleted_at IS NULL
        )
    `
	var exists bool
	if err := r.db.QueryRow(ctx, query, email).Scan(&exists); err != nil {
		return false, fmt.Errorf("checking email existence: %w", err)
	}
	return exists, nil
}
