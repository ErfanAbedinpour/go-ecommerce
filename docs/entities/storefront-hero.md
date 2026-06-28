# Storefront Hero

## Purpose

Configures the homepage hero section — a full-width video background with overlay title, subtitle, and call-to-action buttons. Singleton content entity edited from the admin Context hub.

## Description

Maps to `storefront_hero` table. Expected to be a single active configuration row (upsert pattern). The hero is the primary above-the-fold element on the customer storefront homepage, driving users to featured collections or promotions.

**Implementation status:** Not implemented. Planned in migration `000010_storefront_content`.

## Responsibilities

- Display promotional video background on homepage
- Present headline and subheadline overlay text
- Provide primary and secondary CTA buttons with links
- Toggle hero visibility without deleting content

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `video_url` | string | Yes | `NULL` | Valid URL, max 500, MP4 recommended | Background video URL |
| `title` | string | Yes | `NULL` | Max 255 | Overlay headline |
| `subtitle` | text | Yes | `NULL` | Max 2000 | Supporting text below title |
| `cta_primary_text` | string | Yes | `NULL` | Max 100 | Primary button label |
| `cta_primary_url` | string | Yes | `NULL` | Valid URL or path, max 500 | Primary button link |
| `cta_secondary_text` | string | Yes | `NULL` | Max 100 | Secondary button label |
| `cta_secondary_url` | string | Yes | `NULL` | Valid URL or path, max 500 | Secondary button link |
| `is_active` | bool | No | `true` | — | Show hero on homepage |
| `updated_at` | timestamp | No | `now()` | — | Last edit |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Upload | — | Video uploaded via `/admin/uploads` |
| Homepage aggregate | 1:1 | Included in `GET /store/homepage` |

## Business Rules

1. Singleton: only one hero row; upsert on save.
2. When `is_active = false`, homepage renders without hero section.
3. If `video_url` is null, fallback to static poster image (future field) or hide video.
4. CTA pairs are optional; if text is set, URL should also be set.
5. Video files uploaded via admin upload endpoint; max size per upload config.
6. Cache tag `homepage` invalidated on any hero update.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Read | ❌ Planned | Singleton GET |
| Update (upsert) | ❌ Planned | PUT replaces entire config |
| Create | — | Merged into upsert |
| Delete | — | Use `is_active = false` instead |

## Required APIs

### Admin

All require admin JWT.

#### `GET /api/v1/admin/storefront/hero`

**Response `200`:**
```json
{
  "id": "uuid",
  "video_url": "/uploads/hero-video.mp4",
  "title": "Summer Collection 2026",
  "subtitle": "Discover the latest trends",
  "cta_primary_text": "Shop Now",
  "cta_primary_url": "/products",
  "cta_secondary_text": "Learn More",
  "cta_secondary_url": "/about",
  "is_active": true,
  "updated_at": "2026-06-28T12:00:00Z"
}
```

**Response `404`:** No hero configured yet (return defaults or empty).

---

#### `PUT /api/v1/admin/storefront/hero`

Upsert hero configuration.

**Request:**
```json
{
  "video_url": "/uploads/hero-video.mp4",
  "title": "Summer Collection 2026",
  "subtitle": "Discover the latest trends",
  "cta_primary_text": "Shop Now",
  "cta_primary_url": "/products",
  "cta_secondary_text": "Learn More",
  "cta_secondary_url": "/about",
  "is_active": true
}
```

**Response `200`:** Full hero object.

**Errors:** `422` validation (invalid URL, text too long).

### Storefront (public)

Included in aggregated homepage response:

#### `GET /api/v1/store/homepage`

**Response fragment:**
```json
{
  "hero": {
    "video_url": "/uploads/hero-video.mp4",
    "title": "Summer Collection 2026",
    "subtitle": "Discover the latest trends",
    "cta_primary": { "text": "Shop Now", "url": "/products" },
    "cta_secondary": { "text": "Learn More", "url": "/about" }
  }
}
```

Only returned when `is_active = true`.

## Domain Reference (planned)

- Package: `internal/domain/storefront/hero/`
- Table: `storefront_hero`
- Migration: `000010_storefront_content`
