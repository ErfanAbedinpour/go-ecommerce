package request

// CreateProductRequest is the request body for creating a product.
type CreateProductRequest struct {
	Name             string                    `json:"name" validate:"required,min=1,max=300"`
	Slug             string                    `json:"slug" validate:"omitempty,max=300"`
	SKU              string                    `json:"sku" validate:"required,min=1,max=100"`
	Description      string                    `json:"description" validate:"omitempty"`
	ShortDescription string                    `json:"short_description" validate:"omitempty,max=500"`
	Price            float64                   `json:"price" validate:"required,gte=0"`
	SalePrice        *float64                  `json:"sale_price" validate:"omitempty,gte=0"`
	CategoryID       *string                   `json:"category_id" validate:"omitempty,uuid"`
	Brand            string                    `json:"brand" validate:"omitempty,max=100"`
	IsFeatured       bool                      `json:"is_featured"`
	Status           string                    `json:"status" validate:"omitempty,oneof=draft active archived"`
	Images           []ProductImageRequest     `json:"images" validate:"omitempty,max=10,dive"`
	Attributes       []ProductAttributeRequest `json:"attributes" validate:"omitempty,dive"`
	Inventory        ProductInventoryRequest   `json:"inventory"`
}

// UpdateProductRequest is the request body for updating a product.
type UpdateProductRequest struct {
	Name             *string                    `json:"name" validate:"omitempty,min=1,max=300"`
	Slug             *string                    `json:"slug" validate:"omitempty,max=300"`
	SKU              *string                    `json:"sku" validate:"omitempty,min=1,max=100"`
	Description      *string                    `json:"description"`
	ShortDescription *string                    `json:"short_description" validate:"omitempty,max=500"`
	Price            *float64                   `json:"price" validate:"omitempty,gte=0"`
	SalePrice        *float64                   `json:"sale_price" validate:"omitempty,gte=0"`
	CategoryID       *string                    `json:"category_id" validate:"omitempty,uuid"`
	Brand            *string                    `json:"brand" validate:"omitempty,max=100"`
	IsFeatured       *bool                      `json:"is_featured"`
	Status           *string                    `json:"status" validate:"omitempty,oneof=draft active archived"`
	Images           *[]ProductImageRequest     `json:"images" validate:"omitempty,max=10,dive"`
	Attributes       *[]ProductAttributeRequest `json:"attributes" validate:"omitempty,dive"`
}

// ProductImageRequest holds image data in a product request.
type ProductImageRequest struct {
	URL       string `json:"url" validate:"required,url,max=500"`
	AltText   string `json:"alt_text" validate:"omitempty,max=200"`
	SortOrder int    `json:"sort_order" validate:"gte=0"`
}

// ProductAttributeRequest holds attribute data in a product request.
type ProductAttributeRequest struct {
	Name  string `json:"name" validate:"required,max=100"`
	Value string `json:"value" validate:"required,max=200"`
}

// ProductInventoryRequest holds inventory data in a product request.
type ProductInventoryRequest struct {
	Quantity          int `json:"quantity" validate:"gte=0"`
	LowStockThreshold int `json:"low_stock_threshold" validate:"omitempty,gte=0"`
}

// UpdateInventoryRequest is the request body for inventory adjustment.
type UpdateInventoryRequest struct {
	Quantity          int     `json:"quantity" validate:"required,gte=0"`
	LowStockThreshold *int    `json:"low_stock_threshold" validate:"omitempty,gte=0"`
	AdjustmentReason  string  `json:"adjustment_reason" validate:"omitempty,max=200"`
}
