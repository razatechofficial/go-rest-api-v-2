package order

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
	apperrors "github.com/razatechofficial/go-rest-api-v-2/pkg/errors"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, dto CreateOrderDTO) (*domain.Order, error) {
	status := domain.OrderStatusPending
	if dto.Status != "" {
		status = dto.Status
	}
	now := time.Now().UTC()
	ord := &domain.Order{
		ID:               domain.OrderID(uuid.New().String()),
		UserID:           domain.UserID(dto.UserID),
		Status:           status,
		TotalAmountCents: dto.TotalAmountCents,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.Create(ctx, ord); err != nil {
		return nil, fmt.Errorf("creating order: %w", err)
	}
	logger.Info("order created",
		logger.String("order_id", string(ord.ID)),
		logger.String("user_id", string(ord.UserID)),
	)
	return ord, nil
}

func (s *service) GetByID(ctx context.Context, id domain.OrderID) (*domain.Order, error) {
	ord, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("getting order: %w", err)
	}
	return ord, nil
}

func (s *service) List(ctx context.Context, query OrderListQuery) ([]*domain.Order, int, error) {
	orders, total, err := s.repo.FindAll(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("listing orders: %w", err)
	}
	return orders, total, nil
}

func (s *service) Update(ctx context.Context, id domain.OrderID, dto UpdateOrderDTO) (*domain.Order, error) {
	ord, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dto.Status != "" {
		ord.Status = dto.Status
	}
	if err := s.repo.Update(ctx, ord); err != nil {
		return nil, fmt.Errorf("updating order: %w", err)
	}
	logger.Info("order updated", logger.String("order_id", string(ord.ID)))
	return ord, nil
}

func (s *service) Delete(ctx context.Context, id domain.OrderID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	logger.Info("order deleted", logger.String("order_id", string(id)))
	return nil
}
