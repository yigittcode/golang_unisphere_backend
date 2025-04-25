package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// PastExam error types
var (
	ErrPastExamNotFound = errors.New("past exam not found")
)

// PastExamRepository handles past exam database operations
type PastExamRepository struct {
	db *pgxpool.Pool
	sb squirrel.StatementBuilderType
}

// NewPastExamRepository creates a new PastExamRepository
func NewPastExamRepository(db *pgxpool.Pool) *PastExamRepository {
	return &PastExamRepository{
		db: db,
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

// Helper function to get nullable string from pointer
func getNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

// Helper function to get nullable string from value
func getContentNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// Helper function to get nullable int64
func getNullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: i, Valid: true}
}

// GetAllPastExams retrieves all past exams with pagination and optional filtering/sorting
func (r *PastExamRepository) GetAllPastExams(ctx context.Context, page, pageSize int, filters map[string]interface{}) ([]models.PastExam, int, error) {
	offset := uint64((page - 1) * pageSize)

	// Base select query
	baseSelect := r.sb.Select(
		"pe.id", "pe.year", "pe.term", "pe.department_id", "pe.course_code",
		"pe.title", "pe.content", "pe.file_url", "pe.instructor_id",
		"pe.created_at", "pe.updated_at",
		"d.name as department_name",
		"f.id as faculty_id", "f.name as faculty_name",
		"COALESCE(u.first_name || ' ' || u.last_name, '') as instructor_name",
		"COALESCE(u.email, '') as uploaded_by_email",
	).
		From("past_exams pe").
		Join("departments d ON pe.department_id = d.id").
		Join("faculties f ON d.faculty_id = f.id").
		LeftJoin("instructors i ON pe.instructor_id = i.id").
		LeftJoin("users u ON i.user_id = u.id")

	// Count query (without limit/offset/order)
	countSelect := r.sb.Select("COUNT(*)").
		From("past_exams pe").
		Join("departments d ON pe.department_id = d.id").
		Join("faculties f ON d.faculty_id = f.id").
		LeftJoin("instructors i ON pe.instructor_id = i.id").
		LeftJoin("users u ON i.user_id = u.id")

	// Apply filters
	whereCondition := squirrel.And{}
	if facultyID, ok := filters["faculty_id"].(int); ok && facultyID > 0 {
		whereCondition = append(whereCondition, squirrel.Eq{"f.id": facultyID})
	}
	if departmentID, ok := filters["department_id"].(int); ok && departmentID > 0 {
		whereCondition = append(whereCondition, squirrel.Eq{"pe.department_id": departmentID})
	}
	if year, ok := filters["year"].(int); ok && year > 0 {
		whereCondition = append(whereCondition, squirrel.Eq{"pe.year": year})
	}
	if term, ok := filters["term"].(string); ok && term != "" {
		whereCondition = append(whereCondition, squirrel.Eq{"pe.term": term})
	}
	if courseCode, ok := filters["course_code"].(string); ok && courseCode != "" {
		whereCondition = append(whereCondition, squirrel.ILike{"pe.course_code": "%" + strings.TrimSpace(courseCode) + "%"})
	}
	if title, ok := filters["title"].(string); ok && title != "" {
		whereCondition = append(whereCondition, squirrel.ILike{"pe.title": "%" + strings.TrimSpace(title) + "%"})
	}
	if instructorName, ok := filters["instructor_name"].(string); ok && instructorName != "" {
		whereCondition = append(whereCondition, squirrel.Expr("u.first_name || ' ' || u.last_name ILIKE ?", "%"+strings.TrimSpace(instructorName)+"%"))
	}

	baseSelect = baseSelect.Where(whereCondition)
	countSelect = countSelect.Where(whereCondition)

	// --- Execute Count Query ---
	countSql, countArgs, err := countSelect.ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("Error building count past exams SQL")
		return nil, 0, fmt.Errorf("failed to build count past exams query: %w", err)
	}

	var totalItems int
	err = r.db.QueryRow(ctx, countSql, countArgs...).Scan(&totalItems)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing count past exams query")
		return nil, 0, fmt.Errorf("failed to count past exams: %w", err)
	}

	if totalItems == 0 {
		return []models.PastExam{}, 0, nil
	}

	// --- Apply Sorting and Pagination ---
	sortBy := "created_at"
	sortOrder := "DESC"

	if val, ok := filters["sortBy"]; ok {
		if field, ok := val.(string); ok && isValidSortField(field) {
			sortBy = field
		}
	}
	if val, ok := filters["sortOrder"]; ok {
		if order, ok := val.(string); ok && (strings.ToUpper(order) == "ASC" || strings.ToUpper(order) == "DESC") {
			sortOrder = strings.ToUpper(order)
		}
	}

	// Map model fields to DB columns for sorting
	dbSortColumn := mapSortFieldToColumn(sortBy)

	baseSelect = baseSelect.OrderBy(fmt.Sprintf("%s %s", dbSortColumn, sortOrder)).
		Limit(uint64(pageSize)).
		Offset(offset)

	// --- Execute Main Query ---
	querySql, queryArgs, err := baseSelect.ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("Error building get all past exams SQL")
		return nil, 0, fmt.Errorf("failed to build get past exams query: %w", err)
	}

	rows, err := r.db.Query(ctx, querySql, queryArgs...)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing get all past exams query")
		return nil, 0, fmt.Errorf("failed to query past exams: %w", err)
	}
	defer rows.Close()

	var pastExams []models.PastExam
	for rows.Next() {
		var pastExam models.PastExam
		var departmentName, facultyName, instructorName, uploadedByEmail sql.NullString
		var facultyID, nullableInstructorID sql.NullInt64

		err := rows.Scan(
			&pastExam.ID, &pastExam.Year, &pastExam.Term, &pastExam.DepartmentID,
			&pastExam.CourseCode, &pastExam.Title, &pastExam.Content, &pastExam.FileURL,
			&nullableInstructorID,
			&pastExam.CreatedAt, &pastExam.UpdatedAt,
			&departmentName, &facultyID, &facultyName, &instructorName, &uploadedByEmail,
		)
		if err != nil {
			logger.Error().Err(err).Msg("Error scanning past exam row")
			return nil, 0, fmt.Errorf("failed to scan past exam row: %w", err)
		}

		if departmentName.Valid {
			pastExam.Department = &models.Department{ID: pastExam.DepartmentID, Name: departmentName.String}
		}
		if facultyID.Valid && facultyName.Valid {
			pastExam.FacultyID = facultyID.Int64
			pastExam.Faculty = &models.Faculty{ID: facultyID.Int64, Name: facultyName.String}
		}
		if nullableInstructorID.Valid {
			pastExam.InstructorID = nullableInstructorID.Int64
			if instructorName.Valid {
				pastExam.UploadedByName = instructorName.String
			}
			if uploadedByEmail.Valid {
				pastExam.UploadedByEmail = uploadedByEmail.String
			}
		} else {
			pastExam.InstructorID = 0
		}

		pastExams = append(pastExams, pastExam)
	}

	if err := rows.Err(); err != nil {
		logger.Error().Err(err).Msg("Error iterating past exam rows")
		return nil, 0, fmt.Errorf("error iterating past exam rows: %w", err)
	}

	logger.Info().Int("page", page).Int("pageSize", pageSize).Int("totalItems", totalItems).Int("returnedItems", len(pastExams)).Msg("Successfully fetched past exams")
	return pastExams, totalItems, nil
}

