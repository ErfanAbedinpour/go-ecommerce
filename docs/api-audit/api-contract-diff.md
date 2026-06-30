# API Contract Differences

> Frontend expectations from `docs/features/` and live UI vs actual backend implementation.  
> **Last updated:** 2026-06-30 (post contract normalization)

---

## Open issues

### 1. Product detail variants

**Frontend expects:** `variant_axes[]`, per-SKU `price_toman`/`quantity`, `default_sku_id`, `seo` block

**Backend currently:** `skus[]` with `id`, `code`, `attributes` only; product-level inventory. `?include=reviews_summary,wishlist` **implemented**.

**Priority:** P0 | **Breaking?** No

---

### 2. Order detail (customer)

**Frontend expects:** `*_toman` integers, `variant_label`, `timeline[]`

**Backend currently:** `float64` amounts, no timeline in store DTO.

**Priority:** P1 | **Breaking?** Yes if renaming floats

---

### 3. Checkout preview

**Frontend expects:** `shipping_method`, `shipping_city` → server computes shipping/tax.

**Backend currently:** Client sends `shipping_amount`, `tax_amount`. Shipping **methods** endpoint exists; preview does not use it yet.

**Priority:** P1 | **Breaking?** Yes

---

### 4. Checkout place-order response

**Frontend expects:** `payment_url`, `expires_at`

**Backend currently:** Order metadata only; **`POST /checkout/payment/callback` implemented** but no PSP redirect URL on checkout response.

**Priority:** P0 | **Breaking?** No

---

### 5. Admin blog fields

**Frontend expects:** `excerpt`, `read_time_minutes`, `archived` status

**Backend currently:** `summary`, no read time, `draft|published` only. Approve/reject route aliases implemented.

**Priority:** P2 | **Breaking?** Yes if renaming

---

---

## Resolved since initial audit

| # | Topic | Resolution |
|---|--------|------------|
| — | Homepage aggregate | `categories[]`, `blog_teaser`, extended `stats`, `featured`→`new` slide mapping |
| — | Catalog filters & sort | `category_slug`, `include_children`, `brand`, `on_sale`, `in_stock`, all documented sort values |
| — | Product detail includes | `?include=reviews_summary,wishlist` with optional auth |
| — | Reviews/Q&A slug paths | `productref.ResolveID` on engagement routes |
| — | Wishlist shape | `added_at` alias, `*_toman` on nested product, idempotent add (200) |
| — | Blog store fields | `excerpt` + `cover_image_url` aliases alongside legacy fields |
| — | Theme purchase | Idempotent — returns existing purchase |
| — | Store caching | Redis 5 min on homepage, products, categories, navigation, theme |
| — | Contact rate limit | 3 req/min per IP on `POST /store/contact` |
| — | Catalog indexes | Migration `000016` for bestseller sort |
| — | Account profile | `GET/PUT /api/v1/store/account/profile` with addresses |
| — | About page | `GET /api/v1/store/about` |
| — | Public navigation | `GET /api/v1/store/navigation` |
| — | Related products | `GET /api/v1/store/products/{id}/related` |
| — | Product search autocomplete | `GET /api/v1/store/products/search` |
| — | Public brands | `GET /api/v1/store/brands` |
| — | Shipping methods | `GET /api/v1/store/checkout/shipping-methods` |
| — | Checkout settings | `GET /api/v1/store/settings/checkout` |
| — | Payment callback | `POST /api/v1/store/checkout/payment/callback` |
| — | Wishlist shortcuts | `GET .../wishlist/ids`, `GET .../wishlist/count` |
| — | Blog path aliases | `GET /store/blog`, `GET /store/blog/{slug}/comments` |
| — | Contact inbox stats | `GET /api/v1/admin/contact-messages/stats` |
| — | Contact read/archive aliases | `PATCH .../read`, `PATCH .../archive` |
| — | Comment approve/reject aliases | `PATCH .../approve`, `PATCH .../reject` |

See [missing-endpoints.md](./missing-endpoints.md) for route details.
