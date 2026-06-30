package cart

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/cart"
	domainproduct "app/internal/domain/product"
)

// Service manages server-side shopping carts.
type Service struct {
	repo     domain.Repository
	products domainproduct.Repository
}

// NewService creates a cart Service.
func NewService(repo domain.Repository, products domainproduct.Repository) *Service {
	return &Service{repo: repo, products: products}
}

// AddItemInput holds data for adding a product to the cart.
type AddItemInput struct {
	ProductID uuid.UUID
	SkuID     *uuid.UUID
	Quantity  int
}

// UpdateItemInput holds data for updating a cart line quantity.
type UpdateItemInput struct {
	ProductID uuid.UUID
	SkuID     *uuid.UUID
	Quantity  int
}

// RemoveItemInput identifies a cart line to remove.
type RemoveItemInput struct {
	ProductID uuid.UUID
	SkuID     *uuid.UUID
}

// AddItem adds or increments a product in the cart.
func (s *Service) AddItem(ctx context.Context, owner domain.Owner, input AddItemInput) (*domain.Cart, error) {
	if input.Quantity <= 0 {
		input.Quantity = 1
	}
	if err := s.validateProduct(ctx, input.ProductID, input.SkuID); err != nil {
		return nil, err
	}

	cart, err := s.repo.Get(ctx, owner)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	found := false
	for i, item := range cart.Items {
		if sameLine(item, input.ProductID, input.SkuID) {
			cart.Items[i].Quantity += input.Quantity
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, domain.Item{
			ProductID: input.ProductID,
			SkuID:     input.SkuID,
			Quantity:  input.Quantity,
			AddedAt:   now,
		})
	}

	if err := s.repo.Save(ctx, owner, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

// SetItemQuantity sets the quantity for a cart line (removes when quantity <= 0).
func (s *Service) SetItemQuantity(ctx context.Context, owner domain.Owner, input UpdateItemInput) (*domain.Cart, error) {
	cart, err := s.repo.Get(ctx, owner)
	if err != nil {
		return nil, err
	}

	if input.Quantity <= 0 {
		return s.removeLine(ctx, owner, cart, input.ProductID, input.SkuID)
	}

	if err := s.validateProduct(ctx, input.ProductID, input.SkuID); err != nil {
		return nil, err
	}

	found := false
	for i, item := range cart.Items {
		if sameLine(item, input.ProductID, input.SkuID) {
			cart.Items[i].Quantity = input.Quantity
			found = true
			break
		}
	}
	if !found {
		cart.Items = append(cart.Items, domain.Item{
			ProductID: input.ProductID,
			SkuID:     input.SkuID,
			Quantity:  input.Quantity,
			AddedAt:   time.Now().UTC(),
		})
	}

	if err := s.repo.Save(ctx, owner, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

// RemoveItem deletes a line from the cart.
func (s *Service) RemoveItem(ctx context.Context, owner domain.Owner, input RemoveItemInput) (*domain.Cart, error) {
	cart, err := s.repo.Get(ctx, owner)
	if err != nil {
		return nil, err
	}
	return s.removeLine(ctx, owner, cart, input.ProductID, input.SkuID)
}

// Get returns the raw cart for an owner.
func (s *Service) Get(ctx context.Context, owner domain.Owner) (*domain.Cart, error) {
	return s.repo.Get(ctx, owner)
}

// Clear removes all items from the cart.
func (s *Service) Clear(ctx context.Context, owner domain.Owner) error {
	return s.repo.Delete(ctx, owner)
}

// MergeGuestIntoUser moves guest cart lines into the authenticated user's cart.
func (s *Service) MergeGuestIntoUser(ctx context.Context, guestToken string, userID uuid.UUID) error {
	guestToken = trimToken(guestToken)
	if guestToken == "" {
		return nil
	}

	guestOwner := domain.Owner{GuestToken: guestToken}
	userOwner := domain.Owner{UserID: &userID}

	guestCart, err := s.repo.Get(ctx, guestOwner)
	if err != nil {
		return err
	}
	if len(guestCart.Items) == 0 {
		return s.repo.Delete(ctx, guestOwner)
	}

	userCart, err := s.repo.Get(ctx, userOwner)
	if err != nil {
		return err
	}

	for _, guestItem := range guestCart.Items {
		merged := false
		for i, userItem := range userCart.Items {
			if sameLine(userItem, guestItem.ProductID, guestItem.SkuID) {
				userCart.Items[i].Quantity += guestItem.Quantity
				merged = true
				break
			}
		}
		if !merged {
			userCart.Items = append(userCart.Items, guestItem)
		}
	}

	if err := s.repo.Save(ctx, userOwner, userCart); err != nil {
		return err
	}
	return s.repo.Delete(ctx, guestOwner)
}

// CheckoutItems converts cart lines to storefront checkout inputs.
func CheckoutItems(cart *domain.Cart) []CheckoutItem {
	if cart == nil || len(cart.Items) == 0 {
		return nil
	}
	items := make([]CheckoutItem, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = CheckoutItem{
			ProductID: item.ProductID,
			SkuID:     item.SkuID,
			Quantity:  item.Quantity,
		}
	}
	return items
}

// CheckoutItem is a cart line used by checkout pricing.
type CheckoutItem struct {
	ProductID uuid.UUID
	SkuID     *uuid.UUID
	Quantity  int
}

func (s *Service) validateProduct(ctx context.Context, productID uuid.UUID, skuID *uuid.UUID) error {
	product, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if product.Status != domainproduct.StatusActive {
		return domainproduct.ErrNotFound
	}
	if skuID != nil {
		for _, sku := range product.SKUs {
			if sku.ID == *skuID {
				return nil
			}
		}
		return domainproduct.ErrNotFound
	}
	return nil
}

func (s *Service) removeLine(ctx context.Context, owner domain.Owner, cart *domain.Cart, productID uuid.UUID, skuID *uuid.UUID) (*domain.Cart, error) {
	next := make([]domain.Item, 0, len(cart.Items))
	removed := false
	for _, item := range cart.Items {
		if sameLine(item, productID, skuID) {
			removed = true
			continue
		}
		next = append(next, item)
	}
	if !removed {
		return nil, domain.ErrItemMissing
	}
	cart.Items = next
	if err := s.repo.Save(ctx, owner, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

func sameLine(item domain.Item, productID uuid.UUID, skuID *uuid.UUID) bool {
	if item.ProductID != productID {
		return false
	}
	if skuID == nil && item.SkuID == nil {
		return true
	}
	if skuID != nil && item.SkuID != nil {
		return *skuID == *item.SkuID
	}
	return false
}

func trimToken(token string) string {
	return strings.TrimSpace(token)
}
