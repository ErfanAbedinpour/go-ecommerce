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
	Name   string
	Values []string
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

	now := time.Now().UTC()
	productID := uuid.New()

	attributes, skus, err := generateSKUs(productID, slug, input.Attributes)
	if err != nil {
		return nil, err
	}
	
	// Ensure generated SKUs are unique across the system
	for _, sku := range skus {
		existing, err := s.repo.FindBySKU(ctx, sku.Code)
		if err != nil && err != domain.ErrNotFound {
			return nil, err
		}
		if existing != nil {
			return nil, domain.ErrSKUConflict
		}
	}

	product := &domain.Product{
		ID:               productID,
		CategoryID:       input.CategoryID,
		Name:             input.Name,
		Slug:             slug,
		Description:      input.Description,
		ShortDescription: input.ShortDescription,
		Price:            input.Price,
		SalePrice:        input.SalePrice,
		Brand:            input.Brand,
		IsFeatured:       input.IsFeatured,
		Status:           status,
		Images:           toImages(productID, input.Images),
		Attributes:       attributes,
		SKUs:             skus,
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
		attributes, skus, err := generateSKUs(id, product.Slug, *input.Attributes)
		if err != nil {
			return nil, err
		}
		
		// Ensure generated SKUs are unique across the system
		for _, sku := range skus {
			existing, err := s.repo.FindBySKU(ctx, sku.Code)
			if err != nil && err != domain.ErrNotFound {
				return nil, err
			}
			if existing != nil && existing.ID != id {
				return nil, domain.ErrSKUConflict
			}
		}
		
		product.Attributes = attributes
		product.SKUs = skus
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
	if err := validateStockLevel(filter.StockLevel); err != nil {
		return pagination.Paginated[domain.Product]{}, err
	}

	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.Product]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// GetStats returns product catalog KPI counts.
func (s *Service) GetStats(ctx context.Context) (*domain.Stats, error) {
	return s.repo.GetStats(ctx)
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

func validateStockLevel(level string) error {
	switch level {
	case "", "low", "out":
		return nil
	default:
		return domain.ErrInvalidStockLevel
	}
}

const maxVariantCount = 1000

func generateSKUs(productID uuid.UUID, slug string, inputs []AttributeInput) ([]domain.ProductAttribute, []domain.Sku, error) {
	if len(inputs) == 0 {
		return nil, nil, nil
	}

	var attributes []domain.ProductAttribute
	seenNames := make(map[string]bool)
	totalCombinations := 1

	for _, in := range inputs {
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return nil, nil, domain.ErrEmptyAttributeName
		}
		nameUpper := strings.ToUpper(name)
		if seenNames[nameUpper] {
			return nil, nil, domain.ErrDuplicateAttributeName
		}
		seenNames[nameUpper] = true

		if len(in.Values) == 0 {
			return nil, nil, domain.ErrEmptyAttributeValues
		}

		var values []domain.ProductAttributeValue
		seenValues := make(map[string]bool)
		for _, v := range in.Values {
			val := strings.TrimSpace(v)
			if val == "" {
				return nil, nil, domain.ErrEmptyAttributeValue
			}
			valUpper := strings.ToUpper(val)
			if seenValues[valUpper] {
				return nil, nil, domain.ErrDuplicateAttributeValue
			}
			seenValues[valUpper] = true

			values = append(values, domain.ProductAttributeValue{
				ID:    uuid.New(),
				Value: val,
			})
		}

		totalCombinations *= len(values)
		if totalCombinations > maxVariantCount {
			return nil, nil, domain.ErrMaxVariantsExceeded
		}

		attr := domain.ProductAttribute{
			ID:        uuid.New(),
			ProductID: productID,
			Name:      name,
			Values:    values,
		}
		for i := range attr.Values {
			attr.Values[i].AttributeID = attr.ID
		}
		attributes = append(attributes, attr)
	}

	// Generate cartesian product iteratively to avoid recursion stack issues
	var skus []domain.Sku
	
	// Initialize combinations with an empty combination
	combinations := [][]domain.ProductAttributeValue{{}}
	
	for _, attr := range attributes {
		var nextCombinations [][]domain.ProductAttributeValue
		for _, combo := range combinations {
			for _, val := range attr.Values {
				// Create a new combination by appending the current value
				newCombo := make([]domain.ProductAttributeValue, len(combo), len(combo)+1)
				copy(newCombo, combo)
				newCombo = append(newCombo, val)
				nextCombinations = append(nextCombinations, newCombo)
			}
		}
		combinations = nextCombinations
	}

	for _, combo := range combinations {
		var parts []string
		
		// Prepend slug to ensure global uniqueness
		parts = append(parts, strings.ToUpper(slug))
		
		attrMap := make(map[string]string)
		for i, val := range combo {
			attrName := attributes[i].Name
			attrMap[attrName] = val.Value
			
			// URL-safe part
			part := strings.ToUpper(val.Value)
			part = strings.ReplaceAll(part, " ", "-")
			parts = append(parts, part)
		}
		
		code := strings.Join(parts, "-")
		
		skus = append(skus, domain.Sku{
			ID:         uuid.New(),
			ProductID:  productID,
			Code:       code,
			Attributes: attrMap,
			CreatedAt:  time.Now().UTC(),
		})
	}

	return attributes, skus, nil
}
