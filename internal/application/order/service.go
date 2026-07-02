package order

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"

	domaincoupon "app/internal/domain/coupon"
	domaincustomer "app/internal/domain/customer"
	domain "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	domainsettings "app/internal/domain/settings"
	"app/pkg/apperror"
	"app/pkg/pagination"
)

// Service handles order management use cases.
type Service struct {
	repo        domain.Repository
	products    domainproduct.Repository
	customers   domaincustomer.Repository
	coupons     domaincoupon.Repository
	settings    domainsettings.Repository
	paymentTTL  time.Duration
}

// NewService creates a new order Service.
func NewService(
	repo domain.Repository,
	products domainproduct.Repository,
	customers domaincustomer.Repository,
	coupons domaincoupon.Repository,
	settings domainsettings.Repository,
	paymentTTL time.Duration,
) *Service {
	return &Service{
		repo:       repo,
		products:   products,
		customers:  customers,
		coupons:    coupons,
		settings:   settings,
		paymentTTL: paymentTTL,
	}
}

// List returns a paginated order list.
func (s *Service) List(ctx context.Context, filter domain.ListFilter, page pagination.Params) (pagination.Paginated[domain.ListItem], error) {
	items, total, err := s.repo.List(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.ListItem]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// GetByID returns full order detail.
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.repo.FindByID(ctx, id)
}

// GetInvoice returns printable invoice data for an order.
func (s *Service) GetInvoice(ctx context.Context, id uuid.UUID) (*domain.Invoice, error) {
	o, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	store, err := s.settings.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.Invoice{
		InvoiceNumber: o.OrderNumber,
		IssuedAt:      time.Now().UTC(),
		Store: domain.InvoiceStore{
			Name:    store.Site.Name,
			URL:     store.Site.URL,
			LogoURL: store.Site.LogoURL,
			Email:   store.Contact.Email,
			Phone:   store.Contact.Phone,
			Address: store.Contact.Address,
			City:    store.Contact.City,
			Country: store.Contact.Country,
		},
		Order: o,
	}, nil
}

// CreateItemInput holds a line item for manual order creation.
type CreateItemInput struct {
	ProductID uuid.UUID
	SkuID     *uuid.UUID
	Quantity  int
}

// CreateInput holds data for manual order creation.
type CreateInput struct {
	CustomerID      uuid.UUID
	Items           []CreateItemInput
	CouponCode      string
	ShippingAmount  float64
	TaxAmount       float64
	BillingAddress  domain.Address
	ShippingAddress domain.Address
	PaymentMethod   string
	TransactionID   string
	PaymentStatus   string
	Notes           string
	ChangedBy       uuid.UUID
}

