package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appstorefront "app/internal/application/storefront"
	appstorecontent "app/internal/application/storecontent"
	appsettings "app/internal/application/settings"
	appreview "app/internal/application/productreview"
	apptheme "app/internal/application/theme"
	appwishlist "app/internal/application/wishlist"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/dto/request"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/apperror"
	"app/pkg/pagination"
	"app/pkg/validator"
)

// StoreHandler handles public storefront HTTP endpoints.
type StoreHandler struct {
	storefront            *appstorefront.Service
	storecontent          *appstorecontent.Service
	settings              *appsettings.Service
	theme                 *apptheme.Service
	reviews               *appreview.Service
	wishlist              *appwishlist.Service
	paymentCallbackSecret string
	validator             *validator.Validator
	log                   *slog.Logger
}

// NewStoreHandler creates a new StoreHandler.
func NewStoreHandler(
	storefront *appstorefront.Service,
	storecontent *appstorecontent.Service,
	settings *appsettings.Service,
	theme *apptheme.Service,
	reviews *appreview.Service,
	wishlist *appwishlist.Service,
	paymentCallbackSecret string,
	v *validator.Validator,
	log *slog.Logger,
) *StoreHandler {
	return &StoreHandler{
		storefront:            storefront,
		storecontent:          storecontent,
		settings:              settings,
		theme:                 theme,
		reviews:               reviews,
		wishlist:              wishlist,
		paymentCallbackSecret: paymentCallbackSecret,
		validator:             v,
		log:                   log,
	}
}

// ListProducts godoc
// @Summary      List storefront products
// @Description  Returns active products for the public storefront with search, filters, and sort.
// @Tags         store
// @Produce      json
// @Param        page         query  int     false  "Page number"     default(1)
// @Param        per_page     query  int     false  "Items per page"  default(20)
// @Param        q            query  string  false  "Search query"
// @Param        category_id  query  string  false  "Category ID"
// @Param        sort         query  string  false  "Sort mode"  Enums(bestseller, newest, discounted, price_asc, price_desc)
// @Success      200  {object}  dtoresponse.StoreProductListResponse
// @Router       /api/v1/store/products [get]
func (h *StoreHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter, err := h.storefront.BuildStoreListFilter(r.Context(), r.URL.Query())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	result, err := h.storefront.ListProducts(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToStoreProductListResponse(result))
}

// SearchProducts godoc
// @Summary      Search storefront products
// @Description  Returns quick product suggestions for header autocomplete.
// @Tags         store
// @Produce      json
// @Param        q      query  string  true   "Search query"
// @Param        limit  query  int     false  "Max results"  default(10)
// @Success      200    {object}  appstorefront.ProductSearchResult
// @Failure      400    {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/products/search [get]
func (h *StoreHandler) SearchProducts(w http.ResponseWriter, r *http.Request) {
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		response.Error(w, r, h.log, apperror.Validation("search query is required", map[string]string{"q": "is required"}))
		return
	}

	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	result, err := h.storefront.SearchProducts(r.Context(), query, limit)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, result)
}

// ListRelatedProducts godoc
// @Summary      List related products
// @Description  Returns related active products for a product detail page.
// @Tags         store
// @Produce      json
// @Param        id     path   string  true   "Product ID or slug"
// @Param        limit  query  int     false  "Max results"  default(8)
// @Success      200    {object}  appstorefront.ProductListData
// @Failure      404    {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/products/{id}/related [get]
func (h *StoreHandler) ListRelatedProducts(w http.ResponseWriter, r *http.Request) {
	productRef := chi.URLParam(r, "id")
	limit := 8
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	result, err := h.storefront.ListRelatedProducts(r.Context(), productRef, limit)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, result)
}

// ListBrands godoc
// @Summary      List storefront brands
// @Description  Returns active brands for catalog filters.
// @Tags         store
// @Produce      json
// @Success      200  {object}  appstorefront.StoreBrandList
// @Router       /api/v1/store/brands [get]
func (h *StoreHandler) ListBrands(w http.ResponseWriter, r *http.Request) {
	result, err := h.storefront.ListBrands(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, result)
}

