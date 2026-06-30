package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/order"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

var orderListSorts = map[string]string{
	"created_at":     "orders.created_at",
	"order_number":   "orders.order_number",
	"total":          "orders.total",
	"status":         "orders.status",
	"payment_status": "orders.payment_status",
}

// OrderRepository implements order.Repository using GORM.
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new OrderRepository.
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) FindByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	var m models.OrderModel
	err := r.db.WithContext(ctx).
		Preload("Items").
		Preload("StatusHistory", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Where("id = ?", id).
		First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, order.ErrNotFound
		}
		return nil, err
	}

	customer, err := r.loadCustomerSnapshot(ctx, m.CustomerID)
	if err != nil {
		return nil, err
	}

	return toOrderDomain(&m, customer)
}

func (r *OrderRepository) List(ctx context.Context, filter order.ListFilter, page pagination.Params) ([]order.ListItem, int64, error) {
	base := r.db.WithContext(ctx).
		Table("orders").
		Joins("INNER JOIN customers ON customers.id = orders.customer_id")
	base = r.applyListFilters(base, filter)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type listRow struct {
		ID            uuid.UUID `gorm:"column:id"`
		OrderNumber   string    `gorm:"column:order_number"`
		Status        string    `gorm:"column:status"`
		PaymentStatus string    `gorm:"column:payment_status"`
		Total         float64   `gorm:"column:total"`
		CreatedAt     time.Time `gorm:"column:created_at"`
		CustomerID    uuid.UUID `gorm:"column:customer_id"`
		CustomerName  string    `gorm:"column:customer_name"`
		CustomerEmail string    `gorm:"column:customer_email"`
		ItemCount     int64     `gorm:"column:item_count"`
	}

	var rows []listRow
	err := base.
		Select(`
			orders.id,
			orders.order_number,
			orders.status,
			orders.payment_status,
			orders.total,
			orders.created_at,
			orders.customer_id,
			customers.first_name || ' ' || customers.last_name AS customer_name,
			customers.email AS customer_email,
			(SELECT COUNT(*) FROM order_items WHERE order_items.order_id = orders.id) AS item_count
		`).
		Order(r.orderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]order.ListItem, len(rows))
	for i, row := range rows {
		result[i] = order.ListItem{
			Summary: order.Summary{
				ID:            row.ID,
				OrderNumber:   row.OrderNumber,
				Status:        row.Status,
				PaymentStatus: row.PaymentStatus,
				Total:         row.Total,
				ItemCount:     int(row.ItemCount),
				CreatedAt:     row.CreatedAt,
			},
			CustomerID:    row.CustomerID,
			CustomerName:  row.CustomerName,
			CustomerEmail: row.CustomerEmail,
		}
	}
	return result, total, nil
}

func (r *OrderRepository) Create(ctx context.Context, o *order.Order) error {
	m, err := toOrderModel(o)
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}

		for _, item := range o.Items {
			itemModel := toOrderItemModel(&item)
			if err := tx.Create(itemModel).Error; err != nil {
				return err
			}

			result := tx.Exec(
				"UPDATE inventories SET quantity = quantity - ?, updated_at = NOW() WHERE product_id = ? AND quantity >= ?",
				item.Quantity, item.ProductID, item.Quantity,
			)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return order.ErrInsufficientStock
			}
		}
		return nil
	})
}

func (r *OrderRepository) UpdateNotes(ctx context.Context, id uuid.UUID, notes string, updatedAt time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&models.OrderModel{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"notes":      notes,
			"updated_at": updatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return order.ErrNotFound
	}
	return nil
}

func (r *OrderRepository) NextOrderNumber(ctx context.Context) (string, error) {
	var maxNum int
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(MAX(CAST(NULLIF(regexp_replace(order_number, '[^0-9]', '', 'g'), '') AS INTEGER)), 0)
		FROM orders
	`).Scan(&maxNum).Error
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("ORD-%06d", maxNum+1), nil
}

func (r *OrderRepository) IncrementCouponUsage(ctx context.Context, couponID uuid.UUID) error {
	result := r.db.WithContext(ctx).Exec(
		"UPDATE coupons SET usage_count = usage_count + 1, updated_at = NOW() WHERE id = ?",
		couponID,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("coupon not found")
	}
	return nil
}

func (r *OrderRepository) Update(ctx context.Context, o *order.Order) error {
	result := r.db.WithContext(ctx).
		Model(&models.OrderModel{}).
		Where("id = ?", o.ID).
		Updates(map[string]any{
			"status":         o.Status.String(),
			"payment_status": o.PaymentStatus.String(),
			"updated_at":     o.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return order.ErrNotFound
	}
	return nil
}

func (r *OrderRepository) AddStatusHistory(ctx context.Context, entry *order.StatusHistory) error {
	m := toStatusHistoryModel(entry)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *OrderRepository) RestoreInventory(ctx context.Context, items []order.Item) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			result := tx.Exec(
				"UPDATE inventories SET quantity = quantity + ?, updated_at = NOW() WHERE product_id = ?",
				item.Quantity, item.ProductID,
			)
			if result.Error != nil {
				return result.Error
			}
		}
		return nil
	})
}

func (r *OrderRepository) loadCustomerSnapshot(ctx context.Context, customerID uuid.UUID) (*order.CustomerSnapshot, error) {
	var c models.CustomerModel
	err := r.db.WithContext(ctx).
		Select("id", "email", "first_name", "last_name", "phone").
		Where("id = ?", customerID).
		First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	snap := &order.CustomerSnapshot{
		ID:        c.ID,
		Email:     c.Email,
		FirstName: c.FirstName,
		LastName:  c.LastName,
	}
	if c.Phone != nil {
		snap.Phone = *c.Phone
	}
	return snap, nil
}

func (r *OrderRepository) applyListFilters(query *gorm.DB, filter order.ListFilter) *gorm.DB {
	if filter.Status != "" {
		query = query.Where("orders.status = ?", filter.Status)
	}
	if filter.PaymentStatus != "" {
		query = query.Where("orders.payment_status = ?", filter.PaymentStatus)
	}
	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"LOWER(orders.order_number) LIKE ? OR LOWER(customers.email) LIKE ? OR LOWER(customers.first_name) LIKE ? OR LOWER(customers.last_name) LIKE ? OR LOWER(CONCAT(customers.first_name, ' ', customers.last_name)) LIKE ?",
			pattern, pattern, pattern, pattern, pattern,
		)
	}
	if filter.From != nil {
		query = query.Where("orders.created_at >= ?", filter.From.UTC())
	}
	if filter.To != nil {
		to := filter.To.UTC().Add(24*time.Hour - time.Nanosecond)
		query = query.Where("orders.created_at <= ?", to)
	}
	return query
}

func (r *OrderRepository) orderClause(page pagination.Params) string {
	column, ok := orderListSorts[page.Sort]
	if !ok {
		column = orderListSorts["created_at"]
	}
	orderDir := "DESC"
	if strings.EqualFold(page.Order, "asc") {
		orderDir = "ASC"
	}
	return fmt.Sprintf("%s %s", column, orderDir)
}

func (r *OrderRepository) CountByStatus(ctx context.Context, status order.Status) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.OrderModel{}).
		Where("status = ?", status.String()).
		Count(&count).Error
	return count, err
}

var _ order.Repository = (*OrderRepository)(nil)
