package handler

import (
	"log/slog"
	"net/http"

	appsvc "app/internal/application/order"
	domain "app/internal/domain/order"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

var _ = dtoresponse.OrderDetailResponse{}

// OrderHandler handles order HTTP endpoints.
type OrderHandler struct {
	service   *appsvc.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewOrderHandler creates a new OrderHandler.
func NewOrderHandler(service *appsvc.Service, v *validator.Validator, log *slog.Logger) *OrderHandler {
	return &OrderHandler{service: service, validator: v, log: log}
}

// List godoc
// @Summary      List orders
// @Description  Get a paginated list of orders with optional search and filters.
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Param        page            query  int     false  "Page number"     default(1)
// @Param        per_page        query  int     false  "Items per page"  default(20)
// @Param        sort            query  string  false  "Sort field"      Enums(created_at, order_number, total, status, payment_status)
// @Param        order           query  string  false  "Sort order"      Enums(asc, desc)
// @Param        status          query  string  false  "Filter by order status"
// @Param        payment_status  query  string  false  "Filter by payment status"
// @Param        q               query  string  false  "Search by order number or customer"
// @Success      200  {object}  dtoresponse.OrderListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/orders [get]
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := parseOrderListFilter(r)

	result, err := h.service.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToOrderListResponse(result))
}

// Get godoc
// @Summary      Get order
// @Description  Get full order detail with items, customer, addresses, and status timeline.
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Order ID"
// @Success      200  {object}  dtoresponse.OrderDetailResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/orders/{id} [get]
func (h *OrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	order, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToOrderDetailResponse(order))
}

// UpdateStatus godoc
// @Summary      Update order status
// @Description  Transition order through the fulfillment workflow. Invalid transitions return 422.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                           true  "Order ID"
// @Param        body  body  request.UpdateOrderStatusRequest  true  "Status update"
// @Success      200   {object}  dtoresponse.OrderDetailResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      422   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/orders/{id}/status [patch]
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateOrderStatusRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	changedBy, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	order, err := h.service.UpdateStatus(r.Context(), id, appsvc.UpdateStatusInput{
		Status: req.Status, Note: req.Note, ChangedBy: changedBy,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToOrderDetailResponse(order))
}

// Cancel godoc
// @Summary      Cancel order
// @Description  Cancel a pending or processing order and restore product inventory.
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Order ID"
// @Success      200  {object}  dtoresponse.OrderDetailResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Failure      422  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/orders/{id}/cancel [post]
func (h *OrderHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	changedBy, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	order, err := h.service.Cancel(r.Context(), id, changedBy)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToOrderDetailResponse(order))
}

// Refund godoc
// @Summary      Refund order
// @Description  Process a full or partial refund on a delivered, paid order.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                     true  "Order ID"
// @Param        body  body  request.RefundOrderRequest  true  "Refund data"
// @Success      200   {object}  dtoresponse.OrderDetailResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      422   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/orders/{id}/refund [post]
func (h *OrderHandler) Refund(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.RefundOrderRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	changedBy, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	order, err := h.service.Refund(r.Context(), id, appsvc.RefundInput{
		Amount: req.Amount, Reason: req.Reason, ChangedBy: changedBy,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToOrderDetailResponse(order))
}

func parseOrderListFilter(r *http.Request) domain.ListFilter {
	q := r.URL.Query()
	return domain.ListFilter{
		Status:        q.Get("status"),
		PaymentStatus: q.Get("payment_status"),
		Query:         q.Get("q"),
	}
}
