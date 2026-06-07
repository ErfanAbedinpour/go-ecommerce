package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	appsettings "app/internal/application/settings"
	domain "app/internal/domain/settings"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/validator"
)

var _ = dtoresponse.SiteSettingsResponse{}

// SettingsHandler handles store settings HTTP endpoints.
type SettingsHandler struct {
	service   *appsettings.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(service *appsettings.Service, v *validator.Validator, log *slog.Logger) *SettingsHandler {
	return &SettingsHandler{service: service, validator: v, log: log}
}

// GetSite godoc
// @Summary      Get site settings
// @Description  Returns storefront identity settings (name, URL, logo, favicon).
// @Tags         settings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.SiteSettingsResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/settings/site [get]
func (h *SettingsHandler) GetSite(w http.ResponseWriter, r *http.Request) {
	site, err := h.service.GetSite(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToSiteSettingsResponse(site))
}

// UpdateSite godoc
// @Summary      Update site settings
// @Description  Updates storefront identity settings.
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.UpdateSiteSettingsRequest  true  "Site settings"
// @Success      200   {object}  dtoresponse.SiteSettingsResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/settings/site [put]
func (h *SettingsHandler) UpdateSite(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateSiteSettingsRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	site, err := h.service.UpdateSite(r.Context(), domain.Site{
		Name:       req.Name,
		URL:        req.URL,
		LogoURL:    req.LogoURL,
		FaviconURL: req.FaviconURL,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToSiteSettingsResponse(site))
}

// GetContact godoc
// @Summary      Get contact settings
// @Description  Returns public contact information.
// @Tags         settings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.ContactSettingsResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/settings/contact [get]
func (h *SettingsHandler) GetContact(w http.ResponseWriter, r *http.Request) {
	contact, err := h.service.GetContact(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToContactSettingsResponse(contact))
}

// UpdateContact godoc
// @Summary      Update contact settings
// @Description  Updates public contact information.
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.UpdateContactSettingsRequest  true  "Contact settings"
// @Success      200   {object}  dtoresponse.ContactSettingsResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/settings/contact [put]
func (h *SettingsHandler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateContactSettingsRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	contact, err := h.service.UpdateContact(r.Context(), domain.Contact{
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
		City:    req.City,
		Country: req.Country,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToContactSettingsResponse(contact))
}

// GetSocial godoc
// @Summary      Get social settings
// @Description  Returns social media profile URLs.
// @Tags         settings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.SocialSettingsResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/settings/social [get]
func (h *SettingsHandler) GetSocial(w http.ResponseWriter, r *http.Request) {
	social, err := h.service.GetSocial(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToSocialSettingsResponse(social))
}

// UpdateSocial godoc
// @Summary      Update social settings
// @Description  Updates social media profile URLs.
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.UpdateSocialSettingsRequest  true  "Social settings"
// @Success      200   {object}  dtoresponse.SocialSettingsResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/settings/social [put]
func (h *SettingsHandler) UpdateSocial(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateSocialSettingsRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	social, err := h.service.UpdateSocial(r.Context(), domain.Social{
		Facebook:  req.Facebook,
		Twitter:   req.Twitter,
		Instagram: req.Instagram,
		LinkedIn:  req.LinkedIn,
		YouTube:   req.YouTube,
		TikTok:    req.TikTok,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToSocialSettingsResponse(social))
}

// GetSEO godoc
// @Summary      Get SEO settings
// @Description  Returns search-engine and metadata configuration.
// @Tags         settings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.SEOSettingsResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/settings/seo [get]
func (h *SettingsHandler) GetSEO(w http.ResponseWriter, r *http.Request) {
	seo, err := h.service.GetSEO(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToSEOSettingsResponse(seo))
}

// UpdateSEO godoc
// @Summary      Update SEO settings
// @Description  Updates search-engine and metadata configuration.
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.UpdateSEOSettingsRequest  true  "SEO settings"
// @Success      200   {object}  dtoresponse.SEOSettingsResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/settings/seo [put]
func (h *SettingsHandler) UpdateSEO(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateSEOSettingsRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	seo, err := h.service.UpdateSEO(r.Context(), domain.SEO{
		MetaTitle:         req.MetaTitle,
		MetaDescription:   req.MetaDescription,
		MetaKeywords:      req.MetaKeywords,
		OGImageURL:        req.OGImageURL,
		RobotsTxt:         req.RobotsTxt,
		GoogleAnalyticsID: req.GoogleAnalyticsID,
		SitemapEnabled:    req.SitemapEnabled,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToSEOSettingsResponse(seo))
}

// GetNavigation godoc
// @Summary      Get navigation menu
// @Description  Returns the storefront navigation menu tree.
// @Tags         settings
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.NavigationResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/navigation [get]
func (h *SettingsHandler) GetNavigation(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.GetNavigation(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToNavigationResponse(items))
}

// UpdateNavigation godoc
// @Summary      Update navigation menu
// @Description  Replaces the storefront navigation menu tree.
// @Tags         settings
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.UpdateNavigationRequest  true  "Navigation tree"
// @Success      200   {object}  dtoresponse.NavigationResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/navigation [put]
func (h *SettingsHandler) UpdateNavigation(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateNavigationRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	items, err := h.service.UpdateNavigation(r.Context(), toNavItems(req.Items))
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToNavigationResponse(items))
}

func toNavItems(items []request.NavItemRequest) []domain.NavItem {
	result := make([]domain.NavItem, len(items))
	for i, item := range items {
		id := item.ID
		if id == "" {
			id = uuid.New().String()
		}
		result[i] = domain.NavItem{
			ID:        id,
			Label:     item.Label,
			URL:       item.URL,
			SortOrder: item.SortOrder,
			IsActive:  item.IsActive,
			Children:  toNavItems(item.Children),
		}
	}
	return result
}
