package i18n

import "strings"

// Message and resource keys used across the application.
const (
	KeyErrorNotFound              = "error.not_found"
	KeyErrorInternal              = "error.internal"
	KeyErrorValidation            = "error.validation"
	KeyErrorValidationRequest     = "error.validation_request"
	KeyErrorUnauthorized            = "error.unauthorized"
	KeyErrorForbidden               = "error.forbidden"
	KeyErrorConflict                = "error.conflict"
	KeyErrorUnprocessable           = "error.unprocessable"
	KeyErrorRateLimited             = "error.rate_limited"

	KeyUserInvalidCredentials     = "user.invalid_credentials"
	KeyUserAccountDisabled          = "user.account_disabled"
	KeyUserInvalidToken             = "user.invalid_token"
	KeyUserTokenRevoked             = "user.token_revoked"
	KeyUserForbiddenRole            = "user.forbidden_role"
	KeyUserEmailTaken               = "user.email_taken"
	KeyUserSignupDisabled           = "user.signup_disabled"
	KeyUserInvalidResetToken        = "user.invalid_reset_token"
	KeyUserCannotDeleteSelf         = "user.cannot_delete_self"
	KeyUserLastAdmin                = "user.last_admin"

	KeyCustomerAccountExistsLogin = "customer.account_exists_login_required"
	KeyCustomerEmailTaken         = "customer.email_taken"
	KeyCustomerHasOrders            = "customer.has_orders"
	KeyCustomerInvalidType          = "customer.invalid_type"

	KeyCartEmpty                    = "cart.empty"
	KeyCartItemMissing              = "cart.item_missing"

	KeyOrderInvalidStatusTransition = "order.invalid_status_transition"
	KeyOrderCannotCancel            = "order.cannot_cancel"
	KeyOrderCannotRefund            = "order.cannot_refund"
	KeyOrderInvalidRefundAmount     = "order.invalid_refund_amount"
	KeyOrderInsufficientStock       = "order.insufficient_stock"
	KeyOrderEmpty                   = "order.empty"
	KeyOrderInvalidSKU              = "order.invalid_sku"
	KeyOrderInvalidDateRange        = "order.invalid_date_range"
	KeyOrderPaymentAlreadyPaid      = "order.payment_already_paid"
	KeyOrderPaymentExpired          = "order.payment_expired"

	KeyProductSlugConflict          = "product.slug_conflict"
	KeyProductSKUConflict           = "product.sku_conflict"
	KeyProductHasActiveOrders       = "product.has_active_orders"
	KeyProductInvalidSalePrice      = "product.invalid_sale_price"
	KeyProductInvalidStockLevel     = "product.invalid_stock_level"
	KeyProductEmptyAttributeName    = "product.empty_attribute_name"
	KeyProductDuplicateAttributeName = "product.duplicate_attribute_name"
	KeyProductEmptyAttributeValues  = "product.empty_attribute_values"
	KeyProductEmptyAttributeValue   = "product.empty_attribute_value"
	KeyProductDuplicateAttributeValue = "product.duplicate_attribute_value"
	KeyProductMaxVariantsExceeded   = "product.max_variants_exceeded"

	KeyCategorySlugConflict         = "category.slug_conflict"
	KeyCategoryInvalidParent        = "category.invalid_parent"
	KeyCategoryHasChildren          = "category.has_children"
	KeyCategoryHasProducts          = "category.has_products"

	KeyBrandSlugConflict            = "brand.slug_conflict"
	KeyBrandNameConflict            = "brand.name_conflict"
	KeyBrandHasProducts             = "brand.has_products"

	KeyAttributeSlugConflict        = "attribute.slug_conflict"
	KeyAttributeNameConflict        = "attribute.name_conflict"
	KeyAttributeHasValues           = "attribute.has_values"

	KeyAttributeValueConflict       = "attribute_value.conflict"

	KeyCouponCodeConflict           = "coupon.code_conflict"
	KeyCouponInvalidDiscount        = "coupon.invalid_discount"
	KeyCouponInvalidPercentage      = "coupon.invalid_percentage"
	KeyCouponNotApplicable          = "coupon.not_applicable"
	KeyCouponExpired                = "coupon.expired"
	KeyCouponExhausted              = "coupon.exhausted"
	KeyCouponMinOrderNotMet         = "coupon.min_order_not_met"

	KeyWishlistAlreadyExists        = "wishlist.already_exists"

	KeyProductReviewAlreadyReviewed = "product_review.already_reviewed"

	KeyProductQuestionNotFound      = "product_question.not_found"

	KeyBlogDuplicateSlug            = "blog.duplicate_slug"

	KeyThemeAlreadyPurchased        = "theme.already_purchased"
	KeyThemeNotPurchased            = "theme.not_purchased"
	KeyThemeInactive                = "theme.inactive"

	KeyStoreInvalidSlideType        = "store.invalid_slide_type"
	KeyStoreDuplicateSlideItem      = "store.duplicate_slide_item"

	KeyContactInvalidSource         = "contact.invalid_source"
	KeyContactInvalidStatus         = "contact.invalid_status"

	KeyDashboardInvalidPeriod       = "dashboard.invalid_period"
	KeyDashboardInvalidDateRange    = "dashboard.invalid_date_range"
	KeyDashboardInvalidLimit        = "dashboard.invalid_limit"

	KeyValidationRequired           = "validation.required"
	KeyValidationEmail              = "validation.email"
	KeyValidationMin                = "validation.min"
	KeyValidationMax                = "validation.max"
	KeyValidationGte                = "validation.gte"
	KeyValidationGt                 = "validation.gt"
	KeyValidationUUID               = "validation.uuid"
	KeyValidationOneOf              = "validation.oneof"
	KeyValidationDefault            = "validation.default"

	KeyResourceUser                 = "resource.user"
	KeyResourceCustomer             = "resource.customer"
	KeyResourceProduct              = "resource.product"
	KeyResourceOrder                = "resource.order"
	KeyResourceCategory             = "resource.category"
	KeyResourceBrand                = "resource.brand"
	KeyResourceCoupon               = "resource.coupon"
	KeyResourceCart                 = "resource.cart"
	KeyResourceCartItem             = "resource.cart_item"
	KeyResourceTheme                = "resource.theme"
	KeyResourceStoreStyle           = "resource.store_style"
	KeyResourceWishlistItem         = "resource.wishlist_item"
	KeyResourceProductReview        = "resource.product_review"
	KeyResourceProductQuestion      = "resource.product_question"
	KeyResourceBlogPost             = "resource.blog_post"
	KeyResourceBlogCategory         = "resource.blog_category"
	KeyResourceBlogComment          = "resource.blog_comment"
	KeyResourceContactMessage       = "resource.contact_message"
	KeyResourceProductAttribute     = "resource.product_attribute"
	KeyResourceAttributeValue       = "resource.attribute_value"
	KeyResourceStoreContent         = "resource.store_content"
	KeyResourceStorefrontHero       = "resource.storefront_hero"
	KeyResourceProductSlide         = "resource.product_slide"
	KeyResourceProBanner            = "resource.pro_banner"
	KeyResourcePartnerBrand         = "resource.partner_brand"
	KeyResourceHomepageReview       = "resource.homepage_review"
	KeyResourceFAQItem              = "resource.faq_item"
	KeyResourceSlideItem            = "resource.slide_item"
	KeyResourceParentCategory       = "resource.parent_category"
)

