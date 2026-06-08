package handler

import (
	"log/slog"
	"net/http"

	appuser "app/internal/application/adminuser"
	domain "app/internal/domain/user"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

var _ = dtoresponse.AdminUserResponse{}

// UserHandler handles admin user HTTP endpoints.
type UserHandler struct {
	service   *appuser.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewUserHandler creates a new UserHandler.
func NewUserHandler(service *appuser.Service, v *validator.Validator, log *slog.Logger) *UserHandler {
	return &UserHandler{service: service, validator: v, log: log}
}

// List godoc
// @Summary      List admin users
// @Description  Get a paginated list of admin panel user accounts.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Param        sort      query  string  false  "Sort field"      Enums(created_at, email, first_name, last_name, updated_at)
// @Param        order     query  string  false  "Sort order"      Enums(asc, desc)
// @Param        q         query  string  false  "Search by name or email"
// @Param        role      query  string  false  "Filter by role"  Enums(admin, customer)
// @Success      200  {object}  dtoresponse.AdminUserListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/users [get]
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := parseAdminUserListFilter(r)

	result, err := h.service.List(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToAdminUserListResponse(result))
}

// Create godoc
// @Summary      Create admin user
// @Description  Create a new admin panel user account.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  request.CreateAdminUserRequest  true  "User data"
// @Success      201   {object}  dtoresponse.AdminUserResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/users [post]
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req request.CreateAdminUserRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	user, err := h.service.Create(r.Context(), appuser.CreateInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToAdminUserResponse(user))
}

// Get godoc
// @Summary      Get admin user
// @Description  Get an admin panel user by ID.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "User ID"
// @Success      200  {object}  dtoresponse.AdminUserResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/users/{id} [get]
func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToAdminUserResponse(user))
}

// Update godoc
// @Summary      Update admin user
// @Description  Update an admin panel user account.
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                         true  "User ID"
// @Param        body  body  request.UpdateAdminUserRequest  true  "User data"
// @Success      200   {object}  dtoresponse.AdminUserResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/users/{id} [put]
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateAdminUserRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	user, err := h.service.Update(r.Context(), id, appuser.UpdateInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToAdminUserResponse(user))
}

// Delete godoc
// @Summary      Delete admin user
// @Description  Soft-delete an admin user. Cannot delete yourself or the last active admin.
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "User ID"
// @Success      204
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Failure      422  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/users/{id} [delete]
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	actorID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.service.Delete(r.Context(), id, actorID); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.NoContent(w)
}

func parseAdminUserListFilter(r *http.Request) domain.ListFilter {
	q := r.URL.Query()
	filter := domain.ListFilter{Query: q.Get("q")}
	if role := q.Get("role"); role != "" {
		if parsed, err := domain.ParseRole(role); err == nil {
			filter.Role = &parsed
		}
	}
	return filter
}
