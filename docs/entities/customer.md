# Customer

## Purpose

Represents storefront buyers — both registered accounts and guest checkout profiles. Customers are the purchasers linked to orders, addresses, wishlists, and product reviews.

## Description

The `Customer` aggregate (`internal/domain/customer/entity.go`) maps to the `customers` table. Customers are managed from the admin panel under "Users" and will gain self-service account endpoints for the storefront. Customer records accumulate purchase statistics (`total_orders`, `total_spent`, `last_order_at`) updated when orders are placed.

Embedded value object: `Address` (shipping/billing addresses stored in `customer_addresses`).

**Implementation status:** Admin read/update/delete implemented. Storefront account APIs planned.

## Responsibilities

- Identify buyers across orders and engagement features
- Store contact information and delivery addresses
- Track lifetime purchase metrics for admin insights
- Distinguish registered vs guest checkout profiles

## Attributes

### Customer

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `email` | string | No | — | Required, valid email, max 255 | Contact email (not unique enforced at DB level) |
| `first_name` | string | No | — | Required, max 100 | Given name |
| `last_name` | string | No | — | Required, max 100 | Family name |
| `phone` | string | Yes | `NULL` | Max 20 | Phone number |
| `type` | enum | No | `registered` | `registered` \| `guest` | Account origin |
| `total_orders` | int | No | `0` | ≥ 0 | Denormalized order count |
| `total_spent` | decimal | No | `0` | ≥ 0 | Denormalized lifetime spend |
| `last_order_at` | timestamp | Yes | `NULL` | ISO 8601 UTC | Computed from latest order (domain field; may need DB column) |
| `created_at` | timestamp | No | `now()` | — | First seen |
| `updated_at` | timestamp | No | `now()` | — | Last update |

### Address (embedded)

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | — | Address ID |
| `customer_id` | UUID | No | — | FK → `customers.id` | Owner |
| `type` | enum | No | — | `home` \| `work` \| `billing` \| `shipping` | Address purpose |
| `street` | string | No | — | Required, max 300 | Street address |
| `city` | string | No | — | Required, max 100 | City |
| `state` | string | Yes | `NULL` | Max 100 | State/province |
| `postal_code` | string | No | — | Required, max 20 | ZIP/postal code |
| `country` | string | No | — | Required, ISO 3166-1 alpha-2 (len 2) | Country code |
| `is_default` | bool | No | `false` | — | Default address for type |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Address | 1:N | Customer shipping/billing addresses |
| Order | 1:N | All purchases by this customer |
| WishlistItem (planned) | 1:N | Saved products |
| ProductReview (planned) | 1:N | Reviews submitted (nullable for guests) |

## Business Rules

1. Guest customers are created automatically during guest checkout with `type = guest`.
2. When a user signs up or logs in with the same email or phone as an existing guest profile, the guest customer row is promoted to `type = registered` and linked via `user_id`. Existing orders remain attached to that customer id.
3. Guest checkout is blocked when the submitted email or phone belongs to an active registered account (`ACCOUNT_EXISTS_LOGIN_REQUIRED`).
2. `total_orders` and `total_spent` are incremented atomically when an order is confirmed.
3. Only one address per customer may have `is_default = true` per address type (recommended).
4. Deleting a customer is blocked if they have non-terminal orders (recommended).
5. Customer email is indexed but not unique — guests may share patterns; deduplication is application-level.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ⚠️ Implicit | Created via checkout/signup, not direct admin create |
| Read (list) | ✅ Implemented | Admin list with search |
| Read (single) | ✅ Implemented | Includes addresses + stats |
| Update | ✅ Implemented | Admin can edit profile fields |
| Delete | ✅ Implemented | Hard delete |
| List orders | ✅ Implemented | `GET /admin/customers/{id}/orders` |

## Required APIs

### Admin (implemented)

All require `Authorization: Bearer <jwt>`, `role = admin`.

#### `GET /api/v1/admin/customers`

**Query params:** `page`, `per_page`, `search` (email/name), `type`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "email": "buyer@example.com",
      "first_name": "Jane",
      "last_name": "Doe",
      "full_name": "Jane Doe",
      "phone": "+1234567890",
      "type": "registered",
      "total_orders": 12,
      "total_spent": 1549.99,
      "created_at": "2026-01-15T00:00:00Z",
      "updated_at": "2026-06-20T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 150, "total_pages": 8 }
}
```

---

#### `GET /api/v1/admin/customers/{id}`

**Response `200`:**
```json
{
  "id": "uuid",
  "email": "buyer@example.com",
  "first_name": "Jane",
  "last_name": "Doe",
  "full_name": "Jane Doe",
  "phone": "+1234567890",
  "type": "registered",
  "total_orders": 12,
  "total_spent": 1549.99,
  "created_at": "2026-01-15T00:00:00Z",
  "updated_at": "2026-06-20T00:00:00Z",
  "addresses": [
    {
      "id": "uuid",
      "type": "shipping",
      "street": "123 Main St",
      "city": "Tehran",
      "state": "Tehran",
      "postal_code": "1234567890",
      "country": "IR",
      "is_default": true
    }
  ],
  "stats": {
    "total_orders": 12,
    "total_spent": 1549.99,
    "last_order_at": "2026-06-20T14:30:00Z"
  }
}
```

---

#### `PUT /api/v1/admin/customers/{id}`

**Request:**
```json
{
  "first_name": "Jane",
  "last_name": "Smith",
  "phone": "+9876543210",
  "email": "jane.smith@example.com"
}
```

**Response `200`:** `CustomerDetailResponse`

---

#### `DELETE /api/v1/admin/customers/{id}`

**Response `204`:** No content.

---

#### `GET /api/v1/admin/customers/{id}/orders`

Paginated order history for customer.

**Query params:** `page`, `per_page`, `status`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "order_number": "ORD-20260620-001",
      "status": "delivered",
      "payment_status": "paid",
      "total": 129.99,
      "item_count": 3,
      "created_at": "2026-06-20T14:30:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 10, "total": 12, "total_pages": 2 }
}
```

### Storefront (planned)

#### `GET /api/v1/store/account/profile`

**Auth:** Bearer JWT, `role = customer`

**Response `200`:** Customer profile with addresses.

#### `PUT /api/v1/store/account/profile`

Update name, phone; manage addresses.

#### `GET /api/v1/store/account/orders`

Customer's own order history.

#### `GET /api/v1/store/account/orders/{id}`

Single order detail for authenticated customer.

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| `last_order_at` DB column | Medium | Domain has field; add migration if not computed at query time |
| Customer address CRUD APIs | High | Admin and storefront address management |
| Customer signup links to Customer record | Done | `users` with role=customer linked via `customers.user_id` |
| Public profile endpoint | High | Storefront account section |

## Domain Reference

- Entity: `internal/domain/customer/entity.go`
- Types: `internal/domain/customer/customer_type.go`, `address_type.go`
- Table: `customers`, `customer_addresses`
