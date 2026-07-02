# Order

## Purpose

Represents a customer purchase transaction with line items, pricing breakdown, addresses, payment state, and fulfillment lifecycle. Orders are immutable snapshots — product names, SKUs, and prices are captured at order time.

## Description

The `Order` aggregate (`internal/domain/order/entity.go`) maps to `orders` with child `order_items` and `order_status_history`. Orders link to customers and optionally coupons. Status transitions follow a strict state machine; payment status tracks financial state separately.

Embedded value objects: `Address` (billing/shipping), `Item` (line items), `CustomerSnapshot`, `StatusHistory` (timeline).

**Implementation status:** Full admin CRUD, status management, cancel, refund, invoice. Storefront checkout and date-range filter planned.

## Responsibilities

- Record purchase transactions with price snapshots
- Manage fulfillment lifecycle (pending → delivered)
- Track payment and refund state
- Store billing/shipping addresses as JSONB snapshots
- Maintain audit timeline of status changes
- Generate printable invoices
- Apply coupon discounts at creation time

## Attributes

### Order

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `order_number` | string | No | auto-generated | Unique, max 20 | Human-readable ID (e.g., `ORD-20260628-001`) |
| `customer_id` | UUID | No | — | FK → `customers.id` | Buyer |
| `coupon_id` | UUID | Yes | `NULL` | FK → `coupons.id` SET NULL | Applied coupon |
| `status` | enum | No | `pending` | See status enum | Fulfillment state |
| `payment_status` | enum | No | `unpaid` | `unpaid` \| `paid` \| `refunded` | Payment state |
| `subtotal` | decimal | No | — | ≥ 0 | Sum of line items before discounts |
| `discount_amount` | decimal | No | `0` | ≥ 0 | Coupon discount applied |
| `shipping_amount` | decimal | No | `0` | ≥ 0 | Shipping cost |
| `tax_amount` | decimal | No | `0` | ≥ 0 | Tax |
| `total` | decimal | No | — | ≥ 0 | Final total |
| `notes` | text | Yes | `NULL` | Max 2000 | Internal admin notes |
| `payment_method` | string | Yes | `NULL` | Max 50 | e.g., `card`, `cash`, `bank_transfer` |
| `transaction_id` | string | Yes | `NULL` | Max 100 | External payment reference |
| `billing_address` | JSONB | No | — | Address schema | Billing snapshot |
| `shipping_address` | JSONB | No | — | Address schema | Shipping snapshot |
| `created_at` | timestamp | No | `now()` | — | Placed at |
| `updated_at` | timestamp | No | `now()` | — | Last update |

### Order Status Enum

| Value | Description |
|-------|-------------|
| `pending` | Order placed, awaiting processing |
| `processing` | Being prepared |
| `shipped` | Dispatched |
| `delivered` | Received by customer |
| `cancelled` | Cancelled before completion |
| `refunded` | Refunded after delivery |

### Payment Status Enum

| Value | Description |
|-------|-------------|
| `unpaid` | Payment not received |
| `paid` | Payment confirmed |
| `refunded` | Payment returned |

### Item (line item)

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | — | Line item ID |
| `order_id` | UUID | No | — | FK → `orders.id` CASCADE | Parent order |
| `product_id` | UUID | No | — | FK → `products.id` | Product reference |
| `product_name` | string | No | — | Max 300 | Name at order time |
| `product_sku` | string | No | — | Max 100 | SKU code at order time |
| `quantity` | int | No | — | > 0 | Units ordered |
| `unit_price` | decimal | No | — | ≥ 0 | Price per unit at order time |
| `total_price` | decimal | No | — | ≥ 0 | `quantity × unit_price` |

### StatusHistory

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | — | History entry ID |
| `order_id` | UUID | No | — | FK → `orders.id` CASCADE | Parent order |
| `from_status` | enum | Yes | `NULL` | Status enum | Previous status |
| `to_status` | enum | No | — | Status enum | New status |
| `note` | text | Yes | `NULL` | Max 500 | Change reason |
| `changed_by` | UUID | Yes | `NULL` | FK → `users.id` | Admin who made change |
| `created_at` | timestamp | No | `now()` | — | When changed |

### Address (embedded in order)

| Name | Type | Nullable | Validation | Description |
|------|------|----------|------------|-------------|
| `street` | string | No | Max 300 | Street address |
| `city` | string | No | Max 100 | City |
| `state` | string | Yes | Max 100 | State/province |
| `postal_code` | string | No | Max 20 | Postal code |
| `country` | string | No | Len 2 | ISO country code |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Customer | N:1 | Order buyer |
| Coupon | N:1 | Optional discount |
| Product (via Item) | N:M | Line items |
| User (admin) | N:1 | Status change author |

