package models

import (
	"time"

	"github.com/google/uuid"
)

type BlogCategoryModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"type:varchar(255);not null"`
	Slug        string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	Description *string   `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (BlogCategoryModel) TableName() string { return "blog_categories" }

type BlogPostModel struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Title         string     `gorm:"type:varchar(255);not null"`
	Slug          string     `gorm:"type:varchar(255);not null;uniqueIndex"`
	Content       string     `gorm:"type:text;not null"`
	Summary       *string    `gorm:"type:text"`
	FeaturedImage *string    `gorm:"type:varchar(500)"`
	CategoryID    *uuid.UUID `gorm:"type:uuid"`
	AuthorID      *uuid.UUID `gorm:"type:uuid"`
	Status        string     `gorm:"type:varchar(50);not null;default:'draft'"`
	PublishedAt   *time.Time `gorm:"type:timestamptz"`
	CreatedAt     time.Time  `gorm:"type:timestamptz;not null;autoCreateTime"`
	UpdatedAt     time.Time  `gorm:"type:timestamptz;not null;autoUpdateTime"`
}

func (BlogPostModel) TableName() string { return "blog_posts" }

type BlogCommentModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PostID      uuid.UUID `gorm:"type:uuid;not null"`
	AuthorName  string    `gorm:"type:varchar(255);not null"`
	AuthorEmail string    `gorm:"type:varchar(255);not null"`
	Content     string    `gorm:"type:text;not null"`
	Status      string    `gorm:"type:varchar(50);not null;default:'pending'"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null;autoCreateTime"`
}

func (BlogCommentModel) TableName() string { return "blog_comments" }
