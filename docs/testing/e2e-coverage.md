# E2E Test Coverage

End-to-end tests live in [`tests/e2e/`](../tests/e2e/) and exercise the real HTTP stack: Chi router, middleware, handlers, application services, and PostgreSQL. Tests send JSON requests with varied payloads and assert status codes, error codes, and response bodies.

## How to run

1. Start PostgreSQL (defaults match `docker/docker-compose.yml`):

```bash
docker compose -f docker/docker-compose.yml up -d postgres
```

2. Run e2e tests:

```bash
make test-e2e
```

Or directly:

```bash
go test -race -count=1 -tags=e2e ./tests/e2e/...
```

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DATABASE_URL` | `postgres://ecommerce:ecommerce@localhost:5432/ecommerce?sslmode=disable` | PostgreSQL connection URL |
| `JWT_SECRET` | dev secret (32+ chars) | Must match seeded tokens |
| `E2E_SKIP` | unset | Set to `1` to skip all e2e tests |
| `E2E_REQUIRED` | unset | Set to `1` to fail (exit 1) when DB is unavailable |

Migrations run automatically before the test suite. Seed data provides the default admin account:

- Email: `admin@shop.com`
- Password: `Admin@123456`

All e2e tests use PostgreSQL only (including cart and checkout).

---

## Module coverage

| Module | Status | Test file | Notes |
|--------|--------|-----------|-------|
| **Auth** | ✅ Covered | `tests/e2e/auth_test.go` | Login, signup, refresh, me, forgot-password, validation errors, RBAC |
| **Product (admin)** | ✅ Covered | `tests/e2e/product_test.go` | CRUD lifecycle, validation, slug conflict, search, inventory |
| **Settings (admin)** | ✅ Covered | `tests/e2e/settings_test.go` | Site, contact, social, SEO, navigation; validation + RBAC |
| **Users (admin)** | ✅ Covered | `tests/e2e/users_test.go` | CRUD, validation, duplicate email, cannot delete self, RBAC |
| **Dashboard** | ✅ Covered | `tests/e2e/dashboard_test.go` | Stats, revenue, low stock, recent orders, featured products |
| **Brands (admin)** | ✅ Covered | `tests/e2e/brands_test.go` | CRUD, validation, slug conflict, minimal payload, RBAC |
| **Wishlist** | ✅ Covered | `tests/e2e/wishlist_test.go` | Add/list/count/ids/remove, idempotent add, customer-only |
| **Blog** | ✅ Covered | `tests/e2e/blog_test.go` | Admin categories/posts/comments; storefront list/get/comment flow |
| **Uploads** | ✅ Covered | `tests/e2e/upload_test.go` | Multipart upload, validation, RBAC |
| **Cart** | ✅ Covered | `tests/e2e/cart_test.go` | Add/update/remove/clear, validation, guest + customer |
| **Checkout** | ✅ Covered | `tests/e2e/checkout_test.go` | Preview, place order, coupon preview, payment callback |
| **Orders (admin)** | ✅ Covered | `tests/e2e/orders_test.go` | List/get/notes/invoice/status/cancel/refund, manual create |
| **Coupons (admin)** | ✅ Covered | `tests/e2e/coupons_test.go` | CRUD, validation, duplicate code, activate/deactivate |
| **Categories (admin)** | ✅ Covered | `tests/e2e/categories_test.go` | CRUD, tree, parent/child, slug conflict, delete guards |
| **Customers (admin)** | ✅ Covered | `tests/e2e/customers_test.go` | List, search, get/update/delete, orders history |
| **Storefront (public)** | ✅ Covered | `tests/e2e/storefront_test.go` | Products, categories, brands, homepage, settings |
| **Storefront (admin)** | ✅ Covered | `tests/e2e/storefront_admin_test.go` | Hero, slides, banners, FAQ, reviews, contact, navigation |
| **Product reviews** | ✅ Covered | `tests/e2e/product_reviews_test.go` | Submit, moderation, summary, duplicate guard |
| **Product Q&A** | ✅ Covered | `tests/e2e/product_questions_test.go` | Ask, answer, visibility, admin search |
| **Themes** | ✅ Covered | `tests/e2e/themes_test.go` | List, purchase, store style, public theme sync |
| **Contact (admin inbox)** | ✅ Covered | `tests/e2e/contact_messages_test.go` | List, stats, read/archive/status, delete |
| **Contact form (public)** | ✅ Covered | `tests/e2e/contact_form_test.go` | Submit, validation, sources |
| Health | ⬜ Not tested | — | Covered by unit tests in handler |

