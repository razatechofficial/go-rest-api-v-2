package order

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/razatechofficial/go-rest-api-v-2/internal/domain"
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

func (h *Handler) Create(ctx *gin.Context) {
	var dto CreateOrderDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}
	if fieldErrors := validator.Validate(dto); fieldErrors != nil {
		response.ValidationFailed(ctx, fieldErrors)
		return
	}

	ord, err := h.service.Create(ctx.Request.Context(), dto)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	response.Created(ctx, toResponse(ord))
}

func (h *Handler) GetByID(ctx *gin.Context) {
	id := domain.OrderID(ctx.Param("id"))

	ord, err := h.service.GetByID(ctx.Request.Context(), id)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	response.OK(ctx, toResponse(ord))
}

func (h *Handler) List(ctx *gin.Context) {
	query := OrderListQueryFromContext(ctx)

	orders, total, err := h.service.List(ctx.Request.Context(), query)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	meta := pagination.NewMeta(query.Pagination, total)
	response.List(ctx, toResponseList(orders), meta)
}

func (h *Handler) Update(ctx *gin.Context) {
	id := domain.OrderID(ctx.Param("id"))

	var dto UpdateOrderDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(ctx, "invalid request body")
		return
	}
	if fieldErrors := validator.Validate(dto); fieldErrors != nil {
		response.ValidationFailed(ctx, fieldErrors)
		return
	}

	ord, err := h.service.Update(ctx.Request.Context(), id, dto)
	if err != nil {
		h.handleError(ctx, err)
		return
	}
	response.OK(ctx, toResponse(ord))
}

func (h *Handler) Delete(ctx *gin.Context) {
	id := domain.OrderID(ctx.Param("id"))

	if err := h.service.Delete(ctx.Request.Context(), id); err != nil {
		h.handleError(ctx, err)
		return
	}
	response.NoContent(ctx)
}

func (h *Handler) handleError(ctx *gin.Context, err error) {
	switch {
	case apperrors.IsNotFound(err):
		response.NotFound(ctx, "order not found")
	case apperrors.IsAlreadyExists(err):
		response.Conflict(ctx, "order already exists")
	case apperrors.IsUnauthorized(err):
		response.Unauthorized(ctx, "unauthorized")
	default:
		if apperrors.HTTPStatus(err) == http.StatusInternalServerError {
			logger.Error("unhandled error",
				logger.Err(err),
				logger.String("request_id", ctx.GetString("X-Request-ID")),
				logger.String("path", ctx.Request.URL.Path),
				logger.String("method", ctx.Request.Method),
			)
			response.InternalError(ctx)
			return
		}
		response.FromError(ctx, err)
	}
}
