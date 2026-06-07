package coupon

import "app/pkg/apperror"

var (
	ErrNotFound            = apperror.NotFound("coupon")
	ErrCodeConflict        = apperror.Conflict("coupon code already exists")
	ErrInvalidDiscount     = apperror.Validation("invalid discount value", map[string]string{
		"discount_value": "must be greater than 0",
	})
	ErrInvalidPercentage   = apperror.Validation("invalid percentage discount", map[string]string{
		"discount_value": "percentage discount must be between 0 and 100",
	})
)
