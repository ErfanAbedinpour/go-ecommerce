package storefront

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	apporder "app/internal/application/order"
	domaincoupon "app/internal/domain/coupon"
	domaincart "app/internal/domain/cart"
	domaincustomer "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	"app/internal/domain/user"
	"app/pkg/apperror"
)

// CheckoutItemInput is a cart line for checkout preview or placement.
type CheckoutItemInput struct {
	ProductID uuid.UUID
	SkuID     *uuid.UUID
	Quantity  int
}

// PreviewCheckoutInput holds data for checkout total preview.
type PreviewCheckoutInput struct {
	Owner          domaincart.Owner
	CouponCode     string
	ShippingMethod string
	ShippingCity   string
}

// PreviewLineItem is a validated checkout line with pricing.
type PreviewLineItem struct {
	ProductID         uuid.UUID `json:"product_id"`
	SkuID             *uuid.UUID `json:"sku_id,omitempty"`
	ProductName       string    `json:"product_name"`
	SKUCode           string    `json:"sku_code,omitempty"`
	VariantLabel      string    `json:"variant_label,omitempty"`
	ThumbnailURL      string    `json:"thumbnail_url,omitempty"`
	UnitPriceToman    int64     `json:"unit_price_toman"`
	Quantity          int       `json:"quantity"`
	LineTotalToman    int64     `json:"line_total_toman"`
	IsAvailable       bool      `json:"is_available"`
	AvailableQuantity int       `json:"available_quantity"`
}

// CheckoutSummary holds computed checkout totals.
type CheckoutSummary struct {
	SubtotalToman  int64  `json:"subtotal_toman"`
	DiscountToman  int64  `json:"discount_toman"`
	ShippingToman  int64  `json:"shipping_toman"`
	TaxToman       int64  `json:"tax_toman"`
	TotalToman     int64  `json:"total_toman"`
	Currency       string `json:"currency"`
	CurrencyLabel  string `json:"currency_label"`
}

// PreviewCheckoutOutput is the checkout preview response.
type PreviewCheckoutOutput struct {
	Items   []PreviewLineItem `json:"items"`
	Summary CheckoutSummary   `json:"summary"`
	Coupon  *CouponResult     `json:"coupon,omitempty"`
}

// PlaceCheckoutInput holds data for placing a storefront order.
type PlaceCheckoutInput struct {
	Owner           domaincart.Owner
	CouponCode      string
	Customer        CheckoutCustomerInput
	ShippingAddress domainorder.Address
	BillingAddress  domainorder.Address
	ShippingMethod  string
	ShippingCity    string
	PaymentMethod   string
	Notes           string
	UserID          *uuid.UUID
}

// CheckoutCustomerInput holds guest or authenticated customer data.
type CheckoutCustomerInput struct {
	Email     string
	FirstName string
	LastName  string
	Phone     string
}

// PlaceCheckoutOutput is the placed order response.
type PlaceCheckoutOutput struct {
	OrderID          uuid.UUID `json:"order_id"`
	OrderNumber      string    `json:"order_number"`
	Status           string    `json:"status"`
	PaymentStatus    string    `json:"payment_status"`
	TotalToman       int64     `json:"total_toman"`
	PaymentExpiresAt string    `json:"payment_expires_at,omitempty"`
}

// CheckoutPricingSnapshot holds quote-locked checkout totals and line items.
type CheckoutPricingSnapshot struct {
	Items          []PreviewLineItem
	Subtotal       float64
	DiscountAmount float64
	ShippingAmount float64
	TaxAmount      float64
	Total          float64
	CouponID       *uuid.UUID
	Coupon         *CouponResult
}

// CouponValidateInput holds coupon validation request data.
type CouponValidateInput struct {
	Code          string
	SubtotalToman int64
}

// CouponResult holds coupon validation outcome.
type CouponResult struct {
	Code         string `json:"code,omitempty"`
	IsValid      bool   `json:"is_valid"`
	DiscountToman int64 `json:"discount_toman"`
	Message      string `json:"message,omitempty"`
}

