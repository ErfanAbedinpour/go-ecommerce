package user

import "app/pkg/apperror"

// Role represents the application-level access role assigned to a user.
type Role string

const (
	RoleAdmin    Role = "admin"
	RoleCustomer Role = "customer"
)

// AllRoles returns every valid role value.
func AllRoles() []Role {
	return []Role{RoleAdmin, RoleCustomer}
}

// ParseRole validates and parses a role string.
func ParseRole(value string) (Role, error) {
	switch Role(value) {
	case RoleAdmin, RoleCustomer:
		return Role(value), nil
	default:
		return "", apperror.Validation("invalid role", map[string]string{
			"role": "must be one of: admin, customer",
		})
	}
}

// IsValid reports whether the role is a known value.
func (r Role) IsValid() bool {
	return r == RoleAdmin || r == RoleCustomer
}

// String returns the role as a plain string.
func (r Role) String() string {
	return string(r)
}
