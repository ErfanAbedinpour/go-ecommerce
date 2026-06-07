package response

import (
	"time"

	domain "app/internal/domain/order"
	"app/pkg/pagination"
)

// OrderAddressResponse is an order billing or shipping address.
type OrderAddressResponse struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state,omitempty"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// OrderCustomerResponse is customer info embedded in order detail.
type OrderCustomerResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	FullName  string `json:"full_name"`
	Phone     string `json:"phone,omitempty"`
}

// OrderItemResponse is a line item in order detail.
type OrderItemResponse struct {
	ID          string  `json:"id"`
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	ProductSKU  string  `json:"product_sku"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

// OrderTimelineResponse is a status change history entry.
type OrderTimelineResponse struct {
	ID         string    `json:"id"`
	FromStatus *string   `json:"from_status,omitempty"`
	ToStatus   string    `json:"to_status"`
	Note       string    `json:"note,omitempty"`
	ChangedBy  *string   `json:"changed_by,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// OrderListItemResponse is an order in the admin list view.
type OrderListItemResponse struct {
	ID            string    `json:"id"`
	OrderNumber   string    `json:"order_number"`
	Status        string    `json:"status"`
	PaymentStatus string    `json:"payment_status"`
	Total         float64   `json:"total"`
	ItemCount     int       `json:"item_count"`
	CustomerID    string    `json:"customer_id"`
	CustomerName  string    `json:"customer_name"`
	CustomerEmail string    `json:"customer_email"`
	CreatedAt     time.Time `json:"created_at"`
}

// OrderListResponse is a paginated list of orders.
type OrderListResponse struct {
	Data []OrderListItemResponse `json:"data"`
	Meta pagination.Meta         `json:"meta"`
}

// OrderDetailResponse is the full order detail view.
type OrderDetailResponse struct {
	ID              string                  `json:"id"`
	OrderNumber     string                  `json:"order_number"`
	CustomerID      string                  `json:"customer_id"`
	CouponID        *string                 `json:"coupon_id,omitempty"`
	Status          string                  `json:"status"`
	PaymentStatus   string                  `json:"payment_status"`
	Subtotal        float64                 `json:"subtotal"`
	DiscountAmount  float64                 `json:"discount_amount"`
	ShippingAmount  float64                 `json:"shipping_amount"`
	TaxAmount       float64                 `json:"tax_amount"`
	Total           float64                 `json:"total"`
	Notes           string                  `json:"notes,omitempty"`
	BillingAddress  OrderAddressResponse    `json:"billing_address"`
	ShippingAddress OrderAddressResponse    `json:"shipping_address"`
	Customer        *OrderCustomerResponse  `json:"customer,omitempty"`
	Items           []OrderItemResponse     `json:"items"`
	Timeline        []OrderTimelineResponse `json:"timeline"`
	CreatedAt       time.Time               `json:"created_at"`
	UpdatedAt       time.Time               `json:"updated_at"`
}

// ToOrderListItemResponse maps a list item to API response.
func ToOrderListItemResponse(item domain.ListItem) OrderListItemResponse {
	return OrderListItemResponse{
		ID:            item.ID.String(),
		OrderNumber:   item.OrderNumber,
		Status:        item.Status,
		PaymentStatus: item.PaymentStatus,
		Total:         item.Total,
		ItemCount:     item.ItemCount,
		CustomerID:    item.CustomerID.String(),
		CustomerName:  item.CustomerName,
		CustomerEmail: item.CustomerEmail,
		CreatedAt:     item.CreatedAt,
	}
}

// ToOrderListResponse maps a paginated list to API response.
func ToOrderListResponse(result pagination.Paginated[domain.ListItem]) OrderListResponse {
	items := make([]OrderListItemResponse, len(result.Data))
	for i, item := range result.Data {
		items[i] = ToOrderListItemResponse(item)
	}
	return OrderListResponse{Data: items, Meta: result.Meta}
}

// ToOrderDetailResponse maps a domain order to API response.
func ToOrderDetailResponse(o *domain.Order) OrderDetailResponse {
	resp := OrderDetailResponse{
		ID:              o.ID.String(),
		OrderNumber:     o.OrderNumber,
		CustomerID:      o.CustomerID.String(),
		Status:          o.Status.String(),
		PaymentStatus:   o.PaymentStatus.String(),
		Subtotal:        o.Subtotal,
		DiscountAmount:  o.DiscountAmount,
		ShippingAmount:  o.ShippingAmount,
		TaxAmount:       o.TaxAmount,
		Total:           o.Total,
		Notes:           o.Notes,
		BillingAddress:  toOrderAddressResponse(o.BillingAddress),
		ShippingAddress: toOrderAddressResponse(o.ShippingAddress),
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
	}
	if o.CouponID != nil {
		cid := o.CouponID.String()
		resp.CouponID = &cid
	}
	if o.Customer != nil {
		resp.Customer = &OrderCustomerResponse{
			ID:        o.Customer.ID.String(),
			Email:     o.Customer.Email,
			FirstName: o.Customer.FirstName,
			LastName:  o.Customer.LastName,
			FullName:  o.Customer.FullName(),
			Phone:     o.Customer.Phone,
		}
	}
	resp.Items = make([]OrderItemResponse, len(o.Items))
	for i, item := range o.Items {
		resp.Items[i] = OrderItemResponse{
			ID:          item.ID.String(),
			ProductID:   item.ProductID.String(),
			ProductName: item.ProductName,
			ProductSKU:  item.ProductSKU,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		}
	}
	resp.Timeline = make([]OrderTimelineResponse, len(o.Timeline))
	for i, entry := range o.Timeline {
		tl := OrderTimelineResponse{
			ID:        entry.ID.String(),
			ToStatus:  entry.ToStatus.String(),
			Note:      entry.Note,
			CreatedAt: entry.CreatedAt,
		}
		if entry.FromStatus != nil {
			s := entry.FromStatus.String()
			tl.FromStatus = &s
		}
		if entry.ChangedBy != nil {
			id := entry.ChangedBy.String()
			tl.ChangedBy = &id
		}
		resp.Timeline[i] = tl
	}
	return resp
}

func toOrderAddressResponse(a domain.Address) OrderAddressResponse {
	return OrderAddressResponse{
		Street:     a.Street,
		City:       a.City,
		State:      a.State,
		PostalCode: a.PostalCode,
		Country:    a.Country,
	}
}
