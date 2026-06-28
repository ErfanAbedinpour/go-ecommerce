# Permissions & Roles

## Runtime Authorization (Current)

| Role | Value | Access |
|------|-------|--------|
| Admin | `admin` | All `/api/v1/admin/*` routes |
| Customer | `customer` | `/api/v1/auth/me`, `/api/v1/auth/logout`, future `/api/v1/store/account/*` |
| Public | — | `/api/v1/auth/login`, `/api/v1/auth/signup`, future `/api/v1/store/*` read endpoints |

**Enforcement:** `Authenticate` middleware → `RequireRole(admin)` on admin group.

---

## Access Matrix — Admin Panel

| Feature | Route | Required Role | Notes |
|---------|-------|---------------|-------|
| Dashboard | `/` | admin | |
| Products CRUD | `/products/*` | admin | |
| Product settings | `/products/settings` | admin | |
| Orders | `/orders/*` | admin | |
| Customers | `/users/*` | admin | |
| Coupons | `/coupons` | admin | |
| General settings | `/general-setting` | admin | |
| Navigation | `/navigation` | admin | |
| SEO | `/setting-seo` | admin | |
| Context sections | `/context/*` | admin | New |
| Themes | `/themes`, `/set-style` | admin | New |
| Blog | `/weblog/*` | admin | New |
| Contact inbox | `/contact` | admin | New |
| Admin users | (future UI) | admin | API exists |
| Sign in | `/signin` | public | |
| Sign up | `/signup` | public | Creates customer or admin per config |

---

## Access Matrix — Customer Store

| Feature | Route | Required Role | Notes |
|---------|-------|---------------|-------|
| Homepage | `/` | public | |
| Product catalog | `/products` | public | |
| Product detail | `/products/:id` | public | |
| Checkout | `/checkout` | public* | *Guest checkout allowed (assumption) |
| Account | `/account` | customer | |
| Wishlist | `/account/wishlist` | customer | |
| Blog | `/blog` | public | |
| Submit review | product detail | customer** | **Assumption: verified buyer |
| Submit question | product detail | public | Email required |
| Contact form | homepage/about | public | |
| Add to wishlist | product cards | customer | |

---

## API Permission Mapping (Target)

### Public Store Endpoints (no auth)

```
GET  /api/v1/store/homepage
GET  /api/v1/store/products
GET  /api/v1/store/products/{slug}
GET  /api/v1/store/categories
GET  /api/v1/store/settings
GET  /api/v1/store/theme
GET  /api/v1/store/blog
GET  /api/v1/store/blog/{slug}
POST /api/v1/store/contact
POST /api/v1/store/products/{id}/questions
POST /api/v1/store/coupons/validate
POST /api/v1/store/checkout          # guest allowed
```

### Customer-Authenticated Store Endpoints

```
GET    /api/v1/store/account/profile
PUT    /api/v1/store/account/profile
GET    /api/v1/store/account/orders
GET    /api/v1/store/account/orders/{id}
GET    /api/v1/store/account/wishlist
POST   /api/v1/store/account/wishlist
DELETE /api/v1/store/account/wishlist/{product_id}
POST   /api/v1/store/products/{id}/reviews
```

### Admin Endpoints (all require `admin`)

All existing `/api/v1/admin/*` plus new:

```
# Storefront Content
GET/PUT  /api/v1/admin/storefront/hero
GET/PUT  /api/v1/admin/storefront/product-slides
GET/PUT  /api/v1/admin/storefront/pro-banners
CRUD     /api/v1/admin/storefront/partner-brands
CRUD     /api/v1/admin/storefront/homepage-reviews
GET/PUT  /api/v1/admin/storefront/faq
GET/PUT  /api/v1/admin/storefront/contact-section
GET/PUT  /api/v1/admin/storefront/navigation

# Themes
GET      /api/v1/admin/themes
POST     /api/v1/admin/themes/{id}/purchase
GET/PUT  /api/v1/admin/store-style

# Blog
CRUD     /api/v1/admin/blog/posts
CRUD     /api/v1/admin/blog/categories
GET      /api/v1/admin/blog/comments
PATCH    /api/v1/admin/blog/comments/{id}/approve
PATCH    /api/v1/admin/blog/comments/{id}/reject

# Contact
GET      /api/v1/admin/contact-messages
GET      /api/v1/admin/contact-messages/{id}
PATCH    /api/v1/admin/contact-messages/{id}/read
DELETE   /api/v1/admin/contact-messages/{id}

# Reviews moderation
GET      /api/v1/admin/product-reviews
PATCH    /api/v1/admin/product-reviews/{id}/approve
PATCH    /api/v1/admin/product-reviews/{id}/reject

# Q&A
GET      /api/v1/admin/product-questions
PUT      /api/v1/admin/product-questions/{id}/answer
```

---

## Legacy RBAC Schema (Not Active)

Migration 000001 seeds:

| Role | Permissions |
|------|-------------|
| super_admin | All 21 permissions |
| admin | Most permissions except `users:manage_roles` |
| manager | Read + limited write |
| support | Read-only orders/customers |

**Recommendation:** Defer activation to v2. Document here for future reference.

---

## State Management Requirements

### Admin Panel (Frontend)

| State | Storage | Scope |
|-------|---------|-------|
| Auth tokens | localStorage / httpOnly cookie | Global |
| Current user | React context | Global |
| Sidebar collapse | localStorage | UI preference |
| Product form draft | sessionStorage | Per-session |
| Theme preview | React state | Per-page |
| Table filters/pagination | URL query params | Per-page |

### Customer Store (Frontend)

| State | Storage | Scope |
|-------|---------|-------|
| Auth tokens | localStorage | Global |
| Cart items | localStorage (v1) | Global |
| Wishlist | Server-synced when authenticated | Global |
| Selected variants | React state | Per product page |
| Checkout step | React state / URL | Checkout flow |
| Theme tokens | CSS variables from API | Global |
