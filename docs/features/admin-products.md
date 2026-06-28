# Admin Products

**Routes:** `/products`, `/products/create`, `/products/:id` (edit), `/products/settings`  
**Status:** ✅ Backend implemented (variant/SKU matrix supported)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Products management covers the full catalog lifecycle: list and search products, create/edit with variant attributes and auto-generated SKUs, manage inventory, and configure catalog taxonomy (categories, brands, attribute definitions) on the settings page.

---

## User Flow

### `/products` — Product list

1. Page loads → `GET /admin/products/stats` (KPI cards) + `GET /admin/products` (table).
2. Admin uses search box → debounced `GET /admin/products/search?q=`.
3. Admin applies filters (status, category, brand, featured, stock level) → `GET /admin/products` with query params.
4. Row actions: Edit → `/products/:id`, Delete → `DELETE /admin/products/{id}` with confirmation.

### `/products/create` and edit

1. Load category tree: `GET /admin/categories?tree=true`.
2. Load brands list (optional autocomplete): `GET /admin/brands`.
3. Admin fills form: name, pricing, descriptions, images, variant attributes, inventory.
4. Upload images via `POST /admin/uploads` (multipart `file`).
5. Submit create → `POST /admin/products` or update → `PUT /admin/products/{id}`.
6. Adjust stock separately → `PATCH /admin/products/{id}/inventory`.

### `/products/settings` — Catalog settings

Tabbed UI for:

- **Categories** — CRUD via `/admin/categories`
- **Brands** — CRUD via `/admin/brands`
- **Product attributes** — CRUD via `/admin/product-attributes`
- **Attribute values** — CRUD via `/admin/product-attribute-values?attribute_id=`

---

## Business Logic

### Product lifecycle

| Status | Meaning |
|--------|---------|
| `draft` | Not visible on storefront |
| `active` | Published |
| `archived` | Hidden, retained for order history |

### Variant / SKU generation

When `attributes[]` is provided on create or update:

1. Each attribute has a `name` and one or more `values`.
2. Backend computes the Cartesian product of all attribute value combinations.
3. Each combination becomes a `Sku` row with:
   - `code`: `{slug}-{value1}-{value2}-…` (uppercased, sanitized)
   - `attributes`: map of `{ "Color": "Black", "Size": "M" }`
4. **Max combinations:** 1000 (`ErrMaxVariantsExceeded` if exceeded).
5. Duplicate attribute names or values within a product are rejected.
6. Generated SKU codes must be globally unique (`409 CONFLICT` if collision).

### Inventory

- Inventory is **product-level** (not per-SKU in v1).
- `is_low_stock` = `quantity <= low_stock_threshold`.
- `is_out_of_stock` = `quantity == 0`.
- Stock decremented on order placement; restored on cancel.

### Delete rules

- Soft delete only.
- Fails with `422` if product referenced by active (non-terminal) orders.

### Catalog settings

- Categories support parent/child tree and `products_count`.
- Product attribute definitions are global templates; product-level attributes on create are independent copies.
- Brands on settings page are **product taxonomy brands**, distinct from homepage partner brands (`/context/brands`).

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Create product without attributes | Single implicit product; `skus[]` empty in response |
| Update attributes on product with orders | Regenerates SKUs; existing order line items retain snapshot SKU |
| Search query < 2 chars | Returns empty result set |
| Delete product in active order | `422 UNPROCESSABLE` — `product has active orders` |
| Duplicate SKU code globally | `409 CONFLICT` |
| >1000 variant combinations | `422 VALIDATION` — max variants exceeded |
| Sale price > regular price | Allowed at API level; UI should warn |
| Upload non-image file for product image | Upload handler validates MIME type |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Categories | Product assignment, settings CRUD |
| Brands | Product brand field, settings CRUD |
| Attribute definitions | Settings templates; optional reference |
| Uploads | Image URLs for products and categories |
| Orders | Delete guard, inventory decrement |
| Dashboard | Featured products, low stock widgets |

---

## Required APIs

All require Bearer token + `admin` role.

### Product list & search

#### GET `/api/v1/admin/products`

