package handler

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	appstorecontent "app/internal/application/storecontent"
	appsettings "app/internal/application/settings"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	"app/internal/interfaces/http/response"
	"app/pkg/validator"
)

var _ = dtoresponse.HeroResponse{}

// StorefrontHandler handles admin storefront CMS HTTP endpoints.
type StorefrontHandler struct {
	content   *appstorecontent.Service
	settings  *appsettings.Service
	validator *validator.Validator
	log       *slog.Logger
}

// NewStorefrontHandler creates a new StorefrontHandler.
func NewStorefrontHandler(
	content *appstorecontent.Service,
	settings *appsettings.Service,
	v *validator.Validator,
	log *slog.Logger,
) *StorefrontHandler {
	return &StorefrontHandler{
		content:   content,
		settings:  settings,
		validator: v,
		log:       log,
	}
}

// GetHero godoc
// @Summary      Get storefront hero
// @Description  Returns the singleton homepage hero configuration.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.HeroResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/hero [get]
func (h *StorefrontHandler) GetHero(w http.ResponseWriter, r *http.Request) {
	hero, err := h.content.GetHero(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToHeroResponse(hero))
}

// UpdateHero godoc
// @Summary      Update storefront hero
// @Description  Updates the singleton homepage hero configuration.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.UpdateHeroRequest  true  "Hero settings"
// @Success      200   {object}  dtoresponse.HeroResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/hero [put]
func (h *StorefrontHandler) UpdateHero(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateHeroRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	hero, err := h.content.UpdateHero(r.Context(), appstorecontent.UpdateHeroInput{
		VideoURL:         req.VideoURL,
		Title:            req.Title,
		Subtitle:         req.Subtitle,
		CTAPrimaryText:   req.CTAPrimaryText,
		CTAPrimaryURL:    req.CTAPrimaryURL,
		CTASecondaryText: req.CTASecondaryText,
		CTASecondaryURL:  req.CTASecondaryURL,
		IsActive:         req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToHeroResponse(hero))
}

// ListProductSlides godoc
// @Summary      List product slides
// @Description  Returns all homepage product carousel configurations.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.ProductSlideListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/product-slides [get]
func (h *StorefrontHandler) ListProductSlides(w http.ResponseWriter, r *http.Request) {
	slides, err := h.content.ListProductSlides(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToProductSlideListResponse(slides))
}

// UpdateProductSlide godoc
// @Summary      Update product slide
// @Description  Updates a product carousel configuration by slide type.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slideType  path      string                          true  "Slide type"  Enums(featured, bestseller, discounted)
// @Param        body       body      request.UpdateProductSlideRequest  true  "Slide settings"
// @Success      200        {object}  dtoresponse.ProductSlideResponse
// @Failure      400        {object}  dtoresponse.ErrorResponse
// @Failure      401        {object}  dtoresponse.ErrorResponse
// @Failure      403        {object}  dtoresponse.ErrorResponse
// @Failure      404        {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/product-slides/{slideType} [put]
func (h *StorefrontHandler) UpdateProductSlide(w http.ResponseWriter, r *http.Request) {
	slideType := chi.URLParam(r, "slideType")

	var req request.UpdateProductSlideRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	slide, err := h.content.UpdateProductSlide(r.Context(), slideType, appstorecontent.UpdateProductSlideInput{
		Title:              req.Title,
		AutoplayIntervalMs: req.AutoplayIntervalMs,
		IsActive:           req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToProductSlideResponse(slide))
}

// CreateSlideItem godoc
// @Summary      Add product to slide
// @Description  Adds a product to a product carousel slide.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        slideType  path      string                  true  "Slide type"  Enums(featured, bestseller, discounted)
// @Param        body       body      request.SlideItemRequest  true  "Slide item"
// @Success      201        {object}  dtoresponse.SlideItemResponse
// @Failure      400        {object}  dtoresponse.ErrorResponse
// @Failure      401        {object}  dtoresponse.ErrorResponse
// @Failure      403        {object}  dtoresponse.ErrorResponse
// @Failure      404        {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/product-slides/{slideType}/items [post]
func (h *StorefrontHandler) CreateSlideItem(w http.ResponseWriter, r *http.Request) {
	slideType := chi.URLParam(r, "slideType")

	var req request.SlideItemRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	productID, err := uuid.Parse(req.ProductID)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	item, err := h.content.CreateSlideItem(r.Context(), slideType, appstorecontent.CreateSlideItemInput{
		ProductID: productID,
		SortOrder: req.SortOrder,
		TabLabel:  req.TabLabel,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToSlideItemResponse(*item))
}

// UpdateSlideItem godoc
// @Summary      Update slide item
// @Description  Updates a product slide item sort order or tab label.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                  true  "Slide item ID"
// @Param        body  body      request.SlideItemRequest  true  "Slide item"
// @Success      200   {object}  dtoresponse.SlideItemResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/product-slide-items/{id} [put]
func (h *StorefrontHandler) UpdateSlideItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.SlideItemRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	item, err := h.content.UpdateSlideItem(r.Context(), id, appstorecontent.UpdateSlideItemInput{
		SortOrder: req.SortOrder,
		TabLabel:  req.TabLabel,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToSlideItemResponse(*item))
}

// DeleteSlideItem godoc
// @Summary      Delete slide item
// @Description  Removes a product from a carousel slide.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Slide item ID"
// @Success      204 "No Content"
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/product-slide-items/{id} [delete]
func (h *StorefrontHandler) DeleteSlideItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.content.DeleteSlideItem(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.NoContent(w)
}

// ListProBanners godoc
// @Summary      List pro banners
// @Description  Returns all promotional homepage banners.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.ProBannerListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/pro-banners [get]
func (h *StorefrontHandler) ListProBanners(w http.ResponseWriter, r *http.Request) {
	banners, err := h.content.ListProBanners(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToProBannerListResponse(banners))
}

// CreateProBanner godoc
// @Summary      Create pro banner
// @Description  Creates a promotional homepage banner.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.ProBannerRequest  true  "Banner data"
// @Success      201   {object}  dtoresponse.ProBannerResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/pro-banners [post]
func (h *StorefrontHandler) CreateProBanner(w http.ResponseWriter, r *http.Request) {
	var req request.ProBannerRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	banner, err := h.content.CreateProBanner(r.Context(), appstorecontent.CreateProBannerInput{
		DesktopImageURL: req.DesktopImageURL,
		MobileImageURL:  req.MobileImageURL,
		LinkURL:         req.LinkURL,
		SortOrder:       req.SortOrder,
		IsActive:        req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToProBannerResponse(banner))
}

// UpdateProBanner godoc
// @Summary      Update pro banner
// @Description  Updates a promotional homepage banner.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                  true  "Banner ID"
// @Param        body  body      request.ProBannerRequest  true  "Banner data"
// @Success      200   {object}  dtoresponse.ProBannerResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/pro-banners/{id} [put]
func (h *StorefrontHandler) UpdateProBanner(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.ProBannerRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	banner, err := h.content.UpdateProBanner(r.Context(), id, appstorecontent.UpdateProBannerInput{
		DesktopImageURL: &req.DesktopImageURL,
		MobileImageURL:  &req.MobileImageURL,
		LinkURL:         &req.LinkURL,
		SortOrder:       &req.SortOrder,
		IsActive:        &req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToProBannerResponse(banner))
}

// DeleteProBanner godoc
// @Summary      Delete pro banner
// @Description  Deletes a promotional homepage banner.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Banner ID"
// @Success      204 "No Content"
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/pro-banners/{id} [delete]
func (h *StorefrontHandler) DeleteProBanner(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.content.DeleteProBanner(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.NoContent(w)
}

// ListPartnerBrands godoc
// @Summary      List partner brands
// @Description  Returns all homepage partner brand blocks.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.PartnerBrandListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/partner-brands [get]
func (h *StorefrontHandler) ListPartnerBrands(w http.ResponseWriter, r *http.Request) {
	brands, err := h.content.ListPartnerBrands(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToPartnerBrandListResponse(brands))
}

// CreatePartnerBrand godoc
// @Summary      Create partner brand
// @Description  Creates a homepage partner brand block.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.PartnerBrandRequest  true  "Partner brand data"
// @Success      201   {object}  dtoresponse.PartnerBrandResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/partner-brands [post]
func (h *StorefrontHandler) CreatePartnerBrand(w http.ResponseWriter, r *http.Request) {
	var req request.PartnerBrandRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	brand, err := h.content.CreatePartnerBrand(r.Context(), appstorecontent.CreatePartnerBrandInput{
		Title:       req.Title,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		LinkURL:     req.LinkURL,
		SortOrder:   req.SortOrder,
		IsActive:    req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToPartnerBrandResponse(brand))
}

// UpdatePartnerBrand godoc
// @Summary      Update partner brand
// @Description  Updates a homepage partner brand block.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                       true  "Partner brand ID"
// @Param        body  body      request.PartnerBrandRequest  true  "Partner brand data"
// @Success      200   {object}  dtoresponse.PartnerBrandResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/partner-brands/{id} [put]
func (h *StorefrontHandler) UpdatePartnerBrand(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.PartnerBrandRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	brand, err := h.content.UpdatePartnerBrand(r.Context(), id, appstorecontent.UpdatePartnerBrandInput{
		Title:       &req.Title,
		Description: &req.Description,
		LogoURL:     &req.LogoURL,
		LinkURL:     &req.LinkURL,
		SortOrder:   &req.SortOrder,
		IsActive:    &req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToPartnerBrandResponse(brand))
}

// DeletePartnerBrand godoc
// @Summary      Delete partner brand
// @Description  Deletes a homepage partner brand block.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Partner brand ID"
// @Success      204 "No Content"
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/partner-brands/{id} [delete]
func (h *StorefrontHandler) DeletePartnerBrand(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.content.DeletePartnerBrand(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.NoContent(w)
}

// ListHomepageReviews godoc
// @Summary      List homepage reviews
// @Description  Returns all homepage customer testimonials.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.HomepageReviewListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/homepage-reviews [get]
func (h *StorefrontHandler) ListHomepageReviews(w http.ResponseWriter, r *http.Request) {
	reviews, err := h.content.ListHomepageReviews(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToHomepageReviewListResponse(reviews))
}

// CreateHomepageReview godoc
// @Summary      Create homepage review
// @Description  Creates a homepage customer testimonial.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.HomepageReviewRequest  true  "Review data"
// @Success      201   {object}  dtoresponse.HomepageReviewResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/homepage-reviews [post]
func (h *StorefrontHandler) CreateHomepageReview(w http.ResponseWriter, r *http.Request) {
	var req request.HomepageReviewRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	review, err := h.content.CreateHomepageReview(r.Context(), appstorecontent.CreateHomepageReviewInput{
		CustomerName: req.CustomerName,
		PhotoURL:     req.PhotoURL,
		ReviewText:   req.ReviewText,
		Rating:       req.Rating,
		SortOrder:    req.SortOrder,
		IsActive:     req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToHomepageReviewResponse(review))
}

// UpdateHomepageReview godoc
// @Summary      Update homepage review
// @Description  Updates a homepage customer testimonial.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                         true  "Review ID"
// @Param        body  body      request.HomepageReviewRequest  true  "Review data"
// @Success      200   {object}  dtoresponse.HomepageReviewResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/homepage-reviews/{id} [put]
func (h *StorefrontHandler) UpdateHomepageReview(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.HomepageReviewRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	review, err := h.content.UpdateHomepageReview(r.Context(), id, appstorecontent.UpdateHomepageReviewInput{
		CustomerName: &req.CustomerName,
		PhotoURL:     &req.PhotoURL,
		ReviewText:   &req.ReviewText,
		Rating:       req.Rating,
		SortOrder:    &req.SortOrder,
		IsActive:     &req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToHomepageReviewResponse(review))
}

// DeleteHomepageReview godoc
// @Summary      Delete homepage review
// @Description  Deletes a homepage customer testimonial.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Review ID"
// @Success      204 "No Content"
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/homepage-reviews/{id} [delete]
func (h *StorefrontHandler) DeleteHomepageReview(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.content.DeleteHomepageReview(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.NoContent(w)
}

// GetFAQSection godoc
// @Summary      Get FAQ section
// @Description  Returns the FAQ section image configuration.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.FAQSectionResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/faq [get]
func (h *StorefrontHandler) GetFAQSection(w http.ResponseWriter, r *http.Request) {
	section, err := h.content.GetFAQSection(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToFAQSectionResponse(section))
}

// UpdateFAQSection godoc
// @Summary      Update FAQ section
// @Description  Updates the FAQ section image.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.UpdateFAQSectionRequest  true  "FAQ section"
// @Success      200   {object}  dtoresponse.FAQSectionResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/faq [put]
func (h *StorefrontHandler) UpdateFAQSection(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateFAQSectionRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	section, err := h.content.UpdateFAQSection(r.Context(), appstorecontent.UpdateFAQSectionInput{
		ImageURL: req.ImageURL,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToFAQSectionResponse(section))
}

// ListFAQItems godoc
// @Summary      List FAQ items
// @Description  Returns all FAQ Q&A items.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.FAQItemListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/faq/items [get]
func (h *StorefrontHandler) ListFAQItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.content.ListFAQItems(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToFAQItemListResponse(items))
}

// CreateFAQItem godoc
// @Summary      Create FAQ item
// @Description  Creates an FAQ Q&A item.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.FAQItemRequest  true  "FAQ item"
// @Success      201   {object}  dtoresponse.FAQItemResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/faq/items [post]
func (h *StorefrontHandler) CreateFAQItem(w http.ResponseWriter, r *http.Request) {
	var req request.FAQItemRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	item, err := h.content.CreateFAQItem(r.Context(), appstorecontent.CreateFAQItemInput{
		Question:  req.Question,
		Answer:    req.Answer,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.Created(w, dtoresponse.ToFAQItemResponse(item))
}

// UpdateFAQItem godoc
// @Summary      Update FAQ item
// @Description  Updates an FAQ Q&A item.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      string                true  "FAQ item ID"
// @Param        body  body      request.FAQItemRequest  true  "FAQ item"
// @Success      200   {object}  dtoresponse.FAQItemResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Failure      404   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/faq/items/{id} [put]
func (h *StorefrontHandler) UpdateFAQItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.FAQItemRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	item, err := h.content.UpdateFAQItem(r.Context(), id, appstorecontent.UpdateFAQItemInput{
		Question:  &req.Question,
		Answer:    &req.Answer,
		SortOrder: &req.SortOrder,
		IsActive:  &req.IsActive,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToFAQItemResponse(item))
}

// DeleteFAQItem godoc
// @Summary      Delete FAQ item
// @Description  Deletes an FAQ Q&A item.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "FAQ item ID"
// @Success      204 "No Content"
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/faq/items/{id} [delete]
func (h *StorefrontHandler) DeleteFAQItem(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.content.DeleteFAQItem(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.NoContent(w)
}

// GetContactSection godoc
// @Summary      Get contact section
// @Description  Returns the homepage contact section image URL.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.ContactSectionResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/contact-section [get]
func (h *StorefrontHandler) GetContactSection(w http.ResponseWriter, r *http.Request) {
	imageURL, err := h.settings.GetContactSectionImage(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToContactSectionResponse(imageURL))
}

// UpdateContactSection godoc
// @Summary      Update contact section
// @Description  Updates the homepage contact section image URL.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.ContactSectionRequest  true  "Contact section"
// @Success      200   {object}  dtoresponse.ContactSectionResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/contact-section [put]
func (h *StorefrontHandler) UpdateContactSection(w http.ResponseWriter, r *http.Request) {
	var req request.ContactSectionRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	imageURL, err := h.settings.UpdateContactSectionImage(r.Context(), req.ImageURL)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToContactSectionResponse(imageURL))
}

// GetNavigation godoc
// @Summary      Get storefront navigation
// @Description  Returns the storefront header navigation menu tree.
// @Tags         storefront
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dtoresponse.NavigationResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/navigation [get]
func (h *StorefrontHandler) GetNavigation(w http.ResponseWriter, r *http.Request) {
	items, err := h.settings.GetStorefrontNavigation(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToNavigationResponse(items))
}

// UpdateNavigation godoc
// @Summary      Update storefront navigation
// @Description  Replaces the storefront header navigation menu tree.
// @Tags         storefront
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      request.UpdateNavigationRequest  true  "Navigation tree"
// @Success      200   {object}  dtoresponse.NavigationResponse
// @Failure      400   {object}  dtoresponse.ErrorResponse
// @Failure      401   {object}  dtoresponse.ErrorResponse
// @Failure      403   {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/storefront/navigation [put]
func (h *StorefrontHandler) UpdateNavigation(w http.ResponseWriter, r *http.Request) {
	var req request.UpdateNavigationRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	items, err := h.settings.UpdateStorefrontNavigation(r.Context(), toNavItems(req.Items))
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToNavigationResponse(items))
}
