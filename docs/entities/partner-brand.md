# Partner Brand

## Purpose

Displays partner/sponsor brand logos on the homepage brand carousel. Distinct from catalog **Brand** entity — partner brands are marketing content, not product taxonomy.

## Description

Maps to `partner_brands` table. Each entry has a logo, title, optional description, and link. Ordered by `sort_order` for the scrolling logo strip on the storefront homepage.

**Implementation status:** Not implemented. Planned in migration `000010_storefront_content`.

## Responsibilities

- Showcase partner, sponsor, or featured brand logos
- Link logos to partner websites or internal brand pages
- Control logo visibility and display order
- Provide marketing content separate from product catalog brands

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `title` | string | No | — | Required, max 255 | Brand/partner name |
| `description` | text | Yes | `NULL` | Max 2000 | Optional description (admin/tooltip) |
| `logo_url` | string | No | — | Required, valid URL, max 500 | Logo image |
| `link_url` | string | Yes | `NULL` | Valid URL, max 500 | External or internal link |
| `sort_order` | int | No | `0` | ≥ 0 | Carousel order |
| `is_active` | bool | No | `true` | — | Visible on storefront |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Brand (catalog) | None | Independent entities |
| Homepage aggregate | 1:N | Partner logo list |

## Business Rules

1. `title` and `logo_url` required on create.
2. Only active partners shown on public homepage.
3. `link_url` opens in new tab when external (frontend concern).
4. Recommended logo format: PNG/SVG with transparent background.
5. Maximum partners: 20 (recommended validation).
6. No FK to catalog `brands` table — partner brands are content, not taxonomy.

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

All require admin JWT. Route: `/api/v1/admin/storefront/partner-brands`.

#### `GET /api/v1/admin/storefront/partner-brands`

**Query:** `is_active`, `page`, `per_page`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Nike",
      "description": "Official partner",
      "logo_url": "/uploads/partners/nike.png",
      "link_url": "https://nike.com",
      "sort_order": 0,
      "is_active": true,
      "created_at": "2026-06-01T00:00:00Z",
      "updated_at": "2026-06-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 8, "total_pages": 1 }
}
```

---

#### `POST /api/v1/admin/storefront/partner-brands`

**Request:**
```json
{
  "title": "Nike",
  "description": "Official partner",
  "logo_url": "/uploads/partners/nike.png",
  "link_url": "https://nike.com",
  "sort_order": 0,
  "is_active": true
}
```

**Response `201`:** Partner brand object.

---

#### `GET /api/v1/admin/storefront/partner-brands/{id}`

**Response `200`:** Partner brand object.

---

#### `PUT /api/v1/admin/storefront/partner-brands/{id}`

Partial update.

---

#### `DELETE /api/v1/admin/storefront/partner-brands/{id}`

**Response `204`**

### Storefront (public)

#### `GET /api/v1/store/homepage`

**Response fragment:**
```json
{
  "partner_brands": [
    {
      "id": "uuid",
      "title": "Nike",
      "logo_url": "/uploads/partners/nike.png",
      "link_url": "https://nike.com"
    }
  ]
}
```

## Domain Reference (planned)

- Package: `internal/domain/storefront/partnerbrand/`
- Table: `partner_brands`
