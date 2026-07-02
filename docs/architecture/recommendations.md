# Architecture Recommendations

> **Observation** = derived from UI/code analysis.  
> **Recommendation** = proposed improvement with rationale.

---

## 1. Separate Store API from Admin API

**Observation:** Current router mounts everything under `/api/v1/admin/*`. The new storefront requires public read endpoints.

**Recommendation:** Introduce `/api/v1/store/*` with its own handler package and rate limiting.

**Why:** Clear security boundary. Public endpoints can be cached aggressively. Admin mutations stay isolated. Prevents accidental exposure of admin DTOs on public routes.

---

## 2. Unified Storefront Content Aggregate

**Observation:** Admin `/context/*` manages 7+ independent sections (hero, slides, banners, FAQ, reviews, brands, contact image).

**Recommendation:** Model as a `StorefrontContent` aggregate with typed sections stored in JSONB columns on a singleton row, OR separate tables with a `store_content_sections` registry.

**Why:** Homepage API becomes one `GET /store/homepage` call returning all sections. Reduces waterfall requests from the store frontend. Easier cache invalidation.

**Alternative (CQRS):** Write-optimized per-section admin APIs; read-optimized aggregated homepage projection updated on section save.

---

## 3. Partner Brands vs Catalog Brands

**Observation:** Admin has both `products/settings` brands (product attribute) and `context/brands` partner brands (homepage logos with title + description).

**Recommendation:** Keep as **separate entities**. `brands` table = product taxonomy. `partner_brands` table = marketing/logos for homepage.

**Why:** Different lifecycles, fields, and display contexts. Merging would pollute product filters with marketing-only brands.

---

## 4. SKU Variant Matrix

**Observation:** Migration 008-009 added `skus` table with JSONB attributes. UI shows variant pickers (size, color, weight, pattern) on product detail.

**Recommendation:** 
- Store selected attribute combinations as `skus` rows with unique `code` and per-SKU inventory.
- Product detail API returns `attributes[]` (axes) + `skus[]` (valid combinations with price overrides optional).

**Why:** Prevents invalid variant combinations client-side. Enables per-variant stock and SKU-level pricing.

---

## 5. Server-Side Cart with PostgreSQL

**Status:** Implemented — carts and cart items are stored in PostgreSQL (`carts`, `cart_items` tables).

**Design:** Server-side cart keyed by authenticated `user_id` or anonymous `cart_token` cookie.

**Why:** Enables cross-device sync for logged-in customers and consistent checkout without a separate cache store.

---

## 6. Theme System as Configuration, Not Code

**Observation:** Admin `/themes` shows purchasable themes; `/set-style` customizes 12 colors + font.

**Recommendation:** Themes are **configuration records**, not deployed code. Store frontend reads `active_theme_id` + `style_tokens` JSON and applies CSS variables.

**Why:** No redeploy needed to change appearance. Supports theme marketplace without forking frontend repos.

---

## 7. Event-Driven Side Effects

**Observation:** No domain events implemented. Order placement should trigger email, stock decrement, coupon usage increment.

**Recommendation:** Introduce lightweight in-process event bus (Go channels or `watermill` later). Handlers: `OrderPlacedHandler`, `StockAdjustedHandler`.

**Why:** Decouples core order logic from notifications. Enables future async processing without refactoring aggregates.

---

## 8. CQRS for Dashboard Analytics

**Observation:** Dashboard queries aggregate revenue, counts, growth percentages across orders, products, customers.

**Recommendation:** Keep read models in `dashboard` repository (current approach). Consider materialized views or nightly rollup table at scale.

**Why:** Dashboard reads are expensive joins. Pre-computed rollups improve response time under load.

---

## 9. Validation Layers

**Observation:** Validation currently at HTTP DTO layer.

**Recommendation:** Add domain validation methods on new aggregates (`FAQItem.Validate()`, `BlogPost.Publish()`). DTO validation for format; domain validation for business rules.

**Why:** Reusable across admin and store APIs. Tests can target domain without HTTP.

---

## 10. File Upload Strategy

**Observation:** Admin uploads images/videos for hero, banners, products. Store may need review photos.

**Recommendation:** 
- Extend upload handler with `context` param: `product`, `hero`, `banner`, `blog`, `review`.
- Store files under `/uploads/{context}/{uuid}.{ext}`.
- Validate MIME types and max sizes per context (video: 50MB, image: 5MB).

**Why:** Organized storage, context-specific validation, easier cleanup policies.

---

## 11. i18n Architecture

**Observation:** Store UI is Persian (RTL). Admin UI is English (LTR).

**Recommendation:** 
- Store API returns content as stored (Persian).
- Admin content editors accept any language.
- Add optional `locale` column on translatable entities for future multi-language.

**Why:** Avoids premature i18n framework complexity while keeping schema extensible.

---

## 12. Permission Model Evolution

**Observation:** DB has full RBAC schema (roles, permissions) but runtime uses simple `admin`/`customer` check.

**Recommendation:** 
- **v1:** Keep binary admin guard for admin panel.
- **v2:** Activate DB permissions for granular admin roles (manager, support).

**Why:** RBAC schema already exists. Activating it later avoids migration pain.

---

## 13. API Versioning

**Observation:** All routes under `/api/v1`.

**Recommendation:** When storefront launches, keep v1. Breaking changes → v2 with parallel support period.

**Why:** Two frontends on Vercel can migrate independently.

---

## 14. Scalability Considerations

| Area | Recommendation |
|------|----------------|
| Product catalog | Read replicas + CDN for images |
| Homepage content | PostgreSQL reads; optional materialized view or app-level cache at scale |
| Search | PostgreSQL full-text now; Elasticsearch at 10k+ products |
| Checkout | Idempotency key on order creation to prevent duplicates |
| File storage | S3-compatible object storage for production |

---

## 15. DDD Boundary: Blog as Separate Context

**Observation:** Blog has posts, categories, comments — distinct from catalog.

**Recommendation:** `internal/domain/blog/` as isolated bounded context. No direct imports from `product` domain.

**Why:** Blog can evolve independently (tags, authors, scheduling) without coupling to commerce aggregates.
