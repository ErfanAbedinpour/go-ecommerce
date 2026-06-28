# E-Commerce Platform — Technical Specification

> **Generated:** 2026-06-28  
> **Admin UI:** [shop-panel-react.vercel.app](https://shop-panel-react.vercel.app/)  
> **Store UI:** [store-os-eta.vercel.app](https://store-os-eta.vercel.app/)  
> **Backend:** Go 1.25 · PostgreSQL · Clean Architecture + DDD

This documentation compares the **new UI** (admin panel + customer storefront) against the **existing Go backend** and provides a complete implementation specification for the development team.

---

## Documentation Map

| Folder | Contents |
|--------|----------|
| [architecture/](./architecture/) | System overview, gap analysis, recommendations |
| [features/](./features/) | One file per feature — UI analysis, flows, APIs |
| [entities/](./entities/) | Domain model — attributes, rules, CRUD, APIs |
| [api/](./api/) | Endpoint catalogs grouped by bounded context |
| [workflows/](./workflows/) | End-to-end user and system workflows |
| [database/](./database/) | Schema changes, migrations, indexes |
| [permissions/](./permissions/) | Roles, guards, access matrix |
| [development-roadmap.md](./development-roadmap.md) | Prioritized milestones and tasks |

---

## Executive Summary

### What Changed

The client delivered **two applications**:

1. **Admin Panel** — expanded from core e-commerce management to include **storefront content management (Context)**, **theme marketplace**, **blog CMS**, and **contact message inbox**.
2. **Customer Store (Store OS)** — a **new Persian (RTL) storefront** for building materials (tiles, cement, tools) with product variants, checkout, wishlist, reviews, Q&A, and blog.

### Backend Coverage Today

| Area | Status |
|------|--------|
| Admin: Dashboard, Products, Orders, Customers, Coupons, Settings, SEO, Navigation | ✅ Mostly implemented |
| Admin: Context (Hero, Slides, Banners, Reviews, FAQ, Partner Brands) | ❌ Not implemented |
| Admin: Themes & Style customization | ❌ Not implemented |
| Admin: Blog CMS & Contact inbox | ❌ Not implemented |
| Storefront public API (catalog, cart, checkout, account) | ❌ Not implemented |
| Product variants (SKU matrix) | ⚠️ Partial — DB schema exists; API needs extension |
| Order date-range filter | ❌ Missing query params |
| Wishlist, Reviews, Q&A | ❌ Not implemented |

### Critical Path

1. **Storefront API module** — enables the customer store to function.
2. **Storefront Content (Context) APIs** — enables admin to configure homepage sections.
3. **Theme & Style system** — enables visual customization.
4. **Blog & Contact** — CMS and inbound message handling.

See [development-roadmap.md](./development-roadmap.md) for full prioritization.

---

## Phase 1 — UI & Feature Analysis (Summary)

Full per-feature analysis is in [features/](./features/). Key findings:

### Admin Panel — New Features

| Feature | Route(s) | Backend Status |
|---------|----------|----------------|
| Storefront Context hub | `/context` | ❌ New module |
| Hero Section | `/context/hero` | ❌ New |
| Product Slides (3 carousels) | `/context/product-slides` | ❌ New |
| Pro Banners | `/context/pro-banners` | ❌ New |
| Partner Brands | `/context/brands` | ❌ New (distinct from product `brands` table) |
| Customer Reviews (homepage) | `/context/customer-reviews` | ❌ New |
| FAQ Section | `/context/faq` | ❌ New |
| Contact Us Section image | `/context/contact-us` | ❌ New |
| Store Navigation (storefront) | `/context/navigation` | ⚠️ Overlaps `/navigation` |
| Theme Marketplace | `/themes` | ❌ New |
| Style Customization | `/set-style` | ❌ New |
| Checkout Theme Previews | `/checkout/themes/*` | ❌ New (preview only) |
| Blog CMS | `/weblog/*`, `/posts/create` | ❌ New |
| Contact Messages | `/contact` | ❌ New |

### Admin Panel — Modified Features

| Feature | Change | Backend Impact |
|---------|--------|----------------|
| Products | Variant/SKU support emphasized in UI | Extend product create/update DTOs |
| Orders list | Date range filter added | Add `from`/`to` query params |
| Orders detail | Payment method, transaction ID, internal notes | Fields exist in DB (migration 007); verify DTO exposure |
| Orders create | Manual order creation page | `POST /admin/orders` exists — verify UI contract |
| Users | Customer edit/delete actions | `PUT`/`DELETE` customers exist — wire frontend |

### Customer Store — All New

| Page | Route | Purpose |
|------|-------|---------|
| Homepage | `/` | Hero, categories, product carousels, brands, FAQ, contact, testimonials, blog teaser |
| Product catalog | `/products` | Search, category filters, sort tabs (bestseller/new/discounted) |
| Product detail | `/products/:id` | Variants, gallery, reviews, Q&A, related products, wishlist |
| Checkout | `/checkout` | Multi-step cart review → shipping → payment |
| Account | `/account` | Profile, orders (assumed) |
| Wishlist | `/account/wishlist` | Saved products |
| About | `/about` | Company story + contact form |
| Blog | `/blog`, `/blog/:slug` | Articles |

---

## Phase 2 — Domain Modeling (Summary)

15 existing entities + 12 new entities identified. Full specs in [entities/](./entities/).

**New entities:** `StorefrontHero`, `ProductSlide`, `ProBanner`, `PartnerBrand`, `HomepageReview`, `FAQSection`, `FAQItem`, `BlogPost`, `BlogCategory`, `BlogComment`, `ContactMessage`, `StoreTheme`, `ThemePurchase`, `StoreStyle`, `WishlistItem`, `ProductReview`, `ProductQuestion`.

---

## Phase 3 — System Changes (Summary)

See [architecture/gap-analysis.md](./architecture/gap-analysis.md) and [database/schema-changes.md](./database/schema-changes.md).

**New modules required:**
- `internal/application/storefront/` — public catalog, cart, checkout
- `internal/application/storecontent/` — hero, slides, banners, FAQ, reviews
- `internal/application/blog/`
- `internal/application/contact/`
- `internal/application/theme/`

---

## Phase 4 — Development Roadmap

See [development-roadmap.md](./development-roadmap.md).

---

## Phase 5 — Architecture Suggestions

See [architecture/recommendations.md](./architecture/recommendations.md).

---

## Open Questions

1. **Partner Brands vs Product Brands** — Are homepage partner logos the same `brands` table or a separate marketing entity?
2. **Theme purchases** — Is billing integrated or mock "purchase" for v1?
3. **Cart persistence** — Server-side cart for guests vs localStorage-only?
4. **Payment gateway** — Which Iranian payment provider (Zarinpal, IDPay, etc.)?
5. **Blog comments** — Moderation workflow: auto-publish vs admin approval?
6. **Product reviews** — Who can post: verified buyers only or any authenticated user?
7. **RTL/i18n** — Is Persian the only storefront locale or multi-language?
8. **Store Navigation vs Admin Navigation** — Single source of truth or separate configs?

---

## Assumptions (Documented)

Where UI alone cannot confirm behavior, these assumptions are used throughout this spec:

- **A1:** Storefront content (Context) is a singleton per store (one hero, one FAQ set, etc.).
- **A2:** Product slides reference existing `products` by ID; admin picks products for each carousel.
- **A3:** Pro banners support desktop + mobile images and a click-through URL.
- **A4:** Theme system stores 12 customizable colors + 1 font selection per active theme.
- **A5:** Checkout is 3 steps: Cart Review → Shipping/Address → Payment.
- **A6:** Wishlist requires authenticated customer; guests see login prompt.
- **A7:** Product Q&A allows anonymous questions with email; answers by admin.
- **A8:** Contact form on homepage/about creates `ContactMessage` records visible in admin `/contact`.
