# Brand (Catalog)

## Purpose

Manages the canonical list of product brands used in catalog settings and product assignment. Distinct from **PartnerBrand** (homepage logo carousel for marketing partners).

## Description

The `Brand` aggregate (`internal/domain/brand/entity.go`) maps to the `brands` table. Brands are configured in the admin "Product Settings" section alongside categories and attribute definitions. Products currently store brand as a denormalized string field; future work should link via FK.

**Implementation status:** Fully implemented for admin CRUD.

## Responsibilities

- Maintain a controlled vocabulary of brand names
- Provide slug-based brand identifiers
- Enable brand filtering on product catalog (planned)
- Control brand visibility via `is_active`

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `name` | string | No | — | Required, max 100, unique (soft-delete aware) | Brand display name |
| `slug` | string | No | auto-generated | Unique, max 100, URL-safe | URL segment |
| `description` | text | Yes | `NULL` | — | Brand description |
| `is_active` | bool | No | `true` | — | Visible when true |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |
| `deleted_at` | timestamp | Yes | `NULL` | Soft delete | — |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Product | 1:N (loose) | Products reference brand by name string, not FK |
| PartnerBrand | None | Separate entity for homepage partner logos |

## Business Rules

1. `name` and `slug` must be unique among non-deleted brands.
2. `slug` auto-generated from `name` if omitted.
3. Inactive brands should not appear in product create dropdowns.
4. Deleting a brand does not remove the string from existing products.
5. Brand names are case-sensitive in storage; display normalization is UI concern.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ✅ Implemented | |
| Read (list) | ✅ Implemented | Paginated |
| Read (single) | ✅ Implemented | |
| Update | ✅ Implemented | |
| Delete | ✅ Implemented | Soft delete |
| Public list | ❌ Planned | Storefront brand filter |

## Required APIs

### Admin (implemented)

All require admin JWT.

#### `GET /api/v1/admin/brands`

**Query:** `page`, `per_page`, `search`, `is_active`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Acme",
      "slug": "acme",
      "description": "Premium lifestyle brand",
      "is_active": true,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-06-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 15, "total_pages": 1 }
}
```

---

#### `POST /api/v1/admin/brands`

**Request:**
```json
{
  "name": "Acme",
  "slug": "acme",
  "description": "Premium lifestyle brand",
  "is_active": true
}
```

**Response `201`:** `BrandResponse`

**Errors:** `409` duplicate name/slug, `422` validation.

---

#### `GET /api/v1/admin/brands/{id}`

**Response `200`:** `BrandResponse`

---

#### `PUT /api/v1/admin/brands/{id}`

Partial update.

**Request:**
```json
{
  "name": "Acme Corp",
  "description": "Updated description",
  "is_active": false
}
```

**Response `200`:** `BrandResponse`

---

#### `DELETE /api/v1/admin/brands/{id}`

**Response `204`:** Soft delete.

### Storefront (planned)

#### `GET /api/v1/store/brands`

Active brands for filter sidebar (id, name, slug, product count).

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| Product `brand_id` FK | Medium | Replace string `brand` field on products |
| Brand logo field | Low | Optional `logo_url` for storefront display |
| Product count on list | Low | Denormalized count like categories |

## Domain Reference

- Entity: `internal/domain/brand/entity.go`
- Slug: `internal/domain/brand/slug.go`
- Table: `brands` (migration `000006_catalog_settings.up.sql`)