---

## Auth — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Admin login success | POST | `/api/v1/auth/login` | 200 + token pair |
| Wrong password | POST | `/api/v1/auth/login` | 401 `INVALID_CREDENTIALS` |
| Invalid/missing fields | POST | `/api/v1/auth/login` | 400 `VALIDATION_ERROR` |
| Customer signup | POST | `/api/v1/auth/signup` | 201 + tokens |
| Duplicate email | POST | `/api/v1/auth/signup` | 409 `CONFLICT` |
| Signup validation | POST | `/api/v1/auth/signup` | 400 `VALIDATION_ERROR` |
| Refresh token | POST | `/api/v1/auth/refresh` | 200 + new tokens |
| Invalid refresh token | POST | `/api/v1/auth/refresh` | 401 |
| Current user (admin) | GET | `/api/v1/auth/me` | 200 + profile |
| Unauthorized me | GET | `/api/v1/auth/me` | 401 |
| Forgot password | POST | `/api/v1/auth/forgot-password` | 200 (always) |
| Customer → admin route | GET | `/api/v1/admin/products` | 403 `FORBIDDEN` |

---

## Product (admin) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized list | GET | `/api/v1/admin/products` | 401 |
| Validation errors | POST | `/api/v1/admin/products` | 400 (missing name, bad price, bad URL, bad status) |
| Minimal create | POST | `/api/v1/admin/products` | 201 |
| Full create + CRUD | POST/GET/PUT/PATCH/DELETE | `/api/v1/admin/products/{id}` | 201 → 200 → 204 → 404 |
| List + pagination | GET | `/api/v1/admin/products?page=1&per_page=5` | 200 + meta |
| Search | GET | `/api/v1/admin/products/search?q=...` | 200 + data |
| Slug conflict | POST | `/api/v1/admin/products` | 409 `CONFLICT` |
| Invalid UUID | GET | `/api/v1/admin/products/not-a-uuid` | 500 (current behavior) |
| Not found | GET | `/api/v1/admin/products/{missing-id}` | 404 `NOT_FOUND` |
| Inventory update | PATCH | `/api/v1/admin/products/{id}/inventory` | 200 + stock fields |

---

## Settings (admin) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized reads | GET | `/api/v1/admin/settings/*` | 401 |
| Site settings round-trip | GET/PUT | `/api/v1/admin/settings/site` | 200 |
| Site validation | PUT | `/api/v1/admin/settings/site` | 400 |
| Contact round-trip | PUT | `/api/v1/admin/settings/contact` | 200 |
| Invalid contact email | PUT | `/api/v1/admin/settings/contact` | 400 |
| Social round-trip | GET/PUT | `/api/v1/admin/settings/social` | 200 |
| Invalid social URL | PUT | `/api/v1/admin/settings/social` | 400 |
| SEO round-trip | PUT | `/api/v1/admin/settings/seo` | 200 |
| Navigation round-trip | GET/PUT | `/api/v1/admin/navigation` | 200 + nested items |
| Customer forbidden | GET | `/api/v1/admin/settings/site` | 403 |

---

## Users (admin) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized list | GET | `/api/v1/admin/users` | 401 |
| Validation errors | POST | `/api/v1/admin/users` | 400 (missing email, bad email, short password, missing name, bad role) |
| Full CRUD lifecycle | POST/GET/PUT/DELETE | `/api/v1/admin/users/{id}` | 201 → 200 → 204 → 404 |
| List + search | GET | `/api/v1/admin/users?q=E2E` | 200 + meta |
| Duplicate email | POST | `/api/v1/admin/users` | 409 `CONFLICT` |
| Cannot delete self | DELETE | `/api/v1/admin/users/{self-id}` | 422 `UNPROCESSABLE_ENTITY` |
| Default role is admin | POST | `/api/v1/admin/users` (no role) | 201, role=admin |
| Not found | GET | `/api/v1/admin/users/{missing-id}` | 404 `NOT_FOUND` |
| Customer forbidden | GET | `/api/v1/admin/users` | 403 `FORBIDDEN` |

