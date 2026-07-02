package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	os.Clearenv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.App.Name != "ecommerce-api" {
		t.Errorf("App.Name = %q, want %q", cfg.App.Name, "ecommerce-api")
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 8080)
	}
	wantURL := "postgres://ecommerce:ecommerce@localhost:5432/ecommerce?sslmode=disable"
	if cfg.Database.URL != wantURL {
		t.Errorf("Database.URL = %q, want %q", cfg.Database.URL, wantURL)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	os.Clearenv()
	t.Setenv("APP_ENV", "production")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://ecommerce:ecommerce@db.example.com:5432/ecommerce?sslmode=disable")
	t.Setenv("SWAGGER_USERNAME", "admin")
	t.Setenv("SWAGGER_PASSWORD", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.IsProduction() {
		t.Error("expected IsProduction() to be true")
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("Server.Port = %d, want %d", cfg.Server.Port, 9090)
	}
	wantURL := "postgres://ecommerce:ecommerce@db.example.com:5432/ecommerce?sslmode=disable"
	if cfg.Database.URL != wantURL {
		t.Errorf("Database.URL = %q, want %q", cfg.Database.URL, wantURL)
	}
}

func TestServerConfig_Addr(t *testing.T) {
	cfg := ServerConfig{Host: "0.0.0.0", Port: 8080}
	if got := cfg.Addr(); got != "0.0.0.0:8080" {
		t.Errorf("Addr() = %q, want %q", got, "0.0.0.0:8080")
	}
}

func TestDatabaseConfig_DSN(t *testing.T) {
	cfg := DatabaseConfig{
		URL: "postgres://user:pass@localhost:5432/db?sslmode=disable",
	}
	if got := cfg.DSN(); got != cfg.URL {
		t.Errorf("DSN() = %q, want %q", got, cfg.URL)
	}
}

func TestDatabaseConfig_DatabaseName(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "postgres url",
			url:  "postgres://user:pass@localhost:5432/ecommerce?sslmode=disable",
			want: "ecommerce",
		},
		{
			name: "libpq dsn",
			url:  "host=localhost port=5432 user=user password=pass dbname=ecommerce sslmode=disable",
			want: "ecommerce",
		},
		{
			name:    "missing database name",
			url:     "postgres://user:pass@localhost:5432/?sslmode=disable",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DatabaseConfig{URL: tt.url}
			got, err := cfg.DatabaseName()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("DatabaseName() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("DatabaseName() = %q, want %q", got, tt.want)
			}
		})
	}
}
