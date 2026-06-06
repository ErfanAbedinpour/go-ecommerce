package adminuser

import (
	"testing"

	"github.com/google/uuid"
)

func TestAdminUser_PermissionNames(t *testing.T) {
	user := &AdminUser{
		Roles: []Role{
			{
				Name: "admin",
				Permissions: []Permission{
					{Name: "products:read"},
					{Name: "products:write"},
				},
			},
			{
				Name: "manager",
				Permissions: []Permission{
					{Name: "products:read"},
					{Name: "orders:read"},
				},
			},
		},
	}

	perms := user.PermissionNames()
	if len(perms) != 3 {
		t.Errorf("expected 3 unique permissions, got %d: %v", len(perms), perms)
	}
}

func TestAdminUser_HasPermission(t *testing.T) {
	user := &AdminUser{
		Roles: []Role{
			{Permissions: []Permission{{Name: "products:read"}}},
		},
	}

	if !user.HasPermission("products:read") {
		t.Error("expected HasPermission to return true")
	}
	if user.HasPermission("products:delete") {
		t.Error("expected HasPermission to return false")
	}
}

func TestRefreshToken_IsRevoked(t *testing.T) {
	token := &RefreshToken{}
	if token.IsRevoked() {
		t.Error("expected non-revoked token")
	}
}

func TestAdminUser_FullName(t *testing.T) {
	user := &AdminUser{FirstName: "Admin", LastName: "User", ID: uuid.New()}
	if user.FullName() != "Admin User" {
		t.Errorf("FullName() = %q, want %q", user.FullName(), "Admin User")
	}
}