// GetPastExamByID retrieves a past exam by its ID including related data
func (r *PastExamRepository) GetPastExamByID(ctx context.Context, id int64) (*models.PastExam, error) {
	selectBuilder := r.sb.Select(
		"pe.id", "pe.year", "pe.term", "pe.department_id", "pe.course_code",
		"pe.title", "pe.content", "pe.file_url", "pe.instructor_id",
		"pe.created_at", "pe.updated_at",
		"d.name as department_name",
		"f.id as faculty_id", "f.name as faculty_name",
		"COALESCE(u.first_name || ' ' || u.last_name, '') as instructor_name",
		"COALESCE(u.email, '') as uploaded_by_email",
	).
		From("past_exams pe").
		Join("departments d ON pe.department_id = d.id").
		Join("faculties f ON d.faculty_id = f.id").
		LeftJoin("instructors i ON pe.instructor_id = i.id").
		LeftJoin("users u ON i.user_id = u.id").
		Where(squirrel.Eq{"pe.id": id}).
		Limit(1)

	sqlQuery, args, err := selectBuilder.ToSql()
	if err != nil {
		logger.Error().Err(err).Msg("Error building get past exam by ID SQL")
		return nil, fmt.Errorf("failed to build get past exam query: %w", err)
	}

	var pastExam models.PastExam
	var departmentName, facultyName, instructorName, uploadedByEmail sql.NullString
	var facultyID, nullableInstructorID sql.NullInt64

	err = r.db.QueryRow(ctx, sqlQuery, args...).Scan(
		&pastExam.ID, &pastExam.Year, &pastExam.Term, &pastExam.DepartmentID,
		&pastExam.CourseCode, &pastExam.Title, &pastExam.Content, &pastExam.FileURL,
		&nullableInstructorID,
		&pastExam.CreatedAt, &pastExam.UpdatedAt,
		&departmentName, &facultyID, &facultyName, &instructorName, &uploadedByEmail,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			logger.Warn().Int64("pastExamID", id).Msg("Past exam not found by ID")
			return nil, ErrPastExamNotFound
		}
		logger.Error().Err(err).Int64("pastExamID", id).Msg("Error scanning past exam row by ID")
		return nil, fmt.Errorf("error querying past exam ID=%d: %w", id, err)
	}

	if departmentName.Valid {
		pastExam.Department = &models.Department{ID: pastExam.DepartmentID, Name: departmentName.String}
	}
	if facultyID.Valid && facultyName.Valid {
		pastExam.FacultyID = facultyID.Int64
		pastExam.Faculty = &models.Faculty{ID: facultyID.Int64, Name: facultyName.String}
	}
	if nullableInstructorID.Valid {
		pastExam.InstructorID = nullableInstructorID.Int64
		if instructorName.Valid {
			pastExam.UploadedByName = instructorName.String
		}
		if uploadedByEmail.Valid {
			pastExam.UploadedByEmail = uploadedByEmail.String
		}
	} else {
		pastExam.InstructorID = 0
	}

	return &pastExam, nil
}

