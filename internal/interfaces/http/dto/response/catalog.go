package response

import (
	"time"

	domainattr "app/internal/domain/attributedef"
	domainval "app/internal/domain/attributevalue"
	domainbrand "app/internal/domain/brand"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

// BrandResponse is a brand in API responses.
type BrandResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BrandListResponse is a paginated brand list.
type BrandListResponse struct {
	Data []BrandResponse `json:"data"`
	Meta pagination.Meta `json:"meta"`
}

// CatalogAttributeResponse is a global attribute definition.
type CatalogAttributeResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	SortOrder int       `json:"sort_order"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CatalogAttributeListResponse is a paginated attribute definition list.
type CatalogAttributeListResponse struct {
	Data []CatalogAttributeResponse `json:"data"`
	Meta pagination.Meta          `json:"meta"`
}

// CatalogAttributeValueResponse is a global attribute value.
type CatalogAttributeValueResponse struct {
	ID          string    `json:"id"`
	AttributeID string    `json:"attribute_id"`
	Value       string    `json:"value"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CatalogAttributeValueListResponse is a paginated attribute value list.
type CatalogAttributeValueListResponse struct {
	Data []CatalogAttributeValueResponse `json:"data"`
	Meta pagination.Meta               `json:"meta"`
}

// ProductStatsResponse holds product catalog KPI counts.
type ProductStatsResponse struct {
	Total      int64 `json:"total"`
	Active     int64 `json:"active"`
	Draft      int64 `json:"draft"`
	OutOfStock int64 `json:"out_of_stock"`
}

// UploadResponse is the response for a successful file upload.
type UploadResponse struct {
	URL         string `json:"url"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
}

func ToBrandResponse(b *domainbrand.Brand) BrandResponse {
	return BrandResponse{
		ID:          b.ID.String(),
		Name:        b.Name,
		Slug:        b.Slug,
		Description: b.Description,
		IsActive:    b.IsActive,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func ToBrandListResponse(result pagination.Paginated[domainbrand.Brand]) BrandListResponse {
	items := make([]BrandResponse, len(result.Data))
	for i, b := range result.Data {
		items[i] = ToBrandResponse(&b)
	}
	return BrandListResponse{Data: items, Meta: result.Meta}
}

func ToCatalogAttributeResponse(d *domainattr.Definition) CatalogAttributeResponse {
	return CatalogAttributeResponse{
		ID:        d.ID.String(),
		Name:      d.Name,
		Slug:      d.Slug,
		SortOrder: d.SortOrder,
		IsActive:  d.IsActive,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func ToCatalogAttributeListResponse(result pagination.Paginated[domainattr.Definition]) CatalogAttributeListResponse {
	items := make([]CatalogAttributeResponse, len(result.Data))
	for i, d := range result.Data {
		items[i] = ToCatalogAttributeResponse(&d)
	}
	return CatalogAttributeListResponse{Data: items, Meta: result.Meta}
}

func ToCatalogAttributeValueResponse(v *domainval.Value) CatalogAttributeValueResponse {
	return CatalogAttributeValueResponse{
		ID:          v.ID.String(),
		AttributeID: v.AttributeID.String(),
		Value:       v.Value,
		SortOrder:   v.SortOrder,
		IsActive:    v.IsActive,
		CreatedAt:   v.CreatedAt,
		UpdatedAt:   v.UpdatedAt,
	}
}

func ToCatalogAttributeValueListResponse(result pagination.Paginated[domainval.Value]) CatalogAttributeValueListResponse {
	items := make([]CatalogAttributeValueResponse, len(result.Data))
	for i, v := range result.Data {
		items[i] = ToCatalogAttributeValueResponse(&v)
	}
	return CatalogAttributeValueListResponse{Data: items, Meta: result.Meta}
}

func ToProductStatsResponse(s *domainproduct.Stats) ProductStatsResponse {
	return ProductStatsResponse{
		Total:      s.Total,
		Active:     s.Active,
		Draft:      s.Draft,
		OutOfStock: s.OutOfStock,
	}
}
