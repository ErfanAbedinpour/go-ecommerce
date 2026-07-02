package storefront

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	apporder "app/internal/application/order"
	domaincustomer "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	domainsettings "app/internal/domain/settings"
	"app/pkg/pagination"
)

func TestGetCheckoutSettings_ReturnsDefaults(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil, nil, checkoutSettingsRepoWithDefaults{}, nil, noopMailer{})

	settings, err := svc.GetCheckoutSettings(context.Background())
	if err != nil {
		t.Fatalf("GetCheckoutSettings: %v", err)
	}
	if settings.MinOrderToman != 100000 {
		t.Errorf("min_order_toman = %d, want 100000", settings.MinOrderToman)
	}
	if len(settings.PaymentMethods) != 2 {
		t.Errorf("payment_methods len = %d, want 2", len(settings.PaymentMethods))
	}
	if !settings.CODEEnabled {
		t.Error("cod_enabled should be true")
	}
}

func TestHandlePaymentCallback_ConfirmsPayment(t *testing.T) {
	orderID := uuid.New()
	orderRepo := &paymentCallbackOrderRepo{
		order: &domainorder.Order{
			ID:            orderID,
			OrderNumber:   "ORD-1001",
			PaymentStatus: domainorder.PaymentUnpaid,
			Total:         250000,
			Customer: &domainorder.CustomerSnapshot{
				Email: "buyer@example.com",
			},
		},
	}
	orderSvc := apporder.NewService(orderRepo, nil, nil, nil, checkoutSettingsRepoWithDefaults{}, 25*time.Hour)
	svc := NewService(nil, nil, nil, orderSvc, nil, nil, nil, checkoutSettingsRepoWithDefaults{}, nil, noopMailer{})

	out, err := svc.HandlePaymentCallback(context.Background(), PaymentCallbackInput{
		OrderID:   orderID,
		Authority: "A00000000000000000000000000123456789",
		Status:    "OK",
	}, "")
	if err != nil {
		t.Fatalf("HandlePaymentCallback: %v", err)
	}
	if out.PaymentStatus != "paid" {
		t.Errorf("payment_status = %q, want paid", out.PaymentStatus)
	}
	if orderRepo.order.PaymentStatus != domainorder.PaymentPaid {
		t.Errorf("order payment_status = %q, want paid", orderRepo.order.PaymentStatus)
	}
}

func TestHandlePaymentCallback_AlreadyPaidConflict(t *testing.T) {
	orderID := uuid.New()
	orderRepo := &paymentCallbackOrderRepo{
		order: &domainorder.Order{
			ID:            orderID,
			PaymentStatus: domainorder.PaymentPaid,
		},
	}
	orderSvc := apporder.NewService(orderRepo, nil, nil, nil, checkoutSettingsRepoWithDefaults{}, 25*time.Hour)
	svc := NewService(nil, nil, nil, orderSvc, nil, nil, nil, checkoutSettingsRepoWithDefaults{}, nil, noopMailer{})

	_, err := svc.HandlePaymentCallback(context.Background(), PaymentCallbackInput{
		OrderID:   orderID,
		Authority: "A00000000000000000000000000123456789",
		Status:    "OK",
	}, "")
	if err != domainorder.ErrPaymentAlreadyPaid {
		t.Fatalf("expected ErrPaymentAlreadyPaid, got %v", err)
	}
}

func TestHandlePaymentCallback_FailedPaymentCancelsOrder(t *testing.T) {
	orderID := uuid.New()
	customerID := uuid.New()
	orderRepo := &paymentCallbackOrderRepo{
		order: &domainorder.Order{
			ID:            orderID,
			CustomerID:    customerID,
			Status:        domainorder.StatusPending,
			PaymentStatus: domainorder.PaymentUnpaid,
			Items:         []domainorder.Item{{ProductID: uuid.New(), Quantity: 1}},
		},
	}
	customerRepo := &paymentCallbackCustomerRepo{}
	orderSvc := apporder.NewService(orderRepo, nil, customerRepo, nil, checkoutSettingsRepoWithDefaults{}, 25*time.Hour)
	svc := NewService(nil, nil, nil, orderSvc, nil, customerRepo, nil, checkoutSettingsRepoWithDefaults{}, nil, noopMailer{})

	out, err := svc.HandlePaymentCallback(context.Background(), PaymentCallbackInput{
		OrderID: orderID,
		Status:  "NOK",
	}, "")
	if err != nil {
		t.Fatalf("HandlePaymentCallback: %v", err)
	}
	if out.PaymentStatus != "unpaid" {
		t.Errorf("payment_status = %q, want unpaid", out.PaymentStatus)
	}
	if orderRepo.order.Status != domainorder.StatusCancelled {
		t.Errorf("order status = %q, want cancelled", orderRepo.order.Status)
	}
	if !orderRepo.restored {
		t.Error("expected inventory restore on failed payment")
	}
}

type checkoutSettingsRepoWithDefaults struct{}

