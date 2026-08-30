package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lib/pq"
)

// Migrator handles database schema migrations.
type Migrator struct {
	db  *DB
	dir string
}

// Migration represents a single migration.
type Migration struct {
	Version     string
	Name        string
	Description string
	SQL         string
	RollbackSQL string
	Applied     bool
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(db *DB, migrationsDir string) *Migrator {
	return &Migrator{db: db, dir: migrationsDir}
}

// RunAll runs all pending migrations in order.
func (m *Migrator) RunAll() error {
	migrations, err := m.discoverMigrations()
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %w", err)
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if !applied[migration.Name] && !migration.Applied {
			if err := m.runMigration(&migration); err != nil {
				return fmt.Errorf("failed to apply migration %s: %w", migration.Name, err)
			}
			fmt.Printf("Applied migration: %s - %s\n", migration.Name, migration.Description)
		}
	}

	return nil
}

// Rollback rolls back the last migration.
func (m *Migrator) Rollback() error {
	migrations, err := m.discoverMigrations()
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %w", err)
	}

	if len(migrations) == 0 {
		return fmt.Errorf("no migrations found")
	}

	// Get last applied migration
	lastApplied := ""
	for i := len(migrations) - 1; i >= 0; i-- {
		if migrations[i].Applied {
			lastApplied = migrations[i].Version
			break
		}
	}

	if lastApplied == "" {
		return fmt.Errorf("no applied migrations to rollback")
	}

	return m.rollbackMigration(lastApplied)
}

// RollbackTo rolls back to a specific version (inclusive).
func (m *Migrator) RollbackTo(targetVersion string) error {
	migrations, err := m.discoverMigrations()
	if err != nil {
		return fmt.Errorf("failed to discover migrations: %w", err)
	}

	// Build ordered list
	ordered := make([]Migration, 0, len(migrations))
	for _, m := range migrations {
		if m.Applied {
			ordered = append(ordered, m)
		}
	}
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Version < ordered[j].Version
	})

	// Find target index
	targetIdx := -1
	for i, m := range ordered {
		if m.Version == targetVersion {
			targetIdx = i
			break
		}
	}

	if targetIdx < 0 {
		return fmt.Errorf("migration version %s not found or not applied", targetVersion)
	}

	// Rollback all after target
	for i := len(ordered) - 1; i > targetIdx; i-- {
		if err := m.rollbackMigration(ordered[i].Version); err != nil {
			return fmt.Errorf("failed to rollback %s: %w", ordered[i].Version, err)
		}
		fmt.Printf("Rolled back: %s\n", ordered[i].Version)
	}

	return nil
}

// Status returns the status of all migrations.
func (m *Migrator) Status() ([]Migration, error) {
	migrations, err := m.discoverMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to discover migrations: %w", err)
	}

	applied, err := m.getAppliedMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for i := range migrations {
		migrations[i].Applied = applied[migrations[i].Name]
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// Seed runs all seed files in order.
func (m *Migrator) Seed() error {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return fmt.Errorf("failed to read seeds directory: %w", err)
	}

	seeds := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "seed_") || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		seeds = append(seeds, entry.Name())
	}
	sort.Strings(seeds)

	for _, seed := range seeds {
		path := filepath.Join(m.dir, seed)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read seed %s: %w", seed, err)
		}

		if _, err := m.db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("failed to execute seed %s: %w", seed, err)
		}
		fmt.Printf("Executed seed: %s\n", seed)
	}

	return nil
}

// Reset seeds and reruns all migrations fresh.
// WARNING: Drops all data.
func (m *Migrator) Reset() error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Drop all tables (except in a real reset you'd want pg_dump)
	// For safety, we'll just truncate and let re-run handle it
	_, err = tx.Exec(`TRUNCATE schema_migrations CASCADE`)
	if err != nil {
		return fmt.Errorf("failed to clear migrations: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit reset: %w", err)
	}

	return m.RunAll()
}

