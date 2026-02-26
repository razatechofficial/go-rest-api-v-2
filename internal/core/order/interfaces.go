package order

import (
	"context"

	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
	"github.com/razatechofficial/go-rest-api-v-2/internal/ports"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/pagination"
)

// Service is the public contract of the order module.
type Service interface {
	Create(ctx context.Context, dto CreateOrderDTO) (*domain.Order, error)
	GetByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)
	List(ctx context.Context, query OrderListQuery) ([]*domain.Order, int, error)
	// OrdersByUser implements ports.OrderLister — used by user module to list a user's orders.
	OrdersByUser(ctx context.Context, userID string) ([]*ports.OrderSummary, error)
	Update(ctx context.Context, id domain.OrderID, dto UpdateOrderDTO) (*domain.Order, error)
	Delete(ctx context.Context, id domain.OrderID) error
}

// Repository is the data access contract.
type Repository interface {
	Create(ctx context.Context, order *domain.Order) error
	FindByID(ctx context.Context, id domain.OrderID) (*domain.Order, error)
	FindAll(ctx context.Context, query OrderListQuery) ([]*domain.Order, int, error)
	Update(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, id domain.OrderID) error
}

var _ = pagination.Params{}
