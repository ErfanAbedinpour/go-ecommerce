package models

import (
	"time"

	"github.com/google/uuid"
)

// OrderModel is the GORM model for orders table.
type OrderModel struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderNumber    string     `gorm:"type:varchar(20);uniqueIndex;not null"`
	CustomerID     uuid.UUID  `gorm:"type:uuid;not null;index"`
	CouponID       *uuid.UUID `gorm:"type:uuid"`
	Status         string     `gorm:"type:varchar(20);not null;default:pending"`
	PaymentStatus  string     `gorm:"type:varchar(20);not null;default:unpaid"`
	Subtotal       float64    `gorm:"type:decimal(12,2);not null"`
	DiscountAmount float64    `gorm:"type:decimal(12,2);not null;default:0"`
	ShippingAmount float64    `gorm:"type:decimal(12,2);not null;default:0"`
	TaxAmount      float64    `gorm:"type:decimal(12,2);not null;default:0"`
	Total          float64    `gorm:"type:decimal(12,2);not null"`
	Notes          *string    `gorm:"type:text"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt      time.Time  `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (OrderModel) TableName() string { return "orders" }