---

## Dashboard — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized | GET | `/api/v1/admin/dashboard/*` | 401 |
| Stats KPIs | GET | `/api/v1/admin/dashboard/stats` | 200 + revenue/orders/customers/products/growth |
| Revenue by period | GET | `/api/v1/admin/dashboard/revenue?period=month` | 200 + data array |
| Revenue custom range | GET | `/api/v1/admin/dashboard/revenue?from=...&to=...` | 200 |
| Invalid period | GET | `/api/v1/admin/dashboard/revenue?period=invalid` | 400 `VALIDATION_ERROR` |
| Partial date range | GET | `/api/v1/admin/dashboard/revenue?from=2026-01-01` | 400 `VALIDATION_ERROR` |
| Low stock list | GET | `/api/v1/admin/dashboard/low-stock` | 200 + pagination |
| Recent orders | GET | `/api/v1/admin/dashboard/recent-orders?limit=5` | 200 + data array |
| Invalid recent limit | GET | `/api/v1/admin/dashboard/recent-orders?limit=abc` | 400 `VALIDATION_ERROR` |
| Featured products | GET | `/api/v1/admin/dashboard/featured-products?limit=3` | 200 + data array |
| Invalid featured limit | GET | `/api/v1/admin/dashboard/featured-products?limit=bad` | 400 `VALIDATION_ERROR` |
| Customer forbidden | GET | `/api/v1/admin/dashboard/stats` | 403 `FORBIDDEN` |

---

## Brands (admin) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized list | GET | `/api/v1/admin/brands` | 401 |
| Validation errors | POST | `/api/v1/admin/brands` | 400 (missing/empty/too-long name) |
| Full CRUD lifecycle | POST/GET/PUT/DELETE | `/api/v1/admin/brands/{id}` | 201 → 200 → 204 → 404 |
| List + active filter | GET | `/api/v1/admin/brands?is_active=true` | 200 + meta |
| Slug conflict | POST | `/api/v1/admin/brands` | 409 `CONFLICT` |
| Minimal create | POST | `/api/v1/admin/brands` (name only) | 201 + auto slug |
| Not found | GET | `/api/v1/admin/brands/{missing-id}` | 404 `NOT_FOUND` |
| Customer forbidden | GET | `/api/v1/admin/brands` | 403 `FORBIDDEN` |

---

## Wishlist — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized | GET/POST | `/api/v1/store/account/wishlist` | 401 |
| Admin forbidden | GET | `/api/v1/store/account/wishlist` | 403 `FORBIDDEN` |
| Validation errors | POST | `/api/v1/store/account/wishlist` | 400 (missing/invalid product_id) |
| Add product | POST | `/api/v1/store/account/wishlist` | 201 |
| Idempotent add | POST | `/api/v1/store/account/wishlist` (duplicate) | 200 |
| Count | GET | `/api/v1/store/account/wishlist/count` | 200 + count |
| List IDs | GET | `/api/v1/store/account/wishlist/ids` | 200 + product_ids |
| List items | GET | `/api/v1/store/account/wishlist` | 200 + product summary |
| Remove item | DELETE | `/api/v1/store/account/wishlist/{product_id}` | 204 |
| Product not found | POST | `/api/v1/store/account/wishlist` | 404 `NOT_FOUND` |
| Remove not in list | DELETE | `/api/v1/store/account/wishlist/{product_id}` | 404 `NOT_FOUND` |

---

## Blog — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized admin | GET | `/api/v1/admin/blog/posts` | 401 |
| Category CRUD | POST/GET/PUT/DELETE | `/api/v1/admin/blog/categories/{id}` | 201 → 200 → 204 |
| Category validation | POST | `/api/v1/admin/blog/categories` | 400 |
| Post draft → publish | POST/PUT | `/api/v1/admin/blog/posts/{id}` | draft hidden, published visible |
| Storefront list | GET | `/api/v1/store/blog/posts` | 200 (published only) |
| Storefront get by slug | GET | `/api/v1/store/blog/posts/{slug}` | 200 / 404 for draft |
| Submit comment | POST | `/api/v1/store/blog/posts/{id}/comments` | 201, status=pending |
| Approve comment | PATCH | `/api/v1/admin/blog/comments/{id}/approve` | 204 |
| Approved comments visible | GET | `/api/v1/store/blog/posts/{id}/comments` | 200 + approved |
| Comments by slug | GET | `/api/v1/store/blog/{slug}/comments` | 200 |
| Storefront categories | GET | `/api/v1/store/blog/categories` | 200 |
| Post validation | POST | `/api/v1/admin/blog/posts` | 400 |
| Customer forbidden | GET | `/api/v1/admin/blog/posts` | 403 `FORBIDDEN` |

