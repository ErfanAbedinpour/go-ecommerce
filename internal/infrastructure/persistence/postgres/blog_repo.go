package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"app/internal/domain/blog"
	"app/internal/infrastructure/persistence/models"
	"app/pkg/pagination"
)

type BlogRepository struct {
	db *gorm.DB
}

func NewBlogRepository(db *gorm.DB) *BlogRepository {
	return &BlogRepository{db: db}
}

// category mapping helpers
func toBlogCategoryModel(c *blog.Category) *models.BlogCategoryModel {
	m := &models.BlogCategoryModel{
		ID:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
	}
	if c.Description != "" {
		m.Description = &c.Description
	}
	return m
}

func toBlogCategoryDomain(m *models.BlogCategoryModel) *blog.Category {
	c := &blog.Category{
		ID:        m.ID,
		Name:      m.Name,
		Slug:      m.Slug,
		CreatedAt: m.CreatedAt,
	}
	if m.Description != nil {
		c.Description = *m.Description
	}
	return c
}

// post mapping helpers
func toBlogPostModel(p *blog.Post) *models.BlogPostModel {
	m := &models.BlogPostModel{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		CategoryID:  p.CategoryID,
		AuthorID:    p.AuthorID,
		Status:      string(p.Status),
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
	if p.Summary != "" {
		m.Summary = &p.Summary
	}
	if p.FeaturedImage != "" {
		m.FeaturedImage = &p.FeaturedImage
	}
	return m
}

func toBlogPostDomain(m *models.BlogPostModel) *blog.Post {
	p := &blog.Post{
		ID:          m.ID,
		Title:       m.Title,
		Slug:        m.Slug,
		Content:     m.Content,
		CategoryID:  m.CategoryID,
		AuthorID:    m.AuthorID,
		Status:      blog.PostStatus(m.Status),
		PublishedAt: m.PublishedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
	if m.Summary != nil {
		p.Summary = *m.Summary
	}
	if m.FeaturedImage != nil {
		p.FeaturedImage = *m.FeaturedImage
	}
	return p
}

// comment mapping helpers
func toBlogCommentModel(c *blog.Comment) *models.BlogCommentModel {
	return &models.BlogCommentModel{
		ID:          c.ID,
		PostID:      c.PostID,
		AuthorName:  c.AuthorName,
		AuthorEmail: c.AuthorEmail,
		Content:     c.Content,
		Status:      string(c.Status),
		CreatedAt:   c.CreatedAt,
	}
}

func toBlogCommentDomain(m *models.BlogCommentModel) *blog.Comment {
	return &blog.Comment{
		ID:          m.ID,
		PostID:      m.PostID,
		AuthorName:  m.AuthorName,
		AuthorEmail: m.AuthorEmail,
		Content:     m.Content,
		Status:      blog.CommentStatus(m.Status),
		CreatedAt:   m.CreatedAt,
	}
}

// Category Repository Methods
func (r *BlogRepository) CreateCategory(ctx context.Context, cat *blog.Category) error {
	m := toBlogCategoryModel(cat)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") {
			return blog.ErrDuplicateSlug
		}
		return err
	}
	return nil
}

func (r *BlogRepository) FindCategoryByID(ctx context.Context, id uuid.UUID) (*blog.Category, error) {
	var m models.BlogCategoryModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, blog.ErrCategoryNotFound
		}
		return nil, err
	}
	return toBlogCategoryDomain(&m), nil
}

func (r *BlogRepository) FindCategoryBySlug(ctx context.Context, slug string) (*blog.Category, error) {
	var m models.BlogCategoryModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, blog.ErrCategoryNotFound
		}
		return nil, err
	}
	return toBlogCategoryDomain(&m), nil
}

func (r *BlogRepository) ListCategories(ctx context.Context) ([]blog.Category, error) {
	var modelsList []models.BlogCategoryModel
	err := r.db.WithContext(ctx).Order("name ASC").Find(&modelsList).Error
	if err != nil {
		return nil, err
	}
	res := make([]blog.Category, len(modelsList))
	for i, m := range modelsList {
		res[i] = *toBlogCategoryDomain(&m)
	}
	return res, nil
}

func (r *BlogRepository) UpdateCategory(ctx context.Context, cat *blog.Category) error {
	m := toBlogCategoryModel(cat)
	result := r.db.WithContext(ctx).Model(&models.BlogCategoryModel{}).Where("id = ?", cat.ID).Updates(m)
	if result.Error != nil {
		if strings.Contains(strings.ToLower(result.Error.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(result.Error.Error()), "unique") {
			return blog.ErrDuplicateSlug
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return blog.ErrCategoryNotFound
	}
	return nil
}

func (r *BlogRepository) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.BlogCategoryModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return blog.ErrCategoryNotFound
	}
	return nil
}

// Post Repository Methods
func (r *BlogRepository) CreatePost(ctx context.Context, post *blog.Post) error {
	m := toBlogPostModel(post)
	err := r.db.WithContext(ctx).Create(m).Error
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") {
			return blog.ErrDuplicateSlug
		}
		return err
	}
	return nil
}

func (r *BlogRepository) FindPostByID(ctx context.Context, id uuid.UUID) (*blog.Post, error) {
	var m models.BlogPostModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, blog.ErrPostNotFound
		}
		return nil, err
	}
	return toBlogPostDomain(&m), nil
}

