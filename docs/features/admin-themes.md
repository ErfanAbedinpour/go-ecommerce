# Admin Themes

**Routes:** `/themes`, `/set-style`, `/checkout/themes/*` (preview)  
**Status:** 🆕 Not implemented (proposed)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Theme marketplace lets admins browse purchasable storefront themes, acquire free/paid themes, and customize the active theme's **12 color tokens** and **font family** on `/set-style`. Themes are configuration records — the storefront applies CSS variables from API data without redeploying frontend code.

---

## User Flow

### `/themes` — Theme marketplace

1. Load theme catalog: `GET /admin/themes`.
2. Each card shows preview image, name, description, price (0 = free).
3. Admin clicks **Purchase** / **Activate** on unpurchased theme → `POST /admin/themes/{id}/purchase`.
4. Purchased themes show **Apply** → updates `store_style.active_theme_id`.
5. Link to **Customize** → `/set-style`.

### `/set-style` — Style customization

1. Load current style: `GET /admin/store-style`.
2. Display 12 color pickers + font selector.
3. Live preview updates CSS variables client-side.
4. Save → `PUT /admin/store-style`.
5. Reset to theme defaults → reload defaults from active theme's `default_colors` / `default_font`.

### `/checkout/themes/*` — Checkout previews

- **Frontend-only** static preview routes demonstrating checkout UI per theme.
- No backend API required for preview pages.

---

## Business Logic

### Theme catalog

- Themes seeded in `store_themes` (migration `000011`).
- `price = 0` → free; purchase records ownership without payment gateway (v1 mock purchase).
- Paid themes (`price > 0`): v1 records purchase without real billing; v2 integrates payment.

### Purchases

- `theme_purchases` tracks `(theme_id, purchased_by)` unique per admin user.
- Free themes auto-purchased on first apply.
- Cannot apply theme not purchased (unless free).

### Active style (`store_style` singleton)

- One row per store.
- `active_theme_id` → FK `store_themes`.
- `colors` JSONB — 12 customizable tokens (merged over theme defaults).
- `font_family` — Google Font name or system stack.

### 12 color tokens

Standard design system tokens (stored as hex strings):

| Token | Purpose |
|-------|---------|
| `primary` | Primary brand color |
| `primary_foreground` | Text on primary |
| `secondary` | Secondary surfaces |
| `secondary_foreground` | Text on secondary |
| `accent` | Accent/highlight |
| `accent_foreground` | Text on accent |
| `background` | Page background |
| `foreground` | Default text |
| `card` | Card background |
| `card_foreground` | Text on cards |
| `muted` | Muted backgrounds |
| `border` | Borders/dividers |

### Storefront consumption

- `GET /store/theme` returns active theme + merged colors + font.
- Store applies as CSS variables: `--color-primary: #…`.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Purchase already owned theme | `409` or idempotent 200 |
| Apply unpurchased paid theme | `403` or `422` |
| Invalid hex color | `400` validation |
| No active theme | Apply first free default theme |
| Delete theme with active users | Soft-disable `is_active` on theme, don't hard delete |
| Font not in allowed list | `400` or accept any string (v1: accept any max 100 chars) |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Theme service | Catalog, purchases, style |
| Admin users | `purchased_by` FK |
| Store theme API | Public CSS tokens |
| Cache | Invalidate tag `theme` on style change |

---

## Required APIs

### Proposed — GET `/api/v1/admin/themes`