---

## Uploads — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized | POST | `/api/v1/admin/uploads` | 401 |
| Valid PNG upload | POST | `/api/v1/admin/uploads` (multipart) | 201 + url/filename/size/content_type |
| Missing file | POST | `/api/v1/admin/uploads` | 400 |
| Invalid file type | POST | `/api/v1/admin/uploads` (text/plain) | 400 `VALIDATION_ERROR` |
| Empty file | POST | `/api/v1/admin/uploads` | 400 `VALIDATION_ERROR` |
| Customer forbidden | POST | `/api/v1/admin/uploads` | 403 `FORBIDDEN` |

---

## Cart — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Empty cart | GET | `/api/v1/store/cart` | 200 + empty items |
| Validation errors | POST | `/api/v1/store/cart/items` | 400 (missing/invalid product_id, zero qty) |
| Add + update + remove | POST/PATCH/DELETE | `/api/v1/store/cart/items/{id}` | 200 + cart view |
| Increment on duplicate add | POST | `/api/v1/store/cart/items` | 200, merged quantity |
| Clear cart | DELETE | `/api/v1/store/cart` | 204 |
| Default quantity | POST | `/api/v1/store/cart/items` (no qty) | 200, quantity=1 |
| Product not found | POST | `/api/v1/store/cart/items` | 404 `NOT_FOUND` |
| Remove not in cart | DELETE | `/api/v1/store/cart/items/{id}` | 404 `NOT_FOUND` |
| Authenticated customer cart | GET | `/api/v1/store/cart` | 200 + items |

---

## Checkout — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Preview empty cart | POST | `/api/v1/store/checkout/preview` | 400 `VALIDATION_ERROR` |
| Preview validation | POST | `/api/v1/store/checkout/preview` | 400 (bad/missing shipping) |
| Preview success | POST | `/api/v1/store/checkout/preview` | 200 + items + totals |
| Preview valid coupon | POST | `/api/v1/store/checkout/preview` | 200 + discount |
| Preview invalid coupon | POST | `/api/v1/store/checkout/preview` | 200, `is_valid=false` |
| Guest place order | POST | `/api/v1/store/checkout` | 201 + order_id, cart cleared |
| Authenticated checkout | POST | `/api/v1/store/checkout` | 201 + visible in account orders |
| Checkout validation | POST | `/api/v1/store/checkout` | 400 `VALIDATION_ERROR` |
| Payment callback | POST | `/api/v1/store/checkout/payment/callback` | 200, payment_status=paid |

---

## Orders (admin) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized list | GET | `/api/v1/admin/orders` | 401 |
| Get after storefront checkout | GET | `/api/v1/admin/orders/{id}` | 200 + items |
| List with status filter | GET | `/api/v1/admin/orders?status=pending` | 200 + meta |
| Update notes | PATCH | `/api/v1/admin/orders/{id}/notes` | 200 |
| Get invoice | GET | `/api/v1/admin/orders/{id}/invoice` | 200 |
| Status workflow | PATCH | `/api/v1/admin/orders/{id}/status` | pending → processing → shipped → delivered |
| Invalid status jump | PATCH | `/api/v1/admin/orders/{id}/status` | 422 `INVALID_STATUS_TRANSITION` |
| Cancel order | POST | `/api/v1/admin/orders/{id}/cancel` | 200, status=cancelled |
| Refund delivered order | POST | `/api/v1/admin/orders/{id}/refund` | 200, status=refunded |
| Manual create | POST | `/api/v1/admin/orders` | 201 |
| Create validation | POST | `/api/v1/admin/orders` | 400 `VALIDATION_ERROR` |
| Invalid date range filter | GET | `/api/v1/admin/orders?from=...&to=...` | 400 `VALIDATION_ERROR` |
| Not found | GET | `/api/v1/admin/orders/{missing-id}` | 404 `NOT_FOUND` |
| Customer forbidden | GET | `/api/v1/admin/orders` | 403 `FORBIDDEN` |

