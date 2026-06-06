package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProductModel is the GORM model for products table.
type ProductModel struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CategoryID       *uuid.UUID     `gorm:"type:uuid;index"`
	Name             string         `gorm:"type:varchar(300);not null"`
	Slug             string         `gorm:"type:varchar(300);uniqueIndex;not null"`
	SKU              string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	Description      *string        `gorm:"type:text"`
	ShortDescription *string        `gorm:"type:varchar(500)"`
	Price            float64        `gorm:"type:decimal(12,2);not null"`
	SalePrice        *float64       `gorm:"type:decimal(12,2)"`
	Brand            *string        `gorm:"type:varchar(100)"`
	IsFeatured       bool           `gorm:"not null;default:false"`
	Status           string         `gorm:"type:varchar(20);not null;default:draft;index"`
	CreatedAt        time.Time      `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt        time.Time      `gorm:"type:timestamptz;not null;autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	Images           []ProductImageModel      `gorm:"foreignKey:ProductID"`
	Attributes       []ProductAttributeModel  `gorm:"foreignKey:ProductID"`
	Inventory        *InventoryModel          `gorm:"foreignKey:ProductID"`
}

func (ProductModel) TableName() string { return "products" }

// ProductImageModel is the GORM model for product_images table.
type ProductImageModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index"`
	URL       string    `gorm:"type:varchar(500);not null"`
	AltText   *string   `gorm:"type:varchar(200)"`
	SortOrder int       `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (ProductImageModel) TableName() string { return "product_images" }

// ProductAttributeModel is the GORM model for product_attributes table.
type ProductAttributeModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Value     string    `gorm:"type:varchar(200);not null"`
}

func (ProductAttributeModel) TableName() string { return "product_attributes" }

// InventoryModel is the GORM model for inventories table.
type InventoryModel struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProductID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	Quantity          int       `gorm:"not null;default:0"`
	LowStockThreshold int       `gorm:"not null;default:10"`
	UpdatedAt         time.Time `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (InventoryModel) TableName() string { return "inventories" }

// CategoryModel is a minimal GORM model for category existence checks.
type CategoryModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (CategoryModel) TableName() string { return "categories" }
