# Store Homepage

> **Route:** `/`  
> **UI:** [store-os-eta.vercel.app](https://store-os-eta.vercel.app/)  
> **Locale:** Persian (fa-IR), RTL  
> **Vertical:** Building materials (tiles, cement, tools, fixtures)

---

## Purpose

The homepage is the primary entry point for the customer storefront. It aggregates CMS-managed content sections (hero, carousels, banners, brands, FAQ, testimonials) with live catalog data (categories, product highlights) and engagement surfaces (contact form, blog teaser, stats counters). A single aggregated API minimizes round-trips and enables edge caching.

**Sections (top → bottom):**

| Section | Source | Admin config |
|---------|--------|--------------|
| Hero (video + CTAs) | `storefront_hero` | `/context/hero` |
| Category grid | `categories` (active) | `/products/settings` |
| Product carousels (3 tabs: bestseller / new / discounted) | `product_slides` + `product_slide_items` | `/context/product-slides` |
| Pro banners | `pro_banners` | `/context/pro-banners` |
| Partner brands | `partner_brands` | `/context/brands` |
| FAQ accordion | `faq_sections` + `faq_items` | `/context/faq` |
| Contact form + section image | `faq_sections` image + `POST /store/contact` | `/context/contact-us` |
| Customer testimonials | `homepage_reviews` | `/context/customer-reviews` |
| Blog teaser (latest 3 posts) | `blog_posts` | `/weblog` |
| Stats counters | Computed from catalog + orders | — |

---

## User Flow

```mermaid
flowchart TD
    A[User lands on /] --> B[Load homepage aggregate API]
    B --> C{Sections render RTL}
    C --> D[Hero: play video, tap CTAs]
    C --> E[Browse category cards]
    C --> F[Switch carousel tabs]
    C --> G[Click banner / brand logo]
    C --> H[Expand FAQ items]
    C --> I[Submit contact form]
    C --> J[Read testimonials]
    C --> K[Click blog teaser → /blog/:slug]
    D --> L[Navigate to /products or external URL]
    E --> M[/products?category=slug]
    F --> N[/products/:id]
    G --> N
    K --> O[/blog/:slug]
    I --> P[Success toast + form reset]
```

1. **Initial load:** Fetch `GET /api/v1/store/homepage` (and optionally `GET /api/v1/store/theme` for CSS variables).
2. **Hero:** Autoplay muted video; primary CTA → catalog or campaign URL; secondary CTA → about/contact.
3. **Categories:** Tap card → `/products?category={slug}`.
4. **Carousels:** Tab labels from slide config (`bestseller`, `new`, `discounted`); swipe/arrow through products; card click → product detail.
5. **Contact form:** Inline validation → `POST /api/v1/store/contact` with `source: "homepage"`.
6. **Blog teaser:** "مشاهده همه" → `/blog`.

---

## Business Logic

### Hero

- Return hero only when `is_active = true`.
- Video URL must be HTTPS; fallback to poster image if video fails (frontend).
- CTAs are optional; hide buttons when text/URL null.

### Categories

- Include only `is_active = true` categories with `products_count > 0` (or show all active with zero-count badge disabled).
- Sort by `sort_order ASC`, limit 8–12 for homepage grid.
- Building-materials taxonomy example: کاشی و سرامیک، سیمان و گچ، ابزار، شیرآلات.

### Product carousels

| Tab (`slide_type`) | Product selection rule |
|--------------------|------------------------|
| `bestseller` | Admin-curated via `product_slide_items`, OR fallback: top N by order line-item quantity (30-day window) |
| `new` | Admin-curated OR fallback: `created_at DESC`, `status = active` |
| `discounted` | Admin-curated OR fallback: `sale_price IS NOT NULL AND sale_price < price` |

- Each slide respects `is_active` on slide and product.
- Product cards show: thumbnail, name, price in **Toman**, sale badge if discounted, stock badge if out of stock.
- `autoplay_interval_ms` from `product_slides` (default 4500).

### Stats counters

Computed server-side (cached 5–15 min):

| Counter (Persian label) | Computation |
|-------------------------|-------------|
| محصولات | `COUNT(products) WHERE status = 'active'` |
| مشتریان | `COUNT(customers)` |
| سفارشات تحویل‌شده | `COUNT(orders) WHERE status = 'delivered'` |
| سال تجربه | Static from `store_settings.site.years_experience` or hardcoded |

### Blog teaser

- Latest 3 `blog_posts` where `status = 'published'`, ordered by `published_at DESC`.
- Fields: `title`, `slug`, `excerpt`, `cover_image_url`, `read_time_minutes`, `published_at`.

### Contact form

- Creates `contact_messages` row; `source = 'homepage'`.
- No auth required; rate-limit by IP (e.g. 5/hour).

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Hero inactive or missing | Skip hero section; layout collapses |
| Empty carousel tab | Show empty state: "محصولی یافت نشد" |
| Product in carousel becomes draft/deleted | Filter out on read; admin should refresh slide items |
| Category image missing | Placeholder icon by category slug |
| Video autoplay blocked (mobile) | Show poster + play button |
| Contact spam / duplicate submits | Rate limit; optional honeypot field |
| Stats API slow | Serve stale cache; skeleton loaders on UI |
| RTL number formatting | Prices use `fa-IR` locale; Toman suffix "تومان" |
| Offline / API error | Error boundary with retry; cached theme tokens if available |

---

## Dependencies

### Backend modules (to implement)

| Module | Responsibility |
|--------|----------------|
| `internal/application/storecontent` | Hero, slides, banners, FAQ, homepage reviews |
| `internal/application/storefront` | Category summary, product card DTOs, stats |
| `internal/application/blog` | Teaser posts |
| `internal/application/contact` | Form submission |
| `internal/application/theme` | Active theme tokens (optional separate call) |

### Existing data

- `products`, `categories`, `product_images`, `inventories`, `skus`
- `store_settings` (site name, logo, years_experience)
- `orders`, `customers` (stats)

### Frontend

- RTL layout (`dir="rtl"`, `lang="fa"`)
- Swiper/carousel component
- Video player with muted autoplay
- React Query or SWR for homepage cache (`staleTime: 60_000`)

### Admin (content authoring)

All sections configured via admin `/context/*` routes (separate admin API; not called from store).

---

## Required APIs

> **Base URL:** `/api/v1/store`  
> **Auth:** Public (no Bearer required)

### `GET /api/v1/store/homepage`

Aggregated homepage payload. Cacheable (`Cache-Control: public, max-age=60`, tag `homepage`).

**Response 200**

```json
{
  "hero": {
    "video_url": "https://cdn.example.com/hero.mp4",
    "poster_url": "https://cdn.example.com/hero-poster.jpg",
    "title": "متریال ساختمانی با کیفیت",
    "subtitle": "کاشی، سیمان، ابزار و بیشتر",
    "cta_primary": { "text": "مشاهده محصولات", "url": "/products" },
    "cta_secondary": { "text": "درباره ما", "url": "/about" }
  },
  "categories": [
    {
      "id": "uuid",
      "name": "کاشی و سرامیک",
      "slug": "tiles",
      "image_url": "https://…",
      "products_count": 42
    }
  ],
  "product_slides": [
    {
      "slide_type": "bestseller",
      "title": "پرفروش‌ترین‌ها",
      "tab_label": "پرفروش",
      "autoplay_interval_ms": 4500,
      "products": [/* StoreProductCard[] */]
    },
    {
      "slide_type": "new",
      "title": "جدیدترین محصولات",
      "tab_label": "جدید",
      "autoplay_interval_ms": 4500,
      "products": []
    },
    {
      "slide_type": "discounted",
      "title": "تخفیف‌دار",
      "tab_label": "تخفیف",
      "autoplay_interval_ms": 4500,
      "products": []
    }
  ],
  "pro_banners": [
    {
      "id": "uuid",
      "desktop_image_url": "https://…",
      "mobile_image_url": "https://…",
      "link_url": "/products?category=cement"
    }
  ],
  "partner_brands": [
    {
      "id": "uuid",
      "title": "ایسیکو",
      "description": "شیرآلات بهداشتی",
      "logo_url": "https://…",
      "link_url": "https://…"
    }
  ],
  "faq": {
    "image_url": "https://…",
    "items": [
      { "id": "uuid", "question": "…", "answer": "…", "sort_order": 0 }
    ]
  },
  "contact_section": {
    "image_url": "https://…"
  },
  "testimonials": [
    {
      "id": "uuid",
      "customer_name": "علی رضایی",
      "photo_url": "https://…",
      "review_text": "…",
      "rating": 5
    }
  ],
  "blog_teaser": {
    "posts": [
      {
        "id": "uuid",
        "title": "راهنمای انتخاب کاشی",
        "slug": "tile-guide",
        "excerpt": "…",
        "cover_image_url": "https://…",
        "read_time_minutes": 5,
        "published_at": "2026-06-01T10:00:00Z"
      }
    ]
  },
  "stats": {
    "products_count": 245,
    "customers_count": 3782,
    "delivered_orders_count": 1100,
    "years_experience": 15
  }
}
```

**`StoreProductCard` shape (reused across store APIs)**

```json
{
  "id": "uuid",
  "slug": "ceramic-tile-60x60",
  "name": "کاشی سرامیک ۶۰×۶۰",
  "thumbnail_url": "https://…",
  "price_toman": 450000,
  "sale_price_toman": 399000,
  "discount_percent": 11,
  "is_on_sale": true,
  "is_out_of_stock": false,
  "brand": "مرجان"
}
```

> **Currency:** All storefront prices are integers in **Toman** (1 Toman = 10 Rial). DB stores `DECIMAL`; API converts: `price_toman = ROUND(price_rial / 10)` or store directly as Toman per business rule.

### `POST /api/v1/store/contact`

Used by homepage contact section (also `/about`).

**Request**

```json
{
  "name": "رضا محمدی",
  "email": "reza@example.com",
  "phone": "09121234567",
  "subject": "استعلام قیمت سیمان",
  "message": "…",
  "source": "homepage"
}
```

**Response 201**

```json
{
  "id": "uuid",
  "message": "پیام شما با موفقیت ارسال شد."
}
```

### Supporting public endpoints (optional separate calls)

| Method | Path | Use on homepage |
|--------|------|-----------------|
| `GET` | `/api/v1/store/theme` | CSS variables, font |
| `GET` | `/api/v1/store/settings` | Site name, logo, footer |
| `GET` | `/api/v1/store/navigation` | Header/footer menu |

---

## Database Impact

### Reads

| Table | Access |
|-------|--------|
| `storefront_hero` | Single active row |
| `product_slides`, `product_slide_items` | Join `products`, `product_images`, `inventories` |
| `pro_banners`, `partner_brands`, `homepage_reviews` | Active rows, sorted |
| `faq_sections`, `faq_items` | Active FAQ items |
| `categories` | Active, with product count |
| `blog_posts` | Published, limit 3 |
| `products`, `orders`, `customers` | Stats aggregation |

### Writes

| Table | Trigger |
|-------|---------|
| `contact_messages` | Contact form submit |

### Migrations required

- `000010_storefront_content` (hero, slides, banners, partner_brands, homepage_reviews, faq)
- `000012_blog` (blog teaser)
- `000013_engagement` (contact_messages)

### Indexes

- `blog_posts(status, published_at DESC)`
- `product_slide_items(slide_id, sort_order)`

---

## Validation

### `GET /homepage`

- No input params.
- Server filters inactive/deleted records.

### `POST /contact`

| Field | Rules |
|-------|-------|
| `name` | Required, 2–255 chars |
| `email` | Required, valid email |
| `phone` | Optional, Iranian mobile regex `^09\d{9}$` |
| `subject` | Optional, max 500 chars |
| `message` | Required, 10–5000 chars |
| `source` | Required enum: `homepage`, `about`, `contact_page` |

**Error responses:** Standard `apperror` envelope (`400` validation, `429` rate limit).

---

## Permissions

| Action | Role |
|--------|------|
| View homepage | Public |
| Submit contact form | Public (rate-limited) |
| Manage homepage content | Admin only (`/api/v1/admin/storefront/*`) |

No customer authentication required for homepage viewing.

---

## State Management

### Frontend (Store OS)

| State | Storage | Notes |
|-------|---------|-------|
| Homepage data | React Query cache key `['homepage']` | `staleTime: 60s`, refetch on window focus optional |
| Active carousel tab | React `useState` | Default: first tab (`bestseller`) |
| Expanded FAQ item IDs | React `useState` | Single-expand or multi-expand per UI |
| Contact form fields | React Hook Form | Reset on success |
| Contact submit status | React `useState` | `idle \| submitting \| success \| error` |
| Theme tokens | CSS variables from `/store/theme` | Applied on `:root` |
| Video play state | Component local | Muted autoplay; user gesture for sound |

### Backend caching

| Resource | TTL | Invalidation tag |
|----------|-----|------------------|
| Homepage aggregate | 60s–5min | `homepage` on any context update |
| Stats subsection | 5–15min | `catalog`, `orders` |
| Blog teaser | 5min | `blog` |

### Cart / auth

Homepage does not mutate cart or wishlist. Header cart badge reads from `localStorage` cart (see `store-checkout.md`).
