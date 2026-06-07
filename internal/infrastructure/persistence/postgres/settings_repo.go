package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/settings"
	"app/internal/infrastructure/persistence/models"
)

const storeSettingsID = "f0000000-0000-0000-0000-000000000001"

// SettingsRepository implements settings.Repository using GORM.
type SettingsRepository struct {
	db *gorm.DB
}

// NewSettingsRepository creates a new SettingsRepository.
func NewSettingsRepository(db *gorm.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (r *SettingsRepository) Get(ctx context.Context) (*settings.StoreSettings, error) {
	m, err := r.loadOrCreate(ctx)
	if err != nil {
		return nil, err
	}
	return toStoreSettingsDomain(m)
}

func (r *SettingsRepository) UpdateSite(ctx context.Context, site settings.Site) (*settings.Site, error) {
	m, err := r.loadOrCreate(ctx)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(site)
	if err != nil {
		return nil, err
	}
	m.Site = data
	if err := r.save(ctx, m); err != nil {
		return nil, err
	}
	return &site, nil
}

func (r *SettingsRepository) UpdateContact(ctx context.Context, contact settings.Contact) (*settings.Contact, error) {
	m, err := r.loadOrCreate(ctx)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(contact)
	if err != nil {
		return nil, err
	}
	m.Contact = data
	if err := r.save(ctx, m); err != nil {
		return nil, err
	}
	return &contact, nil
}

func (r *SettingsRepository) UpdateSocial(ctx context.Context, social settings.Social) (*settings.Social, error) {
	m, err := r.loadOrCreate(ctx)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(social)
	if err != nil {
		return nil, err
	}
	m.Social = data
	if err := r.save(ctx, m); err != nil {
		return nil, err
	}
	return &social, nil
}

func (r *SettingsRepository) UpdateSEO(ctx context.Context, seo settings.SEO) (*settings.SEO, error) {
	m, err := r.loadOrCreate(ctx)
	if err != nil {
		return nil, err
	}
	data, err := json.Marshal(seo)
	if err != nil {
		return nil, err
	}
	m.SEO = data
	if err := r.save(ctx, m); err != nil {
		return nil, err
	}
	return &seo, nil
}

func (r *SettingsRepository) UpdateNavigation(ctx context.Context, items []settings.NavItem) ([]settings.NavItem, error) {
	m, err := r.loadOrCreate(ctx)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []settings.NavItem{}
	}
	data, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	m.Navigation = data
	if err := r.save(ctx, m); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *SettingsRepository) loadOrCreate(ctx context.Context) (*models.StoreSettingsModel, error) {
	id, err := uuid.Parse(storeSettingsID)
	if err != nil {
		return nil, err
	}

	var m models.StoreSettingsModel
	err = r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err == nil {
		return &m, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	m = models.StoreSettingsModel{
		ID:         id,
		Site:       []byte("{}"),
		Contact:    []byte("{}"),
		Social:     []byte("{}"),
		SEO:        []byte("{}"),
		Navigation: []byte("[]"),
	}
	if err := r.db.WithContext(ctx).Create(&m).Error; err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *SettingsRepository) save(ctx context.Context, m *models.StoreSettingsModel) error {
	m.UpdatedAt = time.Now().UTC()
	return r.db.WithContext(ctx).Save(m).Error
}

func toStoreSettingsDomain(m *models.StoreSettingsModel) (*settings.StoreSettings, error) {
	var site settings.Site
	var contact settings.Contact
	var social settings.Social
	var seo settings.SEO
	var navigation []settings.NavItem

	if len(m.Site) > 0 {
		if err := json.Unmarshal(m.Site, &site); err != nil {
			return nil, err
		}
	}
	if len(m.Contact) > 0 {
		if err := json.Unmarshal(m.Contact, &contact); err != nil {
			return nil, err
		}
	}
	if len(m.Social) > 0 {
		if err := json.Unmarshal(m.Social, &social); err != nil {
			return nil, err
		}
	}
	if len(m.SEO) > 0 {
		if err := json.Unmarshal(m.SEO, &seo); err != nil {
			return nil, err
		}
	}
	if len(m.Navigation) > 0 {
		if err := json.Unmarshal(m.Navigation, &navigation); err != nil {
			return nil, err
		}
	}
	if navigation == nil {
		navigation = []settings.NavItem{}
	}

	return &settings.StoreSettings{
		Site:       site,
		Contact:    contact,
		Social:     social,
		SEO:        seo,
		Navigation: navigation,
		UpdatedAt:  m.UpdatedAt,
	}, nil
}

var _ settings.Repository = (*SettingsRepository)(nil)
