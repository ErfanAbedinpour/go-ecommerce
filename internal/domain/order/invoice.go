package order

import "time"

// Invoice is a read model for printable order invoices.
type Invoice struct {
	InvoiceNumber string
	IssuedAt      time.Time
	Store         InvoiceStore
	Order         *Order
}

// InvoiceStore holds merchant details shown on invoices.
type InvoiceStore struct {
	Name    string
	URL     string
	LogoURL string
	Email   string
	Phone   string
	Address string
	City    string
	Country string
}
