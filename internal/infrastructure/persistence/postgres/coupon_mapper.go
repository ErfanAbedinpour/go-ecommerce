package postgres

import (
	"app/internal/domain/coupon"
	"app/internal/infrastructure/persistence/models"
)

func toCouponDomain(m *models.CouponModel) *coupon.Coupon {
	c := &coupon.Coupon{
		ID:             m.ID,
		Code:           m.Code,
		DiscountType:   coupon.DiscountType(m.DiscountType),
		DiscountValue:  m.DiscountValue,
		MinOrderAmount: m.MinOrderAmount,
		MaxUsage:       m.MaxUsage,
		UsageCount:     m.UsageCount,
		ExpiresAt:      m.ExpiresAt,
		IsActive:       m.IsActive,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	if m.Note != nil {
		c.Note = *m.Note
	}
	return c
}

func toCouponsDomain(items []models.CouponModel) []coupon.Coupon {
	result := make([]coupon.Coupon, len(items))
	for i, m := range items {
		result[i] = *toCouponDomain(&m)
	}
	return result
}

func toCouponModel(c *coupon.Coupon) *models.CouponModel {
	m := &models.CouponModel{
		ID:             c.ID,
		Code:           c.Code,
		DiscountType:   c.DiscountType.String(),
		DiscountValue:  c.DiscountValue,
		MinOrderAmount: c.MinOrderAmount,
		MaxUsage:       c.MaxUsage,
		UsageCount:     c.UsageCount,
		ExpiresAt:      c.ExpiresAt,
		IsActive:       c.IsActive,
		CreatedAt:      c.CreatedAt,
		UpdatedAt:      c.UpdatedAt,
	}
	if c.Note != "" {
		m.Note = &c.Note
	}
	return m
}
