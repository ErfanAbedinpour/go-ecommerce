# Admin Context — Customer Reviews (Homepage Testimonials)

**Route:** `/context/customer-reviews`  
**Status:** 🆕 Not implemented (proposed)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Manage curated customer testimonials displayed on the storefront homepage. These are **marketing testimonials** (admin-authored), distinct from product-level reviews (`product_reviews`) submitted by buyers on product detail pages.

Each testimonial includes customer name, optional photo, review text, optional star rating, sort order, and active flag.

---

## User Flow

1. Admin opens **Context → Customer Reviews** (`/context/customer-reviews`).
2. List loads: `GET /admin/storefront/homepage-reviews`.
3. **Add review** → form: name, photo upload, text, rating (1–5), active toggle.
4. Upload photo → `POST /admin/uploads?context=review`.
5. Save → `POST /admin/storefront/homepage-reviews`.
6. Edit, delete, reorder entries.
7. Homepage carousel/grid reads active reviews from storefront API.

---

## Business Logic

- Admin-curated content; not tied to real customer accounts or orders.
- Displayed as testimonial carousel on homepage (Persian RTL store).
- Sorted by `sort_order`; inactive hidden on storefront.
- `rating` optional — if null, UI hides stars.
- `photo_url` optional — initials avatar fallback on storefront.
- Recommended max 12 active testimonials for carousel performance.

**Distinction from product reviews:**

| Aspect | Homepage testimonials | Product reviews |
|--------|----------------------|-----------------|
| Table | `homepage_reviews` | `product_reviews` |
| Admin route | `/context/customer-reviews` | Future moderation UI |
| Tied to product | No | Yes |
| Moderation | Direct publish via `is_active` | pending/approved/rejected |

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| No photo | Storefront shows initials or generic avatar |
| Rating out of range | `400` validation |
| Empty review text | `400` required |
| All reviews inactive | Homepage section hidden |
| Duplicate customer names | Allowed |
| Very long review text | Max 2000 chars server-side |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Upload service | Customer photos |
| Storefront content | Review CRUD |
| Store homepage API | Public read |

---

## Required APIs

### Proposed — GET `/api/v1/admin/storefront/homepage-reviews`

**Query:** `page`, `per_page`, `is_active`

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "customer_name": "Ali Rezaei",
      "photo_url": "https://cdn.example.com/reviews/ali.jpg",
      "review_text": "Excellent quality tiles and fast delivery!",
      "rating": 5,
      "sort_order": 0,
      "is_active": true,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 6, "total_pages": 1 }
}
```

**Errors:** `401`, `403`

---

### Proposed — POST `/api/v1/admin/storefront/homepage-reviews`

**Request:**

```json
{
  "customer_name": "Ali Rezaei",
  "photo_url": "https://cdn.example.com/reviews/ali.jpg",
  "review_text": "Excellent quality tiles and fast delivery!",
  "rating": 5,
  "sort_order": 0,
  "is_active": true
}
```

**Response 201:** Created review.

**Errors:** `400`, `401`, `403`

---

### Proposed — GET `/api/v1/admin/storefront/homepage-reviews/{id}`

**Response 200:** Single review.

**Errors:** `404`, `401`, `403`

---

### Proposed — PUT `/api/v1/admin/storefront/homepage-reviews/{id}`

**Request:** Partial update supported.

**Response 200:** Updated review.

**Errors:** `400`, `404`, `401`, `403`

---

### Proposed — DELETE `/api/v1/admin/storefront/homepage-reviews/{id}`

**Response 204**

**Errors:** `404`, `401`, `403`

---

## Database Impact

**Table:** `homepage_reviews` (migration `000010`)

| Column | Notes |
|--------|-------|
| `customer_name` | VARCHAR(255) NOT NULL |
| `photo_url` | VARCHAR(500) nullable |
| `review_text` | TEXT NOT NULL |
| `rating` | SMALLINT 1–5 nullable |
| `sort_order`, `is_active` | Display control |

**Cache:** Invalidate `homepage`.

---

## UI Changes Affecting Backend

| UI element | Backend impact |
|------------|----------------|
| Star rating picker | `rating` integer 1–5 or null |
| Photo upload | Optional `photo_url` |
| Carousel preview | Frontend-only using same data |
| Character counter on review | Enforce max 2000 server-side |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `customer_name` | Required; max 255 chars |
| `photo_url` | Optional; valid URL; max 500 |
| `review_text` | Required; min 10; max 2000 chars |
| `rating` | Optional; integer 1–5 |
| `sort_order` | Integer `>= 0` |
| `is_active` | Boolean |
| Photo upload | Image; max 5 MB |

---

## Permission Requirements

| Action | Role |
|--------|------|
| CRUD homepage reviews | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Review list | React Query | Page |
| Create/edit modal | React state | Modal |
| Photo preview | Blob URL → CDN URL | Form |
| Reorder draft | React state | List |
