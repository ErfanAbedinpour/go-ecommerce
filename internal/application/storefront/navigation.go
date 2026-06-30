package storefront

import (
	"context"

	domainsettings "app/internal/domain/settings"
)

// StoreNavigation is the public storefront navigation menu.
type StoreNavigation struct {
	Items []domainsettings.NavItem `json:"items"`
}

// GetStoreNavigation returns active storefront navigation items.
func (s *Service) GetStoreNavigation(ctx context.Context) (*StoreNavigation, error) {
	store, err := s.settings.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &StoreNavigation{
		Items: filterActiveNavItems(store.StorefrontNavigation),
	}, nil
}

func filterActiveNavItems(items []domainsettings.NavItem) []domainsettings.NavItem {
	if len(items) == 0 {
		return []domainsettings.NavItem{}
	}

	result := make([]domainsettings.NavItem, 0, len(items))
	for _, item := range items {
		if !item.IsActive {
			continue
		}
		item.Children = filterActiveNavItems(item.Children)
		result = append(result, item)
	}
	return result
}
