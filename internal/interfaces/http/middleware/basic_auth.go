package middleware

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
)

// BasicAuth protects routes with HTTP Basic authentication.
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	userBytes := []byte(username)
	passBytes := []byte(password)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := parseBasicAuth(r)
			if !ok ||
				subtle.ConstantTimeCompare([]byte(user), userBytes) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), passBytes) != 1 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Swagger"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func parseBasicAuth(r *http.Request) (username, password string, ok bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Basic ") {
		return "", "", false
	}

	payload, err := base64.StdEncoding.DecodeString(strings.TrimSpace(header[6:]))
	if err != nil {
		return "", "", false
	}

	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
