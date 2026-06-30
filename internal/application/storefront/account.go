package storefront

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	domaincustomer "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	"app/pkg/pagination"
)

// AccountAddress is a customer address on the storefront account profile.
type AccountAddress struct {
	ID         uuid.UUID `json:"id"`
	Type       string    `json:"type"`
	Street     string    `json:"street"`
	City       string    `json:"city"`
	State      string    `json:"state,omitempty"`
	PostalCode string    `json:"postal_code"`
	Country    string    `json:"country"`
	IsDefault  bool      `json:"is_default"`
}

// AccountProfileStats holds aggregate purchase statistics for the account.
type AccountProfileStats struct {
	TotalOrders     int   `json:"total_orders"`
	TotalSpentToman int64 `json:"total_spent_toman"`
}

// AccountProfile is the authenticated customer profile.
type AccountProfile struct {
	ID        uuid.UUID           `json:"id"`
	Email     string              `json:"email"`
	FirstName string              `json:"first_name"`
	LastName  string              `json:"last_name"`
	FullName  string              `json:"full_name"`
	Phone     string              `json:"phone,omitempty"`
	Addresses []AccountAddress    `json:"addresses"`
	Stats     AccountProfileStats   `json:"stats"`
	CreatedAt string              `json:"created_at"`
}

// UpdateAccountProfileInput holds editable profile fields.
type UpdateAccountProfileInput struct {
	FirstName string
	LastName  string
	Phone     string
	Addresses []UpdateAccountAddressInput
}

// UpdateAccountAddressInput holds editable address fields.
type UpdateAccountAddressInput struct {
	ID         *uuid.UUID
	Type       string
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
	IsDefault  bool
}

// GetAccountProfile returns the profile for the authenticated customer.
func (s *Service) GetAccountProfile(ctx context.Context, userID uuid.UUID) (*AccountProfile, error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	addresses, err := s.customers.ListAddresses(ctx, customer.ID)
	if err != nil {
		return nil, err
	}

	return toAccountProfile(customer, addresses), nil
}

// UpdateAccountProfile updates profile fields and replaces saved addresses.
func (s *Service) UpdateAccountProfile(ctx context.Context, userID uuid.UUID, input UpdateAccountProfileInput) (*AccountProfile, error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	customer.FirstName = strings.TrimSpace(input.FirstName)
	customer.LastName = strings.TrimSpace(input.LastName)
	customer.Phone = strings.TrimSpace(input.Phone)
	customer.UpdatedAt = time.Now().UTC()

	if err := s.customers.Update(ctx, customer); err != nil {
		return nil, err
	}

	addresses := toDomainAddresses(customer.ID, input.Addresses)
	if err := s.customers.ReplaceAddresses(ctx, customer.ID, addresses); err != nil {
		return nil, err
	}

	return s.GetAccountProfile(ctx, userID)
}

func toAccountProfile(customer *domaincustomer.Customer, addresses []domaincustomer.Address) *AccountProfile {
	profileAddresses := make([]AccountAddress, len(addresses))
	for i, address := range addresses {
		profileAddresses[i] = AccountAddress{
			ID:         address.ID,
			Type:       address.Type.String(),
			Street:     address.Street,
			City:       address.City,
			State:      address.State,
			PostalCode: address.PostalCode,
			Country:    address.Country,
			IsDefault:  address.IsDefault,
		}
	}

	return &AccountProfile{
		ID:        customer.ID,
		Email:     customer.Email,
		FirstName: customer.FirstName,
		LastName:  customer.LastName,
		FullName:  customer.FullName(),
		Phone:     customer.Phone,
		Addresses: profileAddresses,
		Stats: AccountProfileStats{
			TotalOrders:     customer.TotalOrders,
			TotalSpentToman: toMoneyToman(customer.TotalSpent),
		},
		CreatedAt: customer.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func toDomainAddresses(customerID uuid.UUID, inputs []UpdateAccountAddressInput) []domaincustomer.Address {
	addresses := make([]domaincustomer.Address, len(inputs))
	defaultAssigned := false

	for i, input := range inputs {
		addressType := domaincustomer.AddressType(strings.TrimSpace(input.Type))
		if addressType == "" {
			addressType = domaincustomer.AddressShipping
		}

		country := strings.TrimSpace(input.Country)
		if country == "" {
			country = "IR"
		}

		id := uuid.New()
		if input.ID != nil {
			id = *input.ID
		}

		isDefault := input.IsDefault
		if isDefault {
			if defaultAssigned {
				isDefault = false
			} else {
				defaultAssigned = true
			}
		}

		addresses[i] = domaincustomer.Address{
			ID:         id,
			CustomerID: customerID,
			Type:       addressType,
			Street:     strings.TrimSpace(input.Street),
			City:       strings.TrimSpace(input.City),
			State:      strings.TrimSpace(input.State),
			PostalCode: strings.TrimSpace(input.PostalCode),
			Country:    country,
			IsDefault:  isDefault,
		}
	}

	return addresses
}

// AccountOrderSummary is a customer-facing order list item.
type AccountOrderSummary struct {
	ID            uuid.UUID `json:"id"`
	OrderNumber   string    `json:"order_number"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	TotalToman    int64     `json:"total_toman"`
	ItemCount     int       `json:"item_count"`
	CreatedAt     string    `json:"created_at"`
}

// ListAccountOrders returns paginated orders for the authenticated customer.
func (s *Service) ListAccountOrders(ctx context.Context, userID uuid.UUID, page pagination.Params) (pagination.Paginated[AccountOrderSummary], error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return pagination.Paginated[AccountOrderSummary]{}, err
	}

	summaries, total, err := s.customers.ListOrders(ctx, customer.ID, page)
	if err != nil {
		return pagination.Paginated[AccountOrderSummary]{}, err
	}

	items := make([]AccountOrderSummary, len(summaries))
	for i, summary := range summaries {
		items[i] = AccountOrderSummary{
			ID:            summary.ID,
			OrderNumber:   summary.OrderNumber,
			Status:        summary.Status,
			PaymentStatus: summary.PaymentStatus,
			TotalToman:    toMoneyToman(summary.Total),
			ItemCount:     summary.ItemCount,
			CreatedAt:     summary.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		}
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

// GetAccountOrder returns order detail for the authenticated customer.
func (s *Service) GetAccountOrder(ctx context.Context, userID, orderID uuid.UUID) (*domainorder.Order, error) {
	customer, err := s.customers.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	order, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order.CustomerID != customer.ID {
		return nil, domainorder.ErrNotFound
	}
	return order, nil
}
