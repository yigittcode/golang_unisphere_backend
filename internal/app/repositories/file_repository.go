package repositories

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// Dosya hataları
var (
	ErrFileNotFound = errors.New("file not found")
)

// FileRepository handles database operations for files
type FileRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

// NewFileRepository creates a new FileRepository
func NewFileRepository(db *pgxpool.Pool) *FileRepository {
	return &FileRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// CreateFile creates a new file record in the database
func (r *FileRepository) CreateFile(
	ctx context.Context,
	file *models.File,
) (int64, error) {
	// Build SQL query to insert a file record
	sql, args, err := r.sb.Insert("files").
		Columns("file_name", "file_path", "file_url", "file_size", "file_type", "resource_type", "resource_id", "uploaded_by").
		Values(file.FileName, file.FilePath, file.FileURL, file.FileSize, file.FileType, file.ResourceType, file.ResourceID, file.UploadedBy).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create file SQL")
		return 0, fmt.Errorf("failed to build create file query: %w", err)
	}

	// Execute query
	var fileID int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(&fileID)
	if err != nil {
		logger.Error().Err(err).Str("filename", file.FileName).Msg("Error creating file record")
		return 0, fmt.Errorf("error creating file record: %w", err)
	}

	logger.Info().Int64("fileID", fileID).Str("filename", file.FileName).Msg("File record created")
	return fileID, nil
}

// GetFileByID retrieves a file record by its ID
func (r *FileRepository) GetFileByID(ctx context.Context, fileID int64) (*models.File, error) {
	sql, args, err := r.sb.Select("id", "file_name", "file_path", "file_url", "file_size", "file_type", "resource_type", "resource_id", "uploaded_by", "created_at", "updated_at").
		From("files").
		Where(squirrel.Eq{"id": fileID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get file SQL")
		return nil, fmt.Errorf("failed to build get file query: %w", err)
	}

	var file models.File
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&file.ID,
		&file.FileName,
		&file.FilePath,
		&file.FileURL,
		&file.FileSize,
		&file.FileType,
		&file.ResourceType,
		&file.ResourceID,
		&file.UploadedBy,
		&file.CreatedAt,
		&file.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrFileNotFound
		}
		logger.Error().Err(err).Int64("fileID", fileID).Msg("Error retrieving file")
		return nil, fmt.Errorf("error retrieving file: %w", err)
	}

	return &file, nil
}

// DeleteFile deletes a file record from the database
func (r *FileRepository) DeleteFile(ctx context.Context, fileID int64) error {
	// First get the file path to potentially delete the physical file
	sql, args, err := r.sb.Select("file_path").
		From("files").
		Where(squirrel.Eq{"id": fileID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get file path SQL")
		return fmt.Errorf("failed to build get file path query: %w", err)
	}

	var filePath string
	err = r.db.QueryRow(ctx, sql, args...).Scan(&filePath)
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrFileNotFound
		}
		logger.Error().Err(err).Int64("fileID", fileID).Msg("Error retrieving file path")
		return fmt.Errorf("error retrieving file path: %w", err)
	}

	// Now delete the database record
	sql, args, err = r.sb.Delete("files").
		Where(squirrel.Eq{"id": fileID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building delete file SQL")
		return fmt.Errorf("failed to build delete file query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("fileID", fileID).Msg("Error deleting file record")
		return fmt.Errorf("error deleting file record: %w", err)
	}

	// Delete the physical file from filesystem
	baseDir := "./uploads" // This should match your storage configuration
	fullPath := filepath.Join(baseDir, filepath.Base(filePath))

	if err := os.Remove(fullPath); err != nil {
		if !os.IsNotExist(err) {
			logger.Warn().Err(err).Str("filePath", fullPath).Msg("Error removing physical file")
			// Don't return an error here, we've already deleted the database record
		}
	} else {
		logger.Info().Str("filePath", fullPath).Msg("Physical file deleted successfully")
	}

	logger.Info().Int64("fileID", fileID).Msg("File record deleted")
	return nil
}

// GetFilesByPastExamID retrieves all files associated with a past exam
func (r *FileRepository) GetFilesByPastExamID(ctx context.Context, pastExamID int64) ([]*models.File, error) {
	sql, args, err := r.sb.Select("f.id", "f.file_name", "f.file_path", "f.file_url", "f.file_size",
		"f.file_type", "f.resource_type", "f.resource_id", "f.uploaded_by", "f.created_at", "f.updated_at").
		From("files f").
		Join("past_exam_files pef ON f.id = pef.file_id").
		Where(squirrel.Eq{"pef.past_exam_id": pastExamID}).
		OrderBy("f.created_at DESC").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get past exam files SQL")
		return nil, fmt.Errorf("failed to build get past exam files query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("pastExamID", pastExamID).Msg("Error retrieving past exam files")
		return nil, fmt.Errorf("error retrieving past exam files: %w", err)
	}
	defer rows.Close()

	files := []*models.File{}
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID,
			&file.FileName,
			&file.FilePath,
			&file.FileURL,
			&file.FileSize,
			&file.FileType,
			&file.ResourceType,
			&file.ResourceID,
			&file.UploadedBy,
			&file.CreatedAt,
			&file.UpdatedAt,
		); err != nil {
			logger.Error().Err(err).Msg("Error scanning past exam file row")
			return nil, fmt.Errorf("error scanning past exam file: %w", err)
		}
		files = append(files, &file)
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating past exam files rows")
		return nil, fmt.Errorf("error iterating past exam files: %w", err)
	}

	return files, nil
}

