package storefront

import (
	"context"
	"sync"

	domaincart "app/internal/domain/cart"
)

type memoryCartRepo struct {
	mu   sync.Mutex
	data map[string]*domaincart.Cart
}

func newMemoryCartRepo() *memoryCartRepo {
	return &memoryCartRepo{data: make(map[string]*domaincart.Cart)}
}

func (m *memoryCartRepo) key(owner domaincart.Owner) string {
	if owner.UserID != nil {
		return "user:" + owner.UserID.String()
	}
	return "guest:" + owner.GuestToken
}

func (m *memoryCartRepo) Get(_ context.Context, owner domaincart.Owner) (*domaincart.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cart, ok := m.data[m.key(owner)]
	if !ok {
		return &domaincart.Cart{Items: []domaincart.Item{}}, nil
	}
	cp := *cart
	if cp.Items == nil {
		cp.Items = []domaincart.Item{}
	}
	return &cp, nil
}

func (m *memoryCartRepo) Save(_ context.Context, owner domaincart.Owner, cart *domaincart.Cart) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *cart
	m.data[m.key(owner)] = &cp
	return nil
}

func (m *memoryCartRepo) Delete(_ context.Context, owner domaincart.Owner) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, m.key(owner))
	return nil
}
