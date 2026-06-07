package handler

import (
	"log/slog"
	"net/http"

	appbrand "app/internal/application/brand"
	domain "app/internal/domain/brand"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

var _ = dtoresponse.BrandResponse{}

// BrandHandler handles brand catalog HTTP endpoints.
type BrandHandler struct {
	service   *appbrand.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewBrandHandler creates a new BrandHandler.
func NewBrandHandler(service *appbrand.Service, v *validator.Validator, log *slog.Logger) *BrandHandler {
	return &BrandHandler{service: service, validator: v, log: log}
}

// Create godoc
// @Summary      Create brand
// @Description  Create a new product brand.
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateBrandRequest  true  "Brand data"
// @Success      201   {object}  dtoresponse.BrandResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/brands [post]
func (h *BrandHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateBrandRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	brand, err := h.service.Create(r.Context(), appbrand.CreateInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		IsActive:    req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToBrandResponse(brand))
}

// List godoc
// @Summary      List brands
// @Description  Get a paginated list of product brands.
// @Tags         brands
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Param        is_active query  bool    false  "Filter by active status"
// @Success      200  {object}  dtoresponse.BrandListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/brands [get]
func (h *BrandHandler) List(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := parseBrandListFilter(r)

	result, err := h.service.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToBrandListResponse(result))
}

// Get godoc
// @Summary      Get brand
// @Description  Get a brand by ID.
// @Tags         brands
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      string  true  "Brand ID"
// @Success      200  {object}  dtoresponse.BrandResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/brands/{id} [get]
func (h *BrandHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	brand, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToBrandResponse(brand))
}

// Update godoc
// @Summary      Update brand
// @Description  Update an existing brand.
// @Tags         brands
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                    true  "Brand ID"
// @Param        body  body      request.UpdateBrandRequest  true  "Brand data"
// @Success      200   {object}  dtoresponse.BrandResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/brands/{id} [put]
func (h *BrandHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateBrandRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	brand, err := h.service.Update(r.Context(), id, appbrand.UpdateInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		IsActive:    req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToBrandResponse(brand))
}

// Delete godoc
// @Summary      Delete brand
// @Description  Soft-delete a brand.
// @Tags         brands
// @Produce      json
// @Security     BearerAuth
// @Param        id   path  string  true  "Brand ID"
// @Success      204  "No Content"
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Failure      422  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/brands/{id} [delete]
func (h *BrandHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func parseBrandListFilter(r *http.Request) domain.ListFilter {
	filter := domain.ListFilter{}
	if active := r.URL.Query().Get("is_active"); active != "" {
		val := active == "true"
		filter.IsActive = &val
	}
	return filter
}
