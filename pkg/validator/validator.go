package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"app/pkg/apperror"
	"app/pkg/i18n"
)

// Validator wraps go-playground/validator with application-specific helpers.
type Validator struct {
	validate *validator.Validate
}

// New creates a new Validator instance.
func New() *Validator {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	return &Validator{validate: v}
}

// Validate validates a struct and returns an AppError on failure.
func (v *Validator) Validate(s any) error {
	if err := v.validate.Struct(s); err != nil {
		return toAppError(err)
	}
	return nil
}

func toAppError(err error) *apperror.AppError {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return apperror.ValidationKeyed(i18n.KeyErrorValidation, "validation failed", nil)
	}

	details := make(map[string]string, len(validationErrors))
	for _, fe := range validationErrors {
		field := fe.Field()
		details[field] = formatFieldDetail(fe)
	}

	return apperror.ValidationKeyed(i18n.KeyErrorValidationRequest, "request validation failed", details)
}

func formatFieldDetail(fe validator.FieldError) string {
	key := formatFieldErrorKey(fe)
	switch fe.Tag() {
	case "min", "max", "gte", "gt", "oneof":
		return key + ":" + fe.Param()
	default:
		return key
	}
}

func formatFieldErrorKey(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return i18n.KeyValidationRequired
	case "email":
		return i18n.KeyValidationEmail
	case "min":
		return i18n.KeyValidationMin
	case "max":
		return i18n.KeyValidationMax
	case "gte":
		return i18n.KeyValidationGte
	case "gt":
		return i18n.KeyValidationGt
	case "uuid":
		return i18n.KeyValidationUUID
	case "oneof":
		return i18n.KeyValidationOneOf
	default:
		return i18n.KeyValidationDefault
	}
}

// FormatFieldError returns a human-readable validation message for tests and logging.
func FormatFieldError(fe validator.FieldError) string {
	key := formatFieldErrorKey(fe)
	switch key {
	case i18n.KeyValidationMin, i18n.KeyValidationMax, i18n.KeyValidationGte, i18n.KeyValidationGt, i18n.KeyValidationOneOf, i18n.KeyValidationDefault:
		return fmt.Sprintf("%s (%s)", key, fe.Param())
	default:
		return key
	}
}
