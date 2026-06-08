package request

// CreateAdminUserRequest is the request body for creating an admin user.
type CreateAdminUserRequest struct {
	Email     string `json:"email" validate:"required,email,max=255"`
	Password  string `json:"password" validate:"required,min=8,max=128"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Phone     string `json:"phone" validate:"omitempty,max=20"`
	Role      string `json:"role" validate:"omitempty,oneof=admin customer"`
	IsActive  *bool  `json:"is_active"`
}

// UpdateAdminUserRequest is the request body for updating an admin user.
type UpdateAdminUserRequest struct {
	Email     *string `json:"email" validate:"omitempty,email,max=255"`
	Password  *string `json:"password" validate:"omitempty,min=8,max=128"`
	FirstName *string `json:"first_name" validate:"omitempty,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,max=100"`
	Phone     *string `json:"phone" validate:"omitempty,max=20"`
	Role      *string `json:"role" validate:"omitempty,oneof=admin customer"`
	IsActive  *bool   `json:"is_active"`
}
