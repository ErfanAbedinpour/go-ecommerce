package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Auth     AuthConfig
	SMTP     SMTPConfig
	CORS     CORSConfig
	Upload   UploadConfig
	Log      LogConfig
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Name        string `env:"APP_NAME" envDefault:"ecommerce-api"`
	Environment string `env:"APP_ENV" envDefault:"development"`
	Version     string `env:"APP_VERSION" envDefault:"1.0.0"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host            string        `env:"SERVER_HOST" envDefault:"0.0.0.0"`
	Port            int           `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"15s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"15s"`
	IdleTimeout     time.Duration `env:"SERVER_IDLE_TIMEOUT" envDefault:"60s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"30s"`
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host            string        `env:"DB_HOST" envDefault:"localhost"`
	Port            int           `env:"DB_PORT" envDefault:"5432"`
	User            string        `env:"DB_USER" envDefault:"ecommerce"`
	Password        string        `env:"DB_PASSWORD" envDefault:"ecommerce"`
	Name            string        `env:"DB_NAME" envDefault:"ecommerce"`
	SSLMode         string        `env:"DB_SSL_MODE" envDefault:"disable"`
	MaxOpenConns    int           `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
	MaxIdleConns    int           `env:"DB_MAX_IDLE_CONNS" envDefault:"10"`
	ConnMaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"5m"`
	MigrationsPath  string        `env:"DB_MIGRATIONS_PATH" envDefault:"file://migrations"`
}

// JWTConfig holds JWT authentication settings.
type JWTConfig struct {
	Secret           string        `env:"JWT_SECRET" envDefault:"change-me-in-production-min-32-chars"`
	AccessTokenTTL   time.Duration `env:"JWT_ACCESS_TTL" envDefault:"15m"`
	RefreshTokenTTL  time.Duration `env:"JWT_REFRESH_TTL" envDefault:"168h"`
	Issuer           string        `env:"JWT_ISSUER" envDefault:"ecommerce-api"`
}

// AuthConfig holds account registration and password reset settings.
type AuthConfig struct {
	SignupEnabled     bool          `env:"AUTH_SIGNUP_ENABLED" envDefault:"true"`
	SignupDefaultRole string        `env:"AUTH_SIGNUP_DEFAULT_ROLE" envDefault:"customer"`
	ResetTokenTTL     time.Duration `env:"AUTH_RESET_TOKEN_TTL" envDefault:"1h"`
	AppURL            string        `env:"AUTH_APP_URL" envDefault:"http://localhost:5173"`
	ResetPath         string        `env:"AUTH_RESET_PATH" envDefault:"/reset-password"`
}

// SMTPConfig holds outbound email settings for transactional messages.
type SMTPConfig struct {
	Enabled  bool   `env:"SMTP_ENABLED" envDefault:"false"`
	Host     string `env:"SMTP_HOST" envDefault:"localhost"`
	Port     int    `env:"SMTP_PORT" envDefault:"587"`
	Username string `env:"SMTP_USER"`
	Password string `env:"SMTP_PASSWORD"`
	FromEmail string `env:"SMTP_FROM_EMAIL" envDefault:"noreply@shop.com"`
	FromName  string `env:"SMTP_FROM_NAME" envDefault:"Shop Admin"`
	UseTLS    bool   `env:"SMTP_TLS" envDefault:"true"`
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	AllowedOrigins   []string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:5173"`
	AllowedMethods   []string `env:"CORS_ALLOWED_METHODS" envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	AllowedHeaders   []string `env:"CORS_ALLOWED_HEADERS" envDefault:"Accept,Authorization,Content-Type,X-Request-ID"`
	AllowCredentials bool     `env:"CORS_ALLOW_CREDENTIALS" envDefault:"true"`
	MaxAge           int      `env:"CORS_MAX_AGE" envDefault:"300"`
}

// UploadConfig holds file upload settings.
type UploadConfig struct {
	Dir          string   `env:"UPLOAD_DIR" envDefault:"./uploads"`
	MaxSizeMB    int      `env:"UPLOAD_MAX_SIZE_MB" envDefault:"5"`
	BaseURL      string   `env:"UPLOAD_BASE_URL" envDefault:"http://localhost:8080/uploads"`
	AllowedTypes []string `env:"UPLOAD_ALLOWED_TYPES" envDefault:"image/jpeg,image/png,image/webp,image/gif"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// Load reads configuration from environment variables.
// It optionally loads a .env file if present (ignored in production).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// Addr returns the server listen address.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}