// ValidateCoupon validates a coupon against a subtotal.
func (s *Service) ValidateCoupon(ctx context.Context, input CouponValidateInput) (*CouponResult, error) {
	code := domaincoupon.NormalizeCode(input.Code)
	if code == "" {
		return &CouponResult{IsValid: false, Message: "coupon code is required"}, nil
	}

	subtotal := float64(input.SubtotalToman)
	coupon, err := s.coupons.FindByCode(ctx, code)
	if err != nil {
		if err == domaincoupon.ErrNotFound {
			return &CouponResult{IsValid: false, Message: "coupon not found"}, nil
		}
		return nil, err
	}

	discount, err := coupon.CalculateDiscount(subtotal)
	if err != nil {
		return &CouponResult{
			Code:    coupon.Code,
			IsValid: false,
			Message: couponErrorMessage(err),
		}, nil
	}

	return &CouponResult{
		Code:          coupon.Code,
		IsValid:       true,
		DiscountToman: toMoneyToman(discount),
		Message:       "coupon applied",
	}, nil
}

// PreviewCheckout validates the server cart and computes totals without creating an order.
func (s *Service) PreviewCheckout(ctx context.Context, input PreviewCheckoutInput) (*PreviewCheckoutOutput, error) {
	snapshot, err := s.buildCheckoutSnapshot(ctx, input)
	if err != nil {
		return nil, err
	}

	return &PreviewCheckoutOutput{
		Items: snapshot.Items,
		Summary: CheckoutSummary{
			SubtotalToman: toMoneyToman(snapshot.Subtotal),
			DiscountToman: toMoneyToman(snapshot.DiscountAmount),
			ShippingToman: toMoneyToman(snapshot.ShippingAmount),
			TaxToman:      toMoneyToman(snapshot.TaxAmount),
			TotalToman:    toMoneyToman(snapshot.Total),
			Currency:      "IRT",
			CurrencyLabel: "تومان",
		},
		Coupon: snapshot.Coupon,
	}, nil
}

func (s *Service) buildCheckoutSnapshot(ctx context.Context, input PreviewCheckoutInput) (*CheckoutPricingSnapshot, error) {
	items, err := cartOwnerCheckoutItems(ctx, s.carts, input.Owner)
	if err != nil {
		return nil, err
	}

	lineItems, subtotal, err := s.buildPreviewLines(ctx, items)
	if err != nil {
		return nil, err
	}

	discount := 0.0
	var couponResult *CouponResult
	var couponID *uuid.UUID
	if input.CouponCode != "" {
		code := domaincoupon.NormalizeCode(input.CouponCode)
		result, err := s.ValidateCoupon(ctx, CouponValidateInput{
			Code:          code,
			SubtotalToman: toMoneyToman(subtotal),
		})
		if err != nil {
			return nil, err
		}
		couponResult = result
		if result.IsValid {
			discount = float64(result.DiscountToman)
			coupon, err := s.coupons.FindByCode(ctx, code)
			if err != nil {
				return nil, err
			}
			id := coupon.ID
			couponID = &id
		}
	}

	shipping, err := s.computeShippingAmount(input.ShippingMethod, input.ShippingCity)
	if err != nil {
		return nil, err
	}
	tax := s.computeTaxAmount(subtotal - discount + shipping)
	total := roundMoney(subtotal - discount + shipping + tax)
	if total < 0 {
		total = 0
	}

	if err := s.validateMinOrder(ctx, total); err != nil {
		return nil, err
	}

	return &CheckoutPricingSnapshot{
		Items:          lineItems,
		Subtotal:       subtotal,
		DiscountAmount: discount,
		ShippingAmount: shipping,
		TaxAmount:      tax,
		Total:          total,
		CouponID:       couponID,
		Coupon:         couponResult,
	}, nil
}

