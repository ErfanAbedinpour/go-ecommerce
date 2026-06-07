package brand

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domain "app/internal/domain/brand"
	"app/pkg/pagination"
)

type mockRepo struct {
	brands map[uuid.UUID]*domain.Brand
	slugs  map[string]uuid.UUID
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		brands: make(map[uuid.UUID]*domain.Brand),
		slugs:  make(map[string]uuid.UUID),
	}
}

func (m *mockRepo) Create(_ context.Context, b *domain.Brand) error {
	cp := *b
	m.brands[b.ID] = &cp
	m.slugs[b.Slug] = b.ID
	return nil
}

func (m *mockRepo) Update(_ context.Context, b *domain.Brand) error {
	m.brands[b.ID] = b
	m.slugs[b.Slug] = b.ID
	return nil
}

func (m *mockRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	delete(m.brands, id)
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Brand, error) {
	b, ok := m.brands[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *b
	return &cp, nil
}

func (m *mockRepo) FindBySlug(_ context.Context, slug string) (*domain.Brand, error) {
	id, ok := m.slugs[slug]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return m.FindByID(context.Background(), id)
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter, page pagination.Params) ([]domain.Brand, int64, error) {
	items := make([]domain.Brand, 0, len(m.brands))
	for _, b := range m.brands {
		items = append(items, *b)
	}
	return items, int64(len(items)), nil
}

func (m *mockRepo) HasProducts(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func TestService_Create(t *testing.T) {
	svc := NewService(newMockRepo())
	b, err := svc.Create(context.Background(), CreateInput{Name: "Nike", IsActive: true})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if b.Slug != "nike" {
		t.Errorf("Slug = %q, want nike", b.Slug)
	}
}
