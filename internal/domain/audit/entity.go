package audit

import (
	"context"
	"time"
)

type AuditLog struct {
	ID           string
	AdminUserID  string
	Action       string
	ResourceType string
	ResourceID   string
	OldValue     []byte
	NewValue     []byte
	IPAddress    string
	UserAgent    string
	CreatedAt    time.Time
}

type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
}
