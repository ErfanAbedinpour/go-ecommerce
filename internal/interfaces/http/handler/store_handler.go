package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appstorefront "app/internal/application/storefront"
	appstorecontent "app/internal/application/storecontent"
	appsettings "app/internal/application/settings"
	apptheme "app/internal/application/theme"
	domainproduct "app/internal/domain/product"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/dto/request"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

// StoreHandler handles public storefront HTTP endpoints.
type StoreHandler struct {
	storefront   *appstorefront.Service
	storecontent *appstorecontent.Service
	settings     *appsettings.Service
	theme        *apptheme.Service
	validator    *validator.Validator
	log          *slog.Logger
}

// NewStoreHandler creates a new StoreHandler.
func NewStoreHandler(
	storefront *appstorefront.Service,
	storecontent *appstorecontent.Service,
	settings *appsettings.Service,
	theme *apptheme.Service,
	v *validator.Validator,
	log *slog.Logger,
) *StoreHandler {
	return &StoreHandler{
		storefront:   storefront,
		storecontent: storecontent,
		settings:     settings,
		theme:        theme,
		validator:    v,
		log:          log,
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
	filter := domainproduct.StoreListFilter{
		Query: r.URL.Query().Get("q"),
		Sort:  r.URL.Query().Get("sort"),
	}
	if catID := r.URL.Query().Get("category_id"); catID != "" {
		id, err := uuid.Parse(catID)
		if err != nil {
			response.Error(w, r, h.log, err)
			return
		}
		filter.CategoryID = &id
	}

	result, err := h.storefront.ListProducts(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToStoreProductListResponse(result))
}

// GetProduct godoc
// @Summary      Get storefront product
// @Description  Returns product detail by slug or UUID including variants and SKUs.
// @Tags         store
// @Produce      json
// @Param        slugOrId  path  string  true  "Product slug or UUID"
// @Success      200  {object}  appstorefront.ProductDetail
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/products/{slugOrId} [get]
func (h *StoreHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	slugOrID := chi.URLParam(r, "slugOrId")
	product, err := h.storefront.GetProduct(r.Context(), slugOrID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, product)
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
