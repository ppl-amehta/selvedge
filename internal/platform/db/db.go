package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type DB struct {
	*sql.DB
	Engine   string
	IsSQLite bool
}

// Connect initializes SQLite with industrial factory settings or LibSQL cloud when configured
func Connect(dbPath string) (*DB, error) {
	var db *sql.DB
	var err error
	var engineName string

	tursoURL := os.Getenv("TURSO_DATABASE_URL")
	if tursoURL == "" {
		tursoURL = os.Getenv("DATABASE_URL")
	}

	isSQLite := false
	if tursoURL != "" && (strings.HasPrefix(tursoURL, "libsql://") || strings.HasPrefix(tursoURL, "https://") || strings.HasPrefix(tursoURL, "http://")) {
		token := os.Getenv("TURSO_AUTH_TOKEN")
		u, parseErr := url.Parse(tursoURL)
		if parseErr != nil {
			return nil, fmt.Errorf("invalid LibSQL/Turso URL: %w", parseErr)
		}
		if token != "" {
			q := u.Query()
			q.Set("authToken", token)
			u.RawQuery = q.Encode()
		}
		dsn := u.String()
		db, err = sql.Open("libsql", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open LibSQL/Turso database: %w", err)
		}
		// Redact token from engine description and log outputs
		engineName = fmt.Sprintf("LibSQL / Turso Cloud (%s%s)", u.Host, u.Path)
	} else {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0700); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}

		// DSN with pragmas embedded (including _txlock=immediate to prevent deferred transaction deadlocks)
		dsn := fmt.Sprintf("%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_synchronous=NORMAL&_txlock=immediate", dbPath)
		db, err = sql.Open("sqlite3", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}
		engineName = fmt.Sprintf("Local SQLite WAL (%s)", dbPath)
		isSQLite = true
	}

	// Configure connection pool optimized for single-writer SQLite WAL / cloud concurrency
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(15 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database (%s): %w", engineName, err)
	}

	slog.Info("connected to database", "engine", engineName)
	return &DB{DB: db, Engine: engineName, IsSQLite: isSQLite}, nil
}

// ListAppliedMigrations returns all already-applied migration keys in one query.
func (d *DB) ListAppliedMigrations() (map[string]bool, error) {
	_, err := d.DB.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at INTEGER NOT NULL
	)`)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	rows, err := d.DB.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err == nil {
			applied[version] = true
		}
	}
	return applied, rows.Err()
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

	tx, err := d.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	baseName := filepath.Base(migrationName)
	canonicalKey := "migrations/" + baseName

	var count int
	err = tx.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ? OR version = ?", canonicalKey, baseName).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check migration status for %s: %w", migrationName, err)
	}
	if count > 0 {
		return nil
	}

	if strings.TrimSpace(migrationSql) != "" {
		if d.IsSQLite {
			if _, err := tx.Exec(migrationSql); err != nil {
				return fmt.Errorf("migration statement failed in %s: %w", migrationName, err)
			}
		} else {
			stmts := splitSQLStatements(migrationSql)
			const batchSize = 40
			for i := 0; i < len(stmts); i += batchSize {
				end := i + batchSize
				if end > len(stmts) {
					end = len(stmts)
				}
				chunk := strings.Join(stmts[i:end], ";\n") + ";"
				if _, err := tx.Exec(chunk); err != nil {
					return fmt.Errorf("migration batch failed in %s (stmt %d..%d): %w", migrationName, i+1, end, err)
				}
			}
		}
	}

	_, err = tx.Exec("INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)", canonicalKey, time.Now().UTC().UnixMilli())
	if err != nil {
		return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %s: %w", migrationName, err)
	}

	slog.Info("migration successfully applied", "version", canonicalKey)
	return nil
}

// splitSQLStatements splits a multi-statement SQL script on top-level semicolons,
// ignoring semicolons inside string literals, quoted identifiers, and comments.
func splitSQLStatements(sqlText string) []string {
	var stmts []string
	var cur strings.Builder

	inSingle := false
	inDouble := false
	inLineComment := false
	inBlockComment := false

	n := len(sqlText)
	for i := 0; i < n; i++ {
		ch := sqlText[i]

		if inLineComment {
			cur.WriteByte(ch)
			if ch == '\n' {
				inLineComment = false
			}
			continue
		}
		if inBlockComment {
			cur.WriteByte(ch)
			if ch == '*' && i+1 < n && sqlText[i+1] == '/' {
				cur.WriteByte('/')
				i++
				inBlockComment = false
			}
			continue
		}
		if inSingle {
			cur.WriteByte(ch)
			if ch == '\'' {
				if i+1 < n && sqlText[i+1] == '\'' {
					cur.WriteByte('\'')
					i++
				} else {
					inSingle = false
				}
			}
			continue
		}
		if inDouble {
			cur.WriteByte(ch)
			if ch == '"' {
				if i+1 < n && sqlText[i+1] == '"' {
					cur.WriteByte('"')
					i++
				} else {
					inDouble = false
				}
			}
			continue
		}

		if ch == '-' && i+1 < n && sqlText[i+1] == '-' {
			inLineComment = true
			cur.WriteString("--")
			i++
			continue
		}
		if ch == '/' && i+1 < n && sqlText[i+1] == '*' {
			inBlockComment = true
			cur.WriteString("/*")
			i++
			continue
		}
		if ch == '\'' {
			inSingle = true
			cur.WriteByte(ch)
			continue
		}
		if ch == '"' {
			inDouble = true
			cur.WriteByte(ch)
			continue
		}
		if ch == ';' {
			trimmed := strings.TrimSpace(cur.String())
			if trimmed != "" {
				stmts = append(stmts, trimmed)
			}
			cur.Reset()
			continue
		}
		cur.WriteByte(ch)
	}

	if trimmed := strings.TrimSpace(cur.String()); trimmed != "" {
		stmts = append(stmts, trimmed)
	}
	return stmts
}

// CheckpointWAL truncates the write-ahead log cleanly on shutdown
func (d *DB) CheckpointWAL(ctx context.Context) error {
	if d.IsSQLite {
		_, err := d.DB.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE);")
		if err != nil {
			return fmt.Errorf("wal checkpoint failed: %w", err)
		}
		slog.Info("sqlite wal checkpoint completed (TRUNCATE)")
	}
	return nil
}

// WithTransaction runs a function within an atomic transaction
func (d *DB) WithTransaction(fn func(tx *sql.Tx) error) error {
	if fn == nil {
		return fmt.Errorf("nil transaction function provided")
	}

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
