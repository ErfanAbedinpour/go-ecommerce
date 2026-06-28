# Attribute Value

## Purpose

Stores allowed values for a global attribute definition (e.g., Color → Red, Blue, Green). Provides preset options for product variant configuration in the admin product settings UI.

## Description

The `Value` aggregate (`internal/domain/attributevalue/entity.go`) maps to `product_attribute_values`. Each value belongs to exactly one attribute definition via `attribute_id`. Values have independent sort order and active flags.

**Implementation status:** Fully implemented for admin CRUD at `/admin/product-attribute-values`.

## Responsibilities

- Define preset values for catalog attribute types
- Control value display order
- Gate value availability via `is_active`
- Support product variant matrix configuration

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `attribute_id` | UUID | No | — | FK → `product_attribute_definitions.id` CASCADE | Parent definition |
| `value` | string | No | — | Required, max 200, unique per attribute | Value text (e.g., "Red") |
| `sort_order` | int | No | `0` | ≥ 0 | Display order |
| `is_active` | bool | No | `true` | — | Available for selection |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |
| `deleted_at` | timestamp | Yes | `NULL` | Soft delete | — |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| AttributeDefinition | N:1 | Parent attribute type |
| ProductAttributeValue | 0:N (loose) | Products may use matching value text; no FK |

## Business Rules

1. `(attribute_id, value)` must be unique among non-deleted rows.
2. `attribute_id` must reference an existing, non-deleted definition.
3. Deactivating a value hides it from dropdowns but does not affect existing product SKUs.
4. Deleting a value does not cascade to product variant values.
5. Value text is trimmed; empty strings rejected.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ✅ Implemented | |
| Read (list) | ✅ Implemented | Filterable by `attribute_id` |
| Read (single) | ✅ Implemented | |
| Update | ✅ Implemented | |
| Delete | ✅ Implemented | Soft delete |

## Required APIs

### Admin (implemented)

All require admin JWT. Route prefix: `/api/v1/admin/product-attribute-values`.

#### `GET /api/v1/admin/product-attribute-values`

**Query:** `page`, `per_page`, `attribute_id` (required for filtered list), `is_active`, `search`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "attribute_id": "uuid",
      "value": "Red",
      "sort_order": 0,
      "is_active": true,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-06-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 50, "total": 12, "total_pages": 1 }
}
```

---

#### `POST /api/v1/admin/product-attribute-values`

**Request:**
```json
{
  "attribute_id": "uuid",
  "value": "Red",
  "sort_order": 0,
  "is_active": true
}
```

**Response `201`:** `CatalogAttributeValueResponse`

**Errors:** `404` attribute not found, `409` duplicate value, `422` validation.

---

#### `GET /api/v1/admin/product-attribute-values/{id}`

**Response `200`:** `CatalogAttributeValueResponse`

---

#### `PUT /api/v1/admin/product-attribute-values/{id}`

**Request:**
```json
{
  "value": "Crimson Red",
  "sort_order": 1,
  "is_active": false
}
```

**Response `200`:** `CatalogAttributeValueResponse`

---

#### `DELETE /api/v1/admin/product-attribute-values/{id}`

**Response `204`:** Soft delete.

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| Bulk create values | Low | Import multiple values per attribute |
| Color swatch metadata | Low | Optional `hex_code` or `image_url` for UI |

## Domain Reference

- Entity: `internal/domain/attributevalue/entity.go`
- Service: `internal/application/attributevalue/service.go`
- Table: `product_attribute_values`
