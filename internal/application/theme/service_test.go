package theme

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/theme"
)

type mockThemeRepo struct {
	themes    map[uuid.UUID]*domain.Theme
	purchases map[string]struct{}
	style     *domain.StoreStyle
}

func purchaseKey(themeID, userID uuid.UUID) string {
	return themeID.String() + ":" + userID.String()
}

func (m *mockThemeRepo) ListThemes(ctx context.Context) ([]domain.Theme, error) {
	items := make([]domain.Theme, 0, len(m.themes))
	for _, t := range m.themes {
		items = append(items, *t)
	}
	return items, nil
}

func (m *mockThemeRepo) FindTheme(_ context.Context, id uuid.UUID) (*domain.Theme, error) {
	t, ok := m.themes[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return t, nil
}

func (m *mockThemeRepo) CreatePurchase(_ context.Context, purchase *domain.ThemePurchase) error {
	if m.purchases == nil {
		m.purchases = make(map[string]struct{})
	}
	m.purchases[purchaseKey(purchase.ThemeID, purchase.PurchasedBy)] = struct{}{}
	return nil
}

func (m *mockThemeRepo) HasPurchase(_ context.Context, themeID, userID uuid.UUID) (bool, error) {
	_, ok := m.purchases[purchaseKey(themeID, userID)]
	return ok, nil
}

func (m *mockThemeRepo) ListPurchases(_ context.Context, userID uuid.UUID) ([]domain.ThemePurchase, error) {
	var items []domain.ThemePurchase
	for themeID := range m.themes {
		if _, ok := m.purchases[purchaseKey(themeID, userID)]; ok {
			items = append(items, domain.ThemePurchase{
				ThemeID:     themeID,
				PurchasedBy: userID,
				PurchasedAt: time.Now().UTC(),
			})
		}
	}
	return items, nil
}

func (m *mockThemeRepo) GetStyle(context.Context) (*domain.StoreStyle, error) {
	if m.style == nil {
		return nil, domain.ErrStyleNotFound
	}
	return m.style, nil
}

func (m *mockThemeRepo) UpdateStyle(_ context.Context, style *domain.StoreStyle) error {
	m.style = style
	return nil
}

func TestPurchaseTheme_FreeTheme(t *testing.T) {
	themeID := uuid.New()
	userID := uuid.New()
	repo := &mockThemeRepo{
		themes: map[uuid.UUID]*domain.Theme{
			themeID: {ID: themeID, Name: "Modern Blue", Slug: "modern-blue", Price: 0, IsActive: true},
		},
	}
	svc := NewService(repo)

	purchase, err := svc.PurchaseTheme(context.Background(), themeID, userID)
	if err != nil {
		t.Fatalf("PurchaseTheme() error = %v", err)
	}
	if purchase.ThemeID != themeID {
		t.Fatalf("unexpected purchase: %+v", purchase)
	}
}

func TestUpdateStyle_BlockUnpurchasedPaidTheme(t *testing.T) {
	freeID := uuid.New()
	paidID := uuid.New()
	userID := uuid.New()
	repo := &mockThemeRepo{
		themes: map[uuid.UUID]*domain.Theme{
			freeID: {ID: freeID, Slug: "minimal-light", Price: 0, IsActive: true, DefaultColors: domain.ColorTokens{Primary: "#000"}},
			paidID: {ID: paidID, Slug: "bold-dark", Price: 99000, IsActive: true, DefaultColors: domain.ColorTokens{Primary: "#111"}},
		},
		style: &domain.StoreStyle{
			ID:         uuid.New(),
			Colors:     domain.ColorTokens{Primary: "#fff"},
			FontFamily: "Inter",
			UpdatedAt:  time.Now().UTC(),
		},
	}
	svc := NewService(repo)

	_, err := svc.UpdateStyle(context.Background(), userID, UpdateStyleInput{ActiveThemeID: &paidID})
	if err != domain.ErrNotPurchased {
		t.Fatalf("UpdateStyle() error = %v, want ErrNotPurchased", err)
	}
}

func TestUpdateStyle_AllowsFreeTheme(t *testing.T) {
	freeID := uuid.New()
	userID := uuid.New()
	repo := &mockThemeRepo{
		themes: map[uuid.UUID]*domain.Theme{
			freeID: {ID: freeID, Slug: "modern-blue", Price: 0, IsActive: true, DefaultColors: domain.ColorTokens{Primary: "#2563eb"}},
		},
		style: &domain.StoreStyle{
			ID:         uuid.New(),
			Colors:     domain.ColorTokens{},
			FontFamily: "Inter",
			UpdatedAt:  time.Now().UTC(),
		},
	}
	svc := NewService(repo)

	out, err := svc.UpdateStyle(context.Background(), userID, UpdateStyleInput{ActiveThemeID: &freeID})
	if err != nil {
		t.Fatalf("UpdateStyle() error = %v", err)
	}
	if out.ActiveTheme == nil || out.ActiveTheme.ID != freeID {
		t.Fatalf("expected active theme %v, got %+v", freeID, out.ActiveTheme)
	}
}
