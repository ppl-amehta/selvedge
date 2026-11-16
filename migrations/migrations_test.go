package migrations_test

import (
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/ppl-amehta/selvedge/migrations"
)

var forbiddenStmtRegex = regexp.MustCompile(`(?i)^\s*(PRAGMA|VACUUM|ATTACH|BEGIN|COMMIT)\b`)

func TestEmbeddedMigrations(t *testing.T) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		t.Fatalf("failed to read embedded migrations: %v", err)
	}

	if len(entries) < 8 {
		t.Errorf("expected at least 8 embedded migration files, got %d", len(entries))
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			t.Errorf("unexpected non-sql file in embedded migrations: %s", name)
		}
		names = append(names, name)

		content, err := migrations.FS.ReadFile(name)
		if err != nil {
			t.Fatalf("failed to read embedded file %s: %v", name, err)
		}
		if len(content) == 0 {
			t.Errorf("migration file %s is empty", name)
		}

		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "--") || trimmed == "" {
				continue
			}
			if forbiddenStmtRegex.MatchString(trimmed) {
				t.Errorf("migration %s:%d contains forbidden transaction statement: %s", name, lineNum+1, trimmed)
			}
		}
	}

	if !sort.StringsAreSorted(names) {
		t.Errorf("migration files are not lexically sorted: %v", names)
	}
}
