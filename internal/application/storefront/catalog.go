package storefront

import (
	"context"
	"math"

	"github.com/google/uuid"

	domaincategory "app/internal/domain/category"
	domainproduct "app/internal/domain/product"
	"app/pkg/pagination"
)

// ProductCard is the public storefront product summary DTO.
type ProductCard struct {
	ID               uuid.UUID `json:"id"`
	Slug             string    `json:"slug"`
	Name             string    `json:"name"`
	ThumbnailURL     string    `json:"thumbnail_url"`
	PriceToman       int64     `json:"price_toman"`
	SalePriceToman   *int64    `json:"sale_price_toman,omitempty"`
	DiscountPercent  int       `json:"discount_percent,omitempty"`
	IsOnSale         bool      `json:"is_on_sale"`
	IsOutOfStock     bool      `json:"is_out_of_stock"`
	Brand            string    `json:"brand,omitempty"`
}

// ProductDetail is the public storefront product detail DTO.
type ProductDetail struct {
	ProductCard
	Description      string                          `json:"description"`
	ShortDescription string                          `json:"short_description,omitempty"`
	Images           []domainproduct.Image           `json:"images"`
	Attributes       []domainproduct.ProductAttribute `json:"attributes"`
	SKUs             []domainproduct.Sku             `json:"skus"`
}

// ListProducts returns a paginated list of active storefront products.
func (s *Service) ListProducts(ctx context.Context, filter domainproduct.StoreListFilter, page pagination.Params) (pagination.Paginated[ProductCard], error) {
	items, total, err := s.products.ListStorefront(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[ProductCard]{}, err
	}

	cards := make([]ProductCard, len(items))
	for i, p := range items {
		cards[i] = toProductCard(&p)
	}
	return pagination.NewPaginated(cards, page.Page, page.PerPage, total), nil
}

// GetProduct returns a product by slug or UUID.
func (s *Service) GetProduct(ctx context.Context, slugOrID string) (*ProductDetail, error) {
	var (
		product *domainproduct.Product
		err     error
	)

	if id, parseErr := uuid.Parse(slugOrID); parseErr == nil {
		product, err = s.products.FindByID(ctx, id)
	} else {
		product, err = s.products.FindBySlug(ctx, slugOrID)
	}
	if err != nil {
		return nil, err
	}
	if product.Status != domainproduct.StatusActive {
		return nil, domainproduct.ErrNotFound
	}

	card := toProductCard(product)
	return &ProductDetail{
		ProductCard:      card,
		Description:      product.Description,
		ShortDescription: product.ShortDescription,
		Images:           product.Images,
		Attributes:       product.Attributes,
		SKUs:             product.SKUs,
	}, nil
}

// ListCategories returns the active category tree for the storefront.
func (s *Service) ListCategories(ctx context.Context) ([]domaincategory.Category, error) {
	active := true
	items, err := s.categories.ListAll(ctx, domaincategory.ListFilter{IsActive: &active})
	if err != nil {
		return nil, err
	}

	counts, err := s.categories.ProductCounts(ctx)
	if err != nil {
		return nil, err
	}
	for i := range items {
		items[i].ProductsCount = counts[items[i].ID]
	}
	return buildCategoryTree(items), nil
}

func toProductCard(p *domainproduct.Product) ProductCard {
	card := ProductCard{
		ID:           p.ID,
		Slug:         p.Slug,
		Name:         p.Name,
		PriceToman:   toMoneyToman(p.Price),
		IsOutOfStock: p.Inventory.IsOutOfStock(),
		Brand:        p.Brand,
	}
	if len(p.Images) > 0 {
		card.ThumbnailURL = p.Images[0].URL
	}
	if p.SalePrice != nil && *p.SalePrice >= 0 && *p.SalePrice < p.Price {
		sale := toMoneyToman(*p.SalePrice)
		card.SalePriceToman = &sale
		card.IsOnSale = true
		if p.Price > 0 {
			card.DiscountPercent = int(math.Round((1 - *p.SalePrice/p.Price) * 100))
		}
	}
	return card
}

func toMoneyToman(value float64) int64 {
	return int64(math.Round(value))
}

func buildCategoryTree(items []domaincategory.Category) []domaincategory.Category {
	byID := make(map[uuid.UUID]*domaincategory.Category, len(items))
	roots := make([]domaincategory.Category, 0)

	for i := range items {
		item := items[i]
		item.Children = nil
		byID[item.ID] = &item
	}

	for _, item := range byID {
		if item.ParentID == nil {
			roots = append(roots, *item)
			continue
		}
		if parent, ok := byID[*item.ParentID]; ok {
			parent.Children = append(parent.Children, *item)
		} else {
			roots = append(roots, *item)
		}
	}

	result := make([]domaincategory.Category, 0, len(roots))
	for _, r := range roots {
		if built, ok := byID[r.ID]; ok {
			result = append(result, *built)
		}
	}
	return result
}
