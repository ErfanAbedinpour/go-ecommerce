# Admin Settings

**Routes:** `/general-setting`, `/navigation`, `/setting-seo`  
**Status:** ✅ Backend implemented  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Configure global storefront settings across three admin pages:

- **General** (`/general-setting`) — site identity, contact info, social links
- **Navigation** (`/navigation`) — storefront header menu tree
- **SEO** (`/setting-seo`) — meta tags, analytics, robots, sitemap toggle

Settings are stored in the `store_settings` JSONB structure (or equivalent) and exposed publicly via future `GET /store/settings`.

---

## User Flow

### `/general-setting`

Three sections/tabs on one page:

1. **Site info** — name, URL, logo, favicon  
   - Load: `GET /admin/settings/site`  
   - Save: `PUT /admin/settings/site`  
   - Logo upload: `POST /admin/uploads`

2. **Contact info** — email, phone, address, city, country  
   - `GET/PUT /admin/settings/contact`

3. **Social links** — Facebook, Twitter/X, Instagram, LinkedIn, YouTube, TikTok  
   - `GET/PUT /admin/settings/social`

### `/navigation`

1. Load menu tree: `GET /admin/navigation`.
2. Add/edit/reorder/nest menu items (drag-and-drop).
3. Save entire tree: `PUT /admin/navigation`.

**Note:** Storefront navigation may overlap with `/context/navigation` in UI — backend currently uses single navigation API for storefront menu. Admin panel sidebar nav is frontend-static.

### `/setting-seo`

1. Load: `GET /admin/settings/seo`.
2. Edit meta title/description/keywords, OG image, robots.txt, GA ID, sitemap toggle.
3. Save: `PUT /admin/settings/seo`.

---

## Business Logic

### Site settings

- `name` displayed in admin header, invoices, storefront title.
- `logo_url`, `favicon_url` from upload service URLs.
- `url` canonical store URL for SEO and invoice links.

### Contact settings

- Used on invoice (`GET /orders/{id}/invoice`), storefront footer, contact page.
- All fields optional except business requirements enforced in UI.

### Social settings

- All optional URLs; storefront renders icons only for non-empty links.

### Navigation

- Full tree replace on PUT (not patch per item).
- Max nesting depth: 2 levels recommended (validate in API).
- `is_active = false` hides item on storefront.
- Each item: `label`, `url`, `sort_order`, optional `children[]`.

### SEO settings

- `sitemap_enabled` triggers daily sitemap job (future Milestone 7).
- `robots_txt` served at `/robots.txt` on storefront (future).
- `google_analytics_id` injected in storefront layout.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Empty navigation | Storefront shows minimal nav |
| Circular nav nesting | Reject depth > 2 or validate tree |
| Invalid social URL | `400` validation |
| Missing logo | Storefront shows text name fallback |
| Duplicate sort orders | Accept; sort stable by sort_order then label |
| SEO title empty | Storefront falls back to site name |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Settings repository | JSONB persistence |
| Upload service | Logo, favicon, OG image |
| Order invoice | Site + contact on invoice |
| Store settings API (future) | Public read |
| Sitemap job (future) | `sitemap_enabled` flag |

---

## Required APIs

All require Bearer token + `admin` role.

### Site settings

#### GET `/api/v1/admin/settings/site`

**Response 200:**

```json
{
  "name": "My Shop",
  "url": "https://shop.example.com",
  "logo_url": "https://cdn.example.com/logo.png",
  "favicon_url": "https://cdn.example.com/favicon.ico"
}
```

#### PUT `/api/v1/admin/settings/site`

**Request:**

```json
{
  "name": "My Shop",
  "url": "https://shop.example.com",
  "logo_url": "https://cdn.example.com/logo.png",
  "favicon_url": "https://cdn.example.com/favicon.ico"
}
```

**Response 200:** Updated object.

**Errors:** `400`, `401`, `403`

---

### Contact settings

#### GET `/api/v1/admin/settings/contact`

**Response 200:**

```json
{
  "email": "info@shop.com",
  "phone": "+989121234567",
  "address": "123 Store St",
  "city": "Tehran",
  "country": "Iran"
}
```

#### PUT `/api/v1/admin/settings/contact`

**Request:** Same shape (all fields optional in update).

**Errors:** `400`, `401`, `403`

---

### Social settings

#### GET `/api/v1/admin/settings/social`

**Response 200:**

