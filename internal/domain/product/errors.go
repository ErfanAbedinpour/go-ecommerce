package product

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound          = apperror.NotFound("product")
	ErrSlugConflict      = apperror.ConflictKeyed(i18n.KeyProductSlugConflict, "product slug already exists")
	ErrSKUConflict       = apperror.ConflictKeyed(i18n.KeyProductSKUConflict, "product SKU already exists")
	ErrCategoryNotFound  = apperror.NotFound("category")
	ErrHasActiveOrders   = apperror.UnprocessableKeyed(i18n.KeyProductHasActiveOrders, "cannot delete product referenced by active orders")
	ErrInvalidSalePrice  = apperror.ValidationKeyed(i18n.KeyProductInvalidSalePrice, "invalid sale price", map[string]string{
		"sale_price": "must be less than regular price",
	})
	ErrInvalidStockLevel = apperror.ValidationKeyed(i18n.KeyProductInvalidStockLevel, "invalid stock_level filter", map[string]string{
		"stock_level": "must be one of: all, in_stock, low_stock, out_of_stock",
	})
	ErrEmptyAttributeName      = apperror.ValidationKeyed(i18n.KeyProductEmptyAttributeName, "attribute name cannot be empty", nil)
	ErrDuplicateAttributeName  = apperror.ValidationKeyed(i18n.KeyProductDuplicateAttributeName, "duplicate attribute name", nil)
	ErrEmptyAttributeValues    = apperror.ValidationKeyed(i18n.KeyProductEmptyAttributeValues, "attribute values cannot be empty", nil)
	ErrEmptyAttributeValue     = apperror.ValidationKeyed(i18n.KeyProductEmptyAttributeValue, "attribute value cannot be empty", nil)
	ErrDuplicateAttributeValue = apperror.ValidationKeyed(i18n.KeyProductDuplicateAttributeValue, "duplicate attribute value", nil)
	ErrMaxVariantsExceeded     = apperror.ValidationKeyed(i18n.KeyProductMaxVariantsExceeded, "maximum variant limit exceeded", nil)
)
