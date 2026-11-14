package subscription

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// legacyVersion is the schema version of databases created before migrations
// were tracked: they already have the subscribers table with telegram_chat_type.
const legacyVersion = 2

type migration struct {
	version int
	sql     string
}

// Migrate applies the pending embedded migrations, each in its own transaction.
func Migrate(ctx context.Context, db *sql.DB) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY NOT NULL)`,
	); err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}

	current, err := currentVersion(ctx, db)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		if err := apply(ctx, db, m); err != nil {
			return err
		}
	}
	return nil
}

func currentVersion(ctx context.Context, db *sql.DB) (int, error) {
	var version int
	if err := db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(version), 0) FROM schema_migrations`,
	).Scan(&version); err != nil {
		return 0, fmt.Errorf("reading schema version: %w", err)
	}
	if version > 0 {
		return version, nil
	}

	var legacy int
	if err := db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'subscribers'`,
	).Scan(&legacy); err != nil {
		return 0, fmt.Errorf("detecting legacy schema: %w", err)
	}
	if legacy == 0 {
		return 0, nil
	}
	if _, err := db.ExecContext(ctx,
		`INSERT INTO schema_migrations(version) VALUES (?)`, legacyVersion,
	); err != nil {
		return 0, fmt.Errorf("recording legacy schema version: %w", err)
	}
	return legacyVersion, nil
}

func apply(ctx context.Context, db *sql.DB, m migration) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("migration %d: begin: %w", m.version, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, m.sql); err != nil {
		return fmt.Errorf("migration %d: %w", m.version, err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version) VALUES (?)`, m.version); err != nil {
		return fmt.Errorf("migration %d: recording version: %w", m.version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migration %d: commit: %w", m.version, err)
	}
	return nil
}

func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("reading migrations: %w", err)
	}

	migrations := make([]migration, 0, len(entries))
	for _, e := range entries {
		version, err := strconv.Atoi(strings.TrimSuffix(e.Name(), ".sql"))
		if err != nil {
			return nil, fmt.Errorf("invalid migration name %q: %w", e.Name(), err)
		}
		content, err := migrationsFS.ReadFile("migrations/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("reading migration %q: %w", e.Name(), err)
		}
		migrations = append(migrations, migration{version: version, sql: string(content)})
	}
	sort.Slice(migrations, func(i, j int) bool { return migrations[i].version < migrations[j].version })
	return migrations, nil
}
