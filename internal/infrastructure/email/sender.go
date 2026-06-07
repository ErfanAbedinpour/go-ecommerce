package email

import "context"

// Sender delivers transactional emails.
type Sender interface {
	SendPasswordReset(ctx context.Context, to, resetLink string) error
}
