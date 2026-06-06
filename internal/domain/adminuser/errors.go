package adminuser

import "app/pkg/apperror"

var (
	ErrNotFound          = apperror.NotFound("admin user")
	ErrInvalidCredentials = apperror.New(apperror.CodeInvalidCreds, "invalid email or password", 401)
	ErrAccountDisabled   = apperror.New(apperror.CodeAccountDisabled, "account is disabled", 403)
	ErrInvalidToken      = apperror.New(apperror.CodeInvalidToken, "invalid or expired token", 401)
	ErrTokenRevoked      = apperror.New(apperror.CodeTokenRevoked, "token has been revoked", 401)
)
