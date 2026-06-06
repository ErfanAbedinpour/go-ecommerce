package adminuser

import (
	"time"

	"github.com/google/uuid"
)

// AdminUser represents an admin panel user aggregate root.
type AdminUser struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Phone        string
	IsActive     bool
	LastLoginAt  *time.Time
	Roles        []Role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Role represents a named role with permissions.
type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	Permissions []Permission
}

// Permission represents an atomic access capability.
type Permission struct {
	ID          uuid.UUID
	Name        string
	Description string
}

// RefreshToken represents a stored refresh token for rotation.
type RefreshToken struct {
	ID          uuid.UUID
	AdminUserID uuid.UUID
	TokenHash   string
	FamilyID    uuid.UUID
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time
}

// IsRevoked returns true if the token has been revoked.
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsExpired returns true if the token has expired.
func (t *RefreshToken) IsExpired() bool {
	return time.Now().UTC().After(t.ExpiresAt)
}

// PermissionNames returns a flat list of permission names from all roles.
func (u *AdminUser) PermissionNames() []string {
	seen := make(map[string]struct{})
	var names []string
	for _, role := range u.Roles {
		for _, perm := range role.Permissions {
			if _, ok := seen[perm.Name]; !ok {
				seen[perm.Name] = struct{}{}
				names = append(names, perm.Name)
			}
		}
	}
	return names
}

// HasPermission checks if the user has a specific permission.
func (u *AdminUser) HasPermission(permission string) bool {
	for _, name := range u.PermissionNames() {
		if name == permission {
			return true
		}
	}
	return false
}

// FullName returns the user's full name.
func (u *AdminUser) FullName() string {
	return u.FirstName + " " + u.LastName
}
