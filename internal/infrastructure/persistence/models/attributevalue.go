package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AttributeValueModel is the GORM model for product_attribute_values table.
type AttributeValueModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	AttributeID uuid.UUID      `gorm:"type:uuid;not null;index"`
	Value       string         `gorm:"type:varchar(200);not null"`
	SortOrder   int            `gorm:"not null;default:0"`
	IsActive    bool           `gorm:"not null;default:true"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;not null;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (AttributeValueModel) TableName() string { return "product_attribute_values" }
