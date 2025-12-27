package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SQLiteConfig holds SQLite configuration
type SQLiteConfig struct {
	Path string // Path to SQLite database file
}

// ConnectSQLite connects to a SQLite database
func ConnectSQLite(config SQLiteConfig) error {
	// Ensure directory exists
	dir := filepath.Dir(config.Path)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// SQLite connection with WAL mode for better concurrency
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_cache_size=10000&_foreign_keys=ON", config.Path)

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		return fmt.Errorf("failed to connect to SQLite: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Configure connection pool (SQLite is single-writer, so keep it limited)
	sqlDB.SetMaxOpenConns(1)  // SQLite only supports one writer
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Printf("Connected to SQLite database: %s", config.Path)
	return nil
}

// MigrateSQLite runs migrations for SQLite
func MigrateSQLite(models ...interface{}) error {
	if DB == nil {
		return fmt.Errorf("database not connected")
	}

	log.Println("Running SQLite migrations...")

	if err := DB.AutoMigrate(models...); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	log.Println("SQLite migrations completed successfully")
	return nil
}
