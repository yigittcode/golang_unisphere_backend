package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
)

// ClassNoteRepository handles database operations for class notes
type ClassNoteRepository struct {
	db *pgxpool.Pool
}

// NewClassNoteRepository creates a new ClassNoteRepository
func NewClassNoteRepository(db *pgxpool.Pool) *ClassNoteRepository {
	return &ClassNoteRepository{db: db}
}

// GetAll retrieves all class notes with filtering, sorting and pagination
func (r *ClassNoteRepository) GetAll(ctx context.Context, departmentID *int64, courseCode *string, instructorID *int64, page, pageSize int, sortBy, sortOrder string) ([]models.ClassNote, int64, error) {
	// Build base query
	query := squirrel.Select(
		"id", "course_code", "title", "description", "content",
		"department_id", "user_id", "created_at", "updated_at",
	).
		From("class_notes").
		PlaceholderFormat(squirrel.Dollar)

	// Add filters
	if departmentID != nil {
		query = query.Where("department_id = ?", *departmentID)
	}
	if courseCode != nil {
		query = query.Where("course_code = ?", *courseCode)
	}
	if instructorID != nil {
		query = query.Where("user_id = ?", *instructorID)
	}

	// Add sorting with validation
	// Default to created_at if empty or invalid sort column
	if sortBy == "" {
		sortBy = "created_at"
	}
	
	// Validate sortBy field (whitelist approach for security)
	validSortColumns := map[string]bool{
		"created_at":  true,
		"updated_at":  true,
		"title":       true,
		"course_code": true,
	}
	
	if !validSortColumns[sortBy] {
		sortBy = "created_at" // Default to created_at if invalid
	}
	
	// Validate sortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc" // Default to descending if invalid
	}
	
	// Apply sorting
	query = query.OrderBy(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Add pagination
	offset := (page - 1) * pageSize
	query = query.Limit(uint64(pageSize)).Offset(uint64(offset))

	// Get total count
	countQuery := query.Column("COUNT(*) OVER()")
	sql, args, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var notes []models.ClassNote
	var total int64

	for rows.Next() {
		var note models.ClassNote
		err := rows.Scan(
			&note.ID,
			&note.CourseCode,
			&note.Title,
			&note.Description,
			&note.Content,
			&note.DepartmentID,
			&note.UserID,
			&note.CreatedAt,
			&note.UpdatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning row: %w", err)
		}
		notes = append(notes, note)
	}

	return notes, total, nil
}

// GetByID retrieves a class note by ID
func (r *ClassNoteRepository) GetByID(ctx context.Context, id int64) (*models.ClassNote, error) {
	query := squirrel.Select(
		"id", "course_code", "title", "description", "content",
		"department_id", "user_id", "created_at", "updated_at",
	).
		From("class_notes").
		Where("id = ?", id).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var note models.ClassNote
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&note.ID,
		&note.CourseCode,
		&note.Title,
		&note.Description,
		&note.Content,
		&note.DepartmentID,
		&note.UserID,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	// Dosyaları getir
	files, err := r.GetClassNoteFiles(ctx, note.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting files: %w", err)
	}
	note.Files = files

	return &note, nil
}

// Create creates a new class note
func (r *ClassNoteRepository) Create(ctx context.Context, note *models.ClassNote) (int64, error) {
	query := squirrel.Insert("class_notes").
		Columns(
			"course_code", "title", "description", "content",
			"department_id", "user_id",
		).
		Values(
			note.CourseCode, note.Title, note.Description, note.Content,
			note.DepartmentID, note.UserID,
		).
		Suffix("RETURNING id").
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("error building SQL: %w", err)
	}

	var id int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("error executing query: %w", err)
	}

	return id, nil
}

// Update updates an existing class note
func (r *ClassNoteRepository) Update(ctx context.Context, note *models.ClassNote) error {
	// Log the SQL query for debugging
	fmt.Printf("Updating class note with ID: %d\n", note.ID)
	
	query := squirrel.Update("class_notes").
		Set("course_code", note.CourseCode).
		Set("title", note.Title).
		Set("description", note.Description).
		Set("content", note.Content).
		Set("department_id", note.DepartmentID).
		Set("user_id", note.UserID).
		Set("updated_at", time.Now()).
		Where("id = ?", note.ID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	// Log the SQL query for debugging
	fmt.Printf("Generated SQL: %s\n", sql)
	fmt.Printf("SQL args: %v\n", args)

	result, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		fmt.Printf("Database error: %v\n", err)
		return fmt.Errorf("error executing query: %w", err)
	}

	// Log the affected rows
	rowsAffected := result.RowsAffected()
	fmt.Printf("Rows affected: %d\n", rowsAffected)

	if rowsAffected == 0 {
		// Check if the note exists
		var count int
		checkSQL := "SELECT COUNT(*) FROM class_notes WHERE id = $1"
		err := r.db.QueryRow(ctx, checkSQL, note.ID).Scan(&count)
		if err != nil {
			fmt.Printf("Error checking note existence: %v\n", err)
			return fmt.Errorf("error checking note existence: %w", err)
		}
		
		if count == 0 {
			fmt.Printf("Note with ID %d does not exist\n", note.ID)
			return fmt.Errorf("class note with id %d does not exist", note.ID)
		} else {
			fmt.Printf("Note exists but no rows were affected\n")
			return fmt.Errorf("note exists but no rows were affected, possibly no changes were made")
		}
	}

	return nil
}

// Delete deletes a class note
func (r *ClassNoteRepository) Delete(ctx context.Context, id int64) error {
	query := squirrel.Delete("class_notes").
		Where("id = ?", id).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	result, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

// AddFileToClassNote adds a file to a class note
func (r *ClassNoteRepository) AddFileToClassNote(ctx context.Context, classNoteID int64, fileID int64) error {
	query := squirrel.Insert("class_note_files").
		Columns("class_note_id", "file_id").
		Values(classNoteID, fileID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	return nil
}

// RemoveFileFromClassNote removes a file from a class note
func (r *ClassNoteRepository) RemoveFileFromClassNote(ctx context.Context, classNoteID int64, fileID int64) error {
	query := squirrel.Delete("class_note_files").
		Where("class_note_id = ?", classNoteID).
		Where("file_id = ?", fileID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("error building SQL: %w", err)
	}

	result, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("error executing query: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}

// GetClassNoteFiles gets all files associated with a class note
func (r *ClassNoteRepository) GetClassNoteFiles(ctx context.Context, classNoteID int64) ([]*models.File, error) {
	query := squirrel.Select("f.id", "f.file_name", "f.file_path", "f.file_url",
		"f.file_size", "f.file_type", "f.resource_type", "f.resource_id",
		"f.uploaded_by", "f.created_at", "f.updated_at").
		From("files f").
		Join("class_note_files cnf ON f.id = cnf.file_id").
		Where("cnf.class_note_id = ?", classNoteID).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	var files []*models.File
	for rows.Next() {
		var file models.File
		err := rows.Scan(
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
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		files = append(files, &file)
	}

	return files, nil
}
