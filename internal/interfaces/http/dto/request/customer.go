package request

// UpdateCustomerRequest is the request body for updating a customer profile.
type UpdateCustomerRequest struct {
	Email     *string `json:"email" validate:"omitempty,email,max=255"`
	FirstName *string `json:"first_name" validate:"omitempty,max=100"`
	LastName  *string `json:"last_name" validate:"omitempty,max=100"`
	Phone     *string `json:"phone" validate:"omitempty,max=20"`
	Type      *string `json:"type" validate:"omitempty,oneof=registered guest"`
}
