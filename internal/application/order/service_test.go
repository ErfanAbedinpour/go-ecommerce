package order

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domaincoupon "app/internal/domain/coupon"
	domaincustomer "app/internal/domain/customer"
	domain "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	domainsettings "app/internal/domain/settings"
	"app/pkg/pagination"
)

type mockRepo struct {
	orders    map[uuid.UUID]*domain.Order
	history   []domain.StatusHistory
	restored  bool
	created   bool
	notes     string
	orderNum  string
}

func newMockRepo() *mockRepo {
	return &mockRepo{orders: make(map[uuid.UUID]*domain.Order), orderNum: "ORD-000001"}
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	cp := *o
	return &cp, nil
}

func (m *mockRepo) List(_ context.Context, _ domain.ListFilter, page pagination.Params) ([]domain.ListItem, int64, error) {
	items := make([]domain.ListItem, 0, len(m.orders))
	for _, o := range m.orders {
		items = append(items, domain.ListItem{
			Summary: domain.Summary{
				ID: o.ID, OrderNumber: o.OrderNumber, Status: o.Status.String(),
				PaymentStatus: o.PaymentStatus.String(), Total: o.Total,
			},
			CustomerID: o.CustomerID,
		})
	}
	return items, int64(len(items)), nil
}

func (m *mockRepo) Create(_ context.Context, o *domain.Order) error {
	m.created = true
	m.orders[o.ID] = o
	return nil
}

func (m *mockRepo) Update(_ context.Context, o *domain.Order) error {
	m.orders[o.ID] = o
	return nil
}

func (m *mockRepo) UpdateNotes(_ context.Context, id uuid.UUID, notes string, _ time.Time) error {
	o, ok := m.orders[id]
	if !ok {
		return domain.ErrNotFound
	}
	o.Notes = notes
	m.notes = notes
	return nil
}

func (m *mockRepo) AddStatusHistory(_ context.Context, entry *domain.StatusHistory) error {
	m.history = append(m.history, *entry)
	return nil
}

func (m *mockRepo) RestoreInventory(_ context.Context, _ []domain.Item) error {
	m.restored = true
	return nil
}

func (m *mockRepo) NextOrderNumber(_ context.Context) (string, error) {
	return m.orderNum, nil
}

func (m *mockRepo) IncrementCouponUsage(_ context.Context, _ uuid.UUID) error {
	return nil
}

type mockProductRepo struct {
	products map[uuid.UUID]*domainproduct.Product
}

func (m *mockProductRepo) Create(context.Context, *domainproduct.Product) error { return nil }
func (m *mockProductRepo) Update(context.Context, *domainproduct.Product) error { return nil }
func (m *mockProductRepo) SoftDelete(context.Context, uuid.UUID) error        { return nil }
func (m *mockProductRepo) FindBySlug(context.Context, string) (*domainproduct.Product, error) {
	return nil, domainproduct.ErrNotFound
}
func (m *mockProductRepo) FindBySKU(context.Context, string) (*domainproduct.Product, error) {
	return nil, domainproduct.ErrNotFound
}
func (m *mockProductRepo) List(context.Context, domainproduct.ListFilter, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) Search(context.Context, string, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *mockProductRepo) UpdateInventory(context.Context, uuid.UUID, domainproduct.Inventory) error {
	return nil
}
func (m *mockProductRepo) ExistsInActiveOrders(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockProductRepo) CategoryExists(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}
func (m *mockProductRepo) GetStats(context.Context) (*domainproduct.Stats, error) {
	return nil, nil
}
func (m *mockProductRepo) FindByID(_ context.Context, id uuid.UUID) (*domainproduct.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, domainproduct.ErrNotFound
	}
	return p, nil
}

type mockCustomerRepo struct {
	customers map[uuid.UUID]*domaincustomer.Customer
}

func (m *mockCustomerRepo) FindByID(_ context.Context, id uuid.UUID) (*domaincustomer.Customer, error) {
	c, ok := m.customers[id]
	if !ok {
		return nil, domaincustomer.ErrNotFound
	}
	return c, nil
}
func (m *mockCustomerRepo) List(context.Context, domaincustomer.ListFilter, pagination.Params) ([]domaincustomer.Customer, int64, error) {
	return nil, 0, nil
}
func (m *mockCustomerRepo) ListAddresses(context.Context, uuid.UUID) ([]domaincustomer.Address, error) {
	return nil, nil
}
func (m *mockCustomerRepo) ListOrders(context.Context, uuid.UUID, pagination.Params) ([]domain.Summary, int64, error) {
	return nil, 0, nil
}
func (m *mockCustomerRepo) FindByEmail(context.Context, string) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (m *mockCustomerRepo) Update(context.Context, *domaincustomer.Customer) error { return nil }
func (m *mockCustomerRepo) Delete(context.Context, uuid.UUID) error                { return nil }
func (m *mockCustomerRepo) HasOrders(context.Context, uuid.UUID) (bool, error)     { return false, nil }
func (m *mockCustomerRepo) GetLastOrderAt(context.Context, uuid.UUID) (*time.Time, error) {
	return nil, nil
}

