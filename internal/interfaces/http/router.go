package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"app/docs/swagger"
	"app/internal/config"
	"app/internal/di"
	"app/internal/domain/user"
	"app/internal/interfaces/http/handler"
	appmiddleware "app/internal/interfaces/http/middleware"
)

// NewRouter creates the HTTP router with all middleware and routes.
func NewRouter(c *di.Container) http.Handler {
	r := chi.NewRouter()

	r.Use(appmiddleware.RequestID)
	r.Use(appmiddleware.Recovery(c.Log))
	r.Use(appmiddleware.Logging(c.Log))
	r.Use(appmiddleware.Metrics)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Compress(5))
	r.Use(corsMiddleware(c.Config))

	// ── Public routes (no authentication) ──────────────────────────
	r.Get("/healthz", c.Health.Liveness)
	r.Get("/readyz", c.Health.Readiness)
	r.Handle("/metrics", promhttp.Handler())

	// Swagger UI and OpenAPI spec
	swaggerHandler := httpSwagger.Handler(
		httpSwagger.InstanceName(swagger.SwaggerInfo.InstanceName()),
	)
	if c.Config.IsProduction() {
		r.Route("/swagger", func(r chi.Router) {
			r.Use(appmiddleware.BasicAuth(c.Config.Swagger.Username, c.Config.Swagger.Password))
			r.Handle("/*", swaggerHandler)
		})
	} else {
		r.Get("/swagger/*", swaggerHandler)
	}

	// Local uploads only — S3/MinIO objects are served directly from object storage.
	if c.Config.Upload.Provider == "local" {
		uploadDir := c.Config.Upload.Dir
		r.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir(uploadDir))))
	}

	r.Route("/api/v1", func(r chi.Router) {
		registerPublicRoutes(r, c)
		registerAuthenticatedRoutes(r, c)
		registerStoreRoutes(r, c)
		registerAdminRoutes(r, c)
	})

	return r
}

// registerPublicRoutes — accessible without a token.
func registerPublicRoutes(r chi.Router, c *di.Container) {
	authLimiter := appmiddleware.NewRateLimiter(5, 10)

	r.Route("/auth", func(r chi.Router) {
		r.Use(appmiddleware.RateLimit(authLimiter))
		r.Use(appmiddleware.CartSession())
		r.Post("/login", c.Auth.Login)
		r.Post("/refresh", c.Auth.Refresh)
		r.Post("/signup", c.Auth.Signup)
		r.Post("/forgot-password", c.Auth.ForgotPassword)
		r.Post("/reset-password", c.Auth.ResetPassword)
	})
}

// registerAuthenticatedRoutes — any authenticated user (admin or customer).
func registerAuthenticatedRoutes(r chi.Router, c *di.Container) {
	r.Group(func(r chi.Router) {
		r.Use(appmiddleware.Authenticate(c.JWT))
		r.Post("/auth/logout", c.Auth.Logout)
		r.Get("/auth/me", c.Auth.Me)
	})
}

// registerAdminRoutes — admin role required; enforced at router layer.
func registerAdminRoutes(r chi.Router, c *di.Container) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(appmiddleware.Authenticate(c.JWT))
		r.Use(appmiddleware.RequireRole(user.RoleAdmin))
		r.Use(appmiddleware.AuditLog(c.AuditRepo))

		r.Get("/", handler.AdminIndex)

		registerDashboardRoutes(r, c)
		registerProductRoutes(r, c)
		registerCatalogRoutes(r, c)
		registerCategoryRoutes(r, c)
		registerCouponRoutes(r, c)
		registerCustomerRoutes(r, c)
		registerUserRoutes(r, c)
		registerOrderRoutes(r, c)
		registerSettingsRoutes(r, c)
		registerAdminStorefrontRoutes(r, c)
		registerAdminThemeRoutes(r, c)
		registerAdminEngagementRoutes(r, c)
		registerAdminBlogRoutes(r, c)
	})
}

func registerDashboardRoutes(r chi.Router, c *di.Container) {
	r.Route("/dashboard", func(r chi.Router) {
		r.Get("/stats", c.Dashboard.Stats)
		r.Get("/revenue", c.Dashboard.Revenue)
		r.Get("/low-stock", c.Dashboard.LowStock)
		r.Get("/recent-orders", c.Dashboard.RecentOrders)
		r.Get("/featured-products", c.Dashboard.FeaturedProducts)
	})
}

