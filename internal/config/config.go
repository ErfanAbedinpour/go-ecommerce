package config

import (
	"fmt"
	"net/url"
	"strings"
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
	Payment  PaymentConfig
	Order    OrderConfig
	Swagger  SwaggerConfig
}

// AppConfig holds application-level settings.
type AppConfig struct {
	Name        string `env:"APP_NAME" envDefault:"ecommerce-api"`
	Environment string `env:"APP_ENV" envDefault:"development"`
	Version     string `env:"APP_VERSION" envDefault:"1.0.0"`
	Locale      string `env:"APP_LOCALE" envDefault:"en"`
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
	URL             string        `env:"DATABASE_URL" envDefault:"postgres://ecommerce:ecommerce@localhost:5432/ecommerce?sslmode=disable"`
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
	AllowedOrigins   []string `env:"CORS_ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:5173,http://localhost:5174,http://127.0.0.1:5173,https://store-os-eta.vercel.app,https://shop-panel-react.vercel.app"`
	AllowedMethods   []string `env:"CORS_ALLOWED_METHODS" envDefault:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	AllowedHeaders   []string `env:"CORS_ALLOWED_HEADERS" envDefault:"Accept,Authorization,Content-Type,X-Request-ID"`
	AllowCredentials bool     `env:"CORS_ALLOW_CREDENTIALS" envDefault:"true"`
	MaxAge           int      `env:"CORS_MAX_AGE" envDefault:"300"`
}

// UploadConfig holds file upload settings.
type UploadConfig struct {
	Provider     string   `env:"UPLOAD_PROVIDER" envDefault:"local"` // "local" or "s3"
	Dir          string   `env:"UPLOAD_DIR" envDefault:"./uploads"`
	MaxSizeMB    int      `env:"UPLOAD_MAX_SIZE_MB" envDefault:"5"`
	BaseURL      string   `env:"UPLOAD_BASE_URL" envDefault:"http://localhost:8080/uploads"`
	AllowedTypes []string `env:"UPLOAD_ALLOWED_TYPES" envDefault:"image/jpeg,image/png,image/webp,image/gif,video/mp4"`
	
	// S3 specific configs
	S3Bucket     string   `env:"UPLOAD_S3_BUCKET" envDefault:""`
	S3Region     string   `env:"UPLOAD_S3_REGION" envDefault:"us-east-1"`
	S3Endpoint   string   `env:"UPLOAD_S3_ENDPOINT" envDefault:""` // For minio or custom endpoint
	S3AccessKey  string   `env:"UPLOAD_S3_ACCESS_KEY" envDefault:""`
	S3SecretKey  string   `env:"UPLOAD_S3_SECRET_KEY" envDefault:""`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `env:"LOG_LEVEL" envDefault:"info"`
	Format string `env:"LOG_FORMAT" envDefault:"json"`
}

// PaymentConfig holds payment gateway settings.
type PaymentConfig struct {
	CallbackSecret string `env:"PAYMENT_CALLBACK_SECRET" envDefault:""`
}

// OrderConfig holds order lifecycle settings.
type OrderConfig struct {
	PaymentTTL time.Duration `env:"ORDER_PAYMENT_TTL" envDefault:"25h"`
}

// SwaggerConfig holds Swagger UI access credentials for production.
type SwaggerConfig struct {
	Username string `env:"SWAGGER_USERNAME"`
	Password string `env:"SWAGGER_PASSWORD"`
}

// Load reads configuration from environment variables.
// It optionally loads a .env file if present (ignored in production).
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.Database.MigrationsPath = normalizeMigrationsPath(cfg.Database.MigrationsPath)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks configuration for production safety.
func (c *Config) Validate() error {
	if c.IsProduction() {
		if c.Swagger.Username == "" || c.Swagger.Password == "" {
			return fmt.Errorf("SWAGGER_USERNAME and SWAGGER_PASSWORD are required when APP_ENV=production")
		}
	}
	return nil
}

func normalizeMigrationsPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "file://migrations"
	}
	if strings.Contains(path, "://") {
		return path
	}
	path = strings.TrimPrefix(path, "./")
	return "file://" + path
}

// Addr returns the server listen address.
func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return d.URL
}

// DatabaseName extracts the database name from the connection URL.
func (d DatabaseConfig) DatabaseName() (string, error) {
	if strings.Contains(d.URL, "://") {
		parsed, err := url.Parse(d.URL)
		if err != nil {
			return "", fmt.Errorf("parse database url: %w", err)
		}
		name := strings.TrimPrefix(parsed.Path, "/")
		if name == "" {
			return "", fmt.Errorf("database name not found in DATABASE_URL")
		}
		return name, nil
	}

	for _, part := range strings.Fields(d.URL) {
		if strings.HasPrefix(part, "dbname=") {
			return strings.TrimPrefix(part, "dbname="), nil
		}
	}

	return "", fmt.Errorf("database name not found in DATABASE_URL")
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}
