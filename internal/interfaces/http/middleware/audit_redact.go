package middleware

import (
	"encoding/json"
	"strings"
)

var sensitiveAuditKeys = map[string]struct{}{
	"password":         {},
	"current_password": {},
	"new_password":     {},
	"token":            {},
	"refresh_token":    {},
	"access_token":     {},
	"reset_token":      {},
	"secret":           {},
	"api_key":          {},
	"authorization":    {},
}

func redactJSON(body []byte) []byte {
	if len(body) == 0 {
		return body
	}

	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return []byte(`"[REDACTED]"`)
	}

	for key := range data {
		if _, ok := sensitiveAuditKeys[strings.ToLower(key)]; ok {
			data[key] = "[REDACTED]"
		}
	}

	out, err := json.Marshal(data)
	if err != nil {
		return []byte(`"[REDACTED]"`)
	}
	return out
}
