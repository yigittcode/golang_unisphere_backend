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

// PastExamRepository handles database operations for past exams
type PastExamRepository struct {
	db *pgxpool.Pool
}

// NewPastExamRepository creates a new PastExamRepository
func NewPastExamRepository(db *pgxpool.Pool) *PastExamRepository {
	return &PastExamRepository{db: db}
}

// GetAll retrieves all past exams with filtering and pagination
func (r *PastExamRepository) GetAll(ctx context.Context, departmentID *int64, courseCode *string, year *int, term *string, page, pageSize int) ([]models.PastExam, int64, error) {
	// Build base query
	query := squirrel.Select(
		"id", "year", "term", "course_code", "title", "content",
		"department_id", "instructor_id", "created_at", "updated_at",
	).
		From("past_exams").
		PlaceholderFormat(squirrel.Dollar)

	// Add filters
	if departmentID != nil {
		query = query.Where("department_id = ?", *departmentID)
	}
	if courseCode != nil {
		query = query.Where("course_code = ?", *courseCode)
	}
	if year != nil {
		query = query.Where("year = ?", *year)
	}
	if term != nil {
		query = query.Where("term = ?", *term)
	}

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

	var exams []models.PastExam
	var total int64

	for rows.Next() {
		var exam models.PastExam
		var termStr string
		err := rows.Scan(
			&exam.ID,
			&exam.Year,
			&termStr,
			&exam.CourseCode,
			&exam.Title,
			&exam.Content,
			&exam.DepartmentID,
			&exam.InstructorID,
			&exam.CreatedAt,
			&exam.UpdatedAt,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning row: %w", err)
		}
		exam.Term = models.Term(termStr)
		exams = append(exams, exam)
	}

	return exams, total, nil
}

// GetByID retrieves a past exam by ID
func (r *PastExamRepository) GetByID(ctx context.Context, id int64) (*models.PastExam, error) {
	query := squirrel.Select(
		"id", "year", "term", "course_code", "title", "content",
		"department_id", "instructor_id", "created_at", "updated_at",
	).
		From("past_exams").
		Where("id = ?", id).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var exam models.PastExam
	var termStr string
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&exam.ID,
		&exam.Year,
		&termStr,
		&exam.CourseCode,
		&exam.Title,
		&exam.Content,
		&exam.DepartmentID,
		&exam.InstructorID,
		&exam.CreatedAt,
		&exam.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	exam.Term = models.Term(termStr)

	// Get the files for this exam
	files, err := r.GetPastExamFiles(ctx, exam.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting files: %w", err)
	}
	exam.Files = files

	return &exam, nil
}

// Create creates a new past exam
func (r *PastExamRepository) Create(ctx context.Context, exam *models.PastExam) (int64, error) {
	query := squirrel.Insert("past_exams").
		Columns(
			"year", "term", "course_code", "title", "content",
			"department_id", "instructor_id",
		).
		Values(
			exam.Year, string(exam.Term), exam.CourseCode, exam.Title, exam.Content,
			exam.DepartmentID, exam.InstructorID,
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

// Update updates an existing past exam
func (r *PastExamRepository) Update(ctx context.Context, exam *models.PastExam) error {
	query := squirrel.Update("past_exams").
		Set("year", exam.Year).
		Set("term", string(exam.Term)).
		Set("course_code", exam.CourseCode).
		Set("title", exam.Title).
		Set("content", exam.Content).
		Set("department_id", exam.DepartmentID).
		Set("instructor_id", exam.InstructorID).
		Set("updated_at", time.Now()).
		Where("id = ?", exam.ID).
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

// Delete deletes a past exam
func (r *PastExamRepository) Delete(ctx context.Context, id int64) error {
	query := squirrel.Delete("past_exams").
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

// AddFileToPastExam adds a file to a past exam
func (r *PastExamRepository) AddFileToPastExam(ctx context.Context, pastExamID int64, fileID int64) error {
	query := squirrel.Insert("past_exam_files").
		Columns("past_exam_id", "file_id").
		Values(pastExamID, fileID).
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

// RemoveFileFromPastExam removes a file from a past exam
func (r *PastExamRepository) RemoveFileFromPastExam(ctx context.Context, pastExamID int64, fileID int64) error {
	query := squirrel.Delete("past_exam_files").
		Where("past_exam_id = ?", pastExamID).
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

// GetPastExamFiles gets all files associated with a past exam
func (r *PastExamRepository) GetPastExamFiles(ctx context.Context, pastExamID int64) ([]*models.File, error) {
	query := squirrel.Select("f.id", "f.file_name", "f.file_path", "f.file_url",
		"f.file_size", "f.file_type", "f.resource_type", "f.resource_id",
		"f.uploaded_by", "f.created_at", "f.updated_at").
		From("files f").
		Join("past_exam_files pef ON f.id = pef.file_id").
		Where("pef.past_exam_id = ?", pastExamID).
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
