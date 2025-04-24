package migrations

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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

// MigrateFromFile executes SQL statements from a file
func (m *Migrator) MigrateFromFile(filePath string) error {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	log.Printf("Reading migration file: %s", filePath)

	// Execute the entire SQL file as a single query
	ctx := context.Background()
	_, err = m.db.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("error occurred during SQL migration execution: %w", err)
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

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			filePath := filepath.Join(dirPath, file.Name())
			if err := m.MigrateFromFile(filePath); err != nil {
				return err
			}
		}
	}

	return nil
}
