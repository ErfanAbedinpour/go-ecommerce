# Store Theme

## Purpose

Theme marketplace catalog — predefined visual themes that admins can browse, preview, and purchase (including free themes). Each theme ships with default color tokens and font settings applied to **StoreStyle** on activation.

## Description

Maps to `store_themes` and `theme_purchases` tables. Themes have pricing (`price = 0` for free), preview images, and default design tokens stored as JSONB. Purchases tracked per admin user.

**Implementation status:** Not implemented. Planned in migration `000011_theme_system`.

## Responsibilities

- Catalog available storefront themes
- Define default color palette and font per theme
- Track theme purchases per admin
- Gate theme activation to purchased (or free) themes
- Support theme marketplace UI (`/themes`)

## Attributes

### StoreTheme

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `name` | string | No | — | Required, max 255 | Theme display name |
| `slug` | string | No | auto-generated | Unique, max 255, URL-safe | Theme identifier |
| `description` | text | Yes | `NULL` | Max 2000 | Theme description |
| `preview_image_url` | string | Yes | `NULL` | Valid URL, max 500 | Marketplace preview image |
| `price` | decimal | No | `0` | ≥ 0 | Price in store currency; 0 = free |
| `is_active` | bool | No | `true` | — | Listed in marketplace |
| `default_colors` | JSONB | No | `{}` | 12 color token keys | Default palette |
| `default_font` | string | Yes | `NULL` | Max 100 | Default font family name |
| `created_at` | timestamp | No | `now()` | — | Created |

### Default Color Tokens (JSONB keys)

| Token | Description |
|-------|-------------|
| `primary` | Primary brand color |
| `primary_foreground` | Text on primary |
| `secondary` | Secondary color |
| `secondary_foreground` | Text on secondary |
| `accent` | Accent/highlight |
| `accent_foreground` | Text on accent |
| `background` | Page background |
| `foreground` | Default text |
| `muted` | Muted backgrounds |
| `muted_foreground` | Muted text |
| `border` | Border color |
| `destructive` | Error/danger color |

### ThemePurchase

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `theme_id` | UUID | No | — | FK → `store_themes.id` | Purchased theme |
| `purchased_by` | UUID | No | — | FK → `admin_users.id` | Admin who purchased |
| `purchased_at` | timestamp | No | `now()` | — | Purchase timestamp |

**Unique constraint:** `(theme_id, purchased_by)`

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| StoreStyle | 1:1 active | `active_theme_id` on store_style |
| User (admin) | N:M via purchases | Purchase records |
| ThemePurchase | 1:N | Purchase history |

## Business Rules

1. Free themes (`price = 0`) auto-grant purchase on "Get" action.
2. Paid themes require purchase before activation.
3. Only one theme active at a time (via `store_style.active_theme_id`).
4. Activating a theme copies `default_colors` and `default_font` to `store_style` (user can override later).
5. Inactive themes hidden from marketplace but existing activations remain until changed.
6. Seed migration includes 3+ default themes (minimal, modern, bold).
7. Cache tag `theme` invalidated on purchase or activation.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Read (list) | ❌ Planned | Marketplace catalog |
| Read (single) | ❌ Planned | Theme detail + preview |
| Purchase | ❌ Planned | Record purchase |
| Create/Update/Delete themes | ❌ Future | Seeded; super-admin only in v2 |

## Required APIs

### Admin

All require admin JWT.

#### `GET /api/v1/admin/themes`

List available themes with purchase status for current admin.

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Modern Minimal",
      "slug": "modern-minimal",
      "description": "Clean and contemporary design",
      "preview_image_url": "/themes/modern-minimal/preview.jpg",
      "price": 0,
      "is_active": true,
      "default_colors": {
        "primary": "#2563eb",
        "background": "#ffffff",
        "foreground": "#0f172a"
      },
      "default_font": "Inter",
      "is_purchased": true,
      "is_active_theme": true,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

---

#### `GET /api/v1/admin/themes/{id}`

**Response `200`:** Full theme with all 12 color tokens and preview URLs.

---

#### `POST /api/v1/admin/themes/{id}/purchase`

Purchase or claim a free theme.

**Response `200`:**
```json
{
  "theme_id": "uuid",
  "purchased_at": "2026-06-28T16:00:00Z",
  "message": "Theme purchased successfully"
}
```

**Response `200` (free):**
```json
{
  "theme_id": "uuid",
  "purchased_at": "2026-06-28T16:00:00Z",
  "message": "Free theme added to your library"
}
```

**Errors:** `409` already purchased, `402` payment required (future).

### Storefront (public)

Theme tokens exposed via store style endpoint (not full marketplace).

## Domain Reference (planned)

- Package: `internal/domain/theme/`
- Tables: `store_themes`, `theme_purchases`
- Migration: `000011_theme_system`