// GetShippingMethods godoc
// @Summary      List shipping methods
// @Description  Returns available delivery options for a destination city.
// @Tags         store
// @Produce      json
// @Param        city  query  string  true  "Destination city"
// @Success      200   {object}  appstorefront.ShippingMethodList
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/checkout/shipping-methods [get]
func (h *StoreHandler) GetShippingMethods(w http.ResponseWriter, r *http.Request) {
	city := strings.TrimSpace(r.URL.Query().Get("city"))
	result, err := h.storefront.GetShippingMethods(r.Context(), city)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, result)
}

// GetProduct godoc
// @Summary      Get storefront product
// @Description  Returns product detail by slug or UUID including variants and SKUs.
// @Tags         store
// @Produce      json
// @Param        slugOrId  path   string  true   "Product slug or UUID"
// @Param        include   query  string  false  "Comma-separated embeds: reviews_summary,wishlist"
// @Success      200  {object}  appstorefront.ProductDetailEnriched
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/products/{slugOrId} [get]
func (h *StoreHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	slugOrID := chi.URLParam(r, "slugOrId")
	includes := parseIncludeSet(r.URL.Query().Get("include"))

	opts := appstorefront.ProductDetailOptions{
		IncludeReviewsSummary: includes["reviews_summary"],
		IncludeWishlist:       includes["wishlist"],
	}
	if userID, ok := appmiddleware.GetUserIDOptional(r.Context()); ok {
		opts.UserID = &userID
	}

	product, err := h.storefront.GetProductEnriched(
		r.Context(),
		slugOrID,
		opts,
		h.reviews,
		h.wishlist,
	)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, product)
}

func parseIncludeSet(raw string) map[string]bool {
	out := make(map[string]bool)
	for part := range strings.SplitSeq(raw, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out[part] = true
		}
	}
	return out
}

// ListCategories godoc
// @Summary      List storefront categories
// @Description  Returns the active category tree for the storefront.
// @Tags         store
// @Produce      json
// @Success      200  {array}  dtoresponse.CategoryResponse
// @Router       /api/v1/store/categories [get]
func (h *StoreHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.storefront.ListCategories(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToCategoryTreeResponse(items))
}

// ValidateCoupon godoc
// @Summary      Validate coupon
// @Description  Validates a coupon code against a subtotal in Toman.
// @Tags         store
// @Accept       json
// @Produce      json
// @Param        body  body  request.StoreCouponValidateRequest  true  "Coupon validation"
// @Success      200   {object}  appstorefront.CouponResult
// @Router       /api/v1/store/coupons/validate [post]
func (h *StoreHandler) ValidateCoupon(w http.ResponseWriter, r *http.Request) {
	var req request.StoreCouponValidateRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	result, err := h.storefront.ValidateCoupon(r.Context(), appstorefront.CouponValidateInput{
		Code:          req.Code,
		SubtotalToman: req.SubtotalToman,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, result)
}

// PreviewCheckout godoc
// @Summary      Preview checkout
// @Description  Validates cart items and computes checkout totals without placing an order.
// @Tags         store
// @Accept       json
// @Produce      json
// @Param        body  body  request.StoreCheckoutPreviewRequest  true  "Checkout preview"
// @Success      200   {object}  appstorefront.PreviewCheckoutOutput
// @Router       /api/v1/store/checkout/preview [post]
func (h *StoreHandler) PreviewCheckout(w http.ResponseWriter, r *http.Request) {
	var req request.StoreCheckoutPreviewRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	result, err := h.storefront.PreviewCheckout(r.Context(), appstorefront.PreviewCheckoutInput{
		Items:          toCheckoutItems(req.Items),
		CouponCode:     req.CouponCode,
		ShippingAmount: req.ShippingAmount,
		TaxAmount:      req.TaxAmount,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, result)
}

// Checkout godoc
// @Summary      Place order
// @Description  Creates an unpaid storefront order for guest or authenticated customers.
// @Tags         store
// @Accept       json
// @Produce      json
// @Param        body  body  request.StoreCheckoutRequest  true  "Checkout"
// @Success      201   {object}  appstorefront.PlaceCheckoutOutput
// @Router       /api/v1/store/checkout [post]
func (h *StoreHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req request.StoreCheckoutRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	billing := toOrderAddress(req.ShippingAddress)
	if req.BillingAddress != nil {
		billing = toOrderAddress(*req.BillingAddress)
	}

	input := appstorefront.PlaceCheckoutInput{
		Items:           toCheckoutItems(req.Items),
		CouponCode:      req.CouponCode,
		Customer: appstorefront.CheckoutCustomerInput{
			Email:     req.Customer.Email,
			FirstName: req.Customer.FirstName,
			LastName:  req.Customer.LastName,
			Phone:     req.Customer.Phone,
		},
		ShippingAddress: toOrderAddress(req.ShippingAddress),
		BillingAddress:  billing,
		ShippingAmount:  req.ShippingAmount,
		TaxAmount:       req.TaxAmount,
		PaymentMethod:   req.PaymentMethod,
		Notes:           req.Notes,
	}

	if userID, ok := appmiddleware.GetUserIDOptional(r.Context()); ok {
		input.UserID = &userID
	}

	result, err := h.storefront.PlaceCheckout(r.Context(), input)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, result)
}

// GetCheckoutSettings godoc
// @Summary      Get checkout settings
// @Description  Returns minimum order amount, enabled payment methods, and COD availability.
// @Tags         store
// @Produce      json
// @Success      200  {object}  appstorefront.CheckoutSettingsOutput
// @Router       /api/v1/store/settings/checkout [get]
func (h *StoreHandler) GetCheckoutSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.storefront.GetCheckoutSettings(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, settings)
}

