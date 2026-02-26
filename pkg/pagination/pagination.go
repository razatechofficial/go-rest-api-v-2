// Package pagination provides generic pagination and sort parameters.
// Filters are module-specific — each module defines its own Filter struct in its dto.go.
package pagination

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
	SortAsc      = "asc"
	SortDesc     = "desc"
)

// Params holds universal pagination and sort parameters.
// Module-specific filters are defined in each module's dto.go.
type Params struct {
	Page      int
	Limit     int
	Offset    int
	SortBy    string
	SortOrder string
}

// Meta holds complete pagination metadata for API responses.
// Frontend can show "Showing From to To of Total records".
type Meta struct {
	Total       int  `json:"total"`
	TotalPages  int  `json:"total_pages"`
	Page        int  `json:"page"`
	Limit       int  `json:"limit"`
	From        int  `json:"from"`
	To          int  `json:"to"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// FromContext extracts pagination and sort params from the request query.
//
// Supported: ?page=1&limit=20&sort_by=created_at&sort_order=asc
func FromContext(ctx *gin.Context) Params {
	page := parsePositiveInt(ctx.Query("page"), DefaultPage)
	limit := parsePositiveInt(ctx.Query("limit"), DefaultLimit)
	if limit > MaxLimit {
		limit = MaxLimit
	}

	sortOrder := strings.ToLower(ctx.Query("sort_order"))
	if sortOrder != SortAsc && sortOrder != SortDesc {
		sortOrder = SortDesc
	}

	return Params{
		Page:      page,
		Limit:     limit,
		Offset:    (page - 1) * limit,
		SortBy:    ctx.Query("sort_by"),
		SortOrder: sortOrder,
	}
}

// NewMeta builds pagination metadata for responses.
// total is the count of all matching records across all pages.
func NewMeta(params Params, total int) *Meta {
	totalPages := 0
	if total > 0 {
		totalPages = total / params.Limit
		if total%params.Limit > 0 {
			totalPages++
		}
	}

	from, to := 0, 0
	if total > 0 {
		from = params.Offset + 1
		to = params.Offset + params.Limit
		if to > total {
			to = total
		}
	}

	return &Meta{
		Total:       total,
		TotalPages:  totalPages,
		Page:        params.Page,
		Limit:       params.Limit,
		From:        from,
		To:          to,
		HasNext:     params.Page < totalPages,
		HasPrevious: params.Page > 1,
	}
}

func parsePositiveInt(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return fallback
	}
	return n
}
