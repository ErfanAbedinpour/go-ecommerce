package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

var allowedCustomerSorts = map[string]string{
	"created_at":   "customers.created_at",
	"email":        "COALESCE(users.email, customers.email)",
	"first_name":   "COALESCE(users.first_name, customers.first_name)",
	"last_name":    "COALESCE(users.last_name, customers.last_name)",
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

func (r *CustomerRepository) Create(ctx context.Context, c *customer.Customer) error {
	m := toCustomerModel(c)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *CustomerRepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*customer.Customer, error) {
	var row customerRow
	err := customerSelectQuery(r.db.WithContext(ctx)).
		Where("customers.user_id = ?", userID).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customer.ErrNotFound
		}
		return nil, err
	}
	return toCustomerDomainFromRow(&row), nil
}

func (r *CustomerRepository) FindByID(ctx context.Context, id uuid.UUID) (*customer.Customer, error) {
	var row customerRow
	err := customerSelectQuery(r.db.WithContext(ctx)).
		Where("customers.id = ?", id).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customer.ErrNotFound
		}
		return nil, err
	}
	return toCustomerDomainFromRow(&row), nil
}

func (r *CustomerRepository) List(ctx context.Context, filter customer.ListFilter, page pagination.Params) ([]customer.Customer, int64, error) {
	base := customerSelectQuery(r.db.WithContext(ctx))
	base = r.applyListFilters(base, filter)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []customerRow
	err := base.
		Order(r.customerOrderClause(page)).
		Offset(page.Offset()).
		Limit(page.Limit()).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return toCustomersDomain(rows), total, nil
}

func (r *CustomerRepository) FindByEmail(ctx context.Context, email string) (*customer.Customer, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	var row customerRow
	err := customerSelectQuery(r.db.WithContext(ctx)).
		Where(
			"(customers.type = 'guest' AND LOWER(customers.email) = ?) OR (customers.user_id IS NOT NULL AND LOWER(users.email) = ?)",
			normalized, normalized,
		).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customer.ErrNotFound
		}
		return nil, err
	}
	return toCustomerDomainFromRow(&row), nil
}

func (r *CustomerRepository) FindGuestByEmail(ctx context.Context, email string) (*customer.Customer, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" {
		return nil, customer.ErrNotFound
	}
	var row customerRow
	err := customerSelectQuery(r.db.WithContext(ctx)).
		Where("customers.type = ? AND LOWER(customers.email) = ?", customer.TypeGuest.String(), normalized).
		Order("customers.last_order_at DESC NULLS LAST, customers.created_at DESC").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customer.ErrNotFound
		}
		return nil, err
	}
	return toCustomerDomainFromRow(&row), nil
}

func (r *CustomerRepository) FindGuestByPhone(ctx context.Context, phone string) (*customer.Customer, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, customer.ErrNotFound
	}
	var row customerRow
	err := customerSelectQuery(r.db.WithContext(ctx)).
		Where("customers.type = ? AND TRIM(customers.phone) = ?", customer.TypeGuest.String(), phone).
		Order("customers.last_order_at DESC NULLS LAST, customers.created_at DESC").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customer.ErrNotFound
		}
		return nil, err
	}
	return toCustomerDomainFromRow(&row), nil
}

func (r *CustomerRepository) FindRegisteredByPhone(ctx context.Context, phone string) (*customer.Customer, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return nil, customer.ErrNotFound
	}
	var row customerRow
	err := customerSelectQuery(r.db.WithContext(ctx)).
		Where("customers.user_id IS NOT NULL AND users.phone IS NOT NULL AND TRIM(users.phone) = ?", phone).
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customer.ErrNotFound
		}
		return nil, err
	}
	return toCustomerDomainFromRow(&row), nil
}