---

## Coupons (admin) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized list | GET | `/api/v1/admin/coupons` | 401 |
| Validation errors | POST | `/api/v1/admin/coupons` | 400 (missing code, bad type/value) |
| Full CRUD lifecycle | POST/GET/PUT/DELETE | `/api/v1/admin/coupons/{id}` | 201 → 200 → 204 → 404 |
| List + search | GET | `/api/v1/admin/coupons?q=...` | 200 + meta |
| Activate / deactivate | PATCH | `/api/v1/admin/coupons/{id}/activate\|deactivate` | 200 + is_active |
| Duplicate code | POST | `/api/v1/admin/coupons` | 409 `CONFLICT` |
| Not found | GET | `/api/v1/admin/coupons/{missing-id}` | 404 `NOT_FOUND` |
| Customer forbidden | GET | `/api/v1/admin/coupons` | 403 `FORBIDDEN` |

---

## Categories (admin) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized list | GET | `/api/v1/admin/categories` | 401 |
| Validation errors | POST | `/api/v1/admin/categories` | 400 (missing/empty/too-long name, bad parent_id, bad image_url) |
| Full CRUD lifecycle | POST/GET/PUT/DELETE | `/api/v1/admin/categories/{id}` | 201 → 200 → 204 → 404 |
| List + active filter | GET | `/api/v1/admin/categories?is_active=true` | 200 + meta |
| Nested tree | GET | `/api/v1/admin/categories?tree=true` | 200 + children |
| Parent/child hierarchy | POST | `/api/v1/admin/categories` | child under parent |
| Delete with children | DELETE | `/api/v1/admin/categories/{parent-id}` | 422 `UNPROCESSABLE_ENTITY` |
| Delete with products | DELETE | `/api/v1/admin/categories/{id}` | 422 `UNPROCESSABLE_ENTITY` |
| Slug conflict | POST | `/api/v1/admin/categories` | 409 `CONFLICT` |
| Parent not found | POST | `/api/v1/admin/categories` | 404 `NOT_FOUND` |
| Not found | GET | `/api/v1/admin/categories/{missing-id}` | 404 `NOT_FOUND` |
| Customer forbidden | GET | `/api/v1/admin/categories` | 403 `FORBIDDEN` |

---

## Customers (admin) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized list | GET | `/api/v1/admin/customers` | 401 |
| List + type filter | GET | `/api/v1/admin/customers?type=registered` | 200 + meta |
| Search by email | GET | `/api/v1/admin/customers?q=...` | 200 + matching row |
| Get detail | GET | `/api/v1/admin/customers/{id}` | 200 + stats + addresses |
| Update profile | PUT | `/api/v1/admin/customers/{id}` | 200 + updated fields |
| Delete without orders | DELETE | `/api/v1/admin/customers/{id}` | 204 → 404 |
| Update validation | PUT | `/api/v1/admin/customers/{id}` | 400 (bad email, bad type) |
| Duplicate email | PUT | `/api/v1/admin/customers/{id}` | 409 `CONFLICT` |
| Purchase history | GET | `/api/v1/admin/customers/{id}/orders` | 200 + order after checkout |
| Delete with orders | DELETE | `/api/v1/admin/customers/{id}` | 422 `UNPROCESSABLE_ENTITY` |
| Not found | GET | `/api/v1/admin/customers/{missing-id}` | 404 `NOT_FOUND` |
| Customer forbidden | GET | `/api/v1/admin/customers` | 403 `FORBIDDEN` |

---