// Create creates a manual order, deducts inventory, and records initial status history.
func (s *Service) Create(ctx context.Context, input CreateInput) (*domain.Order, error) {
	if len(input.Items) == 0 {
		return nil, domain.ErrEmptyOrder
	}

	if _, err := s.customers.FindByID(ctx, input.CustomerID); err != nil {
		return nil, err
	}

	items, subtotal, err := s.buildLineItems(ctx, input.Items)
	if err != nil {
		return nil, err
	}

	var couponID *uuid.UUID
	discountAmount := 0.0
	if input.CouponCode != "" {
		coupon, err := s.coupons.FindByCode(ctx, domaincoupon.NormalizeCode(input.CouponCode))
		if err != nil {
			return nil, err
		}
		discountAmount, err = coupon.CalculateDiscount(subtotal)
		if err != nil {
			return nil, err
		}
		id := coupon.ID
		couponID = &id
	}

	shippingAmount := roundMoney(input.ShippingAmount)
	taxAmount := roundMoney(input.TaxAmount)
	total := roundMoney(subtotal - discountAmount + shippingAmount + taxAmount)
	if total < 0 {
		total = 0
	}

	paymentStatus := domain.PaymentUnpaid
	if input.PaymentStatus != "" {
		parsed, err := domain.ParsePaymentStatus(input.PaymentStatus)
		if err != nil {
			return nil, err
		}
		paymentStatus = parsed
	}

	orderNumber, err := s.repo.NextOrderNumber(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	orderID := uuid.New()
	for i := range items {
		items[i].ID = uuid.New()
		items[i].OrderID = orderID
	}

	o := &domain.Order{
		ID:              orderID,
		OrderNumber:     orderNumber,
		CustomerID:      input.CustomerID,
		CouponID:        couponID,
		Status:          domain.StatusPending,
		PaymentStatus:   paymentStatus,
		Subtotal:        subtotal,
		DiscountAmount:  discountAmount,
		ShippingAmount:  shippingAmount,
		TaxAmount:       taxAmount,
		Total:           total,
		Notes:           input.Notes,
		PaymentMethod:   input.PaymentMethod,
		TransactionID:   input.TransactionID,
		BillingAddress:  input.BillingAddress,
		ShippingAddress: input.ShippingAddress,
		Items:           items,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.Create(ctx, o); err != nil {
		return nil, err
	}

	if couponID != nil {
		if err := s.repo.IncrementCouponUsage(ctx, *couponID); err != nil {
			return nil, err
		}
	}

	if err := s.customers.RecordOrderPlaced(ctx, input.CustomerID, total, now); err != nil {
		return nil, err
	}

	note := "Manual order created"
	if err := s.recordHistory(ctx, orderID, nil, domain.StatusPending, note, input.ChangedBy, now); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, orderID)
}

// SnapshotLineItem is a priced checkout line used to create an order without recomputing totals.
type SnapshotLineItem struct {
	ProductID   uuid.UUID
	ProductName string
	ProductSKU  string
	Quantity    int
	UnitPrice   float64
	TotalPrice  float64
}

// CreateFromSnapshotInput holds quote-locked checkout data for storefront order placement.
type CreateFromSnapshotInput struct {
	CustomerID      uuid.UUID
	CouponID        *uuid.UUID
	Subtotal        float64
	DiscountAmount  float64
	ShippingAmount  float64
	TaxAmount       float64
	Total           float64
	Items           []SnapshotLineItem
	BillingAddress  domain.Address
	ShippingAddress domain.Address
	PaymentMethod   string
	Notes           string
}

// CreateFromSnapshot creates a storefront order from a checkout pricing snapshot.
func (s *Service) CreateFromSnapshot(ctx context.Context, input CreateFromSnapshotInput) (*domain.Order, error) {
	if len(input.Items) == 0 {
		return nil, domain.ErrEmptyOrder
	}

	if _, err := s.customers.FindByID(ctx, input.CustomerID); err != nil {
		return nil, err
	}

	if input.CouponID != nil {
		coupon, err := s.coupons.FindByID(ctx, *input.CouponID)
		if err != nil {
			return nil, err
		}
		if _, err := coupon.CalculateDiscount(input.Subtotal); err != nil {
			return nil, err
		}
	}

	subtotal := roundMoney(input.Subtotal)
	discountAmount := roundMoney(input.DiscountAmount)
	shippingAmount := roundMoney(input.ShippingAmount)
	taxAmount := roundMoney(input.TaxAmount)
	total := roundMoney(input.Total)
	expectedTotal := roundMoney(subtotal - discountAmount + shippingAmount + taxAmount)
	if total != expectedTotal {
		return nil, apperror.Conflict("checkout total does not match pricing snapshot")
	}

	orderNumber, err := s.repo.NextOrderNumber(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(s.paymentTTL)
	orderID := uuid.New()

	items := make([]domain.Item, len(input.Items))
	for i, line := range input.Items {
		items[i] = domain.Item{
			ID:          uuid.New(),
			OrderID:     orderID,
			ProductID:   line.ProductID,
			ProductName: line.ProductName,
			ProductSKU:  line.ProductSKU,
			Quantity:    line.Quantity,
			UnitPrice:   line.UnitPrice,
			TotalPrice:  line.TotalPrice,
		}
	}

	o := &domain.Order{
		ID:               orderID,
		OrderNumber:      orderNumber,
		CustomerID:       input.CustomerID,
		CouponID:         input.CouponID,
		Status:           domain.StatusPending,
		PaymentStatus:    domain.PaymentUnpaid,
		Subtotal:         subtotal,
		DiscountAmount:   discountAmount,
		ShippingAmount:   shippingAmount,
		TaxAmount:        taxAmount,
		Total:            total,
		Notes:            input.Notes,
		PaymentMethod:    input.PaymentMethod,
		PaymentExpiresAt: &expiresAt,
		BillingAddress:   input.BillingAddress,
		ShippingAddress:  input.ShippingAddress,
		Items:            items,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.repo.Create(ctx, o); err != nil {
		return nil, err
	}

	if input.CouponID != nil {
		if err := s.repo.IncrementCouponUsage(ctx, *input.CouponID); err != nil {
			return nil, err
		}
	}

	if err := s.customers.RecordOrderPlaced(ctx, input.CustomerID, total, now); err != nil {
		return nil, err
	}

	if err := s.recordHistory(ctx, orderID, nil, domain.StatusPending, "Storefront order created", uuid.Nil, now); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, orderID)
}

func (s *Service) buildLineItems(ctx context.Context, inputs []CreateItemInput) ([]domain.Item, float64, error) {
	items := make([]domain.Item, 0, len(inputs))
	var subtotal float64

	for _, input := range inputs {
		product, err := s.products.FindByID(ctx, input.ProductID)
		if err != nil {
			return nil, 0, err
		}
		if product.Inventory.Quantity < input.Quantity {
			return nil, 0, domain.ErrInsufficientStock
		}

		unitPrice := roundMoney(product.EffectivePrice())
		lineTotal := roundMoney(unitPrice * float64(input.Quantity))
		subtotal += lineTotal

		skuCode, err := resolveProductSKU(product, input.SkuID)
		if err != nil {
			return nil, 0, err
		}

		items = append(items, domain.Item{
			ProductID:   product.ID,
			ProductName: product.Name,
			ProductSKU:  skuCode,
			Quantity:    input.Quantity,
			UnitPrice:   unitPrice,
			TotalPrice:  lineTotal,
		})
	}

	return items, roundMoney(subtotal), nil
}

func resolveProductSKU(product *domainproduct.Product, skuID *uuid.UUID) (string, error) {
	if skuID != nil {
		for _, sku := range product.SKUs {
			if sku.ID == *skuID {
				return sku.Code, nil
			}
		}
		return "", domain.ErrInvalidSKU
	}
	if len(product.SKUs) > 0 {
		return product.SKUs[0].Code, nil
	}
	return "", nil
}

// UpdateNotesInput holds data for updating internal order notes.
type UpdateNotesInput struct {
	Notes string
}

// UpdateNotes saves internal notes without changing order status.
func (s *Service) UpdateNotes(ctx context.Context, id uuid.UUID, input UpdateNotesInput) (*domain.Order, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	if err := s.repo.UpdateNotes(ctx, id, input.Notes, now); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

// UpdateStatusInput holds data for a manual status update.
type UpdateStatusInput struct {
	Status    string
	Note      string
	ChangedBy uuid.UUID
}

// UpdateStatus transitions an order through the fulfillment workflow.
func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, input UpdateStatusInput) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newStatus, err := domain.ParseStatus(input.Status)
	if err != nil {
		return nil, err
	}

	fromStatus := order.Status
	if err := order.TransitionTo(newStatus); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	order.UpdatedAt = now

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	if err := s.recordHistory(ctx, order.ID, &fromStatus, newStatus, input.Note, input.ChangedBy, now); err != nil {
		return nil, err
	}

	if newStatus == domain.StatusCancelled {
		if err := s.revertOrderEffects(ctx, order); err != nil {
			return nil, err
		}
	}

	return s.repo.FindByID(ctx, id)
}

// CancelUnpaidPayment cancels a pending unpaid order after failed payment.
func (s *Service) CancelUnpaidPayment(ctx context.Context, id uuid.UUID, note string) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if order.Status != domain.StatusPending || order.PaymentStatus != domain.PaymentUnpaid {
		return order, nil
	}

	fromStatus := order.Status
	order.Status = domain.StatusCancelled
	order.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	if err := s.recordHistory(ctx, order.ID, &fromStatus, domain.StatusCancelled, note, uuid.Nil, order.UpdatedAt); err != nil {
		return nil, err
	}

	if err := s.revertOrderEffects(ctx, order); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

// ExpireUnpaidOrders cancels unpaid orders past their payment window and restores stock.
func (s *Service) ExpireUnpaidOrders(ctx context.Context, limit int) (int, error) {
	orders, err := s.repo.FindExpiredUnpaid(ctx, time.Now().UTC(), limit)
	if err != nil {
		return 0, err
	}

	expired := 0
	for _, o := range orders {
		if _, err := s.CancelUnpaidPayment(ctx, o.ID, "Payment window expired"); err != nil {
			continue
		}
		expired++
	}
	return expired, nil
}

func (s *Service) revertOrderEffects(ctx context.Context, order *domain.Order) error {
	if err := s.repo.RestoreInventory(ctx, order.Items); err != nil {
		return err
	}
	if order.CouponID != nil {
		if err := s.repo.DecrementCouponUsage(ctx, *order.CouponID); err != nil {
			return err
		}
	}
	return s.customers.RecordOrderCancelled(ctx, order.CustomerID, order.Total)
}

// Cancel cancels a pending or processing order and restores inventory.
func (s *Service) Cancel(ctx context.Context, id uuid.UUID, changedBy uuid.UUID) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !domain.CanCancel(order.Status) {
		return nil, domain.ErrCannotCancel
	}

	fromStatus := order.Status
	order.Status = domain.StatusCancelled
	order.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	if err := s.recordHistory(ctx, order.ID, &fromStatus, domain.StatusCancelled, "Order cancelled", changedBy, order.UpdatedAt); err != nil {
		return nil, err
	}

	if err := s.revertOrderEffects(ctx, order); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

// RefundInput holds data for refunding an order.
type RefundInput struct {
	Amount    float64
	Reason    string
	ChangedBy uuid.UUID
}

// Refund processes a full or partial refund on a delivered, paid order.
func (s *Service) Refund(ctx context.Context, id uuid.UUID, input RefundInput) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !domain.CanRefund(order.Status, order.PaymentStatus) {
		return nil, domain.ErrCannotRefund
	}

	if input.Amount <= 0 || input.Amount > order.Total {
		return nil, domain.ErrInvalidRefundAmount
	}

	fromStatus := order.Status
	order.Status = domain.StatusRefunded
	order.PaymentStatus = domain.PaymentRefunded
	order.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	note := fmt.Sprintf("Refund of %.2f: %s", input.Amount, input.Reason)
	if err := s.recordHistory(ctx, order.ID, &fromStatus, domain.StatusRefunded, note, input.ChangedBy, order.UpdatedAt); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, id)
}

