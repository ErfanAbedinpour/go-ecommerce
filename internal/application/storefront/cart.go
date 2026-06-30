package storefront

import (
	"context"

	appcart "app/internal/application/cart"
	domaincart "app/internal/domain/cart"
)

// CartView is the public cart response with server-computed line pricing.
type CartView struct {
	Items   []PreviewLineItem `json:"items"`
	Summary CheckoutSummary   `json:"summary"`
}

// GetCart returns the server cart with current prices and availability.
func (s *Service) GetCart(ctx context.Context, owner domaincart.Owner) (*CartView, error) {
	cart, err := s.carts.Get(ctx, owner)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return &CartView{
			Items: []PreviewLineItem{},
			Summary: CheckoutSummary{
				Currency:      "IRT",
				CurrencyLabel: "تومان",
			},
		}, nil
	}

	lines, subtotal, err := s.buildPreviewLines(ctx, cartItemsToCheckout(cart))
	if err != nil {
		return nil, err
	}

	return &CartView{
		Items: lines,
		Summary: CheckoutSummary{
			SubtotalToman: toMoneyToman(subtotal),
			TotalToman:    toMoneyToman(subtotal),
			Currency:      "IRT",
			CurrencyLabel: "تومان",
		},
	}, nil
}

func cartItemsToCheckout(cart *domaincart.Cart) []CheckoutItemInput {
	items := make([]CheckoutItemInput, len(cart.Items))
	for i, item := range cart.Items {
		items[i] = CheckoutItemInput{
			ProductID: item.ProductID,
			SkuID:     item.SkuID,
			Quantity:  item.Quantity,
		}
	}
	return items
}

func cartOwnerCheckoutItems(ctx context.Context, carts *appcart.Service, owner domaincart.Owner) ([]CheckoutItemInput, error) {
	cart, err := carts.Get(ctx, owner)
	if err != nil {
		return nil, err
	}
	if len(cart.Items) == 0 {
		return nil, domaincart.ErrEmpty
	}
	return cartItemsToCheckout(cart), nil
}
