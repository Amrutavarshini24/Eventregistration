package main

import (
	"log"

	"github.com/Amrutavarshini24/Eventregistration/cmd/server"
	"github.com/Amrutavarshini24/Eventregistration/internal/database"
)

func main() {
	// Initialize database connection (reads DB_DRIVER, DB_HOST, etc. from .env)
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Run GORM auto-migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Start HTTP server
	srv := server.New(db)
	if err := srv.Run(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
