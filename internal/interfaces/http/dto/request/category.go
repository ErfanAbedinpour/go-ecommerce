package request

// CreateCategoryRequest is the request body for creating a category.
type CreateCategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=200"`
	Slug        string  `json:"slug" validate:"omitempty,max=200"`
	Description string  `json:"description" validate:"omitempty"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid"`
	ImageURL    string  `json:"image_url" validate:"omitempty,url,max=500"`
	SortOrder   int     `json:"sort_order" validate:"gte=0"`
	IsActive    bool    `json:"is_active"`
}

// UpdateCategoryRequest is the request body for updating a category.
type UpdateCategoryRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=200"`
	Slug        *string `json:"slug" validate:"omitempty,max=200"`
	Description *string `json:"description"`
	ParentID    *string `json:"parent_id" validate:"omitempty,uuid"`
	ImageURL    *string `json:"image_url" validate:"omitempty,url,max=500"`
	SortOrder   *int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive    *bool   `json:"is_active"`
}
