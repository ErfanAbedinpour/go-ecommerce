package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/contact"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

// ContactMessageRepository implements contact.Repository using GORM.
type ContactMessageRepository struct {
	db *gorm.DB
}

// NewContactMessageRepository creates a new ContactMessageRepository.
func NewContactMessageRepository(db *gorm.DB) *ContactMessageRepository {
	return &ContactMessageRepository{db: db}
}

func (r *ContactMessageRepository) Create(ctx context.Context, m *contact.Message) error {
	model := toContactMessageModel(m)
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *ContactMessageRepository) FindByID(ctx context.Context, id uuid.UUID) (*contact.Message, error) {
	var m models.ContactMessageModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, contact.ErrNotFound
		}
		return nil, err
	}
	return toContactMessageDomain(&m), nil
}

func (r *ContactMessageRepository) List(ctx context.Context, filter contact.ListFilter, page pagination.Params) ([]contact.Message, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.ContactMessageModel{})
	query = r.applyFilters(query, filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.ContactMessageModel
	err := query.Order(r.orderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toContactMessagesDomain(items), total, nil
}

func (r *ContactMessageRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status contact.Status) error {
	result := r.db.WithContext(ctx).
		Model(&models.ContactMessageModel{}).
		Where("id = ?", id).
		Update("status", string(status))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return contact.ErrNotFound
	}
	return nil
}

func (r *ContactMessageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&models.ContactMessageModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return contact.ErrNotFound
	}
	return nil
}

func (r *ContactMessageRepository) applyFilters(query *gorm.DB, filter contact.ListFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}
	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(email) LIKE ? OR LOWER(subject) LIKE ?",
			pattern, pattern, pattern,
		)
	}
	if filter.From != nil {
		query = query.Where("created_at >= ?", filter.From.UTC())
	}
	if filter.To != nil {
		to := filter.To.UTC().Add(24*time.Hour - time.Nanosecond)
		query = query.Where("created_at <= ?", to)
	}
	return query
}

func (r *ContactMessageRepository) orderClause(page pagination.Params) string {
	allowed := map[string]string{
		"created_at": "created_at",
		"name":       "name",
		"status":     "status",
	}
	column, ok := allowed[page.Sort]
	if !ok {
		column = "created_at"
	}
	order := "DESC"
	if strings.EqualFold(page.Order, "asc") {
		order = "ASC"
	}
	return fmt.Sprintf("%s %s", column, order)
}

func toContactMessageModel(m *contact.Message) *models.ContactMessageModel {
	return &models.ContactMessageModel{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Phone:     nullableString(m.Phone),
		Subject:   nullableString(m.Subject),
		Message:   m.Message,
		Source:    string(m.Source),
		Status:   string(m.Status),
		CreatedAt: m.CreatedAt,
	}
}

func toContactMessageDomain(m *models.ContactMessageModel) *contact.Message {
	msg := &contact.Message{
		ID:        m.ID,
		Name:      m.Name,
		Email:     m.Email,
		Message:   m.Message,
		Source:    contact.Source(m.Source),
		Status:   contact.Status(m.Status),
		CreatedAt: m.CreatedAt,
	}
	if m.Phone != nil {
		msg.Phone = *m.Phone
	}
	if m.Subject != nil {
		msg.Subject = *m.Subject
	}
	return msg
}

func toContactMessagesDomain(items []models.ContactMessageModel) []contact.Message {
	result := make([]contact.Message, len(items))
	for i, m := range items {
		result[i] = *toContactMessageDomain(&m)
	}
	return result
}

var _ contact.Repository = (*ContactMessageRepository)(nil)
