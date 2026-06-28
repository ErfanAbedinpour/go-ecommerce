# Store Settings

## Purpose

Central configuration aggregate for storefront identity, contact information, social links, SEO metadata, and navigation menus. Stored as a singleton document with JSONB sections.

## Description

The `StoreSettings` aggregate (`internal/domain/settings/entity.go`) maps to a single-row `store_settings` table (fixed UUID `f0000000-0000-0000-0000-000000000001`). Settings are split into logical sections exposed as separate API endpoints: Site, Contact, Social, SEO, and Navigation.

**Implementation status:** All admin section endpoints implemented. Storefront read API and storefront navigation split planned.

## Responsibilities

- Define store name, URL, logo, and favicon
- Publish contact information for footer and about pages
- Configure social media profile links
- Manage SEO meta tags, robots.txt, analytics, and sitemap toggle
- Maintain admin and (planned) storefront navigation menus

## Attributes

### StoreSettings (root)

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | fixed singleton | — | Single settings row |
| `site` | JSONB | No | `{}` | Site schema | Identity settings |
| `contact` | JSONB | No | `{}` | Contact schema | Public contact info |
| `social` | JSONB | No | `{}` | Social schema | Social links |
| `seo` | JSONB | No | `{}` | SEO schema | Search engine config |
| `navigation` | JSONB | No | `[]` | NavItem array | Menu tree |
| `storefront_navigation` | JSONB | Yes | `NULL` | **Planned** NavItem array | Separate store nav |
| `created_at` | timestamp | No | `now()` | — | — |
| `updated_at` | timestamp | No | `now()` | — | Last section update |

### Site

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `name` | string | No | — | Required, max 200 | Store display name |
| `url` | string | No | — | Required, valid URL, max 500 | Canonical store URL |
| `logo_url` | string | Yes | `""` | Valid URL, max 500 | Logo image |
| `favicon_url` | string | Yes | `""` | Valid URL, max 500 | Favicon |

### Contact

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `email` | string | Yes | `""` | Email, max 255 | Public email |
| `phone` | string | Yes | `""` | Max 30 | Public phone |
| `address` | string | Yes | `""` | Max 500 | Street address |
| `city` | string | Yes | `""` | Max 100 | City |
| `country` | string | Yes | `""` | Max 100 | Country |

### Social

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `facebook` | string | Yes | `""` | URL, max 500 | Facebook profile URL |
| `twitter` | string | Yes | `""` | URL, max 500 | Twitter/X URL |
| `instagram` | string | Yes | `""` | URL, max 500 | Instagram URL |
| `linkedin` | string | Yes | `""` | URL, max 500 | LinkedIn URL |
| `youtube` | string | Yes | `""` | URL, max 500 | YouTube URL |
| `tiktok` | string | Yes | `""` | URL, max 500 | TikTok URL |

### SEO

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `meta_title` | string | Yes | `""` | Max 200 | Default page title |
| `meta_description` | string | Yes | `""` | Max 500 | Default meta description |
| `meta_keywords` | string | Yes | `""` | Max 500 | Keywords (legacy) |
| `og_image_url` | string | Yes | `""` | URL, max 500 | Open Graph image |
| `robots_txt` | string | Yes | `""` | Max 5000 | robots.txt content |
| `google_analytics_id` | string | Yes | `""` | Max 100 | GA tracking ID |
| `sitemap_enabled` | bool | No | `false` | — | Enable sitemap generation job |

### NavItem

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | string | No | auto-generated | Max 100 | Client-stable item ID |
| `label` | string | No | — | Required, max 200 | Menu label |
| `url` | string | No | — | Required, max 500 | Link target |
| `sort_order` | int | No | `0` | ≥ 0 | Display order |
| `is_active` | bool | No | `true` | — | Visible when true |
| `children` | NavItem[] | No | `[]` | Nested, same schema | Sub-menu items |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Order (invoice) | — | Invoice pulls store name/contact from settings |
| Storefront pages | — | Footer, about, meta tags consume settings |

