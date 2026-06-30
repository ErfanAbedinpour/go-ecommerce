package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"

	appquestion "app/internal/application/productquestion"
	domainquestion "app/internal/domain/productquestion"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

type ProductQuestionHandler struct {
	service   *appquestion.Service
	validator *validator.Validator
	log       *slog.Logger
}

func NewProductQuestionHandler(service *appquestion.Service, v *validator.Validator, log *slog.Logger) *ProductQuestionHandler {
	return &ProductQuestionHandler{
		service:   service,
		validator: v,
		log:       log,
	}
}

func (h *ProductQuestionHandler) resolveProductID(r *http.Request) (uuid.UUID, error) {
	return h.service.ResolveProductID(r.Context(), chi.URLParam(r, "productId"))
}

// Ask godoc
// @Summary      Ask a question
// @Description  Submit a question about a product.
// @Tags         store-questions
// @Accept       json
// @Produce      json
// @Param        productId  path  string                             true  "Product ID"
// @Param        body       body  request.ProductQuestionAskRequest  true  "Question data"
// @Success      201  {object}  dtoresponse.ProductQuestionResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/products/{productId}/questions [post]
func (h *ProductQuestionHandler) Ask(w http.ResponseWriter, r *http.Request) {
	prodID, err := h.resolveProductID(r)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.ProductQuestionAskRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	q, err := h.service.Ask(r.Context(), appquestion.AskInput{
		ProductID:  prodID,
		AskerName:  req.AskerName,
		AskerEmail: req.AskerEmail,
		Question:   req.Question,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.ToProductQuestionResponse(*q))
}

// ListByProduct godoc
// @Summary      List questions
// @Description  Get a paginated list of answered questions for a product.
// @Tags         store-questions
// @Produce      json
// @Param        productId  path   string  true  "Product ID"
// @Param        page       query  int     false  "Page number"     default(1)
// @Param        per_page   query  int     false  "Items per page"  default(20)
// @Success      200  {object}  dtoresponse.ProductQuestionListResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/products/{productId}/questions [get]
func (h *ProductQuestionHandler) ListByProduct(w http.ResponseWriter, r *http.Request) {
	prodID, err := h.resolveProductID(r)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	page := pagination.FromRequest(r)
	result, err := h.service.ListByProduct(r.Context(), prodID, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToProductQuestionListResponse(result))
}

// ListAdmin godoc
// @Summary      List questions (Admin)
// @Description  Get a paginated list of product questions.
// @Tags         admin-questions
// @Produce      json
// @Security     BearerAuth
// @Param        page        query  int     false  "Page number"     default(1)
// @Param        per_page    query  int     false  "Items per page"  default(20)
// @Param        product_id  query  string  false  "Product ID filter"
// @Param        status      query  string  false  "Status filter (open/answered)"
// @Param        q           query  string  false  "Search query by question content or asker name"
// @Success      200  {object}  dtoresponse.AdminProductQuestionListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/questions [get]
func (h *ProductQuestionHandler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := domainquestion.ListFilter{
		Status: r.URL.Query().Get("status"),
		Query:  r.URL.Query().Get("q"),
	}

	if pid := r.URL.Query().Get("product_id"); pid != "" {
		if id, err := uuid.Parse(pid); err == nil {
			filter.ProductID = &id
		}
	}

	result, err := h.service.ListAdmin(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToAdminProductQuestionListResponse(result))
}

// Answer godoc
// @Summary      Answer question (Admin)
// @Description  Answer a product question.
// @Tags         admin-questions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                                true  "Question ID"
// @Param        body  body  request.ProductQuestionAnswerRequest  true  "Answer data"
// @Success      204
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/questions/{id}/answer [post]
func (h *ProductQuestionHandler) Answer(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.ProductQuestionAnswerRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	answeredBy, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	err = h.service.Answer(r.Context(), id, req.Answer, answeredBy)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.NoContent(w)
}

// Delete godoc
// @Summary      Delete question (Admin)
// @Description  Delete a product question.
// @Tags         admin-questions
// @Security     BearerAuth
// @Param        id  path  string  true  "Question ID"
// @Success      204
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/questions/{id} [delete]
func (h *ProductQuestionHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
