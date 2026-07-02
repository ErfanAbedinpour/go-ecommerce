package models

import (
	"time"

	"github.com/google/uuid"
)

// CartModel is the GORM model for carts table.
type CartModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     *uuid.UUID `gorm:"type:uuid"`
	GuestToken *string   `gorm:"type:varchar(255)"`
	UpdatedAt  time.Time `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (CartModel) TableName() string { return "carts" }

// CartItemModel is the GORM model for cart_items table.
type CartItemModel struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CartID    uuid.UUID  `gorm:"type:uuid;not null"`
	ProductID uuid.UUID  `gorm:"type:uuid;not null"`
	SkuID     *uuid.UUID `gorm:"type:uuid"`
	Quantity  int        `gorm:"type:int;not null"`
	AddedAt   time.Time  `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (CartItemModel) TableName() string { return "cart_items" }
