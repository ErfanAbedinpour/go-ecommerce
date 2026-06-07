package handler

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	appsvc "app/internal/application/dashboard"
	domain "app/internal/domain/dashboard"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

var _ = dtoresponse.DashboardStatsResponse{}

// DashboardHandler handles dashboard analytics HTTP endpoints.
type DashboardHandler struct {
	service   *appsvc.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(service *appsvc.Service, v *validator.Validator, log *slog.Logger) *DashboardHandler {
	return &DashboardHandler{service: service, validator: v, log: log}
}

// Stats godoc
// @Summary      Dashboard statistics
// @Description  Get aggregated KPIs: revenue, orders, customers, products, pending orders, and low stock count.
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.DashboardStatsResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/dashboard/stats [get]
func (h *DashboardHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStats(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToDashboardStatsResponse(stats))
}

// Revenue godoc
// @Summary      Revenue analytics
// @Description  Get time-series revenue and order counts. Use period preset or custom from/to dates.
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Param        period  query  string  false  "Preset period"  Enums(today, week, month, year)  default(month)
// @Param        from    query  string  false  "Custom start date (YYYY-MM-DD)"
// @Param        to      query  string  false  "Custom end date (YYYY-MM-DD)"
// @Success      200  {object}  dtoresponse.RevenueAnalyticsResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/dashboard/revenue [get]
func (h *DashboardHandler) Revenue(w http.ResponseWriter, r *http.Request) {
	filter, err := parseRevenueFilter(r)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	points, err := h.service.GetRevenueAnalytics(r.Context(), filter)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToRevenueAnalyticsResponse(points))
}

// LowStock godoc
// @Summary      Low stock products
// @Description  Get paginated products where quantity is at or below the low-stock threshold.
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Param        sort      query  string  false  "Sort field"      Enums(created_at, name, quantity)
// @Param        order     query  string  false  "Sort order"      Enums(asc, desc)
// @Success      200  {object}  dtoresponse.ProductListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/dashboard/low-stock [get]
func (h *DashboardHandler) LowStock(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	result, err := h.service.GetLowStockProducts(r.Context(), page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToLowStockProductListResponse(result))
}

// RecentOrders godoc
// @Summary      Recent orders
// @Description  Get the latest orders for the dashboard feed.
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query  int  false  "Number of orders to return (max 50)"  default(10)
// @Success      200  {object}  dtoresponse.RecentOrdersResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/dashboard/recent-orders [get]
func (h *DashboardHandler) RecentOrders(w http.ResponseWriter, r *http.Request) {
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			response.Error(w, r, h.log, domain.ErrInvalidLimit)
			return
		}
		limit = parsed
	}

	orders, err := h.service.GetRecentOrders(r.Context(), limit)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToRecentOrdersResponse(orders))
}

// FeaturedProducts godoc
// @Summary      Featured products
// @Description  Get active featured products for the dashboard widget.
// @Tags         dashboard
// @Produce      json
// @Security     BearerAuth
// @Param        limit  query  int  false  "Number of products to return (max 20)"  default(5)
// @Success      200  {object}  dtoresponse.FeaturedProductsResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/dashboard/featured-products [get]
func (h *DashboardHandler) FeaturedProducts(w http.ResponseWriter, r *http.Request) {
	limit := 5
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			response.Error(w, r, h.log, domain.ErrInvalidLimit)
			return
		}
		limit = parsed
	}

	products, err := h.service.GetFeaturedProducts(r.Context(), limit)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToFeaturedProductsResponse(products))
}

func parseRevenueFilter(r *http.Request) (domain.RevenueFilter, error) {
	q := r.URL.Query()
	filter := domain.RevenueFilter{
		Period: domain.Period(q.Get("period")),
	}

	fromRaw := q.Get("from")
	toRaw := q.Get("to")
	if fromRaw != "" || toRaw != "" {
		if fromRaw == "" || toRaw == "" {
			return filter, domain.ErrInvalidDateRange
		}
		from, err := time.Parse("2006-01-02", fromRaw)
		if err != nil {
			return filter, domain.ErrInvalidDateRange
		}
		to, err := time.Parse("2006-01-02", toRaw)
		if err != nil {
			return filter, domain.ErrInvalidDateRange
		}
		filter.From = &from
		filter.To = &to
		return filter, nil
	}

	if _, err := domain.ParsePeriod(string(filter.Period)); err != nil {
		return filter, err
	}
	return filter, nil
}