## Storefront (public) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Homepage content | GET | `/api/v1/store/homepage` | 200 |
| Public settings | GET | `/api/v1/store/settings` | 200 |
| Checkout settings | GET | `/api/v1/store/settings/checkout` | 200 |
| Navigation menu | GET | `/api/v1/store/navigation` | 200 |
| About page | GET | `/api/v1/store/about` | 200 |
| Active theme | GET | `/api/v1/store/theme` | 200 |
| Shipping methods | GET | `/api/v1/store/checkout/shipping-methods?city=Tehran` | 200 |
| List active products | GET | `/api/v1/store/products` | 200 + active product |
| Search products | GET | `/api/v1/store/products/search?q=...` | 200 + hits |
| Search validation | GET | `/api/v1/store/products/search` (no q) | 400 `VALIDATION_ERROR` |
| Product by slug/id | GET | `/api/v1/store/products/{slugOrId}` | 200 + detail |
| Related products | GET | `/api/v1/store/products/{id}/related` | 200 + data array |
| Category filter | GET | `/api/v1/store/products?category_id=...` | 200 + filtered product |
| Draft hidden | GET | `/api/v1/store/products/{draft-slug}` | 404 `NOT_FOUND` |
| Active categories tree | GET | `/api/v1/store/categories` | 200 (active only; deactivate via admin PUT) |
| Active brands | GET | `/api/v1/store/brands` | 200 + active brand |
| Product not found | GET | `/api/v1/store/products/{missing-slug}` | 404 `NOT_FOUND` |

---

## Storefront (admin) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized | GET | `/api/v1/admin/storefront/hero` | 401 |
| Hero round-trip | GET/PUT | `/api/v1/admin/storefront/hero` | 200 + updated fields |
| List product slides | GET | `/api/v1/admin/storefront/product-slides` | 200 + seeded slides |
| Update slide | PUT | `/api/v1/admin/storefront/product-slides/{type}` | 200 |
| Slide item CRUD | POST/PUT/DELETE | slide items endpoints | 201 → 200 → 204 |
| Invalid slide type | PUT | `/api/v1/admin/storefront/product-slides/invalid` | 400 `VALIDATION_ERROR` |
| Pro banner CRUD | POST/GET/PUT/DELETE | `/api/v1/admin/storefront/pro-banners/{id}` | 201 → 200 → 204 |
| FAQ section + items | GET/PUT/POST/PUT/DELETE | `/api/v1/admin/storefront/faq*` | full lifecycle |
| Homepage reviews CRUD | POST/GET/PUT/DELETE | `/api/v1/admin/storefront/homepage-reviews/{id}` | 201 → 200 → 204 |
| Contact section | GET/PUT | `/api/v1/admin/storefront/contact-section` | 200 |
| Navigation | GET/PUT | `/api/v1/admin/storefront/navigation` | 200 + items |
| Validation errors | POST | banners/reviews/faq items | 400 `VALIDATION_ERROR` |
| Customer forbidden | GET | `/api/v1/admin/storefront/hero` | 403 `FORBIDDEN` |

---

## Product reviews — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Submit validation | POST | `/api/v1/store/products/{id}/reviews` | 400 (missing author, bad rating, missing content) |
| Guest submit | POST | `/api/v1/store/products/{id}/reviews` | 201, status=pending |
| Pending hidden publicly | GET | `/api/v1/store/products/{id}/reviews` | 200, empty until approved |
| Summary before approval | GET | `/api/v1/store/products/{id}/reviews/summary` | 200, total_count=0 |
| Admin list pending | GET | `/api/v1/admin/reviews?status=pending` | 200 + review |
| Approve review | PATCH | `/api/v1/admin/reviews/{id}/status` | 204 |
| Approved visible | GET | `/api/v1/store/products/{id}/reviews` | 200 + approved review |
| Summary after approval | GET | `/api/v1/store/products/{id}/reviews/summary` | 200, total_count≥1 |
| Reject review | PATCH | `/api/v1/admin/reviews/{id}/status` | 204, status=rejected |
| Delete review | DELETE | `/api/v1/admin/reviews/{id}` | 204 |
| Duplicate (authenticated) | POST | `/api/v1/store/products/{id}/reviews` | 409 `CONFLICT` |
| Product not found | POST/GET | `/api/v1/store/products/{missing-id}/reviews` | 404 `NOT_FOUND` |
| Admin filters | GET | `/api/v1/admin/reviews?product_id&rating&status&q` | 200 + match |
| Admin status validation | PATCH | `/api/v1/admin/reviews/{id}/status` | 400 invalid status |
| Unauthorized / forbidden | GET | `/api/v1/admin/reviews` | 401 / 403 |

---

