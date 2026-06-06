package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"app/internal/domain/adminuser"
	"app/internal/infrastructure/auth"
)

// PasswordVerifier verifies passwords against hashes.
type PasswordVerifier interface {
	Verify(hash, password string) bool
}

// TokenGenerator generates JWT token pairs.
type TokenGenerator interface {
	GenerateTokenPair(input auth.TokenInput) (*auth.TokenPair, string, uuid.UUID, error)
	ValidateRefreshToken(token string) (*auth.Claims, error)
	RefreshTokenTTL() time.Duration
}

// AuthService handles authentication use cases.
type AuthService struct {
	users        adminuser.Repository
	refreshTokens adminuser.RefreshTokenRepository
	hasher       PasswordVerifier
	jwt          TokenGenerator
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	users adminuser.Repository,
	refreshTokens adminuser.RefreshTokenRepository,
	hasher PasswordVerifier,
	jwt TokenGenerator,
) *AuthService {
	return &AuthService{
		users:         users,
		refreshTokens: refreshTokens,
		hasher:        hasher,
		jwt:           jwt,
	}
}

// LoginInput holds login request data.
type LoginInput struct {
	Email    string
	Password string
}

// TokenOutput holds the authentication token response.
type TokenOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Login authenticates an admin user and returns tokens.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*TokenOutput, error) {
	user, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		if err == adminuser.ErrNotFound {
			return nil, adminuser.ErrInvalidCredentials
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, adminuser.ErrAccountDisabled
	}

	if !s.hasher.Verify(user.PasswordHash, input.Password) {
		return nil, adminuser.ErrInvalidCredentials
	}

	tokens, _, _, err := s.generateAndStoreTokens(ctx, user)
	if err != nil {
		return nil, err
	}

	_ = s.users.UpdateLastLogin(ctx, user.ID)

	return tokens, nil
}

// RefreshInput holds refresh token request data.
type RefreshInput struct {
	RefreshToken string
}

// Refresh rotates the refresh token and returns a new token pair.
func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (*TokenOutput, error) {
	claims, err := s.jwt.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, adminuser.ErrInvalidToken
	}

	tokenHash := auth.HashToken(input.RefreshToken)
	stored, err := s.refreshTokens.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if stored.IsRevoked() {
		_ = s.refreshTokens.RevokeFamily(ctx, stored.FamilyID)
		return nil, adminuser.ErrTokenRevoked
	}

	if stored.IsExpired() {
		return nil, adminuser.ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, adminuser.ErrInvalidToken
	}

	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, adminuser.ErrAccountDisabled
	}

	if err := s.refreshTokens.Revoke(ctx, stored.ID); err != nil {
		return nil, err
	}

	tokens, _, _, err := s.generateAndStoreTokensWithFamily(ctx, user, stored.FamilyID)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

// LogoutInput holds logout request data.
type LogoutInput struct {
	RefreshToken string
}

// Logout revokes the refresh token family.
func (s *AuthService) Logout(ctx context.Context, input LogoutInput) error {
	if input.RefreshToken == "" {
		return nil
	}

	tokenHash := auth.HashToken(input.RefreshToken)
	stored, err := s.refreshTokens.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil
	}

	return s.refreshTokens.RevokeFamily(ctx, stored.FamilyID)
}

// CurrentUserOutput holds the current user response.
type CurrentUserOutput struct {
	ID          string   `json:"id"`
	Email       string   `json:"email"`
	FirstName   string   `json:"first_name"`
	LastName    string   `json:"last_name"`
	Phone       string   `json:"phone,omitempty"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// GetCurrentUser returns the authenticated admin user's profile.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*CurrentUserOutput, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Name
	}

	return &CurrentUserOutput{
		ID:          user.ID.String(),
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Phone:       user.Phone,
		Roles:       roles,
		Permissions: user.PermissionNames(),
	}, nil
}

func (s *AuthService) generateAndStoreTokens(ctx context.Context, user *adminuser.AdminUser) (*TokenOutput, string, uuid.UUID, error) {
	return s.generateAndStoreTokensWithFamily(ctx, user, uuid.New())
}

func (s *AuthService) generateAndStoreTokensWithFamily(ctx context.Context, user *adminuser.AdminUser, familyID uuid.UUID) (*TokenOutput, string, uuid.UUID, error) {
	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Name
	}

	tokenInput := auth.TokenInput{
		UserID:      user.ID,
		Email:       user.Email,
		Roles:       roles,
		Permissions: user.PermissionNames(),
	}

	pair, tokenHash, returnedFamilyID, err := s.jwt.GenerateTokenPair(tokenInput)
	if err != nil {
		return nil, "", uuid.Nil, err
	}

	if familyID != uuid.Nil {
		returnedFamilyID = familyID
	}

	refreshToken := &adminuser.RefreshToken{
		ID:          uuid.New(),
		AdminUserID: user.ID,
		TokenHash:   tokenHash,
		FamilyID:    returnedFamilyID,
		ExpiresAt:   time.Now().UTC().Add(s.jwt.RefreshTokenTTL()),
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.refreshTokens.Create(ctx, refreshToken); err != nil {
		return nil, "", uuid.Nil, err
	}

	return &TokenOutput{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
		TokenType:    pair.TokenType,
	}, tokenHash, returnedFamilyID, nil
}
