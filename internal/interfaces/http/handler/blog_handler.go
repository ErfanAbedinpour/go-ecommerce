package handler

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/go-chi/chi/v5"

	appblog "app/internal/application/blog"
	domainblog "app/internal/domain/blog"
	"app/internal/interfaces/http/dto/request"
	dtoresponse "app/internal/interfaces/http/dto/response"
	appmiddleware "app/internal/interfaces/http/middleware"
	"app/internal/interfaces/http/response"
	"app/pkg/pagination"
	"app/pkg/validator"
)

type BlogHandler struct {
	service   *appblog.Service
	validator *validator.Validator
	log       *slog.Logger
}

func NewBlogHandler(service *appblog.Service, v *validator.Validator, log *slog.Logger) *BlogHandler {
	return &BlogHandler{
		service:   service,
		validator: v,
		log:       log,
	}
}

// ── Storefront Handlers ───────────────────────────────────────────

// StoreListPosts godoc
// @Summary      List blog posts
// @Description  Get a paginated list of published blog posts.
// @Tags         blog
// @Produce      json
// @Param        page         query  int     false  "Page number"     default(1)
// @Param        per_page     query  int     false  "Items per page"  default(20)
// @Param        q            query  string  false  "Search query"
// @Param        category_id  query  string  false  "Category ID filter"
// @Success      200  {object}  dtoresponse.BlogPostListResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/blog/posts [get]
func (h *BlogHandler) StoreListPosts(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := domainblog.PostFilter{
		Status: string(domainblog.PostStatusPublished),
		Query:  r.URL.Query().Get("q"),
	}

	if catID := r.URL.Query().Get("category_id"); catID != "" {
		if id, err := uuid.Parse(catID); err == nil {
			filter.CategoryID = &id
		}
	}

	result, err := h.service.ListPosts(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToBlogPostListResponse(result))
}

// StoreGetPost godoc
// @Summary      Get blog post
// @Description  Get details of a single published blog post by slug.
// @Tags         blog
// @Produce      json
// @Param        slug  path  string  true  "Post slug"
// @Success      200  {object}  dtoresponse.BlogPostResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/blog/posts/{slug} [get]
func (h *BlogHandler) StoreGetPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	post, err := h.service.GetPostBySlug(r.Context(), slug)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if post.Status != domainblog.PostStatusPublished {
		response.Error(w, r, h.log, domainblog.ErrPostNotFound)
		return
	}

	response.OK(w, dtoresponse.ToBlogPostResponse(*post))
}

// StoreListCategories godoc
// @Summary      List blog categories
// @Description  Get all blog categories.
// @Tags         blog
// @Produce      json
// @Success      200  {array}   dtoresponse.BlogCategoryResponse
// @Router       /api/v1/store/blog/categories [get]
func (h *BlogHandler) StoreListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListCategories(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToBlogCategoriesResponse(items))
}

// StoreSubmitComment godoc
// @Summary      Submit a comment
// @Description  Submit a comment on a blog post (needs moderation).
// @Tags         blog
// @Accept       json
// @Produce      json
// @Param        postId  path  string                             true  "Post ID"
// @Param        body    body  request.BlogCommentSubmitRequest  true  "Comment data"
// @Success      201  {object}  dtoresponse.BlogCommentResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/blog/posts/{postId}/comments [post]
func (h *BlogHandler) StoreSubmitComment(w http.ResponseWriter, r *http.Request) {
	postID, err := parseUUIDParam(r, "postId")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.BlogCommentSubmitRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	comment, err := h.service.SubmitComment(r.Context(), appblog.SubmitCommentInput{
		PostID:      postID,
		AuthorName:  req.AuthorName,
		AuthorEmail: req.AuthorEmail,
		Content:     req.Content,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.ToBlogCommentResponse(*comment))
}

// StoreListComments godoc
// @Summary      List post comments
// @Description  Get a paginated list of approved comments for a blog post.
// @Tags         blog
// @Produce      json
// @Param        postId    path   string  true  "Post ID"
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Success      200  {object}  dtoresponse.BlogCommentListResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/blog/posts/{postId}/comments [get]
func (h *BlogHandler) StoreListComments(w http.ResponseWriter, r *http.Request) {
	postID, err := parseUUIDParam(r, "postId")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	page := pagination.FromRequest(r)
	result, err := h.service.ListComments(r.Context(), postID, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToBlogCommentListResponse(result))
}

// StoreListPostsAlias godoc
// @Summary      List blog posts (alias)
// @Description  Alias for GET /store/blog/posts for frontend compatibility.
// @Tags         blog
// @Produce      json
// @Param        page         query  int     false  "Page number"     default(1)
// @Param        per_page     query  int     false  "Items per page"  default(20)
// @Param        q            query  string  false  "Search query"
// @Param        category_id  query  string  false  "Category ID filter"
// @Success      200  {object}  dtoresponse.BlogPostListResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/blog [get]
func (h *BlogHandler) StoreListPostsAlias(w http.ResponseWriter, r *http.Request) {
	h.StoreListPosts(w, r)
}

// StoreListCommentsBySlug godoc
// @Summary      List post comments by slug
// @Description  Get approved comments for a published blog post resolved by slug.
// @Tags         blog
// @Produce      json
// @Param        slug      path   string  true  "Post slug"
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Success      200  {object}  dtoresponse.BlogCommentListResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/store/blog/{slug}/comments [get]
func (h *BlogHandler) StoreListCommentsBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	post, err := h.service.GetPostBySlug(r.Context(), slug)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	if post.Status != domainblog.PostStatusPublished {
		response.Error(w, r, h.log, domainblog.ErrPostNotFound)
		return
	}

	page := pagination.FromRequest(r)
	result, err := h.service.ListComments(r.Context(), post.ID, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToBlogCommentListResponse(result))
}

