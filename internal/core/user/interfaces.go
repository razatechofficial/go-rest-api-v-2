package user

import (
	"context"

	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
	"github.com/razatechofficial/go-rest-api-v-2/internal/ports"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/pagination"
)

// Service is the public contract of the user module.
type Service interface {
	Create(ctx context.Context, dto CreateUserDTO) (*domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context, query UserListQuery) ([]*domain.User, int, error)
	// ListOrdersForUser returns orders for the given user via ports.OrderLister (set in container).
	ListOrdersForUser(ctx context.Context, userID string) ([]*ports.OrderSummary, error)
	Update(ctx context.Context, id domain.UserID, dto UpdateUserDTO) (*domain.User, error)
	Delete(ctx context.Context, id domain.UserID) error
	IsActive(ctx context.Context, id domain.UserID) (bool, error)
}

// OrderListerSetter is optional: container uses it to inject OrderLister after construction.
// Not on Service interface so callers don't need to know about it.
type OrderListerSetter interface {
	SetOrderLister(ports.OrderLister)
}

// Repository is the data access contract.
type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	// FindAll takes UserListQuery — builds WHERE clause from user-specific filters
	FindAll(ctx context.Context, query UserListQuery) ([]*domain.User, int, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id domain.UserID) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
}

// keep pagination import used in UserListQuery
var _ = pagination.Params{}