func (checkoutSettingsRepoWithDefaults) Get(context.Context) (*domainsettings.StoreSettings, error) {
	return &domainsettings.StoreSettings{
		Checkout: domainsettings.DefaultCheckout(),
	}, nil
}
func (checkoutSettingsRepoWithDefaults) UpdateSite(context.Context, domainsettings.Site) (*domainsettings.Site, error) {
	return nil, nil
}
func (checkoutSettingsRepoWithDefaults) UpdateContact(context.Context, domainsettings.Contact) (*domainsettings.Contact, error) {
	return nil, nil
}
func (checkoutSettingsRepoWithDefaults) UpdateSocial(context.Context, domainsettings.Social) (*domainsettings.Social, error) {
	return nil, nil
}
func (checkoutSettingsRepoWithDefaults) UpdateSEO(context.Context, domainsettings.SEO) (*domainsettings.SEO, error) {
	return nil, nil
}
func (checkoutSettingsRepoWithDefaults) UpdateNavigation(context.Context, []domainsettings.NavItem) ([]domainsettings.NavItem, error) {
	return nil, nil
}
func (checkoutSettingsRepoWithDefaults) UpdateStorefrontNavigation(context.Context, []domainsettings.NavItem) ([]domainsettings.NavItem, error) {
	return nil, nil
}
func (checkoutSettingsRepoWithDefaults) UpdateContactSectionImage(context.Context, string) (string, error) {
	return "", nil
}

type paymentCallbackOrderRepo struct {
	order    *domainorder.Order
	restored bool
}

func (m *paymentCallbackOrderRepo) FindByID(_ context.Context, id uuid.UUID) (*domainorder.Order, error) {
	if m.order != nil && m.order.ID == id {
		return m.order, nil
	}
	return nil, domainorder.ErrNotFound
}
func (m *paymentCallbackOrderRepo) List(context.Context, domainorder.ListFilter, pagination.Params) ([]domainorder.ListItem, int64, error) {
	return nil, 0, nil
}
func (m *paymentCallbackOrderRepo) Create(context.Context, *domainorder.Order) error { return nil }
func (m *paymentCallbackOrderRepo) Update(_ context.Context, o *domainorder.Order) error {
	m.order = o
	return nil
}
func (m *paymentCallbackOrderRepo) UpdateNotes(context.Context, uuid.UUID, string, time.Time) error {
	return nil
}
func (m *paymentCallbackOrderRepo) AddStatusHistory(context.Context, *domainorder.StatusHistory) error {
	return nil
}
func (m *paymentCallbackOrderRepo) RestoreInventory(context.Context, []domainorder.Item) error {
	m.restored = true
	return nil
}
func (m *paymentCallbackOrderRepo) NextOrderNumber(context.Context) (string, error) {
	return "ORD-1001", nil
}
func (m *paymentCallbackOrderRepo) IncrementCouponUsage(context.Context, uuid.UUID) error { return nil }
func (m *paymentCallbackOrderRepo) DecrementCouponUsage(context.Context, uuid.UUID) error { return nil }
func (m *paymentCallbackOrderRepo) FindExpiredUnpaid(context.Context, time.Time, int) ([]domainorder.Order, error) {
	return nil, nil
}
func (m *paymentCallbackOrderRepo) CountByStatus(_ context.Context, status domainorder.Status) (int64, error) {
	if m.order != nil && m.order.Status == status {
		return 1, nil
	}
	return 0, nil
}

type paymentCallbackCustomerRepo struct{}

func (paymentCallbackCustomerRepo) Create(context.Context, *domaincustomer.Customer) error { return nil }
func (paymentCallbackCustomerRepo) FindByID(context.Context, uuid.UUID) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (paymentCallbackCustomerRepo) FindByEmail(context.Context, string) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (paymentCallbackCustomerRepo) FindGuestByEmail(context.Context, string) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (paymentCallbackCustomerRepo) FindGuestByPhone(context.Context, string) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (paymentCallbackCustomerRepo) FindRegisteredByPhone(context.Context, string) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (paymentCallbackCustomerRepo) FindByUserID(context.Context, uuid.UUID) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (paymentCallbackCustomerRepo) List(context.Context, domaincustomer.ListFilter, pagination.Params) ([]domaincustomer.Customer, int64, error) {
	return nil, 0, nil
}
func (paymentCallbackCustomerRepo) Update(context.Context, *domaincustomer.Customer) error { return nil }
func (paymentCallbackCustomerRepo) Delete(context.Context, uuid.UUID) error                { return nil }
func (paymentCallbackCustomerRepo) HasOrders(context.Context, uuid.UUID) (bool, error)     { return false, nil }
func (paymentCallbackCustomerRepo) GetLastOrderAt(context.Context, uuid.UUID) (*time.Time, error) {
	return nil, nil
}
func (paymentCallbackCustomerRepo) ListAddresses(context.Context, uuid.UUID) ([]domaincustomer.Address, error) {
	return nil, nil
}
func (paymentCallbackCustomerRepo) ReplaceAddresses(context.Context, uuid.UUID, []domaincustomer.Address) error {
	return nil
}
func (paymentCallbackCustomerRepo) ListOrders(context.Context, uuid.UUID, pagination.Params) ([]domainorder.Summary, int64, error) {
	return nil, 0, nil
}
func (paymentCallbackCustomerRepo) Count(context.Context) (int64, error) { return 0, nil }
func (paymentCallbackCustomerRepo) RecordOrderPlaced(context.Context, uuid.UUID, float64, time.Time) error {
	return nil
}
func (paymentCallbackCustomerRepo) RecordOrderCancelled(context.Context, uuid.UUID, float64) error {
	return nil
}
