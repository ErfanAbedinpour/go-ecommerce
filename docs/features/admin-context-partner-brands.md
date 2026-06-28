# Admin Context — Partner Brands

**Route:** `/context/brands`  
**Status:** 🆕 Not implemented (proposed)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Manage partner/trust brand logos displayed on the storefront homepage (e.g. manufacturer logos). These are **marketing entities**, separate from product catalog brands used in `/products/settings`.

Each partner brand has a title, optional description, logo, optional link, sort order, and active flag.

---

## User Flow

1. Admin opens **Context → Partner Brands** (`/context/brands`).
2. List loads: `GET /admin/storefront/partner-brands`.
3. **Add brand** → form with title, description, logo upload, link URL.
4. Upload logo → `POST /admin/uploads?context=brand`.
5. Save → `POST /admin/storefront/partner-brands`.
6. Edit/delete/reorder existing entries.
7. Toggle active to show/hide on homepage.

---

## Business Logic

- **Not the same as** `brands` table (product taxonomy). Partner brands are homepage-only marketing content.
- Displayed as logo grid/carousel on storefront homepage.
- Sorted by `sort_order` ascending; inactive brands hidden on storefront.
- `link_url` optional — clicking logo navigates if set.
- Logo should be transparent PNG/SVG preferred; backend accepts standard images.
- No hard limit on count; UI recommends max 12 visible for layout.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Missing logo on create | `400` — logo required |
| Long description | Truncate on storefront card; full text in admin |
| Duplicate titles | Allowed (different logos) |
| Broken logo URL | Show placeholder on storefront |
| Brand with same name as product brand | Allowed — separate tables |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Upload service | Logo images |
| Storefront content | Partner brand CRUD |
| Store homepage API | Public read |

**Explicit separation:** Do not merge with `GET /admin/brands` (product settings).

---

## Required APIs

### Proposed — GET `/api/v1/admin/storefront/partner-brands`

**Query:** `page`, `per_page`, `is_active` (optional filter)

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "title": "Knauf",
      "description": "Leading drywall and insulation manufacturer",
      "logo_url": "https://cdn.example.com/brands/knauf.png",
      "link_url": "https://knauf.com",
      "sort_order": 0,
      "is_active": true,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-06-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 8, "total_pages": 1 }
}
```

**Errors:** `401`, `403`

---

### Proposed — POST `/api/v1/admin/storefront/partner-brands`

**Request:**

```json
{
  "title": "Knauf",
  "description": "Leading drywall manufacturer",
  "logo_url": "https://cdn.example.com/brands/knauf.png",
  "link_url": "https://knauf.com",
  "sort_order": 0,
  "is_active": true
}
```

**Response 201:** Created brand object.

**Errors:** `400`, `401`, `403`

---

### Proposed — GET `/api/v1/admin/storefront/partner-brands/{id}`

**Response 200:** Single brand.

**Errors:** `404`, `401`, `403`

---

### Proposed — PUT `/api/v1/admin/storefront/partner-brands/{id}`

**Request:** Partial update supported.

**Response 200:** Updated brand.

**Errors:** `400`, `404`, `401`, `403`

---

### Proposed — DELETE `/api/v1/admin/storefront/partner-brands/{id}`

**Response 204**

**Errors:** `404`, `401`, `403`

---

## Database Impact

**Table:** `partner_brands` (migration `000010`)

| Column | Notes |
|--------|-------|
| `title` | VARCHAR(255) NOT NULL |
| `description` | TEXT nullable |
| `logo_url` | VARCHAR(500) NOT NULL |
| `link_url` | VARCHAR(500) nullable |
| `sort_order`, `is_active` | Display control |

**Cache:** Invalidate `homepage` on mutation.

---

## UI Changes Affecting Backend

| UI element | Backend impact |
|------------|----------------|
| Logo uploader | Required field → `logo_url` |
| Description textarea | Optional TEXT field |
| External link field | `link_url`; validate URL format |
| Grid reorder | Batch update `sort_order` |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `title` | Required; max 255 chars |
| `description` | Optional; max 2000 chars |
| `logo_url` | Required; valid URL; max 500 |
| `link_url` | Optional; valid URL; max 500 |
| `sort_order` | Integer `>= 0` |
| `is_active` | Boolean |
| Logo upload | Image types; max 5 MB |

---

## Permission Requirements

| Action | Role |
|--------|------|
| CRUD partner brands | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Brand list | React Query | Page |
| Form modal (create/edit) | React state | Modal |
| Logo preview | Local blob URL → CDN URL after upload | Form |
| Sort draft | React state until save | List |
