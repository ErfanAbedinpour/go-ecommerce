package response

import (
	"time"

	domain "app/internal/domain/coupon"
	"app/pkg/pagination"
)

// CouponResponse is the coupon representation in API responses.
type CouponResponse struct {
	ID             string     `json:"id"`
	Code           string     `json:"code"`
	DiscountType   string     `json:"discount_type"`
	DiscountValue  float64    `json:"discount_value"`
	MinOrderAmount float64    `json:"min_order_amount"`
	MaxUsage       *int       `json:"max_usage,omitempty"`
	UsageCount     int        `json:"usage_count"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	IsActive       bool       `json:"is_active"`
	IsExpired      bool       `json:"is_expired"`
	IsExhausted    bool       `json:"is_exhausted"`
	Note           string     `json:"note,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// CouponListResponse is a paginated list of coupons.
type CouponListResponse struct {
	Data []CouponResponse `json:"data"`
	Meta pagination.Meta  `json:"meta"`
}

// CouponActiveResponse is the activate/deactivate response.
type CouponActiveResponse struct {
	IsActive bool `json:"is_active"`
}

// ToCouponResponse maps a domain coupon to API response.
func ToCouponResponse(c *domain.Coupon) CouponResponse {
	return CouponResponse{
		ID:             c.ID.String(),
		Code:           c.Code,
		DiscountType:   c.DiscountType.String(),
		DiscountValue:  c.DiscountValue,
		MinOrderAmount: c.MinOrderAmount,
		MaxUsage:       c.MaxUsage,
		UsageCount:     c.UsageCount,
		ExpiresAt:      c.ExpiresAt,
		IsActive:       c.IsActive,
		IsExpired:      c.IsExpired(),
		IsExhausted:    c.IsExhausted(),
		Note:           c.Note,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
}

// ToCouponListResponse maps a paginated domain list to API response.
func ToCouponListResponse(result pagination.Paginated[domain.Coupon]) CouponListResponse {
	items := make([]CouponResponse, len(result.Data))
	for i, c := range result.Data {
		items[i] = ToCouponResponse(&c)
	}
	return CouponListResponse{Data: items, Meta: result.Meta}
}
