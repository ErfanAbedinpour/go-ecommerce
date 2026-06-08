package request

// OrderAddressRequest is a billing or shipping address on order create.
type OrderAddressRequest struct {
	Street     string `json:"street" validate:"required,max=300"`
	City       string `json:"city" validate:"required,max=100"`
	State      string `json:"state" validate:"omitempty,max=100"`
	PostalCode string `json:"postal_code" validate:"required,max=20"`
	Country    string `json:"country" validate:"required,len=2"`
}

// CreateOrderItemRequest is a line item for manual order creation.
type CreateOrderItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,gt=0"`
}

// CreateOrderRequest is the request body for manual order creation.
type CreateOrderRequest struct {
	CustomerID      string                   `json:"customer_id" validate:"required,uuid"`
	Items           []CreateOrderItemRequest `json:"items" validate:"required,min=1,dive"`
	CouponCode      string                   `json:"coupon_code" validate:"omitempty,max=50"`
	ShippingAmount  float64                  `json:"shipping_amount" validate:"omitempty,gte=0"`
	TaxAmount       float64                  `json:"tax_amount" validate:"omitempty,gte=0"`
	BillingAddress  OrderAddressRequest      `json:"billing_address" validate:"required"`
	ShippingAddress OrderAddressRequest      `json:"shipping_address" validate:"required"`
	PaymentMethod   string                   `json:"payment_method" validate:"omitempty,max=50"`
	TransactionID   string                   `json:"transaction_id" validate:"omitempty,max=100"`
	PaymentStatus   string                   `json:"payment_status" validate:"omitempty,oneof=unpaid paid"`
	Notes           string                   `json:"notes" validate:"omitempty,max=2000"`
}

// UpdateOrderNotesRequest is the request body for saving internal order notes.
type UpdateOrderNotesRequest struct {
	Notes string `json:"notes" validate:"required,max=2000"`
}

// UpdateOrderStatusRequest is the request body for updating order status.
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending processing shipped delivered cancelled refunded"`
	Note   string `json:"note" validate:"omitempty,max=500"`
}

// RefundOrderRequest is the request body for refunding an order.
type RefundOrderRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
	Reason string  `json:"reason" validate:"required,min=3,max=500"`
}
