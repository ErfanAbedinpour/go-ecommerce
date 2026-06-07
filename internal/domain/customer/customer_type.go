package customer

// CustomerType represents how a customer account was created.
type CustomerType string

const (
	TypeRegistered CustomerType = "registered"
	TypeGuest      CustomerType = "guest"
)

// ParseCustomerType validates and parses a customer type string.
func ParseCustomerType(value string) (CustomerType, error) {
	switch CustomerType(value) {
	case TypeRegistered, TypeGuest:
		return CustomerType(value), nil
	default:
		return "", ErrInvalidType
	}
}

// String returns the customer type as a plain string.
func (t CustomerType) String() string {
	return string(t)
}
