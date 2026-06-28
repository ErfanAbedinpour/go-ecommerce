# Pro Banner

## Purpose

Promotional image banners displayed on the homepage (and potentially other pages). Supports separate desktop and mobile images with optional click-through links.

## Description

Maps to `pro_banners` table. Multiple banners can exist, ordered by `sort_order`. Used for seasonal promotions, category highlights, or partner campaigns in the admin Context hub.

**Implementation status:** Not implemented. Planned in migration `000010_storefront_content`.

## Responsibilities

- Display promotional banner images on storefront
- Support responsive images (desktop + mobile)
- Link banners to internal or external destinations
- Control banner visibility and display order

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `desktop_image_url` | string | No | — | Required, valid URL, max 500 | Desktop banner image |
| `mobile_image_url` | string | Yes | `NULL` | Valid URL, max 500 | Mobile-optimized image (falls back to desktop) |
| `link_url` | string | Yes | `NULL` | Valid URL or path, max 500 | Click-through destination |
| `sort_order` | int | No | `0` | ≥ 0 | Display order |
| `is_active` | bool | No | `true` | — | Visible on storefront |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Upload | — | Images via `/admin/uploads` |
| Homepage aggregate | 1:N | Banner list in homepage API |

## Business Rules

1. At least `desktop_image_url` required; mobile falls back to desktop if null.
2. Only active banners returned on public API, sorted by `sort_order` ASC.
3. `link_url` optional — banner without link is display-only.
4. Recommended image dimensions documented in admin UI (e.g., 1920×600 desktop, 768×400 mobile).
5. Maximum banners limit: 10 (recommended validation).
6. Cache tag `homepage` invalidated on CRUD.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ❌ Planned | |
| Read (list) | ❌ Planned | All banners for admin |
| Read (single) | ❌ Planned | Optional |
| Update | ❌ Planned | |
| Delete | ❌ Planned | Hard delete |
| Reorder | ❌ Planned | Via `sort_order` on update |

## Required APIs

### Admin

All require admin JWT.

#### `GET /api/v1/admin/storefront/pro-banners`

**Query:** `is_active` (optional filter)

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "desktop_image_url": "/uploads/banner-desktop.jpg",
      "mobile_image_url": "/uploads/banner-mobile.jpg",
      "link_url": "/products?category=sale",
      "sort_order": 0,
      "is_active": true,
      "created_at": "2026-06-01T00:00:00Z",
      "updated_at": "2026-06-15T00:00:00Z"
    }
  ]
}
```

---

#### `POST /api/v1/admin/storefront/pro-banners`

**Request:**
```json
{
  "desktop_image_url": "/uploads/banner-desktop.jpg",
  "mobile_image_url": "/uploads/banner-mobile.jpg",
  "link_url": "/products?category=sale",
  "sort_order": 0,
  "is_active": true
}
```

**Response `201`:** Banner object.

---

#### `GET /api/v1/admin/storefront/pro-banners/{id}`

**Response `200`:** Banner object.

---

#### `PUT /api/v1/admin/storefront/pro-banners/{id}`

Partial update.

---

#### `DELETE /api/v1/admin/storefront/pro-banners/{id}`

**Response `204`:** No content.

### Storefront (public)

#### `GET /api/v1/store/homepage`

**Response fragment:**
```json
{
  "pro_banners": [
    {
      "id": "uuid",
      "desktop_image_url": "/uploads/banner-desktop.jpg",
      "mobile_image_url": "/uploads/banner-mobile.jpg",
      "link_url": "/products?category=sale"
    }
  ]
}
```

## Domain Reference (planned)

- Package: `internal/domain/storefront/probanner/`
- Table: `pro_banners`
