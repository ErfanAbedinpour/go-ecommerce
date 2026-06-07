package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	appsvc "app/internal/application/category"
	domain "app/internal/domain/category"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

var _ = dtoresponse.CategoryResponse{}

// CategoryHandler handles category HTTP endpoints.
type CategoryHandler struct {
	service   *appsvc.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewCategoryHandler creates a new CategoryHandler.
func NewCategoryHandler(service *appsvc.Service, v *validator.Validator, log *slog.Logger) *CategoryHandler {
	return &CategoryHandler{service: service, validator: v, log: log}
}

// Create godoc
// @Summary      Create category
// @Description  Create a new product category with optional parent for hierarchy.
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateCategoryRequest  true  "Category data"
// @Success      201   {object}  dtoresponse.CategoryResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/categories [post]
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateCategoryRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	input, err := toCategoryCreateInput(req)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	category, err := h.service.Create(r.Context(), input)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.ToCategoryResponse(category))
}

// Update godoc
// @Summary      Update category
// @Description  Update an existing category. Cannot set parent to self or descendant.
// @Tags         categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                         true  "Category ID"
// @Param        body  body      request.UpdateCategoryRequest  true  "Category data"
// @Success      200   {object}  dtoresponse.CategoryResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Failure      422   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/categories/{id} [put]
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateCategoryRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	input, err := toCategoryUpdateInput(req)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	category, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToCategoryResponse(category))
}

// Delete godoc
// @Summary      Delete category
// @Description  Soft-delete a category. Fails if it has children or assigned products.
// @Tags         categories
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Category ID"
// @Success      204 "No Content"
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Failure      422  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/categories/{id} [delete]
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
// @Summary      Get category
// @Description  Get a category by ID.
// @Tags         categories
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Category ID"
// @Success      200  {object}  dtoresponse.CategoryResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/categories/{id} [get]
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	category, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToCategoryResponse(category))
}

// List godoc
// @Summary      List categories
// @Description  Get paginated categories or a nested tree with tree=true.
// @Tags         categories
// @Produce      json
// @Security     BearerAuth
// @Param        page        query  int     false  "Page number"     default(1)
// @Param        per_page    query  int     false  "Items per page"  default(20)
// @Param        sort        query  string  false  "Sort field"      Enums(created_at, name, sort_order, updated_at)
// @Param        order       query  string  false  "Sort order"      Enums(asc, desc)
// @Param        parent_id   query  string  false  "Filter by parent category ID"
// @Param        is_active   query  bool    false  "Filter by active status"
// @Param        tree        query  bool    false  "Return nested tree structure"
// @Success      200  {object}  dtoresponse.CategoryListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/categories [get]
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := parseCategoryListFilter(r)

	result, err := h.service.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	switch data := result.(type) {
	case []domain.Category:
		response.OK(w, dtoresponse.ToCategoryTreeResponse(data))
	default:
		response.OK(w, dtoresponse.ToCategoryListResponse(data.(pagination.Paginated[domain.Category])))
	}
}

func parseCategoryListFilter(r *http.Request) domain.ListFilter {
	q := r.URL.Query()
	filter := domain.ListFilter{
		Tree: q.Get("tree") == "true",
	}
	if pid := q.Get("parent_id"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			filter.ParentID = &id
		}
	}
	if active := q.Get("is_active"); active != "" {
		val := active == "true"
		filter.IsActive = &val
	}
	return filter
}

func toCategoryCreateInput(req request.CreateCategoryRequest) (appsvc.CreateInput, error) {
	input := appsvc.CreateInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	}
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return input, err
		}
		input.ParentID = &id
	}
	return input, nil
}

func toCategoryUpdateInput(req request.UpdateCategoryRequest) (appsvc.UpdateInput, error) {
	input := appsvc.UpdateInput{
		Name: req.Name, Slug: req.Slug, Description: req.Description,
		ImageURL: req.ImageURL, SortOrder: req.SortOrder, IsActive: req.IsActive,
	}
	if req.ParentID != nil {
		id, err := uuid.Parse(*req.ParentID)
		if err != nil {
			return input, err
		}
		input.ParentID = &id
	}
	return input, nil
}
