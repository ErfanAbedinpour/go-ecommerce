package theme

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines persistence for the theme marketplace and store style.
type Repository interface {
	ListThemes(ctx context.Context) ([]Theme, error)
	FindTheme(ctx context.Context, id uuid.UUID) (*Theme, error)
	CreatePurchase(ctx context.Context, purchase *ThemePurchase) error
	HasPurchase(ctx context.Context, themeID, userID uuid.UUID) (bool, error)
	ListPurchases(ctx context.Context, userID uuid.UUID) ([]ThemePurchase, error)
	GetStyle(ctx context.Context) (*StoreStyle, error)
	UpdateStyle(ctx context.Context, style *StoreStyle) error
}
