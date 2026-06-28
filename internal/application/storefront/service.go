package storefront

import (
	domaincategory "app/internal/domain/category"
	domaincoupon "app/internal/domain/coupon"
	domaincustomer "app/internal/domain/customer"
	domainproduct "app/internal/domain/product"
	domainsettings "app/internal/domain/settings"
	apporder "app/internal/application/order"
)

// Service handles public storefront use cases.
type Service struct {
	products   domainproduct.Repository
	categories domaincategory.Repository
	orders     *apporder.Service
	coupons    domaincoupon.Repository
	customers  domaincustomer.Repository
	settings   domainsettings.Repository
}

// NewService creates a new storefront Service.
func NewService(
	products domainproduct.Repository,
	categories domaincategory.Repository,
	orders *apporder.Service,
	coupons domaincoupon.Repository,
	customers domaincustomer.Repository,
	settings domainsettings.Repository,
) *Service {
	return &Service{
		products:   products,
		categories: categories,
		orders:     orders,
		coupons:    coupons,
		customers:  customers,
		settings:   settings,
	}
}
