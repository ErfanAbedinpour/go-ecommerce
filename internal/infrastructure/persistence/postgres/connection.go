package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"app/internal/config"
)

// DB wraps a GORM database connection with health check support.
type DB struct {
	*gorm.DB
	sqlDB *sql.DB
	log   *slog.Logger
}

// New creates a new PostgreSQL connection via GORM.
func New(cfg config.DatabaseConfig, log *slog.Logger, isDev bool) (*DB, error) {
	logLevel := gormlogger.Silent
	if isDev {
		logLevel = gormlogger.Info
	}

	gormCfg := &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(gormpostgres.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	dbName, err := cfg.DatabaseName()
	if err != nil {
		return nil, fmt.Errorf("database name: %w", err)
	}

	log.Info("database connected", slog.String("database", dbName))

	return &DB{DB: db, sqlDB: sqlDB, log: log}, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	if d.sqlDB != nil {
		return d.sqlDB.Close()
	}
	return nil
}

// Ping checks database connectivity.
func (d *DB) Ping(ctx context.Context) error {
	return d.sqlDB.PingContext(ctx)
}

// RunMigrations applies pending database migrations.
func RunMigrations(cfg config.DatabaseConfig, log *slog.Logger) error {
	sqlDB, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return fmt.Errorf("open database for migration: %w", err)
	}
	defer sqlDB.Close()

	driver, err := migratepostgres.WithInstance(sqlDB, &migratepostgres.Config{})
	if err != nil {
		return fmt.Errorf("create migration driver: %w", err)
	}

	dbName, err := cfg.DatabaseName()
	if err != nil {
		return fmt.Errorf("database name: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(cfg.MigrationsPath, dbName, driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		if strings.Contains(err.Error(), "no migration found for version") {
			return fmt.Errorf(
				"run migrations: %w (database migration version is out of sync with this codebase; run: make db-reset)",
				err,
			)
		}
		return fmt.Errorf("run migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	log.Info("migrations applied",
		slog.Uint64("version", uint64(version)),
		slog.Bool("dirty", dirty),
	)

	return nil
}
