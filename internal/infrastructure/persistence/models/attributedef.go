package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AttributeDefinitionModel is the GORM model for product_attribute_definitions table.
type AttributeDefinitionModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string         `gorm:"type:varchar(100);not null"`
	Slug      string         `gorm:"type:varchar(100);not null"`
	SortOrder int            `gorm:"not null;default:0"`
	IsActive  bool           `gorm:"not null;default:true"`
	CreatedAt time.Time      `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"type:timestamptz;not null;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (AttributeDefinitionModel) TableName() string { return "product_attribute_definitions" }