// ResourceKey maps a resource label used in NotFound() to a translation key.
func ResourceKey(resource string) string {
	switch strings.ToLower(strings.TrimSpace(resource)) {
	case "user":
		return KeyResourceUser
	case "customer":
		return KeyResourceCustomer
	case "product":
		return KeyResourceProduct
	case "order":
		return KeyResourceOrder
	case "category":
		return KeyResourceCategory
	case "parent category":
		return KeyResourceParentCategory
	case "brand":
		return KeyResourceBrand
	case "coupon":
		return KeyResourceCoupon
	case "cart not found", "cart":
		return KeyResourceCart
	case "cart item not found", "cart item":
		return KeyResourceCartItem
	case "theme":
		return KeyResourceTheme
	case "store style":
		return KeyResourceStoreStyle
	case "wishlist item":
		return KeyResourceWishlistItem
	case "product review":
		return KeyResourceProductReview
	case "product question":
		return KeyResourceProductQuestion
	case "blog post":
		return KeyResourceBlogPost
	case "blog category":
		return KeyResourceBlogCategory
	case "blog comment":
		return KeyResourceBlogComment
	case "contact message":
		return KeyResourceContactMessage
	case "product attribute":
		return KeyResourceProductAttribute
	case "attribute value":
		return KeyResourceAttributeValue
	case "store content":
		return KeyResourceStoreContent
	case "storefront hero":
		return KeyResourceStorefrontHero
	case "product slide":
		return KeyResourceProductSlide
	case "pro banner":
		return KeyResourceProBanner
	case "partner brand":
		return KeyResourcePartnerBrand
	case "homepage review":
		return KeyResourceHomepageReview
	case "faq item":
		return KeyResourceFAQItem
	case "slide item":
		return KeyResourceSlideItem
	default:
		return resource
	}
}
