package settings

import (
	"context"
	"testing"

	domain "app/internal/domain/settings"
)

type mockRepo struct {
	store *domain.StoreSettings
}

func (m *mockRepo) Get(_ context.Context) (*domain.StoreSettings, error) {
	if m.store == nil {
		return &domain.StoreSettings{}, nil
	}
	return m.store, nil
}

func (m *mockRepo) UpdateSite(_ context.Context, site domain.Site) (*domain.Site, error) {
	if m.store == nil {
		m.store = &domain.StoreSettings{}
	}
	m.store.Site = site
	return &site, nil
}

func (m *mockRepo) UpdateContact(_ context.Context, contact domain.Contact) (*domain.Contact, error) {
	if m.store == nil {
		m.store = &domain.StoreSettings{}
	}
	m.store.Contact = contact
	return &contact, nil
}

func (m *mockRepo) UpdateSocial(_ context.Context, social domain.Social) (*domain.Social, error) {
	if m.store == nil {
		m.store = &domain.StoreSettings{}
	}
	m.store.Social = social
	return &social, nil
}

func (m *mockRepo) UpdateSEO(_ context.Context, seo domain.SEO) (*domain.SEO, error) {
	if m.store == nil {
		m.store = &domain.StoreSettings{}
	}
	m.store.SEO = seo
	return &seo, nil
}

func (m *mockRepo) UpdateNavigation(_ context.Context, items []domain.NavItem) ([]domain.NavItem, error) {
	if m.store == nil {
		m.store = &domain.StoreSettings{}
	}
	m.store.Navigation = items
	return items, nil
}

func TestService_UpdateSite(t *testing.T) {
	repo := &mockRepo{}
	svc := NewService(repo)

	site, err := svc.UpdateSite(context.Background(), domain.Site{
		Name: "My Shop",
		URL:  "https://shop.example.com",
	})
	if err != nil {
		t.Fatalf("UpdateSite() error = %v", err)
	}
	if site.Name != "My Shop" {
		t.Fatalf("name = %q, want My Shop", site.Name)
	}

	got, err := svc.GetSite(context.Background())
	if err != nil {
		t.Fatalf("GetSite() error = %v", err)
	}
	if got.URL != "https://shop.example.com" {
		t.Fatalf("url = %q", got.URL)
	}
}

func TestService_UpdateNavigation(t *testing.T) {
	svc := NewService(&mockRepo{})

	items, err := svc.UpdateNavigation(context.Background(), []domain.NavItem{
		{ID: "home", Label: "Home", URL: "/", SortOrder: 0, IsActive: true},
	})
	if err != nil {
		t.Fatalf("UpdateNavigation() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d, want 1", len(items))
	}
}
