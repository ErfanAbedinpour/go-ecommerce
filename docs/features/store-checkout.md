# Store Checkout

> **Route:** `/checkout`  
> **UI:** [store-os-eta.vercel.app/checkout](https://store-os-eta.vercel.app/checkout)  
> **Locale:** Persian (fa-IR), RTL  
> **Currency:** Toman (تومان)

---

## Purpose

Checkout converts the shopping cart into a placed order through a **3-step wizard**: (1) cart review, (2) shipping/address, (3) payment. Line items display full variant info (size, color, weight, pattern). Supports guest checkout and authenticated customers. Prices, shipping, discounts, and totals are shown in Persian Toman formatting.

---

## User Flow

```mermaid
flowchart TD
    A[Cart has items] --> B[/checkout step 1]
    B --> C[Review line items + coupon]
    C --> D[Next → step 2]
    D --> E[Shipping address form]
    E --> F{Authenticated?}
    F -->|Yes| G[Pre-fill from profile addresses]
    F -->|No| H[Guest fields: name, email, phone]
    G --> I[Next → step 3]
    H --> I
    I --> J[Payment method selection]
    J --> K[POST /store/checkout/preview]
    K --> L[Confirm totals]
    L --> M[POST /store/checkout]
    M --> N{Payment}
    N -->|Online gateway| O[Redirect to PSP]
    N -->|COD| P[Order confirmed]
    O --> Q[Callback → order confirmed]
    P --> R[/account/orders/:id or thank-you]
    Q --> R
```

### Step 1 — Cart review (`?step=cart`)

- List items from `localStorage` cart with thumbnail, name, variant label, SKU code, unit price (Toman), quantity stepper, line total.
- Remove item / update quantity (re-validate stock via preview API).
- Coupon code input → `POST /store/coupons/validate`.
- Subtotal, discount, shipping estimate, total.
- CTA: "ادامه فرآیند خرید" → step 2.

### Step 2 — Shipping (`?step=shipping`)

- Fields (Persian labels): نام، نام خانوادگی، موبایل، استان، شهر، آدرس کامل، کد پستی
- Optional: save address for logged-in customers.
- Shipping method: پست پیشتاز / پیک (if multiple, radio list).
- CTA → step 3.

### Step 3 — Payment (`?step=payment`)

- Payment methods: درگاه آنلاین (Zarinpal/IDPay placeholder), پرداخت در محل (COD) if enabled.
- Order summary sidebar (sticky on desktop).
- CTA: "پرداخت و ثبت سفارش".
- On success: clear cart, show order number + tracking info.

---

## Business Logic

### Cart source (v1)

- **Client-side cart** in `localStorage` (`store_cart_v1`).
- Server validates on preview/submit; rejects stale prices or insufficient stock.

### Pricing (Toman)

| Field | Calculation |
|-------|-------------|
| Line total | `unit_price_toman * quantity` |
| Subtotal | Sum of line totals |
| Discount | From coupon (`percentage` or `fixed` on subtotal) |
| Shipping | Flat rate or weight-based rules from `store_settings` |
| Tax | VAT if applicable (configurable; default 0% for v1) |
| **Total** | `subtotal - discount + shipping + tax` |

All amounts returned as **integers** (Toman). UI formats with `toLocaleString('fa-IR')`.

### Stock reservation

- On `POST /checkout`: decrement `inventories.quantity` (or SKU stock) inside transaction.
- If insufficient stock → `409 CONFLICT` with per-line errors.

### Coupon validation

- Check: `is_active`, not expired, not exhausted, `min_order_amount`, usage limits per customer.
- Apply discount to subtotal before shipping.

### Order creation

- Generate `order_number` (e.g. `ORD-1403-00001`).
- Snapshot line items: product name, SKU code, variant attributes, unit price at purchase time.
- `status = pending`, `payment_status = unpaid` (or `paid` after gateway callback).
- Guest orders: create or link `customers` record by email.

### Payment gateway (Iran)

- v1: Return `payment_url` from checkout response for redirect.
- Callback endpoint verifies transaction, sets `payment_status = paid`, `transaction_id`.
- COD: skip gateway; `payment_method = 'cod'`.

### Guest vs authenticated

| Aspect | Guest | Customer |
|--------|-------|----------|
| Address | Enter each time | Saved addresses |
| Order history | Email lookup only | `/account` |
| Coupon per-user limits | By email | By `customer_id` |

---

## Edge Cases

| Scenario | Behavior |
|----------|----------|
| Empty cart on checkout | Redirect to `/products` |
| Price changed since add-to-cart | Preview returns updated prices; show notice |
| Product deactivated mid-checkout | Remove line + notify user |
| SKU out of stock | Block submit; highlight line |
| Invalid coupon | Show error; do not apply discount |
| Payment gateway timeout | Order stays `pending`; retry payment link |
| Duplicate submit | Idempotency key `Idempotency-Key` header |
| Browser back from PSP | Show "پرداخت ناموفق" with retry |
| Minimum order amount | Enforce from settings |
| RTL invoice preview | All labels Persian |

---

## Dependencies

### Backend modules

| Module | Role |
|--------|------|
| `internal/application/storefront/checkout` | Preview, place order |
| `internal/application/order` | Order aggregate, stock decrement |
| `internal/application/coupon` | Validate & apply |
| `internal/application/customer` | Guest customer upsert |
| `internal/application/product` | Stock check |

### External

- Payment service provider (Zarinpal / IDPay / etc.)
- SMTP for order confirmation email

### Frontend

- Step wizard state synced to URL `?step=cart|shipping|payment`
- `localStorage` cart read/write
- Persian address province/city data (static JSON)

---

## Required APIs

### `POST /api/v1/store/checkout/preview`

Validate cart and compute totals without creating order.

**Auth:** Optional Bearer (for customer-specific coupon limits)

**Request**

```json
{
  "items": [
    {
      "product_id": "uuid",
      "sku_id": "uuid",
      "quantity": 2
    }
  ],
  "coupon_code": "SUMMER20",
  "shipping_method": "standard",
  "shipping_city": "تهران"
}
```

**Response 200**

```json
{
  "items": [
    {
      "product_id": "uuid",
      "sku_id": "uuid",
      "product_name": "کاشی سرامیک ۶۰×۶۰",
      "sku_code": "TILE-60-WHT-STN",
      "variant_label": "۶۰×۶۰ · سفید · سنگی",
      "thumbnail_url": "https://…",
      "unit_price_toman": 399000,
      "quantity": 2,
      "line_total_toman": 798000,
      "is_available": true,
      "available_quantity": 24
    }
  ],
  "summary": {
    "subtotal_toman": 798000,
    "discount_toman": 159600,
    "shipping_toman": 45000,
    "tax_toman": 0,
    "total_toman": 683400,
    "currency": "IRT",
    "currency_label": "تومان"
  },
  "coupon": {
    "code": "SUMMER20",
    "discount_type": "percentage",
    "discount_value": 20,
    "is_valid": true
  },
  "warnings": []
}
```

**Errors:** `400` invalid items, `409` stock conflict with `details.unavailable_items[]`.

### `POST /api/v1/store/coupons/validate`

**Request**

```json
{
  "code": "SUMMER20",
  "subtotal_toman": 798000
}
```

**Response 200**

```json
{
  "is_valid": true,
  "discount_toman": 159600,
  "message": "کد تخفیف اعمال شد."
}
```

**Response 200 (invalid)**

```json
{
  "is_valid": false,
  "discount_toman": 0,
  "message": "کد تخفیف منقضی شده است."
}
```

### `POST /api/v1/store/checkout`

Place order.

**Headers:** `Idempotency-Key: <uuid>` (recommended)

**Auth:** Optional Bearer

**Request**

```json
{
  "items": [
    { "product_id": "uuid", "sku_id": "uuid", "quantity": 2 }
  ],
  "coupon_code": "SUMMER20",
  "customer": {
    "email": "reza@example.com",
    "first_name": "رضا",
    "last_name": "محمدی",
    "phone": "09121234567"
  },
  "shipping_address": {
    "street": "خیابان ولیعصر، پلاک ۱۲",
    "city": "تهران",
    "state": "تهران",
    "postal_code": "1234567890",
    "country": "IR"
  },
  "billing_address": { "same_as_shipping": true },
  "shipping_method": "standard",
  "payment_method": "online",
  "notes": "تحویل بعد از ظهر"
}
```

**Response 201**

```json
{
  "order_id": "uuid",
  "order_number": "ORD-1403-001248",
  "status": "pending",
  "payment_status": "unpaid",
  "total_toman": 683400,
  "payment_url": "https://gateway.example.com/pay/…",
  "expires_at": "2026-06-28T12:30:00Z"
}
```

**COD response:** `payment_url: null`, `payment_status: "unpaid"`, message about pay on delivery.

### `GET /api/v1/store/checkout/shipping-methods`

**Query:** `city` (optional)

**Response**

```json
{
  "data": [
    {
      "id": "standard",
      "name": "پست پیشتاز",
      "price_toman": 45000,
      "estimated_days": "۳–۵ روز کاری"
    },
    {
      "id": "express",
      "name": "پیک فوری",
      "price_toman": 80000,
      "estimated_days": "۱ روز کاری"
    }
  ]
}
```

### `POST /api/v1/store/checkout/payment/callback`

Payment gateway webhook (server-to-server).

**Auth:** Gateway signature verification

**Effect:** Update `orders.payment_status`, `transaction_id`; emit confirmation email.

### `GET /api/v1/store/settings/checkout`

Public checkout config: enabled payment methods, min order, COD availability.

```json
{
  "min_order_toman": 100000,
  "payment_methods": ["online", "cod"],
  "currency_label": "تومان"
}
```

---

## Database Impact

### Writes (order placement transaction)

| Table | Operation |
|-------|-----------|
| `orders` | INSERT |
| `order_items` | INSERT (with `product_sku`, variant snapshot) |
| `inventories` | UPDATE decrement quantity |
| `coupons` | UPDATE `usage_count` |
| `customers` | INSERT or SELECT for guest |
| `order_status_history` | INSERT initial status |

### Order item extension (recommended)

Add columns to `order_items` for variant snapshot:

```sql
ALTER TABLE order_items ADD COLUMN sku_id UUID REFERENCES skus(id);
ALTER TABLE order_items ADD COLUMN variant_attributes JSONB;
```

### Reads

- `products`, `skus`, `inventories`, `coupons`, `store_settings`

---

## Validation

### Preview / Checkout items

| Rule | Detail |
|------|--------|
| `items` | Min 1, max 50 lines |
| `quantity` | Integer 1–999 per line |
| `product_id`, `sku_id` | Valid UUIDs; SKU must belong to product |
| `coupon_code` | Max 50 chars |

### Customer / address

| Field | Rules |
|-------|-------|
| `email` | Required, valid email |
| `first_name`, `last_name` | Required, 2–100 chars |
| `phone` | Required, `^09\d{9}$` |
| `street` | Required, 10–500 chars |
| `city`, `state` | Required |
| `postal_code` | Required, 10 digits (Iran) |
| `country` | Default `IR` |

### Payment

| `payment_method` | Enum: `online`, `cod` |
|------------------|----------------------|

---

## Permissions

| Action | Role |
|--------|------|
| Preview checkout | Public |
| Validate coupon | Public |
| Place order | Public (guest) or Customer |
| View order after place | Customer (owner) or guest via email+order_number (future) |
| Payment callback | Gateway (signed) |

---

## State Management

### Cart (`localStorage.store_cart_v1`)

```typescript
interface CartState {
  items: CartLineItem[];
  updated_at: string; // ISO
}
```

### Checkout wizard

| State | Storage |
|-------|---------|
| Current step | URL `?step=cart\|shipping\|payment` |
| Shipping form | React Hook Form; persist to `sessionStorage.checkout_draft` |
| Coupon code | React state + preview response |
| Preview summary | React Query from preview API |
| Idempotency key | Generate once per session `sessionStorage.checkout_idempotency` |
| Payment redirect | On return, read `?order_id=` query param |

### Post-success

1. Clear `localStorage.store_cart_v1`
2. Clear `sessionStorage.checkout_draft`
3. Redirect to thank-you page with order number

### Optimistic UI

- Quantity changes trigger debounced `POST /checkout/preview`
- Disable submit while preview loading or stock invalid
