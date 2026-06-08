package adminuser

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	appauth "app/internal/application/auth"
	domain "app/internal/domain/user"
	"app/pkg/pagination"
)

// PasswordHasher hashes plaintext passwords.
type PasswordHasher interface {
	Hash(password string) (string, error)
}

// Service handles admin user management use cases.
type Service struct {
	users         domain.Repository
	refreshTokens domain.RefreshTokenRepository
	hasher        PasswordHasher
}

// NewService creates a new admin user Service.
func NewService(users domain.Repository, refreshTokens domain.RefreshTokenRepository, hasher PasswordHasher) *Service {
	return &Service{users: users, refreshTokens: refreshTokens, hasher: hasher}
}

// CreateInput holds data for creating an admin user.
type CreateInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	Phone     string
	Role      string
	IsActive  *bool
}

// UpdateInput holds partial update data for an admin user.
type UpdateInput struct {
	Email     *string
	Password  *string
	FirstName *string
	LastName  *string
	Phone     *string
	Role      *string
	IsActive  *bool
}

// List returns a paginated admin user list.
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.User], error) {
	items, total, err := s.users.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.User]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// GetByID returns an admin user by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.FindByID(ctx, id)
}

// Create creates a new admin user account.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.User, error) {
	if err := appauth.ValidatePassword(input.Password); err != nil {
		return nil, err
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := s.users.FindByEmail(ctx, email); err == nil {
		return nil, domain.ErrEmailTaken
	} else if err != domain.ErrNotFound {
		return nil, err
	}

	role := domain.RoleAdmin
	if input.Role != "" {
		parsed, err := domain.ParseRole(input.Role)
		if err != nil {
			return nil, err
		}
		role = parsed
	}

	hash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	isActive := true
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	now := time.Now().UTC()
	u := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: hash,
		FirstName:    strings.TrimSpace(input.FirstName),
		LastName:     strings.TrimSpace(input.LastName),
		Phone:        strings.TrimSpace(input.Phone),
		Role:         role,
		IsActive:     isActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Update updates an existing admin user.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.User, error) {
	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Email != nil {
		email := strings.ToLower(strings.TrimSpace(*input.Email))
		existing, err := s.users.FindByEmail(ctx, email)
		if err == nil && existing.ID != id {
			return nil, domain.ErrEmailTaken
		}
		if err != nil && err != domain.ErrNotFound {
			return nil, err
		}
		u.Email = email
	}
	if input.Password != nil {
		if err := appauth.ValidatePassword(*input.Password); err != nil {
			return nil, err
		}
		hash, err := s.hasher.Hash(*input.Password)
		if err != nil {
			return nil, err
		}
		u.PasswordHash = hash
	}
	if input.FirstName != nil {
		u.FirstName = strings.TrimSpace(*input.FirstName)
	}
	if input.LastName != nil {
		u.LastName = strings.TrimSpace(*input.LastName)
	}
	if input.Phone != nil {
		u.Phone = strings.TrimSpace(*input.Phone)
	}
	if input.Role != nil {
		role, err := domain.ParseRole(*input.Role)
		if err != nil {
			return nil, err
		}
		u.Role = role
	}
	if input.IsActive != nil {
		u.IsActive = *input.IsActive
	}

	u.UpdatedAt = time.Now().UTC()
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

// Delete soft-deletes an admin user and revokes refresh tokens.
func (s *Service) Delete(ctx context.Context, id, actorID uuid.UUID) error {
	if id == actorID {
		return domain.ErrCannotDeleteSelf
	}

	u, err := s.users.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if u.Role == domain.RoleAdmin {
		count, err := s.users.CountByRole(ctx, domain.RoleAdmin)
		if err != nil {
			return err
		}
		if count <= 1 {
			return domain.ErrLastAdmin
		}
	}

	if err := s.refreshTokens.RevokeAllByUser(ctx, id); err != nil {
		return err
	}
	return s.users.SoftDelete(ctx, id)
}
