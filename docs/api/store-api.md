# Store API Reference

> **Status:** ❌ Not implemented — specification for development.

Base prefix: `/api/v1/store`

## Public Endpoints (No Auth)

### Homepage & Settings

| Method | Route | Purpose |
|--------|-------|---------|
| GET | `/store/homepage` | Aggregated homepage content (hero, slides, banners, brands, reviews, FAQ, blog teaser, stats) |
| GET | `/store/settings` | Site name, logo, contact, social links |
| GET | `/store/theme` | Active theme + color tokens + font |

### Catalog

| Method | Route | Purpose |
|--------|-------|---------|
| GET | `/store/products` | Product list with search, filters, sort |
| GET | `/store/products/{slugOrId}` | Product detail with variants/SKUs |
| GET | `/store/products/{id}/related` | Related products |
| GET | `/store/products/{id}/reviews` | Approved reviews |
| GET | `/store/products/{id}/questions` | Answered Q&A |
| GET | `/store/categories` | Category tree |

**Product list query params:**

| Param | Type | Description |
|-------|------|-------------|
| `q` | string | Full-text search |
| `category_id` | uuid | Filter by category |
| `sort` | enum | `bestseller`, `newest`, `discounted`, `price_asc`, `price_desc` |
| `page` | int | Page number |
| `per_page` | int | Items per page (max 50) |

### Blog

| Method | Route | Purpose |
|--------|-------|---------|
| GET | `/store/blog` | Published posts list |
| GET | `/store/blog/{slug}` | Single post with content |

### About

| Method | Route | Purpose |
|--------|-------|---------|
| GET | `/store/about` | Company story + contact info |

### Checkout

| Method | Route | Purpose |
|--------|-------|---------|
| POST | `/store/checkout/preview` | Validate cart, compute totals |
| POST | `/store/checkout` | Place order |
| POST | `/store/coupons/validate` | Validate coupon code |

### Engagement (Public Write)

| Method | Route | Purpose |
|--------|-------|---------|
| POST | `/store/contact` | Submit contact form |
| POST | `/store/products/{id}/questions` | Ask product question |
| POST | `/store/blog/{slug}/comments` | Submit blog comment (pending moderation) |

---

## Customer-Authenticated Endpoints

Requires `Authorization: Bearer <token>` with `customer` role.

### Account

| Method | Route | Purpose |
|--------|-------|---------|
| GET | `/store/account/profile` | Customer profile |
| PUT | `/store/account/profile` | Update profile |
| GET | `/store/account/orders` | Order history |
| GET | `/store/account/orders/{id}` | Order detail |

### Wishlist

| Method | Route | Purpose |
|--------|-------|---------|
| GET | `/store/account/wishlist` | List wishlist items |
| POST | `/store/account/wishlist` | Add product `{ "product_id": "uuid" }` |
| DELETE | `/store/account/wishlist/{product_id}` | Remove item |

### Reviews

| Method | Route | Purpose |
|--------|-------|---------|
| POST | `/store/products/{id}/reviews` | Submit review (pending moderation) |

---

## Key Request/Response Examples

### GET /store/homepage

```json
{
  "hero": {
    "video_url": "https://cdn.example.com/hero.mp4",
    "title": "مصالح ساختمانی و کاشی‌کاری پاریاب",
    "subtitle": "کاشی، سرامیک، سیمان و ابزار ساختمانی...",
    "cta_primary": { "text": "مشاهده محصولات", "url": "/products" },
    "cta_secondary": { "text": "مشاوره رایگان", "url": "/about" }
  },
  "categories": [{ "id": "uuid", "name": "کاشی و سرامیک", "slug": "tiles", "image_url": "..." }],
  "product_slides": {
    "featured": { "title": "محصولات پیشنهادی", "tabs": [...], "products": [...] },
    "bestseller": { "title": "پرفروش‌ترین‌ها", "products": [...] },
    "discounted": { "title": "تخفیف‌دار", "products": [...] }
  },
  "partner_brands": [{ "title": "...", "logo_url": "...", "description": "..." }],
  "reviews": [{ "customer_name": "...", "review_text": "...", "photo_url": "..." }],
  "faq": { "image_url": "...", "items": [{ "question": "...", "answer": "..." }] },
  "blog_teaser": [{ "title": "...", "slug": "...", "excerpt": "...", "cover_image_url": "..." }],
  "stats": { "years_experience": 20, "product_count": 500, "provinces_covered": 31 }
}
```

### POST /store/checkout

**Request:**

```json
{
  "items": [
    { "product_id": "uuid", "sku_id": "uuid", "quantity": 2 }
  ],
  "coupon_code": "SUMMER20",
  "shipping_address": {
    "street": "تهران، خیابان آیت‌الله کاشانی",
    "city": "تهران",
    "postal_code": "1234567890",
    "country": "IR"
  },
  "payment_method": "online",
  "customer": {
    "email": "customer@example.com",
    "first_name": "علی",
    "last_name": "محمدی",
    "phone": "09121234567"
  },
  "notes": ""
}
```

**Response 201:**

```json
{
  "order_id": "uuid",
  "order_number": "ORD-2026-001234",
  "status": "pending",
  "payment_status": "unpaid",
  "total_toman": 2930000,
  "payment_url": "https://gateway.example.com/pay/..."
}
```

### GET /store/products/{slugOrId}

```json
{
  "id": "uuid",
  "name": "سرامیک پرسلان ۶۰×۶۰ مات",
  "slug": "porcelain-tile-60x60",
  "description": "...",
  "short_description": "...",
  "price_toman": 1250000,
  "sale_price_toman": 1062500,
  "discount_percent": 15,
  "is_new": true,
  "in_stock": true,
  "images": [{ "url": "...", "alt_text": "" }],
  "attributes": [
    { "name": "سایز", "values": ["۶۰×۶۰", "۸۰×۸۰"] },
    { "name": "رنگ", "values": ["خاکستری روشن"] },
    { "name": "وزن", "values": ["۱۸ کیلو", "۲۲ کیلو"] },
    { "name": "طرح", "values": ["سنگی", "سیمانی", "یکدست"] }
  ],
  "skus": [
    {
      "id": "uuid",
      "code": "SKU-P1-60-GRAY",
      "attributes": { "سایز": "۶۰×۶۰", "رنگ": "خاکستری روشن", "وزن": "۱۸ کیلو", "طرح": "سنگی" },
      "price_toman": 1250000,
      "sale_price_toman": 1062500,
      "quantity": 50
    }
  ],
  "specifications": [
    { "label": "ضخامت", "value": "۹ میلی‌متر" },
    { "label": "گارانتی", "value": "۲ سال" }
  ],
  "review_count": 3,
  "question_count": 3,
  "average_rating": 4.5
}
```

---

## Error Cases

| Endpoint | Error | Code |
|----------|-------|------|
| Checkout | Empty cart | `VALIDATION_ERROR` |
| Checkout | Insufficient stock | `UNPROCESSABLE` |
| Checkout | Invalid coupon | `UNPROCESSABLE` |
| Checkout | Invalid SKU combination | `VALIDATION_ERROR` |
| Wishlist | Not authenticated | `UNAUTHORIZED` |
| Wishlist | Product already in list | `CONFLICT` |
| Review | Not a verified buyer | `FORBIDDEN` |
| Contact | Rate limit exceeded | `TOO_MANY_REQUESTS` |

Full specs: [features/](../features/) and [entities/](../entities/).
