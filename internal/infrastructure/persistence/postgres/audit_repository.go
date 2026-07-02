package postgres

import (
	"context"

	"gorm.io/gorm"

	domain "app/internal/domain/audit"
)

type auditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) domain.Repository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs 
		(user_id, action, resource_type, resource_id, old_value, new_value, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	return r.db.WithContext(ctx).Raw(query,
		log.UserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		log.OldValue,
		log.NewValue,
		log.IPAddress,
		log.UserAgent,
	).Row().Scan(&log.ID, &log.CreatedAt)
}
