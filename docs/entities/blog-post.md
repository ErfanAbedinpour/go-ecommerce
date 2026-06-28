# Blog Post

## Purpose

Editorial content for the storefront blog (`/blog`). Supports categorization, cover images, draft/publish workflow, and SEO-friendly slugs.

## Description

Maps to `blog_posts` table with optional FK to `blog_categories`. Content stored as Markdown/HTML. Status workflow: `draft` → `published` → `archived`.

**Implementation status:** Not implemented. Planned in migration `000012_blog`.

## Responsibilities

- Publish editorial and marketing content
- Organize posts by blog category
- Control publication schedule via `published_at`
- Provide slug-based public URLs
- Support admin content management from `/weblog` UI

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `category_id` | UUID | Yes | `NULL` | FK → `blog_categories.id` SET NULL | Blog category |
| `title` | string | No | — | Required, max 500 | Post title |
| `slug` | string | No | auto-generated | Unique, max 500, URL-safe | Public URL segment |
| `excerpt` | text | Yes | `NULL` | Max 1000 | Listing card summary |
| `content` | text | No | — | Required, min 10 chars | Full post body (Markdown/HTML) |
| `cover_image_url` | string | Yes | `NULL` | Valid URL, max 500 | Hero/thumbnail image |
| `author_name` | string | Yes | `NULL` | Max 255 | Display author (defaults to admin name) |
| `read_time_minutes` | int | Yes | `NULL` | ≥ 1 | Estimated read time |
| `status` | enum | No | `draft` | `draft` \| `published` \| `archived` | Publication state |
| `published_at` | timestamp | Yes | `NULL` | Required when status=published | Go-live timestamp |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| BlogCategory | N:1 | Optional category |
| BlogComment | 1:N | Reader comments on post |

## Business Rules

1. `slug` auto-generated from `title` if omitted; globally unique.
2. `published` status requires `published_at` ≤ now() for public visibility.
3. `archived` posts hidden from public list but accessible by direct URL (optional) or 404.
4. `read_time_minutes` auto-calculated from word count if omitted (~200 WPM).
5. Only `published` posts with `published_at ≤ now()` appear on public blog.
6. Deleting a category sets post `category_id` to NULL.
7. Full-text search on `title` via GIN index (planned).
8. Cache tag `blog` invalidated on publish/unpublish.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ❌ Planned | |
| Read (list) | ❌ Planned | Admin + public |
| Read (single) | ❌ Planned | By ID (admin) or slug (public) |
| Update | ❌ Planned | |
| Delete | ❌ Planned | Soft delete recommended |
| Publish | ❌ Planned | Status transition helper |

## Required APIs

### Admin

All require admin JWT. Route: `/api/v1/admin/blog/posts`.

#### `GET /api/v1/admin/blog/posts`

**Query:** `page`, `per_page`, `status`, `category_id`, `search`, `sort` (default `published_at DESC`)

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "category_id": "uuid",
      "category_name": "News",
      "title": "10 Tips for Summer Shopping",
      "slug": "10-tips-summer-shopping",
      "excerpt": "Make the most of your summer wardrobe...",
      "cover_image_url": "/uploads/blog/summer.jpg",
      "author_name": "Admin User",
      "read_time_minutes": 5,
      "status": "published",
      "published_at": "2026-06-15T10:00:00Z",
      "comment_count": 12,
      "created_at": "2026-06-10T00:00:00Z",
      "updated_at": "2026-06-15T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 45, "total_pages": 3 }
}
```

---

#### `POST /api/v1/admin/blog/posts`

**Request:**
```json
{
  "category_id": "uuid",
  "title": "10 Tips for Summer Shopping",
  "slug": "10-tips-summer-shopping",
  "excerpt": "Make the most of your summer wardrobe...",
  "content": "# Introduction\n\nSummer is here...",
  "cover_image_url": "/uploads/blog/summer.jpg",
  "author_name": "Admin User",
  "read_time_minutes": 5,
  "status": "draft"
}
```

**Response `201`:** Full post object.

---

#### `GET /api/v1/admin/blog/posts/{id}`

**Response `200`:** Full post including `content`.

---

#### `PUT /api/v1/admin/blog/posts/{id}`

Partial update. Setting `status: "published"` auto-sets `published_at` to now if omitted.

---

#### `DELETE /api/v1/admin/blog/posts/{id}`

**Response `204`**

### Storefront (public)

#### `GET /api/v1/store/blog`

**Query:** `page`, `per_page`, `category` (slug), `search`

**Response `200`:** Paginated published posts (no full content).

---

#### `GET /api/v1/store/blog/{slug}`

**Response `200`:**
```json
{
  "id": "uuid",
  "title": "10 Tips for Summer Shopping",
  "slug": "10-tips-summer-shopping",
  "excerpt": "...",
  "content": "# Introduction\n\n...",
  "cover_image_url": "/uploads/blog/summer.jpg",
  "author_name": "Admin User",
  "read_time_minutes": 5,
  "category": { "id": "uuid", "name": "News", "slug": "news" },
  "published_at": "2026-06-15T10:00:00Z"
}
```

## Domain Reference (planned)

- Package: `internal/domain/blog/post/`
- Table: `blog_posts`
- Enum: `blog_post_status`
