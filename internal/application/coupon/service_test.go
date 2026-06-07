package coupon

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domain "app/internal/domain/coupon"
	"app/pkg/pagination"
)

type mockRepo struct {
	coupons map[uuid.UUID]*domain.Coupon
	codes   map[string]uuid.UUID
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		coupons: make(map[uuid.UUID]*domain.Coupon),
		codes:   make(map[string]uuid.UUID),
	}
}

func (m *mockRepo) Create(_ context.Context, c *domain.Coupon) error {
	cp := *c
	m.coupons[c.ID] = &cp
	m.codes[c.Code] = c.ID
	return nil
}

func (m *mockRepo) Update(_ context.Context, c *domain.Coupon) error {
	m.coupons[c.ID] = c
	m.codes[c.Code] = c.ID
	return nil
}

func (m *mockRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	delete(m.coupons, id)
	return nil
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Coupon, error) {
	c, ok := m.coupons[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *c
	return &cp, nil
}

func (m *mockRepo) FindByCode(_ context.Context, code string) (*domain.Coupon, error) {
	id, ok := m.codes[code]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return m.FindByID(context.Background(), id)
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter, page pagination.Params) ([]domain.Coupon, int64, error) {
	items := make([]domain.Coupon, 0, len(m.coupons))
	for _, c := range m.coupons {
		items = append(items, *c)
	}
	return items, int64(len(items)), nil
}

func (m *mockRepo) SetActive(_ context.Context, id uuid.UUID, active bool) error {
	c, ok := m.coupons[id]
	if !ok {
		return domain.ErrNotFound
	}
	c.IsActive = active
	return nil
}

func TestService_Create(t *testing.T) {
	svc := NewService(newMockRepo())
	c, err := svc.Create(context.Background(), CreateInput{
		Code:          "save10",
		DiscountType:  "percentage",
		DiscountValue: 10,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if c.Code != "SAVE10" {
		t.Errorf("Code = %q, want SAVE10", c.Code)
	}
}

func TestService_Create_CodeConflict(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), CreateInput{
		Code: "SAVE10", DiscountType: "percentage", DiscountValue: 10,
	})
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}

	_, err = svc.Create(context.Background(), CreateInput{
		Code: "save10", DiscountType: "percentage", DiscountValue: 5,
	})
	if err != domain.ErrCodeConflict {
		t.Errorf("expected code conflict, got %v", err)
	}
}

func TestService_Create_InvalidPercentage(t *testing.T) {
	svc := NewService(newMockRepo())
	_, err := svc.Create(context.Background(), CreateInput{
		Code: "BIG", DiscountType: "percentage", DiscountValue: 150,
	})
	if err != domain.ErrInvalidPercentage {
		t.Errorf("expected invalid percentage, got %v", err)
	}
}

func TestService_ActivateDeactivate(t *testing.T) {
	svc := NewService(newMockRepo())
	c, _ := svc.Create(context.Background(), CreateInput{
		Code: "OFF", DiscountType: "fixed_amount", DiscountValue: 5, IsActive: false,
	})

	active, err := svc.Activate(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("Activate() error = %v", err)
	}
	if !active.IsActive {
		t.Error("expected active coupon")
	}

	inactive, err := svc.Deactivate(context.Background(), c.ID)
	if err != nil {
		t.Fatalf("Deactivate() error = %v", err)
	}
	if inactive.IsActive {
		t.Error("expected inactive coupon")
	}
}
