package customer

import (
	"time"

	"github.com/google/uuid"
)

// Customer is the aggregate root for storefront buyers.
type Customer struct {
	ID          uuid.UUID
	Email       string
	FirstName   string
	LastName    string
	Phone       string
	Type        CustomerType
	TotalOrders int
	TotalSpent  float64
	LastOrderAt *time.Time
	Addresses   []Address
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// FullName returns the customer's display name.
func (c *Customer) FullName() string {
	return c.FirstName + " " + c.LastName
}

// Address is a customer shipping or billing address.
type Address struct {
	ID         uuid.UUID
	CustomerID uuid.UUID
	Type       AddressType
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
	IsDefault  bool
}
