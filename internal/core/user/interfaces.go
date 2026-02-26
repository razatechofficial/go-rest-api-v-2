package user

import (
	"context"

	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/pagination"
)

// Service is the public contract of the user module.
type Service interface {
	Create(ctx context.Context, dto CreateUserDTO) (*domain.User, error)
	GetByID(ctx context.Context, id domain.UserID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	// List takes UserListQuery — not generic pagination.Params
	// service knows about user-specific filters
	List(ctx context.Context, query UserListQuery) ([]*domain.User, int, error)
	Update(ctx context.Context, id domain.UserID, dto UpdateUserDTO) (*domain.User, error)
	Delete(ctx context.Context, id domain.UserID) error
	IsActive(ctx context.Context, id domain.UserID) (bool, error)
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