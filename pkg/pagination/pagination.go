package pagination

import (
	"net/http"
	"strconv"
)

const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Params holds pagination query parameters.
type Params struct {
	Page    int
	PerPage int
	Sort    string
	Order   string
}

// Meta holds pagination metadata for responses.
type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Paginated wraps data with pagination metadata.
type Paginated[T any] struct {
	Data []T  `json:"data"`
	Meta Meta `json:"meta"`
}

// FromRequest parses pagination params from an HTTP request.
func FromRequest(r *http.Request) Params {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	sort := r.URL.Query().Get("sort")
	order := r.URL.Query().Get("order")

	if page < 1 {
		page = DefaultPage
	}
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	if perPage > MaxPerPage {
		perPage = MaxPerPage
	}
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return Params{
		Page:    page,
		PerPage: perPage,
		Sort:    sort,
		Order:   order,
	}
}

// Offset returns the SQL offset for the current page.
func (p Params) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// Limit returns the SQL limit for the current page.
func (p Params) Limit() int {
	return p.PerPage
}

// NewMeta creates pagination metadata from total count.
func NewMeta(page, perPage int, total int64) Meta {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	if totalPages < 1 {
		totalPages = 1
	}

	return Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// NewPaginated creates a paginated response.
func NewPaginated[T any](data []T, page, perPage int, total int64) Paginated[T] {
	if data == nil {
		data = []T{}
	}
	return Paginated[T]{
		Data: data,
		Meta: NewMeta(page, perPage, total),
	}
}
