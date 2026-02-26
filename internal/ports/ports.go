// Package ports holds application-level contracts (inbound/outbound) used across modules.
// Domain stays pure (entities only); ports live at the application boundary.
// Implementations are wired in the container (with adapters if needed).
package ports

import "context"

// UserChecker is implemented by the user module. Used by order (and others) to check user state.
type UserChecker interface {
	IsUserActive(ctx context.Context, userID string) (bool, error)
}

// OrderLister is implemented by the order module. Used by user (and others) to list orders for a user.
type OrderLister interface {
	OrdersByUser(ctx context.Context, userID string) ([]*OrderSummary, error)
}

// OrderSummary is a DTO so callers (e.g. user module) don't depend on order domain types.
type OrderSummary struct {
	ID               string `json:"id"`
	UserID           string `json:"user_id"`
	Status           string `json:"status"`
	TotalAmountCents int64  `json:"total_amount_cents"`
	CreatedAt        string `json:"created_at"`
}
