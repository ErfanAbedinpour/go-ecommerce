package user

import "app/pkg/apperror"

var (
	ErrNotFound           = apperror.NotFound("user")
	ErrInvalidCredentials = apperror.New(apperror.CodeInvalidCreds, "invalid email or password", 401)
	ErrAccountDisabled    = apperror.New(apperror.CodeAccountDisabled, "account is disabled", 403)
	ErrInvalidToken       = apperror.New(apperror.CodeInvalidToken, "invalid or expired token", 401)
	ErrTokenRevoked       = apperror.New(apperror.CodeTokenRevoked, "token has been revoked", 401)
	ErrForbiddenRole      = apperror.Forbidden("insufficient role for this resource")
	ErrEmailTaken         = apperror.Conflict("email is already registered")
	ErrSignupDisabled     = apperror.Forbidden("signup is currently disabled")
	ErrInvalidResetToken  = apperror.New(apperror.CodeInvalidToken, "invalid or expired reset token", 400)
)
