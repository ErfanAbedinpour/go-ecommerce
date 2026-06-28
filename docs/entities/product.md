# Product

## Purpose

Central catalog entity representing sellable items. Products support rich media, variant attributes, SKU matrix, inventory tracking, and pricing — forming the backbone of both admin catalog management and the public storefront.

## Description

The `Product` aggregate (`internal/domain/product/entity.go`) is the root for images, variant attributes, SKUs, and inventory. Products belong to an optional category and reference a brand name (string field; separate `Brand` entity exists for catalog settings). Status controls visibility: `draft`, `active`, `archived`.

Child entities:
- **Image** — gallery images with sort order
- **ProductAttribute** + **ProductAttributeValue** — variant axes (e.g., Color: Red/Blue, Size: S/M/L)
- **Sku** — concrete purchasable variant with unique code and attribute map
- **Inventory** — stock quantity and low-stock threshold (product-level, not per-SKU currently)

**Implementation status:** Full admin CRUD. SKU auto-generation from attributes exists in domain; create/update DTOs lack explicit SKU input. Public storefront APIs planned.

## Responsibilities

- Define product identity, pricing, and merchandising metadata
- Manage variant attribute matrix and derived SKUs
- Track inventory and low-stock alerts
- Support featured product highlighting on homepage/dashboard
- Provide slug-based public URLs (planned)

## Attributes

### Product

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `category_id` | UUID | Yes | `NULL` | FK → `categories.id` | Product category |
| `name` | string | No | — | Required, 1–300 chars | Display name |
| `slug` | string | No | auto-generated | Unique, max 300, URL-safe | Public URL segment |
| `description` | text | Yes | `NULL` | — | Full HTML/Markdown description |
| `short_description` | string | Yes | `NULL` | Max 500 | Listing card excerpt |
| `price` | decimal | No | — | ≥ 0, required | Base price |
| `sale_price` | decimal | Yes | `NULL` | ≥ 0 if set | Promotional price |
| `brand` | string | Yes | `NULL` | Max 100 | Brand name (denormalized) |
| `is_featured` | bool | No | `false` | — | Homepage/dashboard highlight |
| `status` | enum | No | `draft` | `draft` \| `active` \| `archived` | Lifecycle state |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |
| `deleted_at` | timestamp | Yes | `NULL` | Soft delete | — |

### Image

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | — | Image ID |
| `product_id` | UUID | No | — | FK → `products.id` CASCADE | Parent product |
| `url` | string | No | — | Required, valid URL, max 500 | Image URL (upload or external) |
| `alt_text` | string | Yes | `NULL` | Max 200 | Accessibility text |
| `sort_order` | int | No | `0` | ≥ 0 | Gallery order |
| `created_at` | timestamp | No | `now()` | — | — |

### ProductAttribute

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | — | Attribute ID |
| `product_id` | UUID | No | — | FK → `products.id` CASCADE | Parent product |
| `name` | string | No | — | Required, max 100 | Attribute name (e.g., "Color") |

### ProductAttributeValue

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | — | Value ID |
| `attribute_id` | UUID | No | — | FK → `product_attributes.id` CASCADE | Parent attribute |
| `value` | string | No | — | Required, max 200 | Value (e.g., "Red") |

### Sku (variant)

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | — | SKU ID |
| `product_id` | UUID | No | — | FK → `products.id` CASCADE | Parent product |
| `code` | string | No | — | Unique globally, max 100 | SKU code (e.g., `PROD-COLOR-RED-SIZE-M`) |
| `attributes` | JSONB | No | `{}` | Map of attr name → value | Variant combination |
| `price_override` | decimal | Yes | `NULL` | ≥ 0 | **Planned** per-variant price |
| `sale_price_override` | decimal | Yes | `NULL` | ≥ 0 | **Planned** per-variant sale price |
| `created_at` | timestamp | No | `now()` | — | — |

### Inventory

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | — | Inventory ID |
| `product_id` | UUID | No | — | FK → `products.id` UNIQUE CASCADE | One per product |
| `quantity` | int | No | `0` | ≥ 0 | Stock on hand |
| `low_stock_threshold` | int | No | `10` | ≥ 0 | Alert threshold |
| `updated_at` | timestamp | No | `now()` | — | Last stock change |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Category | N:1 | Optional parent category |
| Brand (catalog) | N:1 (loose) | `brand` string field; not FK to `brands` table |
| Image | 1:N | Product gallery |
| ProductAttribute | 1:N | Variant axes |
| Sku | 1:N | Generated variant combinations |
| Inventory | 1:1 | Stock tracking |
| OrderItem | 1:N | Line items reference product + SKU snapshot |
| ProductSlideItem (planned) | N:M | Featured in homepage carousels |
| WishlistItem (planned) | N:M | Customer wishlists |
| ProductReview (planned) | 1:N | Customer reviews |
| ProductQuestion (planned) | 1:N | Q&A |

## Business Rules

