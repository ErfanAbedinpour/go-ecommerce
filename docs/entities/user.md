# User (Admin)

## Purpose

Represents authenticated back-office operators who manage the ecommerce platform. Distinct from storefront **customers** (buyers), though both share the `admin_users` table with role-based access.

## Description

The `User` aggregate (`internal/domain/user/entity.go`) maps to the `admin_users` database table. Admin users authenticate via JWT (login/refresh) and access all `/api/v1/admin/*` routes. Passwords are stored as bcrypt hashes; they are never exposed in API responses.

Related entities: `RefreshToken` (session rotation), `PasswordResetToken` (forgot-password flow).

**Implementation status:** Fully implemented for admin CRUD and auth.

## Responsibilities

- Authenticate into the admin panel and storefront (when role = `customer`)
- Perform privileged catalog, order, and settings operations (admin role)
- Maintain account profile (name, email, phone, active state)
- Track last login timestamp for audit

## Attributes

### User

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `email` | string | No | — | Required, valid email, max 255, unique | Login identifier |
| `password_hash` | string | No | — | bcrypt hash, never returned via API | Stored credential |
| `first_name` | string | No | — | Required, max 100 | Given name |
| `last_name` | string | No | — | Required, max 100 | Family name |
| `phone` | string | Yes | `NULL` | Max 20 | Optional contact number |
| `role` | enum | No | `customer` | One of: `admin`, `customer` | Access level |
| `is_active` | bool | No | `true` | — | Inactive users cannot log in |
| `last_login_at` | timestamp | Yes | `NULL` | ISO 8601 UTC | Updated on successful login |
| `created_at` | timestamp | No | `now()` | — | Record creation time |
| `updated_at` | timestamp | No | `now()` | — | Last modification time |
| `deleted_at` | timestamp | Yes | `NULL` | Soft-delete marker | Not exposed in API |

### RefreshToken (related)

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | — | Token record ID |
| `user_id` | UUID | No | — | FK → `admin_users.id` | Token owner |
| `token_hash` | string | No | — | SHA-256 of raw token | Lookup key |
| `family_id` | UUID | No | — | — | Rotation family for reuse detection |
| `expires_at` | timestamp | No | — | Future UTC | Expiry |
| `revoked_at` | timestamp | Yes | `NULL` | — | Set on logout or rotation |
| `created_at` | timestamp | No | `now()` | — | Issued at |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| RefreshToken | 1:N | One user may have multiple active/revoked refresh tokens |
| Order (StatusHistory) | 1:N | `changed_by` references admin who updated order status |
| ProductQuestion (planned) | 1:N | `answered_by` references admin who answered a Q&A |
| ThemePurchase (planned) | 1:N | `purchased_by` references admin who bought a theme |

## Business Rules

1. Email must be unique among non-deleted users.
2. Password minimum length 8 characters on create/update.
3. Only `admin` role users may access `/api/v1/admin/*` routes.
4. Deactivating a user (`is_active = false`) invalidates future logins but does not revoke existing tokens until logout/refresh cycle.
5. Admin users cannot delete themselves if they are the last active admin (recommended guard).
6. `role` defaults to `customer` on signup; admin creation requires explicit `role: "admin"`.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create | ✅ Implemented | `POST /admin/users` |
| Read (list) | ✅ Implemented | Paginated with search |
| Read (single) | ✅ Implemented | `GET /admin/users/{id}` |
| Update | ✅ Implemented | Partial update; password optional |
| Delete | ✅ Implemented | Soft delete |
| Activate/Deactivate | ⚠️ Via update | Use `is_active` field |

## Required APIs

### Auth (public)

#### `POST /api/v1/auth/login`

Authenticate and receive access + refresh tokens.

**Request:**
```json
{
  "email": "admin@store.com",
  "password": "securepass123"
}
```

**Response `200`:**
```json
{
  "access_token": "<jwt>",
  "refresh_token": "<opaque>",
  "expires_in": 900,
  "user": {
    "id": "uuid",
    "email": "admin@store.com",
    "first_name": "Admin",
    "last_name": "User",
    "role": "admin"
  }
}
```

**Errors:** `401` invalid credentials, `403` account inactive.

---

#### `POST /api/v1/auth/refresh`

Rotate tokens using refresh token.

**Request:** `{ "refresh_token": "<opaque>" }`

**Response `200`:** Same shape as login.

---

#### `POST /api/v1/auth/signup`

Register a new user (role defaults to `customer` unless configured).

---

#### `POST /api/v1/auth/forgot-password` / `POST /api/v1/auth/reset-password`

Password reset flow via email token.

---

### Authenticated (any role)

#### `GET /api/v1/auth/me`

**Auth:** Bearer JWT

**Response `200`:** `AdminUserResponse` (no password).

#### `POST /api/v1/auth/logout`

**Auth:** Bearer JWT — revokes refresh token family.

---

### Admin CRUD

All endpoints require `Authorization: Bearer <jwt>` and `role = admin`.

#### `GET /api/v1/admin/users`

List admin users with pagination.

**Query params:** `page` (default 1), `per_page` (default 20, max 100), `search` (email/name), `role`, `is_active`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "email": "admin@store.com",
      "first_name": "Admin",
      "last_name": "User",
      "full_name": "Admin User",
      "phone": "+1234567890",
      "role": "admin",
      "is_active": true,
      "last_login_at": "2026-06-28T10:00:00Z",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-06-28T10:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 5, "total_pages": 1 }
}
```

---

#### `POST /api/v1/admin/users`

**Request:**
```json
{
  "email": "newadmin@store.com",
  "password": "securepass123",
  "first_name": "New",
  "last_name": "Admin",
  "phone": "+1234567890",
  "role": "admin",
  "is_active": true
}
```

**Response `201`:** `AdminUserResponse`

**Errors:** `409` email exists, `422` validation.

---

#### `GET /api/v1/admin/users/{id}`

**Response `200`:** `AdminUserResponse`

**Errors:** `404` not found.

---

#### `PUT /api/v1/admin/users/{id}`

Partial update. All fields optional.

**Request:**
```json
{
  "email": "updated@store.com",
  "password": "newpassword123",
  "first_name": "Updated",
  "last_name": "Name",
  "phone": "+9876543210",
  "role": "admin",
  "is_active": false
}
```

**Response `200`:** `AdminUserResponse`

---

#### `DELETE /api/v1/admin/users/{id}`

Soft-delete user.

**Response `204`:** No content.

## Needed Changes

| Change | Priority | Details |
|--------|----------|---------|
| Dedicated admin UI | Low | API exists; no admin user management page yet |
| Granular RBAC | Low | Legacy `roles`/`permissions` tables seeded but inactive |
| Audit log on user changes | Medium | `audit_logs` table exists; wire to user CRUD |

## Domain Reference

- Entity: `internal/domain/user/entity.go`
- Role: `internal/domain/user/role.go`
- Table: `admin_users` (migration `000001_init_schema.up.sql`)
