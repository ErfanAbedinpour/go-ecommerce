package product

import (
	"testing"

	"github.com/google/uuid"
)

func TestProduct_EffectivePrice(t *testing.T) {
	sale := 79.99
	p := &Product{Price: 99.99, SalePrice: &sale}
	if got := p.EffectivePrice(); got != 79.99 {
		t.Errorf("EffectivePrice() = %v, want 79.99", got)
	}

	p2 := &Product{Price: 99.99}
	if got := p2.EffectivePrice(); got != 99.99 {
		t.Errorf("EffectivePrice() = %v, want 99.99", got)
	}
}

func TestInventory_IsLowStock(t *testing.T) {
	inv := Inventory{Quantity: 5, LowStockThreshold: 10}
	if !inv.IsLowStock() {
		t.Error("expected low stock")
	}
	if inv.IsOutOfStock() {
		t.Error("expected not out of stock")
	}
}

func TestParseStatus(t *testing.T) {
	s, err := ParseStatus("active")
	if err != nil || s != StatusActive {
		t.Errorf("ParseStatus() = %v, %v", s, err)
	}
	_, err = ParseStatus("invalid")
	if err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestInventory_IDs(t *testing.T) {
	inv := Inventory{ProductID: uuid.New(), Quantity: 0}
	if !inv.IsOutOfStock() {
		t.Error("expected out of stock")
	}
}
