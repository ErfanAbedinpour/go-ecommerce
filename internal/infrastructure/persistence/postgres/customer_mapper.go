package postgres

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/customer"
	"app/internal/infrastructure/persistence/models"
)

type customerRow struct {
	models.CustomerModel
	UserEmail     *string `gorm:"column:user_email"`
	UserFirstName *string `gorm:"column:user_first_name"`
	UserLastName  *string `gorm:"column:user_last_name"`
	UserPhone     *string `gorm:"column:user_phone"`
}

func customerSelectQuery(db *gorm.DB) *gorm.DB {
	return db.Table("customers").
		Select(`
			customers.*,
			users.email AS user_email,
			users.first_name AS user_first_name,
			users.last_name AS user_last_name,
			users.phone AS user_phone
		`).
		Joins("LEFT JOIN users ON users.id = customers.user_id AND users.deleted_at IS NULL")
}

func toCustomerDomain(m *models.CustomerModel) *customer.Customer {
	return toCustomerDomainFromRow(&customerRow{CustomerModel: *m})
}

func toCustomerDomainFromRow(row *customerRow) *customer.Customer {
	c := &customer.Customer{
		ID:          row.ID,
		UserID:      row.UserID,
		Type:        customer.CustomerType(row.Type),
		TotalOrders: row.TotalOrders,
		TotalSpent:  row.TotalSpent,
		LastOrderAt: row.LastOrderAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}

	if row.UserID != nil {
		if row.UserEmail != nil {
			c.Email = *row.UserEmail
		}
		if row.UserFirstName != nil {
			c.FirstName = *row.UserFirstName
		}
		if row.UserLastName != nil {
			c.LastName = *row.UserLastName
		}
		if row.UserPhone != nil {
			c.Phone = *row.UserPhone
		}
	} else {
		if row.Email != nil {
			c.Email = *row.Email
		}
		if row.FirstName != nil {
			c.FirstName = *row.FirstName
		}
		if row.LastName != nil {
			c.LastName = *row.LastName
		}
		if row.Phone != nil {
			c.Phone = *row.Phone
		}
	}

	return c
}

func toCustomersDomain(rows []customerRow) []customer.Customer {
	result := make([]customer.Customer, len(rows))
	for i := range rows {
		result[i] = *toCustomerDomainFromRow(&rows[i])
	}
	return result
}

func toAddressDomain(m *models.CustomerAddressModel) customer.Address {
	a := customer.Address{
		ID:         m.ID,
		CustomerID: m.CustomerID,
		Type:       customer.AddressType(m.Type),
		Street:     m.Street,
		City:       m.City,
		PostalCode: m.PostalCode,
		Country:    m.Country,
		IsDefault:  m.IsDefault,
	}
	if m.State != nil {
		a.State = *m.State
	}
	return a
}

func toCustomerModel(c *customer.Customer) *models.CustomerModel {
	m := &models.CustomerModel{
		ID:          c.ID,
		UserID:      c.UserID,
		Type:        c.Type.String(),
		TotalOrders: c.TotalOrders,
		TotalSpent:  c.TotalSpent,
		LastOrderAt: c.LastOrderAt,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}

	if c.UserID == nil {
		if c.Email != "" {
			email := c.Email
			m.Email = &email
		}
		if c.FirstName != "" {
			firstName := c.FirstName
			m.FirstName = &firstName
		}
		if c.LastName != "" {
			lastName := c.LastName
			m.LastName = &lastName
		}
		if c.Phone != "" {
			phone := c.Phone
			m.Phone = &phone
		}
	}

	return m
}

func toAddressesDomain(items []models.CustomerAddressModel) []customer.Address {
	result := make([]customer.Address, len(items))
	for i, m := range items {
		result[i] = toAddressDomain(&m)
	}
	return result
}

func toAddressModel(a *customer.Address, customerID uuid.UUID) models.CustomerAddressModel {
	m := models.CustomerAddressModel{
		ID:         a.ID,
		CustomerID: customerID,
		Type:       a.Type.String(),
		Street:     a.Street,
		City:       a.City,
		PostalCode: a.PostalCode,
		Country:    a.Country,
		IsDefault:  a.IsDefault,
	}
	if a.State != "" {
		state := a.State
		m.State = &state
	}
	return m
}

func userIdentityUpdates(c *customer.Customer) map[string]any {
	updates := map[string]any{
		"updated_at": c.UpdatedAt,
	}
	if c.Email != "" {
		updates["email"] = c.Email
	}
	if c.FirstName != "" {
		updates["first_name"] = c.FirstName
	}
	if c.LastName != "" {
		updates["last_name"] = c.LastName
	}
	if c.Phone != "" {
		updates["phone"] = c.Phone
	} else {
		updates["phone"] = nil
	}
	return updates
}
