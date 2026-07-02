package request

// StoreCartAddItemRequest adds a product to the server-side cart.
type StoreCartAddItemRequest struct {
	ProductID string  `json:"product_id" validate:"required,uuid"`
	SkuID     *string `json:"sku_id" validate:"omitempty,uuid"`
	Quantity  int     `json:"quantity" validate:"omitempty,gt=0"`
}

// StoreCartUpdateItemRequest updates a cart line quantity.
type StoreCartUpdateItemRequest struct {
	Quantity int `json:"quantity" validate:"required,gte=0"`
}

// StoreCheckoutValidateCustomerRequest validates guest contact info before checkout.
type StoreCheckoutValidateCustomerRequest struct {
	Email string `json:"email" validate:"required,email"`
	Phone string `json:"phone" validate:"omitempty,max=20"`
}

// StoreCheckoutPreviewRequest validates the server cart and computes totals.
type StoreCheckoutPreviewRequest struct {
	CouponCode     string `json:"coupon_code" validate:"omitempty,max=50"`
	ShippingMethod string `json:"shipping_method" validate:"required,oneof=post courier"`
	ShippingCity   string `json:"shipping_city" validate:"required,max=100"`
}

// StoreCheckoutRequest places an order from the server cart.
type StoreCheckoutRequest struct {
	CouponCode      string                       `json:"coupon_code" validate:"omitempty,max=50"`
	ShippingMethod  string                       `json:"shipping_method" validate:"required,oneof=post courier"`
	ShippingCity    string                       `json:"shipping_city" validate:"required,max=100"`
	Customer        StoreCheckoutCustomerRequest `json:"customer" validate:"required"`
	ShippingAddress OrderAddressRequest          `json:"shipping_address" validate:"required"`
	BillingAddress  *OrderAddressRequest         `json:"billing_address" validate:"omitempty"`
	PaymentMethod   string                       `json:"payment_method" validate:"omitempty,max=50"`
	Notes           string                       `json:"notes" validate:"omitempty,max=2000"`
}

// StoreCheckoutCustomerRequest holds customer info for checkout.
type StoreCheckoutCustomerRequest struct {
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name" validate:"required,max=100"`
	Phone     string `json:"phone" validate:"omitempty,max=20"`
}

// StoreCouponValidateRequest validates a coupon code.
type StoreCouponValidateRequest struct {
	Code          string `json:"code" validate:"required,max=50"`
	SubtotalToman int64  `json:"subtotal_toman" validate:"required,gte=0"`
}

// StorePaymentCallbackRequest is the PSP payment callback payload.
type StorePaymentCallbackRequest struct {
	OrderID   string `json:"order_id" validate:"required,uuid"`
	Authority string `json:"authority" validate:"required,max=100"`
	Status    string `json:"status" validate:"required,oneof=OK NOK ok nok"`
	Signature string `json:"signature" validate:"omitempty,max=128"`
}

// UpdateHeroRequest updates storefront hero settings.
type UpdateHeroRequest struct {
	VideoURL         string `json:"video_url" validate:"omitempty,max=500"`
	Title            string `json:"title" validate:"omitempty,max=255"`
	Subtitle         string `json:"subtitle" validate:"omitempty,max=2000"`
	CTAPrimaryText   string `json:"cta_primary_text" validate:"omitempty,max=100"`
	CTAPrimaryURL    string `json:"cta_primary_url" validate:"omitempty,max=500"`
	CTASecondaryText string `json:"cta_secondary_text" validate:"omitempty,max=100"`
	CTASecondaryURL  string `json:"cta_secondary_url" validate:"omitempty,max=500"`
	IsActive         bool   `json:"is_active"`
}

// UpdateProductSlideRequest updates a product slide configuration.
type UpdateProductSlideRequest struct {
	Title              string `json:"title" validate:"omitempty,max=255"`
	AutoplayIntervalMs int    `json:"autoplay_interval_ms" validate:"omitempty,gte=1000"`
	IsActive           bool   `json:"is_active"`
}

// SlideItemRequest is a product slide item.
type SlideItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	SortOrder int    `json:"sort_order" validate:"omitempty,gte=0"`
	TabLabel  string `json:"tab_label" validate:"omitempty,max=100"`
}

// ProBannerRequest is a promotional banner payload.
type ProBannerRequest struct {
	DesktopImageURL string `json:"desktop_image_url" validate:"required,max=500"`
	MobileImageURL  string `json:"mobile_image_url" validate:"omitempty,max=500"`
	LinkURL         string `json:"link_url" validate:"omitempty,max=500"`
	SortOrder       int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive        bool   `json:"is_active"`
}

// PartnerBrandRequest is a partner brand payload.
type PartnerBrandRequest struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description" validate:"omitempty,max=2000"`
	LogoURL     string `json:"logo_url" validate:"required,max=500"`
	LinkURL     string `json:"link_url" validate:"omitempty,max=500"`
	SortOrder   int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive    bool   `json:"is_active"`
}

// HomepageReviewRequest is a homepage testimonial payload.
type HomepageReviewRequest struct {
	CustomerName string `json:"customer_name" validate:"required,max=255"`
	PhotoURL     string `json:"photo_url" validate:"omitempty,max=500"`
	ReviewText   string `json:"review_text" validate:"required,max=5000"`
	Rating       *int   `json:"rating" validate:"omitempty,gte=1,lte=5"`
	SortOrder    int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive     bool   `json:"is_active"`
}

// UpdateFAQSectionRequest updates FAQ section image.
type UpdateFAQSectionRequest struct {
	ImageURL string `json:"image_url" validate:"omitempty,max=500"`
}

// FAQItemRequest is an FAQ Q&A item.
type FAQItemRequest struct {
	Question  string `json:"question" validate:"required,max=2000"`
	Answer    string `json:"answer" validate:"required,max=5000"`
	SortOrder int    `json:"sort_order" validate:"omitempty,gte=0"`
	IsActive  bool   `json:"is_active"`
}

// ContactSectionRequest updates the contact section image.
type ContactSectionRequest struct {
	ImageURL string `json:"image_url" validate:"omitempty,max=500"`
}

// NavigationItemsRequest wraps navigation menu items.
type NavigationItemsRequest struct {
	Items []NavItemRequest `json:"items" validate:"required,dive"`
}

// UpdateStoreStyleRequest updates active theme and style tokens.
type UpdateStoreStyleRequest struct {
	ActiveThemeID *string           `json:"active_theme_id" validate:"omitempty,uuid"`
	Colors        map[string]string `json:"colors" validate:"omitempty"`
	FontFamily    string            `json:"font_family" validate:"omitempty,max=100"`
}
