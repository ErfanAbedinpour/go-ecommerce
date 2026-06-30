# API Coverage Report

> **Audit date:** 2026-06-30  
> **Last updated:** 2026-06-30 (post contract normalization)

---

## Summary

| Metric | Initial audit | After endpoints | After contract pass |
|--------|--------------:|----------------:|--------------------:|
| **Frontend features / user interactions audited** | 94 | 94 | 94 |
| **Backend API endpoints (router paths)** | ~100 | ~118 | ~118 |
| **Fully implemented (frontend ↔ API contract match)** | 38 (40%) | 52 (55%) | 64 (68%) |
| **Partially implemented (endpoint exists, contract gaps)** | 41 (44%) | 37 (39%) | 25 (27%) |
| **Missing (no endpoint or unusable for UI)** | 15 (16%) | 5 (5%) | 5 (5%) |
| **Overall coverage (Full + Partial endpoint existence)** | **84%** | **95%** | **95%** |
| **Contract-complete coverage (Full only)** | **40%** | **55%** | **68%** |

### By application surface

| Surface | Features | Full | Partial | Missing |
|---------|----------|-----:|--------:|--------:|
| Customer store ([Store OS](https://store-os-eta.vercel.app/)) | 52 | 24 | 23 | 5 |
| Admin panel ([shop-panel-react](https://shop-panel-react.vercel.app/)) | 42 | 28 | 14 | 0 |

### Headline findings

1. **Contract normalization pass complete** — Catalog filters/sort, homepage aggregate (`categories`, `blog_teaser`, stats), product `?include=reviews_summary,wishlist`, wishlist/blog field aliases, slug-based engagement paths, caching, and contact rate limits are implemented.
2. **Remaining contract gaps** — Per-SKU inventory projection, account order `*_toman` fields, checkout `payment_url`, server-side shipping in preview, and product `seo` block.
3. **Checkout payment is partially unblocked** — Callback route and settings exist; PSP redirect URL on place-order still pending.
4. **Admin panel is wire-ready** for core commerce; remaining gaps are CMS pagination, video upload for hero, and admin blog field renames.

> **Storefront UI:** [store-os-eta.vercel.app](https://store-os-eta.vercel.app/)  
> **Admin UI:** [shop-panel-react.vercel.app](https://shop-panel-react.vercel.app/)  
> **Backend source of truth:** `internal/interfaces/http/router.go` + `docs/swagger/swagger.yaml`  
> **Swagger UI:** `http://localhost:8080/swagger/index.html`

---

## Coverage Table

### Authentication (both apps)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Admin login | ✅ | **Full** | `POST /api/v1/auth/login` |
| Admin signup | ✅ | **Full** | `POST /api/v1/auth/signup` |
| Customer signup/login | ✅ | **Full** | Same auth routes; role from token |
| Token refresh | ✅ | **Full** | `POST /api/v1/auth/refresh` |
| Logout | ✅ | **Full** | `POST /api/v1/auth/logout` (auth required) |
| Current user (`/auth/me`) | ✅ | **Full** | `GET /api/v1/auth/me` |
| Forgot password | ✅ | **Full** | `POST /api/v1/auth/forgot-password` |
| Reset password | ✅ | **Full** | `POST /api/v1/auth/reset-password` |

### Customer store — navigation & layout

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Mega-menu categories | ✅ | **Full** | Embedded in `GET /store/homepage` + `GET /store/categories` |
| Store header navigation | ✅ | **Full** | `GET /store/navigation` |
| Theme / CSS variables | ✅ | **Full** | `GET /store/theme` |
| Site settings (footer, contact) | ✅ | **Partial** | `GET /store/settings` — no `whatsapp`/`telegram` in social |
| Mobile bottom nav | — | **N/A** | Client-side routing only |

### Customer store — homepage (`/`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Hero video + CTAs | ✅ | **Partial** | In `GET /store/homepage`; missing `poster_url` |
| Category grid | ✅ | **Full** | `categories[]` embedded in homepage |
| Product carousels (3 tabs) | ✅ | **Full** | `product_slides`; `featured` mapped to `new` |
| Pro banners | ✅ | **Full** | In homepage aggregate |
| Partner brands | ✅ | **Full** | In homepage aggregate |
| Stats counters | ✅ | **Partial** | `products_count`, `customers_count`, `delivered_orders_count`, `years_experience` |
| FAQ accordion | ✅ | **Full** | In homepage aggregate |
| Contact form | ✅ | **Partial** | `POST /store/contact`; rate-limited; response shape differs |
| Customer testimonials | ✅ | **Full** | `testimonials` in homepage |
| Blog teaser (latest 3) | ✅ | **Full** | `blog_teaser.posts[]` in homepage |
| Search shortcut | ✅ | **Full** | `GET /store/products/search` |

### Customer store — catalog (`/products`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Product grid | ✅ | **Partial** | `GET /store/products` — card fields still incomplete |
| Text search | ✅ | **Full** | `q` on list + `GET /store/products/search` |
| Category filter by slug | ✅ | **Full** | `category_slug` + `include_children` |
| Sort: bestseller | ✅ | **Full** | 90-day sales aggregation |
| Sort: newest | ✅ | **Full** | `newest` → `created_at DESC` |
| Sort: discounted | ✅ | **Full** | `discounted` alias mapped |
| Brand filter | ✅ | **Full** | `brand` query param |
| On-sale / in-stock filters | ✅ | **Full** | `on_sale`, `in_stock` params |
| Pagination | ✅ | **Full** | `page`, `per_page` + meta |
| Product card badges | ✅ | **Partial** | Missing `is_new`, `has_variants`, `price_from_toman` |
| Add to wishlist (heart) | ✅ | **Partial** | `POST /store/account/wishlist`; `GET .../wishlist/ids` for badge |
| Add to cart | — | **N/A** | Client-side `localStorage` cart (v1 design) |

### Customer store — product detail (`/products/:id`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Load by slug or UUID | ✅ | **Full** | `GET /store/products/{slugOrId}` |
| Image gallery | ✅ | **Full** | `images[]` in response |
| Variant / SKU selection | ✅ | **Partial** | `skus[]` lacks per-SKU price/stock |
| Reviews tab | ✅ | **Partial** | Slug or UUID in path; no `summary` in list |
| Review summary | ✅ | **Full** | Separate endpoint or `?include=reviews_summary` on product |
| Submit review | ✅ | **Partial** | Guest allowed; docs expect customer-only |
| Q&A tab | ✅ | **Full** | Slug or UUID in path |
| Related products | ✅ | **Full** | `GET /store/products/{id}/related` |
| Wishlist state on detail | ✅ | **Full** | `?include=wishlist` with optional auth |
| SEO meta | ❌ | **Missing** | No `seo` block in product detail |

### Customer store — checkout (`/checkout`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Cart review (step 1) | — | **N/A** | Client-side cart |
| Checkout preview | ✅ | **Partial** | `POST /store/checkout/preview` |
| Coupon validation | ✅ | **Partial** | `POST /store/coupons/validate` |
| Shipping address (step 2) | ✅ | **Partial** | Checkout body + profile addresses |
| Shipping methods | ✅ | **Full** | `GET /store/checkout/shipping-methods?city=` |
| Checkout settings | ✅ | **Full** | `GET /store/settings/checkout` |
| Payment (step 3) | ✅ | **Partial** | `POST /store/checkout`; no `payment_url` |
| Payment gateway callback | ✅ | **Full** | `POST /store/checkout/payment/callback` |
| Guest checkout | ✅ | **Full** | Creates guest customer |
| Order confirmation email | ✅ | **Partial** | Mailer wired; no `payment_url` in response |

### Customer store — account

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Profile view/edit | ✅ | **Full** | `GET/PUT /store/account/profile` |
| Saved addresses | ✅ | **Full** | Part of profile update |
| Order history list | ✅ | **Partial** | `GET /store/account/orders`; no `status` filter |
| Order detail | ✅ | **Partial** | Uses `float64` amounts, not `*_toman`; no `timeline` |
| Wishlist page | ✅ | **Partial** | `GET /store/account/wishlist`; `created_at` not `added_at` |
| Wishlist badge / IDs | ✅ | **Full** | `GET .../wishlist/count`, `GET .../wishlist/ids` |
| Auth guard redirect | ✅ | **Full** | `GET /auth/me` |

### Customer store — blog (`/blog`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Post listing | ✅ | **Full** | `GET /store/blog/posts` + alias `GET /store/blog` |
| Category filter | ✅ | **Full** | `category_id` or `category_slug` + `include_children` |
| Post detail | ✅ | **Partial** | `GET /store/blog/posts/{slug}` |
| Comments list/submit | ✅ | **Partial** | UUID `postId`; slug alias `GET /blog/{slug}/comments` |
| Categories sidebar | ✅ | **Partial** | No `posts_count` per category |

### Customer store — about (`/about`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| About page content | ✅ | **Full** | `GET /store/about` |
| Contact form | ✅ | **Partial** | `POST /store/contact` with `source=about` |

### Admin panel — dashboard (`/`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| KPI stat cards | ✅ | **Full** | `GET /admin/dashboard/stats` |
| Revenue chart | ✅ | **Full** | `GET /admin/dashboard/revenue` |
| Recent orders table | ✅ | **Full** | `GET /admin/dashboard/recent-orders` |
| Low stock table | ✅ | **Full** | `GET /admin/dashboard/low-stock` |
| Featured products table | ✅ | **Full** | `GET /admin/dashboard/featured-products` |

### Admin panel — products

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Product list + filters | ✅ | **Full** | Status, category, brand, stock_level |
| Product stats cards | ✅ | **Full** | `GET /admin/products/stats` |
| Product search | ✅ | **Full** | `GET /admin/products/search` |
| Create / edit product | ✅ | **Partial** | SKU matrix works; per-SKU inventory missing |
| Delete product | ✅ | **Full** | Soft delete |
| Inventory update | ✅ | **Partial** | `adjustment_reason` not persisted |
| Image upload | ✅ | **Partial** | Images only; no video for hero |
| Catalog settings (categories, brands, attrs) | ✅ | **Full** | Full CRUD under admin routes |

### Admin panel — orders

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Order list + date filter | ✅ | **Full** | `from`, `to` query params implemented |
| Order detail | ✅ | **Full** | Timeline, notes, payment fields |
| Status update | ✅ | **Full** | `PATCH .../status` |
| Cancel / refund | ✅ | **Full** | |
| Invoice | ✅ | **Full** | `GET .../invoice` |
| Manual order create | ✅ | **Full** | `POST /admin/orders` |

### Admin panel — customers, coupons, settings

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Customer list/detail/edit/delete | ✅ | **Full** | |
| Customer order history | ✅ | **Full** | |
| Coupon CRUD + activate/deactivate | ✅ | **Full** | |
| Site / contact / social / SEO settings | ✅ | **Full** | |
| Admin navigation | ✅ | **Full** | `GET/PUT /admin/navigation` |

### Admin panel — context / storefront CMS

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Hero section | ✅ | **Partial** | CRUD exists; video upload not supported via uploads API |
| Product slides | ✅ | **Partial** | Item CRUD; no bulk slide update |
| Pro banners | ✅ | **Full** | |
| Partner brands | ✅ | **Partial** | No list pagination |
| Homepage reviews | ✅ | **Partial** | No list pagination |
| FAQ section | ✅ | **Partial** | Split image + items API vs single PUT |
| Contact section image | ✅ | **Full** | |
| Storefront navigation | ✅ | **Full** | |

### Admin panel — themes, blog, contact inbox

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Theme marketplace | ✅ | **Partial** | No pagination on list |
| Theme purchase | ✅ | **Full** | |
| Style customization | ✅ | **Full** | `GET/PUT /admin/store-style` |
| Blog posts CRUD | ✅ | **Partial** | `summary` not `excerpt`; no `read_time_minutes` |
| Blog categories CRUD | ✅ | **Full** | |
| Comment moderation | ✅ | **Full** | `PATCH .../status` + `.../approve` + `.../reject` aliases |
| Contact inbox | ✅ | **Full** | Stats, list, detail, status, read/archive aliases |
| Review / Q&A moderation | ✅ | **Full** | |

---

## Missing Endpoints

All routes catalogued in [missing-endpoints.md](./missing-endpoints.md) are **implemented**. Remaining **feature gaps** (not separate routes):

| Gap | Blocks | Suggested fix |
|-----|--------|---------------|
| `payment_url` on checkout response | PSP redirect before callback | Payment provider integration |
| Product `seo` block | Meta tags on detail | Extend `ProductDetail` DTO |
| Per-SKU price/stock projection | Variant selector on detail | Extend SKU DTO + inventory model |
| `hero.poster_url` | Hero fallback image | CMS field on hero config |

---

## Wrong Contracts

See [api-contract-diff.md](./api-contract-diff.md). Top **open** mismatches:

1. **Toman integer fields** — account order detail still uses `float64` `total` while catalog/wishlist use `*_toman`.
2. **Checkout preview** — client-supplied `shipping_amount`/`tax_amount` vs server-side calculation.
3. **Product variants** — per-SKU `price_toman`/`quantity`, `variant_axes`, `default_sku_id`.
4. **`payment_url`** — missing on place-order response (callback route exists).

**Resolved since contract pass:** catalog sort/filters, homepage aggregate, product `?include=`, wishlist `added_at`/`_*toman`/idempotent add, blog field aliases, engagement slug paths, contact rate limit, Redis cache on categories/navigation/theme.

---

## Missing Fields (response-level)

| Endpoint | Missing fields |
|----------|----------------|
| `GET /store/homepage` | `hero.poster_url` |
| `GET /store/products` (card) | `short_description`, `category`, `is_new`, `has_variants`, `variant_count`, `price_from_toman`, `price_to_toman`, `filters_applied` |
| `GET /store/products/{id}` | `category`, `default_sku_id`, `variant_axes`, per-SKU `price_toman`/`quantity`, `seo` |
| `GET /store/account/orders/{id}` | `*_toman` integers, `variant_label` on items, `timeline[]` |
| `POST /store/checkout` | `payment_url`, `expires_at` |
| Admin blog post | `read_time_minutes`, `excerpt` (has `summary`), `archived` status |

---

## High Priority Issues (blockers)

1. **SKU variant pricing on detail** — variant selector cannot show per-SKU price/stock.
2. **`payment_url` missing on checkout** — callback exists but client cannot redirect to PSP without URL in place-order response.
3. **Account order `*_toman` fields** — store order detail still uses floats.

**Resolved since contract pass:** catalog sort/filters, homepage aggregate, wishlist/blog aliases, product includes, engagement slug paths.

---

## Recommendations (prioritized)

| Priority | Action | Impact | Status |
|----------|--------|--------|--------|
| P0 | Add `payment_url` to checkout response (callback already exists) | Unblocks online payment redirect | Partial |
| P1 | Align account order fields to `*_toman` integers | Prevents frontend adapter bugs | Open |
| P1 | Per-SKU price/stock on product detail + `variant_axes` projection | Unblocks variant selector | Open |
| P2 | Video upload support for hero (`context=hero` on uploads) | Hero video management | Open |
| P3 | Server-side shipping calculation in preview | Checkout accuracy | Open |
| P3 | Persian error messages for coupon/checkout validation | RTL UX | Open |
| P3 | Cache invalidation on category/theme CMS updates | Stale public cache | Open |
| — | Catalog sort/filters, homepage aggregate, product includes, wishlist/blog aliases, contact rate limit | — | **Done** |
| — | Account profile, about, navigation, related products, shipping, callback, wishlist shortcuts, blog/contact admin aliases | — | **Done** |

---

## Methodology

- **Frontend:** Inspected live [Store OS](https://store-os-eta.vercel.app/) pages (home, products, checkout) and [Admin Dashboard](https://shop-panel-react.vercel.app/); cross-referenced `docs/features/*.md` for expected contracts.
- **Backend:** Enumerated all routes in `router.go`; validated handlers, DTOs, and repository logic against Swagger and feature specs.
- **Status definitions:**
  - **Full** — endpoint exists and response/request match documented frontend needs.
  - **Partial** — endpoint exists but missing fields, params, or behavioral differences.
  - **Missing** — no endpoint or endpoint cannot serve the feature.
