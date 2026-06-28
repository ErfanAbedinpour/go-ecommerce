# Admin Blog (Weblog)

**Routes:** `/weblog`, `/weblog/create`, `/posts/create`, `/weblog/settings`, `/weblog/comments`  
**Status:** 🆕 Not implemented (proposed)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Full blog CMS for the admin panel: manage posts (create, edit, publish, archive), organize categories, and moderate reader comments. Published content appears on the storefront at `/blog` and `/blog/:slug`.

`/posts/create` is an alias route for `/weblog/create`.

---

## User Flow

### `/weblog` — Posts list

1. Load posts: `GET /admin/blog/posts`.
2. Filter by status (`draft`, `published`, `archived`), category, search.
3. Row actions: Edit, Delete, Publish/Unpublish.

### `/weblog/create` and `/posts/create` — Create post

1. Load categories: `GET /admin/blog/categories`.
2. Fill title, slug, excerpt, content (markdown/HTML), cover image, author, category.
3. Upload cover → `POST /admin/uploads?context=blog`.
4. Save draft → `POST /admin/blog/posts` with `status: draft`.
5. Publish → set `status: published` + `published_at`.

### `/weblog/settings` — Categories

CRUD at `/admin/blog/categories`.

### `/weblog/comments` — Comment moderation

1. Load pending comments: `GET /admin/blog/comments?status=pending`.
2. Approve → `PATCH /admin/blog/comments/{id}/approve`.
3. Reject → `PATCH /admin/blog/comments/{id}/reject`.

---

## Business Logic

### Post lifecycle

| Status | Meaning |
|--------|---------|
| `draft` | Not visible on storefront |
| `published` | Visible; requires `published_at` |
| `archived` | Hidden; retained for admin |

- `slug` unique globally; auto-generated from title if omitted.
- `read_time_minutes` optional; can be computed client-side (~200 wpm).
- Search uses GIN index on title (migration `000012`).

### Categories

- Separate from product categories.
- Posts optionally linked via `category_id`.
- Inactive categories hidden from storefront filter but posts remain accessible by slug.

### Comments

- Submitted from storefront (`POST /store/blog/{slug}/comments`) — public endpoint.
- Default status: `pending`.
- Only `approved` comments visible on storefront.
- Moderation: approve or reject; no edit in v1.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Publish without cover | Allowed |
| Duplicate slug | `409 CONFLICT` |
| Delete category with posts | `422` or set posts `category_id = null` |
| Delete published post | Soft delete or archive (recommend archive) |
| Comment spam | Rate limit on public POST; admin bulk reject |
| Empty content | `400` validation |
| Future `published_at` | Optional scheduled publish (v2); v1 sets immediately |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Upload service | Cover images |
| Blog module | Posts, categories, comments |
| Store blog API | Public list + detail |
| Cache | Invalidate tag `blog` on publish |

---

## Required APIs

### Posts

#### GET `/api/v1/admin/blog/posts`

**Query:** `page`, `per_page`, `status`, `category_id`, `q`

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "category_id": "uuid",
      "category_name": "Tips",
      "title": "How to Choose Ceramic Tiles",
      "slug": "how-to-choose-ceramic-tiles",
      "excerpt": "A guide for homeowners…",
      "cover_image_url": "https://…",
      "author_name": "Admin",
      "read_time_minutes": 5,
      "status": "published",
      "published_at": "2026-06-01T08:00:00Z",
      "created_at": "2026-05-28T00:00:00Z",
      "updated_at": "2026-06-01T08:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 42, "total_pages": 3 }
}
```

#### POST `/api/v1/admin/blog/posts`

**Request:**

```json
{
  "category_id": "uuid",
  "title": "How to Choose Ceramic Tiles",
  "slug": "how-to-choose-ceramic-tiles",
  "excerpt": "Short summary",
  "content": "# Full markdown content…",
  "cover_image_url": "https://…",
  "author_name": "Admin",
  "read_time_minutes": 5,
  "status": "draft"
}
```

**Response 201:** Post object.

**Errors:** `400`, `409`, `401`, `403`

#### GET `/api/v1/admin/blog/posts/{id}`

**Response 200:** Full post including `content`.

#### PUT `/api/v1/admin/blog/posts/{id}`

**Response 200:** Updated post.

#### DELETE `/api/v1/admin/blog/posts/{id}`

**Response 204**

---

### Categories

| Method | Route |
|--------|-------|
| GET | `/api/v1/admin/blog/categories` |
| POST | `/api/v1/admin/blog/categories` |
| GET | `/api/v1/admin/blog/categories/{id}` |
| PUT | `/api/v1/admin/blog/categories/{id}` |
| DELETE | `/api/v1/admin/blog/categories/{id}` |

**Create category:**

```json
{ "name": "Tips", "slug": "tips", "sort_order": 0, "is_active": true }
```

---

### Comments

#### GET `/api/v1/admin/blog/comments`

**Query:** `page`, `per_page`, `status` (`pending`, `approved`, `rejected`), `post_id`

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "post_id": "uuid",
      "post_title": "How to Choose Ceramic Tiles",
      "author_name": "Sara",
      "author_email": "sara@example.com",
      "content": "Great article!",
      "status": "pending",
      "created_at": "2026-06-01T14:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

#### PATCH `/api/v1/admin/blog/comments/{id}/approve`

**Response 200:** `{ "status": "approved" }`

#### PATCH `/api/v1/admin/blog/comments/{id}/reject`

**Response 200:** `{ "status": "rejected" }`

---

## Database Impact

**Tables:** `blog_categories`, `blog_posts`, `blog_comments` (migration `000012`)

| Table | Key indexes |
|-------|-------------|
| `blog_posts` | `(status, published_at DESC)`, GIN on title |
| `blog_comments` | `(post_id, status)` |

**Cache:** Invalidate `blog` on post publish/unpublish, comment approve.

---

## UI Changes Affecting Backend

| UI element | Backend impact |
|------------|----------------|
| Rich text / markdown editor | `content` TEXT field |
| Slug auto-generate | Send slug or let backend derive |
| Publish button | PUT with `status: published`, set `published_at` |
| Category sidebar on settings | Categories CRUD |
| Comment moderation queue | Filter `status=pending` |
| `/posts/create` alias | Same API as `/weblog/create` |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `title` | Required; max 500 |
| `slug` | Unique; max 500; URL-safe |
| `content` | Required on publish; max 100000 |
| `excerpt` | Optional; max 1000 |
| `status` | `draft` \| `published` \| `archived` |
| `category_id` | Optional UUID |
| `cover_image_url` | Optional URL |
| Comment `content` | Required; max 2000 |

---

## Permission Requirements

| Action | Role |
|--------|------|
| All blog admin APIs | `admin` |
| Submit comment (storefront) | Public |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Post list filters | URL query | `/weblog` |
| Editor draft | `sessionStorage` + auto-save | Create/edit |
| Comment moderation tab | URL `?status=pending` | Comments |
| Category list cache | React Query | Settings |