func (r *CustomerRepository) Update(ctx context.Context, c *customer.Customer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if c.UserID != nil {
			result := tx.Model(&models.UserModel{}).
				Where("id = ? AND deleted_at IS NULL", *c.UserID).
				Updates(userIdentityUpdates(c))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return customer.ErrNotFound
			}
		}

		m := toCustomerModel(c)
		updates := map[string]any{
			"user_id":       m.UserID,
			"type":          m.Type,
			"total_orders":  m.TotalOrders,
			"total_spent":   m.TotalSpent,
			"last_order_at": m.LastOrderAt,
			"updated_at":    m.UpdatedAt,
		}
		if c.UserID != nil {
			updates["email"] = nil
			updates["first_name"] = nil
			updates["last_name"] = nil
			updates["phone"] = nil
		} else {
			updates["email"] = m.Email
			updates["first_name"] = m.FirstName
			updates["last_name"] = m.LastName
			updates["phone"] = m.Phone
		}

		result := tx.Model(&models.CustomerModel{}).
			Where("id = ?", c.ID).
			Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return customer.ErrNotFound
		}
		return nil
	})
}

func (r *CustomerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("customer_id = ?", id).Delete(&models.CustomerAddressModel{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&models.CustomerModel{}, "id = ?", id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return customer.ErrNotFound
		}
		return nil
	})
}

func (r *CustomerRepository) HasOrders(ctx context.Context, customerID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.OrderModel{}).
		Where("customer_id = ?", customerID).
		Count(&count).Error
	return count > 0, err
}

func (r *CustomerRepository) GetLastOrderAt(ctx context.Context, customerID uuid.UUID) (*time.Time, error) {
	var customerModel models.CustomerModel
	err := r.db.WithContext(ctx).
		Select("last_order_at").
		Where("id = ?", customerID).
		First(&customerModel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customer.ErrNotFound
		}
		return nil, err
	}
	if customerModel.LastOrderAt != nil {
		return customerModel.LastOrderAt, nil
	}

	var createdAt time.Time
	err = r.db.WithContext(ctx).
		Model(&models.OrderModel{}).
		Select("created_at").
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Limit(1).
		Scan(&createdAt).Error
	if err != nil {
		return nil, err
	}
	if createdAt.IsZero() {
		return nil, nil
	}
	return &createdAt, nil
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

func (r *CustomerRepository) ReplaceAddresses(ctx context.Context, customerID uuid.UUID, addresses []customer.Address) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("customer_id = ?", customerID).Delete(&models.CustomerAddressModel{}).Error; err != nil {
			return err
		}
		for _, address := range addresses {
			if address.ID == uuid.Nil {
				address.ID = uuid.New()
			}
			address.CustomerID = customerID
			m := toAddressModel(&address, customerID)
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
		return nil
	})
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
			`LOWER(COALESCE(users.email, customers.email)) LIKE ?
			 OR LOWER(COALESCE(users.first_name, customers.first_name)) LIKE ?
			 OR LOWER(COALESCE(users.last_name, customers.last_name)) LIKE ?
			 OR LOWER(CONCAT(COALESCE(users.first_name, customers.first_name), ' ', COALESCE(users.last_name, customers.last_name))) LIKE ?`,
			pattern, pattern, pattern, pattern,
		)
	}
	if filter.Type != nil {
		query = query.Where("customers.type = ?", filter.Type.String())
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

func (r *CustomerRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.CustomerModel{}).Count(&count).Error
	return count, err
}

func (r *CustomerRepository) RecordOrderPlaced(ctx context.Context, customerID uuid.UUID, orderTotal float64, orderedAt time.Time) error {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE customers
		SET total_orders = total_orders + 1,
		    total_spent = total_spent + ?,
		    last_order_at = GREATEST(COALESCE(last_order_at, ?), ?),
		    updated_at = NOW()
		WHERE id = ?
	`, orderTotal, orderedAt, orderedAt, customerID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return customer.ErrNotFound
	}
	return nil
}

func (r *CustomerRepository) RecordOrderCancelled(ctx context.Context, customerID uuid.UUID, orderTotal float64) error {
	result := r.db.WithContext(ctx).Exec(`
		UPDATE customers
		SET total_orders = GREATEST(total_orders - 1, 0),
		    total_spent = GREATEST(total_spent - ?, 0),
		    updated_at = NOW()
		WHERE id = ?
	`, orderTotal, customerID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return customer.ErrNotFound
	}
	return nil
}

var _ customer.Repository = (*CustomerRepository)(nil)
