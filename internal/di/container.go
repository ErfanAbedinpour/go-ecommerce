package di

import (
	"log/slog"
	"os"

	appauth "app/internal/application/auth"
	appcategory "app/internal/application/category"
	appcoupon "app/internal/application/coupon"
	appcustomer "app/internal/application/customer"
	appdashboard "app/internal/application/dashboard"
	apporder "app/internal/application/order"
	appproduct "app/internal/application/product"
	appsettings "app/internal/application/settings"
	"app/internal/config"
	infraauth "app/internal/infrastructure/auth"
	"app/internal/infrastructure/email"
	"app/internal/infrastructure/persistence/postgres"
	"app/internal/interfaces/http/handler"
	"app/pkg/validator"
)

// Container holds all application dependencies wired together.
type Container struct {
	Config    *config.Config
	Log       *slog.Logger
	DB        *postgres.DB
	Validator *validator.Validator

	// Infrastructure
	JWT    *infraauth.JWTService
	Hasher *infraauth.PasswordHasher

	// Services
	AuthService     *appauth.AuthService
	ProductService  *appproduct.Service
	CategoryService *appcategory.Service
	CouponService    *appcoupon.Service
	CustomerService  *appcustomer.Service
	DashboardService *appdashboard.Service
	OrderService     *apporder.Service
	SettingsService  *appsettings.Service

	// Handlers
	Health   *handler.HealthHandler
	Auth     *handler.AuthHandler
	Product  *handler.ProductHandler
	Category *handler.CategoryHandler
	Coupon   *handler.CouponHandler
	Customer  *handler.CustomerHandler
	Dashboard *handler.DashboardHandler
	Order     *handler.OrderHandler
	Settings  *handler.SettingsHandler
}

// New creates and wires the dependency injection container.
func New(cfg *config.Config) (*Container, error) {
	log := newLogger(cfg)

	db, err := postgres.New(cfg.Database, log, cfg.IsDevelopment())
	if err != nil {
		return nil, err
	}

	v := validator.New()

	jwtService := infraauth.NewJWTService(cfg.JWT)
	hasher := infraauth.NewPasswordHasher()

	userRepo := postgres.NewUserRepository(db.DB)
	refreshTokenRepo := postgres.NewRefreshTokenRepository(db.DB)
	resetTokenRepo := postgres.NewPasswordResetRepository(db.DB)
	mailer := email.NewSMTPMailer(cfg.SMTP, log)
	authService := appauth.NewAuthService(
		userRepo,
		refreshTokenRepo,
		resetTokenRepo,
		hasher,
		jwtService,
		mailer,
		cfg.Auth,
	)

	productRepo := postgres.NewProductRepository(db.DB)
	productService := appproduct.NewService(productRepo)

	categoryRepo := postgres.NewCategoryRepository(db.DB)
	categoryService := appcategory.NewService(categoryRepo)

	couponRepo := postgres.NewCouponRepository(db.DB)
	couponService := appcoupon.NewService(couponRepo)

	customerRepo := postgres.NewCustomerRepository(db.DB)
	customerService := appcustomer.NewService(customerRepo)

	dashboardRepo := postgres.NewDashboardRepository(db.DB)
	dashboardService := appdashboard.NewService(dashboardRepo)

	orderRepo := postgres.NewOrderRepository(db.DB)
	orderService := apporder.NewService(orderRepo)

	settingsRepo := postgres.NewSettingsRepository(db.DB)
	settingsService := appsettings.NewService(settingsRepo)

	c := &Container{
		Config:          cfg,
		Log:             log,
		DB:              db,
		Validator:       v,
		JWT:             jwtService,
		Hasher:          hasher,
		AuthService:     authService,
		ProductService:  productService,
		CategoryService: categoryService,
		CouponService:    couponService,
		CustomerService:  customerService,
		DashboardService: dashboardService,
		OrderService:     orderService,
		SettingsService:  settingsService,
		Health:           handler.NewHealthHandler(db, cfg.App.Version),
		Auth:            handler.NewAuthHandler(authService, v, log),
		Product:         handler.NewProductHandler(productService, v, log),
		Category:        handler.NewCategoryHandler(categoryService, v, log),
		Coupon:          handler.NewCouponHandler(couponService, v, log),
		Customer:        handler.NewCustomerHandler(customerService, v, log),
		Dashboard:       handler.NewDashboardHandler(dashboardService, v, log),
		Order:           handler.NewOrderHandler(orderService, v, log),
		Settings:        handler.NewSettingsHandler(settingsService, v, log),
	}

	return c, nil
}

// Close releases all container resources.
func (c *Container) Close() error {
	return c.DB.Close()
}

func newLogger(cfg *config.Config) *slog.Logger {
	level := slog.LevelInfo
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if cfg.Log.Format == "text" || cfg.IsDevelopment() {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
