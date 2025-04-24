package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
)

// PastExam error types
var (
	ErrPastExamNotFound = errors.New("past exam not found")
)

// PastExamRepository handles past exam database operations
type PastExamRepository struct {
	db *pgxpool.Pool
}

// NewPastExamRepository creates a new PastExamRepository
func NewPastExamRepository(db *pgxpool.Pool) *PastExamRepository {
	return &PastExamRepository{
		db: db,
	}
}

// GetAllPastExams retrieves all past exams with pagination and optional filtering
func (r *PastExamRepository) GetAllPastExams(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]models.PastExam, int, error) {
	offset := (page - 1) * pageSize

	// Common base query parts - extracted to constants for readability
	const baseSelectCount = `
		SELECT COUNT(*) 
		FROM past_exams pe
		JOIN departments d ON pe.department_id = d.id
		JOIN faculties f ON d.faculty_id = f.id
		WHERE 1=1
	`

	const baseSelectColumns = `
		SELECT 
			pe.id, pe.year, pe.term, pe.department_id, pe.course_code, 
			pe.title, pe.content, pe.file_url, pe.instructor_id, 
			pe.created_at, pe.updated_at,
			d.name as department_name,
			f.id as faculty_id, f.name as faculty_name,
			u.first_name || ' ' || u.last_name as instructor_name,
			u.email as uploaded_by_email
		FROM past_exams pe
		JOIN departments d ON pe.department_id = d.id
		JOIN faculties f ON d.faculty_id = f.id
		LEFT JOIN instructors i ON pe.instructor_id = i.id
		LEFT JOIN users u ON i.user_id = u.id
		WHERE 1=1
	`

	// Build where clause for filtering
	whereClause, whereArgs := buildWhereClause(filters)

	// Full count query
	countQuery := baseSelectCount
	if whereClause != "" {
		countQuery += " AND " + whereClause
	}

	// Execute count query
	var totalItems int
	err := r.db.QueryRow(ctx, countQuery, whereArgs...).Scan(&totalItems)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count past exams: %w", err)
	}

	// Build main query
	query := baseSelectColumns
	if whereClause != "" {
		query += " AND " + whereClause
	}

	// Add sorting - using prepared statement parameters instead of string concatenation
	// to avoid SQL injection
	sortBy := "created_at"
	sortOrder := "DESC"

	if val, ok := filters["sortBy"]; ok && val != nil {
		if field, ok := val.(string); ok && isValidSortField(field) {
			sortBy = field
		}
	}

	if val, ok := filters["sortOrder"]; ok && val != nil {
		if order, ok := val.(string); ok && (order == "ASC" || order == "DESC") {
			sortOrder = order
		}
	}

	// Build sort and pagination part of the query using a safer approach
	// Add ORDER BY clause based on validated sortBy and sortOrder values
	query += " ORDER BY pe." + sortBy + " " + sortOrder + " LIMIT $" +
		fmt.Sprintf("%d", len(whereArgs)+1) + " OFFSET $" + fmt.Sprintf("%d", len(whereArgs)+2)

	// Add pagination args
	queryArgs := append(whereArgs, pageSize, offset)

	// Execute query
	rows, err := r.db.Query(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query past exams: %w", err)
	}
	defer rows.Close()

	var pastExams []models.PastExam
	for rows.Next() {
		var pastExam models.PastExam
		var departmentName, facultyName, instructorName, uploadedByEmail sql.NullString
		var facultyID sql.NullInt64

		err := rows.Scan(
			&pastExam.ID, &pastExam.Year, &pastExam.Term, &pastExam.DepartmentID,
			&pastExam.CourseCode, &pastExam.Title, &pastExam.Content, &pastExam.FileURL,
			&pastExam.InstructorID, &pastExam.CreatedAt, &pastExam.UpdatedAt,
			&departmentName, &facultyID, &facultyName, &instructorName, &uploadedByEmail,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan past exam row: %w", err)
		}

		// Handle null values
		if departmentName.Valid {
			pastExam.Department = &models.Department{ID: pastExam.DepartmentID, Name: departmentName.String}
		}
		if facultyID.Valid && facultyName.Valid {
			pastExam.FacultyID = facultyID.Int64
			pastExam.Faculty = &models.Faculty{ID: facultyID.Int64, Name: facultyName.String}
		}
		if instructorName.Valid {
			pastExam.UploadedByName = instructorName.String
		}
		if uploadedByEmail.Valid {
			pastExam.UploadedByEmail = uploadedByEmail.String
		}

		pastExams = append(pastExams, pastExam)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating past exam rows: %w", err)
	}

	return pastExams, totalItems, nil
}