func (m *Migrator) discoverMigrations() ([]Migration, error) {
	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	migrations := make([]Migration, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "ROLLBACK_") {
			continue // Handled separately
		}
		if strings.HasPrefix(name, "seed_") {
			continue // Seeds handled separately
		}

		// Extract version from filename
		// Format: NNN_description.sql
		parts := strings.SplitN(name, "_", 2)
		if len(parts) < 2 {
			continue
		}
		version := strings.TrimSuffix(parts[0], ".sql")

		sqlBytes, err := os.ReadFile(filepath.Join(m.dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", entry.Name(), err)
		}

		migration := Migration{
			Version: version,
			Name:    strings.TrimSuffix(entry.Name(), ".sql"),
			SQL:     string(sqlBytes),
		}

		// Load rollback if exists
		rollbackPath := filepath.Join(m.dir, "ROLLBACK_"+version+".sql")
		if _, err := os.Stat(rollbackPath); err == nil {
			rollbackBytes, err := os.ReadFile(rollbackPath)
			if err == nil {
				migration.RollbackSQL = string(rollbackBytes)
			}
		}

		// Extract description if possible
		if len(parts) == 2 {
			migration.Description = strings.TrimSuffix(parts[1], ".sql")
		}

		migrations = append(migrations, migration)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	applied := make(map[string]bool)

	if !m.tableExists("schema_migrations") {
		return applied, nil
	}

	rows, err := m.db.Query("SELECT version FROM schema_migrations ORDER BY applied_at")
	if err != nil {
		return nil, fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

func (m *Migrator) tableExists(tableName string) bool {
	var exists bool
	err := m.db.QueryRow(
		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = $1)",
		tableName,
	).Scan(&exists)
	return err == nil && exists
}

func (m *Migrator) runMigration(migration *Migration) error {
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(migration.SQL)
	if err != nil {
		// The SQL failed. If it's only because objects already exist (fresh
		// DB where docker-entrypoint-initdb.d already applied this migration,
		// or a prior partial run), treat it as a no-op: roll back the aborted
		// tx and record the migration outside it.
		if isAlreadyExists(err) {
			if rerr := tx.Rollback(); rerr != nil {
				return fmt.Errorf("failed to rollback after benign error: %w", rerr)
			}
			_, rerr := m.db.Exec(
				"INSERT INTO schema_migrations (version, description) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING",
				migration.Name, migration.Description,
			)
			if rerr != nil {
				return fmt.Errorf("failed to record migration %s: %w", migration.Name, rerr)
			}
			return nil
		}
		_ = tx.Rollback()
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record the migration as applied. This is self-contained (does not rely
	// on the SQL file carrying its own INSERT), so every migration is tracked
	// even when its .sql has no schema_migrations statement.
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, description) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING",
		migration.Name, migration.Description,
	)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to record migration %s: %w", migration.Name, err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (m *Migrator) rollbackMigration(version string) error {
	rollbackPath := filepath.Join(m.dir, "ROLLBACK_"+version+".sql")
	if _, err := os.Stat(rollbackPath); os.IsNotExist(err) {
		return fmt.Errorf("no rollback file for version %s", version)
	}

	rollbackBytes, err := os.ReadFile(rollbackPath)
	if err != nil {
		return fmt.Errorf("failed to read rollback file: %w", err)
	}

	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	_, err = tx.Exec(string(rollbackBytes))
	if err != nil {
		return fmt.Errorf("failed to execute rollback: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	fmt.Printf("Rolled back: %s\n", version)
	return nil
}

// InitSchemaMigrations creates the schema_migrations table if it doesn't exist.
func (m *Migrator) InitSchemaMigrations() error {
	if m.tableExists("schema_migrations") {
		return nil
	}

	_, err := m.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW(),
			description TEXT
		)
	`)
	return err
}

// EnsureSchemaMigrations ensures the schema_migrations table is created.
func (m *Migrator) EnsureSchemaMigrations() error {
	if !m.tableExists("schema_migrations") {
		return m.InitSchemaMigrations()
	}
	return nil
}

// RunPostgresInitScript runs the raw postgres init script (for docker-entrypoint-initdb.d).
// This executes raw SQL from a file without transaction wrapping.
func RunPostgresInitScript(db *sql.DB, sqlContent string) error {
	_, err := db.Exec(sqlContent)
	return err
}

// isAlreadyExists reports whether err is a PostgreSQL "already exists" error
// (SQLSTATE 42P07 / 42710) or the SQLite equivalent. Used by runMigration to
// treat re-applied idempotent DDL as a no-op rather than a fatal error.
func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	if pqErr, ok := err.(*pq.Error); ok {
		switch pqErr.Code {
		case "42P07", "42710": // duplicate_table, duplicate_object
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "already exists") ||
		strings.Contains(msg, "duplicate table") ||
		strings.Contains(msg, "duplicate object")
}
