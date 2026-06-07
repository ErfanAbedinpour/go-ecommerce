# Frontend API Mapping — [shop-panel-react.vercel.app](https://shop-panel-react.vercel.app/)

> **Validated against:** live UI routes extracted from the deployed bundle (`index-B9c0IS2u.js`) and cross-checked with `README.md` §1.1.  
> **Backend base URL:** `http://localhost:8080/api/v1`  
> **Auth:** Bearer token on all `/admin/*` routes. Admin role required.

## Validation Summary

| Area | UI pages | Backend coverage | Notes |
|------|----------|------------------|-------|
| Authentication | `/signin` | ✅ | Login, signup, forgot/reset password |
| Dashboard `/` | KPIs, chart, recent orders, low stock | ✅ | Growth %, featured products, enriched recent orders |
| Products | `/products`, `/products/create`, `/products/settings` | ✅ | CRUD, stats, stock filter, brands, attributes, uploads |
| Orders | `/orders`, `/orders/:id`, `/orders/:id/invoice`, `/orders/create` | ✅ Partial | List/detail/actions OK; create order + invoice endpoint missing |
| Users | `/users`, `/users/:id` | ✅ Partial | Read-only customers; no admin-user CRUD, no customer edit/delete |
| Coupons | `/coupons` | ✅ Full | All CRUD + activate/deactivate |
| General settings | `/general-setting`, `/navigation` | ✅ | Site, contact, social, navigation APIs |
| SEO | `/setting-seo` | ✅ | SEO settings API |
| Blog | `/weblog/*` | ❌ Out of scope | Marked out-of-scope in README |
| Contact | `/contact` | ❌ Out of scope | Marked out-of-scope in README |
| Categories | (used in product forms / settings) | ✅ Full | No dedicated sidebar page; API exists |
| Audit logs | (not in sidebar) | ❌ None | Phase 9 not implemented |

**Overall:** Core e-commerce admin flows (products, orders, coupons, customers, dashboard analytics) are covered. Settings, CMS, contact, auth signup, admin users, manual order creation, and file uploads are not.

---

## Global Conventions

### Authentication header

```
Authorization: Bearer <access_token>
```

### Standard error envelope

```json
{
  "statusCode": 404,
  "path": "/api/v1/admin/orders/…",
  "error": {
    "code": "NOT_FOUND",
    "message": "order not found",
    "details": {}
  }
}
```

### Pagination meta (list endpoints)

```json
{
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 48,
    "total_pages": 3
  }
}
```

---

## 1. Authentication

### `/signin` — Sign In

| API | Method | Path | Status |
|-----|--------|------|--------|
| Login | `POST` | `/api/v1/auth/login` | ✅ |
| Refresh | `POST` | `/api/v1/auth/refresh` | ✅ |
| Logout | `POST` | `/api/v1/auth/logout` | ✅ |
| Current user | `GET` | `/api/v1/auth/me` | ✅ |

**Login request**

```json
{ "email": "admin@shop.com", "password": "Admin@123456" }
```

**Login / refresh response**

