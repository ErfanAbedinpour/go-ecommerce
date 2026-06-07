package postgres

import (
	"encoding/json"
	"fmt"

	"app/internal/domain/order"
	"app/internal/infrastructure/persistence/models"
)

func toOrderDomain(m *models.OrderModel, customer *order.CustomerSnapshot) (*order.Order, error) {
	o := &order.Order{
		ID:             m.ID,
		OrderNumber:    m.OrderNumber,
		CustomerID:     m.CustomerID,
		CouponID:       m.CouponID,
		Status:         order.Status(m.Status),
		PaymentStatus:  order.PaymentStatus(m.PaymentStatus),
		Subtotal:       m.Subtotal,
		DiscountAmount: m.DiscountAmount,
		ShippingAmount: m.ShippingAmount,
		TaxAmount:      m.TaxAmount,
		Total:          m.Total,
		Customer:       customer,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if m.Notes != nil {
		o.Notes = *m.Notes
	}

	billing, err := parseAddressJSON(m.BillingAddress)
	if err != nil {
		return nil, fmt.Errorf("billing address: %w", err)
	}
	o.BillingAddress = billing

	shipping, err := parseAddressJSON(m.ShippingAddress)
	if err != nil {
		return nil, fmt.Errorf("shipping address: %w", err)
	}
	o.ShippingAddress = shipping

	o.Items = make([]order.Item, len(m.Items))
	for i, item := range m.Items {
		o.Items[i] = toOrderItemDomain(&item)
	}

	o.Timeline = make([]order.StatusHistory, len(m.StatusHistory))
	for i, h := range m.StatusHistory {
		o.Timeline[i] = toStatusHistoryDomain(&h)
	}

	return o, nil
}

func toOrderItemDomain(m *models.OrderItemModel) order.Item {
	return order.Item{
		ID:          m.ID,
		OrderID:     m.OrderID,
		ProductID:   m.ProductID,
		ProductName: m.ProductName,
		ProductSKU:  m.ProductSKU,
		Quantity:    m.Quantity,
		UnitPrice:   m.UnitPrice,
		TotalPrice:  m.TotalPrice,
	}
}

func toStatusHistoryDomain(m *models.OrderStatusHistoryModel) order.StatusHistory {
	h := order.StatusHistory{
		ID:        m.ID,
		OrderID:   m.OrderID,
		ToStatus:  order.Status(m.ToStatus),
		ChangedBy: m.ChangedBy,
		CreatedAt: m.CreatedAt,
	}
	if m.FromStatus != nil {
		s := order.Status(*m.FromStatus)
		h.FromStatus = &s
	}
	if m.Note != nil {
		h.Note = *m.Note
	}
	return h
}

func toStatusHistoryModel(h *order.StatusHistory) *models.OrderStatusHistoryModel {
	m := &models.OrderStatusHistoryModel{
		ID:        h.ID,
		OrderID:   h.OrderID,
		ToStatus:  h.ToStatus.String(),
		ChangedBy: h.ChangedBy,
		CreatedAt: h.CreatedAt,
	}
	if h.FromStatus != nil {
		s := h.FromStatus.String()
		m.FromStatus = &s
	}
	if h.Note != "" {
		m.Note = &h.Note
	}
	return m
}

func parseAddressJSON(raw json.RawMessage) (order.Address, error) {
	var addr order.Address
	if len(raw) == 0 {
		return addr, nil
	}
	if err := json.Unmarshal(raw, &addr); err != nil {
		return addr, err
	}
	return addr, nil
}
