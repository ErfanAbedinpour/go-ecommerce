package storefront

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	apporder "app/internal/application/order"
	domaincoupon "app/internal/domain/coupon"
	domaincustomer "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	domainproduct "app/internal/domain/product"
	domainsettings "app/internal/domain/settings"
	"app/pkg/pagination"
)

type checkoutProductRepo struct {
	products map[uuid.UUID]*domainproduct.Product
}

func (m *checkoutProductRepo) Create(context.Context, *domainproduct.Product) error { return nil }
func (m *checkoutProductRepo) Update(context.Context, *domainproduct.Product) error { return nil }
func (m *checkoutProductRepo) SoftDelete(context.Context, uuid.UUID) error          { return nil }
func (m *checkoutProductRepo) FindBySlug(context.Context, string) (*domainproduct.Product, error) {
	return nil, domainproduct.ErrNotFound
}
func (m *checkoutProductRepo) FindBySKU(context.Context, string) (*domainproduct.Product, error) {
	return nil, domainproduct.ErrNotFound
}
func (m *checkoutProductRepo) List(context.Context, domainproduct.ListFilter, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *checkoutProductRepo) ListStorefront(context.Context, domainproduct.StoreListFilter, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *checkoutProductRepo) Search(context.Context, string, pagination.Params) ([]domainproduct.Product, int64, error) {
	return nil, 0, nil
}
func (m *checkoutProductRepo) CountActive(context.Context) (int64, error) { return 0, nil }
func (m *checkoutProductRepo) UpdateInventory(context.Context, uuid.UUID, domainproduct.Inventory) error {
	return nil
}
func (m *checkoutProductRepo) ExistsInActiveOrders(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *checkoutProductRepo) CategoryExists(context.Context, uuid.UUID) (bool, error) {
	return true, nil
}
func (m *checkoutProductRepo) GetStats(context.Context) (*domainproduct.Stats, error) {
	return nil, nil
}
func (m *checkoutProductRepo) FindByID(_ context.Context, id uuid.UUID) (*domainproduct.Product, error) {
	p, ok := m.products[id]
	if !ok {
		return nil, domainproduct.ErrNotFound
	}
	return p, nil
}

type checkoutCustomerRepo struct {
	created []*domaincustomer.Customer
}

