package storefront

import (
	"context"

	"github.com/google/uuid"

	domainorder "app/internal/domain/order"
	"app/pkg/pagination"
)

// AccountOrderSummary is a customer-facing order list item.
type AccountOrderSummary struct {
	ID            uuid.UUID `json:"id"`
	OrderNumber   string    `json:"order_number"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	TotalToman    int64     `json:"total_toman"`
	ItemCount     int       `json:"item_count"`
	CreatedAt     string    `json:"created_at"`
}

// ListAccountOrders returns paginated orders for the authenticated customer.
func (s *Service) ListAccountOrders(ctx context.Context, userID uuid.UUID, page pagination.Params) (pagination.Paginated[AccountOrderSummary], error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return pagination.Paginated[AccountOrderSummary]{}, err
	}

	summaries, total, err := s.customers.ListOrders(ctx, customer.ID, page)
	if err != nil {
		return pagination.Paginated[AccountOrderSummary]{}, err
	}

	items := make([]AccountOrderSummary, len(summaries))
	for i, summary := range summaries {
		items[i] = AccountOrderSummary{
			ID:            summary.ID,
			OrderNumber:   summary.OrderNumber,
			Status:        summary.Status,
			PaymentStatus: summary.PaymentStatus,
			TotalToman:    toMoneyToman(summary.Total),
			ItemCount:     summary.ItemCount,
			CreatedAt:     summary.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// GetAccountOrder returns order detail for the authenticated customer.
func (s *Service) GetAccountOrder(ctx context.Context, userID, orderID uuid.UUID) (*domainorder.Order, error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.CustomerID != customer.ID {
		return nil, domainorder.ErrNotFound
	}
	return order, nil
}
