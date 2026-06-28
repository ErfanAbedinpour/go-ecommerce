package response

import (
	appstorefront "app/internal/application/storefront"
	"app/pkg/pagination"
)
type StoreProductListResponse struct {
	Data []appstorefront.ProductCard `json:"data"`
	Meta pagination.Meta              `json:"meta"`
}

// StoreAccountOrderListResponse is a paginated customer order list.
type StoreAccountOrderListResponse struct {
	Data []appstorefront.AccountOrderSummary `json:"data"`
	Meta pagination.Meta                    `json:"meta"`
}

// ToStoreProductListResponse maps paginated product cards to API response.
func ToStoreProductListResponse(result pagination.Paginated[appstorefront.ProductCard]) StoreProductListResponse {
	return StoreProductListResponse{
		Data: result.Data,
		Meta: result.Meta,
	}
}

// ToStoreAccountOrderListResponse maps paginated account orders to API response.
func ToStoreAccountOrderListResponse(result pagination.Paginated[appstorefront.AccountOrderSummary]) StoreAccountOrderListResponse {
	return StoreAccountOrderListResponse{
		Data: result.Data,
		Meta: result.Meta,
	}
}
