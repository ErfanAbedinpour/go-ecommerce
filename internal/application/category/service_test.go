package category

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domain "app/internal/domain/category"
	"app/pkg/pagination"
)

type mockRepo struct {
	categories map[uuid.UUID]*domain.Category
	slugs      map[string]uuid.UUID
	children   map[uuid.UUID]bool
	products   map[uuid.UUID]bool
	descendant map[string]bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		categories: make(map[uuid.UUID]*domain.Category),
		slugs:      make(map[string]uuid.UUID),
		children:   make(map[uuid.UUID]bool),
		products:   make(map[uuid.UUID]bool),
		descendant: make(map[string]bool),
	}
}

func (m *mockRepo) Create(_ context.Context, c *domain.Category) error {
	cp := *c
	m.categories[c.ID] = &cp
	m.slugs[c.Slug] = c.ID
	return nil
}

func (m *mockRepo) Update(_ context.Context, c *domain.Category) error {
	m.categories[c.ID] = c
	m.slugs[c.Slug] = c.ID
	return nil
}

func (m *mockRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	delete(m.categories, id)
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Category, error) {
	c, ok := m.categories[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *mockRepo) FindBySlug(_ context.Context, slug string) (*domain.Category, error) {
	id, ok := m.slugs[slug]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return m.FindByID(context.Background(), id)
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter, page pagination.Params) ([]domain.Category, int64, error) {
	items := make([]domain.Category, 0, len(m.categories))
	for _, c := range m.categories {
		items = append(items, *c)
	}
	return items, int64(len(items)), nil
}

func (m *mockRepo) ListAll(_ context.Context, _ domain.ListFilter) ([]domain.Category, error) {
	items, _, err := m.List(context.Background(), domain.ListFilter{}, pagination.Params{})
	return items, err
}

func (m *mockRepo) HasChildren(_ context.Context, id uuid.UUID) (bool, error) {
	return m.children[id], nil
}

func (m *mockRepo) HasProducts(_ context.Context, id uuid.UUID) (bool, error) {
	return m.products[id], nil
}

func (m *mockRepo) IsDescendant(_ context.Context, ancestorID, descendantID uuid.UUID) (bool, error) {
	key := ancestorID.String() + ":" + descendantID.String()
	return m.descendant[key], nil
}

func TestService_Create(t *testing.T) {
	svc := NewService(newMockRepo())
	c, err := svc.Create(context.Background(), CreateInput{Name: "Electronics"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if c.Slug != "electronics" {
		t.Errorf("Slug = %q, want electronics", c.Slug)
	}
}

func TestService_Create_SlugConflict(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{Name: "Electronics"})
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = svc.Create(context.Background(), CreateInput{Name: "Other", Slug: "electronics"})
	if err != domain.ErrSlugConflict {
		t.Errorf("expected slug conflict, got %v", err)
	}
}

func TestService_Update_InvalidParent(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	parent, _ := svc.Create(context.Background(), CreateInput{Name: "Parent"})
	child, _ := svc.Create(context.Background(), CreateInput{Name: "Child", ParentID: &parent.ID})

	repo.descendant[parent.ID.String()+":"+child.ID.String()] = true

	_, err := svc.Update(context.Background(), parent.ID, UpdateInput{ParentID: &child.ID})
	if err != domain.ErrInvalidParent {
		t.Errorf("expected invalid parent, got %v", err)
	}
}

func TestService_Delete_HasChildren(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	c, _ := svc.Create(context.Background(), CreateInput{Name: "Parent"})
	repo.children[c.ID] = true

	err := svc.Delete(context.Background(), c.ID)
	if err != domain.ErrHasChildren {
		t.Errorf("expected has children, got %v", err)
	}
}

func TestService_Delete_HasProducts(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	c, _ := svc.Create(context.Background(), CreateInput{Name: "Category"})
	repo.products[c.ID] = true

	err := svc.Delete(context.Background(), c.ID)
	if err != domain.ErrHasProducts {
		t.Errorf("expected has products, got %v", err)
	}
}

func TestBuildTree(t *testing.T) {
	rootID := uuid.New()
	childID := uuid.New()
	parentID := rootID

	items := []domain.Category{
		{ID: rootID, Name: "Root"},
		{ID: childID, Name: "Child", ParentID: &parentID},
	}

	tree := buildTree(items)
	if len(tree) != 1 {
		t.Fatalf("tree roots = %d, want 1", len(tree))
	}
	if len(tree[0].Children) != 1 {
		t.Fatalf("children = %d, want 1", len(tree[0].Children))
	}
	if tree[0].Children[0].Name != "Child" {
		t.Errorf("child name = %q, want Child", tree[0].Children[0].Name)
	}
}
