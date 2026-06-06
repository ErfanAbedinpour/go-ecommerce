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

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password. Returns JWT access and refresh tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.LoginRequest  true  "Login credentials"
// @Success      200   {object}  response.TokenResponse
// @Failure      400   {object}  response.ErrorResponse
// @Failure      401   {object}  response.ErrorResponse
// @Failure      403   {object}  response.ErrorResponse
// @Router       /api/v1/auth/login [post]
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

// Refresh godoc
// @Summary      Refresh token
// @Description  Exchange a valid refresh token for a new access/refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.RefreshRequest  true  "Refresh token"
// @Success      200   {object}  response.TokenResponse
// @Failure      400   {object}  response.ErrorResponse
// @Failure      401   {object}  response.ErrorResponse
// @Router       /api/v1/auth/refresh [post]
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

// Logout godoc
// @Summary      Logout
// @Description  Revoke the refresh token family. Requires a valid access token.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  request.LogoutRequest  false  "Optional refresh token to revoke"
// @Success      204   "No Content"
// @Failure      401   {object}  response.ErrorResponse
// @Router       /api/v1/auth/logout [post]
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

// Me godoc
// @Summary      Get current user
// @Description  Returns the profile of the authenticated user.
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.CurrentUserResponse
// @Failure      401  {object}  response.ErrorResponse
// @Router       /api/v1/auth/me [get]
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
