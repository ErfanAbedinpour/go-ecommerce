# Missing Endpoints

Endpoints required by the frontend but absent from `router.go` / Swagger.

---

## Store — Account Profile

**Purpose:** Load and edit customer profile and saved addresses on `/account`.

**Suggested endpoint:** `GET /api/v1/store/account/profile`  
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

**Suggested endpoint:** `PUT /api/v1/store/account/profile`  
**Method:** PUT  
**Authentication:** Bearer JWT, role `customer`  
**Request:** `{ first_name, last_name, phone, addresses[] }`  
**Response:** Updated profile object  
**Priority:** P0

---

## Store — About Page

**Purpose:** Render `/about` with company story, team, milestones, stats.

**Suggested endpoint:** `GET /api/v1/store/about`  
**Method:** GET  
**Authentication:** None  
**Request:** —  
**Response:** `{ hero, story, mission, vision, milestones[], team[], stats, contact, social, seo }`  
**Validation:** —  
**Errors:** `500`  
**Priority:** P1

---

## Store — Public Navigation

**Purpose:** CMS-driven header menu on storefront (mega-menu links).

**Suggested endpoint:** `GET /api/v1/store/navigation`  
**Method:** GET  
**Authentication:** None  
**Request:** —  
**Response:** `{ "items": [{ "id", "label", "url", "sort_order", "is_active", "children": [] }] }`  
**Validation:** —  
**Errors:** —  
**Priority:** P1

---

## Store — Related Products

**Purpose:** "محصولات مرتبط" section on product detail page.

**Suggested endpoint:** `GET /api/v1/store/products/{id}/related`  
**Method:** GET  
**Authentication:** None  
**Query:** `limit` (default 8)  
**Response:** `{ "data": [StoreProductCard...] }`  
**Validation:** Product must exist and be active  
**Errors:** `404`  
**Priority:** P2

---

## Store — Product Search (autocomplete)

**Purpose:** Quick search suggestions in header.

**Suggested endpoint:** `GET /api/v1/store/products/search`  
**Method:** GET  
**Authentication:** None  
**Query:** `q` (required), `limit` (default 10)  
**Response:** `{ "data": [{ "id", "slug", "name", "thumbnail_url", "price_toman" }] }`  
**Priority:** P2

---

## Store — Public Brands

**Purpose:** Brand filter chips on catalog (optional).

**Suggested endpoint:** `GET /api/v1/store/brands`  
**Method:** GET  
**Authentication:** None  
**Response:** `{ "data": [{ "id", "name", "slug", "logo_url" }] }`  
**Priority:** P3

---

## Store — Shipping Methods

**Purpose:** Checkout step 2 — select delivery option.

**Suggested endpoint:** `GET /api/v1/store/checkout/shipping-methods`  
**Method:** GET  
**Authentication:** Optional  
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

**Purpose:** Minimum order amount, enabled payment methods, COD availability.

**Suggested endpoint:** `GET /api/v1/store/settings/checkout`  
**Method:** GET  
**Authentication:** None  
**Response:** `{ min_order_toman, payment_methods[], cod_enabled, cod_cities[] }`  
**Priority:** P2

---

## Store — Payment Callback

**Purpose:** PSP redirect callback after online payment.

**Suggested endpoint:** `POST /api/v1/store/checkout/payment/callback`  
**Method:** POST  
**Authentication:** PSP signature verification  
**Request:** Provider-specific (e.g. `authority`, `status`)  
**Response:** `{ order_id, payment_status }`  
**Errors:** `400`, `404`, `409`  
**Priority:** P0

---

## Store — Wishlist Shortcuts

**Purpose:** Header wishlist badge and quick heart-state checks.

**Suggested endpoint:** `GET /api/v1/store/account/wishlist/ids`  
**Method:** GET  
**Authentication:** Bearer, customer  
**Response:** `{ "product_ids": ["uuid", ...] }`  
**Priority:** P3

**Suggested endpoint:** `GET /api/v1/store/account/wishlist/count`  
**Method:** GET  
**Authentication:** Bearer, customer  
**Response:** `{ "count": 12 }`  
**Priority:** P3

---

## Store — Blog Path Aliases (optional)

**Purpose:** Match frontend docs that use `/blog` instead of `/blog/posts`.

**Suggested endpoint:** `GET /api/v1/store/blog` → alias to `GET /store/blog/posts`  
**Suggested endpoint:** `GET /api/v1/store/blog/{slug}/comments` → resolve slug to post ID  
**Priority:** P3 (frontend can adapt to existing paths)

---

## Admin — Contact Inbox Stats

**Purpose:** Unread badge on `/contact` nav item.

**Suggested endpoint:** `GET /api/v1/admin/contact-messages/stats`  
**Method:** GET  
**Authentication:** Bearer, admin  
**Response:** `{ "unread_count": 5, "total_count": 142 }`  
**Priority:** P2

---

## Admin — Contact Read/Archive Aliases (optional)

**Purpose:** Match UI buttons that call dedicated routes.

**Suggested endpoint:** `PATCH /api/v1/admin/contact-messages/{id}/read`  
**Suggested endpoint:** `PATCH /api/v1/admin/contact-messages/{id}/archive`  

*Note: Unified `PATCH .../status` already works; these are convenience aliases.*

**Priority:** P3

---

## Admin — Blog Comment Approve/Reject Aliases (optional)

**Suggested endpoint:** `PATCH /api/v1/admin/blog/comments/{id}/approve`  
**Suggested endpoint:** `PATCH /api/v1/admin/blog/comments/{id}/reject`  

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
