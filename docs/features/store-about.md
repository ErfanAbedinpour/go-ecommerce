# Store About Page

> **Route:** `/about`  
> **UI:** [store-os-eta.vercel.app/about](https://store-os-eta.vercel.app/about)  
> **Locale:** Persian (fa-IR), RTL

---

## Purpose

The about page presents the company's story, mission, values, and team credibility for a building materials retailer. It includes static/markdown content from store settings, visual sections (timeline, stats, team), and a **contact form** that creates the same `contact_messages` records as the homepage form. Builds trust for B2B and B2C buyers purchasing tiles, cement, and tools.

---

## User Flow

```mermaid
flowchart TD
    A[/about] --> B[GET /store/about]
    B --> C[Render company story sections]
    C --> D[Stats / milestones]
    C --> E[Team or values cards]
    C --> F[Embedded contact form]
    F --> G[User fills name, email, phone, message]
    G --> H[POST /store/contact source=about]
    H --> I[Success toast]
    C --> J[Footer links: phone, email, social]
    J --> K[From /store/settings]
```

1. User navigates from header/footer "درباره ما".
2. Page loads CMS content + global contact info.
3. User reads company narrative (RTL typography).
4. Optional: submit contact form for partnership or bulk inquiry.
5. Social icons link to Instagram, WhatsApp, etc. from settings.

---

## Business Logic

### Content sources

| Section | Source |
|---------|--------|
| Page title, hero image | `store_settings.site` extended fields OR dedicated `about` JSON key |
| Company story (rich text) | `store_settings.about.story` (Markdown/HTML) |
| Mission / vision bullets | `store_settings.about.mission`, `vision` |
| Timeline milestones | `store_settings.about.milestones[]` |
| Team members | `store_settings.about.team[]` (name, role, photo) |
| Stats (years, projects) | `store_settings.about.stats` or reuse homepage stats |
| Contact info (phone, email, address) | `store_settings.contact` |
| Social links | `store_settings.social` |

**Assumption:** v1 uses `store_settings` JSONB extension rather than a separate `about_pages` table. Admin edits via General Settings or future `/context/about` page.

### Contact form

- Same endpoint as homepage: `POST /api/v1/store/contact`
- `source = "about"` for admin inbox filtering.
- Rate-limited per IP.

### SEO

- Meta title/description from `store_settings.seo` with about-page overrides.
- Canonical URL `/about`.

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| About content empty in settings | Show sensible Persian placeholder + admin notice in dev |
| Image URLs broken | Fallback placeholder |
| Contact form spam | Rate limit + honeypot field `website` (hidden) |
| Long story text | Typography with readable line-height; table of contents optional |
| Mobile layout | Stack sections vertically; form full width |
| WhatsApp deep link | `https://wa.me/98…` from `contact.whatsapp` |

---

## Dependencies

### Backend

| Module | Role |
|--------|------|
| `internal/application/storefront/content` | About page aggregate |
| `internal/application/settings` | Read `store_settings` |
| `internal/application/contact` | Form submission |

### Tables

- `store_settings` (read)
- `contact_messages` (write on form submit)

### Admin

- Content edited in admin General Settings (v1) or future Context module.
- Messages visible at admin `/contact`.

### Frontend

- RTL prose styling
- Shared `ContactForm` component (homepage + about)
- React Query for about + settings

---

## Required APIs

### `GET /api/v1/store/about`

Aggregated about page content.

**Auth:** Public

**Response 200**

```json
{
  "hero": {
    "title": "درباره ما",
    "subtitle": "بیش از ۱۵ سال تجربه در مصالح ساختمانی",
    "image_url": "https://cdn.example.com/about-hero.jpg"
  },
  "story": {
    "title": "داستان ما",
    "content_html": "<p>ما از سال ۱۳۸۸ …</p>",
    "content_markdown": "ما از سال ۱۳۸۸ …"
  },
  "mission": {
    "title": "ماموریت",
    "text": "تأمین مصالح با کیفیت با بهترین قیمت"
  },
  "vision": {
    "title": "چشم‌انداز",
    "text": "مرجع خرید آنلاین مصالح ساختمانی در ایران"
  },
  "milestones": [
    { "year": "۱۳۸۸", "title": "تأسیس", "description": "شروع فعالیت در تهران" },
    { "year": "۱۴۰۰", "title": "فروشگاه آنلاین", "description": "راه‌اندازی Store OS" }
  ],
  "team": [
    {
      "name": "رضا محمدی",
      "role": "مدیرعامل",
      "photo_url": "https://…"
    }
  ],
  "stats": {
    "years_experience": 15,
    "happy_customers": 3782,
    "products_count": 245,
    "completed_projects": 500
  },
  "contact": {
    "phone": "021-12345678",
    "mobile": "09121234567",
    "email": "info@example.com",
    "address": "تهران، خیابان …",
    "working_hours": "شنبه تا پنج‌شنبه ۹–۱۸"
  },
  "social": {
    "instagram": "https://instagram.com/…",
    "whatsapp": "https://wa.me/989121234567",
    "telegram": "https://t.me/…"
  },
  "seo": {
    "meta_title": "درباره ما | فروشگاه مصالح ساختمانی",
    "meta_description": "…"
  }
}
```

### `POST /api/v1/store/contact`

Shared with homepage (see `store-homepage.md`).

**Request (about page)**

```json
{
  "name": "شرکت ساختمانی آریا",
  "email": "info@arya-build.ir",
  "phone": "09129876543",
  "subject": "همکاری عمده",
  "message": "درخواست لیست قیمت سیمان و کاشی برای پروژه ۵۰۰ واحدی",
  "source": "about"
}
```

**Response 201**

```json
{
  "id": "uuid",
  "message": "پیام شما با موفقیت ارسال شد. به زودی با شما تماس می‌گیریم."
}
```

### `GET /api/v1/store/settings` (supporting)

Public subset for footer/header if not fully included in `/about`:

```json
{
  "site": { "name": "…", "logo_url": "…" },
  "contact": { "phone": "…", "email": "…" },
  "social": { "instagram": "…" }
}
```

---

## Database Impact

### Reads

- `store_settings` — single row `id = f0000000-…`

**Recommended settings extension:**

```json
{
  "about": {
    "hero_title": "…",
    "story_markdown": "…",
    "milestones": [],
    "team": [],
    "stats": {}
  }
}
```

No migration if using existing JSONB columns; optional migration to document schema.

### Writes

| Table | Trigger |
|-------|---------|
| `contact_messages` | Contact form with `source = 'about'` |

---

## Validation

### Contact form (same as homepage)

| Field | Rules |
|-------|-------|
| `name` | Required, 2–255 chars |
| `email` | Required, valid email |
| `phone` | Optional, `^09\d{9}$` |
| `subject` | Optional, max 500 |
| `message` | Required, 10–5000 chars |
| `source` | Must be `"about"` when submitted from this page |
| `website` | Honeypot: must be empty |

### GET /about

- No input parameters.

---

## Permissions

| Action | Role |
|--------|------|
| View about page | Public |
| Submit contact form | Public (rate-limited) |
| Edit about content | Admin (`PUT /api/v1/admin/settings/site` or future context API) |
| Read contact messages | Admin (`/api/v1/admin/contact-messages`) |

---

## State Management

### Page data

| State | Storage |
|-------|---------|
| About content | React Query `['about']`, `staleTime: 300_000` |
| Contact form | React Hook Form local state |
| Submit status | `idle \| submitting \| success \| error` |

### Shared component

```
<ContactForm
  source="about"
  onSuccess={() => toast.success('…')}
/>
```

Reuse validation schema from homepage contact section.

### No user-specific state

About page does not require authentication. No cart/wishlist mutations on this page (unless header components are global).

### Cache invalidation

- Invalidate `['about']` when admin updates settings (not automatic; rely on TTL).

---

## Content Guidelines (for implementers)

Persian copy should emphasize:

- Experience in building materials (کاشی، سیمان، ابزار)
- Delivery coverage (تهران و شهرستان)
- Quality certifications if applicable
- B2B bulk order CTA in contact form placeholder

Example form placeholder: "پیام خود را بنویسید… (استعلام قیمت عمده، زمان تحویل، و غیره)"
