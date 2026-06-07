package customer

// AddressType represents the purpose of a customer address.
type AddressType string

const (
	AddressHome     AddressType = "home"
	AddressWork     AddressType = "work"
	AddressBilling  AddressType = "billing"
	AddressShipping AddressType = "shipping"
)

// String returns the address type as a plain string.
func (t AddressType) String() string {
	return string(t)
}
