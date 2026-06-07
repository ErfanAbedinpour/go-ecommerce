package auth

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"unicode"

	"app/pkg/apperror"
)

func validatePassword(password string) error {
	if len(password) < 8 {
		return apperror.Validation("request validation failed", map[string]string{
			"password": "must be at least 8 characters",
		})
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasLetter || !hasDigit {
		return apperror.Validation("request validation failed", map[string]string{
			"password": "must contain at least one letter and one number",
		})
	}

	return nil
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func buildResetLink(appURL, resetPath, token string) string {
	base := strings.TrimRight(appURL, "/")
	path := resetPath
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path + "?token=" + token
}
