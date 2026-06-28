# Category

## Purpose

Organizes products into a hierarchical taxonomy for navigation, filtering, and merchandising. Categories support parent-child nesting for multi-level menus.

## Description

The `Category` aggregate (`internal/domain/category/entity.go`) maps to the `categories` table with self-referential `parent_id`. Categories have URL slugs, optional images, sort ordering, and an active flag. The `products_count` field is computed at query time for admin list views.

**Implementation status:** Fully implemented for admin CRUD.

## Responsibilities

- Structure the product catalog hierarchy
- Provide slug-based URLs for storefront category pages (planned)
- Control category visibility via `is_active`
- Order categories for navigation display

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `parent_id` | UUID | Yes | `NULL` | FK → `categories.id` SET NULL | Parent category for nesting |
| `name` | string | No | — | Required, max 200 | Display name |
| `slug` | string | No | auto-generated | Unique, max 200, URL-safe | Public URL segment |
| `description` | text | Yes | `NULL` | — | Category description |
| `image_url` | string | Yes | `NULL` | Valid URL, max 500 | Category banner/icon |
| `sort_order` | int | No | `0` | ≥ 0 | Display order among siblings |
| `is_active` | bool | No | `true` | — | Visible when true |
| `products_count` | int64 | — | computed | ≥ 0 | Denormalized count (query-time) |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |
| `deleted_at` | timestamp | Yes | `NULL` | Soft delete | — |

### Children (computed)

Nested `Category` array populated when fetching tree views. Not stored separately.

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Category (parent) | N:1 | Self-referential parent |
| Category (children) | 1:N | Child categories |
| Product | 1:N | Products reference `category_id` |

## Business Rules

1. `slug` auto-generated from `name` if omitted; must be globally unique.
2. A category cannot be its own parent or ancestor (cycle prevention required).
3. Deleting a category with children sets children's `parent_id` to NULL (ON DELETE SET NULL).
4. Deleting a category with products sets product `category_id` to NULL.
5. Inactive categories and their products remain in admin but hidden from storefront (planned).
6. `sort_order` applies among siblings with the same `parent_id`.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ✅ Implemented | |
| Read (list) | ✅ Implemented | Flat paginated list |
| Read (tree) | ✅ Implemented | Nested children in response |
| Read (single) | ✅ Implemented | |
| Update | ✅ Implemented | |
| Delete | ✅ Implemented | Soft delete |
| Public list | ❌ Planned | Storefront category tree |

## Required APIs

### Admin (implemented)

All require admin JWT.

#### `GET /api/v1/admin/categories`

**Query:** `page`, `per_page`, `search`, `parent_id`, `is_active`, `tree` (boolean — return nested)

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "parent_id": null,
      "name": "Electronics",
      "slug": "electronics",
      "description": "Electronic devices",
      "image_url": "/uploads/cat-electronics.jpg",
      "sort_order": 0,
      "is_active": true,
      "products_count": 45,
      "children": [
        {
          "id": "uuid",
          "parent_id": "uuid",
          "name": "Phones",
          "slug": "phones",
          "sort_order": 0,
          "is_active": true,
          "products_count": 20
        }
      ],
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-06-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 50, "total": 12, "total_pages": 1 }
}
```

---

#### `POST /api/v1/admin/categories`

**Request:**
```json
{
  "name": "Electronics",
  "slug": "electronics",
  "description": "Electronic devices and accessories",
  "image_url": "/uploads/cat-electronics.jpg",
  "parent_id": null,
  "sort_order": 0,
  "is_active": true
}
```

**Response `201`:** `CategoryResponse`

---

#### `GET /api/v1/admin/categories/{id}`

**Response `200`:** `CategoryResponse`

---

#### `PUT /api/v1/admin/categories/{id}`

Partial update. All fields optional.

---

#### `DELETE /api/v1/admin/categories/{id}`

**Response `204`:** Soft delete.

### Storefront (planned)

#### `GET /api/v1/store/categories`

Public category tree (active only, with product counts).

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| Cycle detection on parent update | Medium | Prevent circular parent references |
| Public category API | High | Storefront navigation and filtering |
| Category image upload helper | Low | Use existing `/admin/uploads` |

## Domain Reference

- Entity: `internal/domain/category/entity.go`
- Slug: `internal/domain/category/slug.go`
- Table: `categories`
