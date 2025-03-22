package storage

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // SQLite3 driver
)

var DB *sql.DB

// InitDB initializes the database connection
func InitDB() {
	// Get database path from environment variable or use default
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		// If not set, use a default path in the current directory
		dbPath = filepath.Join(".", "finance.db")
	}

	// Make sure the directory exists
	dir := filepath.Dir(dbPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create database directory: %v", err)
		}
	}

	var err error
	// Open SQLite database
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Verify the connection is working
	if err = DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("SQLite database opened at %s", dbPath)

	// Initialize tables
	if err = EnsureUserTableExists(); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}

	if err = EnsureDepositsTableExists(); err != nil {
		log.Fatalf("Failed to create deposits table: %v", err)
	}

	// Enable foreign key support
	_, err = DB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatalf("Failed to enable foreign key support: %v", err)
	}
}

// GetDB returns the database instance
func GetDB() *sql.DB {
	return DB
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
