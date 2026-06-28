package contact

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/contact"
	"app/pkg/pagination"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

type SubmitInput struct {
	Name    string
	Email   string
	Phone   string
	Subject string
	Message string
	Source  string
}

func (s *Service) Submit(ctx context.Context, input SubmitInput) (*domain.Message, error) {
	src := domain.SourceHomepage
	switch input.Source {
	case string(domain.SourceAbout):
		src = domain.SourceAbout
	case string(domain.SourceContactPage):
		src = domain.SourceContactPage
	case "", string(domain.SourceHomepage):
		src = domain.SourceHomepage
	default:
		return nil, domain.ErrInvalidSource
	}

	msg := &domain.Message{
		ID:        uuid.New(),
		Name:      input.Name,
		Email:     input.Email,
		Phone:     input.Phone,
		Subject:   input.Subject,
		Message:   input.Message,
		Source:    src,
		Status:    domain.StatusUnread,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		return nil, err
	}

	return msg, nil
}

func (s *Service) FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.Message], error) {
	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.Message]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	var stat domain.Status
	switch status {
	case string(domain.StatusUnread):
		stat = domain.StatusUnread
	case string(domain.StatusRead):
		stat = domain.StatusRead
	case string(domain.StatusArchived):
		stat = domain.StatusArchived
	default:
		return domain.ErrInvalidStatus
	}

	return s.repo.UpdateStatus(ctx, id, stat)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