```json
{
  "access_token": "eyJ…",
  "refresh_token": "eyJ…",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

**`GET /auth/me` response** — sidebar user menu, session guard

```json
{
  "id": "c0000000-0000-0000-0000-000000000001",
  "email": "admin@shop.com",
  "first_name": "Admin",
  "last_name": "User",
  "role": "admin"
}
```

### `/signup` — Sign Up

| API | Status |
|-----|--------|
| `POST /api/v1/auth/signup` | ✅ |

**Signup request**

```json
{
  "email": "staff@shop.com",
  "password": "Secret123",
  "first_name": "Staff",
  "last_name": "User",
  "phone": "+1234567890"
}
```

**Signup response 201** — same token shape as login.

Role is assigned from `AUTH_SIGNUP_DEFAULT_ROLE` (`admin` or `customer`).

### Forgot / reset password

| API | Status |
|-----|--------|
| `POST /api/v1/auth/forgot-password` | ✅ |
| `POST /api/v1/auth/reset-password` | ✅ |

**Forgot password request**

```json
{ "email": "admin@shop.com" }
```

**Forgot password response**

```json
{ "message": "If an account with that email exists, a password reset link has been sent." }
```

**Reset password request** — token from email link query `?token=…`

```json
{ "token": "abc123", "password": "NewPass456" }
```

---

## 2. Dashboard — `/`

The home page renders **6 KPI cards**, a **sales & revenue chart**, **featured products**, **recent orders**, and **low stock products**.

### KPI cards → `GET /api/v1/admin/dashboard/stats`

| UI label | Response field | Type |
|----------|----------------|------|
| Total Revenue | `total_revenue` | `float64` |
| Total Orders | `total_orders` | `int64` |
| Total Customers | `total_customers` | `int64` |
| Total Products | `total_products` | `int64` |
| Pending Orders | `pending_orders` | `int64` |
| Low Stock Products | `low_stock_count` | `int64` |

**Example response**

```json
{
  "total_revenue": 24780.00,
  "total_orders": 1248,
  "total_customers": 3782,
  "total_products": 245,
  "pending_orders": 32,
  "low_stock_count": 14,
  "growth": {
    "total_revenue": 12.5,
    "total_orders": 8.3,
    "total_customers": 5.1,
    "total_products": 2.0,
    "pending_orders": -4.2,
    "low_stock_count": 10.0
  }
}
```

`growth` values are % change for the last 30 days vs the prior 30 days.

### Sales & revenue chart → `GET /api/v1/admin/dashboard/revenue`

**Query:** `period=today|week|month|year` or `from=YYYY-MM-DD&to=YYYY-MM-DD`

**Response**

```json
{
  "data": [
    { "date": "2026-06-01", "revenue": 1250.00, "orders": 8 },
    { "date": "2026-06-02", "revenue": 980.50, "orders": 5 }
  ]
}
```

| Field | Use on chart |
|-------|----------------|
| `date` | X-axis label |
| `revenue` | Revenue series |
| `orders` | Sales / orders series (UI label: "Sales") |

### Recent orders table → `GET /api/v1/admin/dashboard/recent-orders`

**Query:** `limit` (default 10, max 50)

**Response**

```json
{
  "data": [
    {
      "id": "uuid",
      "order_number": "ORD-001",
      "status": "delivered",
      "payment_status": "paid",
      "total": 129.99,
      "item_count": 2,
      "customer_name": "John Doe",
      "product_name": "Nike Air Max",
      "created_at": "2026-06-01T10:00:00Z"
    }
  ]
}
```

### Low stock widget → `GET /api/v1/admin/dashboard/low-stock`

**Query:** `page`, `per_page`, `sort`, `order`

**Response:** `ProductListResponse` — same shape as product list:

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Nike Air Max",
      "sku": "PROD-001",
      "price": 129.99,
      "status": "active",
      "inventory": {
        "quantity": 3,
        "low_stock_threshold": 10,
        "is_low_stock": true,
        "is_out_of_stock": false
      },
      "images": [{ "url": "https://…", "sort_order": 0 }]
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 14, "total_pages": 1 }
}
```

### Featured products section → `GET /api/v1/admin/dashboard/featured-products`

**Query:** `limit` (default 5, max 20)

**Response:** `{ "data": [ProductResponse] }` — active featured products only.

---

## 3. Products

### `/products` — All Products

| API | Method | Path | Purpose |
|-----|--------|------|---------|
| List | `GET` | `/api/v1/admin/products` | Table + KPI derivation |
| Search | `GET` | `/api/v1/admin/products/search?q=` | Search box |
| Delete | `DELETE` | `/api/v1/admin/products/{id}` | Row action |

**List query params:** `page`, `per_page`, `sort`, `order`, `status`, `category_id`, `brand`, `is_featured`, `stock_level` (`low`|`out`)

**Product KPI cards → `GET /api/v1/admin/products/stats`**

```json
{ "total": 245, "active": 198, "draft": 32, "out_of_stock": 15 }
```

**List response item (`ProductResponse`)**

