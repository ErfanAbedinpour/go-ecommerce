package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

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
// @Param        from            query  string  false  "Filter orders created on or after date (YYYY-MM-DD)"
// @Param        to              query  string  false  "Filter orders created on or before date (YYYY-MM-DD)"
// @Success      200  {object}  dtoresponse.OrderListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/orders [get]
func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter, err := parseOrderListFilter(r)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

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

// Create godoc
// @Summary      Create order
// @Description  Create a manual order with line items, addresses, and optional coupon.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  request.CreateOrderRequest  true  "Order data"
// @Success      201   {object}  dtoresponse.OrderDetailResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      422   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/orders [post]
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateOrderRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	changedBy, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	input, err := toCreateOrderInput(req, changedBy)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	order, err := h.service.Create(r.Context(), input)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToOrderDetailResponse(order))
}

// GetInvoice godoc
// @Summary      Get order invoice
// @Description  Get printable invoice payload with store and order details.
// @Tags         orders
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Order ID"
// @Success      200  {object}  dtoresponse.OrderInvoiceResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/orders/{id}/invoice [get]
func (h *OrderHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	invoice, err := h.service.GetInvoice(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToOrderInvoiceResponse(invoice))
}

// UpdateNotes godoc
// @Summary      Update order notes
// @Description  Save internal order notes without changing status.
// @Tags         orders
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                       true  "Order ID"
// @Param        body  body  request.UpdateOrderNotesRequest  true  "Notes"
// @Success      200   {object}  dtoresponse.OrderDetailResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/orders/{id}/notes [patch]
func (h *OrderHandler) UpdateNotes(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateOrderNotesRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	order, err := h.service.UpdateNotes(r.Context(), id, appsvc.UpdateNotesInput{Notes: req.Notes})
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

func parseOrderListFilter(r *http.Request) (domain.ListFilter, error) {
	q := r.URL.Query()
	filter := domain.ListFilter{
		Status:        q.Get("status"),
		PaymentStatus: q.Get("payment_status"),
		Query:         q.Get("q"),
	}

	fromRaw := q.Get("from")
	toRaw := q.Get("to")
	if fromRaw == "" && toRaw == "" {
		return filter, nil
	}

	var from, to time.Time
	var err error
	if fromRaw != "" {
		from, err = time.Parse("2006-01-02", fromRaw)
		if err != nil {
			return filter, domain.ErrInvalidDateRange
		}
		filter.From = &from
	}
	if toRaw != "" {
		to, err = time.Parse("2006-01-02", toRaw)
		if err != nil {
			return filter, domain.ErrInvalidDateRange
		}
		filter.To = &to
	}
	if filter.From != nil && filter.To != nil && filter.From.After(*filter.To) {
		return filter, domain.ErrInvalidDateRange
	}
	return filter, nil
}

func toCreateOrderInput(req request.CreateOrderRequest, changedBy uuid.UUID) (appsvc.CreateInput, error) {
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return appsvc.CreateInput{}, domain.ErrNotFound
	}

	items := make([]appsvc.CreateItemInput, len(req.Items))
	for i, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			return appsvc.CreateInput{}, domain.ErrNotFound
		}
		items[i] = appsvc.CreateItemInput{ProductID: productID, Quantity: item.Quantity}
	}

	return appsvc.CreateInput{
		CustomerID:      customerID,
		Items:           items,
		CouponCode:      req.CouponCode,
		ShippingAmount:  req.ShippingAmount,
		TaxAmount:       req.TaxAmount,
		BillingAddress:  toOrderAddress(req.BillingAddress),
		ShippingAddress: toOrderAddress(req.ShippingAddress),
		PaymentMethod:   req.PaymentMethod,
		TransactionID:   req.TransactionID,
		PaymentStatus:   req.PaymentStatus,
		Notes:           req.Notes,
		ChangedBy:       changedBy,
	}, nil
}

func toOrderAddress(req request.OrderAddressRequest) domain.Address {
	return domain.Address{
		Street:     req.Street,
		City:       req.City,
		State:      req.State,
		PostalCode: req.PostalCode,
		Country:    req.Country,
	}
}
