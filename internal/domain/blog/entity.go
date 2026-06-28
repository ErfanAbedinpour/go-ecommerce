package blog

import (
	"time"

	"github.com/google/uuid"
)

type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
)

type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"
	CommentStatusApproved CommentStatus = "approved"
	CommentStatusRejected CommentStatus = "rejected"
)

type Category struct {
	ID          uuid.UUID
	Name        string
	Slug        string
	Description string
	CreatedAt   time.Time
}

type Post struct {
	ID            uuid.UUID
	Title         string
	Slug          string
	Content       string
	Summary       string
	FeaturedImage string
	CategoryID    *uuid.UUID
	AuthorID      *uuid.UUID
	Status        PostStatus
	PublishedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Comment struct {
	ID          uuid.UUID
	PostID      uuid.UUID
	AuthorName  string
	AuthorEmail string
	Content     string
	Status      CommentStatus
	CreatedAt   time.Time
}

// PostListItem is a detailed post item with category and author metadata.
type PostListItem struct {
	Post
	CategoryName string
	AuthorName   string
}

// CommentListItem is a detailed comment item with post title metadata.
type CommentListItem struct {
	Comment
	PostTitle string
}
