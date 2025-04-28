package repositories

import (
	"context"
	"fmt"

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
func (r *PastExamRepository) GetAll(ctx context.Context, departmentID *int64, courseCode *string, page, pageSize int) ([]models.PastExam, int64, error) {
	// Build base query
	query := squirrel.Select("id", "course_code", "year", "term", "file_id", "department_id", "instructor_id").
		From("past_exams").
		PlaceholderFormat(squirrel.Dollar)

	// Add filters
	if departmentID != nil {
		query = query.Where("department_id = ?", *departmentID)
	}
	if courseCode != nil {
		query = query.Where("course_code = ?", *courseCode)
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
		err := rows.Scan(
			&exam.ID,
			&exam.CourseCode,
			&exam.Year,
			&exam.Term,
			&exam.FileID,
			&exam.DepartmentID,
			&exam.InstructorID,
			&total,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning row: %w", err)
		}
		exams = append(exams, exam)
	}

	return exams, total, nil
}

// GetByID retrieves a past exam by ID
func (r *PastExamRepository) GetByID(ctx context.Context, id int64) (*models.PastExam, error) {
	query := squirrel.Select("id", "course_code", "year", "term", "file_id", "department_id", "instructor_id").
		From("past_exams").
		Where("id = ?", id).
		PlaceholderFormat(squirrel.Dollar)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL: %w", err)
	}

	var exam models.PastExam
	err = r.db.QueryRow(ctx, sql, args...).Scan(
		&exam.ID,
		&exam.CourseCode,
		&exam.Year,
		&exam.Term,
		&exam.FileID,
		&exam.DepartmentID,
		&exam.InstructorID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error executing query: %w", err)
	}

	return &exam, nil
}

// Create creates a new past exam
func (r *PastExamRepository) Create(ctx context.Context, exam *models.PastExam) (int64, error) {
	query := squirrel.Insert("past_exams").
		Columns("course_code", "year", "term", "file_id", "department_id", "instructor_id").
		Values(exam.CourseCode, exam.Year, exam.Term, exam.FileID, exam.DepartmentID, exam.InstructorID).
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
		Set("course_code", exam.CourseCode).
		Set("year", exam.Year).
		Set("term", exam.Term).
		Set("file_id", exam.FileID).
		Set("department_id", exam.DepartmentID).
		Set("instructor_id", exam.InstructorID).
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
