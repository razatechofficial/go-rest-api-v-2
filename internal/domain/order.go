package domain

import "time"

type OrderID string

// OrderStatus values: pending, confirmed, shipped, cancelled
const (
	OrderStatusPending   = "pending"
	OrderStatusConfirmed = "confirmed"
	OrderStatusShipped   = "shipped"
	OrderStatusCancelled = "cancelled"
)

type Order struct {
	ID                OrderID
	UserID            UserID
	Status            string
	TotalAmountCents  int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}
