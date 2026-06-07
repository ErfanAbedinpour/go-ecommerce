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
