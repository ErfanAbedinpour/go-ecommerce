package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"app/pkg/apperror"
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
		return apperror.Validation("validation failed", nil)
	}

	details := make(map[string]string, len(validationErrors))
	for _, fe := range validationErrors {
		field := fe.Field()
		details[field] = formatFieldError(fe)
	}

	return apperror.Validation("request validation failed", details)
}

func formatFieldError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("must be at most %s characters", fe.Param())
	case "gte":
		return fmt.Sprintf("must be greater than or equal to %s", fe.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fe.Param())
	case "uuid":
		return "must be a valid UUID"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("failed on '%s' validation", fe.Tag())
	}
}