**Query:** `page`, `per_page`, `sort`, `order`, `status`, `category_id`, `brand`, `is_featured`, `stock_level` (`low` \| `out`)

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "category_id": "uuid",
      "name": "Nike Air Max",
      "slug": "nike-air-max",
      "description": "…",
      "short_description": "…",
      "price": 129.99,
      "sale_price": 99.99,
      "brand": "Nike",
      "is_featured": false,
      "status": "active",
      "images": [{ "id": "uuid", "url": "https://…", "alt_text": "", "sort_order": 0 }],
      "attributes": [{ "id": "uuid", "name": "Color", "values": ["Black", "White"] }],
      "skus": [
        { "id": "uuid", "code": "NIKE-AIR-MAX-BLACK", "attributes": { "Color": "Black" } }
      ],
      "inventory": {
        "quantity": 50,
        "low_stock_threshold": 10,
        "is_low_stock": false,
        "is_out_of_stock": false
      },
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 245, "total_pages": 13 }
}
```

#### GET `/api/v1/admin/products/stats`

**Response 200:**

```json
{ "total": 245, "active": 198, "draft": 32, "out_of_stock": 15 }
```

#### GET `/api/v1/admin/products/search?q={query}`

Same response shape as list. Minimum query length: 2 characters.

#### GET `/api/v1/admin/products/{id}`

Single `ProductResponse`.

#### POST `/api/v1/admin/products`

**Request:**

```json
{
  "name": "Nike Air Max",
  "slug": "nike-air-max",
  "description": "Full description",
  "short_description": "Short",
  "price": 129.99,
  "sale_price": 99.99,
  "category_id": "uuid",
  "brand": "Nike",
  "is_featured": false,
  "status": "draft",
  "images": [{ "url": "https://cdn/…/img.jpg", "alt_text": "", "sort_order": 0 }],
  "attributes": [
    { "name": "Color", "values": ["Black", "White"] },
    { "name": "Size", "values": ["M", "L"] }
  ],
  "inventory": { "quantity": 50, "low_stock_threshold": 10 }
}
```

**Response 201:** `ProductResponse`

**Errors:** `400`, `404` (category), `409` (SKU conflict), `422` (variants)

#### PUT `/api/v1/admin/products/{id}`

Partial update supported. Include `attributes` to regenerate SKU matrix.

**Errors:** Same as create + `404`

#### DELETE `/api/v1/admin/products/{id}`

**Response 204**

**Errors:** `404`, `422` (active orders)

#### PATCH `/api/v1/admin/products/{id}/inventory`

**Request:**

```json
{
  "quantity": 100,
  "low_stock_threshold": 15,
  "adjustment_reason": "Restock"
}
```

**Response 200:** Updated `ProductResponse`

---

### Uploads

#### POST `/api/v1/admin/uploads`

**Content-Type:** `multipart/form-data`, field `file`

**Response 200:**

```json
{
  "url": "/uploads/products/uuid.jpg",
  "filename": "photo.jpg",
  "size": 245760,
  "content_type": "image/jpeg"
}
```

---

### Settings — Categories

| Method | Route | Notes |
|--------|-------|-------|
| GET | `/api/v1/admin/categories?tree=true` | Tree for dropdowns |
| GET | `/api/v1/admin/categories` | Flat paginated list |
| POST | `/api/v1/admin/categories` | Create |
| GET | `/api/v1/admin/categories/{id}` | Detail |
| PUT | `/api/v1/admin/categories/{id}` | Update |
| DELETE | `/api/v1/admin/categories/{id}` | Delete (fails if has products/children) |

**Category response includes:** `products_count`, `children[]`, `parent_id`, `sort_order`, `is_active`

---

### Settings — Brands

Full CRUD at `/api/v1/admin/brands` and `/api/v1/admin/brands/{id}`.

---

### Settings — Product attributes

| Method | Route |
|--------|-------|
| GET/POST | `/api/v1/admin/product-attributes` |
| GET/PUT/DELETE | `/api/v1/admin/product-attributes/{id}` |
| GET/POST | `/api/v1/admin/product-attribute-values` |
| GET/PUT/DELETE | `/api/v1/admin/product-attribute-values/{id}` |

**Create attribute value:**

```json
{ "attribute_id": "uuid", "value": "Black", "sort_order": 0, "is_active": true }
```

---

## Database Impact

| Table | Operations |
|-------|------------|
| `products` | INSERT, UPDATE, soft DELETE |
| `product_images` | REPLACE on product update |
| `product_attributes`, `product_attribute_values` | REPLACE on attribute update |
| `skus` | INSERT/REPLACE on variant regeneration |
| `inventory` | INSERT/UPDATE |
| `categories`, `brands` | Settings CRUD |
| `product_attribute_definitions`, `product_attribute_value_definitions` | Settings CRUD |

**Future (migration 000014):** Per-SKU `price_override`, `sale_price_override` on `skus` table.

---

## UI Changes Affecting Backend

| UI feature | Backend impact |
|------------|----------------|
| Variant matrix picker | Send `attributes[]` with name + values array; display returned `skus[]` |
| KPI cards on list page | Use dedicated `/products/stats` endpoint |
| Stock filter tabs | Map to `stock_level=low\|out` query param |
| Image gallery reorder | Send `sort_order` on each image in PUT body |
| Settings tabs | Separate API calls per resource; no combined endpoint |

---

## Validation Requirements

### Product create/update

| Field | Rule |
|-------|------|
| `name` | Required, 1–300 chars |
| `slug` | Optional, max 300; auto-generated from name if omitted |
| `price` | Required, `>= 0` |
| `sale_price` | Optional, `>= 0` |
| `category_id` | Valid UUID if provided |
| `status` | `draft` \| `active` \| `archived` |
| `images` | Max 10; each `url` required, valid URL |
| `attributes[].name` | Required, max 100, unique per product |
| `attributes[].values` | Min 1 value, max 200 chars each, unique per attribute |
| Variant combinations | Max 1000 |
| `inventory.quantity` | `>= 0` |
| `inventory.low_stock_threshold` | `>= 0` |

---

## Permission Requirements

| Action | Role |
|--------|------|
| All product & catalog settings APIs | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Product list filters | URL query params | `/products` |
| Search query | URL `?q=` or local debounced state | `/products` |
| Product form draft | `sessionStorage` | Create/edit |
| Unsaved changes warning | React state | Create/edit |
| Settings active tab | URL hash or query | `/products/settings` |
| Category tree cache | React Query (5 min stale) | Global catalog |
