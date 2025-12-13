package test

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DbmateMigration represents a parsed dbmate migration file
type DbmateMigration struct {
	Version string
	Name    string
	UpSQL   string
	DownSQL string
}

// RunDbmateMigrations runs dbmate format migrations from a directory
// This allows tests to use the same migration files as production
// replacements is a map of strings to replace in the SQL before execution (e.g., database names)
func RunDbmateMigrations(ctx context.Context, dbURL string, migrationsDir string, replacements map[string]string) error {
	// Connect to database
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	// Create schema_migrations table (dbmate uses this)
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	log.Printf("Looking for migrations in: %s", migrationsDir)

	// Read migration files
	migrations, err := readDbmateMigrations(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	log.Printf("Found %d migration files in %s", len(migrations), migrationsDir)

	// Get applied migrations
	appliedVersions, err := getAppliedMigrations(ctx, pool)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	log.Printf("Already applied: %d migrations", len(appliedVersions))

	// Apply pending migrations
	for _, migration := range migrations {
		if _, applied := appliedVersions[migration.Version]; applied {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Applying migration %s: %s", migration.Version, migration.Name)

		// Execute migration in a transaction
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Apply replacements to the SQL
		upSQL := migration.UpSQL
		for old, new := range replacements {
			upSQL = strings.ReplaceAll(upSQL, old, new)
		}

		// Execute the up migration
		if _, err := tx.Exec(ctx, upSQL); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to execute migration %s: %w", migration.Version, err)
		}

		// Record migration in schema_migrations
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", migration.Version); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("failed to record migration %s: %w", migration.Version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migration.Version, err)
		}

		log.Printf("Migration %s applied successfully", migration.Version)
	}

	log.Println("All migrations completed successfully")
	return nil
}

// readDbmateMigrations reads and parses dbmate format migration files
func readDbmateMigrations(dir string) ([]DbmateMigration, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, fmt.Errorf("failed to list migration files: %w", err)
	}

	var migrations []DbmateMigration
	for _, file := range files {
		migration, err := parseDbmateMigration(file)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", file, err)
		}
		migrations = append(migrations, migration)
	}

	// Sort by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// parseDbmateMigration parses a single dbmate migration file
func parseDbmateMigration(filePath string) (DbmateMigration, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return DbmateMigration{}, fmt.Errorf("failed to read file: %w", err)
	}

	// Extract version and name from filename
	// Format: YYYYMMDDHHMMSS_description.sql
	filename := filepath.Base(filePath)
	parts := strings.SplitN(strings.TrimSuffix(filename, ".sql"), "_", 2)
	if len(parts) != 2 {
		return DbmateMigration{}, fmt.Errorf("invalid migration filename format: %s", filename)
	}

	version := parts[0]
	name := parts[1]

	// Split content by migrate markers
	text := string(content)
	upMarker := "-- migrate:up"
	downMarker := "-- migrate:down"

	upIdx := strings.Index(text, upMarker)
	downIdx := strings.Index(text, downMarker)

	if upIdx == -1 {
		return DbmateMigration{}, fmt.Errorf("missing '-- migrate:up' marker in %s", filename)
	}
	if downIdx == -1 {
		return DbmateMigration{}, fmt.Errorf("missing '-- migrate:down' marker in %s", filename)
	}

	// Extract SQL sections
	upSQL := strings.TrimSpace(text[upIdx+len(upMarker) : downIdx])
	downSQL := strings.TrimSpace(text[downIdx+len(downMarker):])

	return DbmateMigration{
		Version: version,
		Name:    name,
		UpSQL:   upSQL,
		DownSQL: downSQL,
	}, nil
}

// getAppliedMigrations returns a map of applied migration versions
func getAppliedMigrations(ctx context.Context, pool *pgxpool.Pool) (map[string]bool, error) {
	rows, err := pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return applied, nil
}

// MustRunDbmateMigrations runs dbmate migrations or panics
func MustRunDbmateMigrations(migrationsDir string, dbURL string, replacements map[string]string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("Running dbmate migrations from %s...", migrationsDir)
	if err := RunDbmateMigrations(ctx, dbURL, migrationsDir, replacements); err != nil {
		log.Fatalf("Failed to run dbmate migrations: %v", err)
	}
	log.Println("Dbmate migrations completed successfully")
}
