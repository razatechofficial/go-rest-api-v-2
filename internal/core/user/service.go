package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
	"github.com/razatechofficial/go-rest-api-v-2/internal/ports"
	apperrors "github.com/razatechofficial/go-rest-api-v-2/pkg/errors"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

type service struct {
	repo        Repository
	orderLister ports.OrderLister
}

func NewService(repo Repository, orderLister ports.OrderLister) Service {
	return &service{repo: repo, orderLister: orderLister}
}

// SetOrderLister injects OrderLister (called by container for two-phase wiring).
func (s *service) SetOrderLister(l ports.OrderLister) {
	s.orderLister = l
}

func (s *service) Create(ctx context.Context, dto CreateUserDTO) (*domain.User, error) {
	dto.Email = strings.ToLower(strings.TrimSpace(dto.Email))
	dto.Name = strings.TrimSpace(dto.Name)

	exists, err := s.repo.ExistsByEmail(ctx, dto.Email)
	if err != nil {
		return nil, fmt.Errorf("checking email: %w", err)
	}
	if exists {
		return nil, apperrors.ErrAlreadyExists
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(dto.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hashing password: %w", err)
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:        domain.UserID(uuid.New().String()),
		Name:      dto.Name,
		Email:     dto.Email,
		Password:  string(hashed),
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	logger.Info("user created",
		logger.String("user_id", string(user.ID)),
		logger.String("email", user.Email),
	)

	return user, nil
}

func (s *service) GetByID(ctx context.Context, id domain.UserID) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("getting user: %w", err)
	}
	return user, nil
}

func (s *service) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.repo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("getting user by email: %w", err)
	}
	return user, nil
}

// List passes the full UserListQuery to repository.
// total is returned alongside users — handler uses it to build pagination.Meta.
// service never builds Meta — that is the handler's responsibility.
func (s *service) List(ctx context.Context, query UserListQuery) ([]*domain.User, int, error) {
	users, total, err := s.repo.FindAll(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("listing users: %w", err)
	}
	return users, total, nil
}

func (s *service) Update(ctx context.Context, id domain.UserID, dto UpdateUserDTO) (*domain.User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if dto.Email != "" {
		dto.Email = strings.ToLower(strings.TrimSpace(dto.Email))
		if dto.Email != user.Email {
			exists, err := s.repo.ExistsByEmail(ctx, dto.Email)
			if err != nil {
				return nil, fmt.Errorf("checking email: %w", err)
			}
			if exists {
				return nil, apperrors.ErrAlreadyExists
			}
			user.Email = dto.Email
		}
	}

	if dto.Name != "" {
		user.Name = strings.TrimSpace(dto.Name)
	}

	user.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}

	logger.Info("user updated",
		logger.String("user_id", string(user.ID)),
	)

	return user, nil
}

func (s *service) Delete(ctx context.Context, id domain.UserID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	logger.Info("user deleted",
		logger.String("user_id", string(id)),
	)
	return nil
}

func (s *service) IsActive(ctx context.Context, id domain.UserID) (bool, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("checking active status: %w", err)
	}
	return user.IsActive, nil
}

func (s *service) ListOrdersForUser(ctx context.Context, userID string) ([]*ports.OrderSummary, error) {
	if s.orderLister == nil {
		return nil, nil
	}
	return s.orderLister.OrdersByUser(ctx, userID)
}
