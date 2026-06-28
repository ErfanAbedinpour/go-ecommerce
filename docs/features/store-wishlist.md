# Store Wishlist

> **Route:** `/account/wishlist`  
> **UI:** [store-os-eta.vercel.app/account/wishlist](https://store-os-eta.vercel.app/account/wishlist)  
> **Locale:** Persian (fa-IR), RTL

---

## Purpose

Wishlist lets authenticated customers save products for later purchase. Items persist server-side per customer. The wishlist page displays saved products as a grid with remove action, stock/price status, and quick add-to-cart. Wishlist toggle (heart icon) also appears on catalog cards and product detail pages.

---

## User Flow

```mermaid
flowchart TD
    A[Heart icon on product] --> B{Authenticated?}
    B -->|No| C[Login modal → returnUrl]
    B -->|Yes| D{In wishlist?}
    D -->|No| E[POST /account/wishlist]
    D -->|Yes| F[DELETE /account/wishlist/:product_id]
    E --> G[Optimistic heart fill]
    F --> H[Optimistic heart empty]
    I[/account/wishlist] --> J{Authenticated?}
    J -->|No| K[Redirect login]
    J -->|Yes| L[GET /account/wishlist]
    L --> M[Render product grid]
    M --> N[Remove item]
    N --> F
    M --> O[Add to cart]
    O --> P[localStorage cart update]
    M --> Q[Click card → /products/:slug]
```

1. **Add from catalog/detail:** Tap heart → API call → toast "به علاقه‌مندی‌ها اضافه شد".
2. **Wishlist page:** Full grid of saved products with current price/stock.
3. **Remove:** Trash icon or heart toggle → DELETE API.
4. **Add to cart:** Opens variant modal if product has SKUs; else direct add.
5. **Empty wishlist:** CTA "مشاهده محصولات" → `/products`.

---

## Business Logic

### Scope

- Wishlist is per **product** (not per SKU) in v1.
- Unique constraint: `(customer_id, product_id)`.
- Only `status = 'active'` products shown; inactive products auto-hidden with optional cleanup job.

### Add behavior

- Idempotent: duplicate POST returns `200` with existing item (no error).
- Product must exist and be active; else `404`.

### Remove behavior

- DELETE is idempotent: removing non-existent item returns `204`.

### Price/stock on list

- Resolve live prices from `products` / `skus` at read time (not snapshotted at save).
- Show "قیمت تغییر کرده" badge if price differs significantly (optional).
- Out-of-stock: disable add-to-cart, show "ناموجود".

### Sync with product pages

- `GET /store/products/{slug}` includes `is_in_wishlist: boolean` when authenticated.
- Catalog batch: optional `GET /store/account/wishlist/ids` for heart state on list pages.

### Guest users

- Heart click → login prompt with `returnUrl` to current page.
- No localStorage wishlist in v1 (assumption A6).

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Product deleted after save | Filter from list; lazy DELETE orphan rows |
| Product moved to draft | Hide from wishlist UI |
| User logs out on wishlist page | Redirect to login |
| Concurrent add from two tabs | Unique constraint prevents duplicate |
| Add out-of-stock to cart | Show error; item stays in wishlist |
| Large wishlist (100+ items) | Paginate list API |
| Price drop | UI may show discount badge (live data) |

---

## Dependencies

### Backend

| Module | Role |
|--------|------|
| `internal/application/storefront/wishlist` | CRUD |
| `internal/application/storefront` | Product card enrichment |
| `internal/application/auth` | Customer auth |

### Tables

- `wishlist_items` (new)
- `products`, `product_images`, `inventories`, `skus`
- `customers`

### Frontend

- Auth guard on `/account/wishlist`
- Shared `WishlistButton` component (catalog, detail, wishlist page)
- Cart integration (`localStorage`)

---

## Required APIs

### `GET /api/v1/store/account/wishlist`

**Auth:** Bearer + `customer`

**Query:** `page`, `per_page` (default 20, max 48)

**Response 200**

```json
{
  "data": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "added_at": "2026-06-01T10:00:00Z",
      "product": {
        "id": "uuid",
        "slug": "ceramic-tile-60x60",
        "name": "کاشی سرامیک ۶۰×۶۰",
        "thumbnail_url": "https://…",
        "brand": "مرجان",
        "price_toman": 450000,
        "sale_price_toman": 399000,
        "price_from_toman": 399000,
        "is_on_sale": true,
        "is_out_of_stock": false,
        "has_variants": true
      }
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 7, "total_pages": 1 }
}
```

### `POST /api/v1/store/account/wishlist`

**Auth:** Bearer + `customer`

**Request**

```json
{
  "product_id": "uuid"
}
```

**Response 201** (new item) / **200** (already exists)

```json
{
  "id": "uuid",
  "product_id": "uuid",
  "added_at": "2026-06-01T10:00:00Z"
}
```

**Errors:** `404` product not found, `400` product not active.

### `DELETE /api/v1/store/account/wishlist/{product_id}`

**Auth:** Bearer + `customer`

**Response 204** No content.

### `GET /api/v1/store/account/wishlist/ids` (optional, for catalog hearts)

**Auth:** Bearer + `customer`

**Response 200**

```json
{
  "product_ids": ["uuid1", "uuid2", "uuid3"]
}
```

Lightweight endpoint to avoid N+1 on product list pages.

### `GET /api/v1/store/account/wishlist/count` (optional)

**Response:** `{ "count": 7 }` for header badge.

---

## Database Impact

### Table: `wishlist_items`

| Column | Type | Notes |
|--------|------|-------|
| id | UUID | PK |
| customer_id | UUID | FK → customers |
| product_id | UUID | FK → products |
| created_at | TIMESTAMPTZ | `added_at` in API |

**Unique:** `(customer_id, product_id)`

**Indexes:** `(customer_id, created_at DESC)`

### Operations

| Operation | SQL |
|-----------|-----|
| List | SELECT with JOIN products, images, inventory |
| Add | INSERT ON CONFLICT DO NOTHING |
| Remove | DELETE WHERE customer_id AND product_id |

### Migration

- `000013_engagement` — creates `wishlist_items`

### Cleanup job (optional)

Periodic job removes wishlist rows where product `status != 'active'` or soft-deleted.

---

## Validation

### POST body

| Field | Rules |
|-------|-------|
| `product_id` | Required, valid UUID, product must be active |

### DELETE path

| Param | Rules |
|-------|-------|
| `product_id` | Valid UUID |

### List query

Standard pagination validation (`page >= 1`, `per_page` 1–48).

---

## Permissions

| Action | Role |
|--------|------|
| View wishlist page | Customer |
| List wishlist items | Customer (own data only) |
| Add to wishlist | Customer |
| Remove from wishlist | Customer (own items only) |
| View heart state on products | Customer |

Public users see empty heart; click triggers login.

---

## State Management

### Server state

| Query key | Endpoint | Use |
|-----------|----------|-----|
| `['wishlist']` | GET full list | Wishlist page |
| `['wishlist','ids']` | GET ids | Catalog/detail hearts |
| `['wishlist','count']` | GET count | Header badge |

### Optimistic updates

On heart toggle:

1. Immediately flip `is_in_wishlist` in product cache
2. Fire POST or DELETE
3. On error: rollback + toast
4. Invalidate `['wishlist']` and `['wishlist','ids']`

### Header badge

- Subscribe to `wishlist.count` from React Query
- Increment/decrement on successful add/remove

### Cart interaction

- Add-to-cart from wishlist uses same `localStorage.store_cart_v1` as catalog
- After add: optional toast "به سبد خرید اضافه شد" (item remains in wishlist)

### UI components

```
WishlistButton({ productId, isInWishlist })
WishlistPage → ProductGrid with remove + addToCart actions
```

### Route protection

```typescript
// Pseudocode
if (!isAuthenticated) redirect(`/login?returnUrl=/account/wishlist`);
```
