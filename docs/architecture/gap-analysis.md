# Gap Analysis — Current System vs New UI

## Legend

| Symbol | Meaning |
|--------|---------|
| ✅ | Fully implemented |
| ⚠️ | Partially implemented |
| ❌ | Not implemented |
| 🆕 | New feature in UI |

---

## Admin Panel Gap Matrix

| UI Route | Feature | Backend | Gap Details |
|----------|---------|---------|-------------|
| `/signin`, `/signup` | Authentication | ✅ | Wire frontend |
| `/` | Dashboard | ✅ | All 5 endpoints exist |
| `/products` | Product list | ✅ | Stats, search, filters, delete |
| `/products/create` | Product create/edit | ⚠️ | SKU variant matrix UI may need extended DTO |
| `/products/settings` | Catalog settings | ✅ | Categories, brands, attributes |
| `/orders` | Order list | ⚠️ | Missing `from`/`to` date filter |
| `/orders/:id` | Order detail | ✅ | Status, cancel, refund, notes, invoice |
| `/orders/create` | Manual order | ✅ | `POST /admin/orders` exists |
| `/orders/:id/invoice` | Print invoice | ✅ | `GET /admin/orders/{id}/invoice` |
| `/users` | Customer list | ✅ | |
| `/users/:id` | Customer detail | ✅ | Includes order history |
| `/coupons` | Coupon CRUD | ✅ | |
| `/general-setting` | Site/contact/social | ✅ | |
| `/navigation` | Admin nav menu | ✅ | |
| `/setting-seo` | SEO settings | ✅ | |
| `/context` 🆕 | Content hub | ❌ | New module |
| `/context/hero` 🆕 | Hero video + CTAs | ❌ | New entity + upload |
| `/context/product-slides` 🆕 | 3 product carousels | ❌ | New entity |
| `/context/pro-banners` 🆕 | Promo banners | ❌ | New entity |
| `/context/brands` 🆕 | Partner brands | ❌ | New or extend brands |
| `/context/customer-reviews` 🆕 | Homepage testimonials | ❌ | New entity |
| `/context/faq` 🆕 | FAQ section | ❌ | New entity |
| `/context/contact-us` 🆕 | Contact section image | ❌ | New or settings extension |
| `/context/navigation` 🆕 | Storefront nav | ⚠️ | May reuse navigation API |
| `/themes` 🆕 | Theme marketplace | ❌ | New module |
| `/set-style` 🆕 | Colors + fonts | ❌ | New module |
| `/checkout/themes/*` 🆕 | Theme previews | ❌ | Static previews |
| `/weblog` 🆕 | Blog posts list | ❌ | New module |
| `/weblog/create` 🆕 | Create post | ❌ | |
| `/weblog/settings` 🆕 | Blog categories | ❌ | |
| `/weblog/comments` 🆕 | Comment moderation | ❌ | |
| `/contact` 🆕 | Contact inbox | ❌ | New module |
| `/posts/create` 🆕 | Alias for blog create | ❌ | |

---

## Customer Store Gap Matrix

| UI Route | Feature | Backend | Gap Details |
|----------|---------|---------|-------------|
| `/` | Homepage | ❌ | Needs aggregated storefront content API |
| `/products` | Product catalog | ❌ | Public product list with filters |
| `/products/:id` | Product detail | ❌ | Variants, images, pricing |
| `/checkout` | Checkout flow | ❌ | Cart + order placement |
| `/account` | Customer account | ⚠️ | Auth exists; no account endpoints |
| `/account/wishlist` | Wishlist | ❌ | New entity |
| `/about` | About page | ❌ | Content from settings + contact form |
| `/blog` | Blog listing | ❌ | Public blog API |
| `/blog/:slug` | Blog post | ❌ | |
| Product reviews tab | Reviews | ❌ | New entity |
| Product Q&A tab | Questions | ❌ | New entity |
| Wishlist button | Add to wishlist | ❌ | |
| Contact forms | Submit inquiry | ❌ | Creates ContactMessage |

---

## Module Change Summary

### Existing Modules — Modifications Required

| Module | Changes |
|--------|---------|
| `product` | Expose SKU variant matrix in create/update; public read endpoints |
| `order` | Add date range filter; storefront order creation |
| `customer` | Account profile endpoints; link to wishlist |
| `settings` | Possibly split storefront navigation from admin navigation |
| `dashboard` | No changes |
| `coupon` | Add public validation endpoint for checkout |
| `auth` | Customer signup/login for storefront |

### New Modules Required

| Module | Priority |
|--------|----------|
| `storefront` (catalog, cart, checkout) | Critical |
| `storecontent` (hero, slides, banners, FAQ, reviews) | High |
| `theme` (marketplace, style tokens) | High |
| `blog` (posts, categories, comments) | Medium |
| `contact` (inbound messages) | Medium |
| `wishlist` | Medium |
| `productreview` | Medium |
| `productquestion` | Low |

---

## Database Changes Required

See [database/schema-changes.md](../database/schema-changes.md).

**Summary:** 12+ new tables, 3 new enum types, 5+ new indexes.

---

## Events & Background Jobs (New)

| Event | Trigger | Handler |
|-------|---------|---------|
| `OrderPlaced` | Checkout complete | Send confirmation email, decrement stock |
| `OrderStatusChanged` | Admin status update | Notify customer |
| `ContactMessageReceived` | Form submit | Notify admin |
| `BlogCommentSubmitted` | Comment post | Queue moderation |
| `LowStockThresholdReached` | Stock update | Dashboard alert (exists) |
| `ThemePurchased` | Theme buy | Record purchase |

| Scheduled Job | Frequency | Purpose |
|---------------|-----------|---------|
| `ExpireCoupons` | Daily | Deactivate expired coupons |
| `AbandonedCartReminder` | Daily | Email reminder (future) |
| `GenerateSitemap` | Daily | If `sitemap_enabled` |
