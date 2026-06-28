# Homepage Review

## Purpose

Curated customer testimonials displayed on the homepage. Admin-managed social proof content with optional photos and star ratings — distinct from **ProductReview** (per-product reviews).

## Description

Maps to `homepage_reviews` table. Each review is a manually entered testimonial (not automatically sourced from product reviews). Displayed in a carousel or grid on the storefront homepage.

**Implementation status:** Not implemented. Planned in migration `000010_storefront_content`.

## Responsibilities

- Display customer testimonials on homepage
- Support optional customer photo and star rating
- Control testimonial visibility and order
- Provide marketing social proof independent of product reviews

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `customer_name` | string | No | — | Required, max 255 | Reviewer display name |
| `photo_url` | string | Yes | `NULL` | Valid URL, max 500 | Reviewer avatar/photo |
| `review_text` | text | No | — | Required, 10–2000 chars | Testimonial content |
| `rating` | smallint | Yes | `NULL` | 1–5 if set | Star rating |
| `sort_order` | int | No | `0` | ≥ 0 | Display order |
| `is_active` | bool | No | `true` | — | Visible on homepage |
| `created_at` | timestamp | No | `now()` | — | Created |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Customer | None | No FK; manually entered names |
| ProductReview | None | Separate review system |
| Homepage aggregate | 1:N | Testimonial list |

## Business Rules

1. `customer_name` and `review_text` required.
2. `rating` optional but if set must be integer 1–5.
3. Only active reviews returned on public API, sorted by `sort_order`.
4. Maximum active reviews: 12 (recommended).
5. Reviews are curated content — not submitted by customers directly.
6. Cache tag `homepage` invalidated on CRUD.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ❌ Planned | |
| Read (list) | ❌ Planned | |
| Read (single) | ❌ Planned | |
| Update | ❌ Planned | |
| Delete | ❌ Planned | |

## Required APIs

### Admin

All require admin JWT. Route: `/api/v1/admin/storefront/homepage-reviews`.

#### `GET /api/v1/admin/storefront/homepage-reviews`

**Query:** `is_active`, `page`, `per_page`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "customer_name": "Sarah Johnson",
      "photo_url": "/uploads/reviews/sarah.jpg",
      "review_text": "Amazing quality and fast shipping!",
      "rating": 5,
      "sort_order": 0,
      "is_active": true,
      "created_at": "2026-06-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 6, "total_pages": 1 }
}
```

---

#### `POST /api/v1/admin/storefront/homepage-reviews`

**Request:**
```json
{
  "customer_name": "Sarah Johnson",
  "photo_url": "/uploads/reviews/sarah.jpg",
  "review_text": "Amazing quality and fast shipping!",
  "rating": 5,
  "sort_order": 0,
  "is_active": true
}
```

**Response `201`:** Review object.

---

#### `GET /api/v1/admin/storefront/homepage-reviews/{id}`

**Response `200`:** Review object.

---

#### `PUT /api/v1/admin/storefront/homepage-reviews/{id}`

Partial update.

---

#### `DELETE /api/v1/admin/storefront/homepage-reviews/{id}`

**Response `204`**

### Storefront (public)

#### `GET /api/v1/store/homepage`

**Response fragment:**
```json
{
  "homepage_reviews": [
    {
      "id": "uuid",
      "customer_name": "Sarah Johnson",
      "photo_url": "/uploads/reviews/sarah.jpg",
      "review_text": "Amazing quality and fast shipping!",
      "rating": 5
    }
  ]
}
```

## Domain Reference (planned)

- Package: `internal/domain/storefront/homepagereview/`
- Table: `homepage_reviews`
