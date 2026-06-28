package handler

import (
	"log/slog"
	"net/http"

	apptheme "app/internal/application/theme"
	domaintheme "app/internal/domain/theme"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/validator"

	"github.com/google/uuid"
)

var _ = dtoresponse.ThemeResponse{}

// ThemeHandler handles admin theme marketplace HTTP endpoints.
type ThemeHandler struct {
	service   *apptheme.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewThemeHandler creates a new ThemeHandler.
func NewThemeHandler(service *apptheme.Service, v *validator.Validator, log *slog.Logger) *ThemeHandler {
	return &ThemeHandler{service: service, validator: v, log: log}
}

// ListThemes godoc
// @Summary      List themes
// @Description  Returns available themes with purchase status for the current admin user.
// @Tags         themes
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.ThemeListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/themes [get]
func (h *ThemeHandler) ListThemes(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	themes, err := h.service.ListThemes(r.Context(), userID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToThemeListResponse(themes))
}

// PurchaseTheme godoc
// @Summary      Purchase theme
// @Description  Records a theme purchase for the current admin user.
// @Tags         themes
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Theme ID"
// @Success      201  {object}  dtoresponse.ThemePurchaseResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Failure      409  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/themes/{id}/purchase [post]
func (h *ThemeHandler) PurchaseTheme(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	themeID, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	purchase, err := h.service.PurchaseTheme(r.Context(), themeID, userID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToThemePurchaseResponse(purchase))
}

// GetStoreStyle godoc
// @Summary      Get store style
// @Description  Returns the current store style with resolved design tokens.
// @Tags         themes
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.StoreStyleResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/store-style [get]
func (h *ThemeHandler) GetStoreStyle(w http.ResponseWriter, r *http.Request) {
	style, err := h.service.GetStyle(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToStoreStyleResponse(style))
}

// UpdateStoreStyle godoc
// @Summary      Update store style
// @Description  Updates active theme, color tokens, and font family.
// @Tags         themes
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.UpdateStoreStyleRequest  true  "Store style"
// @Success      200   {object}  dtoresponse.StoreStyleResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/store-style [put]
func (h *ThemeHandler) UpdateStoreStyle(w http.ResponseWriter, r *http.Request) {
	userID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.UpdateStoreStyleRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	input, err := toUpdateStyleInput(req)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	style, err := h.service.UpdateStyle(r.Context(), userID, input)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToStoreStyleResponse(style))
}

func toUpdateStyleInput(req request.UpdateStoreStyleRequest) (apptheme.UpdateStyleInput, error) {
	input := apptheme.UpdateStyleInput{}

	if req.ActiveThemeID != nil && *req.ActiveThemeID != "" {
		themeID, err := uuid.Parse(*req.ActiveThemeID)
		if err != nil {
			return input, err
		}
		input.ActiveThemeID = &themeID
	}

	if len(req.Colors) > 0 {
		colors := toColorTokens(req.Colors)
		input.Colors = &colors
	}

	if req.FontFamily != "" {
		input.FontFamily = &req.FontFamily
	}

	return input, nil
}

func toColorTokens(m map[string]string) domaintheme.ColorTokens {
	return domaintheme.ColorTokens{
		Primary:             m["primary"],
		PrimaryForeground:   m["primary_foreground"],
		Secondary:           m["secondary"],
		SecondaryForeground: m["secondary_foreground"],
		Accent:              m["accent"],
		AccentForeground:    m["accent_foreground"],
		Background:          m["background"],
		Foreground:          m["foreground"],
		Muted:               m["muted"],
		MutedForeground:     m["muted_foreground"],
		Border:              m["border"],
		Destructive:         m["destructive"],
	}
}
