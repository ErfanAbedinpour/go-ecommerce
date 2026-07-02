package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	"app/internal/config"
	domaincustomer "app/internal/domain/customer"
	"app/internal/domain/user"
	"app/internal/infrastructure/auth"
	"app/internal/infrastructure/email"
)

// TokenGenerator generates JWT token pairs.
type TokenGenerator interface {
	GenerateTokenPair(input auth.TokenInput) (*auth.TokenPair, string, uuid.UUID, error)
	ValidateRefreshToken(token string) (*auth.Claims, error)
	RefreshTokenTTL() time.Duration
}

// AuthService handles authentication use cases.
type AuthService struct {
	users         user.Repository
	customers     domaincustomer.Repository
	refreshTokens user.RefreshTokenRepository
	resetTokens   user.PasswordResetRepository
	hasher        PasswordHasher
	jwt           TokenGenerator
	mailer        email.Sender
	cfg           config.AuthConfig
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	users user.Repository,
	customers domaincustomer.Repository,
	refreshTokens user.RefreshTokenRepository,
	resetTokens user.PasswordResetRepository,
	hasher PasswordHasher,
	jwt TokenGenerator,
	mailer email.Sender,
	cfg config.AuthConfig,
) *AuthService {
	return &AuthService{
		users:         users,
		customers:     customers,
		refreshTokens: refreshTokens,
		resetTokens:   resetTokens,
		hasher:        hasher,
		jwt:           jwt,
		mailer:        mailer,
		cfg:           cfg,
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

// Login authenticates a user and returns tokens with the user id for session side effects.
func (s *AuthService) Login(ctx context.Context, input LoginInput) (*SessionOutput, error) {
	u, err := s.users.FindByEmail(ctx, input.Email)
	if err != nil {
		if err == user.ErrNotFound {
			return nil, user.ErrInvalidCredentials
		}
		return nil, err
	}

	if !u.IsActive {
		return nil, user.ErrAccountDisabled
	}

	if !s.hasher.Verify(u.PasswordHash, input.Password) {
		return nil, user.ErrInvalidCredentials
	}

	tokens, _, _, err := s.generateAndStoreTokens(ctx, u)
	if err != nil {
		return nil, err
	}

	_ = s.users.UpdateLastLogin(ctx, u.ID)

	if err := s.linkOrCreateCustomer(ctx, u); err != nil {
		return nil, err
	}

	return &SessionOutput{TokenOutput: tokens, UserID: u.ID}, nil
}

// RefreshInput holds refresh token request data.
type RefreshInput struct {
	RefreshToken string
}

// Refresh rotates the refresh token and returns a new token pair.
func (s *AuthService) Refresh(ctx context.Context, input RefreshInput) (*TokenOutput, error) {
	claims, err := s.jwt.ValidateRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, user.ErrInvalidToken
	}

	tokenHash := auth.HashToken(input.RefreshToken)
	stored, err := s.refreshTokens.FindByHash(ctx, tokenHash)
	if err != nil {
		return nil, err
	}

	if stored.IsRevoked() {
		_ = s.refreshTokens.RevokeFamily(ctx, stored.FamilyID)
		return nil, user.ErrTokenRevoked
	}

	if stored.IsExpired() {
		return nil, user.ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, user.ErrInvalidToken
	}

	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !u.IsActive {
		return nil, user.ErrAccountDisabled
	}

	if err := s.refreshTokens.Revoke(ctx, stored.ID); err != nil {
		return nil, err
	}

	tokens, _, _, err := s.generateAndStoreTokensWithFamily(ctx, u, stored.FamilyID)
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
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone,omitempty"`
	Role      string `json:"role"`
}

// GetCurrentUser returns the authenticated user's profile.
func (s *AuthService) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*CurrentUserOutput, error) {
	u, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &CurrentUserOutput{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Phone:     u.Phone,
		Role:      u.Role.String(),
	}, nil
}

func (s *AuthService) generateAndStoreTokens(ctx context.Context, u *user.User) (*TokenOutput, string, uuid.UUID, error) {
	return s.generateAndStoreTokensWithFamily(ctx, u, uuid.New())
}

func (s *AuthService) generateAndStoreTokensWithFamily(ctx context.Context, u *user.User, familyID uuid.UUID) (*TokenOutput, string, uuid.UUID, error) {
	tokenInput := auth.TokenInput{
		UserID: u.ID,
		Email:  u.Email,
		Role:   u.Role,
	}

	pair, tokenHash, returnedFamilyID, err := s.jwt.GenerateTokenPair(tokenInput)
	if err != nil {
		return nil, "", uuid.Nil, err
	}

	if familyID != uuid.Nil {
		returnedFamilyID = familyID
	}

	refreshToken := &user.RefreshToken{
		ID:        uuid.New(),
		UserID:    u.ID,
		TokenHash: tokenHash,
		FamilyID:  returnedFamilyID,
		ExpiresAt: time.Now().UTC().Add(s.jwt.RefreshTokenTTL()),
		CreatedAt: time.Now().UTC(),
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
