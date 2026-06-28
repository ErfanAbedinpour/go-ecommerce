package models

import (
	"time"

	"github.com/google/uuid"
)

// StoreThemeModel is the GORM model for store_themes table.
type StoreThemeModel struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name            string    `gorm:"type:varchar(255);not null"`
	Slug            string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Description     *string   `gorm:"type:text"`
	PreviewImageURL *string   `gorm:"type:varchar(500)"`
	Price           float64   `gorm:"type:decimal(10,2);not null;default:0"`
	IsActive        bool      `gorm:"not null;default:true"`
	DefaultColors   []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	DefaultFont     *string   `gorm:"type:varchar(100)"`
	CreatedAt       time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (StoreThemeModel) TableName() string { return "store_themes" }

// ThemePurchaseModel is the GORM model for theme_purchases table.
type ThemePurchaseModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ThemeID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_theme_purchases_theme_user,priority:1"`
	PurchasedBy uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_theme_purchases_theme_user,priority:2"`
	PurchasedAt time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (ThemePurchaseModel) TableName() string { return "theme_purchases" }

// StoreStyleModel is the GORM model for store_style table.
type StoreStyleModel struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ActiveThemeID *uuid.UUID `gorm:"type:uuid"`
	Colors        []byte     `gorm:"type:jsonb;not null;default:'{}'"`
	FontFamily    *string    `gorm:"type:varchar(100)"`
	UpdatedAt     time.Time  `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (StoreStyleModel) TableName() string { return "store_style" }
