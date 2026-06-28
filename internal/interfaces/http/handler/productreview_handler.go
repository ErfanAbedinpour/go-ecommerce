package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	appreview "app/internal/application/productreview"
	domainreview "app/internal/domain/productreview"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

type ProductReviewHandler struct {
	service   *appreview.Service
	validator *validator.Validator
	log       *slog.Logger
}

func NewProductReviewHandler(service *appreview.Service, v *validator.Validator, log *slog.Logger) *ProductReviewHandler {
	return &ProductReviewHandler{
		service:   service,
		validator: v,
		log:       log,
	}
}

// Submit godoc
// @Summary      Submit a review
// @Description  Submit a review for a product. Customer ID is linked if authenticated.
// @Tags         store-reviews
// @Accept       json
// @Produce      json
// @Param        productId  path  string                              true  "Product ID"
// @Param        body       body  request.ProductReviewSubmitRequest  true  "Review data"
// @Success      201  {object}  dtoresponse.ProductReviewResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      409  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/products/{productId}/reviews [post]
func (h *ProductReviewHandler) Submit(w http.ResponseWriter, r *http.Request) {
	prodID, err := parseUUIDParam(r, "productId")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.ProductReviewSubmitRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var userID *uuid.UUID
	if id, ok := appmiddleware.GetUserIDOptional(r.Context()); ok {
		userID = &id
	}

	review, err := h.service.Submit(r.Context(), appreview.SubmitInput{
		ProductID:  prodID,
		UserID:     userID,
		AuthorName: req.AuthorName,
		Rating:     req.Rating,
		Title:      req.Title,
		Content:    req.Content,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.ToProductReviewResponse(*review))
}

// ListByProduct godoc
// @Summary      List reviews
// @Description  Get a paginated list of approved reviews for a product.
// @Tags         store-reviews
// @Produce      json
// @Param        productId  path   string  true  "Product ID"
// @Param        page       query  int     false  "Page number"     default(1)
// @Param        per_page   query  int     false  "Items per page"  default(20)
// @Param        sort       query  string  false  "Sort order (highest/lowest/newest)"
// @Success      200  {object}  dtoresponse.ProductReviewListResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/products/{productId}/reviews [get]
func (h *ProductReviewHandler) ListByProduct(w http.ResponseWriter, r *http.Request) {
	prodID, err := parseUUIDParam(r, "productId")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	page := pagination.FromRequest(r)
	sort := r.URL.Query().Get("sort")

	result, err := h.service.ListByProduct(r.Context(), prodID, sort, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToProductReviewListResponse(result))
}

// GetSummary godoc
// @Summary      Get reviews summary
// @Description  Get aggregate rating stats for a product.
// @Tags         store-reviews
// @Produce      json
// @Param        productId  path  string  true  "Product ID"
// @Success      200  {object}  dtoresponse.ProductReviewSummaryResponse
// @Failure      440  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/products/{productId}/reviews/summary [get]
func (h *ProductReviewHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	prodID, err := parseUUIDParam(r, "productId")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	summary, err := h.service.GetSummary(r.Context(), prodID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToProductReviewSummaryResponse(summary))
}

// ListAdmin godoc
// @Summary      List reviews (Admin)
// @Description  Get a paginated list of reviews for admin moderation.
// @Tags         admin-reviews
// @Produce      json
// @Security     BearerAuth
// @Param        page        query  int     false  "Page number"     default(1)
// @Param        per_page    query  int     false  "Items per page"  default(20)
// @Param        product_id  query  string  false  "Product ID filter"
// @Param        status      query  string  false  "Status filter (pending/approved/rejected)"
// @Param        rating      query  int     false  "Rating filter"
// @Param        q           query  string  false  "Search by author or content"
// @Success      200  {object}  dtoresponse.AdminProductReviewListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/reviews [get]
func (h *ProductReviewHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := domainreview.ListFilter{
		Status: r.URL.Query().Get("status"),
		Query:  r.URL.Query().Get("q"),
	}

	if pid := r.URL.Query().Get("product_id"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			filter.ProductID = &id
		}
	}
	if rating := r.URL.Query().Get("rating"); rating != "" {
		var val int
		if _, err := fmt.Sscanf(rating, "%d", &val); err == nil {
			filter.Rating = &val
		}
	}

	result, err := h.service.ListAdmin(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToAdminProductReviewListResponse(result))
}

// UpdateStatus godoc
// @Summary      Moderate review (Admin)
// @Description  Approve or reject a product review.
// @Tags         admin-reviews
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                                    true  "Review ID"
// @Param        body  body  request.ProductReviewStatusUpdateRequest  true  "Status update"
// @Success      204
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/reviews/{id}/status [patch]
func (h *ProductReviewHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.ProductReviewStatusUpdateRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	err = h.service.UpdateStatus(r.Context(), id, domainreview.Status(req.Status))
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.NoContent(w)
}

// Delete godoc
// @Summary      Delete review (Admin)
// @Description  Delete a product review.
// @Tags         admin-reviews
// @Security     BearerAuth
// @Param        id  path  string  true  "Review ID"
// @Success      204
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/reviews/{id} [delete]
func (h *ProductReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
