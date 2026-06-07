package settings

import "context"

// Repository defines persistence for store settings.
type Repository interface {
	Get(ctx context.Context) (*StoreSettings, error)
	UpdateSite(ctx context.Context, site Site) (*Site, error)
	UpdateContact(ctx context.Context, contact Contact) (*Contact, error)
	UpdateSocial(ctx context.Context, social Social) (*Social, error)
	UpdateSEO(ctx context.Context, seo SEO) (*SEO, error)
	UpdateNavigation(ctx context.Context, items []NavItem) ([]NavItem, error)
}