// PlaceCheckout creates an unpaid storefront order from the server cart.
func (s *Service) PlaceCheckout(ctx context.Context, input PlaceCheckoutInput) (*PlaceCheckoutOutput, error) {
	snapshot, err := s.buildCheckoutSnapshot(ctx, PreviewCheckoutInput{
		Owner:          input.Owner,
		CouponCode:     input.CouponCode,
		ShippingMethod: input.ShippingMethod,
		ShippingCity:   input.ShippingCity,
	})
	if err != nil {
		return nil, err
	}

	if unavailable := unavailablePreviewItems(snapshot.Items); len(unavailable) > 0 {
		return nil, stockConflictError(unavailable)
	}

	customerID, err := s.resolveCheckoutCustomer(ctx, input)
	if err != nil {
		return nil, err
	}

	orderItems := make([]apporder.SnapshotLineItem, len(snapshot.Items))
	for i, item := range snapshot.Items {
		orderItems[i] = apporder.SnapshotLineItem{
			ProductID:   item.ProductID,
			ProductName: item.ProductName,
			ProductSKU:  item.SKUCode,
			Quantity:    item.Quantity,
			UnitPrice:   float64(item.UnitPriceToman),
			TotalPrice:  float64(item.LineTotalToman),
		}
	}

	order, err := s.orders.CreateFromSnapshot(ctx, apporder.CreateFromSnapshotInput{
		CustomerID:      customerID,
		CouponID:        snapshot.CouponID,
		Subtotal:        snapshot.Subtotal,
		DiscountAmount:  snapshot.DiscountAmount,
		ShippingAmount:  snapshot.ShippingAmount,
		TaxAmount:       snapshot.TaxAmount,
		Total:           snapshot.Total,
		Items:           orderItems,
		BillingAddress:  input.BillingAddress,
		ShippingAddress: input.ShippingAddress,
		PaymentMethod:   input.PaymentMethod,
		Notes:           input.Notes,
	})
	if err != nil {
		return nil, err
	}

	if err := s.carts.Clear(ctx, input.Owner); err != nil {
		return nil, err
	}

	out := &PlaceCheckoutOutput{
		OrderID:       order.ID,
		OrderNumber:   order.OrderNumber,
		Status:        order.Status.String(),
		PaymentStatus: order.PaymentStatus.String(),
		TotalToman:    toMoneyToman(order.Total),
	}
	if order.PaymentExpiresAt != nil {
		out.PaymentExpiresAt = order.PaymentExpiresAt.UTC().Format(time.RFC3339)
	}
	return out, nil
}

// ValidateGuestCheckoutCustomerInput holds guest contact info for checkout validation.
type ValidateGuestCheckoutCustomerInput struct {
	Email string
	Phone string
}

// ValidateGuestCheckoutCustomerOutput confirms guest checkout may proceed.
type ValidateGuestCheckoutCustomerOutput struct {
	OK bool `json:"ok"`
}

// ValidateGuestCheckoutCustomer checks whether a guest may continue checkout with the given contact info.
func (s *Service) ValidateGuestCheckoutCustomer(ctx context.Context, input ValidateGuestCheckoutCustomerInput) (*ValidateGuestCheckoutCustomerOutput, error) {
	if err := s.assertGuestCanCheckout(ctx, input.Email, input.Phone); err != nil {
		return nil, err
	}
	return &ValidateGuestCheckoutCustomerOutput{OK: true}, nil
}

func (s *Service) assertGuestCanCheckout(ctx context.Context, email, phone string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	phone = strings.TrimSpace(phone)

	if email != "" {
		if _, err := s.users.FindByEmail(ctx, email); err == nil {
			return domaincustomer.ErrAccountExistsLoginRequired
		} else if err != user.ErrNotFound {
			return err
		}

		if existing, err := s.customers.FindByEmail(ctx, email); err == nil {
			if existing.UserID != nil || existing.Type == domaincustomer.TypeRegistered {
				return domaincustomer.ErrAccountExistsLoginRequired
			}
			if phone != "" && existing.Phone != "" && existing.Phone != phone {
				return domaincustomer.ErrAccountExistsLoginRequired
			}
		} else if err != domaincustomer.ErrNotFound {
			return err
		}
	}

	if phone != "" {
		if _, err := s.users.FindByPhone(ctx, phone); err == nil {
			return domaincustomer.ErrAccountExistsLoginRequired
		} else if err != user.ErrNotFound {
			return err
		}

		if _, err := s.customers.FindRegisteredByPhone(ctx, phone); err == nil {
			return domaincustomer.ErrAccountExistsLoginRequired
		} else if err != domaincustomer.ErrNotFound {
			return err
		}

		if existing, err := s.customers.FindGuestByPhone(ctx, phone); err == nil {
			if email != "" && existing.Email != "" && strings.ToLower(existing.Email) != email {
				return domaincustomer.ErrAccountExistsLoginRequired
			}
		} else if err != domaincustomer.ErrNotFound {
			return err
		}
	}

	return nil
}

