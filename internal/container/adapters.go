// Package container: adapters so user and order modules satisfy ports without importing each other.
package container

import (
	"context"

	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
	"github.com/razatechofficial/go-rest-api-v-2/internal/ports"
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/order"
	"github.com/razatechofficial/go-rest-api-v-2/internal/core/user"
)

// userCheckerAdapter lets order module use user.Service as ports.UserChecker.
type userCheckerAdapter struct {
	user.Service
}

func (a userCheckerAdapter) IsUserActive(ctx context.Context, userID string) (bool, error) {
	return a.Service.IsActive(ctx, domain.UserID(userID))
}

// orderListerAdapter lets user module use order.Service as ports.OrderLister.
type orderListerAdapter struct {
	order.Service
}

func (a orderListerAdapter) OrdersByUser(ctx context.Context, userID string) ([]*ports.OrderSummary, error) {
	return a.Service.OrdersByUser(ctx, userID)
}

// Ensure adapters implement ports.
var (
	_ ports.UserChecker = (*userCheckerAdapter)(nil)
	_ ports.OrderLister = (*orderListerAdapter)(nil)
)
