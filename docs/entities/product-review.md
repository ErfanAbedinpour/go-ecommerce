# Product Review

## Purpose

Customer-submitted reviews and ratings for individual products. Displayed on product detail pages after admin moderation. Supports both registered customers and guest reviewers.

## Description

Maps to `product_reviews` table. Status workflow: `pending` → `approved` or `rejected`. Optional link to `customer_id` for registered buyers; guests provide `author_name` only.

**Implementation status:** Not implemented. Planned in migration `000013_engagement`.

## Responsibilities

- Collect product ratings (1–5 stars) and written reviews
- Queue reviews for admin moderation
- Display approved reviews on product detail page
- Support verified-buyer badge (future: check order history)

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `product_id` | UUID | No | — | FK → `products.id` CASCADE | Reviewed product |
| `customer_id` | UUID | Yes | `NULL` | FK → `customers.id` SET NULL | Reviewer (null = guest) |
| `author_name` | string | No | — | Required, max 255 | Display name |
| `rating` | smallint | No | — | Required, integer 1–5 | Star rating |
| `title` | string | Yes | `NULL` | Max 255 | Review headline |
| `content` | text | No | — | Required, 10–2000 chars | Review body |
| `status` | enum | No | `pending` | `pending` \| `approved` \| `rejected` | Moderation state |
| `created_at` | timestamp | No | `now()` | — | Submitted at |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Product | N:1 | Reviewed product |
| Customer | N:1 | Optional reviewer account |

## Business Rules

1. New reviews default to `status = pending`.
2. Only `approved` reviews shown on public product page.
3. One review per customer per product (unique constraint recommended: `product_id + customer_id` where customer_id not null).
4. Guest reviews allowed without `customer_id`; rate limited per IP.
5. Verified buyer check: customer must have delivered order containing product (recommended for authenticated reviews).
6. Average rating computed from approved reviews only.
7. Rating must be integer 1–5.
8. HTML stripped from `content`.
9. Deleting a product cascades to its reviews.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create (customer) | ❌ Planned | Submit review |
| Read (by product, public) | ❌ Planned | Approved only |
| Read (list, admin) | ❌ Planned | Moderation queue |
| Approve | ❌ Planned | |
| Reject | ❌ Planned | |
| Delete | ❌ Planned | |

## Required APIs

### Storefront

#### `GET /api/v1/store/products/{slug}/reviews`

Public approved reviews for a product.

**Query:** `page`, `per_page`, `sort` (`newest`, `highest`, `lowest`)

**Response `200`:**
```json
{
  "summary": {
    "average_rating": 4.5,
    "total_count": 24,
    "distribution": { "5": 15, "4": 6, "3": 2, "2": 1, "1": 0 }
  },
  "data": [
    {
      "id": "uuid",
      "author_name": "Jane Doe",
      "rating": 5,
      "title": "Excellent quality",
      "content": "Love this product, fits perfectly!",
      "is_verified_buyer": true,
      "created_at": "2026-06-15T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 10, "total": 24, "total_pages": 3 }
}
```

---

#### `POST /api/v1/store/products/{id}/reviews`

Submit a review (customer-authenticated).

**Auth:** Bearer JWT, `role = customer`

**Request:**
```json
{
  "rating": 5,
  "title": "Excellent quality",
  "content": "Love this product, fits perfectly!"
}
```

`author_name` derived from customer profile.

**Response `201`:**
```json
{
  "id": "uuid",
  "status": "pending",
  "message": "Your review has been submitted and is awaiting approval."
}
```

**Errors:** `409` already reviewed, `403` not a verified buyer, `422` validation.

### Admin

All require admin JWT.

#### `GET /api/v1/admin/product-reviews`

**Query:** `page`, `per_page`, `status` (default `pending`), `product_id`, `rating`, `search`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "product_name": "Classic T-Shirt",
      "customer_id": "uuid",
      "author_name": "Jane Doe",
      "rating": 5,
      "title": "Excellent quality",
      "content": "Love this product...",
      "status": "pending",
      "created_at": "2026-06-28T18:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

---

#### `PATCH /api/v1/admin/product-reviews/{id}/approve`

**Response `200`:** `{ "id": "uuid", "status": "approved" }`

---

#### `PATCH /api/v1/admin/product-reviews/{id}/reject`

**Response `200`:** `{ "id": "uuid", "status": "rejected" }`

---

#### `DELETE /api/v1/admin/product-reviews/{id}`

**Response `204`**

## Domain Reference (planned)

- Package: `internal/domain/productreview/`
- Table: `product_reviews`
- Enum: `product_review_status`
