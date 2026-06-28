# Store Account

> **Route:** `/account`  
> **UI:** [store-os-eta.vercel.app/account](https://store-os-eta.vercel.app/account)  
> **Locale:** Persian (fa-IR), RTL

---

## Purpose

The account area lets authenticated customers manage their **profile** and view **order history**. It is the post-login hub linking to wishlist (`/account/wishlist`), order detail, and logout. Unauthenticated visitors are redirected to login with `returnUrl=/account`.

---

## User Flow

```mermaid
flowchart TD
    A[/account] --> B{JWT valid?}
    B -->|No| C[Redirect /login?returnUrl=/account]
    B -->|Yes| D[GET /store/account/profile]
    D --> E[Render profile + orders tabs]
    E --> F{Tab}
    F -->|Profile| G[Edit name, phone, addresses]
    G --> H[PUT /store/account/profile]
    F -->|Orders| I[GET /store/account/orders]
    I --> J[Order list table]
    J --> K[Click row → order detail]
    K --> L[GET /store/account/orders/:id]
    E --> M[Link to /account/wishlist]
```

1. Auth guard on route mount: `GET /api/v1/auth/me` with role `customer`.
2. **Profile tab:** Display email (read-only), editable first/last name, phone; address book with default shipping address.
3. **Orders tab:** Paginated list with order number, date (Jalali display), status badge, total (Toman), item count.
4. Order detail: line items with variant info, addresses, payment status, timeline.
5. Sidebar/nav links: پروفایل، سفارشات، علاقه‌مندی‌ها، خروج.

---

## Business Logic

### Profile

- Email is identity key; change-email flow deferred to v2.
- Phone validated as Iranian mobile.
- Addresses: multiple allowed; one `is_default` for shipping.
- Profile data sourced from `customers` + `customer_addresses` tables.

### Order history

- List orders where `customer_id = current user`.
- Sort `created_at DESC`.
- Status labels in Persian:

| Status | Persian label |
|--------|---------------|
| `pending` | در انتظار |
| `processing` | در حال پردازش |
| `shipped` | ارسال شده |
| `delivered` | تحویل شده |
| `cancelled` | لغو شده |
| `refunded` | مسترد شده |

### Order detail visibility

- Customer may only access own orders (`403` otherwise).
- Include `timeline` from `order_status_history`.
- Line items show variant snapshot: SKU code, variant attributes JSON.

### Registration path

- `POST /api/v1/auth/signup` with `AUTH_SIGNUP_DEFAULT_ROLE=customer`.
- Login: `POST /api/v1/auth/login` → store tokens → fetch profile.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Token expired on page load | Refresh via `POST /auth/refresh`; else redirect login |
| Customer with zero orders | Empty state in orders tab |
| Order still `pending` payment | Show "پرداخت مجدد" link if `payment_url` stored |
| Guest orders before signup | No auto-merge in v1; future: match by email |
| Address delete with only one | Prevent delete or auto-promote another default |
| Admin-cancelled order | Show cancellation reason from timeline note |
| RTL date formatting | `created_at` → Jalali via frontend (`dayjs-jalali`) |

---

## Dependencies

### Backend

| Module | Role |
|--------|------|
| `internal/application/auth` | JWT, `GET /auth/me` |
| `internal/application/storefront/account` | Profile CRUD |
| `internal/application/order` | Customer order queries |

### Tables

- `customers`, `customer_addresses`, `users` (auth link)
- `orders`, `order_items`, `order_status_history`

### Frontend

- Auth context (tokens, user)
- Protected route wrapper
- Jalali date library

---

## Required APIs

### `GET /api/v1/store/account/profile`

**Auth:** Bearer + `customer`

**Response 200**

```json
{
  "id": "uuid",
  "email": "reza@example.com",
  "first_name": "رضا",
  "last_name": "محمدی",
  "full_name": "رضا محمدی",
  "phone": "09121234567",
  "addresses": [
    {
      "id": "uuid",
      "type": "shipping",
      "street": "خیابان ولیعصر، پلاک ۱۲",
      "city": "تهران",
      "state": "تهران",
      "postal_code": "1234567890",
      "country": "IR",
      "is_default": true
    }
  ],
  "stats": {
    "total_orders": 5,
    "total_spent_toman": 12500000
  },
  "created_at": "2025-01-01T00:00:00Z"
}
```

### `PUT /api/v1/store/account/profile`

**Auth:** Bearer + `customer`

**Request**

```json
{
  "first_name": "رضا",
  "last_name": "محمدی",
  "phone": "09121234567",
  "addresses": [
    {
      "id": "uuid",
      "type": "shipping",
      "street": "…",
      "city": "تهران",
      "state": "تهران",
      "postal_code": "1234567890",
      "country": "IR",
      "is_default": true
    }
  ]
}
```

**Response 200:** Same shape as GET profile.

### `GET /api/v1/store/account/orders`

**Auth:** Bearer + `customer`

**Query:** `page`, `per_page`, `status` (optional filter)

**Response 200**

```json
{
  "data": [
    {
      "id": "uuid",
      "order_number": "ORD-1403-001248",
      "status": "delivered",
      "payment_status": "paid",
      "total_toman": 683400,
      "item_count": 3,
      "created_at": "2026-06-01T12:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 10, "total": 5, "total_pages": 1 }
}
```

### `GET /api/v1/store/account/orders/{id}`

**Auth:** Bearer + `customer` (must own order)

**Response 200**

```json
{
  "id": "uuid",
  "order_number": "ORD-1403-001248",
  "status": "delivered",
  "payment_status": "paid",
  "payment_method": "online",
  "transaction_id": "ZP-123456",
  "subtotal_toman": 798000,
  "discount_toman": 159600,
  "shipping_toman": 45000,
  "tax_toman": 0,
  "total_toman": 683400,
  "notes": "",
  "shipping_address": {
    "street": "…",
    "city": "تهران",
    "state": "تهران",
    "postal_code": "1234567890",
    "country": "IR"
  },
  "items": [
    {
      "id": "uuid",
      "product_id": "uuid",
      "product_name": "کاشی سرامیک ۶۰×۶۰",
      "product_sku": "TILE-60-WHT-STN",
      "variant_label": "۶۰×۶۰ · سفید · سنگی",
      "quantity": 2,
      "unit_price_toman": 399000,
      "total_price_toman": 798000
    }
  ],
  "timeline": [
    {
      "from_status": "pending",
      "to_status": "processing",
      "note": "پرداخت تأیید شد",
      "created_at": "2026-06-01T12:05:00Z"
    }
  ],
  "created_at": "2026-06-01T12:00:00Z"
}
```

### Auth endpoints (existing)

| Method | Path | Use |
|--------|------|-----|
| `POST` | `/api/v1/auth/login` | Sign in |
| `POST` | `/api/v1/auth/signup` | Register |
| `POST` | `/api/v1/auth/refresh` | Token refresh |
| `POST` | `/api/v1/auth/logout` | Sign out |
| `GET` | `/api/v1/auth/me` | Session guard |

---

## Database Impact

### Reads

- `customers` JOIN `users` on auth
- `customer_addresses` by `customer_id`
- `orders` WHERE `customer_id = ?`
- `order_items`, `order_status_history`

### Writes

| Table | Operation |
|-------|-----------|
| `customers` | UPDATE name, phone |
| `customer_addresses` | UPSERT/DELETE via profile PUT |

No new tables for account v1.

---

## Validation

### Profile PUT

| Field | Rules |
|-------|-------|
| `first_name`, `last_name` | Required, 2–100 chars |
| `phone` | Optional, `^09\d{9}$` |
| `addresses[].street` | Required if address provided |
| `addresses[].postal_code` | 10 digits |
| `addresses` | Max 5 addresses |

### Order list query

| Param | Rules |
|-------|-------|
| `status` | Valid order status enum |
| `page`, `per_page` | Standard pagination |

---

## Permissions

| Action | Role |
|--------|------|
| View/edit own profile | Customer |
| List own orders | Customer |
| View own order detail | Customer |
| Access `/account` route | Customer (UI guard) |

Admin uses `/api/v1/admin/customers/*` — not accessible from store account.

---

## State Management

### Global auth state

| State | Storage |
|-------|---------|
| `access_token`, `refresh_token` | `localStorage` |
| Current user | React Context from `GET /auth/me` |
| Auth status | `loading \| authenticated \| unauthenticated` |

### Account page state

| State | Storage |
|-------|---------|
| Active tab | URL `?tab=profile\|orders` or path segment |
| Profile form | React Hook Form; dirty tracking |
| Orders list | React Query `['account','orders', page]` |
| Order detail | React Query `['account','orders', id]` |

### Navigation

- Logout: clear tokens, invalidate all queries, redirect `/`
- Link to wishlist: `/account/wishlist` (separate spec)

### Error handling

- `401` → attempt refresh once → login redirect
- `403` on order detail → "دسترسی غیرمجاز"