## Business Rules

1. Only one settings row exists; upsert by fixed ID.
2. Each section update touches only its JSONB column and `updated_at`.
3. Navigation replace is full-tree replacement on PUT (not merge).
4. Inactive nav items and their children are excluded from public rendering.
5. `sitemap_enabled` triggers daily `GenerateSitemap` job when true (planned).
6. Admin navigation and storefront navigation should be separate (planned split).

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Read site | ✅ Implemented | |
| Update site | ✅ Implemented | |
| Read contact | ✅ Implemented | |
| Update contact | ✅ Implemented | |
| Read social | ✅ Implemented | |
| Update social | ✅ Implemented | |
| Read SEO | ✅ Implemented | |
| Update SEO | ✅ Implemented | |
| Read navigation | ✅ Implemented | Admin nav |
| Update navigation | ✅ Implemented | Admin nav |
| Read storefront nav | ❌ Planned | Separate key or endpoint |
| Public read (aggregated) | ❌ Planned | `GET /store/settings` |

## Required APIs

### Admin (implemented)

All require admin JWT.

#### `GET /api/v1/admin/settings/site`

**Response `200`:**
```json
{
  "name": "My Store",
  "url": "https://mystore.com",
  "logo_url": "/uploads/logo.png",
  "favicon_url": "/uploads/favicon.ico"
}
```

#### `PUT /api/v1/admin/settings/site`

**Request:** Same shape as response. `name` and `url` required.

**Response `200`:** Updated `SiteSettingsResponse`

---

#### `GET /api/v1/admin/settings/contact`

**Response `200`:**
```json
{
  "email": "hello@mystore.com",
  "phone": "+98 21 1234 5678",
  "address": "123 Bazaar St",
  "city": "Tehran",
  "country": "Iran"
}
```

#### `PUT /api/v1/admin/settings/contact`

**Response `200`:** `ContactSettingsResponse`

---

#### `GET /api/v1/admin/settings/social`

**Response `200`:** `SocialSettingsResponse` with platform URLs.

#### `PUT /api/v1/admin/settings/social`

**Response `200`:** `SocialSettingsResponse`

---

#### `GET /api/v1/admin/settings/seo`

**Response `200`:**
```json
{
  "meta_title": "My Store — Best Products Online",
  "meta_description": "Shop the best products...",
  "meta_keywords": "ecommerce, shop",
  "og_image_url": "/uploads/og.jpg",
  "robots_txt": "User-agent: *\nAllow: /",
  "google_analytics_id": "G-XXXXXXXXXX",
  "sitemap_enabled": true
}
```

#### `PUT /api/v1/admin/settings/seo`

**Response `200`:** `SEOSettingsResponse`

---

#### `GET /api/v1/admin/navigation`

**Response `200`:**
```json
{
  "items": [
    {
      "id": "nav-1",
      "label": "Dashboard",
      "url": "/",
      "sort_order": 0,
      "is_active": true,
      "children": []
    }
  ]
}
```

#### `PUT /api/v1/admin/navigation`

**Request:** `{ "items": [ ... NavItemRequest[] ] }`

**Response `200`:** `NavigationResponse`

### Storefront (planned)

#### `GET /api/v1/store/settings`

Public aggregated settings (site, contact, social, seo — no admin nav).

#### `GET /api/v1/admin/storefront/navigation` / `PUT`

Separate storefront menu management (`/context/navigation` UI).

#### `GET /api/v1/admin/storefront/contact-section` / `PUT`

Contact page hero image (may extend settings or separate entity).

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| `storefront_navigation` JSONB key | High | Distinguish admin vs store menus |
| Public settings endpoint | High | Storefront footer/header |
| Contact section image | Medium | For `/context/contact-us` UI |
| `updated_at` in responses | Low | Expose per-section or global timestamp |

## Domain Reference

- Entity: `internal/domain/settings/entity.go`
- Table: `store_settings` (migration `000005_store_settings.up.sql`)
