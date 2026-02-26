package user

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/pagination"
)

type CreateUserDTO struct {
	Name     string `json:"name"     binding:"required" validate:"required,min=2,max=100"`
	Email    string `json:"email"    binding:"required" validate:"required,email"`
	Password string `json:"password" binding:"required" validate:"required,min=8,max=72"`
}

type UpdateUserDTO struct {
	Name  string `json:"name"  validate:"omitempty,min=2,max=100"`
	Email string `json:"email" validate:"omitempty,email"`
}


// ── list query ────────────────────────────────────────────────────────────

// UserListQuery holds all query parameters for GET /users.
// Combines universal pagination params with user-specific filters.
type UserListQuery struct {
	// universal — parsed by pagination package
	Pagination pagination.Params

	// user-specific filters
	Search      string     // ?search=john     — matches name or email
	IsActive    *bool      // ?is_active=true   — nil means no filter
	CreatedFrom *time.Time // ?created_from=2024-01-01
	CreatedTo   *time.Time // ?created_to=2024-12-31
}

// UserListQueryFromContext parses all query params for the list endpoint.
// Called in handler — keeps handler clean and parsing logic testable.
func UserListQueryFromContext(ctx *gin.Context) UserListQuery {
	q := UserListQuery{
		Pagination: pagination.FromContext(ctx), // universal params
	}

	// user-specific filters
	q.Search = ctx.Query("search")

	// is_active — three states: nil (all), true, false
	if v := ctx.Query("is_active"); v != "" {
		active := v == "true" || v == "1"
		q.IsActive = &active
	}

	// date range
	q.CreatedFrom = parseDate(ctx.Query("created_from"))
	q.CreatedTo   = parseDate(ctx.Query("created_to"))
	if q.CreatedTo != nil {
		// make end of day inclusive
		endOfDay := q.CreatedTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		q.CreatedTo = &endOfDay
	}

	return q
}

// parseDate is a local helper — only user module needs it here
// if other modules need it, promote to pkg/timeutil/
func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	formats := []string{time.RFC3339, "2006-01-02"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			utc := t.UTC()
			return &utc
		}
	}
	return nil
}