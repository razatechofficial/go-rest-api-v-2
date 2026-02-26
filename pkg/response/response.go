// Package response provides standardized HTTP response helpers.
// Handlers use these instead of gin.H{} for a consistent response shape.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "github.com/razatechofficial/go-rest-api-v-2/pkg/errors"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/pagination"
)

// Response is the standard envelope for all API responses.
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Meta      interface{} `json:"meta,omitempty"`
	Error     *Error      `json:"error,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// Error is the standard error shape in responses.
type Error struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// FieldError is a single field validation failure.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ── success ───────────────────────────────────────────────────────────────

func OK(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{Success: true, Data: data})
}

func Created(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusCreated, Response{Success: true, Data: data})
}

func NoContent(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}

// List sends a paginated list response using pagination.Meta.
// Pass nil for meta to omit meta from the response.
func List(ctx *gin.Context, data interface{}, meta *pagination.Meta) {
	ctx.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// ── error ─────────────────────────────────────────────────────────────────

func BadRequest(ctx *gin.Context, message string) {
	sendError(ctx, http.StatusBadRequest, "BAD_REQUEST", message, nil)
}

func ValidationFailed(ctx *gin.Context, details []FieldError) {
	sendError(ctx, http.StatusUnprocessableEntity, "VALIDATION_FAILED", "validation failed", details)
}

func NotFound(ctx *gin.Context, message string) {
	sendError(ctx, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func Conflict(ctx *gin.Context, message string) {
	sendError(ctx, http.StatusConflict, "CONFLICT", message, nil)
}

func Unauthorized(ctx *gin.Context, message string) {
	sendError(ctx, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func Forbidden(ctx *gin.Context, message string) {
	sendError(ctx, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func InternalError(ctx *gin.Context) {
	sendError(ctx, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "an unexpected error occurred", nil)
}

// FromError maps any error to the correct HTTP response using pkg/errors.
// Handlers can call this instead of a switch on error type.
func FromError(ctx *gin.Context, err error) {
	sendError(ctx,
		apperrors.HTTPStatus(err),
		apperrors.Code(err),
		err.Error(),
		nil,
	)
}

func sendError(ctx *gin.Context, status int, code, message string, details []FieldError) {
	ctx.JSON(status, Response{
		Success:   false,
		RequestID: ctx.GetString("X-Request-ID"),
		Error: &Error{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
