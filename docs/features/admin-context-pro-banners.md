# Admin Context — Pro Banners

**Route:** `/context/pro-banners`  
**Status:** 🆕 Not implemented (proposed)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Manage promotional banner images displayed on the storefront homepage. Each banner supports separate desktop and mobile images, an optional click-through URL, sort order, and active flag. Multiple banners can exist in a ordered list.

---

## User Flow

1. Admin navigates to **Context → Pro Banners**.
2. Page loads banner list: `GET /admin/storefront/pro-banners`.
3. Admin clicks **Add Banner** → empty form.
4. Upload desktop image (required) and optional mobile image via `POST /admin/uploads?context=banner`.
5. Set link URL, sort order, active toggle.
6. Save → `POST /admin/storefront/pro-banners` (create) or `PUT /admin/storefront/pro-banners/{id}` (update).
7. Reorder via drag-and-drop → batch `PUT` with updated `sort_order` or dedicated reorder endpoint.
8. Delete → `DELETE /admin/storefront/pro-banners/{id}`.

---

## Business Logic

- Banners render in `sort_order` ascending; only `is_active = true` shown on storefront.
- **Desktop image required**; mobile image optional (storefront uses desktop as fallback on mobile if mobile absent).
- `link_url` optional — banner non-clickable if empty.
- Recommended image ratios: desktop 16:9 or 21:9; mobile 4:5 or 1:1.
- Max banners: 10 active (soft limit; validate in API).
- Storefront serves via `GET /store/homepage` → `pro_banners[]`.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| No mobile image | Use desktop on all breakpoints |
| Invalid link URL | `400` validation (allow relative paths) |
| All banners inactive | Homepage section hidden |
| Delete last banner | Allowed; empty state in admin |
| Upload wrong dimensions | Warning in UI only; backend accepts any valid image |
| External link | Storefront opens in new tab (frontend behavior) |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Upload service | Image storage (`context=banner`) |
| Storefront content | Banner CRUD |
| Store homepage API | Public read |

---

## Required APIs

### Proposed — GET `/api/v1/admin/storefront/pro-banners`

**Query:** `page`, `per_page` (optional pagination)

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "desktop_image_url": "https://cdn.example.com/banners/desktop-1.jpg",
      "mobile_image_url": "https://cdn.example.com/banners/mobile-1.jpg",
      "link_url": "/products?category=tiles",
      "sort_order": 0,
      "is_active": true,
      "created_at": "2026-06-01T00:00:00Z",
      "updated_at": "2026-06-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 3, "total_pages": 1 }
}
```

**Errors:** `401`, `403`

---

### Proposed — POST `/api/v1/admin/storefront/pro-banners`

**Request:**

```json
{
  "desktop_image_url": "https://cdn.example.com/banners/desktop-1.jpg",
  "mobile_image_url": "https://cdn.example.com/banners/mobile-1.jpg",
  "link_url": "/products",
  "sort_order": 0,
  "is_active": true
}
```

**Response 201:** Created banner object.

**Errors:** `400`, `401`, `403`, `422`

---

### Proposed — GET `/api/v1/admin/storefront/pro-banners/{id}`

**Response 200:** Single banner.

**Errors:** `404`, `401`, `403`

---

### Proposed — PUT `/api/v1/admin/storefront/pro-banners/{id}`

**Request:** Same fields as POST (partial update supported).

**Response 200:** Updated banner.

**Errors:** `400`, `404`, `401`, `403`

---

### Proposed — DELETE `/api/v1/admin/storefront/pro-banners/{id}`

**Response 204**

**Errors:** `404`, `401`, `403`

---

### Supporting — POST `/api/v1/admin/uploads`

**Query:** `context=banner`  
**Constraints:** Image MIME only (`image/jpeg`, `image/png`, `image/webp`); max 5 MB.

---

## Database Impact

**Table:** `pro_banners` (migration `000010`)

| Column | Type | Notes |
|--------|------|-------|
| `desktop_image_url` | VARCHAR(500) NOT NULL | |
| `mobile_image_url` | VARCHAR(500) NULL | |
| `link_url` | VARCHAR(500) NULL | |
| `sort_order` | INT DEFAULT 0 | |
| `is_active` | BOOLEAN DEFAULT true | |

**Cache:** Invalidate `homepage` on any mutation.

---

## UI Changes Affecting Backend

| UI element | Backend impact |
|------------|----------------|
| Desktop/mobile upload slots | Two upload calls; two URL fields |
| Drag reorder | Update `sort_order` on affected rows |
| Link URL field | Optional; support relative paths |
| Preview thumbnails | Frontend-only |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `desktop_image_url` | Required; valid URL; max 500 |
| `mobile_image_url` | Optional; valid URL; max 500 |
| `link_url` | Optional; max 500; relative or absolute |
| `sort_order` | Integer `>= 0` |
| `is_active` | Boolean |
| Max active banners | 10 (recommended `422` if exceeded) |
| Image upload | Max 5 MB; allowed image types |

---

## Permission Requirements

| Action | Role |
|--------|------|
| CRUD pro banners | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Banner list | React Query | Page |
| Edit form / modal | React state | Create/edit |
| Upload previews | Local object URLs until upload completes | Form |
| Reorder draft | React state; commit on save | List |
| Delete confirmation | Modal state | Ephemeral |
