package db

import (
	"context"
	"database/sql"
	"embed"
	"log/slog"
	"path"
	"strings"
	"time"

	"github.com/yz4230/arpmon/internal/set"
)

//go:embed migrations
var schema embed.FS

type Migration struct {
	ID        int
	Name      string
	CreatedAt time.Time
}

func CreateMigrationsTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)

	return err
}

func ListMigrations(ctx context.Context, tx *sql.Tx) ([]*Migration, error) {
	if err := CreateMigrationsTable(ctx, tx); err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, "SELECT id, name, created_at FROM migrations ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var migrations []*Migration
	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.ID, &m.Name, &m.CreatedAt); err != nil {
			return nil, err
		}
		migrations = append(migrations, &m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return migrations, nil
}

func CreateMigration(ctx context.Context, tx *sql.Tx, name string) error {
	if err := CreateMigrationsTable(ctx, tx); err != nil {
		return err
	}

	_, err := tx.ExecContext(ctx, "INSERT INTO migrations (name) VALUES ($1)", name)
	return err
}

func ApplyMigrations(ctx context.Context, tx *sql.Tx) error {
	appliedMigrations, err := ListMigrations(ctx, tx)
	if err != nil {
		return err
	}

	appliedMigrationNames := set.New[string]()
	for _, m := range appliedMigrations {
		appliedMigrationNames.Add(m.Name)
	}

	entries, _ := schema.ReadDir("migrations")
	slog.Info("Migration files found", "count", len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := path.Ext(name)
		if ext != ".sql" {
			continue
		}
		stem := strings.TrimSuffix(name, ext)

		if appliedMigrationNames.Contains(stem) {
			slog.Info("Migration already applied, skipping", "name", stem)
			continue
		}

		data, err := schema.ReadFile(path.Join("migrations", name))
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, string(data)); err != nil {
			return err
		}

		if err := CreateMigration(ctx, tx, stem); err != nil {
			return err
		}

		slog.Info("Migration applied successfully", "name", name)
	}

	return nil
}
