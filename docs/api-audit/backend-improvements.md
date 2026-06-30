# Backend Improvements

Architectural and API design recommendations from the 2026-06-30 audit.  
**Last updated:** 2026-06-30 — contract normalization pass: catalog filters/sort, homepage aggregate, product detail `?include=`, wishlist/blog field aliases, caching, and rate limits implemented.

---

## API Design

### 1. Versioned store DTOs separate from admin DTOs

Customer-facing responses should use a dedicated `dto/response/store_*.go` package with:

- Integer `*_toman` amounts (no floats in JSON)
- Customer-safe field sets (no internal IDs like `coupon_id` on order detail)
- Persian-ready `message` fields on validation errors

Admin DTOs can retain richer internal fields. Sharing `OrderDetailResponse` between admin and store causes contract leakage.

### 2. Aggregated read endpoints for page-level data

The homepage, about page, and product detail benefit from **Backend-for-Frontend (BFF)** aggregates:

| Page | Current calls | Target | Status |
|------|---------------|--------|--------|
| `/` | homepage + categories + theme | Single `GET /store/homepage` with embedded categories | **Done** — categories, blog teaser, extended stats embedded |
| `/about` | settings (partial) | `GET /store/about` | **Done** |
| `/products/:id` | product + reviews summary + wishlist state | Optional `?include=reviews_summary,wishlist` | **Done** — optional auth on GET product |

Reduces waterfall requests on slow mobile networks (primary Store OS audience).

### 3. Slug-first public resource paths

Product detail accepts slug; reviews, Q&A, and wishlist use UUID. **Standardize slug resolution** in a shared middleware or helper:

```go
func ResolveProductID(ctx, slugOrID string) (uuid.UUID, error)
```

Apply to all `/store/products/{ref}/...` sub-routes. **Done** for reviews, Q&A, and related products via `productref.ResolveID`.

### 4. OpenAPI as contract gate

Swagger documents `sort=bestseller|newest|discounted` — **repo now maps all documented values** (migration `000016` adds sort indexes). **Regenerate Swagger** to include `?include=` on product detail and new field aliases, then add CI check:

```bash
go test ./internal/... # integration tests per documented query param
```

Or generate contract tests from `swagger.yaml`.

---

## REST Conventions

### Consistent path naming

| Issue | Current | Recommended | Status |
|-------|---------|-------------|--------|
| Blog list | `/store/blog/posts` + alias `/store/blog` | Document both | **Done** |
| Admin reviews | `/admin/reviews` | Docs say `/admin/product-reviews` — pick one | Open |
| FAQ | Split `/faq` + `/faq/items` | OK for REST; document atomic save pattern for UI | Open |
| Contact moderation | `PATCH .../status` + `/read` + `/archive` | Document aliases | **Done** |
| Blog moderation | `PATCH .../status` + `/approve` + `/reject` | Document aliases | **Done** |

### HTTP verbs and status codes

| Pattern | Recommendation |
|---------|----------------|
| Wishlist duplicate add | `200` idempotent preferred over `409` for UX | **Done** |
| Contact status update | Return `200` + body instead of `204` for SPA state updates |
| Stock conflict on checkout | `409` with structured `unavailable_items[]` |
| Review moderation | `PATCH /status` is fine; document enum values |
| Payment callback | Idempotent; return `409` if already paid (**implemented**) |

### Pagination envelope

All list endpoints should return:

```json
{ "data": [], "meta": { "page", "per_page", "total", "total_pages" } }
```

Blog categories and theme list currently return bare arrays.

---

## Naming

### Field naming standard (store API)

| Concept | Standard name |
|---------|---------------|
| Money | `*_toman` (int64) |
| Timestamps | `created_at`, `updated_at`, `added_at` (wishlist) |
| Images | `thumbnail_url`, `cover_image_url` |
| Text summary | `excerpt` (add alias from `summary`) |
| Booleans | `is_on_sale`, `is_out_of_stock`, `is_verified_buyer` |

Provide **deprecated aliases** during migration (`summary` + `excerpt` both present for one release).

---

## Status Codes & Error Handling

### Structured validation errors

Frontend forms (checkout, contact, profile) need field-level errors:

```json
{
  "statusCode": 400,
  "path": "/api/v1/store/checkout",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "...",
    "details": { "shipping_address.postal_code": "invalid format" }
  }
}
```

Already supported via `apperror.Validation` — ensure all storefront handlers use it consistently.

### Persian user messages

Store-facing errors should return Persian `message` strings for coupon, stock, and contact validation. Keep English `code` for logging.

---

## Response Consistency

### Single product card mapper

One function `ToStoreProductCard(product, opts)` used by:

- Catalog list
- Homepage slides
- Wishlist nested product
- Related products (**implemented** — verify shared mapper)

Eliminates `price` vs `price_toman` drift.

### Order amounts

Store order DTOs should never expose `float64` currency. Use integer Toman throughout the Persian storefront.

---

## Pagination, Filtering, Sorting

### Catalog

**Done** — `StoreListFilter` implements `category_slug`, `include_children`, `brand`, `on_sale`, `in_stock`, and sort mapping (`bestseller`, `newest`, `discounted`, `price_asc`, `price_desc`).