```json
{
  "facebook": "https://facebook.com/myshop",
  "twitter": "https://twitter.com/myshop",
  "instagram": "https://instagram.com/myshop",
  "linkedin": "",
  "youtube": "",
  "tiktok": ""
}
```

#### PUT `/api/v1/admin/settings/social`

**Request:** Same shape; URL validation per field.

**Errors:** `400`, `401`, `403`

---

### Navigation

#### GET `/api/v1/admin/navigation`

**Response 200:**

```json
{
  "items": [
    {
      "id": "nav-home",
      "label": "Home",
      "url": "/",
      "sort_order": 0,
      "is_active": true,
      "children": []
    },
    {
      "id": "nav-products",
      "label": "Products",
      "url": "/products",
      "sort_order": 1,
      "is_active": true,
      "children": [
        {
          "id": "nav-tiles",
          "label": "Tiles",
          "url": "/products?category=tiles",
          "sort_order": 0,
          "is_active": true,
          "children": []
        }
      ]
    }
  ]
}
```

#### PUT `/api/v1/admin/navigation`

**Request:**

```json
{
  "items": [
    {
      "id": "nav-home",
      "label": "Home",
      "url": "/",
      "sort_order": 0,
      "is_active": true,
      "children": []
    }
  ]
}
```

**Response 200:** Updated navigation tree.

**Errors:** `400`, `401`, `403`

---

### SEO settings

#### GET `/api/v1/admin/settings/seo`

**Response 200:**

```json
{
  "meta_title": "My Shop — Building Materials",
  "meta_description": "Premium tiles, cement, and tools",
  "meta_keywords": "tiles, cement, building materials",
  "og_image_url": "https://cdn.example.com/og.jpg",
  "robots_txt": "User-agent: *\nAllow: /",
  "google_analytics_id": "G-XXXXXXXXXX",
  "sitemap_enabled": true
}
```

#### PUT `/api/v1/admin/settings/seo`

**Request:** Same shape.

**Errors:** `400`, `401`, `403`

---

### Uploads (shared)

#### POST `/api/v1/admin/uploads`

Used for logo, favicon, OG image.

**Response 200:**

```json
{
  "url": "/uploads/settings/uuid.png",
  "filename": "logo.png",
  "size": 102400,
  "content_type": "image/png"
}
```

---

## Database Impact

**Table:** `store_settings` (existing)

Settings stored as keyed JSONB sections:

| Key | Content |
|-----|---------|
| `site` | Name, URL, logos |
| `contact` | Public contact info |
| `social` | Social URLs |
| `seo` | Meta + analytics |
| `navigation` | Menu tree |

**Operations:** UPSERT per section on PUT.

**Future:** Split `storefront_navigation` from admin if `/context/navigation` diverges.

---

## UI Changes Affecting Backend

| UI page | Backend section |
|---------|-----------------|
| General → Site tab | `/settings/site` |
| General → Contact tab | `/settings/contact` |
| General → Social tab | `/settings/social` |
| Navigation drag-drop tree | Full `items[]` on PUT |
| SEO form | `/settings/seo` |
| Logo/favicon pickers | Upload then PUT URL |

---

## Validation Requirements

### Site

| Field | Rule |
|-------|------|
| `name` | Required; max 200 |
| `url` | Required; valid URL; max 500 |
| `logo_url`, `favicon_url` | Optional URL; max 500 |

### Contact

| Field | Rule |
|-------|------|
| `email` | Optional; valid email |
| `phone` | Max 30 |
| `address` | Max 500 |
| `city`, `country` | Max 100 |

### Social

Each platform URL: optional, valid URL, max 500.

### Navigation

| Field | Rule |
|-------|------|
| `label` | Required; max 200 |
| `url` | Required; max 500 |
| `sort_order` | `>= 0` |
| Tree depth | Max 2 levels |
| Max items | 50 total |

### SEO

| Field | Rule |
|-------|------|
| `meta_title` | Max 200 |
| `meta_description` | Max 500 |
| `meta_keywords` | Max 500 |
| `og_image_url` | Optional URL |
| `robots_txt` | Max 5000 |
| `google_analytics_id` | Max 100 |

---

## Permission Requirements

| Action | Role |
|--------|------|
| All settings endpoints | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Settings form values | React state per tab | Each settings page |
| Navigation tree draft | React state until save | `/navigation` |
| Dirty / unsaved warning | React state | All settings pages |
| Cached settings | React Query (refetch on mount) | Per section |
| Upload previews | Local blob URLs | General/SEO forms |
