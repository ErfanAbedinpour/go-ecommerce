package storefront

import (
	"context"

	"github.com/google/uuid"

	domainbrand "app/internal/domain/brand"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

// ProductSearchHit is a lightweight product match for autocomplete search.
type ProductSearchHit struct {
	ID           uuid.UUID `json:"id"`
	Slug         string    `json:"slug"`
	Name         string    `json:"name"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	PriceToman   int64     `json:"price_toman"`
}

// ProductSearchResult is the response for storefront product search.
type ProductSearchResult struct {
	Data []ProductSearchHit `json:"data"`
}

// ProductListData wraps a list of product cards.
type ProductListData struct {
	Data []ProductCard `json:"data"`
}

// StoreBrand is a public brand for catalog filters.
type StoreBrand struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	LogoURL string    `json:"logo_url,omitempty"`
}

// StoreBrandList is the public brand list response.
type StoreBrandList struct {
	Data []StoreBrand `json:"data"`
}

// SearchProducts returns quick search suggestions for the storefront header.
func (s *Service) SearchProducts(ctx context.Context, query string, limit int) (*ProductSearchResult, error) {
	items, err := s.products.SearchStorefront(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	hits := make([]ProductSearchHit, len(items))
	for i, item := range items {
		hits[i] = toProductSearchHit(&item)
	}
	return &ProductSearchResult{Data: hits}, nil
}

// ListRelatedProducts returns related active products for a product detail page.
func (s *Service) ListRelatedProducts(ctx context.Context, productRef string, limit int) (*ProductListData, error) {
	productID, err := s.ResolveProductID(ctx, productRef)
	if err != nil {
		return nil, err
	}

	items, err := s.products.ListRelatedStorefront(ctx, productID, limit)
	if err != nil {
		return nil, err
	}

	cards := make([]ProductCard, len(items))
	for i, item := range items {
		cards[i] = toProductCard(&item)
	}
	return &ProductListData{Data: cards}, nil
}

// ListBrands returns active brands for storefront catalog filters.
func (s *Service) ListBrands(ctx context.Context) (*StoreBrandList, error) {
	active := true
	items, _, err := s.brands.List(ctx, domainbrand.ListFilter{IsActive: &active}, pagination.Params{Page: 1, PerPage: 500})
	if err != nil {
		return nil, err
	}

	brands := make([]StoreBrand, len(items))
	for i, item := range items {
		brands[i] = StoreBrand{
			ID:   item.ID,
			Name: item.Name,
			Slug: item.Slug,
		}
	}
	return &StoreBrandList{Data: brands}, nil
}

func toProductSearchHit(p *domainproduct.Product) ProductSearchHit {
	hit := ProductSearchHit{
		ID:         p.ID,
		Slug:       p.Slug,
		Name:       p.Name,
		PriceToman: toMoneyToman(p.Price),
	}
	if p.SalePrice != nil && *p.SalePrice >= 0 && *p.SalePrice < p.Price {
		hit.PriceToman = toMoneyToman(*p.SalePrice)
	}
	if len(p.Images) > 0 {
		hit.ThumbnailURL = p.Images[0].URL
	}
	return hit
}
