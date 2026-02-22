// Package database handles DB connection and migrations setup for PostgreSQL.
package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Amrutavarshini24/Eventregistration/internal/models"
)

// Connect opens a PostgreSQL database connection.
func Connect() (*gorm.DB, error) {
	_ = godotenv.Load()

	cfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Info)}

	var dialector gorm.Dialector
	driver := os.Getenv("DB_DRIVER")

	if driver == "sqlite" {
		sqliteFile := os.Getenv("SQLITE_FILE")
		if sqliteFile == "" {
			sqliteFile = "event_ticketing.db"
		}
		dialector = sqlite.Open(sqliteFile)
	} else {
		dialector = postgres.Open(postgresDSN())
	}

	db, err := gorm.Open(dialector, cfg)
	if err != nil {
		return nil, fmt.Errorf("database.Connect: %w", err)
	}
	log.Printf("✅  Connected to %s database", driver)
	return db, nil
}

// Migrate auto-migrates all models.
func Migrate(db *gorm.DB) error {
	log.Println("Running migrations…")
	if err := db.AutoMigrate(&models.User{}, &models.Event{}, &models.Registration{}); err != nil {
		return fmt.Errorf("database.Migrate: %w", err)
	}
	log.Println("Migrations complete.")
	return nil
}

func postgresDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s client_encoding=UTF8",
		env("DB_HOST", "localhost"),
		env("DB_PORT", "5432"),
		env("DB_USER", "postgres"),
		env("DB_PASSWORD", "postgres"), 
		env("DB_NAME", "event_ticketing"),
		env("DB_SSLMODE", "disable"),
	)
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
