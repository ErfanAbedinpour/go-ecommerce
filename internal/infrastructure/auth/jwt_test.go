package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"app/internal/config"
)

func TestJWTService_GenerateAndValidateAccessToken(t *testing.T) {
	svc := NewJWTService(config.JWTConfig{
		Secret:          "test-secret-key-minimum-32-characters",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	})

	userID := uuid.New()
	input := TokenInput{
		UserID:      userID,
		Email:       "admin@shop.com",
		Roles:       []string{"super_admin"},
		Permissions: []string{"products:read"},
	}

	pair, hash, familyID, err := svc.GenerateTokenPair(input)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}
	if pair.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if pair.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if hash == "" {
		t.Error("expected non-empty token hash")
	}
	if familyID == uuid.Nil {
		t.Error("expected non-nil family ID")
	}

	claims, err := svc.ValidateAccessToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}
	if claims.UserID != userID.String() {
		t.Errorf("UserID = %q, want %q", claims.UserID, userID.String())
	}
	if claims.TokenType != string(TokenTypeAccess) {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, TokenTypeAccess)
	}
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	svc := NewJWTService(config.JWTConfig{
		Secret:          "test-secret-key-minimum-32-characters",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
		Issuer:          "test",
	})

	input := TokenInput{
		UserID: uuid.New(),
		Email:  "admin@shop.com",
	}

	pair, _, _, err := svc.GenerateTokenPair(input)
	if err != nil {
		t.Fatalf("GenerateTokenPair() error = %v", err)
	}

	claims, err := svc.ValidateRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("ValidateRefreshToken() error = %v", err)
	}
	if claims.TokenType != string(TokenTypeRefresh) {
		t.Errorf("TokenType = %q, want %q", claims.TokenType, TokenTypeRefresh)
	}
}

func TestJWTService_InvalidToken(t *testing.T) {
	svc := NewJWTService(config.JWTConfig{
		Secret:         "test-secret-key-minimum-32-characters",
		AccessTokenTTL: 15 * time.Minute,
		Issuer:         "test",
	})

	_, err := svc.ValidateAccessToken("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestHashToken_Deterministic(t *testing.T) {
	token := "test-token-value"
	hash1 := HashToken(token)
	hash2 := HashToken(token)
	if hash1 != hash2 {
		t.Errorf("HashToken not deterministic: %q != %q", hash1, hash2)
	}
	if len(hash1) != 64 {
		t.Errorf("expected SHA-256 hex length 64, got %d", len(hash1))
	}
}
