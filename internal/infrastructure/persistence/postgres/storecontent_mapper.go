package postgres

import (
	"app/internal/domain/storecontent"
	"app/internal/infrastructure/persistence/models"
)

func toHeroDomain(m *models.StorefrontHeroModel) *storecontent.Hero {
	h := &storecontent.Hero{
		ID:       m.ID,
		IsActive: m.IsActive,
		UpdatedAt: m.UpdatedAt,
	}
	if m.VideoURL != nil {
		h.VideoURL = *m.VideoURL
	}
	if m.Title != nil {
		h.Title = *m.Title
	}
	if m.Subtitle != nil {
		h.Subtitle = *m.Subtitle
	}
	if m.CTAPrimaryText != nil {
		h.CTAPrimaryText = *m.CTAPrimaryText
	}
	if m.CTAPrimaryURL != nil {
		h.CTAPrimaryURL = *m.CTAPrimaryURL
	}
	if m.CTASecondaryText != nil {
		h.CTASecondaryText = *m.CTASecondaryText
	}
	if m.CTASecondaryURL != nil {
		h.CTASecondaryURL = *m.CTASecondaryURL
	}
	return h
}

func toHeroModel(h *storecontent.Hero) *models.StorefrontHeroModel {
	m := &models.StorefrontHeroModel{
		ID:        h.ID,
		IsActive:  h.IsActive,
		UpdatedAt: h.UpdatedAt,
	}
	if h.VideoURL != "" {
		m.VideoURL = &h.VideoURL
	}
	if h.Title != "" {
		m.Title = &h.Title
	}
	if h.Subtitle != "" {
		m.Subtitle = &h.Subtitle
	}
	if h.CTAPrimaryText != "" {
		m.CTAPrimaryText = &h.CTAPrimaryText
	}
	if h.CTAPrimaryURL != "" {
		m.CTAPrimaryURL = &h.CTAPrimaryURL
	}
	if h.CTASecondaryText != "" {
		m.CTASecondaryText = &h.CTASecondaryText
	}
	if h.CTASecondaryURL != "" {
		m.CTASecondaryURL = &h.CTASecondaryURL
	}
	return m
}

