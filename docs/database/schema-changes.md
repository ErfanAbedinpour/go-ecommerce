# Database Schema Changes

## Overview

This document lists all database changes required to support the new UI. Existing migrations (000001–000009) cover the admin backend core.

---

## New Tables

### Storefront Content

#### `storefront_hero`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| video_url | VARCHAR(500) | YES | NULL | MP4 URL |
| title | VARCHAR(255) | YES | NULL | Overlay title |
| subtitle | TEXT | YES | NULL | |
| cta_primary_text | VARCHAR(100) | YES | NULL | |
| cta_primary_url | VARCHAR(500) | YES | NULL | |
| cta_secondary_text | VARCHAR(100) | YES | NULL | |
| cta_secondary_url | VARCHAR(500) | YES | NULL | |
| is_active | BOOLEAN | NO | true | |
| updated_at | TIMESTAMPTZ | NO | now() | |

**Indexes:** None (singleton row expected).

---

#### `product_slides`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| slide_type | VARCHAR(50) | NO | — | `featured`, `bestseller`, `discounted` |
| title | VARCHAR(255) | YES | NULL | Section title |
| autoplay_interval_ms | INT | NO | 4500 | |
| is_active | BOOLEAN | NO | true | |
| sort_order | INT | NO | 0 | |
| created_at | TIMESTAMPTZ | NO | now() | |
| updated_at | TIMESTAMPTZ | NO | now() | |

**Unique:** `slide_type` (one row per carousel type).

---

#### `product_slide_items`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| slide_id | UUID | NO | — | FK → product_slides |
| product_id | UUID | NO | — | FK → products |
| sort_order | INT | NO | 0 | |
| tab_label | VARCHAR(100) | YES | NULL | For featured tabs only |

**Indexes:** `(slide_id, sort_order)`

---

#### `pro_banners`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| desktop_image_url | VARCHAR(500) | NO | — | |
| mobile_image_url | VARCHAR(500) | YES | NULL | |
| link_url | VARCHAR(500) | YES | NULL | Click-through |
| sort_order | INT | NO | 0 | |
| is_active | BOOLEAN | NO | true | |
| created_at | TIMESTAMPTZ | NO | now() | |
| updated_at | TIMESTAMPTZ | NO | now() | |

---

#### `partner_brands`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| title | VARCHAR(255) | NO | — | Brand name |
| description | TEXT | YES | NULL | |
| logo_url | VARCHAR(500) | NO | — | |
| link_url | VARCHAR(500) | YES | NULL | |
| sort_order | INT | NO | 0 | |
| is_active | BOOLEAN | NO | true | |
| created_at | TIMESTAMPTZ | NO | now() | |
| updated_at | TIMESTAMPTZ | NO | now() | |

---

#### `homepage_reviews`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| customer_name | VARCHAR(255) | NO | — | |
| photo_url | VARCHAR(500) | YES | NULL | |
| review_text | TEXT | NO | — | |
| rating | SMALLINT | YES | NULL | 1–5 |
| sort_order | INT | NO | 0 | |
| is_active | BOOLEAN | NO | true | |
| created_at | TIMESTAMPTZ | NO | now() | |

---

#### `faq_sections`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| image_url | VARCHAR(500) | YES | NULL | Section image |
| updated_at | TIMESTAMPTZ | NO | now() | |

---

#### `faq_items`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| question | TEXT | NO | — | |
| answer | TEXT | NO | — | |
| sort_order | INT | NO | 0 | |
| is_active | BOOLEAN | NO | true | |

---

### Theme System

#### `store_themes`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| name | VARCHAR(255) | NO | — | |
| slug | VARCHAR(255) | NO | — | UNIQUE |
| description | TEXT | YES | NULL | |
| preview_image_url | VARCHAR(500) | YES | NULL | |
| price | DECIMAL(10,2) | NO | 0 | 0 = free |
| is_active | BOOLEAN | NO | true | |
| default_colors | JSONB | NO | '{}' | 12 color tokens |
| default_font | VARCHAR(100) | YES | NULL | |
| created_at | TIMESTAMPTZ | NO | now() | |

---

#### `theme_purchases`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| theme_id | UUID | NO | — | FK → store_themes |
| purchased_by | UUID | NO | — | FK → admin_users |
| purchased_at | TIMESTAMPTZ | NO | now() | |

**Unique:** `(theme_id, purchased_by)`

---

#### `store_style`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| active_theme_id | UUID | YES | NULL | FK → store_themes |
| colors | JSONB | NO | '{}' | 12 customizable colors |
| font_family | VARCHAR(100) | YES | NULL | |
| updated_at | TIMESTAMPTZ | NO | now() | |

---

### Blog

#### `blog_categories`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| name | VARCHAR(255) | NO | — | |
| slug | VARCHAR(255) | NO | — | UNIQUE |
| sort_order | INT | NO | 0 | |
| is_active | BOOLEAN | NO | true | |

---

