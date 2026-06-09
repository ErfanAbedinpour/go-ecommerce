package response

import (
	"time"

	domain "app/internal/domain/product"
	"app/pkg/pagination"
)

// ProductResponse is the full product representation.
type ProductResponse struct {
	ID               string                    `json:"id"`
	CategoryID       *string                   `json:"category_id,omitempty"`
	Name             string                    `json:"name"`
	Slug             string                    `json:"slug"`
	Description      string                    `json:"description,omitempty"`
	ShortDescription string                    `json:"short_description,omitempty"`
	Price            float64                   `json:"price"`
	SalePrice        *float64                  `json:"sale_price,omitempty"`
	Brand            string                    `json:"brand,omitempty"`
	IsFeatured       bool                      `json:"is_featured"`
	Status           string                    `json:"status"`
	Images           []ProductImageResponse     `json:"images"`
	Attributes       []ProductAttributeResponse `json:"attributes"`
	SKUs             []SkuResponse              `json:"skus"`
	Inventory        ProductInventoryResponse   `json:"inventory"`
	CreatedAt        time.Time                 `json:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at"`
}

// ProductImageResponse is a product image in API responses.
type ProductImageResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	AltText   string `json:"alt_text,omitempty"`
	SortOrder int    `json:"sort_order"`
}

// ProductAttributeResponse is a product attribute in API responses.
type ProductAttributeResponse struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// SkuResponse is a product SKU variant in API responses.
type SkuResponse struct {
	ID         string            `json:"id"`
	Code       string            `json:"code"`
	Attributes map[string]string `json:"attributes"`
}

// ProductInventoryResponse is inventory data in API responses.
type ProductInventoryResponse struct {
	Quantity          int  `json:"quantity"`
	LowStockThreshold int  `json:"low_stock_threshold"`
	IsLowStock        bool `json:"is_low_stock"`
	IsOutOfStock      bool `json:"is_out_of_stock"`
}

// ProductListResponse is a paginated list of products.
type ProductListResponse struct {
	Data []ProductResponse  `json:"data"`
	Meta pagination.Meta    `json:"meta"`
}

// ToProductResponse maps a domain product to an API response.
func ToProductResponse(p *domain.Product) ProductResponse {
	resp := ProductResponse{
		ID:               p.ID.String(),
		Name:             p.Name,
		Slug:             p.Slug,
		Description:      p.Description,
		ShortDescription: p.ShortDescription,
		Price:            p.Price,
		SalePrice:        p.SalePrice,
		Brand:            p.Brand,
		IsFeatured:       p.IsFeatured,
		Status:           p.Status.String(),
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
		Inventory: ProductInventoryResponse{
			Quantity:          p.Inventory.Quantity,
			LowStockThreshold: p.Inventory.LowStockThreshold,
			IsLowStock:        p.Inventory.IsLowStock(),
			IsOutOfStock:      p.Inventory.IsOutOfStock(),
		},
	}
	if p.CategoryID != nil {
		cid := p.CategoryID.String()
		resp.CategoryID = &cid
	}
	for _, img := range p.Images {
		resp.Images = append(resp.Images, ProductImageResponse{
			ID:        img.ID.String(),
			URL:       img.URL,
			AltText:   img.AltText,
			SortOrder: img.SortOrder,
		})
	}
	if resp.Images == nil {
		resp.Images = []ProductImageResponse{}
	}
	for _, attr := range p.Attributes {
		var values []string
		for _, v := range attr.Values {
			values = append(values, v.Value)
		}
		resp.Attributes = append(resp.Attributes, ProductAttributeResponse{
			ID:     attr.ID.String(),
			Name:   attr.Name,
			Values: values,
		})
	}
	if resp.Attributes == nil {
		resp.Attributes = []ProductAttributeResponse{}
	}
	for _, sku := range p.SKUs {
		resp.SKUs = append(resp.SKUs, SkuResponse{
			ID:         sku.ID.String(),
			Code:       sku.Code,
			Attributes: sku.Attributes,
		})
	}
	if resp.SKUs == nil {
		resp.SKUs = []SkuResponse{}
	}
	return resp
}

// ToProductListResponse maps a paginated domain list to API response.
func ToProductListResponse(result pagination.Paginated[domain.Product]) ProductListResponse {
	items := make([]ProductResponse, len(result.Data))
	for i, p := range result.Data {
		items[i] = ToProductResponse(&p)
	}
	return ProductListResponse{Data: items, Meta: result.Meta}
}
