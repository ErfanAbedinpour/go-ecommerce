package blog

import (
	"context"

	"github.com/google/uuid"

	"app/pkg/pagination"
)

type PostFilter struct {
	CategoryID *uuid.UUID
	Status     string
	Query      string
}

type CommentFilter struct {
	PostID *uuid.UUID
	Status string
}

type Repository interface {
	// Post operations
	CreatePost(ctx context.Context, post *Post) error
	FindPostByID(ctx context.Context, id uuid.UUID) (*Post, error)
	FindPostBySlug(ctx context.Context, slug string) (*Post, error)
	ListPosts(ctx context.Context, filter PostFilter, page pagination.Params) ([]PostListItem, int64, error)
	UpdatePost(ctx context.Context, post *Post) error
	DeletePost(ctx context.Context, id uuid.UUID) error

	// Category operations
	CreateCategory(ctx context.Context, cat *Category) error
	FindCategoryByID(ctx context.Context, id uuid.UUID) (*Category, error)
	FindCategoryBySlug(ctx context.Context, slug string) (*Category, error)
	ListCategories(ctx context.Context) ([]Category, error)
	UpdateCategory(ctx context.Context, cat *Category) error
	DeleteCategory(ctx context.Context, id uuid.UUID) error

	// Comment operations
	CreateComment(ctx context.Context, comment *Comment) error
	FindCommentByID(ctx context.Context, id uuid.UUID) (*Comment, error)
	ListComments(ctx context.Context, filter CommentFilter, page pagination.Params) ([]Comment, int64, error)
	ListCommentsAdmin(ctx context.Context, filter CommentFilter, page pagination.Params) ([]CommentListItem, int64, error)
	UpdateCommentStatus(ctx context.Context, id uuid.UUID, status CommentStatus) error
	DeleteComment(ctx context.Context, id uuid.UUID) error
}
