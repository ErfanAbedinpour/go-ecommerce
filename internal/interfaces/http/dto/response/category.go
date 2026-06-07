package response

import (
	"time"

	domain "app/internal/domain/category"
	"app/pkg/pagination"
)

// CategoryResponse is the category representation in API responses.
type CategoryResponse struct {
	ID          string             `json:"id"`
	ParentID    *string            `json:"parent_id,omitempty"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description string             `json:"description,omitempty"`
	ImageURL    string             `json:"image_url,omitempty"`
	SortOrder   int                `json:"sort_order"`
	IsActive      bool               `json:"is_active"`
	ProductsCount int64              `json:"products_count"`
	Children      []CategoryResponse `json:"children,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// CategoryListResponse is a paginated list of categories.
type CategoryListResponse struct {
	Data []CategoryResponse `json:"data"`
	Meta pagination.Meta    `json:"meta"`
}

// CategoryTreeResponse is a nested category tree response.
type CategoryTreeResponse struct {
	Data []CategoryResponse `json:"data"`
}

// ToCategoryResponse maps a domain category to API response.
func ToCategoryResponse(c *domain.Category) CategoryResponse {
	resp := CategoryResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		ImageURL:    c.ImageURL,
		SortOrder:   c.SortOrder,
		IsActive:      c.IsActive,
		ProductsCount: c.ProductsCount,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if c.ParentID != nil {
		pid := c.ParentID.String()
		resp.ParentID = &pid
	}
	for _, child := range c.Children {
		resp.Children = append(resp.Children, ToCategoryResponse(&child))
	}
	return resp
}

// ToCategoryListResponse maps a paginated domain list to API response.
func ToCategoryListResponse(result pagination.Paginated[domain.Category]) CategoryListResponse {
	items := make([]CategoryResponse, len(result.Data))
	for i, c := range result.Data {
		items[i] = ToCategoryResponse(&c)
	}
	return CategoryListResponse{Data: items, Meta: result.Meta}
}

// ToCategoryTreeResponse maps a category tree to API response.
func ToCategoryTreeResponse(items []domain.Category) CategoryTreeResponse {
	result := make([]CategoryResponse, len(items))
	for i, c := range items {
		result[i] = ToCategoryResponse(&c)
	}
	return CategoryTreeResponse{Data: result}
}
