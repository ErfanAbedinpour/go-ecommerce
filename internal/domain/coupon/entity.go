package coupon

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

// Coupon is the aggregate root for discount coupons.
type Coupon struct {
	ID             uuid.UUID
	Code           string
	DiscountType   DiscountType
	DiscountValue  float64
	MinOrderAmount float64
	MaxUsage       *int
	UsageCount     int
	ExpiresAt      *time.Time
	IsActive       bool
	Note           string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// IsExpired reports whether the coupon has passed its expiry date.
func (c *Coupon) IsExpired() bool {
	if c.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*c.ExpiresAt)
}

// IsExhausted reports whether usage limit has been reached.
func (c *Coupon) IsExhausted() bool {
	if c.MaxUsage == nil {
		return false
	}
	return c.UsageCount >= *c.MaxUsage
}

// NormalizeCode uppercases and trims a coupon code.
func NormalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
