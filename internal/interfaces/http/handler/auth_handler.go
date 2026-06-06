package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"app/internal/application/auth"
	"app/internal/interfaces/http/dto/request"
	"app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/validator"
)

// AuthHandler handles authentication HTTP endpoints.
type AuthHandler struct {
	authService *auth.AuthService
	validator   *validator.Validator
	log         *slog.Logger
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authService *auth.AuthService, v *validator.Validator, log *slog.Logger) *AuthHandler {
	return &AuthHandler{authService: authService, validator: v, log: log}
}

// Login handles POST /api/v1/admin/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.LoginRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, h.log, err)
		return
	}

	tokens, err := h.authService.Login(r.Context(), auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.OK(w, tokens)
}

// Refresh handles POST /api/v1/admin/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req request.RefreshRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, h.log, err)
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), auth.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.OK(w, tokens)
}

// Logout handles POST /api/v1/admin/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req request.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, h.log, err)
		return
	}

	if err := h.authService.Logout(r.Context(), auth.LogoutInput{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.NoContent(w)
}

// Me handles GET /api/v1/admin/auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		response.Error(w, h.log, err)
		return
	}

	response.OK(w, user)
}

func decodeAndValidate[T any](r *http.Request, req *T, v *validator.Validator) error {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}
	return v.Validate(req)
}
