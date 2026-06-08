package coupon

import "math"

// CalculateDiscount returns the discount amount for a given order subtotal.
func (c *Coupon) CalculateDiscount(subtotal float64) (float64, error) {
	if !c.IsActive {
		return 0, ErrNotApplicable
	}
	if c.IsExpired() {
		return 0, ErrExpired
	}
	if c.IsExhausted() {
		return 0, ErrExhausted
	}
	if subtotal < c.MinOrderAmount {
		return 0, ErrMinOrderNotMet
	}

	var amount float64
	switch c.DiscountType {
	case DiscountTypePercentage:
		amount = subtotal * (c.DiscountValue / 100)
	case DiscountTypeFixedAmount:
		amount = c.DiscountValue
	default:
		return 0, ErrInvalidDiscount
	}

	if amount > subtotal {
		amount = subtotal
	}
	return roundMoney(amount), nil
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
