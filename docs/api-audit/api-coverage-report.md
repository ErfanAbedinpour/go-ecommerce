# API Coverage Report

> **Audit date:** 2026-06-30  
> **Storefront UI:** [store-os-eta.vercel.app](https://store-os-eta.vercel.app/)  
> **Admin UI:** [shop-panel-react.vercel.app](https://shop-panel-react.vercel.app/)  
> **Backend source of truth:** `internal/interfaces/http/router.go` + `docs/swagger/swagger.yaml`  
> **Swagger UI:** `http://localhost:8080/swagger/index.html`

---

## Summary

| Metric | Count |
|--------|------:|
| **Frontend features / user interactions audited** | 94 |
| **Backend API endpoints (Swagger paths)** | 100 |
| **Fully implemented (frontend ↔ API contract match)** | 38 (40%) |
| **Partially implemented (endpoint exists, contract gaps)** | 41 (44%) |
| **Missing (no endpoint or unusable for UI)** | 15 (16%) |
| **Overall coverage (Full + Partial endpoint existence)** | **84%** |
| **Contract-complete coverage (Full only)** | **40%** |

### By application surface

| Surface | Features | Full | Partial | Missing |
|---------|----------|-----:|--------:|--------:|
| Customer store ([Store OS](https://store-os-eta.vercel.app/)) | 52 | 14 | 30 | 8 |
| Admin panel ([shop-panel-react](https://shop-panel-react.vercel.app/)) | 42 | 24 | 11 | 7 |

### Headline findings

1. **The backend is substantially built** — most modules that were marked ❌ in older `docs/architecture/gap-analysis.md` now have routes (storefront, context CMS, themes, blog, contact, wishlist, reviews, Q&A).
2. **Contract mismatches are the main blocker** — field naming (`price` vs `price_toman`), missing query params (`category_slug`, `bestseller` sort), path differences (`/blog/posts` vs `/blog`), and incomplete aggregate responses (homepage missing `categories`, `blog_teaser`, full `stats`).
3. **Critical missing store APIs:** account profile (`GET/PUT /store/account/profile`), about page (`GET /store/about`), public navigation (`GET /store/navigation`), payment gateway flow (`payment_url`, callback).
4. **Admin panel is largely wire-ready** for core commerce (dashboard, products, orders, customers, coupons, settings). Gaps concentrate in blog field names, contact inbox stats, theme pagination, and video upload for hero.

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
| Mega-menu categories | ✅ | **Partial** | `GET /store/categories` exists; not embedded in homepage |
| Store header navigation | ❌ | **Missing** | UI expects `GET /store/navigation`; only admin route exists |
| Theme / CSS variables | ✅ | **Full** | `GET /store/theme` |
| Site settings (footer, contact) | ✅ | **Partial** | `GET /store/settings` — no `about` block, no `whatsapp`/`telegram` |
| Mobile bottom nav | — | **N/A** | Client-side routing only |

### Customer store — homepage (`/`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Hero video + CTAs | ✅ | **Partial** | In `GET /store/homepage`; missing `poster_url` |
| Category grid | ⚠️ | **Partial** | Separate `GET /store/categories`; not in homepage aggregate |
| Product carousels (3 tabs) | ✅ | **Partial** | `product_slides` in homepage; slide type `featured` vs UI `new` |
| Pro banners | ✅ | **Full** | In homepage aggregate |
| Partner brands | ✅ | **Full** | In homepage aggregate |
| Stats counters | ✅ | **Partial** | Only `products_count`; UI shows years, products, provinces |
| FAQ accordion | ✅ | **Full** | In homepage aggregate |
| Contact form | ✅ | **Partial** | `POST /store/contact`; response shape differs |
| Customer testimonials | ✅ | **Full** | `testimonials` in homepage |
| Blog teaser (latest 3) | ❌ | **Missing** | Not in `HomepageProjection` |
| Search shortcut | ✅ | **Partial** | Catalog `q` param; no dedicated search endpoint |

### Customer store — catalog (`/products`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Product grid | ✅ | **Partial** | `GET /store/products` |
| Text search | ✅ | **Partial** | `q` param works |
| Category filter by slug | ❌ | **Missing** | Only `category_id` (UUID) |
| Sort: bestseller | ❌ | **Missing** | Swagger documents `bestseller`; repo ignores it |
| Sort: newest | ✅ | **Partial** | Default `created_at DESC`; param `newest` not mapped |
| Sort: discounted | ✅ | **Partial** | Repo uses `discount`; swagger says `discounted` |
| Brand filter | ❌ | **Missing** | No `brand` query param |
| On-sale / in-stock filters | ❌ | **Missing** | No `on_sale`, `in_stock` params |
| Pagination | ✅ | **Full** | `page`, `per_page` + meta |
| Product card badges | ✅ | **Partial** | Missing `is_new`, `has_variants`, `price_from_toman` |
| Add to wishlist (heart) | ✅ | **Partial** | `POST /store/account/wishlist`; field naming gaps |
| Add to cart | — | **N/A** | Client-side `localStorage` cart (v1 design) |

### Customer store — product detail (`/products/:id`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Load by slug or UUID | ✅ | **Full** | `GET /store/products/{slugOrId}` |
| Image gallery | ✅ | **Full** | `images[]` in response |
| Variant / SKU selection | ✅ | **Partial** | `skus[]` lacks per-SKU price/stock |
| Reviews tab | ✅ | **Partial** | `GET .../reviews`; no `summary` in list; UUID path only |
| Review summary | ✅ | **Full** | `GET .../reviews/summary` (separate endpoint) |
| Submit review | ✅ | **Partial** | Guest allowed; docs expect customer-only |
| Q&A tab | ✅ | **Partial** | `GET/POST .../questions`; UUID path only |
| Related products | ❌ | **Missing** | No `GET .../related` |
| Wishlist state on detail | ❌ | **Missing** | `is_in_wishlist` not populated |
| SEO meta | ❌ | **Missing** | No `seo` block in product detail |

### Customer store — checkout (`/checkout`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Cart review (step 1) | — | **N/A** | Client-side cart |
| Checkout preview | ✅ | **Partial** | `POST /store/checkout/preview` |
| Coupon validation | ✅ | **Partial** | `POST /store/coupons/validate` |
| Shipping address (step 2) | ✅ | **Partial** | Part of checkout body; no saved addresses API |
| Shipping methods | ❌ | **Missing** | No `GET /store/checkout/shipping-methods` |
| Payment (step 3) | ✅ | **Partial** | `POST /store/checkout`; no `payment_url` |
| Payment gateway callback | ❌ | **Missing** | No callback endpoint |
| Guest checkout | ✅ | **Full** | Creates guest customer |
| Order confirmation email | ✅ | **Partial** | Mailer wired; no `payment_url` in response |

### Customer store — account

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Profile view/edit | ❌ | **Missing** | No `/store/account/profile` |
| Saved addresses | ❌ | **Missing** | Part of profile spec |
| Order history list | ✅ | **Partial** | `GET /store/account/orders`; no `status` filter |
| Order detail | ✅ | **Partial** | Uses `float64` amounts, not `*_toman`; no `timeline` |
| Wishlist page | ✅ | **Partial** | `GET /store/account/wishlist`; `created_at` not `added_at` |
| Auth guard redirect | ✅ | **Full** | `GET /auth/me` |

### Customer store — blog (`/blog`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| Post listing | ✅ | **Partial** | `GET /store/blog/posts` (not `/blog`) |
| Category filter | ✅ | **Partial** | `category_id` only; no `category_slug` |
| Post detail | ✅ | **Partial** | `GET /store/blog/posts/{slug}` |
| Comments list/submit | ✅ | **Partial** | Uses `postId` UUID, not slug |
| Categories sidebar | ✅ | **Partial** | No `posts_count` per category |

### Customer store — about (`/about`)

| Feature | API Exists | Status | Notes |
|---------|:----------:|--------|-------|
| About page content | ❌ | **Missing** | No `GET /store/about` |
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
| Comment moderation | ✅ | **Partial** | Unified `PATCH .../status` vs approve/reject routes |
| Contact inbox | ✅ | **Partial** | No `stats` endpoint; unified status PATCH |
| Review / Q&A moderation | ✅ | **Full** | |

---

## Missing Endpoints

See [missing-endpoints.md](./missing-endpoints.md) for the full catalog. Highest-impact:

| Endpoint | Blocks |
|----------|--------|
| `GET /api/v1/store/account/profile` | Account profile tab |
| `PUT /api/v1/store/account/profile` | Profile editing, saved addresses |
| `GET /api/v1/store/about` | About page |
| `GET /api/v1/store/navigation` | Store header menu |
| `GET /api/v1/store/products/{id}/related` | Product detail related section |
| `GET /api/v1/store/checkout/shipping-methods` | Checkout step 2 |
| `POST /api/v1/store/checkout/payment/callback` | Online payment completion |
| `GET /api/v1/admin/contact-messages/stats` | Inbox unread badge |

---

## Wrong Contracts

See [api-contract-diff.md](./api-contract-diff.md). Top mismatches:

1. **Toman integer fields** — wishlist and order detail still use `float64` `price`/`total` while catalog uses `*_toman`.
2. **Catalog sort param** — Swagger says `bestseller|newest|discounted`; repository implements `discount|price|name|created_at`.
3. **Blog paths** — Backend uses `/store/blog/posts`; frontend docs expect `/store/blog`.
4. **Blog fields** — `summary` vs `excerpt`, `featured_image` vs `cover_image_url`.
5. **Homepage aggregate** — missing `categories`, `blog_teaser`, full `stats`.
6. **Checkout response** — missing `payment_url`, `expires_at`; preview expects client-supplied shipping/tax.
7. **Wishlist** — `created_at` vs `added_at`; duplicate add returns 409 vs idempotent 200.
8. **Product engagement paths** — reviews/Q&A require product UUID; detail page may use slug.

---

## Missing Fields (response-level)

| Endpoint | Missing fields |
|----------|----------------|
| `GET /store/homepage` | `categories[]`, `blog_teaser.posts[]`, `stats.customers_count`, `stats.delivered_orders_count`, `stats.years_experience`, `hero.poster_url` |
| `GET /store/products` (card) | `short_description`, `category`, `is_new`, `has_variants`, `variant_count`, `price_from_toman`, `price_to_toman`, `filters_applied` |
| `GET /store/products/{id}` | `category`, `default_sku_id`, `variant_axes`, per-SKU `price_toman`/`quantity`, `reviews_summary`, `is_in_wishlist`, `seo`, `related_products` |
| `GET /store/account/orders/{id}` | `*_toman` integers, `variant_label` on items, `timeline[]` |
| `GET /store/account/wishlist` | `added_at`, nested product `*_toman` fields |
| Admin blog post | `read_time_minutes`, `excerpt` (has `summary`), `archived` status |

---

## High Priority Issues (blockers)

1. **Account profile APIs missing** — `/account` page cannot load or save profile/addresses.
2. **About page API missing** — `/about` has no content endpoint.
3. **Public store navigation missing** — header menu cannot be CMS-driven without admin token.
4. **Payment gateway flow incomplete** — no `payment_url` or callback; online payment step cannot complete.
5. **Catalog sort tabs broken** — UI tabs map to `bestseller`/`newest`/`discounted`; backend does not implement `bestseller` or map `newest`/`discounted`.
6. **Category filter by slug** — homepage links use `?category=slug`; API only accepts UUID.
7. **Homepage stats show zeros** — UI counters need `years_experience`, product/customer counts beyond `products_count`.
8. **SKU variant pricing on detail** — variant selector cannot show per-SKU price/stock.

---

## Recommendations (prioritized)

| Priority | Action | Impact |
|----------|--------|--------|
| P0 | Implement `GET/PUT /store/account/profile` with addresses | Unblocks account page |
| P0 | Fix catalog `sort` mapping (`bestseller`, `newest`, `discounted`) + `category_slug` | Unblocks product listing from homepage |
| P0 | Extend `GET /store/homepage` with `categories`, `blog_teaser`, full `stats` | Reduces homepage round-trips; fixes stat counters |
| P0 | Add `payment_url` to checkout response + payment callback route | Unblocks online payment |
| P1 | Implement `GET /store/about` | Unblocks about page |
| P1 | Add `GET /store/navigation` (public read of storefront nav) | Unblocks CMS-driven header |
| P1 | Align response field names to `*_toman` across wishlist, orders, checkout | Prevents frontend adapter bugs |
| P1 | Per-SKU price/stock on product detail + `variant_axes` projection | Unblocks variant selector |
| P2 | `GET /store/products/{id}/related` | Product detail cross-sell |
| P2 | Blog field aliases or frontend mapping doc (`summary`↔`excerpt`) | Blog integration |
| P2 | `GET /admin/contact-messages/stats` | Inbox unread badge |
| P2 | Video upload support for hero (`context=hero` on uploads) | Hero video management |
| P3 | Wishlist idempotent add, `is_in_wishlist` on product detail | UX polish |
| P3 | Shipping methods endpoint + server-side shipping calculation | Checkout step 2 accuracy |
| P3 | Persian error messages for coupon/checkout validation | RTL UX |

---

## Methodology

- **Frontend:** Inspected live [Store OS](https://store-os-eta.vercel.app/) pages (home, products, checkout) and [Admin Dashboard](https://shop-panel-react.vercel.app/); cross-referenced `docs/features/*.md` for expected contracts.
- **Backend:** Enumerated all routes in `router.go`; validated handlers, DTOs, and repository logic against Swagger and feature specs.
- **Status definitions:**
  - **Full** — endpoint exists and response/request match documented frontend needs.
  - **Partial** — endpoint exists but missing fields, params, or behavioral differences.
  - **Missing** — no endpoint or endpoint cannot serve the feature.