```go
type StoreListFilter struct {
    Query           string
    CategoryID      *uuid.UUID
    CategorySlug    string
    IncludeChildren bool
    Brand           string
    Sort            string  // mapped: bestseller, newest, discounted, price_asc, price_desc
    OnSale          *bool
    InStock         *bool
}
```

### Admin lists

Add pagination to: themes, partner brands, homepage reviews, FAQ items (when list grows).

### Orders (admin)

`from`/`to` date filter is implemented — update stale documentation.

---

## Performance

### Caching (already started)

Redis cache on `GET /store/products` and `GET /store/homepage` (5 min TTL). **Extended to:**

- `GET /store/categories` (**Done** — invalidate on category CRUD still TODO)
- `GET /store/navigation` (**Done**)
- `GET /store/theme` (**Done** — invalidate on style update still TODO)

Use cache tags or key prefixes for targeted invalidation on admin content updates.

### Homepage stats

**Done** — `customers_count`, `delivered_orders_count`, `years_experience` computed on each homepage load. Future: materialized view or Redis counter for scale.

### N+1 on product slides

**Done** — `BuildHomepage` batch-loads slide products via `FindByIDs`.

---

## Security

### Public endpoints

- Rate-limit `POST /store/contact` (**Done** — 3 req/min burst 10)
- Honeypot field `website` on contact forms
- CAPTCHA on contact/checkout for production

### Customer data isolation

Order detail and wishlist must verify `customer_id` matches JWT subject — audit all store account handlers.

### Upload hardening

- Validate MIME type from content, not just extension
- Max file size per context (image 10MB, video 50MB)
- S3 presigned URLs for large video uploads (future)

### Audit log

Admin audit middleware is wired — ensure sensitive reads (customer PII export) are logged if added.

### Payment callback

`PAYMENT_CALLBACK_SECRET` + HMAC signature verification implemented — enforce secret in production.

---

## Transactions

### Checkout atomicity

`PlaceCheckout` must wrap in a single DB transaction:

1. Validate stock
2. Decrement inventory
3. Create order + items
4. Increment coupon usage
5. Create guest customer if needed

Verify rollback on any failure (partially implemented — add integration test).

### Theme purchase

`PurchaseTheme` is idempotent if user already owns theme (**Done** — returns existing purchase).

---

## Database

### Per-SKU inventory (future)

Current schema has `skus` table but inventory is product-level. For building materials with size/color variants, migrate to `sku_inventory` or stock on `skus` row.

### About content

**Implemented:** JSONB `about` column on `store_settings` (migration `000014`). Checkout settings in JSONB `checkout` (migration `000015`). Add admin edit API when CMS UI is ready.

### Indexes for catalog sort

Add index supporting bestseller query (**Done** — migration `000016`):

```sql
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_orders_status_created ON orders(status, created_at DESC);
```

---

## Scalability

### Stateless API

Current design is stateless — good for horizontal scaling. Redis cache and rate limiter must use shared Redis in multi-instance deploys.

### Background jobs

Move to async:

- Order confirmation email
- Contact message admin notification
- Blog comment moderation queue

Use a simple job table or Redis queue before introducing full worker service.

### Payment webhooks

Payment callback should be idempotent and fast — verify signature, update order, enqueue email, return 200 immediately (**callback handler implemented**; add PSP-specific verification when provider is chosen).

---

## Maintainability

### Documentation sync

These files may still be **stale** relative to `router.go`:

- `docs/architecture/gap-analysis.md`
- `docs/api/admin-api.md`
- `docs/README.md` executive summary (partially)

**Authoritative audit docs:** `docs/api-audit/` (this folder). Regenerate `admin-api.md` from Swagger.

### Feature flags

Use config for:

- `AUTH_SIGNUP_ENABLED`
- COD payment enabled (`store_settings.checkout` JSONB)
- Guest checkout enabled
- Review auto-publish vs moderation
- `PAYMENT_CALLBACK_SECRET` for gateway verification

### Test coverage gaps

Add contract tests for:

- Catalog sort params (each documented value)
- Homepage projection completeness
- Checkout preview → place order → payment callback happy path
- Customer order isolation (403 on other user's order)

---

## Recommended implementation order

```mermaid
flowchart LR
    A[Fix sort + category_slug] --> B[Homepage aggregate]
    B --> C[DTO normalization *_toman]
    C --> D[payment_url + PSP]
    D --> E[Per-SKU inventory]
    F[Account profile]:::done
    G[About + navigation]:::done
    H[Payment callback]:::done
    classDef done fill:#d4edda,stroke:#28a745
```

1. **Next:** Swagger regen, cache invalidation hooks, `payment_url` + PSP  
2. **Then:** `*_toman` normalization on account orders, per-SKU inventory  
3. **Then:** Persian validation messages, `seo` block on product detail  
4. **Later:** Versioned store DTO package, background jobs for email

**Completed:** Account profile, navigation, about, related products, shipping methods, checkout settings, payment callback, wishlist shortcuts, blog/contact admin aliases, catalog filters/sort, homepage aggregate, product `?include=`, wishlist/blog field aliases, theme idempotent purchase, contact rate limit, catalog sort indexes.
