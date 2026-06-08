package postgres

import (
	"app/internal/domain/customer"
	"app/internal/infrastructure/persistence/models"
)

func toCustomerDomain(m *models.CustomerModel) *customer.Customer {
	c := &customer.Customer{
		ID:          m.ID,
		Email:       m.Email,
		FirstName:   m.FirstName,
		LastName:    m.LastName,
		Type:        customer.CustomerType(m.Type),
		TotalOrders: m.TotalOrders,
		TotalSpent:  m.TotalSpent,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.Phone != nil {
		c.Phone = *m.Phone
	}
	return c
}

func toCustomersDomain(items []models.CustomerModel) []customer.Customer {
	result := make([]customer.Customer, len(items))
	for i, m := range items {
		result[i] = *toCustomerDomain(&m)
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
		Email:       c.Email,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		Type:        c.Type.String(),
		TotalOrders: c.TotalOrders,
		TotalSpent:  c.TotalSpent,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if c.Phone != "" {
		m.Phone = &c.Phone
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
