package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	appwishlist "app/internal/application/wishlist"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

type WishlistHandler struct {
	service   *appwishlist.Service
	validator *validator.Validator
	log       *slog.Logger
}

func NewWishlistHandler(service *appwishlist.Service, v *validator.Validator, log *slog.Logger) *WishlistHandler {
	return &WishlistHandler{
		service:   service,
		validator: v,
		log:       log,
	}
}

// Add godoc
// @Summary      Add to wishlist
// @Description  Add a product to the authenticated customer's wishlist.
// @Tags         store-wishlist
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  request.WishlistAddRequest  true  "Product ID to add"
// @Success      210  {object}  map[string]string
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      409  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/account/wishlist [post]
func (h *WishlistHandler) Add(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.WishlistAddRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	prodID, err := uuid.Parse(req.ProductID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	item, err := h.service.Add(r.Context(), userID, prodID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, map[string]string{"id": item.ID.String()})
}

// Remove godoc
// @Summary      Remove from wishlist
// @Description  Remove a product from the authenticated customer's wishlist.
// @Tags         store-wishlist
// @Security     BearerAuth
// @Param        productId  path  string  true  "Product ID to remove"
// @Success      204
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/account/wishlist/{productId} [delete]
func (h *WishlistHandler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	prodID, err := parseUUIDParam(r, "productId")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.service.Remove(r.Context(), userID, prodID); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.NoContent(w)
}

// List godoc
// @Summary      List wishlist
// @Description  Get a paginated list of items in the authenticated customer's wishlist.
// @Tags         store-wishlist
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int  false  "Page number"     default(1)
// @Param        per_page  query  int  false  "Items per page"  default(20)
// @Success      200  {object}  dtoresponse.WishlistListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/account/wishlist [get]
func (h *WishlistHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	page := pagination.FromRequest(r)
	result, err := h.service.List(r.Context(), userID, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToWishlistListResponse(result))
}
