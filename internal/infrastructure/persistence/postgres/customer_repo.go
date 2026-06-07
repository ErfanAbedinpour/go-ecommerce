package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

var allowedCustomerSorts = map[string]string{
	"created_at":   "customers.created_at",
	"email":        "customers.email",
	"first_name":   "customers.first_name",
	"last_name":    "customers.last_name",
	"total_orders": "customers.total_orders",
	"total_spent":  "customers.total_spent",
	"updated_at":   "customers.updated_at",
}

var allowedOrderSorts = map[string]string{
	"created_at":   "orders.created_at",
	"order_number": "orders.order_number",
	"total":        "orders.total",
	"status":       "orders.status",
}

// CustomerRepository implements customer.Repository using GORM.
type CustomerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new CustomerRepository.
func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) FindByID(ctx context.Context, id uuid.UUID) (*customer.Customer, error) {
	var m models.CustomerModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customer.ErrNotFound
		}
		return nil, err
	}
	return toCustomerDomain(&m), nil
}

func (r *CustomerRepository) List(ctx context.Context, filter customer.ListFilter, page pagination.Params) ([]customer.Customer, int64, error) {
	query := r.applyListFilters(r.db.WithContext(ctx).Model(&models.CustomerModel{}), filter)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.CustomerModel
	err := query.
		Order(r.customerOrderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&items).Error
	if err != nil {
		return nil, 0, err
	}
	return toCustomersDomain(items), total, nil
}

func (r *CustomerRepository) ListAddresses(ctx context.Context, customerID uuid.UUID) ([]customer.Address, error) {
	var items []models.CustomerAddressModel
	err := r.db.WithContext(ctx).
		Where("customer_id = ?", customerID).
		Order("is_default DESC, type ASC").
		Find(&items).Error
	if err != nil {
		return nil, err
	}
	return toAddressesDomain(items), nil
}

func (r *CustomerRepository) ListOrders(ctx context.Context, customerID uuid.UUID, page pagination.Params) ([]domainorder.Summary, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.OrderModel{}).
		Where("customer_id = ?", customerID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type orderRow struct {
		models.OrderModel
		ItemCount int64 `gorm:"column:item_count"`
	}

	var rows []orderRow
	err := r.db.WithContext(ctx).
		Table("orders").
		Select("orders.*, COUNT(order_items.id) AS item_count").
		Joins("LEFT JOIN order_items ON order_items.order_id = orders.id").
		Where("orders.customer_id = ?", customerID).
		Group("orders.id").
		Order(r.orderOrderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	result := make([]domainorder.Summary, len(rows))
	for i, row := range rows {
		result[i] = domainorder.Summary{
			ID:            row.ID,
			OrderNumber:   row.OrderNumber,
			Status:        row.Status,
			PaymentStatus: row.PaymentStatus,
			Total:         row.Total,
			ItemCount:     int(row.ItemCount),
			CreatedAt:     row.CreatedAt,
		}
	}
	return result, total, nil
}

func (r *CustomerRepository) applyListFilters(query *gorm.DB, filter customer.ListFilter) *gorm.DB {
	if filter.Query != "" {
		pattern := "%" + strings.ToLower(filter.Query) + "%"
		query = query.Where(
			"LOWER(email) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(CONCAT(first_name, ' ', last_name)) LIKE ?",
			pattern, pattern, pattern, pattern,
		)
	}
	if filter.Type != nil {
		query = query.Where("type = ?", filter.Type.String())
	}
	return query
}

func (r *CustomerRepository) customerOrderClause(page pagination.Params) string {
	column, ok := allowedCustomerSorts[page.Sort]
	if !ok {
		column = allowedCustomerSorts["created_at"]
	}
	order := "DESC"
	if strings.EqualFold(page.Order, "asc") {
		order = "ASC"
	}
	return fmt.Sprintf("%s %s", column, order)
}

func (r *CustomerRepository) orderOrderClause(page pagination.Params) string {
	column, ok := allowedOrderSorts[page.Sort]
	if !ok {
		column = allowedOrderSorts["created_at"]
	}
	order := "DESC"
	if strings.EqualFold(page.Order, "asc") {
		order = "ASC"
	}
	return fmt.Sprintf("%s %s", column, order)
}
