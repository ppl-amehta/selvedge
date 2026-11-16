package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

func TestDatabaseConnectAndPragmas(t *testing.T) {
	testDir := filepath.Join(os.TempDir(), "test_db_platform")
	_ = os.RemoveAll(testDir)
	defer os.RemoveAll(testDir)

	dbPath := filepath.Join(testDir, "test.db")
	database, err := Connect(dbPath)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	if database.Engine == "" {
		t.Errorf("expected engine name to be set")
	}

	// Verify journal_mode is WAL
	var journalMode string
	err = database.QueryRow("PRAGMA journal_mode;").Scan(&journalMode)
	if err != nil {
		t.Fatalf("failed to query journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("expected journal_mode 'wal', got %q", journalMode)
	}

	// Verify foreign keys are enabled
	var foreignKeys int
	err = database.QueryRow("PRAGMA foreign_keys;").Scan(&foreignKeys)
	if err != nil {
		t.Fatalf("failed to query foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Errorf("expected foreign_keys 1, got %d", foreignKeys)
	}

	// Test atomic WithTransaction
	err = database.WithTransaction(func(tx *sql.Tx) error {
		_, err := tx.Exec("CREATE TABLE test_tx (id INTEGER PRIMARY KEY, val TEXT);")
		if err != nil {
			return err
		}
		_, err = tx.Exec("INSERT INTO test_tx (val) VALUES ('committed');")
		return err
	})
	if err != nil {
		t.Fatalf("WithTransaction failed: %v", err)
	}

	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM test_tx WHERE val = 'committed';").Scan(&count)
	if err != nil || count != 1 {
		t.Errorf("expected 1 row committed, got count=%d, err=%v", count, err)
	}
}
