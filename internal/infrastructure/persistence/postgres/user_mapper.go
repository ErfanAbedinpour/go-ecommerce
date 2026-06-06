package postgres

import (
	"app/internal/domain/user"
	"app/internal/infrastructure/persistence/models"
)

func toUserDomain(m *models.UserModel) *user.User {
	u := &user.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		Role:         user.Role(m.Role),
		IsActive:     m.IsActive,
		LastLoginAt:  m.LastLoginAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
	if m.Phone != nil {
		u.Phone = *m.Phone
	}
	return u
}

func toRefreshTokenDomain(m *models.RefreshTokenModel) *user.RefreshToken {
	return &user.RefreshToken{
		ID:        m.ID,
		UserID:    m.AdminUserID,
		TokenHash: m.TokenHash,
		FamilyID:  m.FamilyID,
		ExpiresAt: m.ExpiresAt,
		RevokedAt: m.RevokedAt,
		CreatedAt: m.CreatedAt,
	}
}