func (m *checkoutCustomerRepo) Create(_ context.Context, c *domaincustomer.Customer) error {
	m.created = append(m.created, c)
	return nil
}
func (m *checkoutCustomerRepo) FindByID(_ context.Context, id uuid.UUID) (*domaincustomer.Customer, error) {
	for _, c := range m.created {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, domaincustomer.ErrNotFound
}
func (m *checkoutCustomerRepo) FindByEmail(context.Context, string) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (m *checkoutCustomerRepo) FindByUserID(context.Context, uuid.UUID) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (m *checkoutCustomerRepo) List(context.Context, domaincustomer.ListFilter, pagination.Params) ([]domaincustomer.Customer, int64, error) {
	return nil, 0, nil
}
func (m *checkoutCustomerRepo) Update(context.Context, *domaincustomer.Customer) error { return nil }
func (m *checkoutCustomerRepo) Delete(context.Context, uuid.UUID) error                { return nil }
func (m *checkoutCustomerRepo) HasOrders(context.Context, uuid.UUID) (bool, error)     { return false, nil }
func (m *checkoutCustomerRepo) GetLastOrderAt(context.Context, uuid.UUID) (*time.Time, error) {
	return nil, nil
}
func (m *checkoutCustomerRepo) ListAddresses(context.Context, uuid.UUID) ([]domaincustomer.Address, error) {
	return nil, nil
}
func (m *checkoutCustomerRepo) ListOrders(context.Context, uuid.UUID, pagination.Params) ([]domainorder.Summary, int64, error) {
	return nil, 0, nil
}

type checkoutOrderRepo struct {
	orders map[uuid.UUID]*domainorder.Order
}

func (m *checkoutOrderRepo) FindByID(_ context.Context, id uuid.UUID) (*domainorder.Order, error) {
	o, ok := m.orders[id]
	if !ok {
		return nil, domainorder.ErrNotFound
	}
	return o, nil
}
func (m *checkoutOrderRepo) List(context.Context, domainorder.ListFilter, pagination.Params) ([]domainorder.ListItem, int64, error) {
	return nil, 0, nil
}
func (m *checkoutOrderRepo) Create(_ context.Context, o *domainorder.Order) error {
	if m.orders == nil {
		m.orders = make(map[uuid.UUID]*domainorder.Order)
	}
	m.orders[o.ID] = o
	return nil
}
func (m *checkoutOrderRepo) Update(context.Context, *domainorder.Order) error { return nil }
func (m *checkoutOrderRepo) UpdateNotes(context.Context, uuid.UUID, string, time.Time) error {
	return nil
}
func (m *checkoutOrderRepo) AddStatusHistory(context.Context, *domainorder.StatusHistory) error {
	return nil
}
func (m *checkoutOrderRepo) RestoreInventory(context.Context, []domainorder.Item) error { return nil }
func (m *checkoutOrderRepo) NextOrderNumber(context.Context) (string, error) {
	return "ORD-000001", nil
}
func (m *checkoutOrderRepo) IncrementCouponUsage(context.Context, uuid.UUID) error { return nil }

type checkoutCouponRepo struct {
	coupons map[string]*domaincoupon.Coupon
}

func (m *checkoutCouponRepo) Create(context.Context, *domaincoupon.Coupon) error { return nil }
func (m *checkoutCouponRepo) Update(context.Context, *domaincoupon.Coupon) error { return nil }
func (m *checkoutCouponRepo) SoftDelete(context.Context, uuid.UUID) error        { return nil }
func (m *checkoutCouponRepo) FindByID(context.Context, uuid.UUID) (*domaincoupon.Coupon, error) {
	return nil, domaincoupon.ErrNotFound
}
func (m *checkoutCouponRepo) List(context.Context, domaincoupon.ListFilter, pagination.Params) ([]domaincoupon.Coupon, int64, error) {
	return nil, 0, nil
}
func (m *checkoutCouponRepo) SetActive(context.Context, uuid.UUID, bool) error { return nil }
func (m *checkoutCouponRepo) FindByCode(_ context.Context, code string) (*domaincoupon.Coupon, error) {
	c, ok := m.coupons[code]
	if !ok {
		return nil, domaincoupon.ErrNotFound
	}
	return c, nil
}

type checkoutSettingsRepo struct{}

func (checkoutSettingsRepo) Get(context.Context) (*domainsettings.StoreSettings, error) {
	return &domainsettings.StoreSettings{}, nil
}
func (checkoutSettingsRepo) UpdateSite(context.Context, domainsettings.Site) (*domainsettings.Site, error) {
	return nil, nil
}
func (checkoutSettingsRepo) UpdateContact(context.Context, domainsettings.Contact) (*domainsettings.Contact, error) {
	return nil, nil
}
func (checkoutSettingsRepo) UpdateSocial(context.Context, domainsettings.Social) (*domainsettings.Social, error) {
	return nil, nil
}
func (checkoutSettingsRepo) UpdateSEO(context.Context, domainsettings.SEO) (*domainsettings.SEO, error) {
	return nil, nil
}
func (checkoutSettingsRepo) UpdateNavigation(context.Context, []domainsettings.NavItem) ([]domainsettings.NavItem, error) {
	return nil, nil
}
func (checkoutSettingsRepo) UpdateStorefrontNavigation(context.Context, []domainsettings.NavItem) ([]domainsettings.NavItem, error) {
	return nil, nil
}
func (checkoutSettingsRepo) UpdateContactSectionImage(context.Context, string) (string, error) {
	return "", nil
}

func newCheckoutTestService(products map[uuid.UUID]*domainproduct.Product, coupons map[string]*domaincoupon.Coupon) (*Service, *checkoutOrderRepo, *checkoutCustomerRepo) {
	productRepo := &checkoutProductRepo{products: products}
	customerRepo := &checkoutCustomerRepo{}
	orderRepo := &checkoutOrderRepo{}
	couponRepo := &checkoutCouponRepo{coupons: coupons}
	orderSvc := apporder.NewService(orderRepo, productRepo, customerRepo, couponRepo, checkoutSettingsRepo{})
	return NewService(productRepo, nil, orderSvc, couponRepo, customerRepo, checkoutSettingsRepo{}), orderRepo, customerRepo
}

func TestPreviewCheckout_EmptyCart(t *testing.T) {
	svc, _, _ := newCheckoutTestService(nil, nil)

	_, err := svc.PreviewCheckout(context.Background(), PreviewCheckoutInput{})
	if err != domainorder.ErrEmptyOrder {
		t.Fatalf("PreviewCheckout() error = %v, want ErrEmptyOrder", err)
	}
}

func TestPreviewCheckout_CouponApplied(t *testing.T) {
	productID := uuid.New()
	products := map[uuid.UUID]*domainproduct.Product{
		productID: {
			ID:     productID,
			Name:   "Shirt",
			Price:  100000,
			Status: domainproduct.StatusActive,
			Inventory: domainproduct.Inventory{Quantity: 10},
		},
	}
	coupons := map[string]*domaincoupon.Coupon{
		"SAVE10": {
			Code:          "SAVE10",
			DiscountType:  domaincoupon.DiscountTypePercentage,
			DiscountValue: 10,
			IsActive:      true,
		},
	}
	svc, _, _ := newCheckoutTestService(products, coupons)

	out, err := svc.PreviewCheckout(context.Background(), PreviewCheckoutInput{
		Items:      []CheckoutItemInput{{ProductID: productID, Quantity: 2}},
		CouponCode: "save10",
	})
	if err != nil {
		t.Fatalf("PreviewCheckout() error = %v", err)
	}
	if out.Summary.SubtotalToman != 200000 {
		t.Fatalf("subtotal = %d, want 200000", out.Summary.SubtotalToman)
	}
	if out.Summary.DiscountToman != 20000 {
		t.Fatalf("discount = %d, want 20000", out.Summary.DiscountToman)
	}
	if out.Coupon == nil || !out.Coupon.IsValid {
		t.Fatal("expected valid coupon result")
	}
}

func TestPreviewCheckout_InsufficientStock(t *testing.T) {
	productID := uuid.New()
	products := map[uuid.UUID]*domainproduct.Product{
		productID: {
			ID:        productID,
			Name:      "Hat",
			Price:     50000,
			Status:    domainproduct.StatusActive,
			Inventory: domainproduct.Inventory{Quantity: 1},
		},
	}
	svc, _, _ := newCheckoutTestService(products, nil)

	out, err := svc.PreviewCheckout(context.Background(), PreviewCheckoutInput{
		Items: []CheckoutItemInput{{ProductID: productID, Quantity: 3}},
	})
	if err != nil {
		t.Fatalf("PreviewCheckout() error = %v", err)
	}
	if len(out.Items) != 1 || out.Items[0].IsAvailable {
		t.Fatal("expected line item to be unavailable")
	}
}

func TestPlaceCheckout_GuestCustomerCreated(t *testing.T) {
	productID := uuid.New()
	products := map[uuid.UUID]*domainproduct.Product{
		productID: {
			ID:        productID,
			Name:      "Bag",
			Price:     75000,
			Status:    domainproduct.StatusActive,
			Inventory: domainproduct.Inventory{Quantity: 5},
		},
	}
	svc, orderRepo, customerRepo := newCheckoutTestService(products, nil)

	out, err := svc.PlaceCheckout(context.Background(), PlaceCheckoutInput{
		Items: []CheckoutItemInput{{ProductID: productID, Quantity: 1}},
		Customer: CheckoutCustomerInput{
			Email:     "guest@shop.com",
			FirstName: "Guest",
			LastName:  "Shopper",
		},
		ShippingAddress: domainorder.Address{City: "Tehran"},
		BillingAddress:  domainorder.Address{City: "Tehran"},
		PaymentMethod:   "online",
	})
	if err != nil {
		t.Fatalf("PlaceCheckout() error = %v", err)
	}
	if out.PaymentStatus != domainorder.PaymentUnpaid.String() {
		t.Fatalf("payment_status = %q, want unpaid", out.PaymentStatus)
	}
	if len(orderRepo.orders) != 1 {
		t.Fatal("expected order to be created")
	}
	if len(customerRepo.created) < 1 {
		t.Fatal("expected guest customer to be created")
	}
	last := customerRepo.created[len(customerRepo.created)-1]
	if last.Email != "guest@shop.com" || last.Type != domaincustomer.TypeGuest {
		t.Fatalf("unexpected guest customer: %+v", last)
	}
}
