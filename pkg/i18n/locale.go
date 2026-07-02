package i18n

import "strings"

const (
	LocaleEN = "en"
	LocaleFA = "fa"
)

// ParseLocale normalizes a locale string; unsupported values fall back to English.
func ParseLocale(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case LocaleFA:
		return LocaleFA
	default:
		return LocaleEN
	}
}
