# Store Product Catalog

> **Route:** `/products`  
> **UI:** [store-os-eta.vercel.app/products](https://store-os-eta.vercel.app/products)  
> **Locale:** Persian (fa-IR), RTL

---

## Purpose

The product catalog is the browsable inventory of building materials. Customers search, filter by category, and sort products using tab controls (پرفروش / جدید / تخفیف‌دار). Results render as a responsive product grid with quick actions (add to cart, wishlist heart).

---

## User Flow

```mermaid
flowchart TD
    A[/products] --> B[Read URL query params]
    B --> C[GET /store/products]
    C --> D[Render grid + filters]
    D --> E{User action}
    E -->|Search| F[Debounce q → refetch]
    E -->|Category chip| G[Toggle category_id in URL]
    E -->|Sort tab| H[Set sort param → refetch]
    E -->|Pagination| I[Update page param]
    E -->|Product card| J[/products/:id]
    E -->|Add to cart| K[Local cart update]
    E -->|Wishlist| L{Authenticated?}
    L -->|Yes| M[POST /store/account/wishlist]
    L -->|No| N[Login modal]
```

1. Land on `/products` or arrive from homepage category link (`?category=tiles`).
2. Sidebar/top chips show category tree; selecting filters the grid.
3. Search box filters by name, SKU code, brand, description (debounced 300ms).
4. Sort tabs map to API `sort` param.
5. Infinite scroll or numbered pagination (UI uses pagination per design).
6. Empty state when no matches.

---

## Business Logic

### Visibility rules

- Only `status = 'active'` products appear.
- Soft-deleted products (`deleted_at IS NOT NULL`) excluded.
- Out-of-stock products **shown** with badge; purchasable only if `quantity > 0` at SKU/product level.

### Category filter

- `category_id` or `category_slug` query param.
- Include products in child categories when parent selected (optional `include_children=true`, default `true`).
- Categories with zero active products hidden from filter chips.

### Sort tabs

| UI tab (Persian) | `sort` param | `order` |
|------------------|--------------|---------|
| پرفروش (Bestseller) | `bestseller` | `desc` |
| جدید (New) | `created_at` | `desc` |
| تخفیف‌دار (Discounted) | `discount` | `desc` |

**`bestseller` algorithm:** Sum `order_items.quantity` for delivered/processing orders in last 90 days, grouped by `product_id`. Fallback to `created_at` if no sales data.

**`discount` algorithm:** `sale_price IS NOT NULL AND sale_price < price`, ordered by discount percentage DESC.

### Pricing display

- Prices in **Toman** (integer).
- Show `sale_price_toman` with strikethrough `price_toman` when on sale.
- `discount_percent = ROUND((1 - sale/price) * 100)`.

### Search

- Full-text on `products.name`, `products.description`, `products.brand`, `skus.code`.
- Persian-normalized search recommended (ی/ي, ک/ك unification).

### Product card fields

- Thumbnail (first image by `sort_order`)
- Name, brand
- Price range if SKUs differ: "از ۳۹۹,۰۰۰ تومان"
- Badges: تخفیف, ناموجود, جدید (created < 30 days)

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Invalid `category_slug` | Ignore filter; return all products |
| `sort=bestseller` with no sales | Fall back to `created_at DESC` |
| Product has variants only (no base price) | Min SKU effective price on card |
| All SKUs out of stock | `is_out_of_stock: true` on card |
| Search returns 0 | Empty state + suggest clearing filters |
| Very long product name | Truncate at 2 lines with ellipsis |
| Concurrent filter + search | AND logic: category AND search term |
| Guest wishlist click | Prompt login; optional local "saved IDs" not in v1 |
| URL share with filters | Deep-link restores state from query string |

---

## Dependencies

### Backend

| Module | Role |
|--------|------|
| `internal/application/storefront` | Product list, search, sort |
| `internal/application/category` | Category tree for filters |
| `internal/application/storefront/wishlist` | Heart icon (authenticated) |

### Existing tables

- `products`, `product_images`, `inventories`, `skus`, `categories`
- `order_items`, `orders` (bestseller sort)

### Frontend

- URL-synced filters (`useSearchParams`)
- Debounced search input
- Grid responsive: 2 col mobile, 3–4 col desktop

---

## Required APIs

### `GET /api/v1/store/products`

Paginated product list for storefront.

**Query parameters**

| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Page number |
| `per_page` | int | 20 | Max 48 |
| `q` | string | — | Search term |
| `category_id` | uuid | — | Filter by category |
| `category_slug` | string | — | Alternative to `category_id` |
| `include_children` | bool | true | Include subcategory products |
| `brand` | string | — | Filter by brand name |
| `sort` | string | `created_at` | `created_at`, `bestseller`, `discount`, `price`, `name` |
| `order` | string | `desc` | `asc`, `desc` |
| `on_sale` | bool | — | When true, only discounted |
| `in_stock` | bool | — | When true, only in-stock |

**Response 200**

```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "ceramic-tile-60x60",
      "name": "کاشی سرامیک ۶۰×۶۰",
      "short_description": "…",
      "thumbnail_url": "https://…",
      "brand": "مرجان",
      "category": { "id": "uuid", "name": "کاشی", "slug": "tiles" },
      "price_toman": 450000,
      "sale_price_toman": 399000,
      "price_from_toman": 399000,
      "price_to_toman": 520000,
      "is_on_sale": true,
      "discount_percent": 11,
      "is_out_of_stock": false,
      "is_new": true,
      "has_variants": true,
      "variant_count": 12
    }
  ],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 156,
    "total_pages": 8
  },
  "filters_applied": {
    "q": null,
    "category_slug": "tiles",
    "sort": "bestseller",
    "order": "desc"
  }
}
```

### `GET /api/v1/store/products/search`

Lightweight typeahead (optional; header search).

**Query:** `q` (min 2 chars), `limit` (default 8, max 20)

**Response 200**

```json
{
  "data": [
    {
      "id": "uuid",
      "slug": "…",
      "name": "…",
      "thumbnail_url": "…",
      "price_from_toman": 399000
    }
  ]
}
```

### `GET /api/v1/store/categories`

Category tree for filter sidebar.

**Query:** `tree=true` (default), `with_products_count=true`

**Response 200**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "مصالح ساختمانی",
      "slug": "building-materials",
      "image_url": "https://…",
      "products_count": 120,
      "children": [
        {
          "id": "uuid",
          "name": "کاشی و سرامیک",
          "slug": "tiles",
          "products_count": 42,
          "children": []
        }
      ]
    }
  ]
}
```

### `GET /api/v1/store/brands` (optional)

Distinct active product brands for filter dropdown.

**Response:** `{ "data": ["مرجان", "ایسیکو", "…"] }`

---

## Database Impact

### Reads

- `products` with joins: `product_images`, `inventories`, `categories`, `skus`
- Aggregation on `order_items` for bestseller sort
- `categories` recursive CTE for tree + counts

### Writes

None on catalog browse.

### Indexes (recommended)

```sql
CREATE INDEX idx_products_status_created ON products (status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_products_sale ON products (status) WHERE sale_price IS NOT NULL;
CREATE INDEX idx_products_category ON products (category_id) WHERE status = 'active';
-- GIN for Persian search (optional)
CREATE INDEX idx_products_name_trgm ON products USING gin (name gin_trgm_ops);
```

### Migration

- Existing schema sufficient for v1
- `000015_indexes` for performance

---

## Validation

### Query params

| Param | Validation |
|-------|------------|
| `page` | `>= 1` |
| `per_page` | `1..48` |
| `q` | Max 200 chars; strip HTML |
| `category_id` | Valid UUID or ignored |
| `sort` | Whitelist enum |
| `order` | `asc` \| `desc` |

Invalid `sort` → `400` with `INVALID_SORT_FIELD`.

---

## Permissions

| Action | Role |
|--------|------|
| List/search products | Public |
| View categories | Public |
| Add to wishlist | Customer (Bearer token) |
| Add to cart | Public (client-side cart) |

---

## State Management

### URL query params (source of truth)

```
/products?category=tiles&q=کاشی&sort=bestseller&page=2
```

| Param | Maps to |
|-------|---------|
| `category` | `category_slug` |
| `q` | search |
| `sort` | active tab |
| `page` | pagination |

### React state

| State | Storage |
|-------|---------|
| Product list | React Query `['products', params]` |
| Search input | Local debounced; sync to URL on submit or debounce end |
| Active sort tab | Derived from URL `sort` |
| Selected categories | URL `category` |
| Wishlist product IDs | React Query `['wishlist']` when authenticated |

### Cart

Add-to-cart writes to `localStorage` key `store_cart_v1` (see checkout spec). No catalog API write.