type mockCouponRepo struct {
	coupons map[string]*domaincoupon.Coupon
}

func (m *mockCouponRepo) Create(context.Context, *domaincoupon.Coupon) error { return nil }
func (m *mockCouponRepo) Update(context.Context, *domaincoupon.Coupon) error { return nil }
func (m *mockCouponRepo) SoftDelete(context.Context, uuid.UUID) error        { return nil }
func (m *mockCouponRepo) FindByID(context.Context, uuid.UUID) (*domaincoupon.Coupon, error) {
	return nil, domaincoupon.ErrNotFound
}
func (m *mockCouponRepo) List(context.Context, domaincoupon.ListFilter, pagination.Params) ([]domaincoupon.Coupon, int64, error) {
	return nil, 0, nil
}
func (m *mockCouponRepo) SetActive(context.Context, uuid.UUID, bool) error { return nil }
func (m *mockCouponRepo) FindByCode(_ context.Context, code string) (*domaincoupon.Coupon, error) {
	c, ok := m.coupons[code]
	if !ok {
		return nil, domaincoupon.ErrNotFound
	}
	return c, nil
}

type mockSettingsRepo struct {
	store *domainsettings.StoreSettings
}

func (m *mockSettingsRepo) Get(context.Context) (*domainsettings.StoreSettings, error) {
	return m.store, nil
}
func (m *mockSettingsRepo) UpdateSite(context.Context, domainsettings.Site) (*domainsettings.Site, error) {
	return nil, nil
}
func (m *mockSettingsRepo) UpdateContact(context.Context, domainsettings.Contact) (*domainsettings.Contact, error) {
	return nil, nil
}
func (m *mockSettingsRepo) UpdateSocial(context.Context, domainsettings.Social) (*domainsettings.Social, error) {
	return nil, nil
}
func (m *mockSettingsRepo) UpdateSEO(context.Context, domainsettings.SEO) (*domainsettings.SEO, error) {
	return nil, nil
}
func (m *mockSettingsRepo) UpdateNavigation(context.Context, []domainsettings.NavItem) ([]domainsettings.NavItem, error) {
	return nil, nil
}

func newTestService(repo *mockRepo) *Service {
	return NewService(repo, &mockProductRepo{products: map[uuid.UUID]*domainproduct.Product{}},
		&mockCustomerRepo{customers: map[uuid.UUID]*domaincustomer.Customer{}},
		&mockCouponRepo{coupons: map[string]*domaincoupon.Coupon{}},
		&mockSettingsRepo{store: &domainsettings.StoreSettings{Site: domainsettings.Site{Name: "Shop"}}})
}

func TestService_UpdateStatus(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	id := uuid.New()
	adminID := uuid.New()

	repo.orders[id] = &domain.Order{
		ID: id, Status: domain.StatusPending, PaymentStatus: domain.PaymentUnpaid,
		Items: []domain.Item{{ProductID: uuid.New(), Quantity: 2}},
	}

	order, err := svc.UpdateStatus(context.Background(), id, UpdateStatusInput{
		Status: "processing", Note: "Started", ChangedBy: adminID,
	})
	if err != nil {
		t.Fatalf("UpdateStatus() error = %v", err)
	}
	if order.Status != domain.StatusProcessing {
		t.Errorf("status = %q, want processing", order.Status)
	}
}

func TestService_UpdateStatus_InvalidTransition(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	id := uuid.New()

	repo.orders[id] = &domain.Order{ID: id, Status: domain.StatusShipped, PaymentStatus: domain.PaymentPaid}

	_, err := svc.UpdateStatus(context.Background(), id, UpdateStatusInput{
		Status: "cancelled", ChangedBy: uuid.New(),
	})
	if err != domain.ErrInvalidStatusTransition {
		t.Errorf("expected invalid transition, got %v", err)
	}
}

func TestService_Cancel_RestoresInventory(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	id := uuid.New()

	repo.orders[id] = &domain.Order{
		ID: id, Status: domain.StatusPending, PaymentStatus: domain.PaymentUnpaid,
		Items: []domain.Item{{ProductID: uuid.New(), Quantity: 1}},
	}

	_, err := svc.Cancel(context.Background(), id, uuid.New())
	if err != nil {
		t.Fatalf("Cancel() error = %v", err)
	}
	if !repo.restored {
		t.Error("expected inventory restore")
	}
}

