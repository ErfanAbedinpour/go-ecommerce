package request

type BlogCategoryCreateRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Slug        string `json:"slug" validate:"required,max=255"`
	Description string `json:"description"`
}

type BlogCategoryUpdateRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Slug        string `json:"slug" validate:"required,max=255"`
	Description string `json:"description"`
}

type BlogPostCreateRequest struct {
	Title         string  `json:"title" validate:"required,max=255"`
	Slug          string  `json:"slug" validate:"required,max=255"`
	Content       string  `json:"content" validate:"required"`
	Summary       string  `json:"summary"`
	FeaturedImage string  `json:"featured_image" validate:"omitempty,max=500"`
	CategoryID    *string `json:"category_id" validate:"omitempty,uuid"`
	Status        string  `json:"status" validate:"omitempty,oneof=draft published"`
}

type BlogPostUpdateRequest struct {
	Title         string  `json:"title" validate:"required,max=255"`
	Slug          string  `json:"slug" validate:"required,max=255"`
	Content       string  `json:"content" validate:"required"`
	Summary       string  `json:"summary"`
	FeaturedImage string  `json:"featured_image" validate:"omitempty,max=500"`
	CategoryID    *string `json:"category_id" validate:"omitempty,uuid"`
	Status        string  `json:"status" validate:"omitempty,oneof=draft published"`
}

type BlogCommentSubmitRequest struct {
	AuthorName  string `json:"author_name" validate:"required,max=255"`
	AuthorEmail string `json:"author_email" validate:"required,email,max=255"`
	Content     string `json:"content" validate:"required"`
}

type BlogCommentModerateRequest struct {
	Status string `json:"status" validate:"required,oneof=pending approved rejected"`
}
