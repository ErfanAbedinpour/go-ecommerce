package postgres

import (
	"app/internal/domain/adminuser"
	"app/internal/infrastructure/persistence/models"
)

func toAdminUserDomain(m *models.AdminUserModel) *adminuser.AdminUser {
	u := &adminuser.AdminUser{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		FirstName:    m.FirstName,
		LastName:     m.LastName,
		IsActive:     m.IsActive,
		LastLoginAt:  m.LastLoginAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
	if m.Phone != nil {
		u.Phone = *m.Phone
	}
	for _, r := range m.Roles {
		u.Roles = append(u.Roles, toRoleDomain(r))
	}
	return u
}

func toRoleDomain(m models.RoleModel) adminuser.Role {
	r := adminuser.Role{
		ID:   m.ID,
		Name: m.Name,
	}
	if m.Description != nil {
		r.Description = *m.Description
	}
	for _, p := range m.Permissions {
		r.Permissions = append(r.Permissions, toPermissionDomain(p))
	}
	return r
}

func toPermissionDomain(m models.PermissionModel) adminuser.Permission {
	p := adminuser.Permission{
		ID:   m.ID,
		Name: m.Name,
	}
	if m.Description != nil {
		p.Description = *m.Description
	}
	return p
}

func toRefreshTokenDomain(m *models.RefreshTokenModel) *adminuser.RefreshToken {
	return &adminuser.RefreshToken{
		ID:          m.ID,
		AdminUserID: m.AdminUserID,
		TokenHash:   m.TokenHash,
		FamilyID:    m.FamilyID,
		ExpiresAt:   m.ExpiresAt,
		RevokedAt:   m.RevokedAt,
		CreatedAt:   m.CreatedAt,
	}
}
