# Store Product Detail

> **Route:** `/products/:id` (slug or UUID)  
> **UI:** [store-os-eta.vercel.app/products/:id](https://store-os-eta.vercel.app/products)  
> **Locale:** Persian (fa-IR), RTL

---

## Purpose

Product detail is the conversion page for building materials with **variant axes** (size, color, weight, pattern), image gallery, stock-aware pricing, reviews, Q&A, related products, wishlist, and add-to-cart. Variant selection resolves to a specific SKU with inventory and price.

---

## User Flow

```mermaid
flowchart TD
    A[/products/:id] --> B[GET /store/products/:slug]
    B --> C[Render gallery + variant pickers]
    C --> D[User selects size/color/weight/pattern]
    D --> E[Resolve SKU from selection]
    E --> F{In stock?}
    F -->|Yes| G[Enable Add to Cart]
    F -->|No| H[Disable + show ناموجود]
    G --> I[Add line to localStorage cart]
    C --> J{Tab}
    J -->|Reviews| K[GET reviews + POST if buyer]
    J -->|Q&A| L[GET questions + POST question]
    C --> M[Related products carousel]
    C --> N[Wishlist heart]
    N --> O{Auth?}
    O -->|Yes| P[POST/DELETE wishlist]
    O -->|No| Q[Login prompt]
```

1. Load product by slug (preferred) or UUID.
2. Default variant: first in-stock SKU or first SKU.
3. Gallery updates per variant images if SKU-specific images exist; else product images.
4. Quantity stepper (min 1, max available stock).
5. Tabs: توضیحات (description), نظرات (reviews), پرسش و پاسخ (Q&A).
6. Related products: same category, exclude current.
7. Add to cart includes `sku_id`, variant labels, unit price snapshot.

---

## Business Logic

### Variant matrix

Building materials example attributes:

| Axis (Persian) | Attribute key | Example values |
|----------------|---------------|----------------|
| سایز | `size` | ۳۰×۳۰، ۶۰×۶۰ |
| رنگ | `color` | سفید، بژ |
| وزن | `weight` | ۲۵ کیلو، ۵۰ کیلو |
| طرح | `pattern` | سنگی، مدرن |

- `product_attributes` defines axes; `product_variant_attribute_values` lists allowed values per axis.
- `skus.attributes` JSONB maps axis → value for each valid combination.
- UI disables invalid combinations (gray out unavailable options).

### Pricing

```
effective_price = sku.price_override ?? product.sale_price ?? product.price
price_toman = ROUND(effective_price)  // stored/displayed as Toman
```

### Inventory

- v1: Product-level `inventories.quantity` OR per-SKU stock (future `sku_inventories`).
- `max_quantity` for cart = current stock.
- Show "تنها {n} عدد باقی مانده" when `quantity <= low_stock_threshold`.

### Reviews

- List: `status = 'approved'` only.
- Submit: authenticated customer; **verified buyer** if customer has delivered order containing this product (assumption A from README).
- Guest reviews: not allowed in v1.
- New reviews `status = 'pending'` until admin approves.

### Q&A

- Anyone can ask (name required; email optional but recommended for reply notification).
- Answers visible when `status = 'answered'` and `answer IS NOT NULL`.
- Admin answers via admin API.

### Related products

- Same `category_id`, `status = active`, limit 8, random or `bestseller` within category.

### Wishlist

- Toggle per **product** (not per SKU) in v1.
- Unique `(customer_id, product_id)`.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Invalid slug/id | `404` product not found |
| Draft/archived product | `404` (treat as not found) |
| Partial variant selection | Disable add-to-cart; prompt "لطفاً همه گزینه‌ها را انتخاب کنید" |
| Selected SKU out of stock | Disable cart; suggest another variant |
| Single-SKU product | Hide variant UI or show read-only attributes |
| Price differs per SKU | Update price display on variant change |
| Review pending moderation | Not shown in list; show toast on submit |
| Duplicate wishlist add | Idempotent `201` or `200` |
| Image load failure | Placeholder image |
| Share URL | Canonical slug URL `/products/{slug}` |

---

## Dependencies

### Backend modules

| Module | Role |
|--------|------|
| `internal/application/storefront` | Product detail, related products |
| `internal/application/storefront/review` | List + create reviews |
| `internal/application/storefront/question` | List + create Q&A |
| `internal/application/storefront/wishlist` | Add/remove |
| `internal/application/order` | Verify buyer for reviews |

### Tables

- `products`, `product_images`, `product_attributes`, `product_variant_attribute_values`, `skus`, `inventories`
- `product_reviews`, `product_questions`, `wishlist_items`
- `orders`, `order_items` (verified buyer check)

---

## Required APIs

### `GET /api/v1/store/products/{slugOrId}`

Full product detail for storefront.

**Response 200**

```json
{
  "id": "uuid",
  "slug": "ceramic-tile-60x60",
  "name": "کاشی سرامیک ۶۰×۶۰",
  "description": "…",
  "short_description": "…",
  "brand": "مرجان",
  "category": { "id": "uuid", "name": "کاشی", "slug": "tiles" },
  "images": [
    { "id": "uuid", "url": "https://…", "alt_text": "…", "sort_order": 0 }
  ],
  "variant_axes": [
    {
      "name": "سایز",
      "key": "size",
      "values": ["۳۰×۳۰", "۶۰×۶۰"]
    },
    {
      "name": "رنگ",
      "key": "color",
      "values": ["سفید", "بژ"]
    },
    {
      "name": "وزن",
      "key": "weight",
      "values": ["۲۵ کیلو"]
    },
    {
      "name": "طرح",
      "key": "pattern",
      "values": ["سنگی", "مدرن"]
    }
  ],
  "skus": [
    {
      "id": "uuid",
      "code": "TILE-60-WHT-STN",
      "attributes": { "size": "۶۰×۶۰", "color": "سفید", "weight": "۲۵ کیلو", "pattern": "سنگی" },
      "price_toman": 450000,
      "sale_price_toman": 399000,
      "quantity": 24,
      "is_in_stock": true,
      "is_low_stock": false
    }
  ],
  "default_sku_id": "uuid",
  "price_toman": 450000,
  "sale_price_toman": 399000,
  "is_on_sale": true,
  "seo": {
    "meta_title": "…",
    "meta_description": "…"
  },
  "reviews_summary": {
    "average_rating": 4.5,
    "total_count": 12
  },
  "is_in_wishlist": false
}
```

> `is_in_wishlist` populated when `Authorization: Bearer` present; else `false`.

### `GET /api/v1/store/products/{id}/reviews`

**Query:** `page`, `per_page` (default 10), `sort=rating|created_at`

**Response 200**

```json
{
  "data": [
    {
      "id": "uuid",
      "author_name": "علی ر.",
      "rating": 5,
      "title": "کیفیت عالی",
      "content": "…",
      "created_at": "2026-05-01T12:00:00Z",
      "is_verified_buyer": true
    }
  ],
  "meta": { "page": 1, "per_page": 10, "total": 12, "total_pages": 2 },
  "summary": { "average_rating": 4.5, "total_count": 12 }
}
```

### `POST /api/v1/store/products/{id}/reviews`

**Auth:** Bearer + `customer` role

**Request**

```json
{
  "rating": 5,
  "title": "کیفیت عالی",
  "content": "کاشی بسیار با کیفیت بود."
}
```

**Response 201**

```json
{
  "id": "uuid",
  "status": "pending",
  "message": "نظر شما پس از تأیید نمایش داده می‌شود."
}
```

**Errors:** `403` if not verified buyer (configurable), `409` if already reviewed.

### `GET /api/v1/store/products/{id}/questions`

**Query:** `page`, `per_page`

**Response:** Only `status = answered` items in public list.

```json
{
  "data": [
    {
      "id": "uuid",
      "asker_name": "مریم",
      "question": "برای نمای بیرونی مناسب است؟",
      "answer": "بله، ضد یخ‌زدگی دارد.",
      "answered_at": "2026-05-02T10:00:00Z",
      "created_at": "2026-05-01T09:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 10, "total": 3, "total_pages": 1 }
}
```

### `POST /api/v1/store/products/{id}/questions`

**Auth:** Public

**Request**

```json
{
  "asker_name": "مریم احمدی",
  "asker_email": "maryam@example.com",
  "question": "برای نمای بیرونی مناسب است؟"
}
```

**Response 201**

```json
{
  "id": "uuid",
  "message": "پرسش شما ثبت شد. پاسخ از طریق ایمیل اطلاع‌رسانی می‌شود."
}
```

### `GET /api/v1/store/products/{id}/related`

**Query:** `limit` (default 8)

**Response:** `{ "data": [StoreProductCard] }`

### Wishlist (see `store-wishlist.md`)

| Method | Path |
|--------|------|
| `POST` | `/api/v1/store/account/wishlist` body `{ "product_id": "uuid" }` |
| `DELETE` | `/api/v1/store/account/wishlist/{product_id}` |

---

## Database Impact

### Reads

- Full product graph + SKUs + attributes
- `product_reviews` filtered by `status = 'approved'`
- `product_questions` filtered by `status = 'answered'`
- `wishlist_items` when authenticated

### Writes

| Table | Operation |
|-------|-----------|
| `product_reviews` | INSERT on review submit |
| `product_questions` | INSERT on question submit |

### Migrations

- `000008` / `000009` — SKUs and variant attribute values (exists)
- `000013` — `product_reviews`, `product_questions`, `wishlist_items`
- `000014` — SKU `price_override`, `sale_price_override`

---

## Validation

### Review POST

| Field | Rules |
|-------|-------|
| `rating` | Required, integer 1–5 |
| `title` | Optional, max 255 |
| `content` | Required, 10–2000 chars |

### Question POST

| Field | Rules |
|-------|-------|
| `asker_name` | Required, 2–255 |
| `asker_email` | Optional, valid email |
| `question` | Required, 10–1000 chars |

### Product identifier

- Accept UUID or slug; resolve to single product.

---

## Permissions

| Action | Role |
|--------|------|
| View product detail | Public |
| List reviews / Q&A | Public |
| Submit review | Customer (verified buyer recommended) |
| Submit question | Public |
| Wishlist toggle | Customer |
| Add to cart | Public (client-side) |

---

## State Management

### Per-page React state

| State | Description |
|-------|-------------|
| `selectedAttributes` | `{ size, color, weight, pattern }` |
| `resolvedSkuId` | Derived from selection |
| `quantity` | Default 1, clamped to stock |
| `activeTab` | `description \| reviews \| qa` |
| `activeImageIndex` | Gallery index |
| `wishlistOptimistic` | Optimistic UI on toggle |

### Derived values

```typescript
// Pseudocode
const selectedSku = skus.find(sku =>
  Object.entries(selectedAttributes).every(([k, v]) => sku.attributes[k] === v)
);
const displayPrice = selectedSku?.sale_price_toman ?? selectedSku?.price_toman;
```

### Persistence

| Data | Storage |
|------|---------|
| Product detail | React Query `['product', slug]` |
| Cart add | `localStorage.store_cart_v1` |
| Wishlist | Server + invalidate `['wishlist']` |
| Recently viewed | `localStorage.recent_products` (optional, max 10) |

### Cart line item shape (on add)

```json
{
  "product_id": "uuid",
  "sku_id": "uuid",
  "name": "کاشی سرامیک ۶۰×۶۰",
  "sku_code": "TILE-60-WHT-STN",
  "variant_label": "۶۰×۶۰ · سفید · سنگی",
  "thumbnail_url": "https://…",
  "unit_price_toman": 399000,
  "quantity": 2
}
```
