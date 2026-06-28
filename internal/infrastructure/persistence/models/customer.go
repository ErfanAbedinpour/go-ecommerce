package models

import (
	"time"

	"github.com/google/uuid"
)

// CustomerModel is the GORM model for customers table.
type CustomerModel struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      *uuid.UUID `gorm:"type:uuid;index"`
	Email       string     `gorm:"type:varchar(255);not null"`
	FirstName   string    `gorm:"type:varchar(100);not null"`
	LastName    string    `gorm:"type:varchar(100);not null"`
	Phone       *string   `gorm:"type:varchar(20)"`
	Type        string    `gorm:"type:varchar(20);not null;default:registered"`
	TotalOrders int       `gorm:"not null;default:0"`
	TotalSpent  float64   `gorm:"type:decimal(12,2);not null;default:0"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (CustomerModel) TableName() string { return "customers" }

// CustomerAddressModel is the GORM model for customer_addresses table.
type CustomerAddressModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID uuid.UUID `gorm:"type:uuid;not null;index"`
	Type       string    `gorm:"type:varchar(20);not null"`
	Street     string    `gorm:"type:varchar(300);not null"`
	City       string    `gorm:"type:varchar(100);not null"`
	State      *string   `gorm:"type:varchar(100)"`
	PostalCode string    `gorm:"type:varchar(20);not null"`
	Country    string    `gorm:"type:varchar(2);not null"`
	IsDefault  bool      `gorm:"not null;default:false"`
}

func (CustomerAddressModel) TableName() string { return "customer_addresses" }
