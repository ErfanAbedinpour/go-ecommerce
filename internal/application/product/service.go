package product

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/product"
	"app/pkg/pagination"
)

// Service handles product management use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new product Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// ImageInput holds image data for create/update operations.
type ImageInput struct {
	URL       string
	AltText   string
	SortOrder int
}

// AttributeInput holds attribute data for create/update operations.
type AttributeInput struct {
	Name  string
	Value string
}

// InventoryInput holds inventory data for create/update operations.
type InventoryInput struct {
	Quantity          int
	LowStockThreshold int
}

// CreateInput holds data for creating a product.
type CreateInput struct {
	Name             string
	Slug             string
	SKU              string
	Description      string
	ShortDescription string
	Price            float64
	SalePrice        *float64
	CategoryID       *uuid.UUID
	Brand            string
	IsFeatured       bool
	Status           string
	Images           []ImageInput
	Attributes       []AttributeInput
	Inventory        InventoryInput
}

// UpdateInput holds partial update data for a product.
type UpdateInput struct {
	Name             *string
	Slug             *string
	SKU              *string
	Description      *string
	ShortDescription *string
	Price            *float64
	SalePrice        *float64
	CategoryID       *uuid.UUID
	Brand            *string
	IsFeatured       *bool
	Status           *string
	Images           *[]ImageInput
	Attributes       *[]AttributeInput
}

// InventoryUpdateInput holds inventory adjustment data.
type InventoryUpdateInput struct {
	Quantity          int
	LowStockThreshold *int
	AdjustmentReason  string
}

// Create creates a new product with images, attributes, and inventory.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Product, error) {
	statusStr := input.Status
	if statusStr == "" {
		statusStr = string(domain.StatusDraft)
	}
	status, err := domain.ParseStatus(statusStr)
	if err != nil {
		return nil, err
	}

	if err := validatePricing(input.Price, input.SalePrice); err != nil {
		return nil, err
	}

	if input.CategoryID != nil {
		exists, err := s.repo.CategoryExists(ctx, *input.CategoryID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.ErrCategoryNotFound
		}
	}

	slug := input.Slug
	if slug == "" {
		slug = domain.GenerateSlug(input.Name)
	}

	if err := s.ensureUniqueSlug(ctx, slug, uuid.Nil); err != nil {
		return nil, err
	}
	if err := s.ensureUniqueSKU(ctx, input.SKU, uuid.Nil); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	productID := uuid.New()

	product := &domain.Product{
		ID:               productID,
		CategoryID:       input.CategoryID,
		Name:             input.Name,
		Slug:             slug,
		SKU:              input.SKU,
		Description:      input.Description,
		ShortDescription: input.ShortDescription,
		Price:            input.Price,
		SalePrice:        input.SalePrice,
		Brand:            input.Brand,
		IsFeatured:       input.IsFeatured,
		Status:           status,
		Images:           toImages(productID, input.Images),
		Attributes:       toAttributes(productID, input.Attributes),
		Inventory: domain.Inventory{
			ID:                uuid.New(),
			ProductID:         productID,
			Quantity:          input.Inventory.Quantity,
			LowStockThreshold: defaultLowStockThreshold(input.Inventory.LowStockThreshold),
			UpdatedAt:         now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// Update updates an existing product.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Name != nil {
		product.Name = *input.Name
	}
	if input.Slug != nil {
		if err := s.ensureUniqueSlug(ctx, *input.Slug, id); err != nil {
			return nil, err
		}
		product.Slug = *input.Slug
	}
	if input.SKU != nil {
		if err := s.ensureUniqueSKU(ctx, *input.SKU, id); err != nil {
			return nil, err
		}
		product.SKU = *input.SKU
	}
	if input.Description != nil {
		product.Description = *input.Description
	}
	if input.ShortDescription != nil {
		product.ShortDescription = *input.ShortDescription
	}
	if input.Price != nil {
		product.Price = *input.Price
	}
	if input.SalePrice != nil {
		product.SalePrice = input.SalePrice
	}
	if input.CategoryID != nil {
		exists, err := s.repo.CategoryExists(ctx, *input.CategoryID)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.ErrCategoryNotFound
		}
		product.CategoryID = input.CategoryID
	}
	if input.Brand != nil {
		product.Brand = *input.Brand
	}
	if input.IsFeatured != nil {
		product.IsFeatured = *input.IsFeatured
	}
	if input.Status != nil {
		status, err := domain.ParseStatus(*input.Status)
		if err != nil {
			return nil, err
		}
		product.Status = status
	}
	if input.Images != nil {
		product.Images = toImages(id, *input.Images)
	}
	if input.Attributes != nil {
		product.Attributes = toAttributes(id, *input.Attributes)
	}

	if err := validatePricing(product.Price, product.SalePrice); err != nil {
		return nil, err
	}

	product.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// Delete soft-deletes a product.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}

	hasOrders, err := s.repo.ExistsInActiveOrders(ctx, id)
	if err != nil {
		return err
	}
	if hasOrders {
		return domain.ErrHasActiveOrders
	}

	return s.repo.SoftDelete(ctx, id)
}

// GetByID returns a product by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Product, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns a paginated product list.
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.Product], error) {
	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.Product]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// Search searches products by name, SKU, or description.
func (s *Service) Search(ctx context.Context, query string, page pagination.Params) (pagination.Paginated[domain.Product], error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return pagination.Paginated[domain.Product]{}, nil
	}

	items, total, err := s.repo.Search(ctx, query, page)
	if err != nil {
		return pagination.Paginated[domain.Product]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// UpdateInventory adjusts product inventory levels.
func (s *Service) UpdateInventory(ctx context.Context, id uuid.UUID, input InventoryUpdateInput) (*domain.Inventory, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	inventory := product.Inventory
	inventory.Quantity = input.Quantity
	if input.LowStockThreshold != nil {
		inventory.LowStockThreshold = *input.LowStockThreshold
	}
	inventory.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateInventory(ctx, id, inventory); err != nil {
		return nil, err
	}

	return &inventory, nil
}

func (s *Service) ensureUniqueSlug(ctx context.Context, slug string, excludeID uuid.UUID) error {
	existing, err := s.repo.FindBySlug(ctx, slug)
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return domain.ErrSlugConflict
	}
	return nil
}

func (s *Service) ensureUniqueSKU(ctx context.Context, sku string, excludeID uuid.UUID) error {
	existing, err := s.repo.FindBySKU(ctx, sku)
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return domain.ErrSKUConflict
	}
	return nil
}

func validatePricing(price float64, salePrice *float64) error {
	if salePrice != nil && *salePrice > price {
		return domain.ErrInvalidSalePrice
	}
	return nil
}

func defaultLowStockThreshold(threshold int) int {
	if threshold <= 0 {
		return 10
	}
	return threshold
}

func toImages(productID uuid.UUID, inputs []ImageInput) []domain.Image {
	images := make([]domain.Image, len(inputs))
	for i, in := range inputs {
		images[i] = domain.Image{
			ID:        uuid.New(),
			ProductID: productID,
			URL:       in.URL,
			AltText:   in.AltText,
			SortOrder: in.SortOrder,
		}
	}
	return images
}

func toAttributes(productID uuid.UUID, inputs []AttributeInput) []domain.Attribute {
	attrs := make([]domain.Attribute, len(inputs))
	for i, in := range inputs {
		attrs[i] = domain.Attribute{
			ID:        uuid.New(),
			ProductID: productID,
			Name:      in.Name,
			Value:     in.Value,
		}
	}
	return attrs
}
