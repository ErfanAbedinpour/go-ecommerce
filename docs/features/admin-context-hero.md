# Admin Context — Hero Section

**Route:** `/context/hero`  
**Status:** 🆕 Not implemented (proposed)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Configure the storefront homepage hero: background video, overlay title/subtitle, and up to two call-to-action buttons. Content is a **singleton** — one active hero configuration per store.

---

## User Flow

1. Admin navigates to **Context → Hero** in sidebar.
2. Page loads current hero config (or empty defaults).
3. Admin uploads hero video via file picker → `POST /admin/uploads` with context `hero`.
4. Admin sets title, subtitle, primary/secondary CTA text and URLs.
5. Toggle **Active** to show/hide on storefront.
6. Click **Save** → `PUT /admin/storefront/hero`.
7. Optional **Preview** opens storefront homepage in new tab.

---

## Business Logic

- Only one hero record exists (`storefront_hero` singleton).
- First save creates the row; subsequent saves update in place.
- `video_url` optional — storefront falls back to static image or gradient if absent.
- CTAs are optional; at least one of title or video recommended for non-empty hero.
- Deactivating (`is_active = false`) hides hero on storefront; admin preview may still show draft.
- Video served from upload storage; recommend MP4 H.264, max 50 MB.
- Storefront reads hero via `GET /store/homepage` (aggregated) or `GET /store/homepage/hero`.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Save without video | Allowed; text-only hero |
| Invalid video MIME | Upload rejected `400` |
| Video > 50 MB | Upload rejected |
| CTA URL relative (`/products`) | Allowed |
| CTA URL external | Allowed; storefront opens in same tab or new tab per UI config |
| Broken video URL after file delete | Storefront shows fallback; admin shows warning |
| Concurrent saves | Last write wins |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Upload service | Video file storage |
| Storefront content module | CRUD + cache invalidation tag `homepage` |
| Store homepage API | Public read |

---

## Required APIs

### Proposed — GET `/api/v1/admin/storefront/hero`

**Response 200:**

```json
{
  "id": "uuid",
  "video_url": "https://cdn.example.com/hero/intro.mp4",
  "title": "Premium Building Materials",
  "subtitle": "Quality tiles, cement, and tools delivered nationwide",
  "cta_primary_text": "Shop Now",
  "cta_primary_url": "/products",
  "cta_secondary_text": "Learn More",
  "cta_secondary_url": "/about",
  "is_active": true,
  "updated_at": "2026-06-01T10:00:00Z"
}
```

**Errors:** `401`, `403`

---

### Proposed — PUT `/api/v1/admin/storefront/hero`

**Request:**

```json
{
  "video_url": "https://cdn.example.com/hero/intro.mp4",
  "title": "Premium Building Materials",
  "subtitle": "Quality tiles, cement, and tools",
  "cta_primary_text": "Shop Now",
  "cta_primary_url": "/products",
  "cta_secondary_text": "Learn More",
  "cta_secondary_url": "/about",
  "is_active": true
}
```

**Response 200:** Same as GET body.

**Errors:** `400 VALIDATION_ERROR`, `401`, `403`

---

### Supporting — POST `/api/v1/admin/uploads`

**Form field:** `file`  
**Optional query:** `context=hero`

**Response 200:**

```json
{
  "url": "/uploads/hero/uuid.mp4",
  "filename": "intro.mp4",
  "size": 5242880,
  "content_type": "video/mp4"
}
```

**Errors:** `400` (invalid type/size), `401`, `403`

---

## Database Impact

**Table:** `storefront_hero` (migration `000010_storefront_content`)

| Column | Notes |
|--------|-------|
| `video_url` | VARCHAR(500), nullable |
| `title`, `subtitle` | Overlay copy |
| `cta_primary_*`, `cta_secondary_*` | Button labels + URLs |
| `is_active` | BOOLEAN, default true |
| `updated_at` | Auto-updated |

**Cache:** Invalidate tag `homepage` on PUT.

---

## UI Changes Affecting Backend

| UI element | Backend requirement |
|------------|---------------------|
| Video upload progress | Multipart upload to `/uploads`; store returned URL in PUT body |
| Video preview player | Frontend-only; uses `video_url` |
| Active toggle | Maps to `is_active` |
| Character counters | Enforce max lengths server-side |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `video_url` | Optional; valid URL; max 500 chars |
| `title` | Optional; max 255 chars |
| `subtitle` | Optional; text |
| `cta_primary_text`, `cta_secondary_text` | Optional; max 100 chars |
| `cta_primary_url`, `cta_secondary_url` | Optional; max 500 chars; relative or absolute URL |
| `is_active` | Boolean |
| Upload (video) | MIME: `video/mp4`, `video/webm`; max 50 MB |

---

## Permission Requirements

| Action | Role |
|--------|------|
| View/edit hero | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Form values | React state (controlled form) | Page |
| Upload progress | React state | Upload widget |
| Dirty flag / unsaved warning | React state | Page |
| Saved config | React Query cache | Refetch after PUT |
| Preview URL | Derived from env + storefront base | Preview button |
