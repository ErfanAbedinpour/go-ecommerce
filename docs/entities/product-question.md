# Product Question

## Purpose

Product Q&A — customers ask questions on product detail pages; admins answer from the back office. Displayed in an accordion on the product page once answered.

## Description

Maps to `product_questions` table. Simpler workflow than reviews: `open` → `answered`. Questions can be submitted by anyone (email optional); answers provided by admin users.

**Implementation status:** Not implemented. Planned in migration `000013_engagement`.

## Responsibilities

- Accept customer questions about specific products
- Allow admin to post official answers
- Display answered Q&A on product detail page
- Track which admin answered and when

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `product_id` | UUID | No | — | FK → `products.id` CASCADE | Product in question |
| `asker_name` | string | No | — | Required, max 255 | Questioner display name |
| `asker_email` | string | Yes | `NULL` | Email, max 255 | Optional email for follow-up (not public) |
| `question` | text | No | — | Required, 5–1000 chars | Question text |
| `answer` | text | Yes | `NULL` | Max 2000 chars | Admin response |
| `answered_at` | timestamp | Yes | `NULL` | — | When answer was posted |
| `answered_by` | UUID | Yes | `NULL` | FK → `admin_users.id` SET NULL | Admin who answered |
| `status` | enum | No | `open` | `open` \| `answered` | Q&A state |
| `created_at` | timestamp | No | `now()` | — | Question submitted at |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Product | N:1 | Product being asked about |
| User (admin) | N:1 | Admin who answered |

## Business Rules

1. New questions default to `status = open`.
2. Setting `answer` transitions status to `answered` and sets `answered_at` + `answered_by`.
3. Only `answered` questions visible on public product page.
4. `asker_email` never exposed in public API.
5. Rate limiting: 3 questions per hour per IP per product (recommended).
6. Open questions visible in admin queue only.
7. Answer can be updated (re-answer) while keeping `answered` status.
8. Deleting a product cascades to its questions.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create (public) | ❌ Planned | Submit question |
| Read (by product, public) | ❌ Planned | Answered only |
| Read (list, admin) | ❌ Planned | Open + answered queue |
| Answer | ❌ Planned | Admin response |
| Delete | ❌ Planned | |

## Required APIs

### Storefront (public)

#### `POST /api/v1/store/products/{id}/questions`

Submit a question about a product.

**Request:**
```json
{
  "asker_name": "Curious Shopper",
  "asker_email": "shopper@example.com",
  "question": "Is this product machine washable?"
}
```

**Response `201`:**
```json
{
  "id": "uuid",
  "status": "open",
  "message": "Your question has been submitted. We will respond soon."
}
```

**Errors:** `404` product not found, `422` validation, `429` rate limited.

---

#### `GET /api/v1/store/products/{slug}/questions`

Answered Q&A for product detail page.

**Query:** `page`, `per_page`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "question": "Is this product machine washable?",
      "answer": "Yes, machine wash cold on gentle cycle.",
      "answered_at": "2026-06-20T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 10, "total": 3, "total_pages": 1 }
}
```

Note: `asker_name` omitted from public response for privacy.

### Admin

All require admin JWT.

#### `GET /api/v1/admin/product-questions`

**Query:** `page`, `per_page`, `status` (`open`, `answered`, or omit for all), `product_id`, `search`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "product_name": "Classic T-Shirt",
      "asker_name": "Curious Shopper",
      "asker_email": "shopper@example.com",
      "question": "Is this product machine washable?",
      "answer": null,
      "status": "open",
      "created_at": "2026-06-28T19:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 12, "total_pages": 1 }
}
```

---

#### `PUT /api/v1/admin/product-questions/{id}/answer`

Post or update an answer.

**Request:**
```json
{
  "answer": "Yes, machine wash cold on gentle cycle. Tumble dry low."
}
```

**Response `200`:**
```json
{
  "id": "uuid",
  "question": "Is this product machine washable?",
  "answer": "Yes, machine wash cold on gentle cycle. Tumble dry low.",
  "status": "answered",
  "answered_at": "2026-06-28T20:00:00Z",
  "answered_by": "uuid"
}
```

`answered_by` set from JWT user ID automatically.

---

#### `DELETE /api/v1/admin/product-questions/{id}`

**Response `204`**

## Domain Reference (planned)

- Package: `internal/domain/productquestion/`
- Table: `product_questions`
