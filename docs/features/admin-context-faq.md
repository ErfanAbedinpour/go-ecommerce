# Admin Context — FAQ Section

**Route:** `/context/faq`  
**Status:** 🆕 Not implemented (proposed)  
**Base URL:** `http://localhost:8080/api/v1`

---

## Purpose

Configure the storefront homepage FAQ section: a section hero image plus an ordered list of question/answer pairs. Supports accordion display on the Persian RTL storefront.

---

## User Flow

1. Admin opens **Context → FAQ** (`/context/faq`).
2. Page loads aggregated FAQ config: `GET /admin/storefront/faq`.
3. Upload section image (optional) via `POST /admin/uploads?context=faq`.
4. Add/edit/reorder Q&A items in a list editor.
5. Toggle individual items active/inactive.
6. Save entire section → `PUT /admin/storefront/faq` (section image + items array).
7. Delete individual items inline or by omitting from PUT payload (replace strategy).

---

## Business Logic

- **Singleton section** (`faq_sections`) with optional `image_url`.
- **Many FAQ items** (`faq_items`) — globally ordered, not per-section FK beyond implicit singleton.
- Storefront shows only `is_active = true` items, sorted by `sort_order`.
- Accordion: one or multiple open panels (frontend behavior).
- Empty FAQ list with active section → show image only or hide entire block (storefront config).
- Q&A content stored as plain text; markdown/HTML rendering is frontend responsibility.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| No section image | Text-only FAQ block |
| All items inactive | Hide FAQ on storefront |
| Duplicate questions | Allowed |
| Very long answer | No server truncate; UI may collapse |
| Save with zero items | Allowed — clears FAQ list |
| HTML in answer | Sanitize on storefront render (XSS prevention) |

---

## Dependencies

| Module | Usage |
|--------|-------|
| Upload service | Section image |
| Storefront content | FAQ CRUD |
| Store homepage API | Public read |

---

## Required APIs

### Proposed — GET `/api/v1/admin/storefront/faq`

**Response 200:**

```json
{
  "id": "uuid",
  "image_url": "https://cdn.example.com/faq/section.jpg",
  "items": [
    {
      "id": "uuid",
      "question": "How long does delivery take?",
      "answer": "Delivery within Tehran takes 1-2 business days.",
      "sort_order": 0,
      "is_active": true
    },
    {
      "id": "uuid",
      "question": "Do you offer returns?",
      "answer": "Unopened items can be returned within 7 days.",
      "sort_order": 1,
      "is_active": true
    }
  ],
  "updated_at": "2026-06-01T10:00:00Z"
}
```

**Errors:** `401`, `403`

---

### Proposed — PUT `/api/v1/admin/storefront/faq`

**Request:**

```json
{
  "image_url": "https://cdn.example.com/faq/section.jpg",
  "items": [
    {
      "id": "uuid",
      "question": "How long does delivery take?",
      "answer": "Delivery within Tehran takes 1-2 business days.",
      "sort_order": 0,
      "is_active": true
    },
    {
      "question": "New question without id",
      "answer": "New answer",
      "sort_order": 1,
      "is_active": true
    }
  ]
}
```

**Behavior:** Upsert section; replace items (update by `id`, create if no `id`, delete omitted).

**Response 200:** Full FAQ object as GET.

**Errors:** `400`, `422`, `401`, `403`

---

## Database Impact

**Tables:** `faq_sections`, `faq_items` (migration `000010`)

| Table | Notes |
|-------|-------|
| `faq_sections` | Singleton row; `image_url` optional |
| `faq_items` | `question`, `answer` TEXT NOT NULL; `sort_order`, `is_active` |

**Cache:** Invalidate `homepage`.

---

## UI Changes Affecting Backend

| UI element | Backend impact |
|------------|----------------|
| Section image uploader | `image_url` on section |
| Accordion item editor | Each item = question + answer + active |
| Drag reorder | Updates `sort_order` in PUT items array |
| Add/remove buttons | Include/exclude items in PUT payload |

---

## Validation Requirements

| Field | Rule |
|-------|------|
| `image_url` | Optional; valid URL; max 500 |
| `items` | Max 50 items |
| `items[].question` | Required; min 3; max 500 chars |
| `items[].answer` | Required; min 3; max 5000 chars |
| `items[].sort_order` | Integer `>= 0` |
| `items[].is_active` | Boolean |
| `items[].id` | Optional UUID on update |

---

## Permission Requirements

| Action | Role |
|--------|------|
| View/edit FAQ | `admin` |

---

## State Management

| State | Storage | Scope |
|-------|---------|-------|
| FAQ form (image + items) | React state | Page |
| Expanded accordion in editor | React state | Editor |
| Dirty tracking | React state | Unsaved warning |
| Saved config | React Query | Refetch after PUT |
