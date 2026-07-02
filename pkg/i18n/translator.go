package i18n

import (
	"strings"

	"app/pkg/apperror"
)

// Translator localizes application errors for client responses.
type Translator struct {
	locale string
}

// NewTranslator creates a translator for the given locale.
func NewTranslator(locale string) *Translator {
	return &Translator{locale: ParseLocale(locale)}
}

// Locale returns the active locale.
func (t *Translator) Locale() string {
	return t.locale
}

// Translate returns a copy of the error with localized message and details.
func (t *Translator) Translate(err *apperror.AppError) *apperror.AppError {
	if err == nil {
		return err
	}
	out := *err
	if out.MessageKey != "" {
		if msg := catalogMessage(t.locale, out.MessageKey, out.MessageParams); msg != "" {
			out.Message = msg
		}
	}
	if len(out.Details) > 0 {
		localized := make(map[string]string, len(out.Details))
		for field, detail := range out.Details {
			localized[field] = translateDetail(t.locale, detail, out.MessageParams)
		}
		out.Details = localized
	}
	return &out
}

func translateDetail(locale, detail string, fallbackParams map[string]string) string {
	key := detail
	params := fallbackParams
	if idx := stringsIndex(detail, ":"); idx >= 0 {
		key = detail[:idx]
		param := detail[idx+1:]
		params = map[string]string{"param": param}
	}
	if msg := catalogMessage(locale, key, params); msg != "" {
		return msg
	}
	return detail
}

func stringsIndex(s, sep string) int {
	return strings.Index(s, sep)
}
