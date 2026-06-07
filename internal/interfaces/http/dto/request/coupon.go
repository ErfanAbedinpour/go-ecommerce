package request

import "time"

// CreateCouponRequest is the request body for creating a coupon.
type CreateCouponRequest struct {
	Code           string     `json:"code" validate:"required,min=3,max=50,alphanum"`
	DiscountType   string     `json:"discount_type" validate:"required,oneof=percentage fixed_amount"`
	DiscountValue  float64    `json:"discount_value" validate:"required,gt=0"`
	MinOrderAmount float64    `json:"min_order_amount" validate:"gte=0"`
	MaxUsage       *int       `json:"max_usage" validate:"omitempty,gt=0"`
	ExpiresAt      *time.Time `json:"expires_at"`
	Note           string     `json:"note" validate:"omitempty,max=500"`
	IsActive       bool       `json:"is_active"`
}

// UpdateCouponRequest is the request body for updating a coupon.
type UpdateCouponRequest struct {
	Code           *string    `json:"code" validate:"omitempty,min=3,max=50,alphanum"`
	DiscountType   *string    `json:"discount_type" validate:"omitempty,oneof=percentage fixed_amount"`
	DiscountValue  *float64   `json:"discount_value" validate:"omitempty,gt=0"`
	MinOrderAmount *float64   `json:"min_order_amount" validate:"omitempty,gte=0"`
	MaxUsage       *int       `json:"max_usage" validate:"omitempty,gt=0"`
	ExpiresAt      *time.Time `json:"expires_at"`
	Note           *string    `json:"note" validate:"omitempty,max=500"`
	IsActive       *bool      `json:"is_active"`
}
