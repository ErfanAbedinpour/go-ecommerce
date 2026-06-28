# Wishlist Item

## Purpose

Allows authenticated customers to save products for later purchase. Server-synced wishlist accessible from account page (`/account/wishlist`) and product cards across the storefront.

## Description

Maps to `wishlist_items` table. Junction between `customers` and `products` with unique constraint preventing duplicate entries. Created when a logged-in customer clicks the wishlist heart icon.

**Implementation status:** Not implemented. Planned in migration `000013_engagement`.

## Responsibilities

- Persist customer product favorites server-side
- Prevent duplicate product entries per customer
- Provide wishlist data for account page and product card state
- Support add/remove from product listing and detail pages

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `customer_id` | UUID | No | — | FK → `customers.id` CASCADE | Wishlist owner |
| `product_id` | UUID | No | — | FK → `products.id` CASCADE | Saved product |
| `created_at` | timestamp | No | `now()` | — | Added at |

**Unique constraint:** `(customer_id, product_id)`

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Customer | N:1 | Wishlist owner |
| Product | N:1 | Saved product |

## Business Rules

1. Only authenticated customers (`role = customer`) may manage wishlists.
2. Duplicate add returns `409` or idempotent `200` (recommend idempotent).
3. Only `active` products should appear in wishlist response (filter inactive/archived).
4. Deleting a product removes associated wishlist items (CASCADE).
5. Deleting a customer removes their wishlist (CASCADE).
6. Guest users: wishlist stored in localStorage only (frontend v1); no server record.
7. Wishlist does not reserve inventory.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Read (list) | ❌ Planned | Customer's full wishlist |
| Create (add) | ❌ Planned | Add product |
| Delete (remove) | ❌ Planned | Remove by product_id |
| Check status | ❌ Planned | Is product in wishlist (optional) |

## Required APIs

### Storefront (customer-authenticated)

All require Bearer JWT with `role = customer`.

#### `GET /api/v1/store/account/wishlist`

**Query:** `page`, `per_page`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "product": {
        "name": "Classic T-Shirt",
        "slug": "classic-t-shirt",
        "price": 29.99,
        "sale_price": 24.99,
        "image_url": "/uploads/tee.jpg",
        "is_in_stock": true
      },
      "created_at": "2026-06-20T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

---

#### `POST /api/v1/store/account/wishlist`

Add product to wishlist.

**Request:**
```json
{
  "product_id": "uuid"
}
```

**Response `201`:**
```json
{
  "id": "uuid",
  "product_id": "uuid",
  "created_at": "2026-06-28T17:00:00Z"
}
```

**Response `200`:** Idempotent if already exists.

**Errors:** `404` product not found, `401` unauthorized.

---

#### `DELETE /api/v1/store/account/wishlist/{product_id}`

Remove product from wishlist by product ID.

**Response `204`:** No content.

**Errors:** `404` not in wishlist.

### Optional: batch check endpoint

#### `POST /api/v1/store/account/wishlist/check`

Check which products from a list are wishlisted.

**Request:** `{ "product_ids": ["uuid", "uuid"] }`

**Response `200`:** `{ "wishlisted": ["uuid"] }`

## Domain Reference (planned)

- Package: `internal/domain/wishlist/`
- Table: `wishlist_items`
- Migration: `000013_engagement`