// ── Admin Categories Handlers ─────────────────────────────────────

// AdminListCategories godoc
// @Summary      List categories (Admin)
// @Description  Get list of all blog categories for admin panel.
// @Tags         admin-blog
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}   dtoresponse.BlogCategoryResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/categories [get]
func (h *BlogHandler) AdminListCategories(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListCategories(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}
	response.OK(w, dtoresponse.ToBlogCategoriesResponse(items))
}

// AdminCreateCategory godoc
// @Summary      Create category (Admin)
// @Description  Create a new blog category.
// @Tags         admin-blog
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  request.BlogCategoryCreateRequest  true  "Category data"
// @Success      201  {object}  dtoresponse.BlogCategoryResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      409  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/categories [post]
func (h *BlogHandler) AdminCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req request.BlogCategoryCreateRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	cat, err := h.service.CreateCategory(r.Context(), appblog.CreateCategoryInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.ToBlogCategoryResponse(*cat))
}

// AdminGetCategory godoc
// @Summary      Get category (Admin)
// @Description  Get details of a single blog category by ID.
// @Tags         admin-blog
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Category ID"
// @Success      200  {object}  dtoresponse.BlogCategoryResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/categories/{id} [get]
func (h *BlogHandler) AdminGetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	cat, err := h.service.GetCategory(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToBlogCategoryResponse(*cat))
}

// AdminUpdateCategory godoc
// @Summary      Update category (Admin)
// @Description  Update details of an existing blog category.
// @Tags         admin-blog
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                             true  "Category ID"
// @Param        body  body  request.BlogCategoryUpdateRequest  true  "Category data"
// @Success      200  {object}  dtoresponse.BlogCategoryResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Failure      409  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/categories/{id} [put]
func (h *BlogHandler) AdminUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.BlogCategoryUpdateRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	cat, err := h.service.UpdateCategory(r.Context(), id, appblog.UpdateCategoryInput{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToBlogCategoryResponse(*cat))
}

// AdminDeleteCategory godoc
// @Summary      Delete category (Admin)
// @Description  Delete a blog category by ID.
// @Tags         admin-blog
// @Security     BearerAuth
// @Param        id  path  string  true  "Category ID"
// @Success      204
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/categories/{id} [delete]
func (h *BlogHandler) AdminDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.service.DeleteCategory(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.NoContent(w)
}

// ── Admin Posts Handlers ──────────────────────────────────────────

// AdminListPosts godoc
// @Summary      List posts (Admin)
// @Description  Get a paginated list of all posts (drafts and published).
// @Tags         admin-blog
// @Produce      json
// @Security     BearerAuth
// @Param        page         query  int     false  "Page number"     default(1)
// @Param        per_page     query  int     false  "Items per page"  default(20)
// @Param        q            query  string  false  "Search query"
// @Param        category_id  query  string  false  "Category ID filter"
// @Param        status       query  string  false  "Status filter (draft/published)"
// @Success      200  {object}  dtoresponse.BlogPostListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/posts [get]
func (h *BlogHandler) AdminListPosts(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := domainblog.PostFilter{
		Status: r.URL.Query().Get("status"),
		Query:  r.URL.Query().Get("q"),
	}

	if catID := r.URL.Query().Get("category_id"); catID != "" {
		if id, err := uuid.Parse(catID); err == nil {
			filter.CategoryID = &id
		}
	}

	result, err := h.service.ListPosts(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToBlogPostListResponse(result))
}

