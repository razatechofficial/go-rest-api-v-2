package order

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/pagination"
)

type CreateOrderDTO struct {
	UserID           string `json:"user_id"            binding:"required" validate:"required"`
	TotalAmountCents int64  `json:"total_amount_cents" binding:"required" validate:"required,gte=0"`
	Status           string `json:"status"             validate:"omitempty,oneof=pending confirmed shipped cancelled"`
}

type UpdateOrderDTO struct {
	Status string `json:"status" validate:"omitempty,oneof=pending confirmed shipped cancelled"`
}

// OrderListQuery holds query parameters for GET /orders.
type OrderListQuery struct {
	Pagination   pagination.Params
	UserID       domain.UserID  // filter by user
	Status       string         // ?status=pending
	CreatedFrom  *time.Time
	CreatedTo    *time.Time
}

// OrderListQueryFromContext parses query params for the list endpoint.
func OrderListQueryFromContext(ctx *gin.Context) OrderListQuery {
	q := OrderListQuery{
		Pagination: pagination.FromContext(ctx),
		Status:     ctx.Query("status"),
	}
	if uid := ctx.Query("user_id"); uid != "" {
		q.UserID = domain.UserID(uid)
	}
	q.CreatedFrom = parseDate(ctx.Query("created_from"))
	q.CreatedTo = parseDate(ctx.Query("created_to"))
	if q.CreatedTo != nil {
		endOfDay := q.CreatedTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		q.CreatedTo = &endOfDay
	}
	return q
}

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
