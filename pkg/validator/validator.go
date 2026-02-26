// Package validator provides structured request validation.
// Wraps go-playground/validator with field-level error extraction.
// Returns response.FieldError slice — ready to send directly to client.
package validator

import (
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/razatechofficial/go-rest-api-v-2/pkg/response"
)

var (
	instance *validator.Validate
	once     sync.Once
)

// get returns the singleton validator instance.
// Initialized once — thread safe via sync.Once.
func get() *validator.Validate {
	once.Do(func() {
		instance = validator.New()

		// use json tag name in errors instead of struct field name
		// so errors say "email" not "Email"
		instance.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	})
	return instance
}

// Validate validates a struct and returns field-level errors.
// Returns nil if validation passes.
// Returns []FieldError if validation fails — ready for response.ValidationFailed().
func Validate(s interface{}) []response.FieldError {
	err := get().Struct(s)
	if err == nil {
		return nil
	}

	var fieldErrors []response.FieldError
	for _, e := range err.(validator.ValidationErrors) {
		fieldErrors = append(fieldErrors, response.FieldError{
			Field:   e.Field(),
			Message: fieldMessage(e),
		})
	}
	return fieldErrors
}

// fieldMessage converts a validator error into a human-readable message.
// Clients see these messages — make them clear and actionable.
func fieldMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "max":
		return e.Field() + " must be at most " + e.Param() + " characters"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	case "url":
		return "must be a valid URL"
	case "uuid":
		return e.Field() + " must be a valid UUID"
	case "numeric":
		return e.Field() + " must be a number"
	case "alphanum":
		return e.Field() + " must contain only letters and numbers"
	case "gte":
		return e.Field() + " must be greater than or equal to " + e.Param()
	case "lte":
		return e.Field() + " must be less than or equal to " + e.Param()
	case "len":
		return e.Field() + " must be exactly " + e.Param() + " characters"
	default:
		return e.Field() + " is invalid"
	}
}
