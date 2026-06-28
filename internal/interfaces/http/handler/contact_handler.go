package handler

import (
	"log/slog"
	"net/http"
	"time"

	appcontact "app/internal/application/contact"
	domaincontact "app/internal/domain/contact"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

type ContactHandler struct {
	service   *appcontact.Service
	validator *validator.Validator
	log       *slog.Logger
}

func NewContactHandler(service *appcontact.Service, v *validator.Validator, log *slog.Logger) *ContactHandler {
	return &ContactHandler{
		service:   service,
		validator: v,
		log:       log,
	}
}

// Submit godoc
// @Summary      Submit contact message
// @Description  Submit a contact inquiry form.
// @Tags         store-contact
// @Accept       json
// @Produce      json
// @Param        body  body  request.ContactMessageSubmitRequest  true  "Contact form data"
// @Success      201  {object}  dtoresponse.ContactMessageResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/contact [post]
func (h *ContactHandler) Submit(w http.ResponseWriter, r *http.Request) {
	var req request.ContactMessageSubmitRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	msg, err := h.service.Submit(r.Context(), appcontact.SubmitInput{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Subject: req.Subject,
		Message: req.Message,
		Source:  req.Source,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.ToContactMessageResponse(*msg))
}

// List godoc
// @Summary      List contact messages (Admin)
// @Description  Get a paginated list of contact messages.
// @Tags         admin-contact
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Param        status    query  string  false  "Status filter (unread/read/archived)"
// @Param        source    query  string  false  "Source filter (homepage/about/contact_page)"
// @Param        q         query  string  false  "Search query by name, email, or subject"
// @Param        from      query  string  false  "Filter messages created on or after date (YYYY-MM-DD)"
// @Param        to        query  string  false  "Filter messages created on or before date (YYYY-MM-DD)"
// @Success      200  {object}  dtoresponse.ContactMessageListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/contact-messages [get]
func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	q := r.URL.Query()
	filter := domaincontact.ListFilter{
		Status: q.Get("status"),
		Source: q.Get("source"),
		Query:  q.Get("q"),
	}

	if fromRaw := q.Get("from"); fromRaw != "" {
		if from, err := time.Parse("2006-01-02", fromRaw); err == nil {
			filter.From = &from
		}
	}
	if toRaw := q.Get("to"); toRaw != "" {
		if to, err := time.Parse("2006-01-02", toRaw); err == nil {
			filter.To = &to
		}
	}

	result, err := h.service.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToContactMessageListResponse(result))
}

// Get godoc
// @Summary      Get contact message (Admin)
// @Description  Get details of a single contact message.
// @Tags         admin-contact
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Message ID"
// @Success      200  {object}  dtoresponse.ContactMessageResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/contact-messages/{id} [get]
func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	msg, err := h.service.FindByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToContactMessageResponse(*msg))
}

// UpdateStatus godoc
// @Summary      Update contact status (Admin)
// @Description  Update status of a contact message.
// @Tags         admin-contact
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                                     true  "Message ID"
// @Param        body  body  request.ContactMessageStatusUpdateRequest  true  "Status update"
// @Success      204
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/contact-messages/{id}/status [patch]
func (h *ContactHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.ContactMessageStatusUpdateRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	err = h.service.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.NoContent(w)
}

// Delete godoc
// @Summary      Delete contact message (Admin)
// @Description  Delete a contact message.
// @Tags         admin-contact
// @Security     BearerAuth
// @Param        id  path  string  true  "Message ID"
// @Success      204
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/contact-messages/{id} [delete]
func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
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
