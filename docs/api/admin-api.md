# Admin API Reference

> **Auth:** `Authorization: Bearer <token>` + `admin` role required for all endpoints below.

## Dashboard

| Method | Route | Handler | Status |
|--------|-------|---------|--------|
| GET | `/admin/dashboard/stats` | KPIs + growth % | ✅ |
| GET | `/admin/dashboard/revenue` | Time-series chart data | ✅ |
| GET | `/admin/dashboard/recent-orders` | Recent orders widget | ✅ |
| GET | `/admin/dashboard/low-stock` | Low stock products | ✅ |
| GET | `/admin/dashboard/featured-products` | Featured products | ✅ |

## Products

| Method | Route | Status |
|--------|-------|--------|
| GET | `/admin/products` | ✅ |
| GET | `/admin/products/stats` | ✅ |
| GET | `/admin/products/search?q=` | ✅ |
| POST | `/admin/products` | ✅ |
| GET | `/admin/products/{id}` | ✅ |
| PUT | `/admin/products/{id}` | ✅ |
| DELETE | `/admin/products/{id}` | ✅ |
| PATCH | `/admin/products/{id}/inventory` | ✅ |

**List filters:** `status`, `category_id`, `brand`, `is_featured`, `stock_level` (`low`|`out`)

## Categories, Brands, Attributes

| Method | Route | Status |
|--------|-------|--------|
| CRUD | `/admin/categories` | ✅ |
| CRUD | `/admin/brands` | ✅ |
| CRUD | `/admin/product-attributes` | ✅ |
| CRUD | `/admin/product-attribute-values` | ✅ |

## Orders

| Method | Route | Status |
|--------|-------|--------|
| GET | `/admin/orders` | ✅ |
| POST | `/admin/orders` | ✅ |
| GET | `/admin/orders/{id}` | ✅ |
| GET | `/admin/orders/{id}/invoice` | ✅ |
| PATCH | `/admin/orders/{id}/status` | ✅ |
| PATCH | `/admin/orders/{id}/notes` | ✅ |
| POST | `/admin/orders/{id}/cancel` | ✅ |
| POST | `/admin/orders/{id}/refund` | ✅ |

**List filters:** `status`, `payment_status`, `q`, `from`, `to` (date range — verify implementation)

## Customers

| Method | Route | Status |
|--------|-------|--------|
| GET | `/admin/customers` | ✅ |
| GET | `/admin/customers/{id}` | ✅ |
| PUT | `/admin/customers/{id}` | ✅ |
| DELETE | `/admin/customers/{id}` | ✅ |
| GET | `/admin/customers/{id}/orders` | ✅ |

## Coupons

| Method | Route | Status |
|--------|-------|--------|
| CRUD | `/admin/coupons` | ✅ |
| PATCH | `/admin/coupons/{id}/activate` | ✅ |
| PATCH | `/admin/coupons/{id}/deactivate` | ✅ |

## Admin Users

| Method | Route | Status |
|--------|-------|--------|
| CRUD | `/admin/users` | ✅ |

## Settings

| Method | Route | Status |
|--------|-------|--------|
| GET/PUT | `/admin/settings/site` | ✅ |
| GET/PUT | `/admin/settings/contact` | ✅ |
| GET/PUT | `/admin/settings/social` | ✅ |
| GET/PUT | `/admin/settings/seo` | ✅ |
| GET/PUT | `/admin/navigation` | ✅ |

## Uploads

| Method | Route | Status |
|--------|-------|--------|
| POST | `/admin/uploads` | ✅ Multipart `file` field |

---

## Proposed — Storefront Content

| Method | Route | Status |
|--------|-------|--------|
| GET/PUT | `/admin/storefront/hero` | ❌ |
| GET/PUT | `/admin/storefront/product-slides` | ❌ |
| GET/PUT | `/admin/storefront/pro-banners` | ❌ |
| CRUD | `/admin/storefront/partner-brands` | ❌ |
| CRUD | `/admin/storefront/homepage-reviews` | ❌ |
| GET/PUT | `/admin/storefront/faq` | ❌ |
| GET/PUT | `/admin/storefront/contact-section` | ❌ |
| GET/PUT | `/admin/storefront/navigation` | ❌ |

## Proposed — Themes

| Method | Route | Status |
|--------|-------|--------|
| GET | `/admin/themes` | ❌ |
| POST | `/admin/themes/{id}/purchase` | ❌ |
| GET/PUT | `/admin/store-style` | ❌ |

## Proposed — Blog

| Method | Route | Status |
|--------|-------|--------|
| CRUD | `/admin/blog/posts` | ❌ |
| CRUD | `/admin/blog/categories` | ❌ |
| GET | `/admin/blog/comments` | ❌ |
| PATCH | `/admin/blog/comments/{id}/approve` | ❌ |
| PATCH | `/admin/blog/comments/{id}/reject` | ❌ |

## Proposed — Contact & Moderation

| Method | Route | Status |
|--------|-------|--------|
| GET | `/admin/contact-messages` | ❌ |
| PATCH | `/admin/contact-messages/{id}/read` | ❌ |
| GET | `/admin/product-reviews` | ❌ |
| PATCH | `/admin/product-reviews/{id}/approve` | ❌ |
| GET | `/admin/product-questions` | ❌ |
| PUT | `/admin/product-questions/{id}/answer` | ❌ |

Full request/response specs: see [entities/](../entities/) and [features/](../features/).
