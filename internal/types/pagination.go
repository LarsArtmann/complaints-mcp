package types

import (
	"fmt"
)

// PageRequest provides type-safe pagination request parameters.
type PageRequest struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MinPerPage     = 1
	MaxPerPage     = 100
)

// NewPageRequest creates a validated page request with defaults.
func NewPageRequest(page, perPage int) PageRequest {
	if page < 1 {
		page = DefaultPage
	}

	if perPage < MinPerPage {
		perPage = DefaultPerPage
	}

	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}

	return PageRequest{
		Page:    page,
		PerPage: perPage,
	}
}

// Offset returns the database offset for this page request.
func (pr PageRequest) Offset() int {
	return (pr.Page - 1) * pr.PerPage
}

// Limit returns the limit (per_page) for this page request.
func (pr PageRequest) Limit() int {
	return pr.PerPage
}

// IsValid returns true if the page request has valid values.
func (pr PageRequest) IsValid() bool {
	return pr.Page >= 1 &&
		pr.PerPage >= MinPerPage &&
		pr.PerPage <= MaxPerPage
}

// PageResponse provides type-safe paginated response.
type PageResponse[T any] struct {
	Data       []T  `json:"data"`
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	TotalItems int  `json:"total_items"`
	TotalPages int  `json:"total_pages"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

// NewPageResponse creates a paginated response from data.
func NewPageResponse[T any](data []T, req PageRequest, totalItems int) PageResponse[T] {
	totalPages := totalItems / req.PerPage
	if totalItems%req.PerPage > 0 {
		totalPages++
	}

	if totalPages == 0 {
		totalPages = 1
	}

	return PageResponse[T]{
		Data:       data,
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}
}

// IsEmpty returns true if there are no items in the response.
func (pr PageResponse[T]) IsEmpty() bool {
	return len(pr.Data) == 0
}

// EmptyPageResponse creates an empty paginated response.
func EmptyPageResponse[T any](req PageRequest) PageResponse[T] {
	return PageResponse[T]{
		Data:       []T{},
		Page:       req.Page,
		PerPage:    req.PerPage,
		TotalItems: 0,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    req.Page > 1,
	}
}

// CursorRequest provides cursor-based pagination request.
type CursorRequest struct {
	Cursor string `json:"cursor,omitempty"`
	Limit  int    `json:"limit"`
}

// NewCursorRequest creates a validated cursor request with defaults.
func NewCursorRequest(cursor string, limit int) CursorRequest {
	if limit < MinPerPage {
		limit = DefaultPerPage
	}

	if limit > MaxPerPage {
		limit = MaxPerPage
	}

	return CursorRequest{
		Cursor: cursor,
		Limit:  limit,
	}
}

// CursorResponse provides cursor-based paginated response.
type CursorResponse[T any] struct {
	NextCursor string `json:"next_cursor,omitempty"`
	Data       []T    `json:"data"`
	TotalCount int    `json:"total_count,omitempty"`
	HasMore    bool   `json:"has_more"`
}

// PaginationError represents pagination-related errors.
type PaginationError struct {
	Field   string
	Message string
}

func (e PaginationError) Error() string {
	return fmt.Sprintf("pagination error for %s: %s", e.Field, e.Message)
}