// CreatePastExam inserts a new past exam into the database
func (r *PastExamRepository) CreatePastExam(ctx context.Context, pastExam *models.PastExam) (int64, error) {
	var instructorIDArg interface{}
	if pastExam.InstructorID != 0 {
		instructorIDArg = pastExam.InstructorID
	} else {
		instructorIDArg = nil
	}

	sql, args, err := r.sb.Insert("past_exams").
		Columns(
			"year", "term", "department_id", "course_code", "title", "content",
			"file_url", "instructor_id",
		).
		Values(
			pastExam.Year, pastExam.Term, pastExam.DepartmentID, pastExam.CourseCode,
			pastExam.Title, pastExam.Content, getNullString(pastExam.FileURL),
			instructorIDArg,
		).
		Suffix("RETURNING id").
		ToSql()

	if err != nil {
		logger.Error().Err(err).Msg("Error building create past exam SQL")
		return 0, fmt.Errorf("failed to build create past exam query: %w", err)
	}

	var id int64
	err = r.db.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		logger.Error().Err(err).Msg("Error executing create past exam query")
		return 0, fmt.Errorf("error inserting past exam: %w", err)
	}

	logger.Info().Int64("pastExamID", id).Msg("Past exam created successfully")
	return id, nil
}

// UpdatePastExam updates an existing past exam in the database
func (r *PastExamRepository) UpdatePastExam(ctx context.Context, pastExam *models.PastExam) error {
	var instructorIDArg interface{}
	if pastExam.InstructorID != 0 {
		instructorIDArg = pastExam.InstructorID
	} else {
		instructorIDArg = nil
	}

	sql, args, err := r.sb.Update("past_exams").
		SetMap(map[string]interface{}{
			"year":          pastExam.Year,
			"term":          pastExam.Term,
			"department_id": pastExam.DepartmentID,
			"course_code":   pastExam.CourseCode,
			"title":         pastExam.Title,
			"content":       getContentNullString(pastExam.Content),
			"file_url":      getNullString(pastExam.FileURL),
			"instructor_id": instructorIDArg,
			"updated_at":    time.Now(),
		}).
		Where(squirrel.Eq{"id": pastExam.ID}).
		ToSql()

	if err != nil {
		logger.Error().Err(err).Int64("pastExamID", pastExam.ID).Msg("Error building update past exam SQL")
		return fmt.Errorf("failed to build update past exam query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		// Log error before returning
		logger.Error().Err(err).Int64("pastExamID", pastExam.ID).Msg("Error executing update past exam query")
		return fmt.Errorf("error updating past exam ID=%d: %w", pastExam.ID, err)
	}

	if cmdTag.RowsAffected() == 0 {
		// Log warning before returning
		logger.Warn().Int64("pastExamID", pastExam.ID).Msg("Attempted to update non-existent past exam")
		return ErrPastExamNotFound
	}

	// Log success
	logger.Info().Int64("pastExamID", pastExam.ID).Msg("Past exam updated successfully")
	return nil
}

// DeletePastExam removes a past exam from the database
func (r *PastExamRepository) DeletePastExam(ctx context.Context, id int64) error {
	sql, args, err := r.sb.Delete("past_exams"). // Use squirrel
							Where(squirrel.Eq{"id": id}).
							ToSql()

	if err != nil {
		logger.Error().Err(err).Int64("pastExamID", id).Msg("Error building delete past exam SQL")
		return fmt.Errorf("failed to build delete past exam query: %w", err)
	}

	cmdTag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		logger.Error().Err(err).Int64("pastExamID", id).Msg("Error executing delete past exam query")
		return fmt.Errorf("error deleting past exam ID=%d: %w", id, err)
	}

	if cmdTag.RowsAffected() == 0 {
		logger.Warn().Int64("pastExamID", id).Msg("Attempted to delete non-existent past exam")
		return ErrPastExamNotFound
	}

	logger.Info().Int64("pastExamID", id).Msg("Past exam deleted successfully")
	return nil
}

// mapSortFieldToColumn maps API sort field names to database column names
// Prevents SQL injection by using a predefined map
func mapSortFieldToColumn(field string) string {
	switch field {
	case "year":
		return "pe.year"
	case "term":
		return "pe.term"
	case "courseCode", "course_code":
		return "pe.course_code"
	case "title":
		return "pe.title"
	case "departmentName", "department_name":
		return "d.name"
	case "facultyName", "faculty_name":
		return "f.name"
	case "instructorName", "instructor_name":
		return "instructor_name" // Alias used in select
	case "createdAt", "created_at":
		return "pe.created_at"
	case "updatedAt", "updated_at":
		return "pe.updated_at"
	default:
		return "pe.created_at" // Default sort column
	}
}

// isValidSortField checks if the provided sort field is allowed
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
