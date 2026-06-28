# Attribute Definition

## Purpose

Defines global, reusable product attribute types (e.g., Color, Size, Material) used across the catalog. Attribute definitions provide a controlled vocabulary that product variant attributes can reference.

## Description

The `Definition` aggregate (`internal/domain/attributedef/entity.go`) maps to `product_attribute_definitions`. Each definition has a name, slug, sort order, and active flag. Allowed values are managed separately via **Attribute Value** entities.

Distinct from per-product `ProductAttribute` on the Product entity — definitions are catalog-wide settings; product attributes are instance-specific variant axes.

**Implementation status:** Fully implemented for admin CRUD at `/admin/product-attributes`.

## Responsibilities

- Define reusable attribute type names for the catalog
- Control display order in product settings UI
- Gate availability of attribute types via `is_active`
- Serve as parent for global attribute values

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `name` | string | No | — | Required, max 100, unique (soft-delete aware) | Attribute name (e.g., "Color") |
| `slug` | string | No | auto-generated | Unique, max 100, URL-safe | Identifier |
| `sort_order` | int | No | `0` | ≥ 0 | Display order in settings |
| `is_active` | bool | No | `true` | — | Available for product assignment |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |
| `deleted_at` | timestamp | Yes | `NULL` | Soft delete | — |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| AttributeValue | 1:N | Allowed values for this definition |
| ProductAttribute | 0:N (loose) | Products may use matching name; no FK |

## Business Rules

1. `name` and `slug` unique among non-deleted definitions.
2. `slug` auto-generated from `name` if omitted.
3. Deactivating a definition hides it from product create UI but does not remove existing product attributes.
4. Deleting a definition cascades to its attribute values.
5. Definition names should match product attribute names for consistency (convention, not enforced).

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ✅ Implemented | |
| Read (list) | ✅ Implemented | Paginated |
| Read (single) | ✅ Implemented | |
| Update | ✅ Implemented | |
| Delete | ✅ Implemented | Soft delete |

## Required APIs

### Admin (implemented)

All require admin JWT. Route prefix: `/api/v1/admin/product-attributes`.

#### `GET /api/v1/admin/product-attributes`

**Query:** `page`, `per_page`, `search`, `is_active`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Color",
      "slug": "color",
      "sort_order": 0,
      "is_active": true,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-06-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 8, "total_pages": 1 }
}
```

---

#### `POST /api/v1/admin/product-attributes`

**Request:**
```json
{
  "name": "Color",
  "slug": "color",
  "sort_order": 0,
  "is_active": true
}
```

**Response `201`:** `CatalogAttributeResponse`

---

#### `GET /api/v1/admin/product-attributes/{id}`

**Response `200`:** `CatalogAttributeResponse`

---

#### `PUT /api/v1/admin/product-attributes/{id}`

**Request:**
```json
{
  "name": "Colour",
  "sort_order": 1,
  "is_active": true
}
```

**Response `200`:** `CatalogAttributeResponse`

---

#### `DELETE /api/v1/admin/product-attributes/{id}`

**Response `204`:** Soft delete.

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| Include values in GET single | Low | Eager-load child values for settings UI |
| Link product attributes to definition FK | Medium | Enforce vocabulary consistency |

## Domain Reference

- Entity: `internal/domain/attributedef/entity.go`
- Service: `internal/application/attributedef/service.go`
- Table: `product_attribute_definitions`
