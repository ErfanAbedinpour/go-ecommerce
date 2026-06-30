package request

// StoreAccountAddressRequest is an address on the customer account profile.
type StoreAccountAddressRequest struct {
	ID         *string `json:"id" validate:"omitempty,uuid"`
	Type       string  `json:"type" validate:"omitempty,oneof=home work billing shipping"`
	Street     string  `json:"street" validate:"required,max=300"`
	City       string  `json:"city" validate:"required,max=100"`
	State      string  `json:"state" validate:"omitempty,max=100"`
	PostalCode string  `json:"postal_code" validate:"required,max=20"`
	Country    string  `json:"country" validate:"omitempty,len=2"`
	IsDefault  bool    `json:"is_default"`
}

// UpdateStoreAccountProfileRequest updates the authenticated customer profile.
type UpdateStoreAccountProfileRequest struct {
	FirstName string                       `json:"first_name" validate:"required,max=100"`
	LastName  string                       `json:"last_name" validate:"required,max=100"`
	Phone     string                       `json:"phone" validate:"omitempty,max=20"`
	Addresses []StoreAccountAddressRequest `json:"addresses" validate:"omitempty,dive"`
}
