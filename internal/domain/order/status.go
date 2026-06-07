package order

import "app/pkg/apperror"

// Status represents the fulfillment lifecycle of an order.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusShipped    Status = "shipped"
	StatusDelivered  Status = "delivered"
	StatusCancelled  Status = "cancelled"
	StatusRefunded   Status = "refunded"
)

// PaymentStatus represents the payment state of an order.
type PaymentStatus string

const (
	PaymentUnpaid   PaymentStatus = "unpaid"
	PaymentPaid     PaymentStatus = "paid"
	PaymentRefunded PaymentStatus = "refunded"
)

// ParseStatus validates and parses an order status string.
func ParseStatus(value string) (Status, error) {
	switch Status(value) {
	case StatusPending, StatusProcessing, StatusShipped, StatusDelivered, StatusCancelled, StatusRefunded:
		return Status(value), nil
	default:
		return "", apperror.Validation("invalid order status", map[string]string{
			"status": "must be one of: pending, processing, shipped, delivered, cancelled, refunded",
		})
	}
}

// ParsePaymentStatus validates and parses a payment status string.
func ParsePaymentStatus(value string) (PaymentStatus, error) {
	switch PaymentStatus(value) {
	case PaymentUnpaid, PaymentPaid, PaymentRefunded:
		return PaymentStatus(value), nil
	default:
		return "", apperror.Validation("invalid payment status", map[string]string{
			"payment_status": "must be one of: unpaid, paid, refunded",
		})
	}
}

// String returns the status as a plain string.
func (s Status) String() string { return string(s) }

// String returns the payment status as a plain string.
func (p PaymentStatus) String() string { return string(p) }

// IsTerminal reports whether the order is in a terminal state.
func (s Status) IsTerminal() bool {
	return s == StatusCancelled || s == StatusRefunded || s == StatusDelivered
}

// CanTransition reports whether a status change is allowed.
func CanTransition(from, to Status) bool {
	if from == to {
		return false
	}
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// CanCancel reports whether an order can be cancelled.
func CanCancel(status Status) bool {
	return status == StatusPending || status == StatusProcessing
}

// CanRefund reports whether an order can be refunded.
func CanRefund(status Status, payment PaymentStatus) bool {
	return status == StatusDelivered && payment == PaymentPaid
}

var validTransitions = map[Status][]Status{
	StatusPending:    {StatusProcessing, StatusCancelled},
	StatusProcessing: {StatusShipped, StatusCancelled},
	StatusShipped:    {StatusDelivered},
	StatusDelivered:  {StatusRefunded},
}
