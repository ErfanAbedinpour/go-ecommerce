package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BrandModel is the GORM model for brands table.
type BrandModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `gorm:"type:varchar(100);not null"`
	Slug        string         `gorm:"type:varchar(100);not null"`
	Description *string        `gorm:"type:text"`
	IsActive    bool           `gorm:"not null;default:true"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;not null;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (BrandModel) TableName() string { return "brands" }