func (s *Service) resolveCheckoutCustomer(ctx context.Context, input PlaceCheckoutInput) (uuid.UUID, error) {
	email := strings.ToLower(strings.TrimSpace(input.Customer.Email))
	phone := strings.TrimSpace(input.Customer.Phone)

	if input.UserID != nil {
		if existing, err := s.customers.FindByUserID(ctx, *input.UserID); err == nil {
			return existing.ID, nil
		} else if err != domaincustomer.ErrNotFound {
			return uuid.Nil, err
		}

		if email != "" {
			if existing, err := s.customers.FindByEmail(ctx, email); err == nil {
				if existing.UserID != nil && *existing.UserID != *input.UserID {
					return uuid.Nil, domaincustomer.ErrAccountExistsLoginRequired
				}
				if existing.UserID == nil {
					existing.UserID = input.UserID
					existing.Type = domaincustomer.TypeRegistered
					existing.UpdatedAt = time.Now().UTC()
					if err := s.customers.Update(ctx, existing); err != nil {
						return uuid.Nil, err
					}
				}
				return existing.ID, nil
			} else if err != domaincustomer.ErrNotFound {
				return uuid.Nil, err
			}
		}
	} else {
		if err := s.assertGuestCanCheckout(ctx, email, phone); err != nil {
			return uuid.Nil, err
		}
		return s.resolveGuestCustomer(ctx, input, email, phone)
	}

	now := time.Now().UTC()
	userID := input.UserID
	c := &domaincustomer.Customer{
		ID:        uuid.New(),
		UserID:    userID,
		Type:      domaincustomer.TypeRegistered,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.customers.Create(ctx, c); err != nil {
		return uuid.Nil, err
	}
	return c.ID, nil
}

func (s *Service) resolveGuestCustomer(ctx context.Context, input PlaceCheckoutInput, email, phone string) (uuid.UUID, error) {
	var byEmail, byPhone *domaincustomer.Customer

	if email != "" {
		if guest, err := s.customers.FindGuestByEmail(ctx, email); err == nil {
			byEmail = guest
		} else if err != domaincustomer.ErrNotFound {
			return uuid.Nil, err
		}
	}
	if phone != "" {
		if guest, err := s.customers.FindGuestByPhone(ctx, phone); err == nil {
			byPhone = guest
		} else if err != domaincustomer.ErrNotFound {
			return uuid.Nil, err
		}
	}

	if byEmail != nil && byPhone != nil && byEmail.ID != byPhone.ID {
		return uuid.Nil, domaincustomer.ErrAccountExistsLoginRequired
	}

	existing := byEmail
	if existing == nil {
		existing = byPhone
	}
	if existing != nil {
		if email != "" && existing.Email != "" && strings.ToLower(existing.Email) != email {
			return uuid.Nil, domaincustomer.ErrAccountExistsLoginRequired
		}
		if phone != "" && existing.Phone != "" && existing.Phone != phone {
			return uuid.Nil, domaincustomer.ErrAccountExistsLoginRequired
		}
		return existing.ID, nil
	}

	now := time.Now().UTC()
	c := &domaincustomer.Customer{
		ID:        uuid.New(),
		Type:      domaincustomer.TypeGuest,
		Email:     email,
		FirstName: strings.TrimSpace(input.Customer.FirstName),
		LastName:  strings.TrimSpace(input.Customer.LastName),
		Phone:     phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.customers.Create(ctx, c); err != nil {
		return uuid.Nil, err
	}
	return c.ID, nil
}

func (s *Service) buildPreviewLines(ctx context.Context, items []CheckoutItemInput) ([]PreviewLineItem, float64, error) {
	lines := make([]PreviewLineItem, 0, len(items))
	var subtotal float64

	for _, item := range items {
		if item.Quantity <= 0 {
			continue
		}

		product, err := s.products.FindByID(ctx, item.ProductID)
		if err != nil {
			return nil, 0, err
		}
		if product.Status != domainproduct.StatusActive {
			return nil, 0, domainproduct.ErrNotFound
		}

		skuCode, variantLabel, err := resolveSKU(product, item.SkuID)
		if err != nil {
			return nil, 0, err
		}

		unitPrice := roundMoney(product.EffectivePrice())
		lineTotal := roundMoney(unitPrice * float64(item.Quantity))
		subtotal += lineTotal

		thumbnail := ""
		if len(product.Images) > 0 {
			thumbnail = product.Images[0].URL
		}

		lines = append(lines, PreviewLineItem{
			ProductID:         product.ID,
			SkuID:             item.SkuID,
			ProductName:       product.Name,
			SKUCode:           skuCode,
			VariantLabel:      variantLabel,
			ThumbnailURL:      thumbnail,
			UnitPriceToman:    toMoneyToman(unitPrice),
			Quantity:          item.Quantity,
			LineTotalToman:    toMoneyToman(lineTotal),
			IsAvailable:       product.Inventory.Quantity >= item.Quantity,
			AvailableQuantity: product.Inventory.Quantity,
		})
	}

	return lines, roundMoney(subtotal), nil
}

func resolveSKU(product *domainproduct.Product, skuID *uuid.UUID) (code, variantLabel string, err error) {
	if skuID != nil {
		for _, sku := range product.SKUs {
			if sku.ID == *skuID {
				return sku.Code, formatVariantLabel(sku.Attributes), nil
			}
		}
		return "", "", domainorder.ErrInvalidSKU
	}
	if len(product.SKUs) > 0 {
		sku := product.SKUs[0]
		return sku.Code, formatVariantLabel(sku.Attributes), nil
	}
	return "", "", nil
}

func formatVariantLabel(attrs map[string]string) string {
	if len(attrs) == 0 {
		return ""
	}
	parts := make([]string, 0, len(attrs))
	for _, v := range attrs {
		if v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, " · ")
}

func couponErrorMessage(err error) string {
	switch err {
	case domaincoupon.ErrExpired:
		return "coupon has expired"
	case domaincoupon.ErrExhausted:
		return "coupon usage limit reached"
	case domaincoupon.ErrMinOrderNotMet:
		return "order subtotal does not meet coupon minimum"
	case domaincoupon.ErrNotApplicable:
		return "coupon is not active"
	default:
		return "coupon is not valid"
	}
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

// ShippingMethod is an available delivery option at checkout.
type ShippingMethod struct {
	Code       string `json:"code"`
	Label      string `json:"label"`
	PriceToman int64  `json:"price_toman"`
	EtaDays    string `json:"eta_days"`
}

// ShippingMethodList is the response for available shipping methods.
type ShippingMethodList struct {
	Data []ShippingMethod `json:"data"`
}

// GetShippingMethods returns delivery options for a destination city.
func (s *Service) GetShippingMethods(_ context.Context, city string) (*ShippingMethodList, error) {
	city = strings.TrimSpace(city)
	if city == "" {
		return nil, apperror.Validation("city is required", map[string]string{"city": "is required"})
	}

	methods := []ShippingMethod{
		{
			Code:       "post",
			Label:      "پست پیشتاز",
			PriceToman: 85000,
			EtaDays:    "2-4",
		},
	}

	if isCourierCity(city) {
		methods = append(methods, ShippingMethod{
			Code:       "courier",
			Label:      "پیک",
			PriceToman: 120000,
			EtaDays:    "1",
		})
	}

	return &ShippingMethodList{Data: methods}, nil
}

func isCourierCity(city string) bool {
	normalized := strings.ToLower(strings.TrimSpace(city))
	switch normalized {
	case "tehran", "تهران", "karaj", "کرج":
		return true
	default:
		return false
	}
}

func (s *Service) computeShippingAmount(method, city string) (float64, error) {
	method = strings.TrimSpace(strings.ToLower(method))
	city = strings.TrimSpace(city)
	if method == "" {
		return 0, apperror.Validation("shipping method is required", map[string]string{"shipping_method": "is required"})
	}
	if city == "" {
		return 0, apperror.Validation("shipping city is required", map[string]string{"shipping_city": "is required"})
	}

	switch method {
	case "post":
		return 85000, nil
	case "courier":
		if !isCourierCity(city) {
			return 0, apperror.Validation("courier is not available for this city", map[string]string{"shipping_method": "unavailable for city"})
		}
		return 120000, nil
	default:
		return 0, apperror.Validation("invalid shipping method", map[string]string{"shipping_method": "must be post or courier"})
	}
}

func (s *Service) computeTaxAmount(_ float64) float64 {
	return 0
}

func (s *Service) validateMinOrder(ctx context.Context, total float64) error {
	store, err := s.settings.Get(ctx)
	if err != nil {
		return err
	}
	minOrder := store.Checkout.WithDefaults().MinOrderToman
	if toMoneyToman(total) < minOrder {
		return apperror.Validation("order total is below minimum", map[string]string{
			"total_toman": fmt.Sprintf("must be at least %d", minOrder),
		})
	}
	return nil
}

type unavailableItem struct {
	ProductID         string `json:"product_id"`
	RequestedQuantity int    `json:"requested_quantity"`
	AvailableQuantity int    `json:"available_quantity"`
}

func unavailablePreviewItems(items []PreviewLineItem) []unavailableItem {
	out := make([]unavailableItem, 0)
	for _, item := range items {
		if item.IsAvailable {
			continue
		}
		out = append(out, unavailableItem{
			ProductID:         item.ProductID.String(),
			RequestedQuantity: item.Quantity,
			AvailableQuantity: item.AvailableQuantity,
		})
	}
	return out
}

func stockConflictError(items []unavailableItem) error {
	if len(items) == 0 {
		return apperror.Conflict("some items are unavailable")
	}
	return apperror.Conflict(fmt.Sprintf("some items are unavailable (first: %s)", items[0].ProductID))
}
