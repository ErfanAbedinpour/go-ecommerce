package coupon

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/coupon"
	"app/pkg/pagination"
)

// Service handles coupon management use cases.
type Service struct {
	repo domain.Repository
}

// NewService creates a new coupon Service.
func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// CreateInput holds data for creating a coupon.
type CreateInput struct {
	Code           string
	DiscountType   string
	DiscountValue  float64
	MinOrderAmount float64
	MaxUsage       *int
	ExpiresAt      *time.Time
	Note           string
	IsActive       bool
}

// UpdateInput holds partial update data for a coupon.
type UpdateInput struct {
	Code           *string
	DiscountType   *string
	DiscountValue  *float64
	MinOrderAmount *float64
	MaxUsage       *int
	ExpiresAt      *time.Time
	Note           *string
	IsActive       *bool
}

// Create creates a new coupon.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Coupon, error) {
	discountType, err := domain.ParseDiscountType(input.DiscountType)
	if err != nil {
		return nil, err
	}
	if err := validateDiscount(discountType, input.DiscountValue); err != nil {
		return nil, err
	}

	code := domain.NormalizeCode(input.Code)
	if err := s.ensureUniqueCode(ctx, code, uuid.Nil); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	coupon := &domain.Coupon{
		ID:             uuid.New(),
		Code:           code,
		DiscountType:   discountType,
		DiscountValue:  input.DiscountValue,
		MinOrderAmount: input.MinOrderAmount,
		MaxUsage:       input.MaxUsage,
		UsageCount:     0,
		ExpiresAt:      input.ExpiresAt,
		IsActive:       input.IsActive,
		Note:           input.Note,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Create(ctx, coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}

// Update updates an existing coupon.
func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*domain.Coupon, error) {
	coupon, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.Code != nil {
		code := domain.NormalizeCode(*input.Code)
		if err := s.ensureUniqueCode(ctx, code, id); err != nil {
			return nil, err
		}
		coupon.Code = code
	}
	if input.DiscountType != nil {
		discountType, err := domain.ParseDiscountType(*input.DiscountType)
		if err != nil {
			return nil, err
		}
		coupon.DiscountType = discountType
	}
	if input.DiscountValue != nil {
		coupon.DiscountValue = *input.DiscountValue
	}
	if input.MinOrderAmount != nil {
		coupon.MinOrderAmount = *input.MinOrderAmount
	}
	if input.MaxUsage != nil {
		coupon.MaxUsage = input.MaxUsage
	}
	if input.ExpiresAt != nil {
		coupon.ExpiresAt = input.ExpiresAt
	}
	if input.Note != nil {
		coupon.Note = *input.Note
	}
	if input.IsActive != nil {
		coupon.IsActive = *input.IsActive
	}

	if err := validateDiscount(coupon.DiscountType, coupon.DiscountValue); err != nil {
		return nil, err
	}

	coupon.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, coupon); err != nil {
		return nil, err
	}
	return coupon, nil
}

// Delete soft-deletes a coupon.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return err
	}
	return s.repo.SoftDelete(ctx, id)
}

// GetByID returns a coupon by ID.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Coupon, error) {
	return s.repo.FindByID(ctx, id)
}

// List returns a paginated coupon list.
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.Coupon], error) {
	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.Coupon]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// Activate enables a coupon.
func (s *Service) Activate(ctx context.Context, id uuid.UUID) (*domain.Coupon, error) {
	return s.setActive(ctx, id, true)
}

// Deactivate disables a coupon.
func (s *Service) Deactivate(ctx context.Context, id uuid.UUID) (*domain.Coupon, error) {
	return s.setActive(ctx, id, false)
}

func (s *Service) setActive(ctx context.Context, id uuid.UUID, active bool) (*domain.Coupon, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, err
	}
	if err := s.repo.SetActive(ctx, id, active); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *Service) ensureUniqueCode(ctx context.Context, code string, excludeID uuid.UUID) error {
	existing, err := s.repo.FindByCode(ctx, code)
	if err != nil && err != domain.ErrNotFound {
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return domain.ErrCodeConflict
	}
	return nil
}

func validateDiscount(discountType domain.DiscountType, value float64) error {
	if value <= 0 {
		return domain.ErrInvalidDiscount
	}
	if discountType == domain.DiscountTypePercentage && value > 100 {
		return domain.ErrInvalidPercentage
	}
	return nil
}