func (r *BlogRepository) FindPostBySlug(ctx context.Context, slug string) (*blog.Post, error) {
	var m models.BlogPostModel
	err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, blog.ErrPostNotFound
		}
		return nil, err
	}
	return toBlogPostDomain(&m), nil
}

func (r *BlogRepository) ListPosts(ctx context.Context, filter blog.PostFilter, page pagination.Params) ([]blog.PostListItem, int64, error) {
	base := r.db.WithContext(ctx).Table("blog_posts")

	if filter.CategoryID != nil {
		base = base.Where("blog_posts.category_id = ?", *filter.CategoryID)
	}
	if filter.Status != "" {
		base = base.Where("blog_posts.status = ?", filter.Status)
	}
	if filter.Query != "" {
		pat := "%" + strings.ToLower(filter.Query) + "%"
		base = base.Where("LOWER(blog_posts.title) LIKE ? OR LOWER(blog_posts.content) LIKE ?", pat, pat)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		models.BlogPostModel
		CategoryName string `gorm:"column:category_name"`
		AuthorName   string `gorm:"column:author_name"`
	}

	var rows []row
	err := base.
		Select("blog_posts.*, blog_categories.name AS category_name, (admin_users.first_name || ' ' || admin_users.last_name) AS author_name").
		Joins("LEFT JOIN blog_categories ON blog_categories.id = blog_posts.category_id").
		Joins("LEFT JOIN admin_users ON admin_users.id = blog_posts.author_id").
		Order("blog_posts.created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit()).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	res := make([]blog.PostListItem, len(rows))
	for i, rw := range rows {
		res[i] = blog.PostListItem{
			Post:         *toBlogPostDomain(&rw.BlogPostModel),
			CategoryName: rw.CategoryName,
			AuthorName:   rw.AuthorName,
		}
	}

	return res, total, nil
}

func (r *BlogRepository) UpdatePost(ctx context.Context, post *blog.Post) error {
	m := toBlogPostModel(post)
	result := r.db.WithContext(ctx).Model(&models.BlogPostModel{}).Where("id = ?", post.ID).Updates(m)
	if result.Error != nil {
		if strings.Contains(strings.ToLower(result.Error.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(result.Error.Error()), "unique") {
			return blog.ErrDuplicateSlug
		}
		return result.Error
	}
	if result.RowsAffected == 0 {
		return blog.ErrPostNotFound
	}
	return nil
}

func (r *BlogRepository) DeletePost(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.BlogPostModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return blog.ErrPostNotFound
	}
	return nil
}

// Comment Repository Methods
func (r *BlogRepository) CreateComment(ctx context.Context, comment *blog.Comment) error {
	m := toBlogCommentModel(comment)
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *BlogRepository) FindCommentByID(ctx context.Context, id uuid.UUID) (*blog.Comment, error) {
	var m models.BlogCommentModel
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, blog.ErrCommentNotFound
		}
		return nil, err
	}
	return toBlogCommentDomain(&m), nil
}

func (r *BlogRepository) ListComments(ctx context.Context, filter blog.CommentFilter, page pagination.Params) ([]blog.Comment, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.BlogCommentModel{})

	if filter.PostID != nil {
		query = query.Where("post_id = ?", *filter.PostID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []models.BlogCommentModel
	err := query.Order("created_at DESC").Offset(page.Offset()).Limit(page.Limit()).Find(&items).Error
	if err != nil {
		return nil, 0, err
	}

	res := make([]blog.Comment, len(items))
	for i, m := range items {
		res[i] = *toBlogCommentDomain(&m)
	}
	return res, total, nil
}

func (r *BlogRepository) ListCommentsAdmin(ctx context.Context, filter blog.CommentFilter, page pagination.Params) ([]blog.CommentListItem, int64, error) {
	base := r.db.WithContext(ctx).Table("blog_comments")

	if filter.PostID != nil {
		base = base.Where("blog_comments.post_id = ?", *filter.PostID)
	}
	if filter.Status != "" {
		base = base.Where("blog_comments.status = ?", filter.Status)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	type row struct {
		models.BlogCommentModel
		PostTitle string `gorm:"column:post_title"`
	}

	var rows []row
	err := base.
		Select("blog_comments.*, blog_posts.title AS post_title").
		Joins("INNER JOIN blog_posts ON blog_posts.id = blog_comments.post_id").
		Order("blog_comments.created_at DESC").
		Offset(page.Offset()).
		Limit(page.Limit()).
		Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	res := make([]blog.CommentListItem, len(rows))
	for i, rw := range rows {
		res[i] = blog.CommentListItem{
			Comment:   *toBlogCommentDomain(&rw.BlogCommentModel),
			PostTitle: rw.PostTitle,
		}
	}
	return res, total, nil
}

func (r *BlogRepository) UpdateCommentStatus(ctx context.Context, id uuid.UUID, status blog.CommentStatus) error {
	result := r.db.WithContext(ctx).
		Model(&models.BlogCommentModel{}).
		Where("id = ?", id).
		Update("status", string(status))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return blog.ErrCommentNotFound
	}
	return nil
}

func (r *BlogRepository) DeleteComment(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.BlogCommentModel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return blog.ErrCommentNotFound
	}
	return nil
}

var _ blog.Repository = (*BlogRepository)(nil)
