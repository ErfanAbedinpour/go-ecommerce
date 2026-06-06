package user

import (
	"testing"

	"github.com/google/uuid"
)

func TestUser_IsAdmin(t *testing.T) {
	tests := []struct {
		role Role
		want bool
	}{
		{RoleAdmin, true},
		{RoleCustomer, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.role), func(t *testing.T) {
			u := &User{Role: tt.role}
			if got := u.IsAdmin(); got != tt.want {
				t.Errorf("IsAdmin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRole(t *testing.T) {
	role, err := ParseRole("admin")
	if err != nil {
		t.Fatalf("ParseRole() error = %v", err)
	}
	if role != RoleAdmin {
		t.Errorf("role = %q, want %q", role, RoleAdmin)
	}

	_, err = ParseRole("invalid")
	if err == nil {
		t.Error("expected error for invalid role")
	}
}

func TestUser_FullName(t *testing.T) {
	u := &User{FirstName: "Jane", LastName: "Doe", ID: uuid.New()}
	if u.FullName() != "Jane Doe" {
		t.Errorf("FullName() = %q, want %q", u.FullName(), "Jane Doe")
	}
}
