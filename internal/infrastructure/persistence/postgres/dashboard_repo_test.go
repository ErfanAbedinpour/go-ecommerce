package postgres

import (
	"context"
	"os"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestDashboardRepository_GetStats(t *testing.T) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "host=localhost user=admin password=admin dbname=ecommerce-db port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Skipf("database not available: %v", err)
	}

	repo := NewDashboardRepository(db)
	stats, err := repo.GetStats(context.Background())
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}
	if stats.TotalProducts < 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
}
