package coupon

import (
	"testing"
	"time"
)

func TestCoupon_CalculateDiscount(t *testing.T) {
	tests := []struct {
		name      string
		coupon    Coupon
		subtotal  float64
		want      float64
		wantError error
	}{
		{
			name: "percentage discount",
			coupon: Coupon{
				IsActive: true, DiscountType: DiscountTypePercentage, DiscountValue: 10,
			},
			subtotal: 100,
			want:     10,
		},
		{
			name: "fixed amount capped at subtotal",
			coupon: Coupon{
				IsActive: true, DiscountType: DiscountTypeFixedAmount, DiscountValue: 150,
			},
			subtotal: 100,
			want:     100,
		},
		{
			name: "inactive coupon",
			coupon: Coupon{
				IsActive: false, DiscountType: DiscountTypePercentage, DiscountValue: 10,
			},
			subtotal:  100,
			wantError: ErrNotApplicable,
		},
		{
			name: "below minimum order",
			coupon: Coupon{
				IsActive: true, DiscountType: DiscountTypeFixedAmount, DiscountValue: 5, MinOrderAmount: 50,
			},
			subtotal:  40,
			wantError: ErrMinOrderNotMet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.coupon.CalculateDiscount(tt.subtotal)
			if tt.wantError != nil {
				if err != tt.wantError {
					t.Fatalf("CalculateDiscount() error = %v, want %v", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("CalculateDiscount() unexpected error = %v", err)
			}
			if got != tt.want {
				t.Errorf("CalculateDiscount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCoupon_CalculateDiscount_Expired(t *testing.T) {
	expired := time.Now().UTC().Add(-time.Hour)
	c := Coupon{
		IsActive: true, DiscountType: DiscountTypePercentage, DiscountValue: 10, ExpiresAt: &expired,
	}
	_, err := c.CalculateDiscount(100)
	if err != ErrExpired {
		t.Errorf("expected expired error, got %v", err)
	}
}