func registerProductRoutes(r chi.Router, c *di.Container) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/stats", c.Product.Stats)
		r.Get("/search", c.Product.Search)
		r.Get("/", c.Product.List)
		r.Post("/", c.Product.Create)
		r.Get("/{id}", c.Product.Get)
		r.Put("/{id}", c.Product.Update)
		r.Delete("/{id}", c.Product.Delete)
		r.Patch("/{id}/inventory", c.Product.UpdateInventory)
	})
}

func registerCatalogRoutes(r chi.Router, c *di.Container) {
	r.Post("/uploads", c.Upload.Upload)

	r.Route("/brands", func(r chi.Router) {
		r.Get("/", c.Brand.List)
		r.Post("/", c.Brand.Create)
		r.Get("/{id}", c.Brand.Get)
		r.Put("/{id}", c.Brand.Update)
		r.Delete("/{id}", c.Brand.Delete)
	})

	r.Route("/product-attributes", func(r chi.Router) {
		r.Get("/", c.Attribute.ListAttributes)
		r.Post("/", c.Attribute.CreateAttribute)
		r.Get("/{id}", c.Attribute.GetAttribute)
		r.Put("/{id}", c.Attribute.UpdateAttribute)
		r.Delete("/{id}", c.Attribute.DeleteAttribute)
	})

	r.Route("/product-attribute-values", func(r chi.Router) {
		r.Get("/", c.Attribute.ListValues)
		r.Post("/", c.Attribute.CreateValue)
		r.Get("/{id}", c.Attribute.GetValue)
		r.Put("/{id}", c.Attribute.UpdateValue)
		r.Delete("/{id}", c.Attribute.DeleteValue)
	})
}

func registerCategoryRoutes(r chi.Router, c *di.Container) {
	r.Route("/categories", func(r chi.Router) {
		r.Get("/", c.Category.List)
		r.Post("/", c.Category.Create)
		r.Get("/{id}", c.Category.Get)
		r.Put("/{id}", c.Category.Update)
		r.Delete("/{id}", c.Category.Delete)
	})
}

func registerCustomerRoutes(r chi.Router, c *di.Container) {
	r.Route("/customers", func(r chi.Router) {
		r.Get("/", c.Customer.List)
		r.Get("/{id}/orders", c.Customer.ListOrders)
		r.Put("/{id}", c.Customer.Update)
		r.Delete("/{id}", c.Customer.Delete)
		r.Get("/{id}", c.Customer.Get)
	})
}

func registerUserRoutes(r chi.Router, c *di.Container) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/", c.User.List)
		r.Post("/", c.User.Create)
		r.Get("/{id}", c.User.Get)
		r.Put("/{id}", c.User.Update)
		r.Delete("/{id}", c.User.Delete)
	})
}

func registerOrderRoutes(r chi.Router, c *di.Container) {
	r.Route("/orders", func(r chi.Router) {
		r.Get("/", c.Order.List)
		r.Post("/", c.Order.Create)
		r.Patch("/{id}/status", c.Order.UpdateStatus)
		r.Patch("/{id}/notes", c.Order.UpdateNotes)
		r.Post("/{id}/cancel", c.Order.Cancel)
		r.Post("/{id}/refund", c.Order.Refund)
		r.Get("/{id}/invoice", c.Order.GetInvoice)
		r.Get("/{id}", c.Order.Get)
	})
}

func registerSettingsRoutes(r chi.Router, c *di.Container) {
	r.Route("/settings", func(r chi.Router) {
		r.Get("/site", c.Settings.GetSite)
		r.Put("/site", c.Settings.UpdateSite)
		r.Get("/contact", c.Settings.GetContact)
		r.Put("/contact", c.Settings.UpdateContact)
		r.Get("/social", c.Settings.GetSocial)
		r.Put("/social", c.Settings.UpdateSocial)
		r.Get("/seo", c.Settings.GetSEO)
		r.Put("/seo", c.Settings.UpdateSEO)
	})
	r.Route("/navigation", func(r chi.Router) {
		r.Get("/", c.Settings.GetNavigation)
		r.Put("/", c.Settings.UpdateNavigation)
	})
}

func registerCouponRoutes(r chi.Router, c *di.Container) {
	r.Route("/coupons", func(r chi.Router) {
		r.Get("/", c.Coupon.List)
		r.Post("/", c.Coupon.Create)
		r.Get("/{id}", c.Coupon.Get)
		r.Put("/{id}", c.Coupon.Update)
		r.Delete("/{id}", c.Coupon.Delete)
		r.Patch("/{id}/activate", c.Coupon.Activate)
		r.Patch("/{id}/deactivate", c.Coupon.Deactivate)
	})
}

