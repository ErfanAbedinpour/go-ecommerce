package product

import "app/pkg/apperror"

var (
	ErrNotFound          = apperror.NotFound("product")
	ErrSlugConflict      = apperror.Conflict("product slug already exists")
	ErrSKUConflict       = apperror.Conflict("product SKU already exists")
	ErrCategoryNotFound  = apperror.NotFound("category")
	ErrHasActiveOrders   = apperror.Unprocessable("cannot delete product referenced by active orders")
	ErrInvalidSalePrice  = apperror.Validation("invalid sale price", map[string]string{
		"sale_price": "must be less than or equal to price",
	})
	ErrInvalidStockLevel = apperror.Validation("invalid stock_level filter", map[string]string{
		"stock_level": "must be one of: low, out",
	})
	ErrEmptyAttributeName      = apperror.Validation("attribute name cannot be empty", nil)
	ErrDuplicateAttributeName  = apperror.Validation("duplicate attribute name", nil)
	ErrEmptyAttributeValues    = apperror.Validation("attribute values cannot be empty", nil)
	ErrEmptyAttributeValue     = apperror.Validation("attribute value cannot be empty", nil)
	ErrDuplicateAttributeValue = apperror.Validation("duplicate attribute value", nil)
	ErrMaxVariantsExceeded     = apperror.Validation("maximum variant limit exceeded", nil)
)
