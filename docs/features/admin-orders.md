# Admin Orders

**Routes:** `/orders`, `/orders/:id`, `/orders/create`, `/orders/:id/invoice`  
**Status:** ✅ Backend implemented  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Order management lets admins view and filter all orders, inspect full order detail with timeline, create manual orders (phone/in-store sales), update fulfillment status, save internal notes, process refunds, cancel orders, and print invoices.

---

## User Flow

### `/orders` — Order list

1. Load paginated list: `GET /admin/orders`.
2. Apply filters: status, payment status, search (`q`), date range (`from`/`to` or UI presets today/week/month).
3. Click row → `/orders/:id`.

### `/orders/:id` — Order detail

1. Load `GET /admin/orders/{id}`.
2. View customer info, line items, addresses, payment fields, timeline.
3. Actions:
   - Change status → `PATCH /admin/orders/{id}/status`
   - Save internal note → `PATCH /admin/orders/{id}/notes`
   - Cancel → `POST /admin/orders/{id}/cancel`
   - Refund → `POST /admin/orders/{id}/refund`
4. Print invoice → navigate to `/orders/:id/invoice`.

### `/orders/create` — Manual order

1. Search/select customer (from customers list).
2. Add line items (product search).
3. Enter billing/shipping addresses, shipping/tax amounts.
4. Optional: coupon code, payment method, transaction ID, payment status.
5. Submit → `POST /admin/orders` → redirect to detail.

### `/orders/:id/invoice` — Print invoice

1. Load `GET /admin/orders/{id}/invoice`.
2. Render print layout with store branding + order payload.
3. Browser print (`window.print()`).

---

## Business Logic

### Order statuses (fulfillment)

```
pending → processing → shipped → delivered
pending → cancelled
processing → cancelled
delivered → refunded (also sets payment_status = refunded)
```

| Status | Terminal? |
|--------|-----------|
| `pending` | No |
| `processing` | No |
| `shipped` | No |
| `delivered` | Yes |
| `cancelled` | Yes |
| `refunded` | Yes |

Invalid transitions return `422 INVALID_STATUS`.

### Payment status

| Value | Meaning |
|-------|---------|
| `unpaid` | Awaiting payment |
| `paid` | Payment received |
| `refunded` | Refund processed |

Refund transition requires `status = delivered` AND `payment_status = paid`.

### Cancel

- Allowed when `status` is `pending` or `processing`.
- Restores product inventory for all line items.

### Manual order creation

1. Validates customer exists.
2. Resolves products; uses first SKU code if product has variants.
3. Applies coupon if valid (checks expiry, usage limit, min order).
4. Computes subtotal, discount, shipping, tax, total.
5. Decrements inventory per line item.
6. Records initial timeline entry.
7. Stores `payment_method`, `transaction_id`, `payment_status` from request.

### Date filter

- `from`: orders with `created_at >= from 00:00:00 UTC`
- `to`: orders with `created_at <= to 23:59:59 UTC`
- UI presets map to computed `from`/`to` dates client-side.

### Invoice

- `invoice_number` generated server-side (e.g. `INV-{order_number}`).
- Includes store info from site + contact settings.
- Embeds full `OrderDetailResponse`.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Refund amount > order total | `422` validation error |
| Cancel shipped order | `422` — not cancellable |
| Create order with insufficient stock | `422` — insufficient inventory |
| Invalid coupon on create | `422` — coupon invalid/expired/exhausted |
| Customer not found on create | `404` |
| Product deleted after order placed | Line items retain snapshot name/SKU |
| Empty date range results | Empty paginated list |
| `from` > `to` | `400` validation error |
| Partial refund | Allowed; multiple refunds not tracked separately in v1 — use single refund endpoint |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Customers | Customer lookup for create + detail |
| Products | Line item resolution, stock check/decrement |
| Coupons | Discount on manual create |
| Settings (site/contact) | Invoice store branding |
| Inventory | Stock decrement/restore |
| Dashboard | Recent orders widget |

---

## Required APIs

All require Bearer token + `admin` role.

### GET `/api/v1/admin/orders`

**Query:**

| Param | Type | Notes |
|-------|------|-------|
| `page`, `per_page` | int | Pagination |
| `sort` | string | `created_at`, `order_number`, `total`, `status`, `payment_status` |
| `order` | string | `asc`, `desc` |
| `status` | string | Fulfillment status filter |
| `payment_status` | string | Payment filter |
| `q` | string | Search order number or customer name/email |
| `from` | string | `YYYY-MM-DD` |
| `to` | string | `YYYY-MM-DD` |

**Response 200:**

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

**Errors:** `401`, `403`

---

### GET `/api/v1/admin/orders/{id}`

**Response 200:**