func TestService_Cancel_NotAllowed(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	id := uuid.New()
	repo.orders[id] = &domain.Order{ID: id, Status: domain.StatusShipped}

	_, err := svc.Cancel(context.Background(), id, uuid.New())
	if err != domain.ErrCannotCancel {
		t.Errorf("expected cannot cancel, got %v", err)
	}
}

func TestService_Refund(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	id := uuid.New()

	repo.orders[id] = &domain.Order{
		ID: id, Status: domain.StatusDelivered, PaymentStatus: domain.PaymentPaid, Total: 100,
	}

	order, err := svc.Refund(context.Background(), id, RefundInput{
		Amount: 50, Reason: "Partial refund", ChangedBy: uuid.New(),
	})
	if err != nil {
		t.Fatalf("Refund() error = %v", err)
	}
	if order.Status != domain.StatusRefunded {
		t.Errorf("status = %q, want refunded", order.Status)
	}
	if order.PaymentStatus != domain.PaymentRefunded {
		t.Errorf("payment = %q, want refunded", order.PaymentStatus)
	}
}

func TestService_Refund_InvalidAmount(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	id := uuid.New()
	repo.orders[id] = &domain.Order{
		ID: id, Status: domain.StatusDelivered, PaymentStatus: domain.PaymentPaid, Total: 100,
	}

	_, err := svc.Refund(context.Background(), id, RefundInput{Amount: 200, Reason: "Too much", ChangedBy: uuid.New()})
	if err != domain.ErrInvalidRefundAmount {
		t.Errorf("expected invalid amount, got %v", err)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc := newTestService(newMockRepo())
	_, err := svc.GetByID(context.Background(), uuid.New())
	if err != domain.ErrNotFound {
		t.Errorf("expected not found, got %v", err)
	}
}

func TestService_Create(t *testing.T) {
	repo := newMockRepo()
	customerID := uuid.New()
	productID := uuid.New()

	svc := NewService(repo,
		&mockProductRepo{products: map[uuid.UUID]*domainproduct.Product{
			productID: {
				ID: productID, Name: "Shirt", SKU: "SHIRT-1", Price: 25,
				Inventory: domainproduct.Inventory{Quantity: 10},
			},
		}},
		&mockCustomerRepo{customers: map[uuid.UUID]*domaincustomer.Customer{
			customerID: {ID: customerID},
		}},
		&mockCouponRepo{coupons: map[string]*domaincoupon.Coupon{}},
		&mockSettingsRepo{store: &domainsettings.StoreSettings{}},
	)

	order, err := svc.Create(context.Background(), CreateInput{
		CustomerID: customerID,
		Items:      []CreateItemInput{{ProductID: productID, Quantity: 2}},
		BillingAddress: domain.Address{
			Street: "1 Main", City: "NYC", PostalCode: "10001", Country: "US",
		},
		ShippingAddress: domain.Address{
			Street: "1 Main", City: "NYC", PostalCode: "10001", Country: "US",
		},
		PaymentMethod: "card",
		TransactionID: "TXN-123",
		PaymentStatus: "paid",
		ChangedBy:     uuid.New(),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !repo.created {
		t.Error("expected repo.Create to be called")
	}
	if order.Subtotal != 50 {
		t.Errorf("subtotal = %v, want 50", order.Subtotal)
	}
	if order.PaymentMethod != "card" {
		t.Errorf("payment_method = %q, want card", order.PaymentMethod)
	}
}

func TestService_Create_EmptyItems(t *testing.T) {
	svc := newTestService(newMockRepo())
	_, err := svc.Create(context.Background(), CreateInput{CustomerID: uuid.New(), ChangedBy: uuid.New()})
	if err != domain.ErrEmptyOrder {
		t.Errorf("expected empty order error, got %v", err)
	}
}

func TestService_UpdateNotes(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	id := uuid.New()
	repo.orders[id] = &domain.Order{ID: id, Status: domain.StatusPending}

	order, err := svc.UpdateNotes(context.Background(), id, UpdateNotesInput{Notes: "VIP customer"})
	if err != nil {
		t.Fatalf("UpdateNotes() error = %v", err)
	}
	if order.Notes != "VIP customer" {
		t.Errorf("notes = %q, want VIP customer", order.Notes)
	}
}

func TestService_GetInvoice(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)
	id := uuid.New()
	repo.orders[id] = &domain.Order{
		ID: id, OrderNumber: "ORD-000001", Status: domain.StatusPending,
	}

	invoice, err := svc.GetInvoice(context.Background(), id)
	if err != nil {
		t.Fatalf("GetInvoice() error = %v", err)
	}
	if invoice.InvoiceNumber != "ORD-000001" {
		t.Errorf("invoice number = %q, want ORD-000001", invoice.InvoiceNumber)
	}
	if invoice.Store.Name != "Shop" {
		t.Errorf("store name = %q, want Shop", invoice.Store.Name)
	}
}
