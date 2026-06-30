package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	domain "app/internal/domain/cart"
)

const cartTTL = 30 * 24 * time.Hour

// CartRepository stores carts in Redis.
type CartRepository struct {
	cache *RedisCache
}

// NewCartRepository creates a Redis-backed cart repository.
func NewCartRepository(cache *RedisCache) *CartRepository {
	return &CartRepository{cache: cache}
}

func (r *CartRepository) Get(ctx context.Context, owner domain.Owner) (*domain.Cart, error) {
	data, err := r.cache.Get(ctx, cartKey(owner))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return &domain.Cart{Items: []domain.Item{}}, nil
	}

	var cart domain.Cart
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil, err
	}
	if cart.Items == nil {
		cart.Items = []domain.Item{}
	}
	return &cart, nil
}

func (r *CartRepository) Save(ctx context.Context, owner domain.Owner, cart *domain.Cart) error {
	if cart == nil {
		cart = &domain.Cart{Items: []domain.Item{}}
	}
	cart.UpdatedAt = time.Now().UTC()
	data, err := json.Marshal(cart)
	if err != nil {
		return err
	}
	return r.cache.Set(ctx, cartKey(owner), data, cartTTL)
}

func (r *CartRepository) Delete(ctx context.Context, owner domain.Owner) error {
	return r.cache.Delete(ctx, cartKey(owner))
}

func cartKey(owner domain.Owner) string {
	if owner.UserID != nil {
		return fmt.Sprintf("cart:user:%s", owner.UserID.String())
	}
	return fmt.Sprintf("cart:guest:%s", owner.GuestToken)
}

var _ domain.Repository = (*CartRepository)(nil)
