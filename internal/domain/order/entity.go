package order

import (
	"time"

	"github.com/google/uuid"
)

// Order is the aggregate root for customer purchases.
type Order struct {
	ID             uuid.UUID
	OrderNumber    string
	CustomerID     uuid.UUID
	CouponID       *uuid.UUID
	Status         Status
	PaymentStatus  PaymentStatus
	Subtotal       float64
	DiscountAmount float64
	ShippingAmount float64
	TaxAmount      float64
	Total           float64
	Notes           string
	PaymentMethod   string
	TransactionID      string
	PaymentExpiresAt   *time.Time
	BillingAddress  Address
	ShippingAddress Address
	Items          []Item
	Customer       *CustomerSnapshot
	Timeline       []StatusHistory
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Item is a line item snapshot on an order.
type Item struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	ProductName string
	ProductSKU  string
	Quantity    int
	UnitPrice   float64
	TotalPrice  float64
}

// CustomerSnapshot is embedded customer info on order detail.
type CustomerSnapshot struct {
	ID        uuid.UUID
	Email     string
	FirstName string
	LastName  string
	Phone     string
}

// FullName returns the customer's display name.
func (c *CustomerSnapshot) FullName() string {
	if c == nil {
		return ""
	}
	return c.FirstName + " " + c.LastName
}

// StatusHistory is a timeline entry for order status changes.
type StatusHistory struct {
	ID         uuid.UUID
	OrderID    uuid.UUID
	FromStatus *Status
	ToStatus   Status
	Note       string
	ChangedBy  *uuid.UUID
	CreatedAt  time.Time
}

// TransitionTo validates and applies a new status.
func (o *Order) TransitionTo(to Status) error {
	if !CanTransition(o.Status, to) {
		return ErrInvalidStatusTransition
	}
	if to == StatusRefunded && o.PaymentStatus != PaymentPaid {
		return ErrCannotRefund
	}
	o.Status = to
	if to == StatusRefunded {
		o.PaymentStatus = PaymentRefunded
	}
	return nil
}
