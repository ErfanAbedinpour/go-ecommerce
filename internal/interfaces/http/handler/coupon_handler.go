package handler

import (
	"log/slog"
	"net/http"

	appsvc "app/internal/application/coupon"
	domain "app/internal/domain/coupon"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

var _ = dtoresponse.CouponResponse{}

// CouponHandler handles coupon HTTP endpoints.
type CouponHandler struct {
	service   *appsvc.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewCouponHandler creates a new CouponHandler.
func NewCouponHandler(service *appsvc.Service, v *validator.Validator, log *slog.Logger) *CouponHandler {
	return &CouponHandler{service: service, validator: v, log: log}
}

// Create godoc
// @Summary      Create coupon
// @Description  Create a new discount coupon.
// @Tags         coupons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateCouponRequest  true  "Coupon data"
// @Success      201   {object}  dtoresponse.CouponResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/coupons [post]
func (h *CouponHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateCouponRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	coupon, err := h.service.Create(r.Context(), appsvc.CreateInput{
		Code:           req.Code,
		DiscountType:   req.DiscountType,
		DiscountValue:  req.DiscountValue,
		MinOrderAmount: req.MinOrderAmount,
		MaxUsage:       req.MaxUsage,
		ExpiresAt:      req.ExpiresAt,
		Note:           req.Note,
		IsActive:       req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.ToCouponResponse(coupon))
}

// Update godoc
// @Summary      Update coupon
// @Description  Update an existing coupon.
// @Tags         coupons
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                       true  "Coupon ID"
// @Param        body  body      request.UpdateCouponRequest  true  "Coupon data"
// @Success      200   {object}  dtoresponse.CouponResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/coupons/{id} [put]
func (h *CouponHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateCouponRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	coupon, err := h.service.Update(r.Context(), id, appsvc.UpdateInput{
		Code:           req.Code,
		DiscountType:   req.DiscountType,
		DiscountValue:  req.DiscountValue,
		MinOrderAmount: req.MinOrderAmount,
		MaxUsage:       req.MaxUsage,
		ExpiresAt:      req.ExpiresAt,
		Note:           req.Note,
		IsActive:       req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToCouponResponse(coupon))
}

// Delete godoc
// @Summary      Delete coupon
// @Description  Soft-delete a coupon.
// @Tags         coupons
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Coupon ID"
// @Success      204 "No Content"
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/coupons/{id} [delete]
func (h *CouponHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.NoContent(w)
}

// Get godoc
// @Summary      Get coupon
// @Description  Get a coupon by ID.
// @Tags         coupons
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Coupon ID"
// @Success      200  {object}  dtoresponse.CouponResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/coupons/{id} [get]
func (h *CouponHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	coupon, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToCouponResponse(coupon))
}

// List godoc
// @Summary      List coupons
// @Description  Get a paginated list of coupons with optional search.
// @Tags         coupons
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Param        is_active query  bool    false  "Filter by active status"
// @Param        q         query  string  false  "Search by code or note"
// @Success      200  {object}  dtoresponse.CouponListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/coupons [get]
func (h *CouponHandler) List(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := parseCouponListFilter(r)

	result, err := h.service.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToCouponListResponse(result))
}

// Activate godoc
// @Summary      Activate coupon
// @Description  Enable a coupon for use.
// @Tags         coupons
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Coupon ID"
// @Success      200  {object}  dtoresponse.CouponActiveResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/coupons/{id}/activate [patch]
func (h *CouponHandler) Activate(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	coupon, err := h.service.Activate(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.CouponActiveResponse{IsActive: coupon.IsActive})
}

// Deactivate godoc
// @Summary      Deactivate coupon
// @Description  Disable a coupon.
// @Tags         coupons
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Coupon ID"
// @Success      200  {object}  dtoresponse.CouponActiveResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/coupons/{id}/deactivate [patch]
func (h *CouponHandler) Deactivate(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	coupon, err := h.service.Deactivate(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.CouponActiveResponse{IsActive: coupon.IsActive})
}

func parseCouponListFilter(r *http.Request) domain.ListFilter {
	q := r.URL.Query()
	filter := domain.ListFilter{Query: q.Get("q")}
	if active := q.Get("is_active"); active != "" {
		val := active == "true"
		filter.IsActive = &val
	}
	return filter
}
