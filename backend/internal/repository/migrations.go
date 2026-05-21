package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func RunMigrations(ctx context.Context, db *sql.DB, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		files = append(files, filepath.Join(migrationsDir, entry.Name()))
	}
	sort.Strings(files)

	for _, file := range files {
		query, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filepath.Base(file), err)
		}

		if _, err := db.ExecContext(ctx, string(query)); err != nil {
			return fmt.Errorf("run migration %s: %w", filepath.Base(file), err)
		}
	}

	return nil
}
