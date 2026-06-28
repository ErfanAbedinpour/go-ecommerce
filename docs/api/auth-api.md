# Auth API Reference

Base prefix: `/api/v1/auth`

## Public Endpoints

### POST /auth/login

Authenticate and receive tokens.

**Request:**
```json
{
  "email": "admin@shop.com",
  "password": "Admin@123456"
}
```

**Response 200:**
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

**Errors:** `401 UNAUTHORIZED` — invalid credentials

---

### POST /auth/refresh

Rotate access token using refresh token.

**Request:**
```json
{
  "refresh_token": "eyJ..."
}
```

**Response 200:** Same shape as login.

**Errors:** `401 UNAUTHORIZED` — invalid or revoked refresh token

---

### POST /auth/signup

Create new account (configurable via `AUTH_SIGNUP_ENABLED`).

**Request:**
```json
{
  "email": "user@example.com",
  "password": "Secret123",
  "first_name": "Ali",
  "last_name": "Mohammadi",
  "phone": "+989121234567"
}
```

**Response 201:** Token shape same as login.

**Role:** Assigned from `AUTH_SIGNUP_DEFAULT_ROLE` (`admin` or `customer`).

---

### POST /auth/forgot-password

**Request:** `{ "email": "user@example.com" }`

**Response 200:** `{ "message": "If an account with that email exists, a password reset link has been sent." }`

Always returns 200 to prevent email enumeration.

---

### POST /auth/reset-password

**Request:**
```json
{
  "token": "reset-token-from-email",
  "password": "NewPassword123"
}
```

**Response 200:** `{ "message": "Password has been reset successfully." }`

---

## Authenticated Endpoints

Require `Authorization: Bearer <access_token>`.

### POST /auth/logout

Revoke refresh token family.

**Request:** `{ "refresh_token": "eyJ..." }`

**Response 204:** No content.

---

### GET /auth/me

Current user profile.

**Response 200:**
```json
{
  "id": "uuid",
  "email": "admin@shop.com",
  "first_name": "Admin",
  "last_name": "User",
  "role": "admin"
}
```

---

## JWT Claims

| Claim | Description |
|-------|-------------|
| `sub` | User ID |
| `email` | User email |
| `role` | `admin` or `customer` |
| `exp` | Expiration timestamp |

## Token Lifetimes

| Token | Default TTL |
|-------|-------------|
| Access | 15 minutes |
| Refresh | 7 days |

Refresh tokens use family-based rotation — reusing a revoked token revokes the entire family.