func registerStoreRoutes(r chi.Router, c *di.Container) {
	contactLimiter := appmiddleware.NewRateLimiter(3, 10)

	r.Route("/store", func(r chi.Router) {
		r.Get("/products", c.Store.ListProducts)
		r.Get("/products/search", c.Store.SearchProducts)
		r.Get("/homepage", c.Store.GetHomepage)
		r.Get("/categories", c.Store.ListCategories)
		r.Get("/navigation", c.Store.GetNavigation)
		r.Get("/theme", c.Store.GetTheme)

		r.Get("/products/{id}/related", c.Store.ListRelatedProducts)

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.OptionalAuthenticate(c.JWT))
			r.Get("/products/{slugOrId}", c.Store.GetProduct)
		})

		r.Get("/brands", c.Store.ListBrands)
		r.Get("/settings", c.Store.GetSettings)
		r.Get("/settings/checkout", c.Store.GetCheckoutSettings)
		r.Get("/about", c.Store.GetAbout)
		r.Get("/checkout/shipping-methods", c.Store.GetShippingMethods)
		r.Post("/coupons/validate", c.Store.ValidateCoupon)

		// Public engagement & blog storefront routes
		r.Get("/products/{productId}/reviews", c.ProductReview.ListByProduct)
		r.Get("/products/{productId}/reviews/summary", c.ProductReview.GetSummary)
		r.Get("/products/{productId}/questions", c.ProductQuestion.ListByProduct)
		r.Post("/products/{productId}/questions", c.ProductQuestion.Ask)

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.RateLimit(contactLimiter))
			r.Post("/contact", c.Contact.Submit)
		})

		r.Get("/blog/posts", c.Blog.StoreListPosts)
		r.Get("/blog/posts/{slug}", c.Blog.StoreGetPost)
		r.Get("/blog/categories", c.Blog.StoreListCategories)
		r.Get("/blog/posts/{postId}/comments", c.Blog.StoreListComments)
		r.Post("/blog/posts/{postId}/comments", c.Blog.StoreSubmitComment)
		r.Get("/blog", c.Blog.StoreListPostsAlias)
		r.Get("/blog/{slug}/comments", c.Blog.StoreListCommentsBySlug)

		checkoutLimiter := appmiddleware.NewRateLimiter(2, 5)

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.CartSession())
			r.Use(appmiddleware.OptionalAuthenticate(c.JWT))

			r.Get("/cart", c.Cart.GetCart)
			r.Post("/cart/items", c.Cart.AddCartItem)
			r.Patch("/cart/items/{productId}", c.Cart.UpdateCartItem)
			r.Delete("/cart/items/{productId}", c.Cart.RemoveCartItem)
			r.Delete("/cart", c.Cart.ClearCart)

			r.Group(func(r chi.Router) {
				r.Use(appmiddleware.RateLimit(checkoutLimiter))
				r.Post("/checkout/validate-customer", c.Store.ValidateGuestCheckoutCustomer)
				r.Post("/checkout/preview", c.Store.PreviewCheckout)
				r.Post("/checkout", c.Store.Checkout)
				r.Post("/checkout/payment/callback", c.Store.PaymentCallback)
			})

			r.Post("/products/{productId}/reviews", c.ProductReview.Submit)
		})

		r.Group(func(r chi.Router) {
			r.Use(appmiddleware.Authenticate(c.JWT))
			r.Use(appmiddleware.RequireCustomer())
			r.Get("/account/profile", c.Store.GetAccountProfile)
			r.Put("/account/profile", c.Store.UpdateAccountProfile)
			r.Get("/account/orders", c.Store.ListAccountOrders)
			r.Get("/account/orders/{id}", c.Store.GetAccountOrder)
			r.Post("/account/wishlist", c.Wishlist.Add)
			r.Delete("/account/wishlist/{productId}", c.Wishlist.Remove)
			r.Get("/account/wishlist", c.Wishlist.List)
			r.Get("/account/wishlist/ids", c.Wishlist.ListIDs)
			r.Get("/account/wishlist/count", c.Wishlist.Count)
		})
	})
}

