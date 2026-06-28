# Admin Contact Inbox

**Route:** `/contact`  
**Status:** 🆕 Not implemented (proposed)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Inbox for inbound contact form submissions from the storefront (homepage, about page, contact section). Admins view messages, mark as read, archive, and delete. Messages are created by the public `POST /store/contact` endpoint — not from the admin panel.

---

## User Flow

1. Admin opens **Contact** in sidebar (`/contact`).
2. List loads with unread count badge: `GET /admin/contact-messages?status=unread`.
3. Default view shows all messages, newest first.
4. Filter tabs: All | Unread | Read | Archived.
5. Click row → expand detail or navigate to `/contact/:id` (if routed).
6. Mark read → `PATCH /admin/contact-messages/{id}/read`.
7. Archive → `PATCH /admin/contact-messages/{id}/archive` (proposed) or update status.
8. Delete → `DELETE /admin/contact-messages/{id}`.

---

## Business Logic

### Message sources

| Source | Origin |
|--------|--------|
| `homepage` | Homepage contact form |
| `about` | About page form |
| `contact_page` | Dedicated contact page |

### Status lifecycle

```
unread → read → archived
```

- New submissions always `unread`.
- `read` set when admin opens message or explicitly marks read.
- `archived` for handled messages; hidden from default inbox filter.
- Delete is hard delete (v1); consider soft delete in v2.

### Notifications (future)

- Event `ContactMessageReceived` → email notify admin (Milestone 5.5).

### Storefront submission

Public endpoint creates record:

```json
POST /api/v1/store/contact
{
  "name": "…",
  "email": "…",
  "phone": "…",
  "subject": "…",
  "message": "…",
  "source": "homepage"
}
```

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Spam submissions | Rate limit on public POST; honeypot field (frontend) |
| Missing phone/subject | Optional fields; store null |
| Invalid email on submit | `400` on storefront POST |
| Delete unread message | Allowed |
| Empty inbox | Empty state UI |
| Long message body | Store full TEXT; UI scroll in detail |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Contact module | Message CRUD |
| Store contact API | Public POST creates messages |
| Email service | Optional admin notification (future) |

---

## Required APIs

### Proposed — GET `/api/v1/admin/contact-messages`

**Query:**

| Param | Notes |
|-------|-------|
| `page`, `per_page` | Pagination |
| `status` | `unread`, `read`, `archived` |
| `source` | Filter by source |
| `q` | Search name, email, subject, message |
| `from`, `to` | Date range on `created_at` |

**Response 200:**

```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Mohammad Hosseini",
      "email": "m.hosseini@example.com",
      "phone": "+989121234567",
      "subject": "Bulk order inquiry",
      "message": "I need a quote for 500 bags of cement…",
      "source": "homepage",
      "status": "unread",
      "created_at": "2026-06-01T09:30:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 34, "total_pages": 2 }
}
```

**Errors:** `401`, `403`

---

### Proposed — GET `/api/v1/admin/contact-messages/{id}`

**Response 200:** Single message (full body).

**Side effect:** Optionally auto-mark as `read` on GET (configurable).

**Errors:** `404`, `401`, `403`

---

### Proposed — PATCH `/api/v1/admin/contact-messages/{id}/read`

**Request:** Empty body.

**Response 200:**

```json
{ "id": "uuid", "status": "read" }
```

**Errors:** `404`, `401`, `403`

---

### Proposed — PATCH `/api/v1/admin/contact-messages/{id}/archive`

**Response 200:**

```json
{ "id": "uuid", "status": "archived" }
```

**Errors:** `404`, `401`, `403`

---

### Proposed — DELETE `/api/v1/admin/contact-messages/{id}`

**Response 204**

**Errors:** `404`, `401`, `403`

---

### Proposed — GET `/api/v1/admin/contact-messages/stats`

**Response 200:**

```json
{ "unread_count": 5, "total_count": 34 }
```

For sidebar badge.

---

### Public — POST `/api/v1/store/contact`

**Auth:** None (rate limited)

**Request:**

```json
{
  "name": "Mohammad Hosseini",
  "email": "m.hosseini@example.com",
  "phone": "+989121234567",
  "subject": "Bulk order inquiry",
  "message": "Message body…",
  "source": "homepage"
}
```

**Response 201:**

```json
{ "message": "Thank you. We will respond shortly." }
```

**Errors:** `400`, `429` (rate limit)

---

## Database Impact

**Table:** `contact_messages` (migration `000013`)

| Column | Notes |
|--------|-------|
| `name`, `email` | Required |
| `phone`, `subject` | Optional |
| `message` | TEXT NOT NULL |
| `source` | Default `homepage` |
| `status` | `unread`, `read`, `archived` |

**Index:** `(status, created_at DESC)`

---

## UI Changes Affecting Backend

| UI element | Backend impact |
|------------|----------------|
| Unread badge in sidebar | Stats endpoint or count in list meta |
| Filter tabs | `status` query param |
| Search box | `q` param |
| Mark read on open | GET detail or PATCH read |
| Reply via email | Out of scope v1 — external mail client using `email` field |

---

## Validation Requirements

### Admin (read/update/delete)

- `id` valid UUID

### Storefront POST

| Field | Rule |
|-------|------|
| `name` | Required; max 255 |
| `email` | Required; valid email |
| `phone` | Optional; max 50 |
| `subject` | Optional; max 500 |
| `message` | Required; min 10; max 5000 |
| `source` | Enum: `homepage`, `about`, `contact_page` |

---

## Permission Requirements

| Action | Role |
|--------|------|
| View/manage messages | `admin` |
| Submit contact form | Public (rate limited) |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| Inbox filters | URL query `?status=unread` | `/contact` |
| Selected message | URL param or React state | Detail panel |
| Unread count | React Query (poll 60s or websocket future) | Sidebar badge |
| List pagination | URL query | Inbox |