```json
{
  "id": "uuid",
  "category_id": "uuid",
  "name": "Nike Air Max",
  "slug": "nike-air-max",
  "sku": "PROD-001",
  "description": "…",
  "short_description": "…",
  "price": 129.99,
  "sale_price": 99.99,
  "brand": "Nike",
  "is_featured": false,
  "status": "active",
  "images": [{ "id": "uuid", "url": "https://…", "alt_text": "", "sort_order": 0 }],
  "attributes": [{ "id": "uuid", "name": "Color", "value": "Black" }],
  "inventory": {
    "quantity": 50,
    "low_stock_threshold": 10,
    "is_low_stock": false,
    "is_out_of_stock": false
  },
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

**UI KPI cards** (total / active / draft / out of stock): derive client-side from list filters or add a dedicated stats endpoint (not implemented).


### `/products/create` and product edit

| API | Method | Path |
|-----|--------|------|
| Create | `POST` | `/api/v1/admin/products` |
| Get one | `GET` | `/api/v1/admin/products/{id}` |
| Update | `PUT` | `/api/v1/admin/products/{id}` |
| Update stock | `PATCH` | `/api/v1/admin/products/{id}/inventory` |

**Create request (abbreviated)**

```json
{
  "name": "Nike Air Max",
  "sku": "PROD-001",
  "price": 129.99,
  "sale_price": 99.99,
  "category_id": "uuid",
  "brand": "Nike",
  "is_featured": false,
  "status": "draft",
  "images": [{ "url": "https://cdn/…/img.jpg", "sort_order": 0 }],
  "attributes": [{ "name": "Color", "value": "Black" }],
  "inventory": { "quantity": 50, "low_stock_threshold": 10 }
}
```

**Category dropdown:** `GET /api/v1/admin/categories?tree=true` or flat list.

**File upload:** `POST /api/v1/admin/uploads` (multipart `file` field) → `{ "url", "filename", "size", "content_type" }`

### `/products/settings` — Product Settings

| API | Status |
|-----|--------|
| `GET/POST/PUT/DELETE /api/v1/admin/categories` | ✅ includes `products_count` |
| `GET/POST/PUT/DELETE /api/v1/admin/product-attributes` | ✅ |
| `GET/POST/PUT/DELETE /api/v1/admin/product-attribute-values` | ✅ (`?attribute_id=`) |
| `GET/POST/PUT/DELETE /api/v1/admin/brands` | ✅ |

**Category response** now includes `products_count` for settings tables.

**Attribute value create body**

```json
{ "attribute_id": "uuid", "value": "Black", "sort_order": 0, "is_active": true }
```

---

## 4. Orders

### `/orders` — All Orders

| API | `GET /api/v1/admin/orders` |

**Query:** `page`, `per_page`, `sort`, `order`, `status`, `payment_status`, `q`

**Gap:** UI has date-range filter (today / this week / this month). Orders list has no `from` / `to` date params.

**List response**

```json
{
  "data": [
    {
      "id": "uuid",
      "order_number": "ORD-001",
      "status": "pending",
      "payment_status": "unpaid",
      "total": 199.99,
      "item_count": 3,
      "customer_id": "uuid",
      "customer_name": "John Doe",
      "customer_email": "john@example.com",
      "created_at": "2026-06-01T12:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 48, "total_pages": 3 }
}
```

### `/orders/:id` — Order Detail

| API | Method | Path |
|-----|--------|------|
| Detail | `GET` | `/api/v1/admin/orders/{id}` |
| Change status | `PATCH` | `/api/v1/admin/orders/{id}/status` |
| Cancel | `POST` | `/api/v1/admin/orders/{id}/cancel` |
| Refund | `POST` | `/api/v1/admin/orders/{id}/refund` |

**Detail response (`OrderDetailResponse`)**

```json
{
  "id": "uuid",
  "order_number": "ORD-001",
  "customer_id": "uuid",
  "coupon_id": null,
  "status": "processing",
  "payment_status": "paid",
  "subtotal": 180.00,
  "discount_amount": 10.00,
  "shipping_amount": 15.00,
  "tax_amount": 14.99,
  "total": 199.99,
  "notes": "",
  "billing_address": {
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "US"
  },
  "shipping_address": { "street": "…", "city": "…", "postal_code": "…", "country": "US" },
  "customer": {
    "id": "uuid",
    "email": "john@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "full_name": "John Doe",
    "phone": "+1…"
  },
  "items": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "product_name": "Nike Air Max",
      "product_sku": "PROD-001",
      "quantity": 2,
      "unit_price": 90.00,
      "total_price": 180.00
    }
  ],
  "timeline": [
    {
      "id": "uuid",
      "from_status": "pending",
      "to_status": "processing",
      "note": "Payment confirmed",
      "changed_by": "admin-uuid",
      "created_at": "2026-06-01T12:05:00Z"
    }
  ],
  "created_at": "2026-06-01T12:00:00Z",
  "updated_at": "2026-06-01T12:05:00Z"
}
```

**Status update request**

```json
{ "status": "shipped", "note": "Shipped via FedEx" }
```

**Refund request**

```json
{ "amount": 99.99, "reason": "Customer request" }
```

**Gaps:**
- UI shows payment method / transaction ID (`TXN: …`) — not stored in backend.
- UI "Save internal note" without status change — no `PATCH /orders/{id}/notes` endpoint.

### `/orders/:id/invoice` — Print Invoice

| API | Status |
|-----|--------|
| `GET /api/v1/admin/orders/{id}/invoice` | ❌ Not implemented |

**Workaround:** Use `GET /api/v1/admin/orders/{id}` — contains all printable fields (items, addresses, totals, customer). Frontend formats invoice layout client-side.

### `/orders/create` — Manual Order Creation

| API | Status |
|-----|--------|
| `POST /api/v1/admin/orders` | ❌ Not implemented |

---

## 5. Users — `/users`, `/users/:id`

The UI **Users** section maps to **storefront customers** (not admin panel accounts).

| API | Method | Path | UI use |
|-----|--------|------|--------|
| List | `GET` | `/api/v1/admin/customers` | `/users` table |
| Detail | `GET` | `/api/v1/admin/customers/{id}` | `/users/:id` profile |
| Order history | `GET` | `/api/v1/admin/customers/{id}/orders` | Order history tab |

**Customer list item**

```json
{
  "id": "uuid",
  "email": "john@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "full_name": "John Doe",
  "phone": "+1…",
  "type": "registered",
  "total_orders": 5,
  "total_spent": 499.95,
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

**Customer detail** adds `addresses[]` and `stats`:

```json
{
  "…CustomerResponse fields…",
  "addresses": [
    {
      "id": "uuid",
      "type": "home",
      "street": "123 Main St",
      "city": "NYC",
      "postal_code": "10001",
      "country": "US",
      "is_default": true
    }
  ],
  "stats": { "total_orders": 5, "total_spent": 499.95 }
}
```

**Gaps:**
- UI has edit/delete user actions — no `PUT` / `DELETE` customer endpoints.
- UI shows `last_order` date on detail — not a dedicated field (derive from order history).
- **Admin users** (`admin_users` table) are separate; no `/api/v1/admin/users` API for staff management.

---

## 6. Coupons — `/coupons`

| API | Method | Path |
|-----|--------|------|
| List | `GET` | `/api/v1/admin/coupons` |
| Create | `POST` | `/api/v1/admin/coupons` |
| Get | `GET` | `/api/v1/admin/coupons/{id}` |
| Update | `PUT` | `/api/v1/admin/coupons/{id}` |
| Delete | `DELETE` | `/api/v1/admin/coupons/{id}` |
| Activate | `PATCH` | `/api/v1/admin/coupons/{id}/activate` |
| Deactivate | `PATCH` | `/api/v1/admin/coupons/{id}/deactivate` |

**Coupon response**

```json
{
  "id": "uuid",
  "code": "SUMMER20",
  "discount_type": "percentage",
  "discount_value": 20,
  "min_order_amount": 50,
  "max_usage": 100,
  "usage_count": 12,
  "expires_at": "2026-12-31T23:59:59Z",
  "is_active": true,
  "is_expired": false,
  "is_exhausted": false,
  "note": "Summer sale",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

**Activate/deactivate response**

```json
{ "is_active": true }
```

✅ Fully covers the coupons page.

---

## 7. General Settings

### `/general-setting` — Base Information

| API | Method | Path |
|-----|--------|------|
| Site settings | `GET/PUT` | `/api/v1/admin/settings/site` |
| Contact settings | `GET/PUT` | `/api/v1/admin/settings/contact` |
| Social settings | `GET/PUT` | `/api/v1/admin/settings/social` |
| File upload (logo) | `POST` | `/api/v1/admin/uploads` ✅ |

**Site settings (PUT body)**

```json
{
  "name": "My Shop",
  "url": "https://shop.example.com",
  "logo_url": "https://cdn.example.com/logo.png",
  "favicon_url": "https://cdn.example.com/favicon.ico"
}
```

### `/navigation` — Navigation Settings

| API | Method | Path |
|-----|--------|------|
| Menu tree | `GET/PUT` | `/api/v1/admin/navigation` |

**Navigation (PUT body)**

```json
{
  "items": [
    { "label": "Home", "url": "/", "sort_order": 0, "is_active": true, "children": [] }
  ]
}
```

---

## 8. SEO — `/setting-seo`

| API | Method | Path |
|-----|--------|------|
| SEO settings | `GET/PUT` | `/api/v1/admin/settings/seo` |

---

## 9. Blog / Contact (out of v1 scope)

| UI route | Status |
|----------|--------|
| `/weblog`, `/weblog/create`, `/weblog/settings`, `/weblog/comments` | ❌ Out of scope per README |
| `/contact` | ❌ Out of scope per README |

---

## 10. Categories (supporting API)

No dedicated sidebar page, but required for product forms and settings.

| API | Method | Path |
|-----|--------|------|
| List / tree | `GET` | `/api/v1/admin/categories?tree=true` |
| CRUD | `POST/GET/PUT/DELETE` | `/api/v1/admin/categories[/{id}]` |

**Category response**

```json
{
  "id": "uuid",
  "parent_id": null,
  "name": "Electronics",
  "slug": "electronics",
  "description": "",
  "image_url": "https://…",
  "sort_order": 0,
  "is_active": true,
  "children": [],
  "created_at": "…",
  "updated_at": "…"
}
```

---

## 11. Infrastructure (not UI pages)

| Endpoint | Purpose |
|----------|---------|
| `GET /healthz` | Liveness |
| `GET /readyz` | Readiness |
| `GET /metrics` | Prometheus |
| `GET /swagger/index.html` | API docs |

---

## Quick Reference — Page → API map

```
/signin          → POST /auth/login, GET /auth/me
/                → GET /dashboard/stats, /revenue, /recent-orders, /low-stock, /featured-products
/products        → GET /products, GET /products/search, DELETE /products/{id}
/products/create → POST /products, GET /categories?tree=true
/products/settings → /categories, /brands, /product-attributes, /product-attribute-values, POST /uploads
/orders          → GET /orders
/orders/:id      → GET /orders/{id}, PATCH /status, POST /cancel, POST /refund
/orders/:id/invoice → GET /orders/{id} (workaround)
/orders/create   → ❌ POST /orders
/users           → GET /customers
/users/:id       → GET /customers/{id}, GET /customers/{id}/orders
/coupons         → Full /coupons CRUD + activate/deactivate
/general-setting → GET/PUT /settings/site, /contact, /social
/navigation      → GET/PUT /navigation
/setting-seo     → GET/PUT /settings/seo
/signup          → POST /auth/signup
```

---

## Recommended integration order for frontend

1. Wire auth (`login` → store tokens → attach Bearer header).
2. Dashboard (stats + revenue + recent orders + low stock).
3. Products list/create/edit with categories tree.
4. Orders list/detail with status actions.
5. Customers list/detail.
6. Coupons CRUD.
7. Defer settings, SEO, blog, contact until backend Phase 11+ APIs exist.
