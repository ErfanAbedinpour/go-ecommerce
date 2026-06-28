package storecontent

import "app/pkg/apperror"

var (
	ErrNotFound         = apperror.NotFound("store content")
	ErrHeroNotFound     = apperror.NotFound("storefront hero")
	ErrSlideNotFound    = apperror.NotFound("product slide")
	ErrBannerNotFound   = apperror.NotFound("pro banner")
	ErrBrandNotFound    = apperror.NotFound("partner brand")
	ErrReviewNotFound   = apperror.NotFound("homepage review")
	ErrFAQItemNotFound  = apperror.NotFound("faq item")
	ErrSlideItemNotFound = apperror.NotFound("slide item")
	ErrInvalidSlideType = apperror.Validation("invalid slide type", map[string]string{
		"slide_type": "must be one of: featured, bestseller, discounted",
	})
	ErrDuplicateSlideItem = apperror.Conflict("product already exists in this slide")
)
