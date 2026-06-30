package contact

import (
	"context"
	"testing"

	"github.com/google/uuid"

	domain "app/internal/domain/contact"
	"app/pkg/pagination"
)

type mockRepo struct {
	stats *domain.InboxStats
}

func (m *mockRepo) Create(context.Context, *domain.Message) error { return nil }
func (m *mockRepo) FindByID(context.Context, uuid.UUID) (*domain.Message, error) {
	return nil, nil
}
func (m *mockRepo) List(context.Context, domain.ListFilter, pagination.Params) ([]domain.Message, int64, error) {
	return nil, 0, nil
}
func (m *mockRepo) CountStats(context.Context) (*domain.InboxStats, error) {
	return m.stats, nil
}
func (m *mockRepo) UpdateStatus(context.Context, uuid.UUID, domain.Status) error { return nil }
func (m *mockRepo) Delete(context.Context, uuid.UUID) error                      { return nil }

func TestGetStats(t *testing.T) {
	svc := NewService(&mockRepo{
		stats: &domain.InboxStats{UnreadCount: 5, TotalCount: 142},
	})

	out, err := svc.GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if out.UnreadCount != 5 {
		t.Errorf("unread_count = %d, want 5", out.UnreadCount)
	}
	if out.TotalCount != 142 {
		t.Errorf("total_count = %d, want 142", out.TotalCount)
	}
}
