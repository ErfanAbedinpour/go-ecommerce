# FAQ

## Purpose

Manages the Frequently Asked Questions section on the storefront — a section-level image plus an ordered list of question/answer pairs displayed in an accordion UI.

## Description

Maps to `faq_sections` (singleton section config with optional image) and `faq_items` (individual Q&A entries). Edited from admin Context hub (`/context/faq`).

**Implementation status:** Not implemented. Planned in migration `000010_storefront_content`.

## Responsibilities

- Display FAQ accordion on homepage or dedicated section
- Provide section hero/banner image
- Manage ordered list of questions and answers
- Control individual FAQ item visibility

## Attributes

### FaqSection (singleton)

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `image_url` | string | Yes | `NULL` | Valid URL, max 500 | Section decorative image |
| `updated_at` | timestamp | No | `now()` | — | Last section update |

### FaqItem

| Name | Type | Nullable | Default | Validation | Description |
|------|------|----------|---------|------------|-------------|
| `id` | UUID | No | `gen_random_uuid()` | UUID v4 | Primary key |
| `question` | text | No | — | Required, 5–500 chars | FAQ question |
| `answer` | text | No | — | Required, 10–5000 chars | FAQ answer (HTML allowed) |
| `sort_order` | int | No | `0` | ≥ 0 | Display order |
| `is_active` | bool | No | `true` | — | Visible when true |

## Relationships

| Related Entity | Cardinality | Description |
|----------------|-------------|-------------|
| Homepage aggregate | 1:1 section + 1:N items | FAQ block in homepage or about |

## Business Rules

1. FAQ section is singleton (one row in `faq_sections`).
2. Items managed as full list replacement on PUT (or individual CRUD).
3. Only active items returned on public API, sorted by `sort_order`.
4. Minimum 1 active item recommended when FAQ section is displayed.
5. HTML sanitization required for `answer` field on public render.
6. Cache tag `homepage` invalidated on update.

## Required CRUD Operations

| Operation | Status | Notes |
|-----------|--------|-------|
| Read section + items | ❌ Planned | Combined GET |
| Update section | ❌ Planned | Image URL |
| Create item | ❌ Planned | Or bulk via PUT |
| Update items | ❌ Planned | Bulk replace |
| Delete item | ❌ Planned | |

## Required APIs

### Admin

All require admin JWT.

#### `GET /api/v1/admin/storefront/faq`

**Response `200`:**
```json
{
  "section": {
    "id": "uuid",
    "image_url": "/uploads/faq-section.jpg",
    "updated_at": "2026-06-28T00:00:00Z"
  },
  "items": [
    {
      "id": "uuid",
      "question": "What is your return policy?",
      "answer": "We offer 30-day returns on all unused items...",
      "sort_order": 0,
      "is_active": true
    }
  ]
}
```

---

#### `PUT /api/v1/admin/storefront/faq`

Replace entire FAQ configuration (section + items).

**Request:**
```json
{
  "image_url": "/uploads/faq-section.jpg",
  "items": [
    {
      "id": "uuid",
      "question": "What is your return policy?",
      "answer": "We offer 30-day returns...",
      "sort_order": 0,
      "is_active": true
    },
    {
      "question": "How long does shipping take?",
      "answer": "Standard shipping takes 3-5 business days.",
      "sort_order": 1,
      "is_active": true
    }
  ]
}
```

**Response `200`:** Full FAQ object. Items without `id` are created; omitted existing IDs are deleted.

**Errors:** `422` validation.

### Storefront (public)

#### `GET /api/v1/store/homepage`

**Response fragment:**
```json
{
  "faq": {
    "image_url": "/uploads/faq-section.jpg",
    "items": [
      {
        "id": "uuid",
        "question": "What is your return policy?",
        "answer": "We offer 30-day returns on all unused items..."
      }
    ]
  }
}
```

## Domain Reference (planned)

- Package: `internal/domain/storefront/faq/`
- Tables: `faq_sections`, `faq_items`
