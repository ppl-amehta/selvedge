package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

// RunMigrations executes SQL schema scripts in order
func (d *DB) RunMigrations(migrationSql string) error {
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
			return fmt.Errorf("migration statement failed: %s: %w", q, err)
		}
	}

	return tx.Commit()
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
