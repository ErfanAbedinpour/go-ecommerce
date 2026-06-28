# API Reference Index

Base URL: `http://localhost:8080/api/v1`

## Authentication

All authenticated requests require:

```
Authorization: Bearer <access_token>
```

## Endpoint Catalogs

| Document | Prefix | Auth |
|----------|--------|------|
| [admin-api.md](./admin-api.md) | `/admin/*` | Bearer + `admin` role |
| [store-api.md](./store-api.md) | `/store/*` | Public / customer |
| [auth-api.md](./auth-api.md) | `/auth/*` | Mixed |

## Conventions

### Pagination

```
?page=1&per_page=20&sort=created_at&order=desc
```

### Error Envelope

```json
{
  "statusCode": 400,
  "path": "/api/v1/...",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable message",
    "details": {}
  }
}
```

### Common Error Codes

| Code | HTTP | When |
|------|------|------|
| `VALIDATION_ERROR` | 400 | Invalid request body or params |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient role |
| `NOT_FOUND` | 404 | Resource not found |
| `CONFLICT` | 409 | Duplicate slug, SKU, coupon code |
| `UNPROCESSABLE` | 422 | Business rule violation |

### List Response Shape

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

## Implementation Status

| API Group | Status |
|-----------|--------|
| Auth | ✅ Implemented |
| Admin Dashboard | ✅ Implemented |
| Admin Products | ✅ Implemented |
| Admin Orders | ✅ Implemented |
| Admin Customers | ✅ Implemented |
| Admin Coupons | ✅ Implemented |
| Admin Settings | ✅ Implemented |
| Admin Users | ✅ Implemented |
| Admin Uploads | ✅ Implemented |
| Store Public | ❌ Not implemented |
| Store Account | ❌ Not implemented |
| Admin Storefront Content | ❌ Not implemented |
| Admin Themes | ❌ Not implemented |
| Admin Blog | ❌ Not implemented |
| Admin Contact Inbox | ❌ Not implemented |

See individual entity docs in [entities/](../entities/) for full endpoint specifications.
