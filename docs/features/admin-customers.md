# Admin Customers (Users)

**Route:** `/users`, `/users/:id`  
**Status:** ✅ Backend implemented  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Manage **storefront customers** (not admin staff accounts). The sidebar label "Users" maps to the customers table — registered shoppers and guest checkout records. Admins view customer lists, profiles, purchase history, and can edit or delete customers without orders.

**Note:** Staff and customer accounts both live in the `users` table; role distinguishes access. Customer commerce data is in `customers` (`/api/v1/admin/customers`).

---

## User Flow

### `/users` — Customer list

1. Load paginated list: `GET /admin/customers`.
2. Search by name/email: `q` param.
3. Filter by type: `registered` | `guest`.
4. Sort by created date, total orders, total spent.
5. Click row → `/users/:id`.

### `/users/:id` — Customer detail

1. Load profile: `GET /admin/customers/{id}`.
2. View addresses, stats (`total_orders`, `total_spent`, `last_order_at`).
3. **Order history tab:** `GET /admin/customers/{id}/orders`.
4. **Edit** → `PUT /admin/customers/{id}`.
5. **Delete** → `DELETE /admin/customers/{id}` (blocked if customer has orders).

---

## Business Logic

### Customer types

| Type | Meaning |
|------|---------|
| `registered` | Account with email/password auth |
| `guest` | Created during guest checkout |

### Aggregates

- `total_orders`, `total_spent` denormalized or computed on read.
- `last_order_at` on detail stats — latest order timestamp.

### Delete policy

- Delete allowed only when `total_orders = 0`.
- Otherwise `422 UNPROCESSABLE`.

### Addresses

- Read-only in admin v1 (managed by customer on storefront account future).
- Displayed on detail for support context.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Edit email to existing email | `409 CONFLICT` |
| Delete customer with orders | `422` |
| Guest customer | Shown with type badge |
| Customer with no addresses | Empty addresses array |
| Search no results | Empty paginated list |
| Invalid UUID in URL | `400` |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Orders | Purchase history, delete guard |
| Auth | Registered customers linked to auth (future storefront) |
| Dashboard | Customer count KPI |

---

## Required APIs

All require Bearer token + `admin` role.

### GET `/api/v1/admin/customers`

**Query:** `page`, `per_page`, `sort`, `order`, `q`, `type`

**Sort fields:** `created_at`, `email`, `first_name`, `last_name`, `total_orders`, `total_spent`, `updated_at`

**Response 200:**

```json
{
  "data": [
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
  ],
  "meta": { "page": 1, "per_page": 20, "total": 3782, "total_pages": 190 }
}
```

**Errors:** `401`, `403`

---

### GET `/api/v1/admin/customers/{id}`

**Response 200:**

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
  "updated_at": "2026-06-01T00:00:00Z",
  "addresses": [
    {
      "id": "uuid",
      "type": "home",
      "street": "123 Main St",
      "city": "NYC",
      "state": "NY",
      "postal_code": "10001",
      "country": "US",
      "is_default": true
    }
  ],
  "stats": {
    "total_orders": 5,
    "total_spent": 499.95,
    "last_order_at": "2026-05-15T14:00:00Z"
  }
}
```

**Errors:** `404`, `401`, `403`

---

### GET `/api/v1/admin/customers/{id}/orders`

**Query:** `page`, `per_page`, `sort`, `order`

**Response 200:**

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
      "created_at": "2026-05-15T14:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

**Errors:** `404`, `401`, `403`

---

### PUT `/api/v1/admin/customers/{id}`

**Request:**

```json
{
  "email": "john.doe@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+1234567890",
  "type": "registered"
}
```

All fields optional (partial update).

**Response 200:** `CustomerDetailResponse`

**Errors:** `400`, `404`, `409`, `401`, `403`

---

### DELETE `/api/v1/admin/customers/{id}`

**Response 204**

**Errors:** `404`, `422` (has orders), `401`, `403`

---

## Database Impact

**Tables:** `customers`, `customer_addresses` (read)

| Operation | Notes |
|-----------|-------|
| UPDATE customers | Profile fields |
| DELETE customers | Hard delete when no orders |

**Related:** `orders.customer_id` FK prevents orphan issues via delete guard.

---

## UI Changes Affecting Backend

| UI element | Backend mapping |
|------------|----------------|
| "Users" sidebar label | Maps to `/customers` API |
| Last order date on profile | `stats.last_order_at` |
| Edit profile modal | PUT with changed fields only |
| Delete button | DELETE; show error if 422 |
| Order history tab | Paginated orders endpoint |
| Guest vs registered badge | `type` field |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `email` | Valid email; max 255; unique |
| `first_name`, `last_name` | Max 100 each |
| `phone` | Max 20 |
| `type` | `registered` \| `guest` |

---

## Permission Requirements

| Action | Role |
|--------|------|
| All customer endpoints | `admin` |

Admin staff management (`/admin/users`) also requires `admin` but is a separate UI.

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| List search/filters | URL query | `/users` |
| Customer detail | React Query by `id` | `/users/:id` |
| Order history pagination | URL query on tab | Detail |
| Edit form | React state in modal | Detail |
