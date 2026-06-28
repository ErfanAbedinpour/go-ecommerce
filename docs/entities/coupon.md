# Coupon

## Purpose

Manages discount codes that reduce order totals at checkout. Supports percentage and fixed-amount discounts with usage limits, minimum order thresholds, and expiry dates.

## Description

The `Coupon` aggregate (`internal/domain/coupon/entity.go`) maps to the `coupons` table. Coupon codes are normalized (uppercase, trimmed) on storage. The entity exposes `IsExpired()` and `IsExhausted()` computed checks used during validation.

**Implementation status:** Full admin CRUD with activate/deactivate. Public validation endpoint for checkout planned.

## Responsibilities

- Define discount rules (type, value, constraints)
- Track usage count against max usage limit
- Control availability via `is_active` and `expires_at`
- Apply discounts during order creation (admin manual + planned checkout)

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `code` | string | No | — | Required, 3–50 chars, alphanumeric, unique | Discount code (stored uppercase) |
| `discount_type` | enum | No | — | `percentage` \| `fixed_amount` | Calculation method |
| `discount_value` | decimal | No | — | > 0; ≤ 100 if percentage | Discount amount or percent |
| `min_order_amount` | decimal | No | `0` | ≥ 0 | Minimum subtotal to qualify |
| `max_usage` | int | Yes | `NULL` | > 0 if set | Total redemption limit |
| `usage_count` | int | No | `0` | ≥ 0 | Times redeemed |
| `expires_at` | timestamp | Yes | `NULL` | Future UTC if set | Expiry date |
| `is_active` | bool | No | `true` | — | Manually enabled/disabled |
| `note` | text | Yes | `NULL` | Max 500 | Internal admin note |
| `created_at` | timestamp | No | `now()` | — | Created |
| `updated_at` | timestamp | No | `now()` | — | Updated |
| `deleted_at` | timestamp | Yes | `NULL` | Soft delete | — |

### Computed (API response only)

| Name | Type | Description |
|------|------|-------------|
| `is_expired` | bool | `expires_at` in the past |
| `is_exhausted` | bool | `usage_count >= max_usage` when limit set |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Order | 1:N | Orders referencing `coupon_id` |

## Business Rules

1. Code normalized via `NormalizeCode()`: uppercase + trim.
2. Percentage discounts: `discount_value` must be 0–100; applied as `subtotal × (value / 100)`.
3. Fixed amount: discount capped at subtotal (cannot go negative).
4. Coupon valid only when: `is_active`, not expired, not exhausted, `subtotal ≥ min_order_amount`.
5. `usage_count` incremented atomically on successful order with coupon.
6. One coupon per order.
7. Expired coupons should be deactivated by scheduled job `ExpireCoupons` (planned).

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ✅ Implemented | |
| Read (list) | ✅ Implemented | Paginated |
| Read (single) | ✅ Implemented | |
| Update | ✅ Implemented | |
| Delete | ✅ Implemented | Soft delete |
| Activate | ✅ Implemented | `PATCH /{id}/activate` |
| Deactivate | ✅ Implemented | `PATCH /{id}/deactivate` |
| Validate (public) | ❌ Planned | Checkout validation |

## Required APIs

### Admin (implemented)

All require admin JWT.

#### `GET /api/v1/admin/coupons`

**Query:** `page`, `per_page`, `search` (code), `is_active`, `discount_type`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "code": "SUMMER20",
      "discount_type": "percentage",
      "discount_value": 20,
      "min_order_amount": 50,
      "max_usage": 100,
      "usage_count": 45,
      "expires_at": "2026-08-31T23:59:59Z",
      "is_active": true,
      "is_expired": false,
      "is_exhausted": false,
      "note": "Summer sale campaign",
      "created_at": "2026-06-01T00:00:00Z",
      "updated_at": "2026-06-15T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 10, "total_pages": 1 }
}
```

---

#### `POST /api/v1/admin/coupons`

**Request:**
```json
{
  "code": "SUMMER20",
  "discount_type": "percentage",
  "discount_value": 20,
  "min_order_amount": 50,
  "max_usage": 100,
  "expires_at": "2026-08-31T23:59:59Z",
  "note": "Summer sale campaign",
  "is_active": true
}
```

**Response `201`:** `CouponResponse`

---

#### `GET /api/v1/admin/coupons/{id}`

**Response `200`:** `CouponResponse`

---

#### `PUT /api/v1/admin/coupons/{id}`

Partial update. All fields optional.

---

#### `DELETE /api/v1/admin/coupons/{id}`

**Response `204`:** Soft delete.

---

#### `PATCH /api/v1/admin/coupons/{id}/activate`

**Response `200`:** `{ "is_active": true }`

---

#### `PATCH /api/v1/admin/coupons/{id}/deactivate`

**Response `200`:** `{ "is_active": false }`

### Storefront (planned)

#### `POST /api/v1/store/coupons/validate`

Validate coupon without redeeming.

**Request:**
```json
{
  "code": "SUMMER20",
  "subtotal": 75.00
}
```

**Response `200`:**
```json
{
  "valid": true,
  "code": "SUMMER20",
  "discount_type": "percentage",
  "discount_value": 20,
  "discount_amount": 15.00,
  "message": "Coupon applied successfully"
}
```

**Response `200` (invalid):**
```json
{
  "valid": false,
  "message": "Coupon has expired"
}
```

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| Public validate endpoint | High | Checkout flow dependency |
| Scheduled expiry job | Medium | `ExpireCoupons` daily cron |
| Per-customer usage limit | Low | Optional `max_usage_per_customer` |

## Domain Reference

- Entity: `internal/domain/coupon/entity.go`
- Discount: `internal/domain/coupon/discount.go`, `discount_type.go`
- Table: `coupons`
