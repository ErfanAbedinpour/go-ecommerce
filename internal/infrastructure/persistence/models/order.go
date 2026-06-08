package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OrderModel is the GORM model for orders table.
type OrderModel struct {
	ID              uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderNumber     string          `gorm:"type:varchar(20);uniqueIndex;not null"`
	CustomerID      uuid.UUID       `gorm:"type:uuid;not null;index"`
	CouponID        *uuid.UUID      `gorm:"type:uuid"`
	Status          string          `gorm:"type:varchar(20);not null;default:pending"`
	PaymentStatus   string          `gorm:"type:varchar(20);not null;default:unpaid"`
	Subtotal        float64         `gorm:"type:decimal(12,2);not null"`
	DiscountAmount  float64         `gorm:"type:decimal(12,2);not null;default:0"`
	ShippingAmount  float64         `gorm:"type:decimal(12,2);not null;default:0"`
	TaxAmount       float64         `gorm:"type:decimal(12,2);not null;default:0"`
	Total           float64         `gorm:"type:decimal(12,2);not null"`
	Notes           *string         `gorm:"type:text"`
	PaymentMethod   *string         `gorm:"type:varchar(50)"`
	TransactionID   *string         `gorm:"type:varchar(100)"`
	BillingAddress  json.RawMessage `gorm:"type:jsonb;not null"`
	ShippingAddress json.RawMessage `gorm:"type:jsonb;not null"`
	CreatedAt       time.Time       `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt       time.Time       `gorm:"type:timestamptz;not null;autoUpdateTime"`
	Items           []OrderItemModel       `gorm:"foreignKey:OrderID"`
	StatusHistory   []OrderStatusHistoryModel `gorm:"foreignKey:OrderID"`
}

func (OrderModel) TableName() string { return "orders" }

// OrderItemModel is the GORM model for order_items table.
type OrderItemModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID     uuid.UUID `gorm:"type:uuid;not null;index"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null"`
	ProductName string    `gorm:"type:varchar(300);not null"`
	ProductSKU  string    `gorm:"type:varchar(100);not null"`
	Quantity    int       `gorm:"not null"`
	UnitPrice   float64   `gorm:"type:decimal(12,2);not null"`
	TotalPrice  float64   `gorm:"type:decimal(12,2);not null"`
}

func (OrderItemModel) TableName() string { return "order_items" }

// OrderStatusHistoryModel is the GORM model for order_status_history table.
type OrderStatusHistoryModel struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	FromStatus *string    `gorm:"type:varchar(20)"`
	ToStatus   string     `gorm:"type:varchar(20);not null"`
	Note       *string    `gorm:"type:text"`
	ChangedBy  *uuid.UUID `gorm:"type:uuid"`
	CreatedAt  time.Time  `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (OrderStatusHistoryModel) TableName() string { return "order_status_history" }
