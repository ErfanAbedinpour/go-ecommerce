# Admin Coupons

**Route:** `/coupons`  
**Status:** ✅ Backend fully implemented  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Create and manage discount coupons for the storefront checkout and manual admin orders. Supports percentage and fixed-amount discounts, usage limits, expiry dates, minimum order amounts, and activate/deactivate toggles without deletion.

---

## User Flow

1. Admin opens `/coupons`.
2. List loads: `GET /admin/coupons`.
3. **Create coupon** → modal/form → `POST /admin/coupons`.
4. **Edit** → `GET /admin/coupons/{id}` + `PUT /admin/coupons/{id}`.
5. **Toggle active** → `PATCH …/activate` or `PATCH …/deactivate`.
6. **Delete** → `DELETE /admin/coupons/{id}` (if no usage or policy allows).

List displays: code, discount, usage count/limit, expiry, active status, computed flags (`is_expired`, `is_exhausted`).

---

## Business Logic

### Discount types

| Type | Calculation |
|------|-------------|
| `percentage` | `discount = subtotal * (discount_value / 100)` |
| `fixed_amount` | `discount = discount_value` (capped at subtotal) |

### Validation at checkout / manual order

1. Coupon must be `is_active = true`.
2. Not expired (`expires_at > now` or null).
3. Not exhausted (`usage_count < max_usage` or `max_usage` null).
4. Order subtotal `>= min_order_amount`.
5. On successful order: increment `usage_count`.

### Activate / deactivate

- Deactivate preserves coupon for history; validation fails for new orders.
- Reactivate checks expiry still valid.

### Delete

- Allowed when no orders reference coupon, or soft policy (implementation may block if `usage_count > 0`).

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Duplicate coupon code | `409 CONFLICT` |
| Percentage > 100 | Validate max 100 on create |
| Expired coupon | `is_expired: true`; auto-deactivate optional (daily job) |
| max_usage reached | `is_exhausted: true`; reject at checkout |
| Coupon on manual order | `coupon_code` in `POST /admin/orders` |
| Fixed discount > order total | Discount capped at subtotal |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Orders | Coupon application on create |
| Checkout (future) | `POST /store/coupons/validate` |
| Scheduled job | `ExpireCoupons` daily (future) |

---

## Required APIs

All require Bearer token + `admin` role.

### GET `/api/v1/admin/coupons`

**Query:** `page`, `per_page`, `sort`, `order`, `is_active`, `q`

**Response 200:**

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
      "usage_count": 12,
      "expires_at": "2026-12-31T23:59:59Z",
      "is_active": true,
      "is_expired": false,
      "is_exhausted": false,
      "note": "Summer sale",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-06-01T00:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 15, "total_pages": 1 }
}
```

**Errors:** `401`, `403`

---

### POST `/api/v1/admin/coupons`

**Request:**

```json
{
  "code": "SUMMER20",
  "discount_type": "percentage",
  "discount_value": 20,
  "min_order_amount": 50,
  "max_usage": 100,
  "expires_at": "2026-12-31T23:59:59Z",
  "note": "Summer sale",
  "is_active": true
}
```

**Response 201:** Coupon object.

**Errors:** `400`, `409`, `401`, `403`

---

### GET `/api/v1/admin/coupons/{id}`

**Response 200:** Single coupon.

**Errors:** `404`, `401`, `403`

---

### PUT `/api/v1/admin/coupons/{id}`

**Request:** Partial fields (all optional pointers in DTO).

**Response 200:** Updated coupon.

**Errors:** `400`, `404`, `409`, `401`, `403`

---

### DELETE `/api/v1/admin/coupons/{id}`

**Response 204**

**Errors:** `404`, `422` (if referenced), `401`, `403`

---

### PATCH `/api/v1/admin/coupons/{id}/activate`

**Response 200:**

```json
{ "is_active": true }
```

**Errors:** `404`, `401`, `403`

---

### PATCH `/api/v1/admin/coupons/{id}/deactivate`

**Response 200:**

```json
{ "is_active": false }
```

**Errors:** `404`, `401`, `403`

---

## Database Impact

**Table:** `coupons` (existing)

| Column | Notes |
|--------|-------|
| `code` | UNIQUE |
| `usage_count` | Incremented on order |
| `max_usage`, `expires_at` | Nullable = unlimited / no expiry |

**Related:** `orders.coupon_id` FK

---

## UI Changes Affecting Backend

| UI element | Backend mapping |
|------------|-----------------|
| Code input | Uppercase normalization recommended client-side |
| Discount type toggle | `percentage` vs `fixed_amount` |
| Usage progress bar | `usage_count / max_usage` |
| Expired badge | Use `is_expired` computed field |
| Active switch | activate/deactivate PATCH endpoints |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `code` | Required; 3–50 chars; alphanumeric |
| `discount_type` | `percentage` \| `fixed_amount` |
| `discount_value` | Required; `> 0`; percentage `<= 100` |
| `min_order_amount` | `>= 0` |
| `max_usage` | Optional; `> 0` if set |
| `expires_at` | Optional ISO 8601 datetime |
| `note` | Max 500 chars |

---

## Permission Requirements

| Action | Role |
|--------|------|
| Full coupon CRUD | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Coupon list | React Query | `/coupons` |
| Create/edit modal | React state | Modal |
| Active filter | URL query | List |
| Optimistic toggle | React state; revert on error | Row switch |
