package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

// Connect initializes SQLite with industrial factory settings (WAL mode, busy timeout, foreign keys)
func Connect(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	// DSN with pragmas embedded
	dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_synchronous=NORMAL", dbPath)
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Configure connection pool with WAL concurrency support
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlite database: %w", err)
	}

	log.Printf("✓ Connected to SQLite database at: %s (WAL mode active)", dbPath)
	return &DB{db}, nil
}

// RunMigrations executes SQL schema scripts in order with schema_migrations tracking
func (d *DB) RunMigrations(migrationName string, migrationSql string) error {
	// Ensure migration tracking table exists
	_, err := d.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at INTEGER NOT NULL
	)`)
	if err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	// Check if already applied
	var count int
	err = d.DB.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", migrationName).Scan(&count)
	if err == nil && count > 0 {
		log.Printf("Migration %s already applied, skipping", migrationName)
		return nil
	}

	queries := strings.Split(migrationSql, ";")

	// 1. Run any PRAGMA statements first outside transaction
	for _, query := range queries {
		q := strings.TrimSpace(query)
		if q == "" {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(q), "PRAGMA") {
			if _, err := d.DB.Exec(q); err != nil {
				log.Printf("Warning executing pragma %s: %v", q, err)
			}
		}
	}

	tx, err := d.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, query := range queries {
		q := strings.TrimSpace(query)
		if q == "" || strings.HasPrefix(strings.ToUpper(q), "PRAGMA") {
			continue
		}
		if _, err := tx.Exec(q); err != nil {
			return fmt.Errorf("migration statement failed in %s: %s: %w", migrationName, q, err)
		}
	}

	_, err = tx.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", migrationName, timeNowMs())
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
	}

	return tx.Commit()
}

func timeNowMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// WithTransaction runs a function within an atomic transaction
func (d *DB) WithTransaction(fn func(tx *sql.Tx) error) error {
	tx, err := d.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
