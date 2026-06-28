# Store Style

## Purpose

Active storefront visual customization — the currently applied theme plus user overrides for colors and font. Singleton configuration edited from admin `/set-style` UI and consumed by the storefront as CSS variables.

## Description

Maps to `store_style` table (singleton row). References an optional `active_theme_id` from **StoreTheme**. Custom `colors` JSONB overrides theme defaults; `font_family` overrides theme default font.

**Implementation status:** Not implemented. Planned in migration `000011_theme_system`.

## Responsibilities

- Track currently active theme
- Store customized color palette (12 tokens)
- Store customized font family
- Provide design tokens to storefront frontend
- Reset to theme defaults on theme change

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key (singleton) |
| `active_theme_id` | UUID | Yes | `NULL` | FK → `store_themes.id` SET NULL | Currently active theme |
| `colors` | JSONB | No | `{}` | 12 color token keys | Custom color overrides |
| `font_family` | string | Yes | `NULL` | Max 100 | Custom font family |
| `updated_at` | timestamp | No | `now()` | — | Last style change |

### Color Tokens (same as StoreTheme)

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

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| StoreTheme | N:1 | Active theme source for defaults |

## Business Rules

1. Singleton row — upsert pattern like store settings.
2. On theme activation: copy theme `default_colors` and `default_font` to style row.
3. Color values must be valid CSS colors (hex, rgb, hsl).
4. Partial color updates merge into existing `colors` JSONB.
5. `font_family` must be from allowed list or Google Fonts catalog (validation).
6. Storefront applies tokens as CSS custom properties: `--color-primary`, etc.
7. Cache tag `theme` invalidated on any style update.
8. Preview mode (checkout theme previews) uses read-only theme defaults without saving.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Read | ❌ Planned | Current style + resolved tokens |
| Update | ❌ Planned | Colors and/or font |
| Activate theme | ❌ Planned | Sets theme + resets defaults |
| Reset to theme defaults | ❌ Planned | Clears overrides |

## Required APIs

### Admin

All require admin JWT.

#### `GET /api/v1/admin/store-style`

**Response `200`:**
```json
{
  "id": "uuid",
  "active_theme": {
    "id": "uuid",
    "name": "Modern Minimal",
    "slug": "modern-minimal"
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
    "muted": "#f8fafc",
    "muted_foreground": "#64748b",
    "border": "#e2e8f0",
    "destructive": "#ef4444"
  },
  "font_family": "Inter",
  "updated_at": "2026-06-28T16:00:00Z"
}
```

Resolved colors merge theme defaults with overrides (overrides win).

---

#### `PUT /api/v1/admin/store-style`

Update color overrides and/or font. Does not change active theme.

**Request:**
```json
{
  "colors": {
    "primary": "#7c3aed",
    "accent": "#f59e0b"
  },
  "font_family": "Poppins"
}
```

**Response `200`:** Full resolved style object.

---

#### `PUT /api/v1/admin/store-style/activate-theme`

Activate a purchased theme and reset to its defaults.

**Request:**
```json
{
  "theme_id": "uuid"
}
```

**Response `200`:** Full style object with theme defaults applied.

**Errors:** `403` theme not purchased, `404` theme not found.

---

#### `POST /api/v1/admin/store-style/reset`

Reset colors and font to active theme defaults.

**Response `200`:** Full style object.

### Storefront (public)

#### `GET /api/v1/store/theme`

Public design tokens for CSS variable injection.

**Response `200`:**
```json
{
  "colors": {
    "primary": "#7c3aed",
    "background": "#ffffff",
    "foreground": "#0f172a"
  },
  "font_family": "Poppins"
}
```

No theme metadata exposed publicly — only resolved tokens.

## Domain Reference (planned)

- Package: `internal/domain/theme/style/`
- Table: `store_style`
- Migration: `000011_theme_system`
