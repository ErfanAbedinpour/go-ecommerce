package i18n

import "strings"

var catalogs = map[string]map[string]string{
	LocaleEN: catalogEN,
	LocaleFA: catalogFA,
}

func catalogMessage(locale, key string, params map[string]string) string {
	locale = ParseLocale(locale)
	cat, ok := catalogs[locale]
	if !ok {
		cat = catalogEN
	}
	msg, ok := cat[key]
	if !ok {
		if msg, ok = catalogEN[key]; ok {
			return applyParams(msg, params, catalogEN)
		}
		return ""
	}
	return applyParams(msg, params, cat)
}

func applyParams(template string, params map[string]string, cat map[string]string) string {
	if template == "" {
		return ""
	}
	out := template
	for k, v := range params {
		value := v
		if translated, ok := cat[v]; ok {
			value = translated
		} else if translated, ok := catalogEN[v]; ok {
			value = translated
		}
		out = strings.ReplaceAll(out, "{"+k+"}", value)
	}
	return out
}
