package theme

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/theme"
)

// Service handles theme marketplace and store style use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new theme Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// ThemeListItem is a theme with ownership flags for admin listing.
type ThemeListItem struct {
	domain.Theme
	IsPurchased   bool `json:"is_purchased"`
	IsActiveTheme bool `json:"is_active_theme"`
}

// ListThemes returns available themes with purchase status for an admin user.
func (s *Service) ListThemes(ctx context.Context, userID uuid.UUID) ([]ThemeListItem, error) {
	themes, err := s.repo.ListThemes(ctx)
	if err != nil {
		return nil, err
	}

	style, err := s.repo.GetStyle(ctx)
	if err != nil && err != domain.ErrStyleNotFound {
		return nil, err
	}

	purchases, err := s.repo.ListPurchases(ctx, userID)
	if err != nil {
		return nil, err
	}
	owned := make(map[uuid.UUID]struct{}, len(purchases))
	for _, p := range purchases {
		owned[p.ThemeID] = struct{}{}
	}

	items := make([]ThemeListItem, len(themes))
	for i, t := range themes {
		_, isPurchased := owned[t.ID]
		if t.Price == 0 {
			isPurchased = true
		}
		isActiveTheme := style != nil && style.ActiveThemeID != nil && *style.ActiveThemeID == t.ID
		items[i] = ThemeListItem{
			Theme:         t,
			IsPurchased:   isPurchased,
			IsActiveTheme: isActiveTheme,
		}
	}
	return items, nil
}

// PurchaseTheme records a theme purchase for an admin user.
func (s *Service) PurchaseTheme(ctx context.Context, themeID, userID uuid.UUID) (*domain.ThemePurchase, error) {
	t, err := s.repo.FindTheme(ctx, themeID)
	if err != nil {
		return nil, err
	}
	if !t.IsActive {
		return nil, domain.ErrThemeInactive
	}

	has, err := s.repo.HasPurchase(ctx, themeID, userID)
	if err != nil {
		return nil, err
	}
	if has {
		purchases, err := s.repo.ListPurchases(ctx, userID)
		if err != nil {
			return nil, err
		}
		for i := range purchases {
			if purchases[i].ThemeID == themeID {
				return &purchases[i], nil
			}
		}
	}

	if t.Price == 0 {
		purchase := &domain.ThemePurchase{
			ID:          uuid.New(),
			ThemeID:     themeID,
			PurchasedBy: userID,
			PurchasedAt: time.Now().UTC(),
		}
		if err := s.repo.CreatePurchase(ctx, purchase); err != nil {
			return nil, err
		}
		return purchase, nil
	}

	purchase := &domain.ThemePurchase{
		ID:          uuid.New(),
		ThemeID:     themeID,
		PurchasedBy: userID,
		PurchasedAt: time.Now().UTC(),
	}
	if err := s.repo.CreatePurchase(ctx, purchase); err != nil {
		return nil, err
	}
	return purchase, nil
}

// StyleOutput is the admin store style response.
type StyleOutput struct {
	Style       *domain.StoreStyle `json:"style"`
	ActiveTheme *domain.Theme      `json:"active_theme,omitempty"`
	Colors      domain.ColorTokens `json:"colors"`
	FontFamily  string             `json:"font_family"`
}

// GetStyle returns the current store style with resolved tokens.
func (s *Service) GetStyle(ctx context.Context) (*StyleOutput, error) {
	style, err := s.repo.GetStyle(ctx)
	if err != nil {
		return nil, err
	}

	var activeTheme *domain.Theme
	if style.ActiveThemeID != nil {
		activeTheme, err = s.repo.FindTheme(ctx, *style.ActiveThemeID)
		if err != nil && err != domain.ErrNotFound {
			return nil, err
		}
	}

	colors, font := style.ResolvedStyle(activeTheme)
	return &StyleOutput{
		Style:       style,
		ActiveTheme: activeTheme,
		Colors:      colors,
		FontFamily:  font,
	}, nil
}

// UpdateStyleInput holds partial style update data.
type UpdateStyleInput struct {
	ActiveThemeID *uuid.UUID
	Colors        *domain.ColorTokens
	FontFamily    *string
}

// UpdateStyle updates store style colors and/or font.
func (s *Service) UpdateStyle(ctx context.Context, userID uuid.UUID, input UpdateStyleInput) (*StyleOutput, error) {
	style, err := s.repo.GetStyle(ctx)
	if err != nil {
		return nil, err
	}

	if input.ActiveThemeID != nil {
		t, err := s.repo.FindTheme(ctx, *input.ActiveThemeID)
		if err != nil {
			return nil, err
		}
		if t.Price > 0 {
			has, err := s.repo.HasPurchase(ctx, t.ID, userID)
			if err != nil {
				return nil, err
			}
			if !has {
				return nil, domain.ErrNotPurchased
			}
		}
		style.ActiveThemeID = input.ActiveThemeID
		style.Colors = t.DefaultColors
		if t.DefaultFont != "" {
			style.FontFamily = t.DefaultFont
		}
	}

	if input.Colors != nil {
		style.Colors = domain.MergeColors(style.Colors, *input.Colors)
	}
	if input.FontFamily != nil {
		style.FontFamily = *input.FontFamily
	}
	style.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateStyle(ctx, style); err != nil {
		return nil, err
	}
	return s.GetStyle(ctx)
}

// PublicThemeOutput is the public storefront theme token response.
type PublicThemeOutput struct {
	Colors     domain.ColorTokens `json:"colors"`
	FontFamily string             `json:"font_family"`
}

// GetPublicTheme returns resolved design tokens for the storefront.
func (s *Service) GetPublicTheme(ctx context.Context) (*PublicThemeOutput, error) {
	output, err := s.GetStyle(ctx)
	if err != nil {
		return nil, err
	}
	return &PublicThemeOutput{
		Colors:     output.Colors,
		FontFamily: output.FontFamily,
	}, nil
}
