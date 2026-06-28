# Store Blog

> **Routes:** `/blog`, `/blog/:slug`  
> **UI:** [store-os-eta.vercel.app/blog](https://store-os-eta.vercel.app/blog)  
> **Locale:** Persian (fa-IR), RTL

---

## Purpose

The blog provides educational and SEO content for building materials customers — installation guides, product comparisons, industry news. The listing page supports category filter and search; the detail page renders full article content with related posts and optional comments. Homepage shows a 3-post teaser via `blog_teaser` in homepage API.

---

## User Flow

### Listing (`/blog`)

```mermaid
flowchart TD
    A[/blog] --> B[GET /store/blog]
    B --> C[Render post cards grid]
    C --> D{Filter}
    D -->|Category| E[?category=slug]
    D -->|Search| F[?q=term]
    E --> G[Refetch list]
    F --> G
    C --> H[Click card]
    H --> I[/blog/:slug]
```

1. Load paginated published posts, newest first.
2. Sidebar or chips: blog categories (راهنما، اخبار، مقایسه محصول).
3. Search debounced → filter by title/excerpt.
4. Card: cover image, title, excerpt, category, read time, Jalali date.
5. Pagination or infinite scroll.

### Detail (`/blog/:slug`)

```mermaid
flowchart TD
    A[/blog/:slug] --> B[GET /store/blog/:slug]
    B --> C[Render article]
    C --> D[Related posts]
    C --> E[Comments section]
    E --> F[Submit comment form]
    F --> G[POST /store/blog/:slug/comments]
    G --> H[Pending moderation message]
```

1. Load post by slug; `404` if draft/archived.
2. Render `content_html` or markdown-rendered body.
3. Author, published date (Jalali), read time, category breadcrumb.
4. Related posts (same category, limit 3).
5. Comments: show approved only; submit creates `pending` comment.

---

## Business Logic

### Visibility

- Only posts with `status = 'published'` and `published_at <= NOW()`.
- Draft/archived → `404` on public API.

### Sorting

- Default: `published_at DESC`.
- Optional: `sort=popular` (comment count or view count — views deferred to v2).

### Categories

- Filter by `category_slug` or `category_id`.
- Inactive categories hidden.

### Content rendering

- `content` stored as Markdown or HTML in `blog_posts.content`.
- API returns both `content_html` (sanitized) and raw `content_markdown` for client choice.
- Cover image required for cards; fallback placeholder.

### Comments

- Public submit: name required, email optional.
- `status = 'pending'` on create; visible after admin approval.
- Admin moderates via `/api/v1/admin/blog/comments`.

### Related posts

- Same `category_id`, exclude current, limit 3, `published_at DESC`.

### Homepage teaser

- Subset of blog list (latest 3) embedded in `GET /store/homepage` → `blog_teaser.posts`.

### SEO per post

- `meta_title`, `meta_description` optional fields on post or derived from title/excerpt.
- Open Graph: `cover_image_url`, title, description.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Invalid slug | `404` |
| Post unpublished after share | `404`; social caches may stale |
| Empty blog | Empty state "مطلبی یافت نشد" |
| Category with no posts | Hide category filter or show zero state |
| XSS in content | Sanitize HTML server-side (bluemonday) |
| Comment spam | Rate limit; honeypot; moderation queue |
| Very long articles | Table of contents from H2 headings (frontend optional) |
| Missing cover image | Default building-materials placeholder |
| Persian slug | URL-encode; support Unicode slugs |

---

## Dependencies

### Backend modules

| Module | Role |
|--------|------|
| `internal/application/blog` | Public list, detail, comments |
| `internal/application/storecontent` | Homepage teaser (or blog service) |

### Tables

- `blog_posts`, `blog_categories`, `blog_comments`

### Admin (content authoring)

- `/weblog` — CRUD posts
- `/weblog/settings` — categories
- `/weblog/comments` — moderation

### Frontend

- Markdown renderer (e.g. `react-markdown`) with RTL prose
- Jalali date formatting
- SEO `<head>` from post meta

---

## Required APIs

### `GET /api/v1/store/blog`

Paginated blog post list.

**Query parameters**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | |
| `per_page` | int | 12 | Max 24 |
| `category_slug` | string | — | Filter by category |
| `category_id` | uuid | — | Alternative filter |
| `q` | string | — | Search title/excerpt |

**Response 200**

```json
{
  "data": [
    {
      "id": "uuid",
      "title": "راهنمای انتخاب کاشی کف",
      "slug": "floor-tile-guide",
      "excerpt": "نکات مهم در انتخاب کاشی برای آشپزخانه و سرویس…",
      "cover_image_url": "https://…",
      "category": {
        "id": "uuid",
        "name": "راهنما",
        "slug": "guides"
      },
      "author_name": "تیم تحریریه",
      "read_time_minutes": 7,
      "published_at": "2026-06-01T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 12, "total": 34, "total_pages": 3 }
}
```

### `GET /api/v1/store/blog/{slug}`

Single published post.

**Response 200**

```json
{
  "id": "uuid",
  "title": "راهنمای انتخاب کاشی کف",
  "slug": "floor-tile-guide",
  "excerpt": "…",
  "content_html": "<h2>مقدمه</h2><p>…</p>",
  "content_markdown": "## مقدمه\n\n…",
  "cover_image_url": "https://…",
  "category": {
    "id": "uuid",
    "name": "راهنما",
    "slug": "guides"
  },
  "author_name": "تیم تحریریه",
  "read_time_minutes": 7,
  "published_at": "2026-06-01T10:00:00Z",
  "updated_at": "2026-06-02T08:00:00Z",
  "seo": {
    "meta_title": "راهنمای انتخاب کاشی کف | بلاگ",
    "meta_description": "…"
  },
  "related_posts": [
    {
      "id": "uuid",
      "title": "مقایسه کاشی سرامیک و پرسلان",
      "slug": "ceramic-vs-porcelain",
      "cover_image_url": "https://…",
      "published_at": "2026-05-20T10:00:00Z"
    }
  ]
}
```

**Response 404:** Post not found or not published.

### `GET /api/v1/store/blog/categories`

Active categories for filter UI.

**Response 200**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "راهنما",
      "slug": "guides",
      "posts_count": 12
    },
    {
      "id": "uuid",
      "name": "اخبار",
      "slug": "news",
      "posts_count": 8
    }
  ]
}
```

### `GET /api/v1/store/blog/{slug}/comments`

Approved comments only.

**Query:** `page`, `per_page`

**Response 200**

```json
{
  "data": [
    {
      "id": "uuid",
      "author_name": "علی",
      "content": "مطلب بسیار مفیدی بود.",
      "created_at": "2026-06-03T14:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

### `POST /api/v1/store/blog/{slug}/comments`

Submit comment (moderation required).

**Auth:** Public (rate-limited)

**Request**

```json
{
  "author_name": "علی رضایی",
  "author_email": "ali@example.com",
  "content": "مطلب بسیار مفیدی بود. ممنون."
}
```

**Response 201**

```json
{
  "id": "uuid",
  "status": "pending",
  "message": "نظر شما پس از تأیید نمایش داده می‌شود."
}
```

---

## Database Impact

### Reads

| Table | Query |
|-------|-------|
| `blog_posts` | `status = 'published'`, ordered by `published_at` |
| `blog_categories` | Active categories with post counts |
| `blog_comments` | `status = 'approved'` for public list |

### Writes

| Table | Operation |
|-------|-----------|
| `blog_comments` | INSERT on comment submit |

### Schema (from migration 000012)

See [database/schema-changes.md](../database/schema-changes.md) for `blog_categories`, `blog_posts`, `blog_comments`.

### Indexes

```sql
CREATE INDEX idx_blog_posts_published ON blog_posts (status, published_at DESC);
CREATE INDEX idx_blog_comments_post ON blog_comments (post_id, status);
```

---

## Validation

### List query

| Param | Rules |
|-------|-------|
| `page` | >= 1 |
| `per_page` | 1–24 |
| `q` | Max 200 chars |

### Comment POST

| Field | Rules |
|-------|-------|
| `author_name` | Required, 2–255 chars |
| `author_email` | Optional, valid email |
| `content` | Required, 5–2000 chars |

### Slug param

- Resolve post by slug; case-sensitive match on stored slug.

---

## Permissions

| Action | Role |
|--------|------|
| List/read published posts | Public |
| List categories | Public |
| Read approved comments | Public |
| Submit comment | Public (rate-limited) |
| Create/edit/publish posts | Admin |
| Moderate comments | Admin |

---

## State Management

### Listing page

| State | Storage |
|-------|---------|
| Post list | React Query `['blog', { page, category, q }]` |
| Active category filter | URL `?category=guides` |
| Search query | URL `?q=` debounced |
| Pagination | URL `?page=` |

### Detail page

| State | Storage |
|-------|---------|
| Post data | React Query `['blog', slug]` |
| Comments | React Query `['blog', slug, 'comments']` |
| Comment form | React Hook Form |
| Submit status | Local `idle \| submitting \| success` |

### Cache strategy

| Resource | staleTime |
|----------|-----------|
| Blog list | 60s |
| Post detail | 120s |
| Categories | 300s |

Invalidate on admin publish (not automatic from store; use TTL).

### SEO / meta

```typescript
// Pseudocode — set from post.seo on detail mount
useEffect(() => {
  document.title = post.seo.meta_title ?? post.title;
}, [post]);
```

### Related navigation

- Breadcrumb: خانه → بلاگ → {category} → {title}
- Back to list preserves filters via browser history

---

## Sample content categories (building materials)

| Slug | Persian name | Example topics |
|------|--------------|----------------|
| `guides` | راهنما | Tile installation, cement mixing ratios |
| `news` | اخبار | Price updates, new arrivals |
| `comparisons` | مقایسه | Ceramic vs porcelain, tool brands |
| `projects` | پروژه‌ها | Case studies, before/after |

---

## API summary table

| Method | Path | Auth | Purpose |
|--------|------|------|---------|
| `GET` | `/api/v1/store/blog` | Public | Post list |
| `GET` | `/api/v1/store/blog/categories` | Public | Category filters |
| `GET` | `/api/v1/store/blog/{slug}` | Public | Post detail + related |
| `GET` | `/api/v1/store/blog/{slug}/comments` | Public | Approved comments |
| `POST` | `/api/v1/store/blog/{slug}/comments` | Public | Submit comment |
