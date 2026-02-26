// Package errors provides application-level error types and utilities.
// Two concerns:
//  1. Sentinel errors — shared across all modules (ErrNotFound etc)
//  2. Error utilities — AppError, wrapping, HTTPStatus, Code
//
// Modules import this package for sentinel errors. Handlers use HTTPStatus(err) and Code(err).
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ── sentinel errors ───────────────────────────────────────────────────────
// These are the only errors modules should return from their service layer.
// Handlers use errors.Is() and HTTPStatus() to map these to HTTP status codes.
// Never return raw DB or framework errors from services.

var (
	ErrNotFound           = errors.New("resource not found")
	ErrAlreadyExists      = errors.New("resource already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInactive           = errors.New("resource is inactive")
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
)

// ── app error type ────────────────────────────────────────────────────────

// AppError is a structured error with HTTP status, machine-readable code, and message.
// Use when you need richer context beyond sentinels.
//
// Example:
//
//	return errors.NewAppError(
//	    http.StatusBadRequest,
//	    "INVALID_DATE_RANGE",
//	    "start date must be before end date",
//	)
type AppError struct {
	HTTPStatus int    // HTTP status code to return
	Code       string // machine-readable code for clients
	Message    string // human-readable message
	Err        error  // optional underlying error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new AppError.
func NewAppError(status int, code, message string) *AppError {
	return &AppError{
		HTTPStatus: status,
		Code:       code,
		Message:    message,
	}
}

// Wrap wraps an existing error with AppError context.
func Wrap(err error, status int, code, message string) *AppError {
	return &AppError{
		HTTPStatus: status,
		Code:       code,
		Message:    message,
		Err:        err,
	}
}

// ── utility functions ─────────────────────────────────────────────────────

// IsNotFound returns true if the error is or wraps ErrNotFound.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists returns true if the error is or wraps ErrAlreadyExists.
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsUnauthorized returns true if the error is or wraps ErrUnauthorized.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// HTTPStatus returns the HTTP status code for err.
// Uses AppError when present; otherwise maps sentinels; else 500.
func HTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrInactive):
		return http.StatusForbidden
	case errors.Is(err, ErrInvalidInput):
		return http.StatusUnprocessableEntity
	case errors.Is(err, ErrInvalidCredentials):
		return http.StatusUnauthorized
	case errors.Is(err, ErrTokenExpired):
		return http.StatusUnauthorized
	case errors.Is(err, ErrTokenInvalid):
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

// Code returns the machine-readable error code for err.
// Returns "INTERNAL_SERVER_ERROR" when not an AppError or known sentinel.
func Code(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return "NOT_FOUND"
	case errors.Is(err, ErrAlreadyExists):
		return "CONFLICT"
	case errors.Is(err, ErrUnauthorized):
		return "UNAUTHORIZED"
	case errors.Is(err, ErrForbidden):
		return "FORBIDDEN"
	case errors.Is(err, ErrInactive):
		return "FORBIDDEN"
	case errors.Is(err, ErrInvalidInput):
		return "INVALID_INPUT"
	case errors.Is(err, ErrInvalidCredentials):
		return "INVALID_CREDENTIALS"
	case errors.Is(err, ErrTokenExpired):
		return "TOKEN_EXPIRED"
	case errors.Is(err, ErrTokenInvalid):
		return "TOKEN_INVALID"
	default:
		return "INTERNAL_SERVER_ERROR"
	}
}
