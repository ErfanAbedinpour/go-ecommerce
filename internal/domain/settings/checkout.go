package settings

// Checkout holds public storefront checkout configuration.
type Checkout struct {
	MinOrderToman  int64    `json:"min_order_toman"`
	PaymentMethods []string `json:"payment_methods"`
	CODEEnabled    bool     `json:"cod_enabled"`
	CODCities      []string `json:"cod_cities"`
	CurrencyLabel  string   `json:"currency_label"`
}

// DefaultCheckout returns sensible checkout defaults when none are configured.
func DefaultCheckout() Checkout {
	return Checkout{
		MinOrderToman:  100000,
		PaymentMethods: []string{"online", "cod"},
		CODEEnabled:    true,
		CODCities:      []string{"تهران", "کرج", "Tehran", "Karaj"},
		CurrencyLabel:  "تومان",
	}
}

// WithDefaults fills zero values from DefaultCheckout.
func (c Checkout) WithDefaults() Checkout {
	def := DefaultCheckout()
	if c.MinOrderToman <= 0 {
		c.MinOrderToman = def.MinOrderToman
	}
	if len(c.PaymentMethods) == 0 {
		c.PaymentMethods = def.PaymentMethods
	}
	if len(c.CODCities) == 0 {
		c.CODCities = def.CODCities
	}
	if c.CurrencyLabel == "" {
		c.CurrencyLabel = def.CurrencyLabel
	}
	return c
}
