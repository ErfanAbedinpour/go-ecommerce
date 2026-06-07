package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CouponModel is the GORM model for coupons table.
type CouponModel struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Code           string         `gorm:"type:varchar(50);uniqueIndex;not null"`
	DiscountType   string         `gorm:"type:varchar(20);not null"`
	DiscountValue  float64        `gorm:"type:decimal(12,2);not null"`
	MinOrderAmount float64        `gorm:"type:decimal(12,2);not null;default:0"`
	MaxUsage       *int           `gorm:"type:int"`
	UsageCount     int            `gorm:"not null;default:0"`
	ExpiresAt      *time.Time     `gorm:"type:timestamptz"`
	IsActive       bool           `gorm:"not null;default:true"`
	Note           *string        `gorm:"type:text"`
	CreatedAt      time.Time      `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"type:timestamptz;not null;autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (CouponModel) TableName() string { return "coupons" }