// PaymentCallback godoc
// @Summary      Payment gateway callback
// @Description  Handles PSP redirect callback after online payment and updates order payment status.
// @Tags         store
// @Accept       json
// @Produce      json
// @Param        body  body  request.StorePaymentCallbackRequest  true  "Payment callback"
// @Success      200   {object}  appstorefront.PaymentCallbackOutput
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/checkout/payment/callback [post]
func (h *StoreHandler) PaymentCallback(w http.ResponseWriter, r *http.Request) {
	var req request.StorePaymentCallbackRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	signature := req.Signature
	if signature == "" {
		signature = r.Header.Get("X-Payment-Signature")
	}

	result, err := h.storefront.HandlePaymentCallback(r.Context(), appstorefront.PaymentCallbackInput{
		OrderID:   orderID,
		Authority: req.Authority,
		Status:    req.Status,
		Signature: signature,
	}, h.paymentCallbackSecret)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, result)
}

// GetHomepage godoc
// @Summary      Get homepage content
// @Description  Returns aggregated homepage sections for the storefront.
// @Tags         store
// @Produce      json
// @Success      200  {object}  appstorecontent.HomepageProjection
// @Router       /api/v1/store/homepage [get]
func (h *StoreHandler) GetHomepage(w http.ResponseWriter, r *http.Request) {
	data, err := h.storecontent.BuildHomepage(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, data)
}

// GetSettings godoc
// @Summary      Get public store settings
// @Description  Returns public site, contact, social, and SEO settings.
// @Tags         store
// @Produce      json
// @Success      200  {object}  appsettings.PublicSettings
// @Router       /api/v1/store/settings [get]
func (h *StoreHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	data, err := h.settings.GetPublicSettings(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, data)
}

// GetTheme godoc
// @Summary      Get active storefront theme
// @Description  Returns active theme slug, color tokens, and font for CSS variables.
// @Tags         store
// @Produce      json
// @Success      200  {object}  apptheme.PublicThemeOutput
// @Router       /api/v1/store/theme [get]
func (h *StoreHandler) GetTheme(w http.ResponseWriter, r *http.Request) {
	data, err := h.theme.GetPublicTheme(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, data)
}

