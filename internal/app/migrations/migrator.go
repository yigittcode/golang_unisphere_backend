package migrations

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

// Migrator manages database migrations
type Migrator struct {
	db *pgxpool.Pool
}

// NewMigrator creates a new migrator
func NewMigrator(db *pgxpool.Pool) *Migrator {
	return &Migrator{
		db: db,
	}
}

// ensureMigrationTableExists creates the migration tracking table if it doesn't exist
func (m *Migrator) ensureMigrationTableExists(ctx context.Context) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := m.db.Exec(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create migration tracking table: %w", err)
	}
	return nil
}

// isMigrationApplied checks if a specific migration has already been applied
func (m *Migrator) isMigrationApplied(ctx context.Context, version string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1);`
	err := m.db.QueryRow(ctx, query, version).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check migration status: %w", err)
	}
	return exists, nil
}

// recordMigration marks a migration as applied
func (m *Migrator) recordMigration(ctx context.Context, version string) error {
	_, err := m.db.Exec(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)`,
		version, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}
	return nil
}

// MigrateFromFile executes SQL statements from a file
func (m *Migrator) MigrateFromFile(filePath string) error {
	ctx := context.Background()

	// Ensure migration tracking table exists
	if err := m.ensureMigrationTableExists(ctx); err != nil {
		return err
	}

	// Extract version from filename (e.g., "001_init.sql" => "001")
	filename := filepath.Base(filePath)
	version := strings.Split(filename, "_")[0]

	// Check if migration was already applied
	migrationApplied, err := m.isMigrationApplied(ctx, version)
	if err != nil {
		return err
	}

	if migrationApplied {
		log.Printf("Migration %s already applied, skipping", filename)
		return nil
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	log.Printf("Reading migration file: %s", filePath)

	// Start a transaction for the migration
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Execute the migration
	_, err = tx.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("error occurred during SQL migration execution: %w", err)
	}

	// Record the migration as applied
	if err := m.recordMigration(ctx, version); err != nil {
		return err
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Migration file successfully applied: %s", filePath)
	return nil
}

// MigrateFromDirectory finds and executes all SQL files in a directory
func (m *Migrator) MigrateFromDirectory(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read migration directory: %w", err)
	}

	// Collect SQL files
	var sqlFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			sqlFiles = append(sqlFiles, file.Name())
		}
	}

	// Sort files to ensure they're executed in order
	sort.Strings(sqlFiles)

	// Apply migrations in order
	for _, file := range sqlFiles {
		filePath := filepath.Join(dirPath, file)
		if err := m.MigrateFromFile(filePath); err != nil {
			return err
		}
	}

	return nil
}
