# eCommerce Admin Backend — Implementation Plan

> **Target UI:** [shop-panel-react.vercel.app](https://shop-panel-react.vercel.app/)  
> **Stack:** Go 1.25 · GORM · PostgreSQL · REST · Docker · OpenAPI/Swagger  
> **Architecture:** Clean Architecture + DDD + Repository + Service Layer

---

## Table of Contents

1. [Admin Panel Analysis](#1-admin-panel-analysis)
2. [Architectural Decisions](#2-architectural-decisions)
3. [Bounded Contexts & DDD Model](#3-bounded-contexts--ddd-model)
4. [Database Schema](#4-database-schema)
5. [API Design](#5-api-design)
6. [Project Structure](#6-project-structure)
7. [Implementation Phases](#7-implementation-phases)
8. [Observability & Production Concerns](#8-observability--production-concerns)

---

## 1. Admin Panel Analysis

### 1.1 UI Routes Discovered

| Route                 | Feature                                                   |
| --------------------- | --------------------------------------------------------- |
| `/`                   | Dashboard — KPIs, revenue chart, recent orders, low stock |
| `/signin`, `/signup`  | Authentication                                            |
| `/products`           | Product list with search & filters                        |
| `/products/create`    | Create product                                            |
| `/products/settings`  | Product attributes (color, size, weight, brand)           |
| `/orders`             | Order list                                                |
| `/orders/:id`         | Order detail, timeline, customer info                     |
| `/orders/:id/invoice` | Printable invoice                                         |
| `/orders/create`      | Manual order creation                                     |
| `/users`              | Customer & admin user list                                |
| `/users/:id`          | User profile, addresses, order history                    |
| `/coupons`            | Coupon management                                         |
| `/general-setting`    | Store settings                                            |
| `/navigation`         | Menu management                                           |
| `/weblog/*`           | Blog CMS (out of scope for v1 backend)                    |
| `/contact`            | Contact messages (out of scope for v1 backend)            |
| `/setting-seo`        | SEO settings (out of scope for v1 backend)                |

### 1.2 Business Capabilities Required

#### Dashboard

- Total Revenue, Total Orders, Total Customers, Total Products
- Pending Orders count
- Low Stock Products list
- Revenue analytics (daily/weekly/monthly time series)
- Recent orders feed
- Sales & revenue chart data

#### Products

- CRUD with pagination, search, filtering (category, brand, status, stock level)
- Fields: name, slug, SKU, description, short description, price, sale price
- Stock quantity with low-stock / out-of-stock thresholds
- Multiple images (ordered gallery)
- Category assignment
- Brand assignment
- Product attributes (color, size, weight, material — configurable)
- Featured product flag
- Status: draft, active, archived

#### Categories

- Hierarchical categories (parent/child)
- CRUD with slug, name, description, image
- Active/inactive status

#### Orders

- List with search (order ID, customer), status & payment filters
- Order detail: items, customer, billing/shipping addresses, payment info
- Status workflow: `pending → processing → shipped → delivered`
- Terminal states: `cancelled`, `refunded`
- Payment status: `unpaid`, `paid`, `refunded`
- Order timeline (status change history)
- Cancel order (pre-shipment)
- Refund order (post-payment)
- Invoice generation data

#### Coupons

- CRUD with code, discount type (percentage | fixed_amount)
- Discount value, min order amount, max usage count, expiry date
- Active/inactive toggle
- Usage tracking

#### Customers

- List with search, pagination
- Profile: name, email, phone, addresses
- Purchase history (linked orders)
- Customer type: registered | guest

#### Admin Users

- CRUD for admin accounts
- Single role per user: `admin` or `customer`
- Route-level authorization guards (not DB permission lookups)
- Separate from storefront customers

#### Audit Logs

- Immutable log of admin actions
- Actor, action, resource type/id, before/after snapshot, IP, timestamp

---

## 2. Architectural Decisions

### 2.1 Clean Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│  interfaces/          HTTP handlers, middleware, DTOs   │
├─────────────────────────────────────────────────────────┤
│  application/         Use cases, application services   │
├─────────────────────────────────────────────────────────┤
│  domain/              Entities, value objects, ports   │
├─────────────────────────────────────────────────────────┤
│  infrastructure/      GORM repos, JWT, logging, metrics │
└─────────────────────────────────────────────────────────┘
```

**Justification:** Dependency rule flows inward. Domain has zero infrastructure imports. Use cases depend on repository interfaces (ports) defined in domain. Infrastructure implements those ports. This enables unit testing use cases with mocks and swapping PostgreSQL for another store without touching business logic.

### 2.2 DDD Tactical Patterns

| Pattern            | Application                                                        |
| ------------------ | ------------------------------------------------------------------ |
| **Aggregate Root** | `Product`, `Order`, `Coupon`, `AdminUser` — consistency boundaries |
| **Entity**         | `OrderItem`, `ProductImage`, `Role`, `Permission`                  |
| **Value Object**   | `Money`, `Email`, `Slug`, `Address`, `Discount`                    |
| **Domain Event**   | `OrderStatusChanged`, `StockAdjusted`, `CouponRedeemed`            |
| **Repository**     | One interface per aggregate root                                   |
| **Domain Service** | `PricingService`, `InventoryService`, `CouponValidator`            |

**Justification:** Aggregates enforce invariants (e.g., order total = sum of items; stock cannot go negative). Value objects encapsulate validation (email format, money arithmetic). Domain events decouple side effects (audit logging, analytics updates).

### 2.3 Dependency Injection

Manual constructor injection via a `Container` struct in `internal/di`. No reflection-based frameworks (wire/fx avoided for explicitness and debuggability).

**Justification:** Explicit wiring is easier to trace in production incidents. Each component's dependencies are visible at compile time.

### 2.4 HTTP Framework: `chi` + `net/http`

**Justification:** `chi` is lightweight, idiomatic, supports middleware composition, and stays close to stdlib. No heavy framework magic. Compatible with `httprouter`-style routing while remaining `net/http` compatible.

### 2.5 ORM: GORM

**Justification:** Mature PostgreSQL support, migrations via golang-migrate, preloading for relations. Repository layer wraps GORM to keep domain models free of GORM tags (mapping in infrastructure).

### 2.6 Authentication: JWT (access + refresh tokens)

- Access token: 15 min TTL, HS256 signed
- Refresh token: 7 day TTL, stored hashed in DB, rotatable
- Logout invalidates refresh token family

**Justification:** Stateless access tokens scale horizontally. Refresh token rotation prevents replay attacks. DB-backed refresh tokens enable revocation.

### 2.7 RBAC Model (Application-Layer Route Guards)

```
User ──has──> Role (admin | customer)
                │
                ▼
         Router Middleware Guards
```

Authorization is enforced at the **router/application layer**, not the database layer:

| Guard             | Applied To                                               | Roles Allowed       |
| ----------------- | -------------------------------------------------------- | ------------------- |
| **Public**        | `/healthz`, `/api/v1/auth/login`, `/api/v1/auth/refresh` | None (no token)     |
| **Authenticated** | `/api/v1/auth/logout`, `/api/v1/auth/me`                 | `admin`, `customer` |
| **Admin**         | `/api/v1/admin/*`                                        | `admin` only        |

- Each user has exactly **one role**: `admin` or `customer` (stored as a column on `admin_users`).
- Role is embedded in the JWT at login and validated by middleware — no DB permission lookups per request.
- Route guards: `Authenticate` → `RequireRole(user.RoleAdmin)`.

**Justification:** Simple two-role model matches the admin panel vs storefront split. Router-level guards are explicit, testable, and follow the principle that authorization policy lives in the application layer, not in SQL joins.

### 2.8 Error Handling

Custom `AppError` type with HTTP status, error code, message, and optional field-level details. Never leak internal errors to clients.

```go
type AppError struct {
    Code       string            `json:"code"`
    Message    string            `json:"message"`
    Status     int               `json:"-"`
    Details    map[string]string `json:"details,omitempty"`
}
```

### 2.9 Configuration: Environment variables + `.env` file

Using `caarlos0/env` for struct-based parsing. Twelve-factor compliant.

### 2.10 Logging: `slog` (stdlib structured logging)

JSON output in production, text in development. Request ID propagation via middleware.

### 2.11 API Versioning

All admin routes under `/api/v1/admin/...`

### 2.12 Pagination Standard

```json
{
  "data": [],
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

Query params: `page`, `per_page` (max 100), `sort`, `order` (asc|desc).

---

## 3. Bounded Contexts & DDD Model

### 3.1 Bounded Contexts

```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Identity &  │  │   Catalog    │  │   Ordering   │
│  Access Mgmt │  │   Context    │  │   Context    │
└──────────────┘  └──────────────┘  └──────────────┘
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│  Promotion   │  │  Analytics   │  │    Audit     │
│   Context    │  │   Context    │  │   Context    │
└──────────────┘  └──────────────┘  └──────────────┘
```

### 3.2 Domain Model

#### Identity & Access Context

| Type         | Name           | Description                    |
| ------------ | -------------- | ------------------------------ |
| Aggregate    | `AdminUser`    | Admin account with credentials |
| Entity       | `Role`         | Named role bundle              |
| Entity       | `Permission`   | Atomic capability              |
| Entity       | `RefreshToken` | Token rotation record          |
| Value Object | `Email`        | Validated email                |
| Value Object | `PasswordHash` | bcrypt hash                    |
| Value Object | `HashedToken`  | SHA-256 of refresh token       |

**Use Cases:**

- `LoginAdmin`, `RefreshToken`, `LogoutAdmin`, `GetCurrentAdmin`
- `CreateAdminUser`, `UpdateAdminUser`, `DeleteAdminUser`
- `AssignRoles`, `ManagePermissions`

**Application Services:** `AuthService`, `AdminUserService`

**Infrastructure:** `JWTService`, `BcryptHasher`, `GormAdminUserRepository`

#### Catalog Context

| Type         | Name               | Description                    |
| ------------ | ------------------ | ------------------------------ |
| Aggregate    | `Product`          | Product with images, inventory |
| Aggregate    | `Category`         | Hierarchical category          |
| Entity       | `ProductImage`     | Image URL with sort order      |
| Entity       | `Inventory`        | Stock level per product        |
| Entity       | `ProductAttribute` | Key-value attribute            |
| Value Object | `Money`            | Amount + currency              |
| Value Object | `Slug`             | URL-safe identifier            |
| Value Object | `SKU`              | Stock keeping unit             |

**Use Cases:**

- `CreateProduct`, `UpdateProduct`, `DeleteProduct`, `GetProduct`, `ListProducts`, `SearchProducts`
- `AdjustInventory`, `GetLowStockProducts`
- `CreateCategory`, `UpdateCategory`, `DeleteCategory`, `ListCategories`

**Application Services:** `ProductService`, `CategoryService`, `InventoryService`

#### Ordering Context

| Type         | Name                 | Description                    |
| ------------ | -------------------- | ------------------------------ |
| Aggregate    | `Order`              | Order with items and status    |
| Entity       | `OrderItem`          | Line item snapshot             |
| Entity       | `OrderStatusHistory` | Timeline entry                 |
| Entity       | `Customer`           | Buyer profile                  |
| Value Object | `Address`            | Shipping/billing address       |
| Value Object | `OrderNumber`        | Human-readable ID              |
| Domain Event | `OrderStatusChanged` | Triggers audit + notifications |

**Use Cases:**

- `ListOrders`, `GetOrder`, `UpdateOrderStatus`, `CancelOrder`, `RefundOrder`
- `ListCustomers`, `GetCustomer`, `GetCustomerPurchaseHistory`

**Application Services:** `OrderService`, `CustomerService`

#### Promotion Context

| Type         | Name         | Description            |
| ------------ | ------------ | ---------------------- |
| Aggregate    | `Coupon`     | Discount coupon        |
| Value Object | `Discount`   | Type + value           |
| Value Object | `CouponCode` | Uppercase alphanumeric |

**Use Cases:**

- `CreateCoupon`, `UpdateCoupon`, `DeleteCoupon`, `GetCoupon`, `ListCoupons`
- `ActivateCoupon`, `DeactivateCoupon`

**Application Services:** `CouponService`

#### Analytics Context

| Type       | Name               | Description       |
| ---------- | ------------------ | ----------------- |
| Read Model | `DashboardStats`   | Aggregated KPIs   |
| Read Model | `RevenueDataPoint` | Time-series point |

**Use Cases:**

- `GetDashboardStats`, `GetRevenueAnalytics`

**Application Services:** `DashboardService` (read-only, queries via raw SQL/GORM)

#### Audit Context

| Type   | Name       | Description             |
| ------ | ---------- | ----------------------- |
| Entity | `AuditLog` | Immutable action record |

**Use Cases:**

- `ListAuditLogs`, `GetAuditLog`

**Application Services:** `AuditService` (append-only writer + reader)

---

## 4. Database Schema

### 4.1 ER Overview

```
admin_users ──M:N── admin_user_roles ──M:N── roles ──M:N── role_permissions ──M:N── permissions
admin_users ──1:N── refresh_tokens
admin_users ──1:N── audit_logs

categories (self-referential parent_id)
products ──N:1── categories
products ──1:N── product_images
products ──1:1── inventories
products ──1:N── product_attributes

customers ──1:N── customer_addresses
customers ──1:N── orders
orders ──1:N── order_items
orders ──1:N── order_status_history
orders ──N:1── coupons (nullable)

coupons (standalone)
```

### 4.2 Table Definitions

#### `admin_users`

| Column        | Type         | Constraints                   |
| ------------- | ------------ | ----------------------------- |
| id            | UUID         | PK, DEFAULT gen_random_uuid() |
| email         | VARCHAR(255) | UNIQUE, NOT NULL              |
| password_hash | VARCHAR(255) | NOT NULL                      |
| first_name    | VARCHAR(100) | NOT NULL                      |
| last_name     | VARCHAR(100) | NOT NULL                      |
| phone         | VARCHAR(20)  | NULLABLE                      |
| is_active     | BOOLEAN      | DEFAULT true                  |
| last_login_at | TIMESTAMPTZ  | NULLABLE                      |
| created_at    | TIMESTAMPTZ  | NOT NULL                      |
| updated_at    | TIMESTAMPTZ  | NOT NULL                      |
| deleted_at    | TIMESTAMPTZ  | NULLABLE (soft delete)        |

#### `roles`

| Column      | Type        | Constraints                                   |
| ----------- | ----------- | --------------------------------------------- |
| id          | UUID        | PK                                            |
| name        | VARCHAR(50) | UNIQUE (super_admin, admin, manager, support) |
| description | TEXT        |                                               |
| created_at  | TIMESTAMPTZ | NOT NULL                                      |
| updated_at  | TIMESTAMPTZ | NOT NULL                                      |

#### `permissions`

| Column      | Type         | Constraints                 |
| ----------- | ------------ | --------------------------- |
| id          | UUID         | PK                          |
| name        | VARCHAR(100) | UNIQUE (e.g. products:read) |
| description | TEXT         |                             |
| created_at  | TIMESTAMPTZ  | NOT NULL                    |

#### `admin_user_roles` (join)

| Column                               | Type | Constraints      |
| ------------------------------------ | ---- | ---------------- |
| admin_user_id                        | UUID | FK → admin_users |
| role_id                              | UUID | FK → roles       |
| PRIMARY KEY (admin_user_id, role_id) |      |                  |

#### `role_permissions` (join)

| Column                               | Type | Constraints      |
| ------------------------------------ | ---- | ---------------- |
| role_id                              | UUID | FK → roles       |
| permission_id                        | UUID | FK → permissions |
| PRIMARY KEY (role_id, permission_id) |      |                  |

#### `refresh_tokens`

| Column        | Type         | Constraints                |
| ------------- | ------------ | -------------------------- |
| id            | UUID         | PK                         |
| admin_user_id | UUID         | FK → admin_users           |
| token_hash    | VARCHAR(255) | NOT NULL                   |
| family_id     | UUID         | NOT NULL (rotation family) |
| expires_at    | TIMESTAMPTZ  | NOT NULL                   |
| revoked_at    | TIMESTAMPTZ  | NULLABLE                   |
| created_at    | TIMESTAMPTZ  | NOT NULL                   |

#### `categories`

| Column      | Type         | Constraints               |
| ----------- | ------------ | ------------------------- |
| id          | UUID         | PK                        |
| parent_id   | UUID         | FK → categories, NULLABLE |
| name        | VARCHAR(200) | NOT NULL                  |
| slug        | VARCHAR(200) | UNIQUE, NOT NULL          |
| description | TEXT         | NULLABLE                  |
| image_url   | VARCHAR(500) | NULLABLE                  |
| sort_order  | INT          | DEFAULT 0                 |
| is_active   | BOOLEAN      | DEFAULT true              |
| created_at  | TIMESTAMPTZ  | NOT NULL                  |
| updated_at  | TIMESTAMPTZ  | NOT NULL                  |
| deleted_at  | TIMESTAMPTZ  | NULLABLE                  |

#### `products`

| Column            | Type          | Constraints                        |
| ----------------- | ------------- | ---------------------------------- |
| id                | UUID          | PK                                 |
| category_id       | UUID          | FK → categories, NULLABLE          |
| name              | VARCHAR(300)  | NOT NULL                           |
| slug              | VARCHAR(300)  | UNIQUE, NOT NULL                   |
| sku               | VARCHAR(100)  | UNIQUE, NOT NULL                   |
| description       | TEXT          | NULLABLE                           |
| short_description | VARCHAR(500)  | NULLABLE                           |
| price             | DECIMAL(12,2) | NOT NULL, CHECK >= 0               |
| sale_price        | DECIMAL(12,2) | NULLABLE, CHECK >= 0               |
| brand             | VARCHAR(100)  | NULLABLE                           |
| is_featured       | BOOLEAN       | DEFAULT false                      |
| status            | VARCHAR(20)   | NOT NULL (draft, active, archived) |
| created_at        | TIMESTAMPTZ   | NOT NULL                           |
| updated_at        | TIMESTAMPTZ   | NOT NULL                           |
| deleted_at        | TIMESTAMPTZ   | NULLABLE                           |

#### `product_images`

| Column     | Type         | Constraints                      |
| ---------- | ------------ | -------------------------------- |
| id         | UUID         | PK                               |
| product_id | UUID         | FK → products, ON DELETE CASCADE |
| url        | VARCHAR(500) | NOT NULL                         |
| alt_text   | VARCHAR(200) | NULLABLE                         |
| sort_order | INT          | DEFAULT 0                        |
| created_at | TIMESTAMPTZ  | NOT NULL                         |

#### `inventories`

| Column              | Type        | Constraints                     |
| ------------------- | ----------- | ------------------------------- |
| id                  | UUID        | PK                              |
| product_id          | UUID        | FK → products, UNIQUE           |
| quantity            | INT         | NOT NULL, DEFAULT 0, CHECK >= 0 |
| low_stock_threshold | INT         | DEFAULT 10                      |
| updated_at          | TIMESTAMPTZ | NOT NULL                        |

#### `product_attributes`

| Column     | Type         | Constraints                      |
| ---------- | ------------ | -------------------------------- |
| id         | UUID         | PK                               |
| product_id | UUID         | FK → products, ON DELETE CASCADE |
| name       | VARCHAR(100) | NOT NULL                         |
| value      | VARCHAR(200) | NOT NULL                         |

#### `customers`

| Column       | Type          | Constraints          |
| ------------ | ------------- | -------------------- |
| id           | UUID          | PK                   |
| email        | VARCHAR(255)  | NOT NULL             |
| first_name   | VARCHAR(100)  | NOT NULL             |
| last_name    | VARCHAR(100)  | NOT NULL             |
| phone        | VARCHAR(20)   | NULLABLE             |
| type         | VARCHAR(20)   | DEFAULT 'registered' |
| total_orders | INT           | DEFAULT 0            |
| total_spent  | DECIMAL(12,2) | DEFAULT 0            |
| created_at   | TIMESTAMPTZ   | NOT NULL             |
| updated_at   | TIMESTAMPTZ   | NOT NULL             |

#### `customer_addresses`

| Column      | Type         | Constraints                     |
| ----------- | ------------ | ------------------------------- |
| id          | UUID         | PK                              |
| customer_id | UUID         | FK → customers                  |
| type        | VARCHAR(20)  | (home, work, billing, shipping) |
| street      | VARCHAR(300) | NOT NULL                        |
| city        | VARCHAR(100) | NOT NULL                        |
| state       | VARCHAR(100) | NULLABLE                        |
| postal_code | VARCHAR(20)  | NOT NULL                        |
| country     | VARCHAR(2)   | NOT NULL (ISO 3166-1 alpha-2)   |
| is_default  | BOOLEAN      | DEFAULT false                   |

#### `orders`

| Column           | Type          | Constraints            |
| ---------------- | ------------- | ---------------------- |
| id               | UUID          | PK                     |
| order_number     | VARCHAR(20)   | UNIQUE, NOT NULL       |
| customer_id      | UUID          | FK → customers         |
| coupon_id        | UUID          | FK → coupons, NULLABLE |
| status           | VARCHAR(20)   | NOT NULL               |
| payment_status   | VARCHAR(20)   | NOT NULL               |
| subtotal         | DECIMAL(12,2) | NOT NULL               |
| discount_amount  | DECIMAL(12,2) | DEFAULT 0              |
| shipping_amount  | DECIMAL(12,2) | DEFAULT 0              |
| tax_amount       | DECIMAL(12,2) | DEFAULT 0              |
| total            | DECIMAL(12,2) | NOT NULL               |
| notes            | TEXT          | NULLABLE               |
| billing_address  | JSONB         | NOT NULL               |
| shipping_address | JSONB         | NOT NULL               |
| created_at       | TIMESTAMPTZ   | NOT NULL               |
| updated_at       | TIMESTAMPTZ   | NOT NULL               |

#### `order_items`

| Column       | Type          | Constraints                    |
| ------------ | ------------- | ------------------------------ |
| id           | UUID          | PK                             |
| order_id     | UUID          | FK → orders, ON DELETE CASCADE |
| product_id   | UUID          | FK → products                  |
| product_name | VARCHAR(300)  | NOT NULL (snapshot)            |
| product_sku  | VARCHAR(100)  | NOT NULL (snapshot)            |
| quantity     | INT           | NOT NULL, CHECK > 0            |
| unit_price   | DECIMAL(12,2) | NOT NULL                       |
| total_price  | DECIMAL(12,2) | NOT NULL                       |

#### `order_status_history`

| Column      | Type        | Constraints                |
| ----------- | ----------- | -------------------------- |
| id          | UUID        | PK                         |
| order_id    | UUID        | FK → orders                |
| from_status | VARCHAR(20) | NULLABLE                   |
| to_status   | VARCHAR(20) | NOT NULL                   |
| note        | TEXT        | NULLABLE                   |
| changed_by  | UUID        | FK → admin_users, NULLABLE |
| created_at  | TIMESTAMPTZ | NOT NULL                   |

#### `coupons`

| Column           | Type          | Constraints                 |
| ---------------- | ------------- | --------------------------- |
| id               | UUID          | PK                          |
| code             | VARCHAR(50)   | UNIQUE, NOT NULL            |
| discount_type    | VARCHAR(20)   | (percentage, fixed_amount)  |
| discount_value   | DECIMAL(12,2) | NOT NULL                    |
| min_order_amount | DECIMAL(12,2) | DEFAULT 0                   |
| max_usage        | INT           | NULLABLE (NULL = unlimited) |
| usage_count      | INT           | DEFAULT 0                   |
| expires_at       | TIMESTAMPTZ   | NULLABLE                    |
| is_active        | BOOLEAN       | DEFAULT true                |
| note             | TEXT          | NULLABLE                    |
| created_at       | TIMESTAMPTZ   | NOT NULL                    |
| updated_at       | TIMESTAMPTZ   | NOT NULL                    |
| deleted_at       | TIMESTAMPTZ   | NULLABLE                    |

#### `audit_logs`

| Column        | Type         | Constraints      |
| ------------- | ------------ | ---------------- |
| id            | UUID         | PK               |
| admin_user_id | UUID         | FK → admin_users |
| action        | VARCHAR(50)  | NOT NULL         |
| resource_type | VARCHAR(50)  | NOT NULL         |
| resource_id   | VARCHAR(100) | NOT NULL         |
| old_value     | JSONB        | NULLABLE         |
| new_value     | JSONB        | NULLABLE         |
| ip_address    | VARCHAR(45)  | NULLABLE         |
| user_agent    | TEXT         | NULLABLE         |
| created_at    | TIMESTAMPTZ  | NOT NULL         |

### 4.3 Indexes

```sql
CREATE INDEX idx_products_category ON products(category_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_slug ON products(slug);
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created ON orders(created_at DESC);
CREATE INDEX idx_audit_logs_admin ON audit_logs(admin_user_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
CREATE INDEX idx_coupons_code ON coupons(code);
CREATE INDEX idx_refresh_tokens_hash ON refresh_tokens(token_hash);
```

---

## 5. API Design

### Route Guard Tiers

| Tier          | Base Path                                    | Guard Middleware                      | Description                    |
| ------------- | -------------------------------------------- | ------------------------------------- | ------------------------------ |
| Public        | `/api/v1/auth/login`, `/api/v1/auth/refresh` | None                                  | No token required              |
| Authenticated | `/api/v1/auth/logout`, `/api/v1/auth/me`     | `Authenticate`                        | Any role (`admin`, `customer`) |
| Admin         | `/api/v1/admin/*`                            | `Authenticate` + `RequireRole(admin)` | Admin panel APIs               |

### 5.1 Authentication

#### POST `/auth/login`

|                    |                                                                             |
| ------------------ | --------------------------------------------------------------------------- |
| **Auth**           | Public                                                                      |
| **Request**        | `{ "email": "admin@shop.com", "password": "string" }`                       |
| **Validation**     | email: required, valid format; password: required, min 8 chars              |
| **Response 200**   | `{ "access_token", "refresh_token", "expires_in", "token_type": "Bearer" }` |
| **Business Rules** | Account must be active; bcrypt password verify; update last_login_at        |
| **Errors**         | 401 INVALID_CREDENTIALS, 403 ACCOUNT_DISABLED                               |

#### POST `/auth/refresh`

|                    |                                                            |
| ------------------ | ---------------------------------------------------------- |
| **Auth**           | Public                                                     |
| **Request**        | `{ "refresh_token": "string" }`                            |
| **Validation**     | refresh_token: required                                    |
| **Response 200**   | New token pair                                             |
| **Business Rules** | Rotate refresh token; revoke old; check family not revoked |
| **Errors**         | 401 INVALID_REFRESH_TOKEN, 401 TOKEN_REVOKED               |

#### POST `/auth/logout`

|                    |                                    |
| ------------------ | ---------------------------------- |
| **Auth**           | Bearer token                       |
| **Request**        | `{ "refresh_token": "string" }`    |
| **Response 204**   | No content                         |
| **Business Rules** | Revoke entire refresh token family |

#### GET `/auth/me`

|                  |                                                                 |
| ---------------- | --------------------------------------------------------------- |
| **Auth**         | Bearer token                                                    |
| **Response 200** | `{ "id", "email", "first_name", "last_name", "role": "admin" }` |

---

### 5.2 Dashboard

#### GET `/dashboard/stats`

|                  |                                                                                                                 |
| ---------------- | --------------------------------------------------------------------------------------------------------------- |
| **Auth**         | Admin role                                                                                                      |
| **Response 200** | `{ "total_revenue", "total_orders", "total_customers", "total_products", "pending_orders", "low_stock_count" }` |

#### GET `/dashboard/revenue`

|                  |                                                   |
| ---------------- | ------------------------------------------------- |
| **Auth**         | Admin role                                        |
| **Query**        | `period` (today, week, month, year), `from`, `to` |
| **Response 200** | `{ "data": [{ "date", "revenue", "orders" }] }`   |

#### GET `/dashboard/low-stock`

|                  |                                                    |
| ---------------- | -------------------------------------------------- |
| **Auth**         | `products:read`                                    |
| **Query**        | `page`, `per_page`                                 |
| **Response 200** | Paginated product list where quantity <= threshold |

#### GET `/dashboard/recent-orders`

|                  |                              |
| ---------------- | ---------------------------- |
| **Auth**         | `orders:read`                |
| **Query**        | `limit` (default 10, max 50) |
| **Response 200** | `{ "data": [OrderSummary] }` |

---

### 5.3 Products

#### POST `/products`

|                    |                                                                                                                                      |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------ |
| **Auth**           | `products:write`                                                                                                                     |
| **Request**        | See ProductCreateRequest DTO below                                                                                                   |
| **Validation**     | name: required, 1-300; slug: required, unique; sku: required, unique; price: required, >= 0; category_id: valid UUID; images: max 10 |
| **Response 201**   | ProductResponse                                                                                                                      |
| **Business Rules** | Auto-create inventory record; generate slug if omitted; audit log                                                                    |

#### PUT `/products/{id}`

|                  |                                |
| ---------------- | ------------------------------ |
| **Auth**         | `products:write`               |
| **Request**      | ProductUpdateRequest (partial) |
| **Response 200** | ProductResponse                |
| **Errors**       | 404 PRODUCT_NOT_FOUND          |

#### DELETE `/products/{id}`

|                    |                                                  |
| ------------------ | ------------------------------------------------ |
| **Auth**           | `products:delete`                                |
| **Response 204**   | Soft delete                                      |
| **Business Rules** | Cannot delete if active orders reference product |

#### GET `/products/{id}`

|                  |                                                    |
| ---------------- | -------------------------------------------------- |
| **Auth**         | `products:read`                                    |
| **Response 200** | ProductResponse with images, inventory, attributes |

#### GET `/products`

|                  |                                                                                      |
| ---------------- | ------------------------------------------------------------------------------------ |
| **Auth**         | `products:read`                                                                      |
| **Query**        | `page`, `per_page`, `sort`, `order`, `status`, `category_id`, `brand`, `is_featured` |
| **Response 200** | Paginated ProductResponse list                                                       |

#### GET `/products/search`

|                  |                                                   |
| ---------------- | ------------------------------------------------- |
| **Auth**         | `products:read`                                   |
| **Query**        | `q` (required, min 2 chars), `page`, `per_page`   |
| **Response 200** | Paginated search results (name, sku, description) |

#### PATCH `/products/{id}/inventory`

|                    |                                                                                  |
| ------------------ | -------------------------------------------------------------------------------- |
| **Auth**           | `inventory:write`                                                                |
| **Request**        | `{ "quantity": 100, "low_stock_threshold": 10, "adjustment_reason": "restock" }` |
| **Validation**     | quantity: required, >= 0                                                         |
| **Response 200**   | InventoryResponse                                                                |
| **Business Rules** | Log adjustment; emit low-stock event if below threshold                          |

**ProductCreateRequest:**

```json
{
  "name": "Nike Air Max",
  "slug": "nike-air-max-black",
  "sku": "PROD-001",
  "description": "Full description",
  "short_description": "Short desc",
  "price": 129.99,
  "sale_price": 99.99,
  "category_id": "uuid",
  "brand": "Nike",
  "is_featured": false,
  "status": "draft",
  "images": [{ "url": "https://...", "alt_text": "...", "sort_order": 0 }],
  "attributes": [{ "name": "Color", "value": "Black" }],
  "inventory": { "quantity": 50, "low_stock_threshold": 10 }
}
```

---

### 5.4 Categories

#### POST `/categories`

|                  |                                                                                          |
| ---------------- | ---------------------------------------------------------------------------------------- |
| **Auth**         | `categories:write`                                                                       |
| **Request**      | `{ "name", "slug", "description", "parent_id", "image_url", "sort_order", "is_active" }` |
| **Validation**   | name: required; slug: unique                                                             |
| **Response 201** | CategoryResponse                                                                         |

#### PUT `/categories/{id}`

|                    |                                         |
| ------------------ | --------------------------------------- |
| **Auth**           | `categories:write`                      |
| **Business Rules** | Cannot set parent to self or descendant |

#### DELETE `/categories/{id}`

|                    |                                                      |
| ------------------ | ---------------------------------------------------- |
| **Auth**           | `categories:delete`                                  |
| **Business Rules** | Cannot delete if products assigned or children exist |

#### GET `/categories/{id}`

|          |                   |
| -------- | ----------------- |
| **Auth** | `categories:read` |

#### GET `/categories`

|           |                                                                             |
| --------- | --------------------------------------------------------------------------- |
| **Auth**  | `categories:read`                                                           |
| **Query** | `page`, `per_page`, `parent_id`, `is_active`, `tree` (bool — return nested) |

---

### 5.5 Orders

#### GET `/orders`

|                  |                                                                                |
| ---------------- | ------------------------------------------------------------------------------ |
| **Auth**         | `orders:read`                                                                  |
| **Query**        | `page`, `per_page`, `status`, `payment_status`, `q` (order number or customer) |
| **Response 200** | Paginated OrderSummary list                                                    |

#### GET `/orders/{id}`

|                  |                                                               |
| ---------------- | ------------------------------------------------------------- |
| **Auth**         | `orders:read`                                                 |
| **Response 200** | OrderDetailResponse with items, customer, timeline, addresses |

#### PATCH `/orders/{id}/status`

|                    |                                                                                                      |
| ------------------ | ---------------------------------------------------------------------------------------------------- |
| **Auth**           | `orders:write`                                                                                       |
| **Request**        | `{ "status": "shipped", "note": "Shipped via FedEx" }`                                               |
| **Validation**     | Valid status transition                                                                              |
| **Business Rules** | State machine: pending→processing→shipped→delivered; any→cancelled (pre-shipped); delivered→refunded |
| **Errors**         | 422 INVALID_STATUS_TRANSITION                                                                        |

#### POST `/orders/{id}/cancel`

|                    |                                            |
| ------------------ | ------------------------------------------ |
| **Auth**           | `orders:cancel`                            |
| **Business Rules** | Only pending/processing; restore inventory |

#### POST `/orders/{id}/refund`

|                    |                                                     |
| ------------------ | --------------------------------------------------- |
| **Auth**           | `orders:refund`                                     |
| **Request**        | `{ "amount": 99.99, "reason": "Customer request" }` |
| **Business Rules** | Only paid/delivered orders; partial or full refund  |

---

### 5.6 Coupons

#### POST `/coupons`

|                  |                                                                                                        |
| ---------------- | ------------------------------------------------------------------------------------------------------ |
| **Auth**         | `coupons:write`                                                                                        |
| **Request**      | `{ "code", "discount_type", "discount_value", "min_order_amount", "max_usage", "expires_at", "note" }` |
| **Validation**   | code: required, 3-50, alphanumeric; discount_value: > 0; percentage: <= 100                            |
| **Response 201** | CouponResponse                                                                                         |

#### PUT `/coupons/{id}`

|          |                 |
| -------- | --------------- |
| **Auth** | `coupons:write` |

#### DELETE `/coupons/{id}`

|                  |                  |
| ---------------- | ---------------- |
| **Auth**         | `coupons:delete` |
| **Response 204** | Soft delete      |

#### GET `/coupons/{id}`

|          |                |
| -------- | -------------- |
| **Auth** | `coupons:read` |

#### GET `/coupons`

|           |                                      |
| --------- | ------------------------------------ |
| **Auth**  | `coupons:read`                       |
| **Query** | `page`, `per_page`, `is_active`, `q` |

#### PATCH `/coupons/{id}/activate`

|                  |                         |
| ---------------- | ----------------------- |
| **Auth**         | `coupons:write`         |
| **Response 200** | `{ "is_active": true }` |

#### PATCH `/coupons/{id}/deactivate`

|                  |                          |
| ---------------- | ------------------------ |
| **Auth**         | `coupons:write`          |
| **Response 200** | `{ "is_active": false }` |

---

### 5.7 Customers

#### GET `/customers`

|           |                                       |
| --------- | ------------------------------------- |
| **Auth**  | `customers:read`                      |
| **Query** | `page`, `per_page`, `q` (name, email) |

#### GET `/customers/{id}`

|                  |                                      |
| ---------------- | ------------------------------------ |
| **Auth**         | `customers:read`                     |
| **Response 200** | CustomerDetail with addresses, stats |

#### GET `/customers/{id}/orders`

|                  |                         |
| ---------------- | ----------------------- |
| **Auth**         | `customers:read`        |
| **Query**        | `page`, `per_page`      |
| **Response 200** | Paginated order history |

---

### 5.8 Admin Users

#### POST `/users`

|                  |                                                                               |
| ---------------- | ----------------------------------------------------------------------------- |
| **Auth**         | `users:write`                                                                 |
| **Request**      | `{ "email", "password", "first_name", "last_name", "phone", "role_ids": [] }` |
| **Validation**   | email: unique; password: min 8, complexity rules                              |
| **Response 201** | AdminUserResponse (no password)                                               |

#### PUT `/users/{id}`

|          |               |
| -------- | ------------- |
| **Auth** | `users:write` |

#### DELETE `/users/{id}`

|                    |                                                    |
| ------------------ | -------------------------------------------------- |
| **Auth**           | `users:delete`                                     |
| **Business Rules** | Cannot delete self; cannot delete last super_admin |

#### GET `/users/{id}`

|          |              |
| -------- | ------------ |
| **Auth** | `users:read` |

#### GET `/users`

|          |              |
| -------- | ------------ |
| **Auth** | `users:read` |

#### PUT `/users/{id}/roles`

|             |                                      |
| ----------- | ------------------------------------ |
| **Auth**    | `users:manage_roles`                 |
| **Request** | `{ "role_ids": ["uuid1", "uuid2"] }` |

#### GET `/permissions`

|                  |                         |
| ---------------- | ----------------------- |
| **Auth**         | `users:read`            |
| **Response 200** | List of all permissions |

#### GET `/roles`

|                  |                                |
| ---------------- | ------------------------------ |
| **Auth**         | `users:read`                   |
| **Response 200** | List of roles with permissions |

---

### 5.9 Audit Logs

#### GET `/audit-logs`

|           |                                                                              |
| --------- | ---------------------------------------------------------------------------- |
| **Auth**  | `audit:read`                                                                 |
| **Query** | `page`, `per_page`, `admin_user_id`, `resource_type`, `action`, `from`, `to` |

#### GET `/audit-logs/{id}`

|          |              |
| -------- | ------------ |
| **Auth** | `audit:read` |

---

### 5.10 Standard Error Response

```json
{
  "statusCode": 400,
  "path": "/api/v1/auth/login",
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": {
      "email": "must be a valid email address",
      "price": "must be greater than or equal to 0"
    }
  }
}
```

| HTTP Status | Code                 | When                     |
| ----------- | -------------------- | ------------------------ |
| 400         | VALIDATION_ERROR     | Invalid input            |
| 401         | UNAUTHORIZED         | Missing/invalid token    |
| 403         | FORBIDDEN            | Insufficient permissions |
| 404         | NOT_FOUND            | Resource not found       |
| 409         | CONFLICT             | Duplicate slug/sku/email |
| 422         | UNPROCESSABLE_ENTITY | Business rule violation  |
| 429         | RATE_LIMITED         | Too many requests        |
| 500         | INTERNAL_ERROR       | Unexpected server error  |

---

## 6. Project Structure

```
ecommerce/
├── cmd/
│   └── api/
│       └── main.go                    # Entry point, graceful shutdown
├── internal/
│   ├── config/
│   │   └── config.go                  # Env-based configuration
│   ├── di/
│   │   └── container.go               # Dependency injection container
│   ├── domain/
│   │   ├── adminuser/
│   │   │   ├── entity.go
│   │   │   ├── repository.go          # Port interface
│   │   │   └── errors.go
│   │   ├── product/
│   │   ├── category/
│   │   ├── order/
│   │   ├── coupon/
│   │   ├── customer/
│   │   ├── audit/
│   │   └── shared/
│   │       ├── money.go
│   │       ├── email.go
│   │       └── pagination.go
│   ├── application/
│   │   ├── auth/
│   │   │   ├── login.go
│   │   │   ├── refresh.go
│   │   │   └── logout.go
│   │   ├── product/
│   │   ├── category/
│   │   ├── order/
│   │   ├── coupon/
│   │   ├── customer/
│   │   ├── dashboard/
│   │   ├── adminuser/
│   │   └── audit/
│   ├── infrastructure/
│   │   ├── persistence/
│   │   │   ├── postgres/
│   │   │   │   ├── connection.go
│   │   │   │   ├── admin_user_repo.go
│   │   │   │   ├── product_repo.go
│   │   │   │   └── ...
│   │   │   └── models/                # GORM models (DB layer)
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   └── bcrypt.go
│   │   └── audit/
│   │       └── writer.go
│   └── interfaces/
│       └── http/
│           ├── router.go
│           ├── middleware/
│           │   ├── auth.go
│           │   ├── rbac.go
│           │   ├── logging.go
│           │   ├── recovery.go
│           │   ├── request_id.go
│           │   └── cors.go
│           ├── handler/
│           │   ├── auth_handler.go
│           │   ├── product_handler.go
│           │   └── ...
│           ├── dto/
│           │   ├── request/
│           │   └── response/
│           └── response/
│               └── response.go        # Standard response helpers
├── pkg/
│   ├── apperror/
│   │   └── error.go
│   ├── validator/
│   │   └── validator.go
│   └── pagination/
│       └── pagination.go
├── migrations/
│   ├── 000001_init_schema.up.sql
│   ├── 000001_init_schema.down.sql
│   └── 000002_seed_data.up.sql
├── docs/
│   └── swagger/                       # Generated OpenAPI spec
├── tests/
│   ├── integration/
│   └── testutil/
├── docker/
│   ├── Dockerfile
│   └── docker-compose.yml
├── .env.example
├── Makefile
├── go.mod
├── go.sum
└── IMPLEMENTATION_PLAN.md
```

---

## 7. Implementation Phases

### Phase 1: Project Foundation ✅

- [x] Implementation plan
- [x] Go module setup with dependencies
- [x] Configuration management (`config` package)
- [x] Structured logging (`slog`)
- [x] PostgreSQL connection with GORM
- [x] Database migrations (golang-migrate)
- [x] HTTP server with chi router
- [x] Health check (`/healthz`) and readiness (`/readyz`)
- [x] Prometheus metrics endpoint (`/metrics`)
- [x] Graceful shutdown
- [x] Docker + Docker Compose
- [x] Standard error handling (`apperror` package)
- [x] Request ID middleware
- [x] CORS middleware
- [x] DI container skeleton

### Phase 2: Authentication & Authorization ✅

- [x] Domain models: AdminUser, Role, Permission, RefreshToken
- [x] GORM models + migration
- [x] Seed roles and permissions
- [x] JWT service (access + refresh)
- [x] Bcrypt password hashing
- [x] Login, Refresh, Logout, GetCurrentUser use cases
- [x] Auth middleware (Bearer token validation)
- [x] Role-based route guards (admin | customer)
- [x] Auth HTTP handlers
- [x] Unit tests for auth use cases
- [ ] Integration tests for auth endpoints (Phase 10)

### Phase 3: Product Management ✅

- [x] Domain: Product, ProductImage, Inventory, ProductAttribute
- [x] Repository implementation
- [x] CRUD + search use cases
- [x] Inventory management use case
- [x] Product HTTP handlers with validation
- [ ] Audit log on mutations (Phase 9)
- [x] Tests

### Phase 4: Category Management ✅

- [x] Domain: Category (hierarchical)
- [x] CRUD use cases with tree support
- [x] HTTP handlers
- [x] Tests

### Phase 5: Order Management ⬜

- [ ] Domain: Order, OrderItem, OrderStatusHistory
- [ ] Status state machine
- [ ] List, detail, update status, cancel, refund use cases
- [ ] HTTP handlers
- [ ] Tests

### Phase 6: Coupon Management ✅

- [x] Domain: Coupon with discount value objects
- [x] CRUD + activate/deactivate use cases
- [x] HTTP handlers
- [x] Tests

### Phase 7: Dashboard Analytics ⬜

- [ ] Read-model queries (raw SQL aggregations)
- [ ] Dashboard stats, revenue analytics, low stock, recent orders
- [ ] HTTP handlers
- [ ] Tests

### Phase 8: Customer Management ⬜

- [ ] Domain: Customer, CustomerAddress
- [ ] List, detail, purchase history use cases
- [ ] HTTP handlers
- [ ] Tests

### Phase 9: Audit Logging ⬜

- [ ] Domain: AuditLog
- [ ] Audit writer middleware/decorator
- [ ] List and detail use cases
- [ ] HTTP handlers
- [ ] Tests

### Phase 10: Testing & Production Hardening ⬜

- [x] Swagger/OpenAPI generation (swaggo)
- [ ] Integration test suite with testcontainers
- [ ] Rate limiting middleware
- [ ] Request validation library (go-playground/validator)
- [ ] Makefile with dev/test/migrate targets
- [ ] README with setup instructions
- [ ] Seed data for development

---

## 8. Observability & Production Concerns

### Health Checks

- **Liveness** (`/healthz`): Returns 200 if process is alive
- **Readiness** (`/readyz`): Returns 200 if PostgreSQL connection is healthy

### Metrics (Prometheus)

- `http_requests_total` (method, path, status)
- `http_request_duration_seconds` (histogram)
- `db_connections_active`
- `db_query_duration_seconds`

### Graceful Shutdown

1. Receive SIGINT/SIGTERM
2. Stop accepting new connections
3. Wait for in-flight requests (30s timeout)
4. Close database connections
5. Flush logs

### Security

- bcrypt cost factor 12
- JWT secret from env (min 32 chars)
- Rate limiting on auth endpoints (10 req/min)
- CORS restricted to configured origins
- No sensitive data in logs
- SQL injection prevented via parameterized queries (GORM)
- Input validation on all endpoints

---

_Document version: 1.0 — Generated from admin panel analysis of shop-panel-react.vercel.app_
