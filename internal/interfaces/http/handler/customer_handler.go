package handler

import (
	"log/slog"
	"net/http"

	appsvc "app/internal/application/customer"
	domain "app/internal/domain/customer"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

var _ = dtoresponse.CustomerDetailResponse{}

// CustomerHandler handles customer HTTP endpoints.
type CustomerHandler struct {
	service   *appsvc.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewCustomerHandler creates a new CustomerHandler.
func NewCustomerHandler(service *appsvc.Service, v *validator.Validator, log *slog.Logger) *CustomerHandler {
	return &CustomerHandler{service: service, validator: v, log: log}
}

// List godoc
// @Summary      List customers
// @Description  Get a paginated list of customers with optional search by name or email.
// @Tags         customers
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Param        sort      query  string  false  "Sort field"      Enums(created_at, email, first_name, last_name, total_orders, total_spent, updated_at)
// @Param        order     query  string  false  "Sort order"      Enums(asc, desc)
// @Param        q         query  string  false  "Search by name or email"
// @Param        type      query  string  false  "Filter by customer type"  Enums(registered, guest)
// @Success      200  {object}  dtoresponse.CustomerListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/customers [get]
func (h *CustomerHandler) List(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := parseCustomerListFilter(r)

	result, err := h.service.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToCustomerListResponse(result))
}

// Get godoc
// @Summary      Get customer
// @Description  Get a customer by ID with addresses and purchase statistics.
// @Tags         customers
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Customer ID"
// @Success      200  {object}  dtoresponse.CustomerDetailResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/customers/{id} [get]
func (h *CustomerHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	customer, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToCustomerDetailResponse(customer))
}

// ListOrders godoc
// @Summary      Customer purchase history
// @Description  Get paginated order history for a customer.
// @Tags         customers
// @Produce      json
// @Security     BearerAuth
// @Param        id        path   string  true   "Customer ID"
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Param        sort      query  string  false  "Sort field"      Enums(created_at, order_number, total, status)
// @Param        order     query  string  false  "Sort order"      Enums(asc, desc)
// @Success      200  {object}  dtoresponse.CustomerOrderListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/customers/{id}/orders [get]
func (h *CustomerHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	page := pagination.FromRequest(r)
	result, err := h.service.GetPurchaseHistory(r.Context(), id, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToCustomerOrderListResponse(result))
}

// Update godoc
// @Summary      Update customer
// @Description  Update a storefront customer profile.
// @Tags         customers
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                       true  "Customer ID"
// @Param        body  body  request.UpdateCustomerRequest  true  "Customer data"
// @Success      200   {object}  dtoresponse.CustomerDetailResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/customers/{id} [put]
func (h *CustomerHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateCustomerRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	customer, err := h.service.Update(r.Context(), id, appsvc.UpdateInput{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Type:      req.Type,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToCustomerDetailResponse(customer))
}

// Delete godoc
// @Summary      Delete customer
// @Description  Delete a customer without existing orders.
// @Tags         customers
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Customer ID"
// @Success      204
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Failure      422  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/customers/{id} [delete]
func (h *CustomerHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

func parseCustomerListFilter(r *http.Request) domain.ListFilter {
	q := r.URL.Query()
	filter := domain.ListFilter{Query: q.Get("q")}
	if t := q.Get("type"); t != "" {
		if ct, err := domain.ParseCustomerType(t); err == nil {
			filter.Type = &ct
		}
	}
	return filter
}
