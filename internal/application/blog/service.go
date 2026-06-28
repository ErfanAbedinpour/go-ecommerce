package blog

import (
	"context"
	"time"

	"github.com/google/uuid"

	domain "app/internal/domain/blog"
	"app/pkg/pagination"
)

type Service struct {
	repo domain.Repository
}

func NewService(repo domain.Repository) *Service {
	return &Service{repo: repo}
}

// ── Category Methods ──────────────────────────────────────────────

type CreateCategoryInput struct {
	Name        string
	Slug        string
	Description string
}

func (s *Service) CreateCategory(ctx context.Context, input CreateCategoryInput) (*domain.Category, error) {
	cat := &domain.Category{
		ID:          uuid.New(),
		Name:        input.Name,
		Slug:        input.Slug,
		Description: input.Description,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.CreateCategory(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

type UpdateCategoryInput struct {
	Name        string
	Slug        string
	Description string
}

func (s *Service) UpdateCategory(ctx context.Context, id uuid.UUID, input UpdateCategoryInput) (*domain.Category, error) {
	cat, err := s.repo.FindCategoryByID(ctx, id)
	if err != nil {
		return nil, err
	}
	cat.Name = input.Name
	cat.Slug = input.Slug
	cat.Description = input.Description
	if err := s.repo.UpdateCategory(ctx, cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *Service) GetCategory(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	return s.repo.FindCategoryByID(ctx, id)
}

func (s *Service) ListCategories(ctx context.Context) ([]domain.Category, error) {
	return s.repo.ListCategories(ctx)
}

func (s *Service) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteCategory(ctx, id)
}

// ── Post Methods ──────────────────────────────────────────────────

type CreatePostInput struct {
	Title         string
	Slug          string
	Content       string
	Summary       string
	FeaturedImage string
	CategoryID    *uuid.UUID
	AuthorID      uuid.UUID
	Status        string
}

func (s *Service) CreatePost(ctx context.Context, input CreatePostInput) (*domain.Post, error) {
	status := domain.PostStatusDraft
	if input.Status == string(domain.PostStatusPublished) {
		status = domain.PostStatusPublished
	}

	var pubAt *time.Time
	if status == domain.PostStatusPublished {
		now := time.Now().UTC()
		pubAt = &now
	}

	if input.CategoryID != nil {
		if _, err := s.repo.FindCategoryByID(ctx, *input.CategoryID); err != nil {
			return nil, err
		}
	}

	post := &domain.Post{
		ID:            uuid.New(),
		Title:         input.Title,
		Slug:          input.Slug,
		Content:       input.Content,
		Summary:       input.Summary,
		FeaturedImage: input.FeaturedImage,
		CategoryID:    input.CategoryID,
		AuthorID:      &input.AuthorID,
		Status:        status,
		PublishedAt:   pubAt,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	if err := s.repo.CreatePost(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

type UpdatePostInput struct {
	Title         string
	Slug          string
	Content       string
	Summary       string
	FeaturedImage string
	CategoryID    *uuid.UUID
	Status        string
}

func (s *Service) UpdatePost(ctx context.Context, id uuid.UUID, input UpdatePostInput) (*domain.Post, error) {
	post, err := s.repo.FindPostByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if input.CategoryID != nil {
		if _, err := s.repo.FindCategoryByID(ctx, *input.CategoryID); err != nil {
			return nil, err
		}
	}

	status := domain.PostStatusDraft
	if input.Status == string(domain.PostStatusPublished) {
		status = domain.PostStatusPublished
	}

	var pubAt = post.PublishedAt
	if status == domain.PostStatusPublished && post.Status != domain.PostStatusPublished {
		now := time.Now().UTC()
		pubAt = &now
	} else if status == domain.PostStatusDraft {
		pubAt = nil
	}

	post.Title = input.Title
	post.Slug = input.Slug
	post.Content = input.Content
	post.Summary = input.Summary
	post.FeaturedImage = input.FeaturedImage
	post.CategoryID = input.CategoryID
	post.Status = status
	post.PublishedAt = pubAt
	post.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdatePost(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *Service) GetPostByID(ctx context.Context, id uuid.UUID) (*domain.Post, error) {
	return s.repo.FindPostByID(ctx, id)
}

func (s *Service) GetPostBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	return s.repo.FindPostBySlug(ctx, slug)
}

func (s *Service) ListPosts(ctx context.Context, filter domain.PostFilter, page pagination.Params) (pagination.Paginated[domain.PostListItem], error) {
	items, total, err := s.repo.ListPosts(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.PostListItem]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

func (s *Service) DeletePost(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeletePost(ctx, id)
}

// ── Comment Methods ───────────────────────────────────────────────

type SubmitCommentInput struct {
	PostID      uuid.UUID
	AuthorName  string
	AuthorEmail string
	Content     string
}

func (s *Service) SubmitComment(ctx context.Context, input SubmitCommentInput) (*domain.Comment, error) {
	if _, err := s.repo.FindPostByID(ctx, input.PostID); err != nil {
		return nil, err
	}

	comment := &domain.Comment{
		ID:          uuid.New(),
		PostID:      input.PostID,
		AuthorName:  input.AuthorName,
		AuthorEmail: input.AuthorEmail,
		Content:     input.Content,
		Status:      domain.CommentStatusPending,
		CreatedAt:   time.Now().UTC(),
	}

	if err := s.repo.CreateComment(ctx, comment); err != nil {
		return nil, err
	}
	return comment, nil
}

func (s *Service) ListComments(ctx context.Context, postID uuid.UUID, page pagination.Params) (pagination.Paginated[domain.Comment], error) {
	filter := domain.CommentFilter{
		PostID: &postID,
		Status: string(domain.CommentStatusApproved),
	}
	items, total, err := s.repo.ListComments(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.Comment]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

func (s *Service) ListCommentsAdmin(ctx context.Context, filter domain.CommentFilter, page pagination.Params) (pagination.Paginated[domain.CommentListItem], error) {
	items, total, err := s.repo.ListCommentsAdmin(ctx, filter, page)
	if err != nil {
		return pagination.Paginated[domain.CommentListItem]{}, err
	}
	return pagination.NewPaginated(items, page.Page, page.PerPage, total), nil
}

func (s *Service) ModerateComment(ctx context.Context, id uuid.UUID, status string) error {
	var stat domain.CommentStatus
	switch status {
	case string(domain.CommentStatusApproved):
		stat = domain.CommentStatusApproved
	case string(domain.CommentStatusRejected):
		stat = domain.CommentStatusRejected
	case string(domain.CommentStatusPending):
		stat = domain.CommentStatusPending
	default:
		return domain.ErrCommentNotFound
	}
	return s.repo.UpdateCommentStatus(ctx, id, stat)
}

func (s *Service) DeleteComment(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteComment(ctx, id)
}
