package settings

import (
	"context"

	domain "app/internal/domain/settings"
)

// Service handles store settings use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new settings Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// GetSite returns site identity settings.
func (s *Service) GetSite(ctx context.Context) (*domain.Site, error) {
	store, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &store.Site, nil
}

// UpdateSite updates site identity settings.
func (s *Service) UpdateSite(ctx context.Context, site domain.Site) (*domain.Site, error) {
	return s.repo.UpdateSite(ctx, site)
}

// GetContact returns contact settings.
func (s *Service) GetContact(ctx context.Context) (*domain.Contact, error) {
	store, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &store.Contact, nil
}

// UpdateContact updates contact settings.
func (s *Service) UpdateContact(ctx context.Context, contact domain.Contact) (*domain.Contact, error) {
	return s.repo.UpdateContact(ctx, contact)
}

// GetSocial returns social media settings.
func (s *Service) GetSocial(ctx context.Context) (*domain.Social, error) {
	store, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &store.Social, nil
}

// UpdateSocial updates social media settings.
func (s *Service) UpdateSocial(ctx context.Context, social domain.Social) (*domain.Social, error) {
	return s.repo.UpdateSocial(ctx, social)
}

// GetSEO returns SEO settings.
func (s *Service) GetSEO(ctx context.Context) (*domain.SEO, error) {
	store, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &store.SEO, nil
}

// UpdateSEO updates SEO settings.
func (s *Service) UpdateSEO(ctx context.Context, seo domain.SEO) (*domain.SEO, error) {
	return s.repo.UpdateSEO(ctx, seo)
}

// GetNavigation returns the navigation menu tree.
func (s *Service) GetNavigation(ctx context.Context) ([]domain.NavItem, error) {
	store, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return store.Navigation, nil
}

// UpdateNavigation replaces the navigation menu tree.
func (s *Service) UpdateNavigation(ctx context.Context, items []domain.NavItem) ([]domain.NavItem, error) {
	return s.repo.UpdateNavigation(ctx, items)
}

// PublicSettings is the storefront-safe settings projection.
type PublicSettings struct {
	Site    domain.Site    `json:"site"`
	Contact domain.Contact `json:"contact"`
	Social  domain.Social  `json:"social"`
	SEO     domain.SEO     `json:"seo"`
}

// GetPublicSettings returns public storefront settings.
func (s *Service) GetPublicSettings(ctx context.Context) (*PublicSettings, error) {
	store, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return &PublicSettings{
		Site:    store.Site,
		Contact: store.Contact,
		Social:  store.Social,
		SEO:     store.SEO,
	}, nil
}

// GetStorefrontNavigation returns the storefront navigation menu.
func (s *Service) GetStorefrontNavigation(ctx context.Context) ([]domain.NavItem, error) {
	store, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return store.StorefrontNavigation, nil
}

// UpdateStorefrontNavigation replaces the storefront navigation menu.
func (s *Service) UpdateStorefrontNavigation(ctx context.Context, items []domain.NavItem) ([]domain.NavItem, error) {
	return s.repo.UpdateStorefrontNavigation(ctx, items)
}

// GetContactSectionImage returns the contact section image URL.
func (s *Service) GetContactSectionImage(ctx context.Context) (string, error) {
	store, err := s.repo.Get(ctx)
	if err != nil {
		return "", err
	}
	return store.ContactSectionImageURL, nil
}

// UpdateContactSectionImage updates the contact section image URL.
func (s *Service) UpdateContactSectionImage(ctx context.Context, imageURL string) (string, error) {
	return s.repo.UpdateContactSectionImage(ctx, imageURL)
}
