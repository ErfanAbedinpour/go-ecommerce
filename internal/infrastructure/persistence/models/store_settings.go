package models

import (
	"time"

	"github.com/google/uuid"
)

// StoreSettingsModel is the GORM model for the store_settings singleton row.
type StoreSettingsModel struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey"`
	Site                   []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	Contact                []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	Social                 []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	SEO                    []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	About                  []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	Checkout               []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	Navigation             []byte    `gorm:"type:jsonb;not null;default:'[]'"`
	StorefrontNavigation   []byte    `gorm:"type:jsonb;not null;default:'[]'"`
	ContactSectionImageURL *string   `gorm:"type:varchar(500)"`
	CreatedAt              time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt              time.Time `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (StoreSettingsModel) TableName() string { return "store_settings" }