```json
{
  "id": "uuid",
  "order_number": "ORD-001",
  "customer_id": "uuid",
  "coupon_id": null,
  "status": "processing",
  "payment_status": "paid",
  "payment_method": "card",
  "transaction_id": "TXN-123456",
  "subtotal": 180.00,
  "discount_amount": 10.00,
  "shipping_amount": 15.00,
  "tax_amount": 14.99,
  "total": 199.99,
  "notes": "VIP customer",
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

**Errors:** `401`, `403`, `404`

---

### POST `/api/v1/admin/orders`

**Request:**

```json
{
  "customer_id": "uuid",
  "items": [{ "product_id": "uuid", "quantity": 2 }],
  "coupon_code": "SUMMER20",
  "shipping_amount": 15.00,
  "tax_amount": 14.99,
  "billing_address": {
    "street": "123 Main St",
    "city": "New York",
    "state": "NY",
    "postal_code": "10001",
    "country": "US"
  },
  "shipping_address": {
    "street": "123 Main St",
    "city": "New York",
    "postal_code": "10001",
    "country": "US"
  },
  "payment_method": "card",
  "transaction_id": "TXN-123456",
  "payment_status": "paid",
  "notes": "Phone order"
}
```

**Response 201:** `OrderDetailResponse`

**Errors:** `400`, `404`, `422`

---

### PATCH `/api/v1/admin/orders/{id}/status`

**Request:**

```json
{ "status": "shipped", "note": "Shipped via FedEx" }
```

**Response 200:** `OrderDetailResponse`

**Errors:** `400`, `404`, `422` (invalid transition)

---

### PATCH `/api/v1/admin/orders/{id}/notes`

**Request:**

```json
{ "notes": "Internal note text" }
```

**Response 200:** `OrderDetailResponse`

**Errors:** `400`, `404`

---

### POST `/api/v1/admin/orders/{id}/cancel`

No body. Restores inventory.

**Response 200:** `OrderDetailResponse`

**Errors:** `404`, `422`

---

### POST `/api/v1/admin/orders/{id}/refund`

**Request:**

```json
{ "amount": 99.99, "reason": "Customer request" }
```

**Response 200:** `OrderDetailResponse`

**Errors:** `400`, `404`, `422`

---

### GET `/api/v1/admin/orders/{id}/invoice`

**Response 200:**

```json
{
  "invoice_number": "INV-ORD-001",
  "issued_at": "2026-06-01T14:00:00Z",
  "store": {
    "name": "My Shop",
    "url": "https://shop.example.com",
    "logo_url": "https://cdn.example.com/logo.png",
    "email": "info@shop.com",
    "phone": "+1…",
    "address": "123 Store St",
    "city": "NYC",
    "country": "US"
  },
  "order": { /* OrderDetailResponse */ }
}
```

**Errors:** `401`, `403`, `404`

---

## Database Impact

| Table | Operations |
|-------|------------|
| `orders` | INSERT (create), UPDATE (status, notes, payment fields) |
| `order_items` | INSERT on create |
| `order_status_history` | INSERT on status change |
| `inventory` | UPDATE decrement (create) / increment (cancel) |
| `coupons` | UPDATE `usage_count` on create with coupon |

**Index:** `idx_orders_created_at` on `orders(created_at DESC)` for date filters.

---

## UI Changes Affecting Backend

| UI element | Backend mapping |
|------------|-----------------|
| Date filter presets (today/week/month) | Compute `from`/`to` dates client-side |
| Payment method dropdown | `payment_method` on create; read-only on detail |
| Transaction ID field | `transaction_id` on create |
| Internal notes textarea | `PATCH …/notes` (separate from status note) |
| Status dropdown | Only show valid next statuses per transition map |
| Invoice print page | Use dedicated invoice endpoint for store branding |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `customer_id` | Required UUID on create |
| `items` | Min 1 item; `product_id` UUID; `quantity > 0` |
| Addresses | `street`, `city`, `postal_code` required; `country` ISO 2-letter |
| `payment_status` | `unpaid` \| `paid` |
| `status` (update) | Valid enum + allowed transition |
| `notes` | Max 2000 chars |
| `note` (status) | Max 500 chars |
| Refund `amount` | `> 0`, `<= order.total` |
| Refund `reason` | 3–500 chars |
| `from`/`to` | Valid dates, `from <= to` |

---

## Permission Requirements

| Action | Role |
|--------|------|
| All order endpoints | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| List filters (status, dates, search) | URL query params | `/orders` |
| Order detail | React Query keyed by `id` | `/orders/:id` |
| Create form draft | `sessionStorage` | `/orders/create` |
| Selected customer/products | React state | Create flow |
| Invoice print layout | Ephemeral; data from invoice API | `/orders/:id/invoice` |
| Optimistic status update | Optional React state; revert on 422 | Detail page |
