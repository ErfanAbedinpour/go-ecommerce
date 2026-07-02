package storefront

import (
	domainbrand "app/internal/domain/brand"
	domaincategory "app/internal/domain/category"
	domaincoupon "app/internal/domain/coupon"
	domaincustomer "app/internal/domain/customer"
	domainproduct "app/internal/domain/product"
	domainsettings "app/internal/domain/settings"
	"app/internal/domain/user"
	apporder "app/internal/application/order"
	appcart "app/internal/application/cart"
	"app/internal/infrastructure/email"
)

// Service handles public storefront use cases.
type Service struct {
	products   domainproduct.Repository
	categories domaincategory.Repository
	brands     domainbrand.Repository
	orders     *apporder.Service
	coupons    domaincoupon.Repository
	customers  domaincustomer.Repository
	users      user.Repository
	settings   domainsettings.Repository
	carts      *appcart.Service
	mailer     email.Sender
}

// NewService creates a new storefront Service.
func NewService(
	products domainproduct.Repository,
	categories domaincategory.Repository,
	brands domainbrand.Repository,
	orders *apporder.Service,
	coupons domaincoupon.Repository,
	customers domaincustomer.Repository,
	users user.Repository,
	settings domainsettings.Repository,
	carts *appcart.Service,
	mailer email.Sender,
) *Service {
	return &Service{
		products:   products,
		categories: categories,
		brands:     brands,
		orders:     orders,
		coupons:    coupons,
		customers:  customers,
		users:      users,
		settings:   settings,
		carts:      carts,
		mailer:     mailer,
	}
}
