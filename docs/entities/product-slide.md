# Product Slide

## Purpose

Manages homepage product carousels — three distinct slide types (`featured`, `bestseller`, `discounted`) each displaying a curated list of products with configurable autoplay and optional tab labels.

## Description

Maps to `product_slides` (carousel config) and `product_slide_items` (product membership). One row per `slide_type` (unique constraint). Each slide references products by ID with sort order; featured slides support `tab_label` for tabbed UI.

**Implementation status:** Not implemented. Planned in migration `000010_storefront_content`.

## Responsibilities

- Configure three homepage product carousels
- Curate which products appear in each carousel
- Control section titles, autoplay timing, and visibility
- Order products within each carousel

## Attributes

### ProductSlide

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `slide_type` | enum | No | — | `featured` \| `bestseller` \| `discounted` | Carousel identifier (unique) |
| `title` | string | Yes | `NULL` | Max 255 | Section heading |
| `autoplay_interval_ms` | int | No | `4500` | 1000–30000 | Milliseconds between slides |
| `is_active` | bool | No | `true` | — | Show section on homepage |
| `sort_order` | int | No | `0` | ≥ 0 | Section order on homepage |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |

### ProductSlideItem

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `slide_id` | UUID | No | — | FK → `product_slides.id` CASCADE | Parent carousel |
| `product_id` | UUID | No | — | FK → `products.id` | Featured product |
| `sort_order` | int | No | `0` | ≥ 0 | Position in carousel |
| `tab_label` | string | Yes | `NULL` | Max 100 | Tab label (featured type only) |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Product | N:M via items | Products in carousel |
| Homepage aggregate | 1:3 | Three slide sections |

## Business Rules

1. Exactly one `product_slides` row per `slide_type` value.
2. `bestseller` and `discounted` slides may auto-populate from product queries (optional enhancement); manual curation is primary.
3. Only `active` products should appear in public response.
4. Duplicate `product_id` within same slide rejected.
5. Deleting a product removes its slide items (FK cascade or orphan cleanup).
6. `tab_label` only meaningful for `featured` slide type.
7. Minimum 1 product required when slide is active (validation).
8. Cache tag `homepage` invalidated on update.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Read all slides | ❌ Planned | Returns 3 slides with items |
| Update slide | ❌ Planned | PUT per slide type or bulk |
| Manage items | ❌ Planned | Replace item list on update |

## Required APIs

### Admin

All require admin JWT.

#### `GET /api/v1/admin/storefront/product-slides`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "slide_type": "featured",
      "title": "Featured Products",
      "autoplay_interval_ms": 4500,
      "is_active": true,
      "sort_order": 0,
      "items": [
        {
          "id": "uuid",
          "product_id": "uuid",
          "product_name": "Classic T-Shirt",
          "product_slug": "classic-t-shirt",
          "product_image": "/uploads/tee.jpg",
          "product_price": 29.99,
          "sort_order": 0,
          "tab_label": "New Arrivals"
        }
      ],
      "created_at": "2026-06-01T00:00:00Z",
      "updated_at": "2026-06-28T00:00:00Z"
    }
  ]
}
```

---

#### `PUT /api/v1/admin/storefront/product-slides/{slide_type}`

Update a single carousel by type.

**Path param:** `slide_type` = `featured` | `bestseller` | `discounted`

**Request:**
```json
{
  "title": "Featured Products",
  "autoplay_interval_ms": 5000,
  "is_active": true,
  "sort_order": 0,
  "items": [
    {
      "product_id": "uuid",
      "sort_order": 0,
      "tab_label": "New Arrivals"
    },
    {
      "product_id": "uuid",
      "sort_order": 1,
      "tab_label": "Trending"
    }
  ]
}
```

**Response `200`:** Updated slide with enriched product preview.

**Errors:** `404` product not found, `422` validation.

---

#### `PUT /api/v1/admin/storefront/product-slides`

Bulk update all three slides in one request (alternative).

**Request:** `{ "slides": [ ... ] }`

### Storefront (public)

#### `GET /api/v1/store/homepage`

**Response fragment:**
```json
{
  "product_slides": [
    {
      "slide_type": "featured",
      "title": "Featured Products",
      "autoplay_interval_ms": 4500,
      "items": [
        {
          "product": {
            "id": "uuid",
            "name": "Classic T-Shirt",
            "slug": "classic-t-shirt",
            "price": 29.99,
            "sale_price": 24.99,
            "image_url": "/uploads/tee.jpg"
          },
          "tab_label": "New Arrivals"
        }
      ]
    }
  ]
}
```

## Domain Reference (planned)

- Package: `internal/domain/storefront/productslide/`
- Tables: `product_slides`, `product_slide_items`
- Enum: `slide_type`
