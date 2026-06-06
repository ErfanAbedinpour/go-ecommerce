package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"app/internal/config"
	"app/internal/domain/user"
)

// TokenType distinguishes access and refresh tokens.
type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

// Claims holds JWT claims for authenticated users.
type Claims struct {
	jwt.RegisteredClaims
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
}

// TokenPair holds an access/refresh token pair.
type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	TokenType    string
}

// JWTService handles JWT token generation and validation.
type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	issuer          string
}

// NewJWTService creates a new JWTService.
func NewJWTService(cfg config.JWTConfig) *JWTService {
	return &JWTService{
		secret:          []byte(cfg.Secret),
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		issuer:          cfg.Issuer,
	}
}

// TokenInput holds data needed to generate tokens.
type TokenInput struct {
	UserID uuid.UUID
	Email  string
	Role   user.Role
}

// GenerateTokenPair creates a new access and refresh token pair.
func (s *JWTService) GenerateTokenPair(input TokenInput) (*TokenPair, string, uuid.UUID, error) {
	familyID := uuid.New()
	jti := uuid.New().String()

	accessToken, err := s.generateToken(input, TokenTypeAccess, s.accessTokenTTL, "")
	if err != nil {
		return nil, "", uuid.Nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.generateToken(input, TokenTypeRefresh, s.refreshTokenTTL, jti)
	if err != nil {
		return nil, "", uuid.Nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenTTL.Seconds()),
		TokenType:    "Bearer",
	}, HashToken(refreshToken), familyID, nil
}

// ValidateAccessToken parses and validates an access token.
func (s *JWTService) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != string(TokenTypeAccess) {
		return nil, fmt.Errorf("invalid token type")
	}
	return claims, nil
}

// ValidateRefreshToken parses and validates a refresh token.
func (s *JWTService) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenType != string(TokenTypeRefresh) {
		return nil, fmt.Errorf("invalid token type")
	}
	return claims, nil
}

// RefreshTokenTTL returns the refresh token TTL duration.
func (s *JWTService) RefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}

func (s *JWTService) generateToken(input TokenInput, tokenType TokenType, ttl time.Duration, jti string) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   input.UserID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID:    input.UserID.String(),
		Email:     input.Email,
		Role:      input.Role.String(),
		TokenType: string(tokenType),
	}
	if jti != "" {
		claims.RegisteredClaims.ID = jti
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return claims, nil
}

// HashToken creates a SHA-256 hash of a token for secure storage.
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
