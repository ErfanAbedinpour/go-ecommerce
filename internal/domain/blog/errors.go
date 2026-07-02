package blog

import (
	"app/pkg/apperror"
	"app/pkg/i18n"
)

var (
	ErrPostNotFound     = apperror.NotFound("blog post")
	ErrCategoryNotFound = apperror.NotFound("blog category")
	ErrCommentNotFound  = apperror.NotFound("blog comment")
	ErrDuplicateSlug    = apperror.ConflictKeyed(i18n.KeyBlogDuplicateSlug, "slug is already in use")
)
