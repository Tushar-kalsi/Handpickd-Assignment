package repository

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"runtime"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/tushar-kalsi/product-views/internal/config"
)

// DB represents the database connection
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection
func NewDB(cfg *config.Config) (*DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return &DB{conn: db}, nil
}

// GetDB returns the database connection
func (db *DB) GetDB() *sql.DB {
	return db.conn
}

// Close closes the database connection
func (db *DB) Close() {
	if db.conn != nil {
		db.conn.Close()
	}
}

// RunMigrations runs database migrations
func (db *DB) RunMigrations() error {
	// Get the directory of the current file
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	migrationsPath := filepath.Join(dir, "..", "..", "migrations")

	// Set the dialect
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Run migrations
	if err := goose.Up(db.conn, migrationsPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// GetConn returns the underlying sql.DB connection
// This is useful for passing to repositories
func (db *DB) GetConn() *sql.DB {
	return db.conn
}
