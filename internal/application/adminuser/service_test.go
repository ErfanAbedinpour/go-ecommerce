package adminuser

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domain "app/internal/domain/user"
	"app/pkg/pagination"
)

type mockHasher struct{}

func (mockHasher) Hash(password string) (string, error) {
	return "hash:" + password, nil
}

type mockUserRepo struct {
	users map[uuid.UUID]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[uuid.UUID]*domain.User)}
}

func (m *mockUserRepo) Create(_ context.Context, u *domain.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			cp := *u
			return &cp, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *mockUserRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.User, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *u
	return &cp, nil
}

func (m *mockUserRepo) List(_ context.Context, _ domain.ListFilter, _ pagination.Params) ([]domain.User, int64, error) {
	items := make([]domain.User, 0, len(m.users))
	for _, u := range m.users {
		items = append(items, *u)
	}
	return items, int64(len(items)), nil
}

func (m *mockUserRepo) Update(_ context.Context, u *domain.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) SoftDelete(_ context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserRepo) CountByRole(_ context.Context, role domain.Role) (int64, error) {
	var count int64
	for _, u := range m.users {
		if u.Role == role && u.IsActive {
			count++
		}
	}
	return count, nil
}

func (m *mockUserRepo) UpdateLastLogin(context.Context, uuid.UUID) error { return nil }
func (m *mockUserRepo) UpdatePassword(context.Context, uuid.UUID, string) error {
	return nil
}

type mockRefreshRepo struct{}

func (mockRefreshRepo) Create(context.Context, *domain.RefreshToken) error { return nil }
func (mockRefreshRepo) FindByHash(context.Context, string) (*domain.RefreshToken, error) {
	return nil, domain.ErrNotFound
}
func (mockRefreshRepo) RevokeFamily(context.Context, uuid.UUID) error { return nil }
func (mockRefreshRepo) Revoke(context.Context, uuid.UUID) error       { return nil }
func (mockRefreshRepo) RevokeAllByUser(context.Context, uuid.UUID) error {
	return nil
}

func TestService_Create(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo, mockRefreshRepo{}, mockHasher{})

	user, err := svc.Create(context.Background(), CreateInput{
		Email: "staff@shop.com", Password: "Secret1a", FirstName: "Staff", LastName: "User", Role: "admin",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if user.Role != domain.RoleAdmin {
		t.Errorf("role = %q, want admin", user.Role)
	}
}

func TestService_Delete_CannotDeleteSelf(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo, mockRefreshRepo{}, mockHasher{})
	id := uuid.New()
	repo.users[id] = &domain.User{ID: id, Role: domain.RoleAdmin, IsActive: true}

	err := svc.Delete(context.Background(), id, id)
	if err != domain.ErrCannotDeleteSelf {
		t.Errorf("expected cannot delete self, got %v", err)
	}
}

func TestService_Delete_LastAdmin(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewService(repo, mockRefreshRepo{}, mockHasher{})
	id := uuid.New()
	actor := uuid.New()
	repo.users[id] = &domain.User{ID: id, Role: domain.RoleAdmin, IsActive: true}

	err := svc.Delete(context.Background(), id, actor)
	if err != domain.ErrLastAdmin {
		t.Errorf("expected last admin error, got %v", err)
	}
}
