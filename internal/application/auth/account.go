package auth

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	domaincustomer "app/internal/domain/customer"
	"app/internal/domain/user"
	"app/internal/infrastructure/auth"
)

// PasswordHasher hashes and verifies passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

// SignupInput holds registration request data.
type SignupInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
}

// Signup registers a new user and returns an authentication token pair.
func (s *AuthService) Signup(ctx context.Context, input SignupInput) (*TokenOutput, error) {
	if !s.cfg.SignupEnabled {
		return nil, user.ErrSignupDisabled
	}

	if err := ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, user.ErrEmailTaken
	} else if err != user.ErrNotFound {
		return nil, err
	}

	role, err := user.ParseRole(s.cfg.SignupDefaultRole)
	if err != nil {
		return nil, err
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	u := &user.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		FirstName:    strings.TrimSpace(input.FirstName),
		LastName:     strings.TrimSpace(input.LastName),
		Phone:        strings.TrimSpace(input.Phone),
		Role:         role,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}

	if role == user.RoleCustomer {
		userID := u.ID
		now := time.Now().UTC()
		customer := &domaincustomer.Customer{
			ID:        uuid.New(),
			UserID:    &userID,
			Email:     email,
			FirstName: u.FirstName,
			LastName:  u.LastName,
			Phone:     u.Phone,
			Type:      domaincustomer.TypeRegistered,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.customers.Create(ctx, customer); err != nil {
			return nil, err
		}
	}

	tokens, _, _, err := s.generateAndStoreTokens(ctx, u)
	return tokens, err
}

// ForgotPasswordInput holds a password reset request.
type ForgotPasswordInput struct {
	Email string
}

// MessageOutput is a generic success message response.
type MessageOutput struct {
	Message string `json:"message"`
}

const forgotPasswordMessage = "If an account with that email exists, a password reset link has been sent."

// ForgotPassword initiates a password reset flow for the given email.
func (s *AuthService) ForgotPassword(ctx context.Context, input ForgotPasswordInput) (*MessageOutput, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))

	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if err == user.ErrNotFound {
			return &MessageOutput{Message: forgotPasswordMessage}, nil
		}
		return nil, err
	}

	if !u.IsActive {
		return &MessageOutput{Message: forgotPasswordMessage}, nil
	}

	rawToken, err := generateResetToken()
	if err != nil {
		return nil, err
	}

	if err := s.resetTokens.InvalidateByUser(ctx, u.ID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	resetToken := &user.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    u.ID,
		TokenHash: auth.HashToken(rawToken),
		ExpiresAt: now.Add(s.cfg.ResetTokenTTL),
		CreatedAt: now,
	}

	if err := s.resetTokens.Create(ctx, resetToken); err != nil {
		return nil, err
	}

	resetLink := buildResetLink(s.cfg.AppURL, s.cfg.ResetPath, rawToken)
	if err := s.mailer.SendPasswordReset(ctx, u.Email, resetLink); err != nil {
		return nil, err
	}

	return &MessageOutput{Message: forgotPasswordMessage}, nil
}

// ResetPasswordInput holds password reset confirmation data.
type ResetPasswordInput struct {
	Token    string
	Password string
}

const resetPasswordMessage = "Password has been reset successfully. You can now sign in with your new password."

// ResetPassword validates a reset token and updates the user's password.
func (s *AuthService) ResetPassword(ctx context.Context, input ResetPasswordInput) (*MessageOutput, error) {
	if err := ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	tokenHash := auth.HashToken(strings.TrimSpace(input.Token))
	stored, err := s.resetTokens.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if !stored.IsValid() {
		return nil, user.ErrInvalidResetToken
	}

	u, err := s.users.FindByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}

	if !u.IsActive {
		return nil, user.ErrAccountDisabled
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	if err := s.users.UpdatePassword(ctx, u.ID, hash); err != nil {
		return nil, err
	}

	if err := s.resetTokens.MarkUsed(ctx, stored.ID); err != nil {
		return nil, err
	}

	_ = s.refreshTokens.RevokeAllByUser(ctx, u.ID)

	return &MessageOutput{Message: resetPasswordMessage}, nil
}
