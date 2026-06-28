# Contact Message

## Purpose

Stores inbound contact form submissions from the storefront (homepage, about page, contact page). Provides an admin inbox for reading and managing customer inquiries.

## Description

Maps to `contact_messages` table. Messages have a source tag indicating which form submitted them and a status workflow: `unread` → `read` → `archived`.

**Implementation status:** Not implemented. Planned in migration `000013_engagement`.

## Responsibilities

- Capture contact form submissions from storefront
- Track read/unread state for admin inbox
- Categorize by submission source page
- Notify admins on new messages (planned event)

## Attributes

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `name` | string | No | — | Required, max 255 | Sender name |
| `email` | string | No | — | Required, valid email, max 255 | Sender email |
| `phone` | string | Yes | `NULL` | Max 50 | Optional phone |
| `subject` | string | Yes | `NULL` | Max 500 | Message subject |
| `message` | text | No | — | Required, 10–5000 chars | Message body |
| `source` | enum | No | `homepage` | `homepage` \| `about` \| `contact_page` | Form origin |
| `status` | enum | No | `unread` | `unread` \| `read` \| `archived` | Inbox state |
| `created_at` | timestamp | No | `now()` | — | Received at |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Customer | None | No FK; may match by email |
| StoreSettings | — | Contact page content from settings |

## Business Rules

1. New messages default to `status = unread`.
2. Opening a message in admin sets `status = read` (auto or manual).
3. Archived messages hidden from default inbox view.
4. Rate limiting: 3 submissions per hour per IP (recommended).
5. Email format validated; honeypot field for spam prevention (frontend).
6. Event `ContactMessageReceived` notifies admin (planned).
7. No public read API — admin only.
8. Messages are never hard-deleted in v1; archive instead.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Create (public) | ❌ Planned | Form submission |
| Read (list) | ❌ Planned | Admin inbox |
| Read (single) | ❌ Planned | Auto-mark read |
| Mark read | ❌ Planned | PATCH status |
| Archive | ❌ Planned | PATCH status |
| Delete | ❌ Planned | Admin purge |

## Required APIs

### Storefront (public)

#### `POST /api/v1/store/contact`

**Request:**
```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "phone": "+1234567890",
  "subject": "Product inquiry",
  "message": "I would like to know if you ship internationally.",
  "source": "contact_page"
}
```

**Response `201`:**
```json
{
  "id": "uuid",
  "message": "Thank you for your message. We will respond shortly."
}
```

**Errors:** `422` validation, `429` rate limited.

### Admin

All require admin JWT. Route: `/api/v1/admin/contact-messages`.

#### `GET /api/v1/admin/contact-messages`

**Query:** `page`, `per_page`, `status` (default `unread`), `source`, `search`, `from`, `to`

**Response `200`:**
```json
{
  "data": [
    {
      "id": "uuid",
      "name": "Jane Doe",
      "email": "jane@example.com",
      "phone": "+1234567890",
      "subject": "Product inquiry",
      "message_preview": "I would like to know if...",
      "source": "contact_page",
      "status": "unread",
      "created_at": "2026-06-28T15:00:00Z"
    }
  ],
  "meta": { "page": 1, "per_page": 20, "total": 25, "total_pages": 2 }
}
```

---

#### `GET /api/v1/admin/contact-messages/{id}`

Returns full message. Auto-sets `status = read` if currently `unread`.

**Response `200`:**
```json
{
  "id": "uuid",
  "name": "Jane Doe",
  "email": "jane@example.com",
  "phone": "+1234567890",
  "subject": "Product inquiry",
  "message": "I would like to know if you ship internationally.",
  "source": "contact_page",
  "status": "read",
  "created_at": "2026-06-28T15:00:00Z"
}
```

---

#### `PATCH /api/v1/admin/contact-messages/{id}/read`

**Response `200`:** `{ "status": "read" }`

---

#### `PATCH /api/v1/admin/contact-messages/{id}/archive`

**Response `200`:** `{ "status": "archived" }`

---

#### `DELETE /api/v1/admin/contact-messages/{id}`

**Response `204`:** Permanent delete (admin purge).

## Domain Reference (planned)

- Package: `internal/domain/contact/`
- Table: `contact_messages`
- Enum: `contact_message_status`
