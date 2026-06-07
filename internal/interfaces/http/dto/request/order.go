package request

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
