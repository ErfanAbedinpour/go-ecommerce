package i18n

import (
	"net/http"
	"testing"

	"app/pkg/apperror"
)

func TestTranslate_FA_AccountExistsLoginRequired(t *testing.T) {
	tr := NewTranslator(LocaleFA)
	err := tr.Translate(domaincustomerAccountExistsError())

	if err.Message != "با این ایمیل یا شماره تماس حساب کاربری فعال وجود دارد. لطفاً ابتدا وارد شوید." {
		t.Fatalf("message = %q", err.Message)
	}
}

func TestTranslate_EN_AccountExistsLoginRequired(t *testing.T) {
	tr := NewTranslator(LocaleEN)
	err := tr.Translate(domaincustomerAccountExistsError())

	if err.Message != "An account with this email or phone already exists. Please log in first." {
		t.Fatalf("message = %q", err.Message)
	}
}

func TestTranslate_ValidationDetailWithParam(t *testing.T) {
	tr := NewTranslator(LocaleFA)
	err := tr.Translate(&apperror.AppError{
		Code:       apperror.CodeValidation,
		MessageKey: KeyErrorValidationRequest,
		Message:    "request validation failed",
		Status:     http.StatusBadRequest,
		Details: map[string]string{
			"password": KeyValidationMin + ":8",
		},
	})

	if err.Details["password"] != "باید حداقل 8 کاراکتر باشد" {
		t.Fatalf("detail = %q", err.Details["password"])
	}
}

func domaincustomerAccountExistsError() *apperror.AppError {
	return apperror.Keyed(
		apperror.CodeAccountExistsLogin,
		KeyCustomerAccountExistsLogin,
		"An account with this email or phone already exists. Please log in first.",
		http.StatusConflict,
	)
}
