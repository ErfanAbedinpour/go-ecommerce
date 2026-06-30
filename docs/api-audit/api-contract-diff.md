# API Contract Differences

> Frontend expectations from `docs/features/` and live UI vs actual backend implementation.

---

## 1. Homepage aggregate

**Frontend expects** (`GET /api/v1/store/homepage`): `categories[]`, `blog_teaser.posts[]`, full `stats` (`customers_count`, `delivered_orders_count`, `years_experience`), `hero.poster_url`, slide type `new`.

**Backend currently returns:** `product_slides`, `stats.products_count` only; slide type `featured` instead of `new`.

**Missing:** `categories`, `blog_teaser`, extended stats, `poster_url`

**Recommended change:** Extend `BuildHomepage()` in `storecontent/homepage.go`.

**Priority:** P0 | **Breaking?** No

---

## 2. Catalog sort parameters

**Frontend expects:** `sort=bestseller|newest|discounted`

**Backend currently:** Only `discount`, `price` ASC, `name`, default `created_at`. Swagger lists `bestseller|newest|discounted|price_asc|price_desc` but repo ignores most.

**Missing:** Bestseller aggregation, param aliases, `price_desc`

**Recommended change:** Sort mapping in handler + repo implementation.

**Priority:** P0 | **Breaking?** No

---

## 3. Category filter by slug

**Frontend expects:** `?category_slug=tiles&include_children=true`

**Backend currently:** `category_id` UUID only.

**Missing:** `category_slug`, `include_children`, `brand`, `on_sale`, `in_stock`

**Priority:** P0 | **Breaking?** No

---

## 4. Product detail variants

**Frontend expects:** `variant_axes[]`, per-SKU `price_toman`/`quantity`, `default_sku_id`, `is_in_wishlist`, `reviews_summary`

**Backend currently:** `skus[]` with `id`, `code`, `attributes` only; product-level inventory.

**Priority:** P0 | **Breaking?** No

---

## 5. Account profile

**Frontend expects:** `GET/PUT /api/v1/store/account/profile` with addresses.

**Backend currently:** Endpoints do not exist.

**Priority:** P0 | **Breaking?** No (new)

---

## 6. Order detail (customer)

**Frontend expects:** `*_toman` integers, `variant_label`, `timeline[]`

**Backend currently:** `float64` amounts, no timeline in store DTO.

**Priority:** P1 | **Breaking?** Yes if renaming floats

---

## 7. Checkout preview

**Frontend expects:** `shipping_method`, `shipping_city` → server computes shipping/tax.

**Backend currently:** Client sends `shipping_amount`, `tax_amount`.

**Priority:** P1 | **Breaking?** Yes

---

## 8. Checkout response

**Frontend expects:** `payment_url`, `expires_at`

**Backend currently:** Order metadata only, no PSP redirect.

**Priority:** P0 | **Breaking?** No

---

## 9. Wishlist shape

**Frontend expects:** `added_at`, nested `product.price_toman`

**Backend currently:** `created_at`, `product.price` (float)

**Priority:** P1 | **Breaking?** No

---

## 10. Blog paths and fields

**Frontend expects:** `/store/blog`, fields `excerpt`, `cover_image_url`

**Backend currently:** `/store/blog/posts`, fields `summary`, `featured_image`

**Priority:** P2 | **Breaking?** No with aliases

---

## 11. About page

**Frontend expects:** `GET /api/v1/store/about`

**Backend currently:** Missing.

**Priority:** P1 | **Breaking?** No

---

## 12. Public navigation

**Frontend expects:** `GET /api/v1/store/navigation`

**Backend currently:** Admin-only `GET /admin/storefront/navigation`

**Priority:** P1 | **Breaking?** No

---

## 13. Admin blog fields

**Frontend expects:** `excerpt`, `read_time_minutes`, `archived` status

**Backend currently:** `summary`, no read time, `draft|published` only

**Priority:** P2 | **Breaking?** Yes if renaming

---

## 14. Contact inbox stats

**Frontend expects:** `GET /admin/contact-messages/stats`

**Backend currently:** Missing.

**Priority:** P2 | **Breaking?** No

---

## 15. Reviews path param

**Frontend uses slug in URL;** engagement APIs require product UUID.

**Priority:** P1 | **Breaking?** No
