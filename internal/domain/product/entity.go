package product

import (
	"time"

	"github.com/google/uuid"
)

// Product is the aggregate root for catalog products.
type Product struct {
	ID               uuid.UUID
	CategoryID       *uuid.UUID
	Name             string
	Slug             string
	Description      string
	ShortDescription string
	Price            float64
	SalePrice        *float64
	Brand            string
	IsFeatured       bool
	Status           Status
	Images           []Image
	Attributes       []ProductAttribute
	SKUs             []Sku
	Inventory        Inventory
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Image represents a product gallery image.
type Image struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	URL       string
	AltText   string
	SortOrder int
	CreatedAt time.Time
}

// ProductAttribute represents a product attribute (e.g. Color).
type ProductAttribute struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Name      string
	Values    []ProductAttributeValue
}

// ProductAttributeValue represents a specific value for a product attribute (e.g. Red).
type ProductAttributeValue struct {
	ID          uuid.UUID
	AttributeID uuid.UUID
	Value       string
}

// Sku represents a specific variant of a product.
type Sku struct {
	ID         uuid.UUID
	ProductID  uuid.UUID
	Code       string
	Attributes map[string]string
	CreatedAt  time.Time
}

// Inventory tracks stock levels for a product.
type Inventory struct {
	ID                uuid.UUID
	ProductID         uuid.UUID
	Quantity          int
	LowStockThreshold int
	UpdatedAt         time.Time
}

// IsLowStock reports whether quantity is at or below the threshold.
func (i Inventory) IsLowStock() bool {
	return i.Quantity <= i.LowStockThreshold
}

// IsOutOfStock reports whether quantity is zero.
func (i Inventory) IsOutOfStock() bool {
	return i.Quantity == 0
}

// EffectivePrice returns sale price when set, otherwise regular price.
func (p *Product) EffectivePrice() float64 {
	if p.SalePrice != nil && *p.SalePrice >= 0 {
		return *p.SalePrice
	}
	return p.Price
}
