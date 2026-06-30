package storefront

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	apporder "app/internal/application/order"
	domaincoupon "app/internal/domain/coupon"
	domaincustomer "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
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
	Items          []CheckoutItemInput
	CouponCode     string
	ShippingAmount float64
	TaxAmount      float64
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
	Items           []CheckoutItemInput
	CouponCode      string
	Customer        CheckoutCustomerInput
	ShippingAddress domainorder.Address
	BillingAddress  domainorder.Address
	ShippingAmount  float64
	TaxAmount       float64
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
	OrderID       uuid.UUID `json:"order_id"`
	OrderNumber   string    `json:"order_number"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	TotalToman    int64     `json:"total_toman"`
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

// PreviewCheckout validates cart items and computes totals without creating an order.
func (s *Service) PreviewCheckout(ctx context.Context, input PreviewCheckoutInput) (*PreviewCheckoutOutput, error) {
	if len(input.Items) == 0 {
		return nil, domainorder.ErrEmptyOrder
	}

	lineItems, subtotal, err := s.buildPreviewLines(ctx, input.Items)
	if err != nil {
		return nil, err
	}

	discount := 0.0
	var couponResult *CouponResult
	if input.CouponCode != "" {
		result, err := s.ValidateCoupon(ctx, CouponValidateInput{
			Code:          input.CouponCode,
			SubtotalToman: toMoneyToman(subtotal),
		})
		if err != nil {
			return nil, err
		}
		couponResult = result
		if result.IsValid {
			discount = float64(result.DiscountToman)
		}
	}

	shipping := roundMoney(input.ShippingAmount)
	tax := roundMoney(input.TaxAmount)
	total := roundMoney(subtotal - discount + shipping + tax)
	if total < 0 {
		total = 0
	}

	return &PreviewCheckoutOutput{
		Items: lineItems,
		Summary: CheckoutSummary{
			SubtotalToman: toMoneyToman(subtotal),
			DiscountToman: toMoneyToman(discount),
			ShippingToman: toMoneyToman(shipping),
			TaxToman:      toMoneyToman(tax),
			TotalToman:    toMoneyToman(total),
			Currency:      "IRT",
			CurrencyLabel: "تومان",
		},
		Coupon: couponResult,
	}, nil
}

// PlaceCheckout creates an unpaid storefront order for a guest or authenticated customer.
func (s *Service) PlaceCheckout(ctx context.Context, input PlaceCheckoutInput) (*PlaceCheckoutOutput, error) {
	if len(input.Items) == 0 {
		return nil, domainorder.ErrEmptyOrder
	}

	customerID, err := s.resolveCheckoutCustomer(ctx, input)
	if err != nil {
		return nil, err
	}

	orderItems := make([]apporder.CreateItemInput, len(input.Items))
	for i, item := range input.Items {
		orderItems[i] = apporder.CreateItemInput{
			ProductID: item.ProductID,
			SkuID:     item.SkuID,
			Quantity:  item.Quantity,
		}
	}

	order, err := s.orders.Create(ctx, apporder.CreateInput{
		CustomerID:      customerID,
		Items:           orderItems,
		CouponCode:      input.CouponCode,
		ShippingAmount:  input.ShippingAmount,
		TaxAmount:       input.TaxAmount,
		BillingAddress:  input.BillingAddress,
		ShippingAddress: input.ShippingAddress,
		PaymentMethod:   input.PaymentMethod,
		PaymentStatus:   domainorder.PaymentUnpaid.String(),
		Notes:           input.Notes,
	})
	if err != nil {
		return nil, err
	}

	if s.mailer != nil {
		emailTo := input.Customer.Email
		if emailTo == "" {
			customer, err := s.customers.FindByID(ctx, customerID)
			if err == nil && customer != nil {
				emailTo = customer.Email
			}
		}
		if emailTo != "" {
			go func() {
				_ = s.mailer.SendOrderConfirmation(context.Background(), emailTo, order.OrderNumber, order.Total)
			}()
		}
	}

	return &PlaceCheckoutOutput{
		OrderID:       order.ID,
		OrderNumber:   order.OrderNumber,
		Status:        order.Status.String(),
		PaymentStatus: order.PaymentStatus.String(),
		TotalToman:    toMoneyToman(order.Total),
	}, nil
}

func (s *Service) resolveCheckoutCustomer(ctx context.Context, input PlaceCheckoutInput) (uuid.UUID, error) {
	email := strings.ToLower(strings.TrimSpace(input.Customer.Email))

	if input.UserID != nil {
		if existing, err := s.customers.FindByUserID(ctx, *input.UserID); err == nil {
			return existing.ID, nil
		} else if err != domaincustomer.ErrNotFound {
			return uuid.Nil, err
		}
	}

	if email != "" {
		if existing, err := s.customers.FindByEmail(ctx, email); err == nil {
			if input.UserID != nil && existing.UserID == nil {
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

	now := time.Now().UTC()
	customerType := domaincustomer.TypeGuest
	if input.UserID != nil {
		customerType = domaincustomer.TypeRegistered
	}

	c := &domaincustomer.Customer{
		ID:        uuid.New(),
		UserID:    input.UserID,
		Email:     email,
		FirstName: strings.TrimSpace(input.Customer.FirstName),
		LastName:  strings.TrimSpace(input.Customer.LastName),
		Phone:     strings.TrimSpace(input.Customer.Phone),
		Type:      customerType,
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