// GetFilesByClassNoteID retrieves all files associated with a class note
func (r *FileRepository) GetFilesByClassNoteID(ctx context.Context, classNoteID int64) ([]*models.File, error) {
	sql, args, err := r.sb.Select("f.id", "f.file_name", "f.file_path", "f.file_url", "f.file_size",
		"f.file_type", "f.resource_type", "f.resource_id", "f.uploaded_by", "f.created_at", "f.updated_at").
		From("files f").
		Join("class_note_files cnf ON f.id = cnf.file_id").
		Where(squirrel.Eq{"cnf.class_note_id": classNoteID}).
		OrderBy("f.created_at DESC").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building get class note files SQL")
		return nil, fmt.Errorf("failed to build get class note files query: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("classNoteID", classNoteID).Msg("Error retrieving class note files")
		return nil, fmt.Errorf("error retrieving class note files: %w", err)
	}
	defer rows.Close()

	files := []*models.File{}
	for rows.Next() {
		var file models.File
		if err := rows.Scan(
			&file.ID,
			&file.FileName,
			&file.FilePath,
			&file.FileURL,
			&file.FileSize,
			&file.FileType,
			&file.ResourceType,
			&file.ResourceID,
			&file.UploadedBy,
			&file.CreatedAt,
			&file.UpdatedAt,
		); err != nil {
			logger.Error().Err(err).Msg("Error scanning class note file row")
			return nil, fmt.Errorf("error scanning class note file: %w", err)
		}
		files = append(files, &file)
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating class note files rows")
		return nil, fmt.Errorf("error iterating class note files: %w", err)
	}

	return files, nil
}

// AddFileToPastExam adds a file to a past exam
func (r *FileRepository) AddFileToPastExam(ctx context.Context, pastExamID, fileID int64) error {
	sql, args, err := r.sb.Insert("past_exam_files").
		Columns("past_exam_id", "file_id").
		Values(pastExamID, fileID).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building add file to past exam SQL")
		return fmt.Errorf("failed to build add file to past exam query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("pastExamID", pastExamID).Int64("fileID", fileID).Msg("Error adding file to past exam")
		return fmt.Errorf("error adding file to past exam: %w", err)
	}

	logger.Info().Int64("pastExamID", pastExamID).Int64("fileID", fileID).Msg("File added to past exam")
	return nil
}

// AddFileToClassNote adds a file to a class note
func (r *FileRepository) AddFileToClassNote(ctx context.Context, classNoteID, fileID int64) error {
	sql, args, err := r.sb.Insert("class_note_files").
		Columns("class_note_id", "file_id").
		Values(classNoteID, fileID).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building add file to class note SQL")
		return fmt.Errorf("failed to build add file to class note query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("classNoteID", classNoteID).Int64("fileID", fileID).Msg("Error adding file to class note")
		return fmt.Errorf("error adding file to class note: %w", err)
	}

	logger.Info().Int64("classNoteID", classNoteID).Int64("fileID", fileID).Msg("File added to class note")
	return nil
}

// RemoveFileFromPastExam removes a file from a past exam
func (r *FileRepository) RemoveFileFromPastExam(ctx context.Context, pastExamID, fileID int64) error {
	sql, args, err := r.sb.Delete("past_exam_files").
		Where(squirrel.Eq{"past_exam_id": pastExamID, "file_id": fileID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building remove file from past exam SQL")
		return fmt.Errorf("failed to build remove file from past exam query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("pastExamID", pastExamID).Int64("fileID", fileID).Msg("Error removing file from past exam")
		return fmt.Errorf("error removing file from past exam: %w", err)
	}

	logger.Info().Int64("pastExamID", pastExamID).Int64("fileID", fileID).Msg("File removed from past exam")
	return nil
}

// RemoveFileFromClassNote removes a file from a class note
func (r *FileRepository) RemoveFileFromClassNote(ctx context.Context, classNoteID, fileID int64) error {
	sql, args, err := r.sb.Delete("class_note_files").
		Where(squirrel.Eq{"class_note_id": classNoteID, "file_id": fileID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building remove file from class note SQL")
		return fmt.Errorf("failed to build remove file from class note query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("classNoteID", classNoteID).Int64("fileID", fileID).Msg("Error removing file from class note")
		return fmt.Errorf("error removing file from class note: %w", err)
	}

	logger.Info().Int64("classNoteID", classNoteID).Int64("fileID", fileID).Msg("File removed from class note")
	return nil
}

// BatchRemoveFilesByPastExamID removes all files associated with a past exam
func (r *FileRepository) BatchRemoveFilesByPastExamID(ctx context.Context, pastExamID int64) error {
	sql, args, err := r.sb.Delete("past_exam_files").
		Where(squirrel.Eq{"past_exam_id": pastExamID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building batch remove files from past exam SQL")
		return fmt.Errorf("failed to build batch remove files from past exam query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("pastExamID", pastExamID).Msg("Error batch removing files from past exam")
		return fmt.Errorf("error batch removing files from past exam: %w", err)
	}

	logger.Info().Int64("pastExamID", pastExamID).Msg("Files batch removed from past exam")
	return nil
}

// BatchRemoveFilesByClassNoteID removes all files associated with a class note
func (r *FileRepository) BatchRemoveFilesByClassNoteID(ctx context.Context, classNoteID int64) error {
	sql, args, err := r.sb.Delete("class_note_files").
		Where(squirrel.Eq{"class_note_id": classNoteID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building batch remove files from class note SQL")
		return fmt.Errorf("failed to build batch remove files from class note query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("classNoteID", classNoteID).Msg("Error batch removing files from class note")
		return fmt.Errorf("error batch removing files from class note: %w", err)
	}

	logger.Info().Int64("classNoteID", classNoteID).Msg("Files batch removed from class note")
	return nil
}