## Business Rules

### Status Transitions

```
pending    → processing, cancelled
processing → shipped, cancelled
shipped    → delivered
delivered  → refunded
```

1. Refund requires `payment_status = paid` and `status = delivered`.
2. Cancel allowed only for `pending` or `processing`.
3. Refund sets `payment_status = refunded`.
4. Terminal states: `cancelled`, `refunded`, `delivered` (no further transitions except delivered → refunded).
5. `order_number` auto-generated, unique.
6. Line item prices are snapshots; product price changes do not affect existing orders.
7. Coupon `usage_count` incremented on order creation when coupon applied.
8. Customer `total_orders` and `total_spent` updated on order confirmation.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create (manual) | ✅ Implemented | Admin creates order for customer |
| Create (checkout) | ❌ Planned | Storefront checkout |
| Read (list) | ✅ Implemented | Missing `from`/`to` date filter |
| Read (single) | ✅ Implemented | Full detail with timeline |
| Update status | ✅ Implemented | `PATCH /{id}/status` |
| Update notes | ✅ Implemented | `PATCH /{id}/notes` |
| Cancel | ✅ Implemented | `POST /{id}/cancel` |
| Refund | ✅ Implemented | `POST /{id}/refund` |
| Invoice | ✅ Implemented | `GET /{id}/invoice` |
| Delete | ❌ Not supported | Orders are permanent records |

## Required APIs

### Admin (implemented)

All require admin JWT.

#### `GET /api/v1/admin/orders`

**Query:** `page`, `per_page`, `status`, `payment_status`, `search` (order number/customer), `customer_id`, `from` (**planned**), `to` (**planned**)

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "order_number": "ORD-20260628-001",
      "status": "processing",
      "payment_status": "paid",
      "total": 159.97,
      "item_count": 3,
      "customer_id": "uuid",
      "customer_name": "Jane Doe",
      "customer_email": "jane@example.com",
      "created_at": "2026-06-28T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 500, "total_pages": 25 }
}
```

---

#### `POST /api/v1/admin/orders`

Manual order creation.

**Request:**
```json
{
  "customer_id": "uuid",
  "items": [
    { "product_id": "uuid", "quantity": 2 }
  ],
  "coupon_code": "SUMMER20",
  "shipping_amount": 9.99,
  "tax_amount": 12.00,
  "billing_address": {
    "street": "123 Main St",
    "city": "Tehran",
    "state": "Tehran",
    "postal_code": "1234567890",
    "country": "IR"
  },
  "shipping_address": {
    "street": "123 Main St",
    "city": "Tehran",
    "postal_code": "1234567890",
    "country": "IR"
  },
  "payment_method": "card",
  "transaction_id": "txn_abc123",
  "payment_status": "paid",
  "notes": "VIP customer — expedite"
}
```

**Response `201`:** `OrderDetailResponse`

---

#### `GET /api/v1/admin/orders/{id}`

**Response `200`:** Full `OrderDetailResponse` with items, customer snapshot, timeline.

---

#### `PATCH /api/v1/admin/orders/{id}/status`

**Request:**
```json
{
  "status": "shipped",
  "note": "Shipped via DHL, tracking #12345"
}
```

**Response `200`:** `OrderDetailResponse`

**Errors:** `422` invalid transition.

---

#### `PATCH /api/v1/admin/orders/{id}/notes`

**Request:** `{ "notes": "Internal note text" }`

---

#### `POST /api/v1/admin/orders/{id}/cancel`

Cancel pending/processing order.

**Response `200`:** `OrderDetailResponse`

---

#### `POST /api/v1/admin/orders/{id}/refund`

**Request:**
```json
{
  "amount": 159.97,
  "reason": "Customer returned all items"
}
```

**Response `200`:** `OrderDetailResponse`

---

#### `GET /api/v1/admin/orders/{id}/invoice`

**Response `200`:** `OrderInvoiceResponse` with store info and order detail.

### Storefront (planned)

#### `POST /api/v1/store/checkout`

Guest and authenticated checkout. Creates order, decrements inventory.

#### `GET /api/v1/store/account/orders`

Customer order history.

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| Date range filter (`from`/`to`) | High | Admin order list UI requirement |
| Storefront checkout | Critical | `POST /store/checkout` |
| SKU selection on line items | High | Currently uses product default SKU |
| Order confirmation email | Medium | Event `OrderPlaced` handler |
| Stock decrement on place | High | Transactional inventory update |

## Domain Reference

- Entity: `internal/domain/order/entity.go`
- Status: `internal/domain/order/status.go`
- Tables: `orders`, `order_items`, `order_status_history`