func registerAdminStorefrontRoutes(r chi.Router, c *di.Container) {
	r.Route("/storefront", func(r chi.Router) {
		r.Get("/hero", c.Storefront.GetHero)
		r.Put("/hero", c.Storefront.UpdateHero)

		r.Get("/product-slides", c.Storefront.ListProductSlides)
		r.Put("/product-slides/{slideType}", c.Storefront.UpdateProductSlide)
		r.Post("/product-slides/{slideType}/items", c.Storefront.CreateSlideItem)
		r.Put("/product-slide-items/{id}", c.Storefront.UpdateSlideItem)
		r.Delete("/product-slide-items/{id}", c.Storefront.DeleteSlideItem)

		r.Get("/pro-banners", c.Storefront.ListProBanners)
		r.Post("/pro-banners", c.Storefront.CreateProBanner)
		r.Put("/pro-banners/{id}", c.Storefront.UpdateProBanner)
		r.Delete("/pro-banners/{id}", c.Storefront.DeleteProBanner)

		r.Get("/partner-brands", c.Storefront.ListPartnerBrands)
		r.Post("/partner-brands", c.Storefront.CreatePartnerBrand)
		r.Put("/partner-brands/{id}", c.Storefront.UpdatePartnerBrand)
		r.Delete("/partner-brands/{id}", c.Storefront.DeletePartnerBrand)

		r.Get("/homepage-reviews", c.Storefront.ListHomepageReviews)
		r.Post("/homepage-reviews", c.Storefront.CreateHomepageReview)
		r.Put("/homepage-reviews/{id}", c.Storefront.UpdateHomepageReview)
		r.Delete("/homepage-reviews/{id}", c.Storefront.DeleteHomepageReview)

		r.Get("/faq", c.Storefront.GetFAQSection)
		r.Put("/faq", c.Storefront.UpdateFAQSection)
		r.Get("/faq/items", c.Storefront.ListFAQItems)
		r.Post("/faq/items", c.Storefront.CreateFAQItem)
		r.Put("/faq/items/{id}", c.Storefront.UpdateFAQItem)
		r.Delete("/faq/items/{id}", c.Storefront.DeleteFAQItem)

		r.Get("/contact-section", c.Storefront.GetContactSection)
		r.Put("/contact-section", c.Storefront.UpdateContactSection)

		r.Get("/navigation", c.Storefront.GetNavigation)
		r.Put("/navigation", c.Storefront.UpdateNavigation)
	})
}

func registerAdminThemeRoutes(r chi.Router, c *di.Container) {
	r.Get("/themes", c.Theme.ListThemes)
	r.Post("/themes/{id}/purchase", c.Theme.PurchaseTheme)
	r.Get("/store-style", c.Theme.GetStoreStyle)
	r.Put("/store-style", c.Theme.UpdateStoreStyle)
}

func registerAdminEngagementRoutes(r chi.Router, c *di.Container) {
	r.Route("/reviews", func(r chi.Router) {
		r.Get("/", c.ProductReview.ListAdmin)
		r.Patch("/{id}/status", c.ProductReview.UpdateStatus)
		r.Delete("/{id}", c.ProductReview.Delete)
	})

	r.Route("/questions", func(r chi.Router) {
		r.Get("/", c.ProductQuestion.ListAdmin)
		r.Post("/{id}/answer", c.ProductQuestion.Answer)
		r.Delete("/{id}", c.ProductQuestion.Delete)
	})

	r.Route("/contact-messages", func(r chi.Router) {
		r.Get("/stats", c.Contact.Stats)
		r.Get("/", c.Contact.List)
		r.Get("/{id}", c.Contact.Get)
		r.Patch("/{id}/read", c.Contact.MarkRead)
		r.Patch("/{id}/archive", c.Contact.MarkArchive)
		r.Patch("/{id}/status", c.Contact.UpdateStatus)
		r.Delete("/{id}", c.Contact.Delete)
	})
}

func registerAdminBlogRoutes(r chi.Router, c *di.Container) {
	r.Route("/blog", func(r chi.Router) {
		r.Route("/categories", func(r chi.Router) {
			r.Get("/", c.Blog.AdminListCategories)
			r.Post("/", c.Blog.AdminCreateCategory)
			r.Get("/{id}", c.Blog.AdminGetCategory)
			r.Put("/{id}", c.Blog.AdminUpdateCategory)
			r.Delete("/{id}", c.Blog.AdminDeleteCategory)
		})

		r.Route("/posts", func(r chi.Router) {
			r.Get("/", c.Blog.AdminListPosts)
			r.Post("/", c.Blog.AdminCreatePost)
			r.Get("/{id}", c.Blog.AdminGetPost)
			r.Put("/{id}", c.Blog.AdminUpdatePost)
			r.Delete("/{id}", c.Blog.AdminDeletePost)
		})

		r.Route("/comments", func(r chi.Router) {
			r.Get("/", c.Blog.AdminListComments)
			r.Patch("/{id}/approve", c.Blog.AdminApproveComment)
			r.Patch("/{id}/reject", c.Blog.AdminRejectComment)
			r.Patch("/{id}/status", c.Blog.AdminModerateComment)
			r.Delete("/{id}", c.Blog.AdminDeleteComment)
		})
	})
}

func corsMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAge:           cfg.CORS.MaxAge,
	})
}