**Query:** `page`, `per_page`

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Modern Minimal",
      "slug": "modern-minimal",
      "description": "Clean lines for building material stores",
      "preview_image_url": "https://cdn.example.com/themes/modern.jpg",
      "price": 0,
      "is_active": true,
      "default_colors": {
        "primary": "#2563eb",
        "primary_foreground": "#ffffff",
        "secondary": "#f1f5f9",
        "secondary_foreground": "#0f172a",
        "accent": "#f59e0b",
        "accent_foreground": "#ffffff",
        "background": "#ffffff",
        "foreground": "#0f172a",
        "card": "#ffffff",
        "card_foreground": "#0f172a",
        "muted": "#f8fafc",
        "border": "#e2e8f0"
      },
      "default_font": "Inter",
      "is_purchased": true,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 6, "total_pages": 1 }
}
```

**Errors:** `401`, `403`

---

### Proposed — POST `/api/v1/admin/themes/{id}/purchase`

**Request:** Empty body (v1 mock purchase).

**Response 201:**

```json
{
  "theme_id": "uuid",
  "purchased_at": "2026-06-01T12:00:00Z"
}
```

**Errors:** `404`, `409` (already purchased), `401`, `403`

---

### Proposed — GET `/api/v1/admin/store-style`

**Response 200:**

```json
{
  "id": "uuid",
  "active_theme_id": "uuid",
  "active_theme": {
    "id": "uuid",
    "name": "Modern Minimal",
    "slug": "modern-minimal",
    "preview_image_url": "https://…"
  },
  "colors": {
    "primary": "#2563eb",
    "primary_foreground": "#ffffff",
    "secondary": "#f1f5f9",
    "secondary_foreground": "#0f172a",
    "accent": "#f59e0b",
    "accent_foreground": "#ffffff",
    "background": "#ffffff",
    "foreground": "#0f172a",
    "card": "#ffffff",
    "card_foreground": "#0f172a",
    "muted": "#f8fafc",
    "border": "#e2e8f0"
  },
  "font_family": "Inter",
  "updated_at": "2026-06-01T10:00:00Z"
}
```

**Errors:** `401`, `403`

---

### Proposed — PUT `/api/v1/admin/store-style`

**Request:**

```json
{
  "active_theme_id": "uuid",
  "colors": {
    "primary": "#1d4ed8",
    "primary_foreground": "#ffffff",
    "secondary": "#f1f5f9",
    "secondary_foreground": "#0f172a",
    "accent": "#ea580c",
    "accent_foreground": "#ffffff",
    "background": "#fafafa",
    "foreground": "#171717",
    "card": "#ffffff",
    "card_foreground": "#171717",
    "muted": "#f5f5f5",
    "border": "#d4d4d4"
  },
  "font_family": "Vazirmatn"
}
```

**Response 200:** Updated style object.

**Errors:** `400`, `403` (theme not purchased), `404`, `401`

---

### Public (storefront) — GET `/api/v1/store/theme`

**Auth:** None

**Response 200:** Same shape as `store-style` (read-only).

---

## Database Impact

**Tables:** `store_themes`, `theme_purchases`, `store_style` (migration `000011`)

| Table | Notes |
|-------|-------|
| `store_themes` | Catalog + default tokens |
| `theme_purchases` | UNIQUE (theme_id, purchased_by) |
| `store_style` | Singleton; JSONB `colors` |

**Seed:** 4–6 default themes on migration.

**Cache:** Invalidate `theme` on PUT store-style or purchase.

---

## UI Changes Affecting Backend

| UI element | Backend impact |
|------------|----------------|
| Color pickers (×12) | Full `colors` object on PUT |
| Font dropdown | `font_family` string |
| Theme preview cards | `preview_image_url` from catalog |
| Purchase button | POST purchase; refresh `is_purchased` |
| Live preview | Client-side CSS vars; no API until save |
| Checkout theme previews | Static routes; no API |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `active_theme_id` | Valid UUID; theme exists and purchased (or free) |
| `colors.*` | Valid hex `#RGB` or `#RRGGBB`; all 12 keys required on full update |
| `font_family` | Max 100 chars |
| Theme `price` | `>= 0` decimal |

---

## Permission Requirements

| Action | Role |
|--------|------|
| Browse/purchase/apply themes | `admin` |
| Public store theme read | None |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Theme catalog | React Query | `/themes` |
| Live preview colors/font | React state (ephemeral) | `/set-style` |
| Saved style | React Query | `/set-style` |
| Purchase loading per card | React state | `/themes` |
| CSS variable injection | DOM / preview iframe | Preview |
