package theme

import (
	"time"

	"github.com/google/uuid"
)

// ColorTokens holds the 12 storefront design color tokens.
type ColorTokens struct {
	Primary             string `json:"primary"`
	PrimaryForeground   string `json:"primary_foreground"`
	Secondary           string `json:"secondary"`
	SecondaryForeground string `json:"secondary_foreground"`
	Accent              string `json:"accent"`
	AccentForeground    string `json:"accent_foreground"`
	Background          string `json:"background"`
	Foreground          string `json:"foreground"`
	Muted               string `json:"muted"`
	MutedForeground     string `json:"muted_foreground"`
	Border              string `json:"border"`
	Destructive         string `json:"destructive"`
}

// Theme is a marketplace theme catalog entry.
type Theme struct {
	ID              uuid.UUID
	Name            string
	Slug            string
	Description     string
	PreviewImageURL string
	Price           float64
	IsActive        bool
	DefaultColors   ColorTokens
	DefaultFont     string
	CreatedAt       time.Time
}

// ThemePurchase records an admin user's theme purchase.
type ThemePurchase struct {
	ID          uuid.UUID
	ThemeID     uuid.UUID
	PurchasedBy uuid.UUID
	PurchasedAt time.Time
}

// StoreStyle is the singleton active storefront style configuration.
type StoreStyle struct {
	ID            uuid.UUID
	ActiveThemeID *uuid.UUID
	Colors        ColorTokens
	FontFamily    string
	UpdatedAt     time.Time
}

// MergeColors returns tokens with overrides applied on top of defaults.
func MergeColors(defaults, overrides ColorTokens) ColorTokens {
	merged := defaults
	if overrides.Primary != "" {
		merged.Primary = overrides.Primary
	}
	if overrides.PrimaryForeground != "" {
		merged.PrimaryForeground = overrides.PrimaryForeground
	}
	if overrides.Secondary != "" {
		merged.Secondary = overrides.Secondary
	}
	if overrides.SecondaryForeground != "" {
		merged.SecondaryForeground = overrides.SecondaryForeground
	}
	if overrides.Accent != "" {
		merged.Accent = overrides.Accent
	}
	if overrides.AccentForeground != "" {
		merged.AccentForeground = overrides.AccentForeground
	}
	if overrides.Background != "" {
		merged.Background = overrides.Background
	}
	if overrides.Foreground != "" {
		merged.Foreground = overrides.Foreground
	}
	if overrides.Muted != "" {
		merged.Muted = overrides.Muted
	}
	if overrides.MutedForeground != "" {
		merged.MutedForeground = overrides.MutedForeground
	}
	if overrides.Border != "" {
		merged.Border = overrides.Border
	}
	if overrides.Destructive != "" {
		merged.Destructive = overrides.Destructive
	}
	return merged
}

// ResolvedStyle returns merged colors and effective font for public consumption.
func (s *StoreStyle) ResolvedStyle(theme *Theme) (ColorTokens, string) {
	font := s.FontFamily
	colors := s.Colors
	if theme != nil {
		colors = MergeColors(theme.DefaultColors, s.Colors)
		if font == "" {
			font = theme.DefaultFont
		}
	}
	return colors, font
}
