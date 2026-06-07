package request

// CreateBrandRequest is the request body for creating a brand.
type CreateBrandRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Slug        string `json:"slug" validate:"omitempty,max=100"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	IsActive    bool   `json:"is_active"`
}

// UpdateBrandRequest is the request body for updating a brand.
type UpdateBrandRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Slug        *string `json:"slug" validate:"omitempty,max=100"`
	Description *string `json:"description" validate:"omitempty,max=1000"`
	IsActive    *bool   `json:"is_active"`
}

// CreateProductAttributeRequest is the request body for creating a global attribute definition.
type CreateProductAttributeRequest struct {
	Name      string `json:"name" validate:"required,min=1,max=100"`
	Slug      string `json:"slug" validate:"omitempty,max=100"`
	SortOrder int    `json:"sort_order" validate:"gte=0"`
	IsActive  bool   `json:"is_active"`
}

// UpdateProductAttributeRequest is the request body for updating a global attribute definition.
type UpdateProductAttributeRequest struct {
	Name      *string `json:"name" validate:"omitempty,min=1,max=100"`
	Slug      *string `json:"slug" validate:"omitempty,max=100"`
	SortOrder *int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive  *bool   `json:"is_active"`
}

// CreateProductAttributeValueRequest is the request body for creating an attribute value.
type CreateProductAttributeValueRequest struct {
	AttributeID string `json:"attribute_id" validate:"required,uuid"`
	Value       string `json:"value" validate:"required,min=1,max=200"`
	SortOrder   int    `json:"sort_order" validate:"gte=0"`
	IsActive    bool   `json:"is_active"`
}

// UpdateProductAttributeValueRequest is the request body for updating an attribute value.
type UpdateProductAttributeValueRequest struct {
	Value     *string `json:"value" validate:"omitempty,min=1,max=200"`
	SortOrder *int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive  *bool   `json:"is_active"`
}