// ConfirmPaymentInput holds data for a payment gateway callback.
type ConfirmPaymentInput struct {
	OrderID       uuid.UUID
	TransactionID string
}

// ConfirmPayment marks an unpaid order as paid after successful gateway callback.
// Future: transition status to processing and send confirmation email via OnPaymentConfirmed.
func (s *Service) ConfirmPayment(ctx context.Context, input ConfirmPaymentInput) (*domain.Order, error) {
	order, err := s.repo.FindByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}

	if order.Status == domain.StatusCancelled {
		return nil, apperror.Unprocessable("order cannot be paid in its current state")
	}

	if order.PaymentExpiresAt != nil && time.Now().UTC().After(*order.PaymentExpiresAt) {
		return nil, domain.ErrPaymentExpired
	}

	switch order.PaymentStatus {
	case domain.PaymentPaid:
		return nil, domain.ErrPaymentAlreadyPaid
	case domain.PaymentRefunded:
		return nil, apperror.Unprocessable("order cannot be paid in its current state")
	}

	now := time.Now().UTC()
	order.PaymentStatus = domain.PaymentPaid
	order.TransactionID = strings.TrimSpace(input.TransactionID)
	order.UpdatedAt = now

	if err := s.repo.Update(ctx, order); err != nil {
		return nil, err
	}

	return s.repo.FindByID(ctx, input.OrderID)
}

func (s *Service) recordHistory(ctx context.Context, orderID uuid.UUID, from *domain.Status, to domain.Status, note string, changedBy uuid.UUID, at time.Time) error {
	entry := &domain.StatusHistory{
		ID:         uuid.New(),
		OrderID:    orderID,
		FromStatus: from,
		ToStatus:   to,
		Note:       note,
		ChangedBy:  &changedBy,
		CreatedAt:  at,
	}
	return s.repo.AddStatusHistory(ctx, entry)
}

func roundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}
