package email

import "context"

// Sender delivers transactional emails.
type Sender interface {
	SendPasswordReset(ctx context.Context, to, resetLink string) error
	SendOrderConfirmation(ctx context.Context, to string, orderNumber string, total float64) error
}
