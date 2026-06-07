package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CategoryModel is the GORM model for categories table.
type CategoryModel struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ParentID    *uuid.UUID     `gorm:"type:uuid;index"`
	Name        string         `gorm:"type:varchar(200);not null"`
	Slug        string         `gorm:"type:varchar(200);uniqueIndex;not null"`
	Description *string        `gorm:"type:text"`
	ImageURL    *string        `gorm:"type:varchar(500)"`
	SortOrder   int            `gorm:"not null;default:0"`
	IsActive    bool           `gorm:"not null;default:true"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;not null;autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (CategoryModel) TableName() string { return "categories" }
