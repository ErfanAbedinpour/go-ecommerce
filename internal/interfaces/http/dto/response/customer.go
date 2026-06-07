package response

import (
	"time"

	domain "app/internal/domain/customer"
	domainorder "app/internal/domain/order"
	"app/pkg/pagination"
)

// CustomerResponse is the customer representation in list API responses.
type CustomerResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	FullName    string    `json:"full_name"`
	Phone       string    `json:"phone,omitempty"`
	Type        string    `json:"type"`
	TotalOrders int       `json:"total_orders"`
	TotalSpent  float64   `json:"total_spent"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CustomerAddressResponse is a customer address in API responses.
type CustomerAddressResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	IsDefault  bool   `json:"is_default"`
}

// CustomerStatsResponse holds aggregate purchase statistics.
type CustomerStatsResponse struct {
	TotalOrders int     `json:"total_orders"`
	TotalSpent  float64 `json:"total_spent"`
}

// CustomerDetailResponse is the detailed customer view with addresses and stats.
type CustomerDetailResponse struct {
	CustomerResponse
	Addresses []CustomerAddressResponse `json:"addresses"`
	Stats     CustomerStatsResponse     `json:"stats"`
}

// CustomerListResponse is a paginated list of customers.
type CustomerListResponse struct {
	Data []CustomerResponse `json:"data"`
	Meta pagination.Meta    `json:"meta"`
}

// OrderSummaryResponse is a compact order view for purchase history.
type OrderSummaryResponse struct {
	ID            string    `json:"id"`
	OrderNumber   string    `json:"order_number"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	Total         float64   `json:"total"`
	ItemCount     int       `json:"item_count"`
	CreatedAt     time.Time `json:"created_at"`
}

// CustomerOrderListResponse is a paginated list of customer orders.
type CustomerOrderListResponse struct {
	Data []OrderSummaryResponse `json:"data"`
	Meta pagination.Meta        `json:"meta"`
}

// ToCustomerResponse maps a domain customer to API response.
func ToCustomerResponse(c *domain.Customer) CustomerResponse {
	return CustomerResponse{
		ID:          c.ID.String(),
		Email:       c.Email,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		FullName:    c.FullName(),
		Phone:       c.Phone,
		Type:        c.Type.String(),
		TotalOrders: c.TotalOrders,
		TotalSpent:  c.TotalSpent,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

// ToCustomerAddressResponse maps a domain address to API response.
func ToCustomerAddressResponse(a domain.Address) CustomerAddressResponse {
	return CustomerAddressResponse{
		ID:         a.ID.String(),
		Type:       a.Type.String(),
		Street:     a.Street,
		City:       a.City,
		State:      a.State,
		PostalCode: a.PostalCode,
		Country:    a.Country,
		IsDefault:  a.IsDefault,
	}
}

// ToCustomerDetailResponse maps a domain customer with addresses to API response.
func ToCustomerDetailResponse(c *domain.Customer) CustomerDetailResponse {
	addresses := make([]CustomerAddressResponse, len(c.Addresses))
	for i, a := range c.Addresses {
		addresses[i] = ToCustomerAddressResponse(a)
	}
	return CustomerDetailResponse{
		CustomerResponse: ToCustomerResponse(c),
		Addresses:        addresses,
		Stats: CustomerStatsResponse{
			TotalOrders: c.TotalOrders,
			TotalSpent:  c.TotalSpent,
		},
	}
}

// ToCustomerListResponse maps a paginated domain list to API response.
func ToCustomerListResponse(result pagination.Paginated[domain.Customer]) CustomerListResponse {
	items := make([]CustomerResponse, len(result.Data))
	for i, c := range result.Data {
		items[i] = ToCustomerResponse(&c)
	}
	return CustomerListResponse{Data: items, Meta: result.Meta}
}

// ToOrderSummaryResponse maps an order summary to API response.
func ToOrderSummaryResponse(o domainorder.Summary) OrderSummaryResponse {
	return OrderSummaryResponse{
		ID:            o.ID.String(),
		OrderNumber:   o.OrderNumber,
		Status:        o.Status,
		PaymentStatus: o.PaymentStatus,
		Total:         o.Total,
		ItemCount:     o.ItemCount,
		CreatedAt:     o.CreatedAt,
	}
}

// ToCustomerOrderListResponse maps a paginated order list to API response.
func ToCustomerOrderListResponse(result pagination.Paginated[domainorder.Summary]) CustomerOrderListResponse {
	items := make([]OrderSummaryResponse, len(result.Data))
	for i, o := range result.Data {
		items[i] = ToOrderSummaryResponse(o)
	}
	return CustomerOrderListResponse{Data: items, Meta: result.Meta}
}