// AdminCreatePost godoc
// @Summary      Create post (Admin)
// @Description  Create a new blog post.
// @Tags         admin-blog
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body  request.BlogPostCreateRequest  true  "Post data"
// @Success      201  {object}  dtoresponse.BlogPostResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      409  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/posts [post]
func (h *BlogHandler) AdminCreatePost(w http.ResponseWriter, r *http.Request) {
	var req request.BlogPostCreateRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	authorID, err := appmiddleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var catID *uuid.UUID
	if req.CategoryID != nil {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			response.Error(w, r, h.log, err)
			return
		}
		catID = &id
	}

	post, err := h.service.CreatePost(r.Context(), appblog.CreatePostInput{
		Title:         req.Title,
		Slug:          req.Slug,
		Content:       req.Content,
		Summary:       req.Summary,
		FeaturedImage: req.FeaturedImage,
		CategoryID:    catID,
		AuthorID:      authorID,
		Status:        req.Status,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.Created(w, dtoresponse.ToBlogPostResponse(*post))
}

// AdminGetPost godoc
// @Summary      Get post (Admin)
// @Description  Get details of a single post by ID.
// @Tags         admin-blog
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  string  true  "Post ID"
// @Success      200  {object}  dtoresponse.BlogPostResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/posts/{id} [get]
func (h *BlogHandler) AdminGetPost(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	post, err := h.service.GetPostByID(r.Context(), id)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToBlogPostResponse(*post))
}

// AdminUpdatePost godoc
// @Summary      Update post (Admin)
// @Description  Update details of an existing blog post.
// @Tags         admin-blog
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                         true  "Post ID"
// @Param        body  body  request.BlogPostUpdateRequest  true  "Post data"
// @Success      200  {object}  dtoresponse.BlogPostResponse
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Failure      409  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/posts/{id} [put]
func (h *BlogHandler) AdminUpdatePost(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.BlogPostUpdateRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var catID *uuid.UUID
	if req.CategoryID != nil {
		id, err := uuid.Parse(*req.CategoryID)
		if err != nil {
			response.Error(w, r, h.log, err)
			return
		}
		catID = &id
	}

	post, err := h.service.UpdatePost(r.Context(), id, appblog.UpdatePostInput{
		Title:         req.Title,
		Slug:          req.Slug,
		Content:       req.Content,
		Summary:       req.Summary,
		FeaturedImage: req.FeaturedImage,
		CategoryID:    catID,
		Status:        req.Status,
	})
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToBlogPostResponse(*post))
}

// AdminDeletePost godoc
// @Summary      Delete post (Admin)
// @Description  Delete a blog post by ID.
// @Tags         admin-blog
// @Security     BearerAuth
// @Param        id  path  string  true  "Post ID"
// @Success      204
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/posts/{id} [delete]
func (h *BlogHandler) AdminDeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.service.DeletePost(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.NoContent(w)
}

// ── Admin Comments Handlers ───────────────────────────────────────

// AdminListComments godoc
// @Summary      List comments (Admin)
// @Description  Get a paginated list of all blog comments.
// @Tags         admin-blog
// @Produce      json
// @Security     BearerAuth
// @Param        page      query  int     false  "Page number"     default(1)
// @Param        per_page  query  int     false  "Items per page"  default(20)
// @Param        post_id   query  string  false  "Post ID filter"
// @Param        status    query  string  false  "Status filter (pending/approved/rejected)"
// @Success      200  {object}  dtoresponse.AdminBlogCommentListResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/comments [get]
func (h *BlogHandler) AdminListComments(w http.ResponseWriter, r *http.Request) {
	page := pagination.FromRequest(r)
	filter := domainblog.CommentFilter{
		Status: r.URL.Query().Get("status"),
	}

	if postID := r.URL.Query().Get("post_id"); postID != "" {
		if id, err := uuid.Parse(postID); err == nil {
			filter.PostID = &id
		}
	}

	result, err := h.service.ListCommentsAdmin(r.Context(), filter, page)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.OK(w, dtoresponse.ToAdminBlogCommentListResponse(result))
}

// AdminModerateComment godoc
// @Summary      Moderate comment (Admin)
// @Description  Approve or reject a blog comment by ID.
// @Tags         admin-blog
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path  string                              true  "Comment ID"
// @Param        body  body  request.BlogCommentModerateRequest  true  "Moderation data"
// @Success      204
// @Failure      400  {object}  dtoresponse.ErrorResponse
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/comments/{id}/status [patch]
func (h *BlogHandler) AdminModerateComment(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	var req request.BlogCommentModerateRequest
	if err := decodeAndValidate(r, &req, h.validator); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	err = h.service.ModerateComment(r.Context(), id, req.Status)
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.NoContent(w)
}

// AdminDeleteComment godoc
// @Summary      Delete comment (Admin)
// @Description  Delete a blog comment by ID.
// @Tags         admin-blog
// @Security     BearerAuth
// @Param        id  path  string  true  "Comment ID"
// @Success      204
// @Failure      401  {object}  dtoresponse.ErrorResponse
// @Failure      403  {object}  dtoresponse.ErrorResponse
// @Failure      404  {object}  dtoresponse.ErrorResponse
// @Router       /api/v1/admin/blog/comments/{id} [delete]
func (h *BlogHandler) AdminDeleteComment(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	if err := h.service.DeleteComment(r.Context(), id); err != nil {
		response.Error(w, r, h.log, err)
		return
	}

	response.NoContent(w)
}
