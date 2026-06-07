package coupon

import "app/pkg/apperror"

// DiscountType represents how a coupon discount is calculated.
type DiscountType string

const (
	DiscountTypePercentage  DiscountType = "percentage"
	DiscountTypeFixedAmount DiscountType = "fixed_amount"
)

// ParseDiscountType validates and parses a discount type string.
func ParseDiscountType(value string) (DiscountType, error) {
	switch DiscountType(value) {
	case DiscountTypePercentage, DiscountTypeFixedAmount:
		return DiscountType(value), nil
	default:
		return "", apperror.Validation("invalid discount type", map[string]string{
			"discount_type": "must be one of: percentage, fixed_amount",
		})
	}
}

// IsValid reports whether the discount type is known.
func (d DiscountType) IsValid() bool {
	return d == DiscountTypePercentage || d == DiscountTypeFixedAmount
}

// String returns the discount type as a plain string.
func (d DiscountType) String() string {
	return string(d)
}