1. `slug` auto-generated from `name` if omitted; must remain unique.
2. `sale_price` must be ≤ `price` when both set (recommended validation).
3. `EffectivePrice()` returns `sale_price` when set and ≥ 0, else `price`.
4. SKUs are regenerated from attribute Cartesian product on create/update.
5. Each SKU `code` is globally unique.
6. Only `active` products appear on public storefront (planned).
7. Inventory decrement occurs on order placement (planned event handler).
8. `is_low_stock` when `quantity ≤ low_stock_threshold`; `is_out_of_stock` when `quantity = 0`.
9. Maximum 10 images per product.
10. Archived products are hidden from catalog but retain order history references.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ✅ Implemented | With images, attributes, inventory |
| Read (list) | ✅ Implemented | Paginated, filterable |
| Read (search) | ✅ Implemented | Full-text search endpoint |
| Read (single) | ✅ Implemented | Full aggregate |
| Read (stats) | ✅ Implemented | KPI counts |
| Update | ✅ Implemented | Partial; images/attributes replace |
| Delete | ✅ Implemented | Soft delete |
| Update inventory | ✅ Implemented | `PATCH /{id}/inventory` |
| Public list/detail | ❌ Planned | Storefront endpoints |

## Required APIs

### Admin (implemented)

All require admin JWT.

#### `GET /api/v1/admin/products/stats`

**Response `200`:**
```json
{ "total": 250, "active": 180, "draft": 45, "out_of_stock": 12 }
```

---

#### `GET /api/v1/admin/products/search`

**Query:** `q` (search term), `limit` (default 10)

**Response `200`:** Array of compact product matches.

---

#### `GET /api/v1/admin/products`

**Query:** `page`, `per_page`, `status`, `category_id`, `is_featured`, `search`, `sort`, `order`

**Response `200`:** `ProductListResponse`

---

#### `POST /api/v1/admin/products`

**Request:**
```json
{
  "name": "Classic T-Shirt",
  "slug": "classic-t-shirt",
  "description": "Premium cotton tee",
  "short_description": "Soft cotton tee",
  "price": 29.99,
  "sale_price": 24.99,
  "category_id": "uuid",
  "brand": "Acme",
  "is_featured": true,
  "status": "active",
  "images": [
    { "url": "/uploads/abc.jpg", "alt_text": "Front view", "sort_order": 0 }
  ],
  "attributes": [
    { "name": "Color", "values": ["Red", "Blue"] },
    { "name": "Size", "values": ["S", "M", "L"] }
  ],
  "inventory": { "quantity": 100, "low_stock_threshold": 10 }
}
```

**Response `201`:** `ProductResponse` (includes auto-generated SKUs)

---

#### `GET /api/v1/admin/products/{id}`

**Response `200`:**
```json
{
  "id": "uuid",
  "category_id": "uuid",
  "name": "Classic T-Shirt",
  "slug": "classic-t-shirt",
  "description": "Premium cotton tee",
  "short_description": "Soft cotton tee",
  "price": 29.99,
  "sale_price": 24.99,
  "brand": "Acme",
  "is_featured": true,
  "status": "active",
  "images": [{ "id": "uuid", "url": "/uploads/abc.jpg", "alt_text": "Front view", "sort_order": 0 }],
  "attributes": [{ "id": "uuid", "name": "Color", "values": ["Red", "Blue"] }],
  "skus": [
    { "id": "uuid", "code": "CLASSIC-T-SHIRT-RED-S", "attributes": { "Color": "Red", "Size": "S" } }
  ],
  "inventory": { "quantity": 100, "low_stock_threshold": 10, "is_low_stock": false, "is_out_of_stock": false },
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-06-28T00:00:00Z"
}
```

---

#### `PUT /api/v1/admin/products/{id}`

Partial update. `images` and `attributes` arrays replace existing when provided.

---

#### `DELETE /api/v1/admin/products/{id}`

**Response `204`**

---

#### `PATCH /api/v1/admin/products/{id}/inventory`

**Request:**
```json
{
  "quantity": 85,
  "low_stock_threshold": 15,
  "adjustment_reason": "Stock count correction"
}
```

**Response `200`:** Updated inventory fields.

### Storefront (planned)

#### `GET /api/v1/store/products`

Public catalog with filters: `category`, `brand`, `min_price`, `max_price`, `search`, `sort`, pagination.

#### `GET /api/v1/store/products/{slug}`

Public product detail with variants, images, reviews.

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| SKU in create/update DTO | High | Accept explicit SKU codes and per-variant pricing |
| `price_override` / `sale_price_override` on SKUs | High | Migration `000014_sku_pricing` |
| Link `brand` to `brands.id` FK | Medium | Replace string with FK |
| Per-SKU inventory | Medium | Currently product-level only |
| Public catalog endpoints | Critical | Storefront product list/detail |
| Brand filter on product list | Low | Query param exists partially |

## Domain Reference

- Entity: `internal/domain/product/entity.go`
- Status: `internal/domain/product/status.go`
- Tables: `products`, `product_images`, `product_attributes`, `product_variant_attribute_values`, `skus`, `inventories`
