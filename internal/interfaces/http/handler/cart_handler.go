package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	appcart "app/internal/application/cart"
	appstorefront "app/internal/application/storefront"
	domaincart "app/internal/domain/cart"
	"app/internal/interfaces/http/dto/request"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/apperror"
	"app/pkg/validator"
)

// CartHandler handles server-side shopping cart HTTP endpoints.
type CartHandler struct {
	carts      *appcart.Service
	storefront *appstorefront.Service
	validator  *validator.Validator
	log        *slog.Logger
}

// NewCartHandler creates a CartHandler.
func NewCartHandler(
	carts *appcart.Service,
	storefront *appstorefront.Service,
	v *validator.Validator,
	log *slog.Logger,
) *CartHandler {
	return &CartHandler{
		carts:      carts,
		storefront: storefront,
		validator:  v,
		log:        log,
	}
}

func (h *CartHandler) resolveOwner(w http.ResponseWriter, r *http.Request) (domaincart.Owner, bool) {
	owner, ok := appmiddleware.GetCartOwner(r.Context())
	if !ok {
		response.Error(w, r, h.log, apperror.Internal("cart session missing"))
		return owner, false
	}
	if err := h.mergeGuestCartIfNeeded(r, owner); err != nil {
		response.Error(w, r, h.log, err)
		return owner, false
	}
	return owner, true
}

func (h *CartHandler) mergeGuestCartIfNeeded(r *http.Request, owner domaincart.Owner) error {
	if owner.UserID == nil {
		return nil
	}
	token := appmiddleware.GetCartGuestToken(r.Context())
	return h.carts.MergeGuestIntoUser(r.Context(), token, *owner.UserID)
}

// GetCart godoc
// @Summary      Get cart
// @Description  Returns the server-side cart with current prices and availability.
// @Tags         store-cart
// @Produce      json
// @Success      200  {object}  appstorefront.CartView
// @Router       /api/v1/store/cart [get]
func (h *CartHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	owner, ok := h.resolveOwner(w, r)
	if !ok {
		return
	}
	view, err := h.storefront.GetCart(r.Context(), owner)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, view)
}

// AddCartItem godoc
// @Summary      Add item to cart
// @Description  Adds a product to the server-side cart (increments quantity when line exists).
// @Tags         store-cart
// @Accept       json
// @Produce      json
// @Param        body  body  request.StoreCartAddItemRequest  true  "Cart item"
// @Success      200   {object}  appstorefront.CartView
// @Router       /api/v1/store/cart/items [post]
func (h *CartHandler) AddCartItem(w http.ResponseWriter, r *http.Request) {
	owner, ok := h.resolveOwner(w, r)
	if !ok {
		return
	}

	var req request.StoreCartAddItemRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	skuID, err := parseOptionalUUID(req.SkuID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if _, err := h.carts.AddItem(r.Context(), owner, appcart.AddItemInput{
		ProductID: productID,
		SkuID:     skuID,
		Quantity:  req.Quantity,
	}); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	view, err := h.storefront.GetCart(r.Context(), owner)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, view)
}

// UpdateCartItem godoc
// @Summary      Update cart item quantity
// @Description  Sets cart line quantity; removes the line when quantity is zero.
// @Tags         store-cart
// @Accept       json
// @Produce      json
// @Param        productId  path  string  true  "Product ID"
// @Param        body       body  request.StoreCartUpdateItemRequest  true  "Quantity"
// @Success      200        {object}  appstorefront.CartView
// @Router       /api/v1/store/cart/items/{productId} [patch]
func (h *CartHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	owner, ok := h.resolveOwner(w, r)
	if !ok {
		return
	}

	productID, err := parseUUIDParam(r, "productId")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.StoreCartUpdateItemRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	skuID, err := parseOptionalUUIDQuery(r, "sku_id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if _, err := h.carts.SetItemQuantity(r.Context(), owner, appcart.UpdateItemInput{
		ProductID: productID,
		SkuID:     skuID,
		Quantity:  req.Quantity,
	}); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	view, err := h.storefront.GetCart(r.Context(), owner)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, view)
}

// RemoveCartItem godoc
// @Summary      Remove cart item
// @Description  Removes a product line from the server-side cart.
// @Tags         store-cart
// @Produce      json
// @Param        productId  path  string  true  "Product ID"
// @Success      200        {object}  appstorefront.CartView
// @Router       /api/v1/store/cart/items/{productId} [delete]
func (h *CartHandler) RemoveCartItem(w http.ResponseWriter, r *http.Request) {
	owner, ok := h.resolveOwner(w, r)
	if !ok {
		return
	}

	productID, err := parseUUIDParam(r, "productId")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	skuID, err := parseOptionalUUIDQuery(r, "sku_id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if _, err := h.carts.RemoveItem(r.Context(), owner, appcart.RemoveItemInput{
		ProductID: productID,
		SkuID:     skuID,
	}); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	view, err := h.storefront.GetCart(r.Context(), owner)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, view)
}

// ClearCart godoc
// @Summary      Clear cart
// @Description  Removes all items from the server-side cart.
// @Tags         store-cart
// @Success      204
// @Router       /api/v1/store/cart [delete]
func (h *CartHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	owner, ok := h.resolveOwner(w, r)
	if !ok {
		return
	}
	if err := h.carts.Clear(r.Context(), owner); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.NoContent(w)
}

func parseOptionalUUID(raw *string) (*uuid.UUID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func parseOptionalUUIDQuery(r *http.Request, key string) (*uuid.UUID, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}
