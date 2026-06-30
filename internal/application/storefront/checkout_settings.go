package storefront

import (
	"context"

	domainsettings "app/internal/domain/settings"
)

// GetCheckoutSettings returns public checkout configuration for the storefront.
func (s *Service) GetCheckoutSettings(ctx context.Context) (CheckoutSettingsOutput, error) {
	store, err := s.settings.Get(ctx)
	if err != nil {
		return CheckoutSettingsOutput{}, err
	}
	return CheckoutSettingsOutput(store.Checkout.WithDefaults()), nil
}

// CheckoutSettingsOutput is the public checkout settings response.
type CheckoutSettingsOutput = domainsettings.Checkout
