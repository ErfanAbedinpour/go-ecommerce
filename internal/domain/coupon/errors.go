package coupon

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound            = apperror.NotFound("coupon")
	ErrCodeConflict        = apperror.ConflictKeyed(i18n.KeyCouponCodeConflict, "coupon code already exists")
	ErrInvalidDiscount     = apperror.ValidationKeyed(i18n.KeyCouponInvalidDiscount, "invalid discount value", map[string]string{
		"discount_value": "must be greater than zero",
	})
	ErrInvalidPercentage   = apperror.ValidationKeyed(i18n.KeyCouponInvalidPercentage, "invalid percentage discount", map[string]string{
		"discount_value": "must be between 1 and 100 for percentage coupons",
	})
	ErrNotApplicable = apperror.UnprocessableKeyed(i18n.KeyCouponNotApplicable, "coupon is not active")
	ErrExpired       = apperror.UnprocessableKeyed(i18n.KeyCouponExpired, "coupon has expired")
	ErrExhausted     = apperror.UnprocessableKeyed(i18n.KeyCouponExhausted, "coupon usage limit reached")
	ErrMinOrderNotMet = apperror.UnprocessableKeyed(i18n.KeyCouponMinOrderNotMet, "order subtotal does not meet coupon minimum")
)
