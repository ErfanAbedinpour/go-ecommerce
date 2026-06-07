package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"app/internal/application/auth"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/dto/request"
	"app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/validator"
)

var _ = dtoresponse.TokenResponse{}

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
// @Success      200   {object}  dtoresponse.TokenResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req request.LoginRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	tokens, err := h.authService.Login(r.Context(), auth.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
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
// @Success      200   {object}  dtoresponse.TokenResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req request.RefreshRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), auth.RefreshInput{
		RefreshToken: req.RefreshToken,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
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
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req request.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.authService.Logout(r.Context(), auth.LogoutInput{
		RefreshToken: req.RefreshToken,
	}); err != nil {
		response.Error(w, r, h.log, err)
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
// @Success      200  {object}  dtoresponse.CurrentUserResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/auth/me [get]
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, user)
}

// Signup godoc
// @Summary      Sign up
// @Description  Register a new account. Returns JWT tokens on success. Role is assigned from AUTH_SIGNUP_DEFAULT_ROLE.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.SignupRequest  true  "Registration details"
// @Success      201   {object}  dtoresponse.TokenResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      409   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/auth/signup [post]
func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req request.SignupRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	tokens, err := h.authService.Signup(r.Context(), auth.SignupInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, tokens)
}

// ForgotPassword godoc
// @Summary      Forgot password
// @Description  Request a password reset email. Always returns success to avoid email enumeration.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.ForgotPasswordRequest  true  "Account email"
// @Success      200   {object}  dtoresponse.MessageResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req request.ForgotPasswordRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	result, err := h.authService.ForgotPassword(r.Context(), auth.ForgotPasswordInput{
		Email: req.Email,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, result)
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Set a new password using the token from the reset email link.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      request.ResetPasswordRequest  true  "Reset token and new password"
// @Success      200   {object}  dtoresponse.MessageResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/auth/reset-password [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req request.ResetPasswordRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	result, err := h.authService.ResetPassword(r.Context(), auth.ResetPasswordInput{
		Token:    req.Token,
		Password: req.Password,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, result)
}

func decodeAndValidate[T any](r *http.Request, req *T, v *validator.Validator) error {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}
	return v.Validate(req)
}
