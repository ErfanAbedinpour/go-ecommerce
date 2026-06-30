package response

import (
	"time"

	domain "app/internal/domain/blog"
	"app/pkg/pagination"
)

type BlogCategoryResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToBlogCategoryResponse(c domain.Category) BlogCategoryResponse {
	return BlogCategoryResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
	}
}

func ToBlogCategoriesResponse(list []domain.Category) []BlogCategoryResponse {
	res := make([]BlogCategoryResponse, len(list))
	for i, c := range list {
		res[i] = ToBlogCategoryResponse(c)
	}
	return res
}

type BlogPostResponse struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Slug          string     `json:"slug"`
	Content       string     `json:"content"`
	Summary       string     `json:"summary,omitempty"`
	Excerpt       string     `json:"excerpt,omitempty"`
	FeaturedImage string     `json:"featured_image,omitempty"`
	CoverImageURL string     `json:"cover_image_url,omitempty"`
	CategoryID    *string    `json:"category_id,omitempty"`
	AuthorID      *string    `json:"author_id,omitempty"`
	Status        string     `json:"status"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func ToBlogPostResponse(p domain.Post) BlogPostResponse {
	var catID, authID *string
	if p.CategoryID != nil {
		cStr := p.CategoryID.String()
		catID = &cStr
	}
	if p.AuthorID != nil {
		aStr := p.AuthorID.String()
		authID = &aStr
	}
	return BlogPostResponse{
		ID:            p.ID.String(),
		Title:         p.Title,
		Slug:          p.Slug,
		Content:       p.Content,
		Summary:       p.Summary,
		Excerpt:       p.Summary,
		FeaturedImage: p.FeaturedImage,
		CoverImageURL: p.FeaturedImage,
		CategoryID:    catID,
		AuthorID:      authID,
		Status:        string(p.Status),
		PublishedAt:   p.PublishedAt,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

type BlogPostListItemResponse struct {
	BlogPostResponse
	CategoryName string `json:"category_name,omitempty"`
	AuthorName   string `json:"author_name,omitempty"`
}

func ToBlogPostListItemResponse(item domain.PostListItem) BlogPostListItemResponse {
	return BlogPostListItemResponse{
		BlogPostResponse: ToBlogPostResponse(item.Post),
		CategoryName:     item.CategoryName,
		AuthorName:       item.AuthorName,
	}
}

type BlogPostListResponse struct {
	Data []BlogPostListItemResponse `json:"data"`
	Meta pagination.Meta            `json:"meta"`
}

func ToBlogPostListResponse(paginated pagination.Paginated[domain.PostListItem]) BlogPostListResponse {
	data := make([]BlogPostListItemResponse, len(paginated.Data))
	for i, item := range paginated.Data {
		data[i] = ToBlogPostListItemResponse(item)
	}
	return BlogPostListResponse{
		Data: data,
		Meta: paginated.Meta,
	}
}

type BlogCommentResponse struct {
	ID          string    `json:"id"`
	PostID      string    `json:"post_id"`
	AuthorName  string    `json:"author_name"`
	AuthorEmail string    `json:"author_email"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func ToBlogCommentResponse(c domain.Comment) BlogCommentResponse {
	return BlogCommentResponse{
		ID:          c.ID.String(),
		PostID:      c.PostID.String(),
		AuthorName:  c.AuthorName,
		AuthorEmail: c.AuthorEmail,
		Content:     c.Content,
		Status:      string(c.Status),
		CreatedAt:   c.CreatedAt,
	}
}

type BlogCommentListResponse struct {
	Data []BlogCommentResponse `json:"data"`
	Meta pagination.Meta       `json:"meta"`
}

func ToBlogCommentListResponse(paginated pagination.Paginated[domain.Comment]) BlogCommentListResponse {
	data := make([]BlogCommentResponse, len(paginated.Data))
	for i, item := range paginated.Data {
		data[i] = ToBlogCommentResponse(item)
	}
	return BlogCommentListResponse{
		Data: data,
		Meta: paginated.Meta,
	}
}

type BlogCommentListItemResponse struct {
	BlogCommentResponse
	PostTitle string `json:"post_title"`
}

func ToBlogCommentListItemResponse(item domain.CommentListItem) BlogCommentListItemResponse {
	return BlogCommentListItemResponse{
		BlogCommentResponse: ToBlogCommentResponse(item.Comment),
		PostTitle:           item.PostTitle,
	}
}

type AdminBlogCommentListResponse struct {
	Data []BlogCommentListItemResponse `json:"data"`
	Meta pagination.Meta               `json:"meta"`
}

func ToAdminBlogCommentListResponse(paginated pagination.Paginated[domain.CommentListItem]) AdminBlogCommentListResponse {
	data := make([]BlogCommentListItemResponse, len(paginated.Data))
	for i, item := range paginated.Data {
		data[i] = ToBlogCommentListItemResponse(item)
	}
	return AdminBlogCommentListResponse{
		Data: data,
		Meta: paginated.Meta,
	}
}
