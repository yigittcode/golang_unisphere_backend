package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// FileRepository handles database operations for files
type FileRepository struct {
	db *pgxpool.Pool
}

// NewFileRepository creates a new FileRepository
func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{db: db}
}

// GetByID retrieves a file by ID
func (r *FileRepository) GetByID(ctx context.Context, id int64) (*models.File, error) {
	query := `
		SELECT id, file_name, file_url, file_size, file_type, uploaded_by, created_at
		FROM files
		WHERE id = $1
	`

	var file models.File
	err := r.db.QueryRow(ctx, query, id).Scan(
		&file.ID,
		&file.FileName,
		&file.FileURL,
		&file.FileSize,
		&file.FileType,
		&file.UploadedBy,
		&file.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error getting file: %w", err)
	}

	return &file, nil
}

// Create creates a new file
func (r *FileRepository) Create(ctx context.Context, file *models.File) (int64, error) {
	query := `
		INSERT INTO files (file_name, file_url, file_size, file_type, uploaded_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query,
		file.FileName,
		file.FileURL,
		file.FileSize,
		file.FileType,
		file.UploadedBy,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error creating file: %w", err)
	}

	return id, nil
}

// Update updates an existing file
func (r *FileRepository) Update(ctx context.Context, file *models.File) error {
	query := `
		UPDATE files
		SET file_name = $1, file_url = $2, file_size = $3, file_type = $4
		WHERE id = $5
	`

	result, err := r.db.Exec(ctx, query,
		file.FileName,
		file.FileURL,
		file.FileSize,
		file.FileType,
		file.ID,
	)

	if err != nil {
		return fmt.Errorf("error updating file: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrResourceNotFound
	}

	return nil
}

// Delete deletes a file
func (r *FileRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM files WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting file: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrResourceNotFound
	}

	return nil
}
