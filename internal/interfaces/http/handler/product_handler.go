package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appsvc "app/internal/application/product"
	domain "app/internal/domain/product"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/dto/request"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

var _ = dtoresponse.ProductResponse{}

// ProductHandler handles product HTTP endpoints.
type ProductHandler struct {
	service   *appsvc.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewProductHandler creates a new ProductHandler.
func NewProductHandler(service *appsvc.Service, v *validator.Validator, log *slog.Logger) *ProductHandler {
	return &ProductHandler{service: service, validator: v, log: log}
}

// Create godoc
// @Summary      Create product
// @Description  Create a new product with images, attributes, and inventory.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.CreateProductRequest  true  "Product data"
// @Success      201   {object}  dtoresponse.ProductResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/products [post]
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateProductRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	input, err := toCreateInput(req)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	product, err := h.service.Create(r.Context(), input)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.ToProductResponse(product))
}

// Update godoc
// @Summary      Update product
// @Description  Update an existing product. Supports partial updates.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                        true  "Product ID"
// @Param        body  body      request.UpdateProductRequest  true  "Product data"
// @Success      200   {object}  dtoresponse.ProductResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/products/{id} [put]
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateProductRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	input, err := toUpdateInput(req)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	product, err := h.service.Update(r.Context(), id, input)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToProductResponse(product))
}

// Delete godoc
// @Summary      Delete product
// @Description  Soft-delete a product. Fails if referenced by active orders.
// @Tags         products
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Product ID"
// @Success      204 "No Content"
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Failure      422  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/products/{id} [delete]
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
// @Summary      Get product
// @Description  Get a product by ID with images, attributes, and inventory.
// @Tags         products
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Product ID"
// @Success      200  {object}  dtoresponse.ProductResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/products/{id} [get]
func (h *ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	product, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToProductResponse(product))
}

// Stats godoc
// @Summary      Product statistics
// @Description  Get product catalog KPI counts for dashboard cards.
// @Tags         products
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.ProductStatsResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/products/stats [get]
func (h *ProductHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToProductStatsResponse(stats))
}

// List godoc
// @Summary      List products
// @Description  Get a paginated list of products with optional filters.
// @Tags         products
// @Produce      json
// @Security     BearerAuth
// @Param        page          query  int     false  "Page number"        default(1)
// @Param        per_page      query  int     false  "Items per page"     default(20)
// @Param        sort          query  string  false  "Sort field"         Enums(created_at, name, price, updated_at)
// @Param        order         query  string  false  "Sort order"         Enums(asc, desc)
// @Param        status        query  string  false  "Filter by status"   Enums(draft, active, archived)
// @Param        category_id   query  string  false  "Filter by category"
// @Param        brand         query  string  false  "Filter by brand"
// @Param        is_featured   query  bool    false  "Filter featured products"
// @Param        stock_level   query  string  false  "Filter by stock level"  Enums(low, out)
// @Success      200  {object}  dtoresponse.ProductListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/products [get]
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := parseListFilter(r)

	result, err := h.service.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToProductListResponse(result))
}

// Search godoc
// @Summary      Search products
// @Description  Search products by name, SKU, or description.
// @Tags         products
// @Produce      json
// @Security     BearerAuth
// @Param        q         query  string  true   "Search query (min 2 chars)"
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Success      200  {object}  dtoresponse.ProductListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/products/search [get]
func (h *ProductHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	page := pagination.FromRequest(r)

	result, err := h.service.Search(r.Context(), query, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToProductListResponse(result))
}

// UpdateInventory godoc
// @Summary      Update product inventory
// @Description  Adjust stock quantity and low-stock threshold for a product.
// @Tags         products
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                        true  "Product ID"
// @Param        body  body  request.UpdateInventoryRequest  true  "Inventory data"
// @Success      200   {object}  dtoresponse.ProductInventoryResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/products/{id}/inventory [patch]
func (h *ProductHandler) UpdateInventory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateInventoryRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	inventory, err := h.service.UpdateInventory(r.Context(), id, appsvc.InventoryUpdateInput{
		Quantity:          req.Quantity,
		LowStockThreshold: req.LowStockThreshold,
		AdjustmentReason:  req.AdjustmentReason,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ProductInventoryResponse{
		Quantity:          inventory.Quantity,
		LowStockThreshold: inventory.LowStockThreshold,
		IsLowStock:        inventory.IsLowStock(),
		IsOutOfStock:      inventory.IsOutOfStock(),
	})
}

func parseUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}

func parseListFilter(r *http.Request) domain.ListFilter {
	q := r.URL.Query()
	filter := domain.ListFilter{
		Status:     q.Get("status"),
		Brand:      q.Get("brand"),
		StockLevel: q.Get("stock_level"),
	}
	if cid := q.Get("category_id"); cid != "" {
		if id, err := uuid.Parse(cid); err == nil {
			filter.CategoryID = &id
		}
	}
	if featured := q.Get("is_featured"); featured != "" {
		val := featured == "true"
		filter.IsFeatured = &val
	}
	return filter
}

func toCreateInput(req request.CreateProductRequest) (appsvc.CreateInput, error) {
	input := appsvc.CreateInput{
		Name:             req.Name,
		Slug:             req.Slug,
		SKU:              req.SKU,
		Description:      req.Description,
		ShortDescription: req.ShortDescription,
		Price:            req.Price,
		SalePrice:        req.SalePrice,
		Brand:            req.Brand,
		IsFeatured:       req.IsFeatured,
		Status:           req.Status,
		Inventory: appsvc.InventoryInput{
			Quantity:          req.Inventory.Quantity,
			LowStockThreshold: req.Inventory.LowStockThreshold,
		},
	}
	if req.CategoryID != nil {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return input, err
		}
		input.CategoryID = &id
	}
	for _, img := range req.Images {
		input.Images = append(input.Images, appsvc.ImageInput{
			URL: img.URL, AltText: img.AltText, SortOrder: img.SortOrder,
		})
	}
	for _, attr := range req.Attributes {
		input.Attributes = append(input.Attributes, appsvc.AttributeInput{
			Name: attr.Name, Value: attr.Value,
		})
	}
	return input, nil
}

func toUpdateInput(req request.UpdateProductRequest) (appsvc.UpdateInput, error) {
	input := appsvc.UpdateInput{
		Name: req.Name, Slug: req.Slug, SKU: req.SKU,
		Description: req.Description, ShortDescription: req.ShortDescription,
		Price: req.Price, SalePrice: req.SalePrice, Brand: req.Brand,
		IsFeatured: req.IsFeatured, Status: req.Status,
	}
	if req.CategoryID != nil {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			return input, err
		}
		input.CategoryID = &id
	}
	if req.Images != nil {
		images := make([]appsvc.ImageInput, len(*req.Images))
		for i, img := range *req.Images {
			images[i] = appsvc.ImageInput{URL: img.URL, AltText: img.AltText, SortOrder: img.SortOrder}
		}
		input.Images = &images
	}
	if req.Attributes != nil {
		attrs := make([]appsvc.AttributeInput, len(*req.Attributes))
		for i, attr := range *req.Attributes {
			attrs[i] = appsvc.AttributeInput{Name: attr.Name, Value: attr.Value}
		}
		input.Attributes = &attrs
	}
	return input, nil
}
