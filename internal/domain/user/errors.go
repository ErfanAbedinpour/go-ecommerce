package user

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrNotFound           = apperror.NotFound("user")
	ErrInvalidCredentials = apperror.Keyed(apperror.CodeInvalidCreds, i18n.KeyUserInvalidCredentials, "invalid email or password", 401)
	ErrAccountDisabled    = apperror.Keyed(apperror.CodeAccountDisabled, i18n.KeyUserAccountDisabled, "account is disabled", 403)
	ErrInvalidToken       = apperror.Keyed(apperror.CodeInvalidToken, i18n.KeyUserInvalidToken, "invalid or expired token", 401)
	ErrTokenRevoked       = apperror.Keyed(apperror.CodeTokenRevoked, i18n.KeyUserTokenRevoked, "token has been revoked", 401)
	ErrForbiddenRole      = apperror.Keyed(apperror.CodeForbidden, i18n.KeyUserForbiddenRole, "insufficient role for this resource", 403)
	ErrEmailTaken         = apperror.ConflictKeyed(i18n.KeyUserEmailTaken, "email is already registered")
	ErrSignupDisabled     = apperror.Keyed(apperror.CodeForbidden, i18n.KeyUserSignupDisabled, "signup is currently disabled", 403)
	ErrInvalidResetToken  = apperror.Keyed(apperror.CodeInvalidToken, i18n.KeyUserInvalidResetToken, "invalid or expired reset token", 400)
	ErrCannotDeleteSelf   = apperror.UnprocessableKeyed(i18n.KeyUserCannotDeleteSelf, "cannot delete your own account")
	ErrLastAdmin          = apperror.UnprocessableKeyed(i18n.KeyUserLastAdmin, "cannot delete the last active admin account")
)
