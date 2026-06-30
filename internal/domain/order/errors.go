package order

import "app/pkg/apperror"

var (
	ErrNotFound = apperror.NotFound("order")

	ErrInvalidStatusTransition = apperror.New(
		apperror.CodeInvalidStatus,
		"invalid order status transition",
		422,
	)

	ErrCannotCancel = apperror.Unprocessable("order cannot be cancelled in its current state")

	ErrCannotRefund = apperror.Unprocessable("order cannot be refunded in its current state")

	ErrInvalidRefundAmount = apperror.Validation("invalid refund amount", map[string]string{
		"amount": "must be greater than 0 and not exceed order total",
	})

	ErrInsufficientStock = apperror.Unprocessable("insufficient product stock")

	ErrEmptyOrder = apperror.Validation("order must contain at least one item", map[string]string{
		"items": "required",
	})

	ErrInvalidSKU = apperror.Validation("invalid product sku", map[string]string{
		"sku_id": "sku does not belong to product",
	})

	ErrInvalidDateRange = apperror.Validation("invalid date range", map[string]string{
		"from": "must be before or equal to to",
	})

	ErrPaymentAlreadyPaid = apperror.Conflict("order is already paid")
)
