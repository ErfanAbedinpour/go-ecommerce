//go:build e2e

package e2e

import (
	"context"
	"fmt"
	"log/slog"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"app/internal/config"
	"app/internal/di"
	"app/internal/infrastructure/persistence/postgres"
	apphttp "app/internal/interfaces/http"
)

var (
	testServer *httptest.Server
	testClient *client
	testRunID  int64
)

func TestMain(m *testing.M) {
	if os.Getenv("E2E_SKIP") == "1" {
		fmt.Fprintln(os.Stderr, "e2e: skipped (E2E_SKIP=1)")
		os.Exit(0)
	}

	if err := chdirToProjectRoot(); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: setup failed: %v\n", err)
		exitOnMissingDeps(1)
	}

	testRunID = time.Now().UnixNano()

	cfg, cleanup, err := setupTestEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: %v\n", err)
		exitOnMissingDeps(0)
	}
	defer cleanup()

	if err := postgres.RunMigrations(cfg.Database, slog.Default()); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: migrations failed: %v\n", err)
		os.Exit(1)
	}

	container, err := di.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "e2e: container init failed: %v\n", err)
		os.Exit(1)
	}
	defer container.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := container.DB.Ping(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "e2e: database ping failed: %v\n", err)
		exitOnMissingDeps(0)
	}

	testServer = httptest.NewServer(apphttp.NewRouter(container))
	defer testServer.Close()

	testClient = newClient(testServer.URL)

	os.Exit(m.Run())
}

func chdirToProjectRoot() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return os.Chdir(wd)
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return fmt.Errorf("go.mod not found from %s", wd)
		}
		wd = parent
	}
}

func setupTestEnv() (*config.Config, func(), error) {
	uploadDir, err := os.MkdirTemp("", "e2e-uploads-*")
	if err != nil {
		return nil, nil, fmt.Errorf("create upload dir: %w", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(uploadDir)
	}

	setDefaultEnv("APP_ENV", "development")
	setDefaultEnv("DATABASE_URL", "postgres://ecommerce:ecommerce@localhost:5432/ecommerce?sslmode=disable")
	setDefaultEnv("DB_MIGRATIONS_PATH", "file://migrations")
	setDefaultEnv("JWT_SECRET", "dev-secret-change-in-production-min-32-chars")
	setDefaultEnv("AUTH_SIGNUP_ENABLED", "true")
	setDefaultEnv("SMTP_ENABLED", "false")
	setDefaultEnv("UPLOAD_PROVIDER", "local")
	setDefaultEnv("UPLOAD_DIR", uploadDir)
	setDefaultEnv("UPLOAD_BASE_URL", "http://localhost:8080/uploads")

	cfg, err := config.Load()
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("load config: %w", err)
	}

	return cfg, cleanup, nil
}

func setDefaultEnv(key, value string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}

func exitOnMissingDeps(code int) {
	if os.Getenv("E2E_REQUIRED") == "1" {
		os.Exit(1)
	}
	os.Exit(code)
}

func adminClient(t *testing.T) *client {
	t.Helper()
	token := testClient.loginAdmin(t)
	return testClient.withToken(token)
}