## Product Q&A — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Ask validation | POST | `/api/v1/store/products/{id}/questions` | 400 (missing name/question, bad email) |
| Ask question | POST | `/api/v1/store/products/{id}/questions` | 201, status=open |
| Open hidden publicly | GET | `/api/v1/store/products/{id}/questions` | 200, empty until answered |
| Admin list open | GET | `/api/v1/admin/questions?status=open` | 200 + question |
| Answer question | POST | `/api/v1/admin/questions/{id}/answer` | 204 |
| Answered visible | GET | `/api/v1/store/products/{id}/questions` | 200 + Q&A |
| Delete question | DELETE | `/api/v1/admin/questions/{id}` | 204 |
| Admin search | GET | `/api/v1/admin/questions?q=...` | 200 + match |
| Answer validation | POST | `/api/v1/admin/questions/{id}/answer` | 400 missing answer |
| Product not found | POST/GET | `/api/v1/store/products/{missing-id}/questions` | 404 `NOT_FOUND` |
| Unauthorized / forbidden | GET | `/api/v1/admin/questions` | 401 / 403 |

---

## Themes — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized list | GET | `/api/v1/admin/themes` | 401 |
| List seeded themes | GET | `/api/v1/admin/themes` | 200 + modern-blue active |
| Purchase theme | POST | `/api/v1/admin/themes/{id}/purchase` | 201 |
| Idempotent purchase | POST | `/api/v1/admin/themes/{id}/purchase` | 201 |
| Get store style | GET | `/api/v1/admin/store-style` | 200 + colors/font |
| Update active theme + tokens | PUT | `/api/v1/admin/store-style` | 200 |
| Public theme sync | GET | `/api/v1/store/theme` | 200 matches admin style |
| Update colors only | PUT | `/api/v1/admin/store-style` | 200 + merged primary |
| Purchase not found | POST | `/api/v1/admin/themes/{missing-id}/purchase` | 404 `NOT_FOUND` |
| Invalid theme id on style | PUT | `/api/v1/admin/store-style` | 400 / 404 |
| Customer forbidden | GET | `/api/v1/admin/themes` | 403 `FORBIDDEN` |

---

## Contact (admin inbox) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Unauthorized list | GET | `/api/v1/admin/contact-messages` | 401 |
| Inbox stats | GET | `/api/v1/admin/contact-messages/stats` | 200 + unread/total |
| List + filters | GET | `/api/v1/admin/contact-messages?status&source` | 200 + meta |
| Search by subject | GET | `/api/v1/admin/contact-messages?q=...` | 200 + match |
| Get message | GET | `/api/v1/admin/contact-messages/{id}` | 200 |
| Mark read | PATCH | `/api/v1/admin/contact-messages/{id}/read` | 204 |
| Archive | PATCH | `/api/v1/admin/contact-messages/{id}/archive` | 204 |
| Update status | PATCH | `/api/v1/admin/contact-messages/{id}/status` | 204 |
| Delete message | DELETE | `/api/v1/admin/contact-messages/{id}` | 204 → 404 |
| Status validation | PATCH | `.../status` (invalid) | 400 `VALIDATION_ERROR` |
| Not found | GET | `/api/v1/admin/contact-messages/{missing-id}` | 404 `NOT_FOUND` |
| Customer forbidden | GET | `/api/v1/admin/contact-messages` | 403 `FORBIDDEN` |

---

## Contact form (public) — scenarios covered

| Scenario | Method | Endpoint | Expected |
|----------|--------|----------|----------|
| Submit success | POST | `/api/v1/store/contact` | 201, status=unread |
| Default source | POST | `/api/v1/store/contact` (no source) | 201, source=homepage |
| Validation errors | POST | `/api/v1/store/contact` | 400 (missing name/message, bad email, bad source) |
| All sources | POST | `/api/v1/store/contact` | 201 for homepage/about/contact_page |
| Rate limit | POST | `/api/v1/store/contact` (11 rapid) | 10×201 then 429 `RATE_LIMITED` (in `contact_messages_test.go`) |

---

## Adding new module tests

1. Create `tests/e2e/<module>_test.go` with `//go:build e2e`.
2. Reuse helpers from `tests/e2e/client.go`, `adminClient(t)`, `customerClient(t)`, and `storeClient(t)`.
3. Update this document: move the module from **Not tested** to **Covered** and list scenarios.
4. Run `make test-e2e` locally before committing.

Last updated: 2026-06-30
