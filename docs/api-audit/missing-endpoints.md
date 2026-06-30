# Missing Endpoints

Endpoints required by the frontend but absent from `router.go` / Swagger.

---

## Store — Account Profile

**Status:** ✅ Implemented

**Purpose:** Load and edit customer profile and saved addresses on `/account`.

**Endpoint:** `GET /api/v1/store/account/profile`  
**Method:** GET  
**Authentication:** Bearer JWT, role `customer`  
**Request:** —  
**Response:**

```json
{
  "email": "user@shop.com",
  "first_name": "علی",
  "last_name": "رضایی",
  "phone": "09121234567",
  "addresses": [{ "id", "label", "city", "address_line", "postal_code", "is_default" }],
  "stats": { "total_orders": 5, "total_spent_toman": 12500000 }
}
```

**Validation:** Token required  
**Errors:** `401`, `403`  
**Priority:** P0

---

**Endpoint:** `PUT /api/v1/store/account/profile`  
**Method:** PUT  
**Authentication:** Bearer JWT, role `customer`  
**Request:** `{ first_name, last_name, phone, addresses[] }`  
**Response:** Updated profile object  
**Priority:** P0

---

## Store — About Page

**Status:** ✅ Implemented

**Purpose:** Render `/about` with company story, team, milestones, stats.

**Endpoint:** `GET /api/v1/store/about`  
**Method:** GET  
**Authentication:** None  
**Request:** —  
**Response:** `{ hero, story, mission, vision, milestones[], team[], stats, contact, social, seo }`  
**Validation:** —  
**Errors:** `500`  
**Priority:** P1

---

## Store — Public Navigation

**Status:** ✅ Implemented

**Purpose:** CMS-driven header menu on storefront (mega-menu links).

**Endpoint:** `GET /api/v1/store/navigation`  
**Method:** GET  
**Authentication:** None  
**Request:** —  
**Response:** `{ "items": [{ "id", "label", "url", "sort_order", "is_active", "children": [] }] }`  
**Validation:** —  
**Errors:** —  
**Priority:** P1

---

## Store — Related Products

**Status:** ✅ Implemented

**Purpose:** "محصولات مرتبط" section on product detail page.

**Endpoint:** `GET /api/v1/store/products/{id}/related`  
**Method:** GET  
**Authentication:** None  
**Query:** `limit` (default 8)  
**Response:** `{ "data": [StoreProductCard...] }`  
**Validation:** Product must exist and be active (accepts UUID or slug)  
**Errors:** `404`  
**Priority:** P2

---

## Store — Product Search (autocomplete)

**Status:** ✅ Implemented

**Purpose:** Quick search suggestions in header.

**Endpoint:** `GET /api/v1/store/products/search`  
**Method:** GET  
**Authentication:** None  
**Query:** `q` (required), `limit` (default 10)  
**Response:** `{ "data": [{ "id", "slug", "name", "thumbnail_url", "price_toman" }] }`  
**Priority:** P2

---

## Store — Public Brands

**Status:** ✅ Implemented

**Purpose:** Brand filter chips on catalog (optional).

**Endpoint:** `GET /api/v1/store/brands`  
**Method:** GET  
**Authentication:** None  
**Response:** `{ "data": [{ "id", "name", "slug", "logo_url" }] }`  
**Priority:** P3

---

## Store — Shipping Methods

**Status:** ✅ Implemented

**Purpose:** Checkout step 2 — select delivery option.

**Endpoint:** `GET /api/v1/store/checkout/shipping-methods`  
**Method:** GET  
**Authentication:** None  
**Query:** `city` (required)  
**Response:**

```json
{
  "data": [
    { "code": "post", "label": "پست پیشتاز", "price_toman": 85000, "eta_days": "2-4" },
    { "code": "courier", "label": "پیک", "price_toman": 120000, "eta_days": "1" }
  ]
}
```

**Priority:** P1

---

## Store — Checkout Settings

**Status:** ✅ Implemented

**Purpose:** Minimum order amount, enabled payment methods, COD availability.

**Endpoint:** `GET /api/v1/store/settings/checkout`  
**Method:** GET  
**Authentication:** None  
**Response:** `{ min_order_toman, payment_methods[], cod_enabled, cod_cities[], currency_label }`  
**Priority:** P2

---

## Store — Payment Callback

**Status:** ✅ Implemented

**Purpose:** PSP redirect callback after online payment.

**Endpoint:** `POST /api/v1/store/checkout/payment/callback`  
**Method:** POST  
**Authentication:** PSP signature verification (`signature` body field or `X-Payment-Signature` header; HMAC-SHA256 when `PAYMENT_CALLBACK_SECRET` is set)  
**Request:** `{ order_id, authority, status, signature? }` — `status` is `OK` or `NOK`  
**Response:** `{ order_id, payment_status }`  
**Errors:** `400`, `401`, `404`, `409`  
**Priority:** P0

---

## Store — Wishlist Shortcuts

**Status:** ✅ Implemented

**Purpose:** Header wishlist badge and quick heart-state checks.

**Endpoint:** `GET /api/v1/store/account/wishlist/ids`  
**Method:** GET  
**Authentication:** Bearer, customer  
**Response:** `{ "product_ids": ["uuid", ...] }`  
**Priority:** P3

**Endpoint:** `GET /api/v1/store/account/wishlist/count`  
**Method:** GET  
**Authentication:** Bearer, customer  
**Response:** `{ "count": 12 }`  
**Priority:** P3

---

## Store — Blog Path Aliases (optional)

**Status:** ✅ Implemented

**Purpose:** Match frontend docs that use `/blog` instead of `/blog/posts`.

**Endpoint:** `GET /api/v1/store/blog` → alias to `GET /store/blog/posts`  
**Endpoint:** `GET /api/v1/store/blog/{slug}/comments` → resolve slug to post ID  
**Priority:** P3 (frontend can adapt to existing paths)

---

## Admin — Contact Inbox Stats

**Status:** ✅ Implemented

**Purpose:** Unread badge on `/contact` nav item.

**Endpoint:** `GET /api/v1/admin/contact-messages/stats`  
**Method:** GET  
**Authentication:** Bearer, admin  
**Response:** `{ "unread_count": 5, "total_count": 142 }`  
**Priority:** P2

---

## Admin — Contact Read/Archive Aliases (optional)

**Status:** ✅ Implemented

**Purpose:** Match UI buttons that call dedicated routes.

**Endpoint:** `PATCH /api/v1/admin/contact-messages/{id}/read`  
**Endpoint:** `PATCH /api/v1/admin/contact-messages/{id}/archive`  

*Note: Unified `PATCH .../status` already works; these are convenience aliases.*

**Priority:** P3

---

## Admin — Blog Comment Approve/Reject Aliases (optional)

**Status:** ✅ Implemented

**Endpoint:** `PATCH /api/v1/admin/blog/comments/{id}/approve`  
**Endpoint:** `PATCH /api/v1/admin/blog/comments/{id}/reject`  

*Note: `PATCH .../status` with `{ status: "approved" }` already works.*

**Priority:** P3

---

## Summary by priority

| Priority | Count | Endpoints |
|----------|------:|-----------|
| P0 | 4 | account profile GET/PUT, payment callback |
| P1 | 4 | about, navigation, shipping-methods, related products |
| P2 | 4 | contact stats, checkout settings, product search, brands |
| P3 | 6 | wishlist shortcuts, blog aliases, moderation aliases |
