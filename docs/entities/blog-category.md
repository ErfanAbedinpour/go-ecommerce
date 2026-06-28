# Blog Category

## Purpose

Organizes blog posts into topical groups for navigation and filtering on the storefront blog and admin weblog settings.

## Description

Maps to `blog_categories` table. Simple flat categories (no nesting). Each category has a unique slug for URL filtering.

**Implementation status:** Not implemented. Planned in migration `000012_blog`.

## Responsibilities

- Group blog posts by topic
- Provide slug-based filter on public blog
- Control category visibility
- Order categories for blog sidebar/navigation

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `name` | string | No | — | Required, max 255 | Category display name |
| `slug` | string | No | auto-generated | Unique, max 255, URL-safe | URL filter segment |
| `sort_order` | int | No | `0` | ≥ 0 | Display order |
| `is_active` | bool | No | `true` | — | Visible when true |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| BlogPost | 1:N | Posts in this category |

## Business Rules

1. `name` and `slug` must be unique.
2. `slug` auto-generated from `name` if omitted.
3. Inactive categories hidden from public filter but existing posts retain `category_id`.
4. Deleting a category sets post `category_id` to NULL (ON DELETE SET NULL).
5. Category with published posts should warn before delete (admin UI).

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ❌ Planned | |
| Read (list) | ❌ Planned | |
| Read (single) | ❌ Planned | Optional |
| Update | ❌ Planned | |
| Delete | ❌ Planned | |

## Required APIs

### Admin

All require admin JWT. Route: `/api/v1/admin/blog/categories`.

#### `GET /api/v1/admin/blog/categories`

**Query:** `is_active`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "News",
      "slug": "news",
      "sort_order": 0,
      "is_active": true,
      "post_count": 15
    }
  ]
}
```

---

#### `POST /api/v1/admin/blog/categories`

**Request:**
```json
{
  "name": "News",
  "slug": "news",
  "sort_order": 0,
  "is_active": true
}
```

**Response `201`:** Category object.

---

#### `GET /api/v1/admin/blog/categories/{id}`

**Response `200`:** Category object.

---

#### `PUT /api/v1/admin/blog/categories/{id}`

Partial update.

---

#### `DELETE /api/v1/admin/blog/categories/{id}`

**Response `204`**

**Errors:** `409` if category has posts (optional strict mode).

### Storefront (public)

Categories included in blog list response and post detail. Optional dedicated endpoint:

#### `GET /api/v1/store/blog/categories`

Active categories with post counts.

## Domain Reference (planned)

- Package: `internal/domain/blog/category/`
- Table: `blog_categories`
