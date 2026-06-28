# Admin Context — Product Slides

**Route:** `/context/product-slides`  
**Status:** 🆕 Not implemented (proposed)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Manage three homepage product carousels:

1. **Featured** — tabbed carousel; each tab has a label and product set
2. **Bestseller** — single carousel of top-selling products
3. **Discounted** — products with active sale pricing

Each carousel has configurable title, autoplay interval, active flag, and ordered product items.

---

## User Flow

1. Admin opens **Context → Product Slides**.
2. UI shows three sections (tabs or accordion): Featured, Bestseller, Discounted.
3. For each carousel:
   - Edit section title and autoplay interval.
   - Search/add products via product picker (`GET /admin/products/search`).
   - Drag to reorder items.
   - For **Featured** only: assign `tab_label` per item group.
4. Toggle carousel active state.
5. Save section → `PUT /admin/storefront/product-slides/{slide_type}` or bulk `PUT /admin/storefront/product-slides`.

---

## Business Logic

- Exactly **three** slide configs exist, keyed by `slide_type`: `featured`, `bestseller`, `discounted` (unique constraint).
- Items reference existing `products.id`; deleted/archived products filtered out on storefront read.
- **Featured tabs:** Items with the same `tab_label` form one tab; `sort_order` determines order within tab; tabs ordered by first item's sort order.
- **Bestseller / Discounted:** Flat ordered list; backend may optionally auto-populate discounted slide from `sale_price IS NOT NULL` if admin enables "auto mode" (future).
- Default `autoplay_interval_ms`: 4500.
- Max items per carousel: 20 (recommended validation).
- Storefront renders via `GET /store/homepage` → `product_slides[]`.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Product deleted after assignment | Skip on storefront; show warning badge in admin |
| Product set to draft | Hidden on storefront; visible in admin with status badge |
| Duplicate product in same carousel | Reject or dedupe on save |
| Empty carousel + active | Storefront hides section |
| Featured item without tab_label | Default tab "Featured" |
| More than 20 items | `422` validation error |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Products | Product picker, validation product exists |
| Storefront content | Slide CRUD |
| Store homepage API | Public aggregation |

---

## Required APIs

### Proposed — GET `/api/v1/admin/storefront/product-slides`

Returns all three carousels.

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "slide_type": "featured",
      "title": "Featured Products",
      "autoplay_interval_ms": 4500,
      "is_active": true,
      "sort_order": 0,
      "items": [
        {
          "id": "uuid",
          "product_id": "uuid",
          "product_name": "Ceramic Tile 60x60",
          "product_image": "https://…",
          "tab_label": "Tiles",
          "sort_order": 0
        }
      ],
      "updated_at": "2026-06-01T10:00:00Z"
    },
    {
      "slide_type": "bestseller",
      "title": "Best Sellers",
      "items": []
    },
    {
      "slide_type": "discounted",
      "title": "Special Offers",
      "items": []
    }
  ]
}
```

**Errors:** `401`, `403`

---

### Proposed — GET `/api/v1/admin/storefront/product-slides/{slide_type}`

**Path:** `slide_type` = `featured` | `bestseller` | `discounted`

**Response 200:** Single slide object (same shape as array item above).

**Errors:** `404`, `401`, `403`

---

### Proposed — PUT `/api/v1/admin/storefront/product-slides/{slide_type}`

**Request:**

```json
{
  "title": "Featured Products",
  "autoplay_interval_ms": 5000,
  "is_active": true,
  "items": [
    { "product_id": "uuid", "tab_label": "Tiles", "sort_order": 0 },
    { "product_id": "uuid", "tab_label": "Tiles", "sort_order": 1 },
    { "product_id": "uuid", "tab_label": "Tools", "sort_order": 2 }
  ]
}
```

**Response 200:** Updated slide with enriched product fields.

**Errors:** `400`, `404` (product), `422`, `401`, `403`

---

## Database Impact

**Tables:** `product_slides`, `product_slide_items` (migration `000010`)

| Table | Notes |
|-------|-------|
| `product_slides` | One row per `slide_type` (UNIQUE) |
| `product_slide_items` | FK → `product_slides`, `products`; `tab_label` for featured only |

**Operations on save:** Replace all items for slide (delete + insert in transaction).

**Index:** `(slide_id, sort_order)`

**Cache:** Invalidate `homepage`.

---

## UI Changes Affecting Backend

| UI element | Backend impact |
|------------|----------------|
| Product search modal | Uses existing `/admin/products/search` |
| Drag-and-drop reorder | Send updated `sort_order` array on save |
| Featured tab editor | Group items by `tab_label` client-side; persist flat list |
| Autoplay slider | Maps to `autoplay_interval_ms` (min 2000, max 15000) |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `slide_type` | Enum: `featured`, `bestseller`, `discounted` |
| `title` | Optional; max 255 chars |
| `autoplay_interval_ms` | Integer 2000–15000 |
| `items` | Max 20 per slide |
| `items[].product_id` | Required UUID; product must exist |
| `items[].tab_label` | Optional; max 100; required for featured grouping |
| `items[].sort_order` | Integer `>= 0` |

---

## Permission Requirements

| Action | Role |
|--------|------|
| View/edit product slides | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Active carousel tab (Featured/Bestseller/Discounted) | URL hash or local state | Page |
| Draft items per carousel | React state until save | Page |
| Product picker modal | React state | Modal |
| Saved slides | React Query | Refetch after PUT |
| Unsaved changes | React state + `beforeunload` | Page |
