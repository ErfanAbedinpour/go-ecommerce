package order

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound = apperror.NotFound("order")

	ErrInvalidStatusTransition = apperror.Keyed(
		apperror.CodeInvalidStatus,
		i18n.KeyOrderInvalidStatusTransition,
		"invalid order status transition",
		422,
	)
	ErrCannotCancel = apperror.UnprocessableKeyed(i18n.KeyOrderCannotCancel, "order cannot be cancelled in its current state")
	ErrCannotRefund = apperror.UnprocessableKeyed(i18n.KeyOrderCannotRefund, "order cannot be refunded in its current state")
	ErrInvalidRefundAmount = apperror.ValidationKeyed(i18n.KeyOrderInvalidRefundAmount, "invalid refund amount", map[string]string{
		"amount": "must be greater than zero and not exceed order total",
	})
	ErrInsufficientStock = apperror.UnprocessableKeyed(i18n.KeyOrderInsufficientStock, "insufficient product stock")
	ErrEmptyOrder = apperror.ValidationKeyed(i18n.KeyOrderEmpty, "order must contain at least one item", map[string]string{
		"items": "at least one item is required",
	})
	ErrInvalidSKU = apperror.ValidationKeyed(i18n.KeyOrderInvalidSKU, "invalid product sku", map[string]string{
		"sku_id": "sku does not belong to the product",
	})
	ErrInvalidDateRange = apperror.ValidationKeyed(i18n.KeyOrderInvalidDateRange, "invalid date range", map[string]string{
		"date_range": "start date must be before end date",
	})
	ErrPaymentAlreadyPaid = apperror.ConflictKeyed(i18n.KeyOrderPaymentAlreadyPaid, "order is already paid")
	ErrPaymentExpired     = apperror.UnprocessableKeyed(i18n.KeyOrderPaymentExpired, "order payment window has expired")
)