func toProductSlideDomain(m *models.ProductSlideModel) *storecontent.ProductSlide {
	s := &storecontent.ProductSlide{
		ID:                 m.ID,
		SlideType:          storecontent.SlideType(m.SlideType),
		AutoplayIntervalMs: m.AutoplayIntervalMs,
		IsActive:           m.IsActive,
		SortOrder:          m.SortOrder,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
	if m.Title != nil {
		s.Title = *m.Title
	}
	for _, item := range m.Items {
		s.Items = append(s.Items, toSlideItemDomain(&item))
	}
	return s
}

func toProductSlidesDomain(items []models.ProductSlideModel) []storecontent.ProductSlide {
	result := make([]storecontent.ProductSlide, len(items))
	for i, m := range items {
		result[i] = *toProductSlideDomain(&m)
	}
	return result
}

func toSlideItemDomain(m *models.ProductSlideItemModel) storecontent.SlideItem {
	item := storecontent.SlideItem{
		ID:        m.ID,
		SlideID:   m.SlideID,
		ProductID: m.ProductID,
		SortOrder: m.SortOrder,
	}
	if m.TabLabel != nil {
		item.TabLabel = *m.TabLabel
	}
	return item
}

func toSlideItemModel(item *storecontent.SlideItem) *models.ProductSlideItemModel {
	m := &models.ProductSlideItemModel{
		ID:        item.ID,
		SlideID:   item.SlideID,
		ProductID: item.ProductID,
		SortOrder: item.SortOrder,
	}
	if item.TabLabel != "" {
		m.TabLabel = &item.TabLabel
	}
	return m
}

func toProBannerDomain(m *models.ProBannerModel) *storecontent.ProBanner {
	b := &storecontent.ProBanner{
		ID:              m.ID,
		DesktopImageURL: m.DesktopImageURL,
		SortOrder:       m.SortOrder,
		IsActive:        m.IsActive,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
	if m.MobileImageURL != nil {
		b.MobileImageURL = *m.MobileImageURL
	}
	if m.LinkURL != nil {
		b.LinkURL = *m.LinkURL
	}
	return b
}

func toProBannersDomain(items []models.ProBannerModel) []storecontent.ProBanner {
	result := make([]storecontent.ProBanner, len(items))
	for i, m := range items {
		result[i] = *toProBannerDomain(&m)
	}
	return result
}

func toProBannerModel(b *storecontent.ProBanner) *models.ProBannerModel {
	m := &models.ProBannerModel{
		ID:              b.ID,
		DesktopImageURL: b.DesktopImageURL,
		SortOrder:       b.SortOrder,
		IsActive:        b.IsActive,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
	if b.MobileImageURL != "" {
		m.MobileImageURL = &b.MobileImageURL
	}
	if b.LinkURL != "" {
		m.LinkURL = &b.LinkURL
	}
	return m
}

func toPartnerBrandDomain(m *models.PartnerBrandModel) *storecontent.PartnerBrand {
	b := &storecontent.PartnerBrand{
		ID:        m.ID,
		Title:     m.Title,
		LogoURL:   m.LogoURL,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	if m.Description != nil {
		b.Description = *m.Description
	}
	if m.LinkURL != nil {
		b.LinkURL = *m.LinkURL
	}
	return b
}

func toPartnerBrandsDomain(items []models.PartnerBrandModel) []storecontent.PartnerBrand {
	result := make([]storecontent.PartnerBrand, len(items))
	for i, m := range items {
		result[i] = *toPartnerBrandDomain(&m)
	}
	return result
}

func toPartnerBrandModel(b *storecontent.PartnerBrand) *models.PartnerBrandModel {
	m := &models.PartnerBrandModel{
		ID:        b.ID,
		Title:     b.Title,
		LogoURL:   b.LogoURL,
		SortOrder: b.SortOrder,
		IsActive:  b.IsActive,
		CreatedAt: b.CreatedAt,
		UpdatedAt: b.UpdatedAt,
	}
	if b.Description != "" {
		m.Description = &b.Description
	}
	if b.LinkURL != "" {
		m.LinkURL = &b.LinkURL
	}
	return m
}

func toHomepageReviewDomain(m *models.HomepageReviewModel) *storecontent.HomepageReview {
	r := &storecontent.HomepageReview{
		ID:           m.ID,
		CustomerName: m.CustomerName,
		ReviewText:   m.ReviewText,
		SortOrder:    m.SortOrder,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
	}
	if m.PhotoURL != nil {
		r.PhotoURL = *m.PhotoURL
	}
	if m.Rating != nil {
		v := int(*m.Rating)
		r.Rating = &v
	}
	return r
}

func toHomepageReviewsDomain(items []models.HomepageReviewModel) []storecontent.HomepageReview {
	result := make([]storecontent.HomepageReview, len(items))
	for i, m := range items {
		result[i] = *toHomepageReviewDomain(&m)
	}
	return result
}

func toHomepageReviewModel(r *storecontent.HomepageReview) *models.HomepageReviewModel {
	m := &models.HomepageReviewModel{
		ID:           r.ID,
		CustomerName: r.CustomerName,
		ReviewText:   r.ReviewText,
		SortOrder:    r.SortOrder,
		IsActive:     r.IsActive,
		CreatedAt:    r.CreatedAt,
	}
	if r.PhotoURL != "" {
		m.PhotoURL = &r.PhotoURL
	}
	if r.Rating != nil {
		v := int16(*r.Rating)
		m.Rating = &v
	}
	return m
}

func toFAQSectionDomain(m *models.FAQSectionModel) *storecontent.FAQSection {
	s := &storecontent.FAQSection{
		ID:        m.ID,
		UpdatedAt: m.UpdatedAt,
	}
	if m.ImageURL != nil {
		s.ImageURL = *m.ImageURL
	}
	return s
}

func toFAQSectionModel(s *storecontent.FAQSection) *models.FAQSectionModel {
	m := &models.FAQSectionModel{
		ID:        s.ID,
		UpdatedAt: s.UpdatedAt,
	}
	if s.ImageURL != "" {
		m.ImageURL = &s.ImageURL
	}
	return m
}

func toFAQItemDomain(m *models.FAQItemModel) *storecontent.FAQItem {
	return &storecontent.FAQItem{
		ID:        m.ID,
		Question:  m.Question,
		Answer:    m.Answer,
		SortOrder: m.SortOrder,
		IsActive:  m.IsActive,
	}
}

func toFAQItemsDomain(items []models.FAQItemModel) []storecontent.FAQItem {
	result := make([]storecontent.FAQItem, len(items))
	for i, m := range items {
		result[i] = *toFAQItemDomain(&m)
	}
	return result
}

func toFAQItemModel(item *storecontent.FAQItem) *models.FAQItemModel {
	return &models.FAQItemModel{
		ID:        item.ID,
		Question:  item.Question,
		Answer:    item.Answer,
		SortOrder: item.SortOrder,
		IsActive:  item.IsActive,
	}
}