// ListAccountOrders godoc
// @Summary      List customer orders
// @Description  Returns order history for the authenticated customer.
// @Tags         store
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int  false  "Page number"     default(1)
// @Param        per_page  query  int  false  "Items per page"  default(20)
// @Success      200  {object}  dtoresponse.StoreAccountOrderListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/account/orders [get]
func (h *StoreHandler) ListAccountOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	page := pagination.FromRequest(r)
	result, err := h.storefront.ListAccountOrders(r.Context(), userID, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToStoreAccountOrderListResponse(result))
}

// GetAccountOrder godoc
// @Summary      Get customer order
// @Description  Returns order detail for the authenticated customer.
// @Tags         store
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Order ID"
// @Success      200  {object}  dtoresponse.OrderDetailResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/account/orders/{id} [get]
func (h *StoreHandler) GetAccountOrder(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	orderID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	order, err := h.storefront.GetAccountOrder(r.Context(), userID, orderID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToOrderDetailResponse(order))
}

// GetAccountProfile godoc
// @Summary      Get customer profile
// @Description  Returns profile, addresses, and purchase stats for the authenticated customer.
// @Tags         store
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  appstorefront.AccountProfile
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/account/profile [get]
func (h *StoreHandler) GetAccountProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	profile, err := h.storefront.GetAccountProfile(r.Context(), userID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, profile)
}

// UpdateAccountProfile godoc
// @Summary      Update customer profile
// @Description  Updates profile fields and replaces saved addresses for the authenticated customer.
// @Tags         store
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  request.UpdateStoreAccountProfileRequest  true  "Profile update"
// @Success      200   {object}  appstorefront.AccountProfile
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/account/profile [put]
func (h *StoreHandler) UpdateAccountProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateStoreAccountProfileRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	profile, err := h.storefront.UpdateAccountProfile(r.Context(), userID, toUpdateAccountProfileInput(req))
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, profile)
}

// GetAbout godoc
// @Summary      Get about page content
// @Description  Returns aggregated about page content for the public storefront.
// @Tags         store
// @Produce      json
// @Success      200  {object}  appstorefront.AboutPage
// @Router       /api/v1/store/about [get]
func (h *StoreHandler) GetAbout(w http.ResponseWriter, r *http.Request) {
	page, err := h.storefront.GetAboutPage(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, page)
}

// GetNavigation godoc
// @Summary      Get storefront navigation
// @Description  Returns the active storefront navigation menu.
// @Tags         store
// @Produce      json
// @Success      200  {object}  appstorefront.StoreNavigation
// @Router       /api/v1/store/navigation [get]
func (h *StoreHandler) GetNavigation(w http.ResponseWriter, r *http.Request) {
	navigation, err := h.storefront.GetStoreNavigation(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, navigation)
}

func toUpdateAccountProfileInput(req request.UpdateStoreAccountProfileRequest) appstorefront.UpdateAccountProfileInput {
	addresses := make([]appstorefront.UpdateAccountAddressInput, len(req.Addresses))
	for i, address := range req.Addresses {
		var id *uuid.UUID
		if address.ID != nil && *address.ID != "" {
			if parsed, err := uuid.Parse(*address.ID); err == nil {
				id = &parsed
			}
		}
		addresses[i] = appstorefront.UpdateAccountAddressInput{
			ID:         id,
			Type:       address.Type,
			Street:     address.Street,
			City:       address.City,
			State:      address.State,
			PostalCode: address.PostalCode,
			Country:    address.Country,
			IsDefault:  address.IsDefault,
		}
	}

	return appstorefront.UpdateAccountProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Addresses: addresses,
	}
}

func toCheckoutItems(items []request.StoreCheckoutItemRequest) []appstorefront.CheckoutItemInput {
	result := make([]appstorefront.CheckoutItemInput, len(items))
	for i, item := range items {
		productID, _ := uuid.Parse(item.ProductID)
		var skuID *uuid.UUID
		if item.SkuID != nil && *item.SkuID != "" {
			if id, err := uuid.Parse(*item.SkuID); err == nil {
				skuID = &id
			}
		}
		result[i] = appstorefront.CheckoutItemInput{
			ProductID: productID,
			SkuID:     skuID,
			Quantity:  item.Quantity,
		}
	}
	return result
}
