# Blog Comment

## Purpose

Reader comments on blog posts with moderation workflow. Comments are submitted by visitors and require admin approval before public display.

## Description

Maps to `blog_comments` table. Status workflow: `pending` → `approved` or `rejected`. Moderated from admin `/weblog/comments` UI.

**Implementation status:** Not implemented. Planned in migration `000012_blog`.

## Responsibilities

- Accept visitor comments on blog posts
- Queue comments for moderation
- Display approved comments on public post pages
- Reject spam or inappropriate content

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `post_id` | UUID | No | — | FK → `blog_posts.id` CASCADE | Parent blog post |
| `author_name` | string | No | — | Required, max 255 | Commenter display name |
| `author_email` | string | Yes | `NULL` | Email, max 255 | Optional email (not public) |
| `content` | text | No | — | Required, 3–2000 chars | Comment body |
| `status` | enum | No | `pending` | `pending` \| `approved` \| `rejected` | Moderation state |
| `created_at` | timestamp | No | `now()` | — | Submitted at |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| BlogPost | N:1 | Comment belongs to post |

## Business Rules

1. New comments default to `status = pending`.
2. Only `approved` comments visible on public post page.
3. `author_email` never exposed in public API (admin only).
4. Rate limiting on comment submission (recommended: 5/hour per IP).
5. HTML stripped from `content`; plain text or sanitized markdown only.
6. Deleting a post cascades to its comments.
7. Event `BlogCommentSubmitted` queues moderation notification (planned).
8. Rejected comments retained for audit; not shown publicly.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create (public) | ❌ Planned | POST on blog post |
| Read (list) | ❌ Planned | Admin moderation queue |
| Approve | ❌ Planned | Status transition |
| Reject | ❌ Planned | Status transition |
| Delete | ❌ Planned | Hard delete |

## Required APIs

### Admin

All require admin JWT.

#### `GET /api/v1/admin/blog/comments`

**Query:** `page`, `per_page`, `status` (default `pending`), `post_id`, `search`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "post_id": "uuid",
      "post_title": "10 Tips for Summer Shopping",
      "author_name": "John Reader",
      "author_email": "john@example.com",
      "content": "Great article, thanks!",
      "status": "pending",
      "created_at": "2026-06-28T14:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 8, "total_pages": 1 }
}
```

---

#### `PATCH /api/v1/admin/blog/comments/{id}/approve`

**Response `200`:**
```json
{
  "id": "uuid",
  "status": "approved"
}
```

---

#### `PATCH /api/v1/admin/blog/comments/{id}/reject`

**Response `200`:**
```json
{
  "id": "uuid",
  "status": "rejected"
}
```

---

#### `DELETE /api/v1/admin/blog/comments/{id}`

**Response `204`**

### Storefront

#### `POST /api/v1/store/blog/{slug}/comments`

Public comment submission.

**Request:**
```json
{
  "author_name": "John Reader",
  "author_email": "john@example.com",
  "content": "Great article, thanks!"
}
```

**Response `201`:**
```json
{
  "id": "uuid",
  "message": "Your comment has been submitted and is awaiting moderation."
}
```

**Errors:** `404` post not found, `422` validation, `429` rate limited.

---

#### `GET /api/v1/store/blog/{slug}/comments`

Approved comments only (included in post detail or separate).

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "author_name": "John Reader",
      "content": "Great article, thanks!",
      "created_at": "2026-06-28T14:00:00Z"
    }
  ]
}
```

## Domain Reference (planned)

- Package: `internal/domain/blog/comment/`
- Table: `blog_comments`
- Enum: `comment_status`
