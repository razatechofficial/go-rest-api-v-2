package order

import (
	"time"

	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
)

type OrderResponse struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Status           string `json:"status"`
	TotalAmountCents int64  `json:"total_amount_cents"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

func toResponse(o *domain.Order) *OrderResponse {
	return &OrderResponse{
		ID:               string(o.ID),
		UserID:           string(o.UserID),
		Status:           o.Status,
		TotalAmountCents: o.TotalAmountCents,
		CreatedAt:        o.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        o.UpdatedAt.Format(time.RFC3339),
	}
}

func toResponseList(orders []*domain.Order) []*OrderResponse {
	result := make([]*OrderResponse, len(orders))
	for i, o := range orders {
		result[i] = toResponse(o)
	}
	return result
}
