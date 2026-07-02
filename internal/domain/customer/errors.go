package customer

import (
	"net/http"

	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound                   = apperror.NotFound("customer")
	ErrEmailTaken                 = apperror.ConflictKeyed(i18n.KeyCustomerEmailTaken, "customer email is already in use")
	ErrHasOrders                  = apperror.UnprocessableKeyed(i18n.KeyCustomerHasOrders, "cannot delete customer with existing orders")
	ErrInvalidType                = apperror.ValidationKeyed(i18n.KeyCustomerInvalidType, "invalid customer type", map[string]string{
		"type": "must be one of: registered, guest",
	})
	ErrAccountExistsLoginRequired = apperror.Keyed(
		apperror.CodeAccountExistsLogin,
		i18n.KeyCustomerAccountExistsLogin,
		"An account with this email or phone already exists. Please log in first.",
		http.StatusConflict,
	)
)
