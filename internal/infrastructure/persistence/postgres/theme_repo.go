package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/theme"
	"app/internal/infrastructure/persistence/models"
)

const storeStyleID = "e1000000-0000-0000-0000-000000000001"

// ThemeRepository implements theme.Repository using GORM.
type ThemeRepository struct {
	db *gorm.DB
}

// NewThemeRepository creates a new ThemeRepository.
func NewThemeRepository(db *gorm.DB) *ThemeRepository {
	return &ThemeRepository{db: db}
}

func (r *ThemeRepository) ListThemes(ctx context.Context) ([]theme.Theme, error) {
	var items []models.StoreThemeModel
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toThemesDomain(items)
}

func (r *ThemeRepository) FindTheme(ctx context.Context, id uuid.UUID) (*theme.Theme, error) {
	var m models.StoreThemeModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, theme.ErrNotFound
		}
		return nil, err
	}
	return toThemeDomain(&m)
}

func (r *ThemeRepository) CreatePurchase(ctx context.Context, purchase *theme.ThemePurchase) error {
	m := &models.ThemePurchaseModel{
		ID:          purchase.ID,
		ThemeID:     purchase.ThemeID,
		PurchasedBy: purchase.PurchasedBy,
		PurchasedAt: purchase.PurchasedAt,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return theme.ErrAlreadyPurchased
		}
		return err
	}
	return nil
}

func (r *ThemeRepository) HasPurchase(ctx context.Context, themeID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.ThemePurchaseModel{}).
		Where("theme_id = ? AND purchased_by = ?", themeID, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *ThemeRepository) ListPurchases(ctx context.Context, userID uuid.UUID) ([]theme.ThemePurchase, error) {
	var items []models.ThemePurchaseModel
	err := r.db.WithContext(ctx).
		Where("purchased_by = ?", userID).
		Order("purchased_at DESC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toThemePurchasesDomain(items), nil
}

func (r *ThemeRepository) GetStyle(ctx context.Context) (*theme.StoreStyle, error) {
	id, err := uuid.Parse(storeStyleID)
	if err != nil {
		return nil, err
	}

	var m models.StoreStyleModel
	err = r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, theme.ErrStyleNotFound
		}
		return nil, err
	}
	return toStoreStyleDomain(&m)
}

func (r *ThemeRepository) UpdateStyle(ctx context.Context, style *theme.StoreStyle) error {
	colors, err := json.Marshal(style.Colors)
	if err != nil {
		return err
	}

	updates := map[string]any{
		"active_theme_id": style.ActiveThemeID,
		"colors":          colors,
		"updated_at":      time.Now().UTC(),
	}
	if style.FontFamily != "" {
		updates["font_family"] = style.FontFamily
	}

	result := r.db.WithContext(ctx).
		Model(&models.StoreStyleModel{}).
		Where("id = ?", style.ID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return theme.ErrStyleNotFound
	}
	return nil
}

var _ theme.Repository = (*ThemeRepository)(nil)

func toThemeDomain(m *models.StoreThemeModel) (*theme.Theme, error) {
	t := &theme.Theme{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		Price:     m.Price,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
	}
	if m.Description != nil {
		t.Description = *m.Description
	}
	if m.PreviewImageURL != nil {
		t.PreviewImageURL = *m.PreviewImageURL
	}
	if m.DefaultFont != nil {
		t.DefaultFont = *m.DefaultFont
	}
	if len(m.DefaultColors) > 0 {
		if err := json.Unmarshal(m.DefaultColors, &t.DefaultColors); err != nil {
			return nil, err
		}
	}
	return t, nil
}

func toThemesDomain(items []models.StoreThemeModel) ([]theme.Theme, error) {
	result := make([]theme.Theme, len(items))
	for i, m := range items {
		t, err := toThemeDomain(&m)
		if err != nil {
			return nil, err
		}
		result[i] = *t
	}
	return result, nil
}

func toThemePurchasesDomain(items []models.ThemePurchaseModel) []theme.ThemePurchase {
	result := make([]theme.ThemePurchase, len(items))
	for i, m := range items {
		result[i] = theme.ThemePurchase{
			ID:          m.ID,
			ThemeID:     m.ThemeID,
			PurchasedBy: m.PurchasedBy,
			PurchasedAt: m.PurchasedAt,
		}
	}
	return result
}

func toStoreStyleDomain(m *models.StoreStyleModel) (*theme.StoreStyle, error) {
	s := &theme.StoreStyle{
		ID:            m.ID,
		ActiveThemeID: m.ActiveThemeID,
		UpdatedAt:     m.UpdatedAt,
	}
	if m.FontFamily != nil {
		s.FontFamily = *m.FontFamily
	}
	if len(m.Colors) > 0 {
		if err := json.Unmarshal(m.Colors, &s.Colors); err != nil {
			return nil, err
		}
	}
	return s, nil
}