#### `blog_posts`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| category_id | UUID | YES | NULL | FK → blog_categories |
| title | VARCHAR(500) | NO | — | |
| slug | VARCHAR(500) | NO | — | UNIQUE |
| excerpt | TEXT | YES | NULL | |
| content | TEXT | NO | — | Markdown/HTML |
| cover_image_url | VARCHAR(500) | YES | NULL | |
| author_name | VARCHAR(255) | YES | NULL | |
| read_time_minutes | INT | YES | NULL | |
| status | VARCHAR(20) | NO | 'draft' | draft, published, archived |
| published_at | TIMESTAMPTZ | YES | NULL | |
| created_at | TIMESTAMPTZ | NO | now() | |
| updated_at | TIMESTAMPTZ | NO | now() | |

**Indexes:** `(status, published_at DESC)`, GIN on `title` for search.

---

#### `blog_comments`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| post_id | UUID | NO | — | FK → blog_posts |
| author_name | VARCHAR(255) | NO | — | |
| author_email | VARCHAR(255) | YES | NULL | |
| content | TEXT | NO | — | |
| status | VARCHAR(20) | NO | 'pending' | pending, approved, rejected |
| created_at | TIMESTAMPTZ | NO | now() | |

**Indexes:** `(post_id, status)`

---

### Engagement

#### `contact_messages`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| name | VARCHAR(255) | NO | — | |
| email | VARCHAR(255) | NO | — | |
| phone | VARCHAR(50) | YES | NULL | |
| subject | VARCHAR(500) | YES | NULL | |
| message | TEXT | NO | — | |
| source | VARCHAR(50) | NO | 'homepage' | homepage, about, contact_page |
| status | VARCHAR(20) | NO | 'unread' | unread, read, archived |
| created_at | TIMESTAMPTZ | NO | now() | |

**Indexes:** `(status, created_at DESC)`

---

#### `wishlist_items`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| customer_id | UUID | NO | — | FK → customers |
| product_id | UUID | NO | — | FK → products |
| created_at | TIMESTAMPTZ | NO | now() | |

**Unique:** `(customer_id, product_id)`

---

#### `product_reviews`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| product_id | UUID | NO | — | FK → products |
| customer_id | UUID | YES | NULL | NULL = guest |
| author_name | VARCHAR(255) | NO | — | |
| rating | SMALLINT | NO | — | 1–5 |
| title | VARCHAR(255) | YES | NULL | |
| content | TEXT | NO | — | |
| status | VARCHAR(20) | NO | 'pending' | pending, approved, rejected |
| created_at | TIMESTAMPTZ | NO | now() | |

**Indexes:** `(product_id, status)`

---

#### `product_questions`

| Column | Type | Nullable | Default | Notes |
|--------|------|----------|---------|-------|
| id | UUID | NO | gen_random_uuid() | PK |
| product_id | UUID | NO | — | FK → products |
| asker_name | VARCHAR(255) | NO | — | |
| asker_email | VARCHAR(255) | YES | NULL | |
| question | TEXT | NO | — | |
| answer | TEXT | YES | NULL | Admin response |
| answered_at | TIMESTAMPTZ | YES | NULL | |
| answered_by | UUID | YES | NULL | FK → admin_users |
| status | VARCHAR(20) | NO | 'open' | open, answered |
| created_at | TIMESTAMPTZ | NO | now() | |

---

## Modifications to Existing Tables

### `orders`

Add index for date range queries:

```sql
CREATE INDEX idx_orders_created_at ON orders (created_at DESC);
```

### `store_settings`

Extend JSONB `navigation` or add `storefront_navigation` key to distinguish admin vs store nav.

### `skus`

Add optional `price_override` and `sale_price_override` columns for per-variant pricing.

```sql
ALTER TABLE skus ADD COLUMN price_override DECIMAL(10,2);
ALTER TABLE skus ADD COLUMN sale_price_override DECIMAL(10,2);
```

---

## New Enums

```sql
CREATE TYPE blog_post_status AS ENUM ('draft', 'published', 'archived');
CREATE TYPE comment_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE contact_message_status AS ENUM ('unread', 'read', 'archived');
CREATE TYPE product_review_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE slide_type AS ENUM ('featured', 'bestseller', 'discounted');
```

---

## Migration Plan

| Migration | Contents |
|-----------|----------|
| 000010_storefront_content | hero, slides, slide_items, banners, partner_brands, homepage_reviews, faq |
| 000011_theme_system | store_themes, theme_purchases, store_style + seed themes |
| 000012_blog | blog_categories, blog_posts, blog_comments |
| 000013_engagement | contact_messages, wishlist_items, product_reviews, product_questions |
| 000014_sku_pricing | skus price overrides |
| 000015_indexes | Performance indexes |

---

## Cache Invalidation Tags

| Tag | Invalidated When |
|-----|------------------|
| `homepage` | Any context section update |
| `catalog` | Product/category change |
| `theme` | Style or theme purchase |
| `blog` | Post publish/unpublish |
