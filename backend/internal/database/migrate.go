package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations menjalankan semua file SQL di folder migrations secara berurutan.
// Menggunakan tabel schema_migrations untuk tracking yang sudah dijalankan.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	// Buat tabel tracking jika belum ada
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	// Baca semua file migration dari direktori
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		// Jika direktori tidak ada, skip (mungkin belum setup)
		fmt.Printf("Warning: migrations dir %s not found, skipping\n", migrationsDir)
		return nil
	}

	// Sort berdasarkan nama file
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := entry.Name()

		// Cek apakah sudah dijalankan
		var count int
		err := pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM schema_migrations WHERE version = $1",
			version,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if count > 0 {
			continue // Sudah dijalankan
		}

		// Baca dan jalankan SQL
		content, err := os.ReadFile(filepath.Join(migrationsDir, version))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("run migration %s: %w", version, err)
		}

		// Catat sebagai sudah dijalankan
		if _, err := pool.Exec(ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			version,
		); err != nil {
			return fmt.Errorf("record migration %s: %w", version, err)
		}

		fmt.Printf("Migration applied: %s\n", version)
	}

	return nil
}
