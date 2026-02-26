package user

import (
	"github.com/gin-gonic/gin"

	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
	"github.com/razatechofficial/go-rest-api-v-2/internal/ports"
	apperrors "github.com/razatechofficial/go-rest-api-v-2/pkg/errors"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/logger"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/pagination"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/response"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/validator"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Create handles POST /api/v1/users
func (h *Handler) Create(ctx *gin.Context) {
	var dto CreateUserDTO

	if err := ctx.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}

	if fieldErrors := validator.Validate(dto); fieldErrors != nil {
		response.ValidationFailed(ctx, fieldErrors)
		return
	}

	user, err := h.service.Create(ctx.Request.Context(), dto)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	response.Created(ctx, toResponse(user))
}

// GetByID handles GET /api/v1/users/:id
func (h *Handler) GetByID(ctx *gin.Context) {
	id := domain.UserID(ctx.Param("id"))

	user, err := h.service.GetByID(ctx.Request.Context(), id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	response.OK(ctx, toResponse(user))
}

// List handles GET /api/v1/users
func (h *Handler) List(ctx *gin.Context) {
	// parse universal pagination params + user-specific filters
	// all query string parsing lives in dto.go — handler stays clean
	query := UserListQueryFromContext(ctx)

	users, total, err := h.service.List(ctx.Request.Context(), query)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	// handler builds Meta — service and repository never touch Meta
	meta := pagination.NewMeta(query.Pagination, total)

	response.List(ctx, toResponseList(users), meta)
}

// Update handles PUT /api/v1/users/:id
func (h *Handler) Update(ctx *gin.Context) {
	id := domain.UserID(ctx.Param("id"))

	var dto UpdateUserDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}

	if fieldErrors := validator.Validate(dto); fieldErrors != nil {
		response.ValidationFailed(ctx, fieldErrors)
		return
	}

	user, err := h.service.Update(ctx.Request.Context(), id, dto)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	response.OK(ctx, toResponse(user))
}

// Delete handles DELETE /api/v1/users/:id
func (h *Handler) Delete(ctx *gin.Context) {
	id := domain.UserID(ctx.Param("id"))

	if err := h.service.Delete(ctx.Request.Context(), id); err != nil {
		h.handleError(ctx, err)
		return
	}

	response.NoContent(ctx)
}

// ListOrders handles GET /api/v1/users/:id/orders (list orders for a user via ports.OrderLister).
func (h *Handler) ListOrders(ctx *gin.Context) {
	userID := ctx.Param("id")

	orders, err := h.service.ListOrdersForUser(ctx.Request.Context(), userID)
	if err != nil {
		h.handleError(ctx, err)
		return
	}

	if orders == nil {
		orders = []*ports.OrderSummary{}
	}
	response.OK(ctx, orders)
}

// handleError maps service errors to HTTP responses.
// Module-specific messages for known errors.
// Falls back to response.FromError() for everything else.
func (h *Handler) handleError(ctx *gin.Context, err error) {
	switch {
	case apperrors.IsNotFound(err):
		response.NotFound(ctx, "user not found")

	case apperrors.IsAlreadyExists(err):
		response.Conflict(ctx, "email already in use")

	case apperrors.IsUnauthorized(err):
		response.Unauthorized(ctx, "unauthorized")

	default:
		// unexpected error — log it before responding
		if apperrors.HTTPStatus(err) == 500 {
			logger.Error("unhandled error",
				logger.Err(err),
				logger.String("request_id", ctx.GetString("X-Request-ID")),
				logger.String("path", ctx.Request.URL.Path),
				logger.String("method", ctx.Request.Method),
			)
			response.InternalError(ctx)
			return
		}
		// known error — let response.FromError map it
		response.FromError(ctx, err)
	}
}