// GetPastExamByID retrieves a past exam by its ID
func (r *PastExamRepository) GetPastExamByID(ctx context.Context, id int64) (*models.PastExam, error) {
	query := `
		SELECT 
			pe.id, pe.year, pe.term, pe.department_id, pe.course_code, 
			pe.title, pe.content, pe.file_url, pe.instructor_id, 
			pe.created_at, pe.updated_at,
			d.name as department_name,
			f.id as faculty_id, f.name as faculty_name,
			u.first_name || ' ' || u.last_name as instructor_name,
			u.email as uploaded_by_email
		FROM past_exams pe
		JOIN departments d ON pe.department_id = d.id
		JOIN faculties f ON d.faculty_id = f.id
		LEFT JOIN instructors i ON pe.instructor_id = i.id
		LEFT JOIN users u ON i.user_id = u.id
		WHERE pe.id = $1
	`

	var pastExam models.PastExam
	var departmentName, facultyName, instructorName, uploadedByEmail sql.NullString
	var facultyID sql.NullInt64

	err := r.db.QueryRow(ctx, query, id).Scan(
		&pastExam.ID, &pastExam.Year, &pastExam.Term, &pastExam.DepartmentID,
		&pastExam.CourseCode, &pastExam.Title, &pastExam.Content, &pastExam.FileURL,
		&pastExam.InstructorID, &pastExam.CreatedAt, &pastExam.UpdatedAt,
		&departmentName, &facultyID, &facultyName, &instructorName, &uploadedByEmail,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPastExamNotFound
		}
		return nil, fmt.Errorf("error querying past exam ID=%d: %w", id, err)
	}

	// Handle null values
	if departmentName.Valid {
		pastExam.Department = &models.Department{ID: pastExam.DepartmentID, Name: departmentName.String}
	}
	if facultyID.Valid && facultyName.Valid {
		pastExam.FacultyID = facultyID.Int64
		pastExam.Faculty = &models.Faculty{ID: facultyID.Int64, Name: facultyName.String}
	}
	if instructorName.Valid {
		pastExam.UploadedByName = instructorName.String
	}
	if uploadedByEmail.Valid {
		pastExam.UploadedByEmail = uploadedByEmail.String
	}

	return &pastExam, nil
}

// CreatePastExam inserts a new past exam into the database
func (r *PastExamRepository) CreatePastExam(ctx context.Context, pastExam *models.PastExam) (int64, error) {
	// InstructorID zaten bulunuyor, öğretim görevlisini aramaya gerek yok
	query := `
		INSERT INTO past_exams (
			year, term, department_id, course_code, title, content, 
			file_url, instructor_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query,
		pastExam.Year, pastExam.Term, pastExam.DepartmentID, pastExam.CourseCode,
		pastExam.Title, pastExam.Content, pastExam.FileURL, pastExam.InstructorID,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error inserting past exam: %w", err)
	}

	return id, nil
}

// UpdatePastExam updates an existing past exam in the database
func (r *PastExamRepository) UpdatePastExam(ctx context.Context, pastExam *models.PastExam) error {
	query := `
		UPDATE past_exams SET
			year = $1, term = $2, department_id = $3, course_code = $4, 
			title = $5, content = $6, file_url = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
	`

	cmdTag, err := r.db.Exec(ctx, query,
		pastExam.Year, pastExam.Term, pastExam.DepartmentID, pastExam.CourseCode,
		pastExam.Title, pastExam.Content, pastExam.FileURL, pastExam.ID,
	)
	if err != nil {
		return fmt.Errorf("error updating past exam ID=%d: %w", pastExam.ID, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrPastExamNotFound
	}

	return nil
}

// DeletePastExam removes a past exam from the database
func (r *PastExamRepository) DeletePastExam(ctx context.Context, id int64) error {
	query := "DELETE FROM past_exams WHERE id = $1"

	cmdTag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting past exam ID=%d: %w", id, err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrPastExamNotFound
	}

	return nil
}

// buildWhereClause constructs the WHERE clause for filtering past exams
func buildWhereClause(filters map[string]interface{}) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	// Add conditions based on the filters
	for key, value := range filters {
		// Skip sort and paging related parameters
		if key == "sortBy" || key == "sortOrder" || value == nil {
			continue
		}

		switch key {
		case "year":
			conditions = append(conditions, "pe.year = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, value)
			argIndex++
		case "term":
			conditions = append(conditions, "pe.term = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, value)
			argIndex++
		case "departmentId":
			conditions = append(conditions, "pe.department_id = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, value)
			argIndex++
		case "facultyId":
			conditions = append(conditions, "d.faculty_id = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, value)
			argIndex++
		case "courseCode":
			conditions = append(conditions, "pe.course_code = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, value)
			argIndex++
		case "instructorId":
			conditions = append(conditions, "pe.instructor_id = $"+fmt.Sprintf("%d", argIndex))
			args = append(args, value)
			argIndex++
		case "search":
			// Search in title, content, and course code (case insensitive)
			searchTerm := "%" + fmt.Sprintf("%v", value) + "%"
			conditions = append(conditions, "(LOWER(pe.title) LIKE LOWER($"+fmt.Sprintf("%d", argIndex)+
				") OR LOWER(pe.content) LIKE LOWER($"+fmt.Sprintf("%d", argIndex)+
				") OR LOWER(pe.course_code) LIKE LOWER($"+fmt.Sprintf("%d", argIndex)+"))")
			args = append(args, searchTerm)
			argIndex++
		}
	}

	whereClause := strings.Join(conditions, " AND ")
	return whereClause, args
}

// isValidSortField checks if a sort field is valid to prevent SQL injection
func isValidSortField(field string) bool {
	// List of column names that are allowed to be used for sorting
	// This is a security measure to prevent SQL injection
	allowedFields := []string{
		"id", "year", "term", "department_id",
		"course_code", "title", "created_at", "updated_at",
	}

	// Check if the provided field is in the list of allowed fields
	for _, allowedField := range allowedFields {
		if field == allowedField {
			return true
		}
	}

	return false
}
