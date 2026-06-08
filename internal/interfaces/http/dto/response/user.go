package response

import (
	"time"

	domain "app/internal/domain/user"
	"app/pkg/pagination"
)

// AdminUserResponse is an admin user in API responses (no password).
type AdminUserResponse struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	FullName    string     `json:"full_name"`
	Phone       string     `json:"phone,omitempty"`
	Role        string     `json:"role"`
	IsActive    bool       `json:"is_active"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// AdminUserListResponse is a paginated list of admin users.
type AdminUserListResponse struct {
	Data []AdminUserResponse `json:"data"`
	Meta pagination.Meta     `json:"meta"`
}

// ToAdminUserResponse maps a domain user to API response.
func ToAdminUserResponse(u *domain.User) AdminUserResponse {
	return AdminUserResponse{
		ID:          u.ID.String(),
		Email:       u.Email,
		FirstName:   u.FirstName,
		LastName:    u.LastName,
		FullName:    u.FullName(),
		Phone:       u.Phone,
		Role:        u.Role.String(),
		IsActive:    u.IsActive,
		LastLoginAt: u.LastLoginAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// ToAdminUserListResponse maps a paginated user list to API response.
func ToAdminUserListResponse(result pagination.Paginated[domain.User]) AdminUserListResponse {
	items := make([]AdminUserResponse, len(result.Data))
	for i, u := range result.Data {
		items[i] = ToAdminUserResponse(&u)
	}
	return AdminUserListResponse{Data: items, Meta: result.Meta}
}
