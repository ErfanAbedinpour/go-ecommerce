package di

import (
	"log/slog"
	"os"

	appauth "app/internal/application/auth"
	"app/internal/config"
	infraauth "app/internal/infrastructure/auth"
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
	AuthService *appauth.AuthService

	// Handlers
	Health *handler.HealthHandler
	Auth   *handler.AuthHandler
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
	authService := appauth.NewAuthService(userRepo, refreshTokenRepo, hasher, jwtService)

	c := &Container{
		Config:      cfg,
		Log:         log,
		DB:          db,
		Validator:   v,
		JWT:         jwtService,
		Hasher:      hasher,
		AuthService: authService,
		Health:      handler.NewHealthHandler(db, cfg.App.Version),
		Auth:        handler.NewAuthHandler(authService, v, log),
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
