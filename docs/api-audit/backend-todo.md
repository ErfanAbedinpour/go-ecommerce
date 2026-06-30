# Backend Implementation Checklist

> Generated from API audit comparing [Store OS](https://store-os-eta.vercel.app/) + [Admin Panel](https://shop-panel-react.vercel.app/) against current Go backend.  
> **Last updated:** 2026-06-30 — contract normalization pass complete (catalog, homepage, product includes, wishlist/blog aliases).
> Check items that are **not yet done** or need **contract fixes**.

---

# Authentication

- [x] Login endpoint (`POST /api/v1/auth/login`)
- [x] Signup endpoint (`POST /api/v1/auth/signup`)
- [x] Refresh token endpoint (`POST /api/v1/auth/refresh`)
- [x] Logout endpoint (`POST /api/v1/auth/logout`)
- [x] Current user endpoint (`GET /api/v1/auth/me`)
- [x] Forgot password (`POST /api/v1/auth/forgot-password`)
- [x] Reset password (`POST /api/v1/auth/reset-password`)
- [ ] Return `role` in login/signup response for client routing (verify DTO matches UI)
- [ ] Customer-specific signup default role env (`AUTH_SIGNUP_DEFAULT_ROLE=customer`)

---

# Store — Homepage & Layout

- [x] Aggregated homepage (`GET /api/v1/store/homepage`)
- [x] Public store navigation (`GET /api/v1/store/navigation`)
- [x] Embed `categories[]` in homepage response (or document separate call)
- [x] Add `blog_teaser.posts[]` (latest 3 published posts)
- [x] Extend `stats` with `customers_count`, `delivered_orders_count`, `years_experience`
- [ ] Add `hero.poster_url` fallback image
- [x] Align `product_slides[].slide_type` enum: `featured` → `new` (or document mapping)
- [x] Theme tokens (`GET /api/v1/store/theme`)
- [x] Public settings (`GET /api/v1/store/settings`)
- [x] Contact form submit (`POST /api/v1/store/contact`)
- [ ] Require `source` enum validation (`homepage|about|contact_page`)
- [ ] Simplify contact response to `{ id, message }` or document full object
- [x] Rate-limit `POST /store/contact` by IP

---

# Store — Product Catalog

- [x] Product list (`GET /api/v1/store/products`)
- [x] Pagination (`page`, `per_page`)
- [x] Search (`q`)
- [x] Category filter by UUID (`category_id`)
- [x] Category filter by slug (`category_slug`)
- [x] Include child categories (`include_children`, default true)
- [x] Brand filter (`brand`)
- [x] On-sale filter (`on_sale`)
- [x] In-stock filter (`in_stock`)
- [x] Sort: `bestseller` (90-day sales aggregation)
- [x] Sort: map `newest` → `created_at DESC`
- [x] Sort: map `discounted` → existing `discount` logic
- [x] Sort: `price_desc` (currently only `price` ASC)
- [ ] Product card fields: `short_description`, `category`, `is_new`, `has_variants`, `variant_count`, `price_from_toman`, `price_to_toman`
- [ ] Response echo `filters_applied`
- [x] Category tree (`GET /api/v1/store/categories`)
- [ ] Support `tree` / `with_products_count` query params (or document always-on behavior)
- [x] Product search autocomplete (`GET /api/v1/store/products/search`)
- [x] Public brands list (`GET /api/v1/store/brands`)

---

# Store — Product Detail

- [x] Get product by slug or UUID (`GET /api/v1/store/products/{slugOrId}`)
- [ ] Per-SKU `price_toman`, `sale_price_toman`, `quantity`, `is_in_stock`
- [ ] `variant_axes[]` projection from attributes
- [ ] `default_sku_id`
- [ ] Nested `category { id, name, slug }`
- [x] `reviews_summary` embedded in detail response (`?include=reviews_summary`)
- [x] `is_in_wishlist` (optional auth middleware on GET + `?include=wishlist`)
- [ ] `seo` block
- [x] Related products (`GET /api/v1/store/products/{id}/related`)

---

# Store — Reviews & Q&A

- [x] List reviews (`GET /api/v1/store/products/{productId}/reviews`)
- [x] Review summary (`GET /api/v1/store/products/{productId}/reviews/summary`)
- [x] Submit review (`POST /api/v1/store/products/{productId}/reviews`)
- [x] Accept product slug in path (or document UUID-only)
- [ ] Embed `summary` in list response
- [ ] Add `is_verified_buyer` flag
- [ ] Enforce customer-only review submission (403 for non-buyers)
- [ ] Map `sort=rating` alias to `highest`/`lowest`
- [x] List questions (`GET /api/v1/store/products/{productId}/questions`)
- [x] Ask question (`POST /api/v1/store/products/{productId}/questions`)
- [ ] Hide internal fields (`status`, `product_id`) from public list

---

# Store — Checkout

- [x] Checkout preview (`POST /api/v1/store/checkout/preview`)
- [x] Place order (`POST /api/v1/store/checkout`)
- [x] Coupon validation (`POST /api/v1/store/coupons/validate`)
- [ ] Server-side shipping calculation from `shipping_method` + `shipping_city`
- [ ] Remove client-supplied `shipping_amount` / `tax_amount` from preview (or document)
- [ ] `warnings[]` in preview response for price changes
- [ ] `payment_url` and `expires_at` in checkout response (callback route implemented; PSP redirect URL still pending)
- [ ] Idempotency-Key header support
- [ ] `409 CONFLICT` with `unavailable_items[]` for stock errors
- [x] Shipping methods (`GET /api/v1/store/checkout/shipping-methods?city=`)
- [x] Checkout settings (`GET /api/v1/store/settings/checkout`)
- [x] Payment callback (`POST /api/v1/store/checkout/payment/callback`)
- [ ] Persian validation messages for coupons
- [x] Order confirmation email (mailer wired)

---

# Store — Account

- [x] Profile read (`GET /api/v1/store/account/profile`)
- [x] Profile update (`PUT /api/v1/store/account/profile`)
- [x] Address book replace within profile update
- [x] Order list (`GET /api/v1/store/account/orders`)
- [ ] Order list `status` filter
- [x] Order detail (`GET /api/v1/store/account/orders/{id}`)
- [ ] Order detail: `*_toman` integer fields
- [ ] Order detail: `variant_label` on line items
- [ ] Order detail: `timeline[]` from status history
- [ ] Retry payment link for unpaid orders

---

# Store — Wishlist

- [x] List wishlist (`GET /api/v1/store/account/wishlist`)
- [x] Add item (`POST /api/v1/store/account/wishlist`)
- [x] Remove item (`DELETE /api/v1/store/account/wishlist/{productId}`)
- [x] Rename `created_at` → `added_at` (or alias both)
- [x] Nested product card with `*_toman` fields
- [x] Idempotent add (200 on duplicate vs 409)
- [x] Wishlist IDs shortcut (`GET .../wishlist/ids`)
- [x] Wishlist count (`GET .../wishlist/count`)

---

# Store — Blog

- [x] List posts (`GET /api/v1/store/blog/posts`)
- [x] Post detail (`GET /api/v1/store/blog/posts/{slug}`)
- [x] Categories (`GET /api/v1/store/blog/categories`)
- [x] Comments list/submit (`/blog/posts/{postId}/comments`)
- [ ] `category_slug` filter on list
- [ ] `posts_count` on categories
- [x] `excerpt` alias for `summary`
- [x] `cover_image_url` alias for `featured_image`
- [ ] `read_time_minutes` on posts
- [ ] `related_posts[]` on detail
- [ ] `content_html` / `content_markdown` split
- [x] Blog list alias (`GET /api/v1/store/blog`)
- [x] Comment routes by slug (`GET /api/v1/store/blog/{slug}/comments`)

---

# Store — About

- [x] About page aggregate (`GET /api/v1/store/about`)
- [x] `about` JSONB in `store_settings` (migration `000014`)
- [ ] Admin API to edit about content (future)
- [x] Contact form reuse (`POST /api/v1/store/contact`)

---

# Admin — Dashboard

- [x] Stats (`GET /api/v1/admin/dashboard/stats`)
- [x] Revenue chart (`GET /api/v1/admin/dashboard/revenue`)
- [x] Recent orders (`GET /api/v1/admin/dashboard/recent-orders`)
- [x] Low stock (`GET /api/v1/admin/dashboard/low-stock`)
- [x] Featured products (`GET /api/v1/admin/dashboard/featured-products`)

---

# Admin — Products

- [x] List with filters (`GET /api/v1/admin/products`)
- [x] Stats (`GET /api/v1/admin/products/stats`)
- [x] Search (`GET /api/v1/admin/products/search`)
- [x] CRUD (`POST`, `GET`, `PUT`, `DELETE`)
- [x] Inventory patch (`PATCH .../inventory`)
- [x] SKU matrix from `attributes[]` on create/update
- [ ] Per-SKU inventory tracking
- [ ] Persist `adjustment_reason` on inventory updates
- [ ] Top-level `sku` convenience field on list response
- [x] Image upload (`POST /api/v1/admin/uploads`)
- [ ] Video upload for hero (`context=hero` query param)
- [ ] Upload `context` param for organized storage paths

---

# Admin — Orders

- [x] List with filters including `from`/`to` dates
- [x] Detail with timeline
- [x] Status update, notes, cancel, refund
- [x] Invoice PDF/JSON
- [x] Manual order create

---

# Admin — Customers

- [x] List, detail, update, delete
- [x] Customer order history

---

# Admin — Coupons

- [x] Full CRUD + activate/deactivate

---

# Admin — Settings

- [x] Site, contact, social, SEO GET/PUT
- [x] Admin navigation GET/PUT

---

# Admin — Storefront Context (CMS)

- [x] Hero GET/PUT
- [x] Product slides + items CRUD
- [x] Pro banners CRUD
- [x] Partner brands CRUD
- [x] Homepage reviews CRUD
- [x] FAQ section image + items CRUD
- [x] Contact section image GET/PUT
- [x] Storefront navigation GET/PUT
- [ ] Bulk `PUT /storefront/product-slides`
- [ ] Atomic `PUT /storefront/faq` with embedded items (optional)
- [ ] Pagination on partner brands / homepage reviews lists
- [ ] Video upload for hero section

---

# Admin — Themes

- [x] List themes
- [x] Purchase theme
- [x] Get/update store style
- [ ] Pagination on theme list (`page`, `per_page`, `meta`)

---

# Admin — Blog

- [x] Posts CRUD
- [x] Categories CRUD
- [x] Comment moderation (`PATCH .../status`)
- [ ] `read_time_minutes` field
- [ ] `archived` post status
- [ ] `excerpt` field alias (currently `summary`)
- [x] Separate approve/reject routes (`PATCH .../comments/{id}/approve|reject`)

---

# Admin — Contact Inbox

- [x] List with filters (`status`, `source`, `q`, `from`, `to`)
- [x] Detail, status update, delete
- [x] Inbox stats (`GET /api/v1/admin/contact-messages/stats`)
- [x] Mark read alias (`PATCH .../contact-messages/{id}/read`)
- [x] Archive alias (`PATCH .../contact-messages/{id}/archive`)
- [ ] Auto-mark-read on GET detail
- [ ] Return body on status PATCH (currently 204)

---

# Admin — Engagement Moderation

- [x] Product reviews list, status update, delete
- [x] Product questions list, answer, delete

---

# Admin — Users (staff)

- [x] Admin user CRUD (`/api/v1/admin/users`)

---

# Infrastructure & Cross-Cutting

- [x] JWT authentication middleware
- [x] Role guards (admin, customer)
- [x] Rate limiting on auth and checkout
- [x] Redis response cache on product list + homepage
- [x] Audit log middleware on admin mutations
- [ ] Consistent `*_toman` integer types across all store DTOs
- [ ] OpenAPI/Swagger sync with actual sort param values and newly added routes
- [ ] Update stale docs (`docs/architecture/gap-analysis.md`, `docs/api/admin-api.md`)
