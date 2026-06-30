package storefront

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	domaincustomer "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	domainsettings "app/internal/domain/settings"
	"app/pkg/pagination"
)

type accountCustomerRepo struct {
	byUser    map[uuid.UUID]*domaincustomer.Customer
	addresses map[uuid.UUID][]domaincustomer.Address
}

func newAccountCustomerRepo(customer *domaincustomer.Customer) *accountCustomerRepo {
	repo := &accountCustomerRepo{
		byUser:    make(map[uuid.UUID]*domaincustomer.Customer),
		addresses: make(map[uuid.UUID][]domaincustomer.Address),
	}
	if customer != nil && customer.UserID != nil {
		cp := *customer
		repo.byUser[*customer.UserID] = &cp
	}
	return repo
}

func (m *accountCustomerRepo) Create(context.Context, *domaincustomer.Customer) error { return nil }
func (m *accountCustomerRepo) FindByID(context.Context, uuid.UUID) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (m *accountCustomerRepo) FindByEmail(context.Context, string) (*domaincustomer.Customer, error) {
	return nil, domaincustomer.ErrNotFound
}
func (m *accountCustomerRepo) FindByUserID(_ context.Context, userID uuid.UUID) (*domaincustomer.Customer, error) {
	c, ok := m.byUser[userID]
	if !ok {
		return nil, domaincustomer.ErrNotFound
	}
	cp := *c
	return &cp, nil
}
func (m *accountCustomerRepo) List(context.Context, domaincustomer.ListFilter, pagination.Params) ([]domaincustomer.Customer, int64, error) {
	return nil, 0, nil
}
func (m *accountCustomerRepo) Update(_ context.Context, c *domaincustomer.Customer) error {
	if c.UserID == nil {
		return domaincustomer.ErrNotFound
	}
	cp := *c
	m.byUser[*c.UserID] = &cp
	return nil
}
func (m *accountCustomerRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (m *accountCustomerRepo) HasOrders(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}
func (m *accountCustomerRepo) GetLastOrderAt(context.Context, uuid.UUID) (*time.Time, error) {
	return nil, nil
}
func (m *accountCustomerRepo) ListAddresses(_ context.Context, customerID uuid.UUID) ([]domaincustomer.Address, error) {
	return m.addresses[customerID], nil
}
func (m *accountCustomerRepo) ReplaceAddresses(_ context.Context, customerID uuid.UUID, addresses []domaincustomer.Address) error {
	cp := make([]domaincustomer.Address, len(addresses))
	copy(cp, addresses)
	m.addresses[customerID] = cp
	return nil
}
func (m *accountCustomerRepo) ListOrders(context.Context, uuid.UUID, pagination.Params) ([]domainorder.Summary, int64, error) {
	return nil, 0, nil
}

type accountSettingsRepo struct {
	settings domainsettings.StoreSettings
}

func (m *accountSettingsRepo) Get(context.Context) (*domainsettings.StoreSettings, error) {
	cp := m.settings
	return &cp, nil
}
func (m *accountSettingsRepo) UpdateSite(context.Context, domainsettings.Site) (*domainsettings.Site, error) {
	return nil, nil
}
func (m *accountSettingsRepo) UpdateContact(context.Context, domainsettings.Contact) (*domainsettings.Contact, error) {
	return nil, nil
}
func (m *accountSettingsRepo) UpdateSocial(context.Context, domainsettings.Social) (*domainsettings.Social, error) {
	return nil, nil
}
func (m *accountSettingsRepo) UpdateSEO(context.Context, domainsettings.SEO) (*domainsettings.SEO, error) {
	return nil, nil
}
func (m *accountSettingsRepo) UpdateNavigation(context.Context, []domainsettings.NavItem) ([]domainsettings.NavItem, error) {
	return nil, nil
}
func (m *accountSettingsRepo) UpdateStorefrontNavigation(context.Context, []domainsettings.NavItem) ([]domainsettings.NavItem, error) {
	return nil, nil
}
func (m *accountSettingsRepo) UpdateContactSectionImage(context.Context, string) (string, error) {
	return "", nil
}

func TestGetAccountProfile(t *testing.T) {
	userID := uuid.New()
	customerID := uuid.New()
	customer := &domaincustomer.Customer{
		ID:          customerID,
		UserID:      &userID,
		Email:       "user@shop.com",
		FirstName:   "Ali",
		LastName:    "Rezaei",
		Phone:       "09121234567",
		TotalOrders: 3,
		TotalSpent:  1250000,
		CreatedAt:   time.Now().UTC(),
	}

	customers := newAccountCustomerRepo(customer)
	customers.addresses[customerID] = []domaincustomer.Address{
		{
			ID:         uuid.New(),
			CustomerID: customerID,
			Type:       domaincustomer.AddressShipping,
			Street:     "Valiasr St",
			City:       "Tehran",
			PostalCode: "1234567890",
			Country:    "IR",
			IsDefault:  true,
		},
	}

	svc := NewService(nil, nil, nil, nil, nil, customers, &accountSettingsRepo{}, noopMailer{})

	profile, err := svc.GetAccountProfile(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetAccountProfile() error = %v", err)
	}
	if profile.Email != "user@shop.com" || profile.Stats.TotalOrders != 3 {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if len(profile.Addresses) != 1 {
		t.Fatalf("expected 1 address, got %d", len(profile.Addresses))
	}
}

func TestUpdateAccountProfile(t *testing.T) {
	userID := uuid.New()
	customerID := uuid.New()
	customer := &domaincustomer.Customer{
		ID:        customerID,
		UserID:    &userID,
		Email:     "user@shop.com",
		FirstName: "Ali",
		LastName:  "Rezaei",
		CreatedAt: time.Now().UTC(),
	}

	customers := newAccountCustomerRepo(customer)
	svc := NewService(nil, nil, nil, nil, nil, customers, &accountSettingsRepo{}, noopMailer{})

	profile, err := svc.UpdateAccountProfile(context.Background(), userID, UpdateAccountProfileInput{
		FirstName: "Hassan",
		LastName:  "Karimi",
		Phone:     "09129876543",
		Addresses: []UpdateAccountAddressInput{
			{
				Type:       "shipping",
				Street:     "Azadi Blvd",
				City:       "Tehran",
				PostalCode: "1111111111",
				Country:    "IR",
				IsDefault:  true,
			},
		},
	})
	if err != nil {
		t.Fatalf("UpdateAccountProfile() error = %v", err)
	}
	if profile.FirstName != "Hassan" || profile.Phone != "09129876543" {
		t.Fatalf("unexpected profile: %+v", profile)
	}
	if len(profile.Addresses) != 1 || profile.Addresses[0].Street != "Azadi Blvd" {
		t.Fatalf("unexpected addresses: %+v", profile.Addresses)
	}
}

func TestGetStoreNavigation_FiltersInactive(t *testing.T) {
	svc := NewService(nil, nil, nil, nil, nil, nil, &accountSettingsRepo{
		settings: domainsettings.StoreSettings{
			StorefrontNavigation: []domainsettings.NavItem{
				{ID: "1", Label: "Active", URL: "/products", SortOrder: 1, IsActive: true},
				{ID: "2", Label: "Hidden", URL: "/hidden", SortOrder: 2, IsActive: false},
			},
		},
	}, noopMailer{})

	navigation, err := svc.GetStoreNavigation(context.Background())
	if err != nil {
		t.Fatalf("GetStoreNavigation() error = %v", err)
	}
	if len(navigation.Items) != 1 || navigation.Items[0].Label != "Active" {
		t.Fatalf("unexpected navigation: %+v", navigation.Items)
	}
}
