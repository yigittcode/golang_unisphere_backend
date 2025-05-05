package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// These errors have been moved to apperrors package
// Use apperrors.ErrUserNotFound, apperrors.ErrEmailAlreadyExists, and apperrors.ErrIdentifierExists instead

// IUserRepository defines the interface for user-related database operations
type IUserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id int64) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id int64) error

	// Authentication
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)

	// Profile
	UpdateProfile(ctx context.Context, userID int64, firstName, lastName, email string) error
	UpdateProfilePhotoFileID(ctx context.Context, userID int64, fileID *int64) error

	// Department
	GetDepartmentNameByID(ctx context.Context, departmentID int64) (string, error)

	// Email verification
	SetEmailVerified(ctx context.Context, userID int64, verified bool) error
	IsEmailVerified(ctx context.Context, userID int64) (bool, error)

	// For backward compatibility - to be deprecated
	CreateInstructor(ctx context.Context, instructor *models.Instructor) error
	GetInstructorByUserID(ctx context.Context, userID int64) (*models.Instructor, error)
	CreateStudent(ctx context.Context, student *models.Student) error
	GetStudentByUserID(ctx context.Context, userID int64) (*models.Student, error)

	// Advanced filtering
	FindByFilter(ctx context.Context, departmentID *int64, roleType *models.RoleType, email *string, name *string, page, pageSize int) ([]*models.User, int64, error)
}

// UserRepository combines all user-related repositories
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// GetDB returns the database connection pool
func (r *UserRepository) GetDB() *pgxpool.Pool {
	return r.db
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (int64, error) {
	query := `
		INSERT INTO users 
		(email, password, first_name, last_name, role_type, is_active, email_verified, department_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query,
		user.Email, user.Password, user.FirstName, user.LastName,
		user.RoleType, user.IsActive, user.EmailVerified, user.DepartmentID).Scan(&id)

	if err != nil {
		// Check for duplicate email error
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
			return 0, apperrors.ErrEmailAlreadyExists
		}
		return 0, fmt.Errorf("error creating user: %w", err)
	}

	return id, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password, first_name, last_name, created_at, updated_at, 
		role_type, is_active, last_login_at, department_id, profile_photo_file_id
		FROM users
		WHERE email = $1
	`

	var user models.User
	var lastLoginAt *time.Time

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.RoleType, &user.IsActive, &lastLoginAt,
		&user.DepartmentID, &user.ProfilePhotoFileID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("error retrieving user by email: %w", err)
	}

	user.LastLoginAt = lastLoginAt

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT id, email, password, first_name, last_name, created_at, updated_at, 
		role_type, is_active, last_login_at, department_id, profile_photo_file_id
		FROM users
		WHERE id = $1
	`

	var user models.User
	var lastLoginAt *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.RoleType, &user.IsActive, &lastLoginAt,
		&user.DepartmentID, &user.ProfilePhotoFileID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("error retrieving user by ID: %w", err)
	}

	user.LastLoginAt = lastLoginAt

	return &user, nil
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking email existence: %w", err)
	}

	return exists, nil
}

// GetDepartmentNameByID retrieves department name by ID
func (r *UserRepository) GetDepartmentNameByID(ctx context.Context, departmentID int64) (string, error) {
	query := `
		SELECT name FROM departments WHERE id = $1
	`

	var name string
	err := r.db.QueryRow(ctx, query, departmentID).Scan(&name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.ErrDepartmentNotFound
		}
		return "", fmt.Errorf("error retrieving department name: %w", err)
	}

	return name, nil
}

// UpdateLastLogin updates the last login time
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	query := `
		UPDATE users SET last_login_at = NOW() WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("error updating last login time: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

// UpdateProfile updates a user's basic profile information
func (r *UserRepository) UpdateProfile(ctx context.Context, userID int64, firstName, lastName, email string) error {
	query := `
		UPDATE users 
		SET first_name = $2, last_name = $3, email = $4, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, firstName, lastName, email)
	if err != nil {
		// Check for duplicate email error
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
			return apperrors.ErrEmailAlreadyExists
		}
		return fmt.Errorf("error updating user profile: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

// UpdateProfilePhotoFileID updates the profile photo file ID for a given user
func (r *UserRepository) UpdateProfilePhotoFileID(ctx context.Context, userID int64, fileID *int64) error {
	query := `
		UPDATE users SET profile_photo_file_id = $2, updated_at = NOW() WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, fileID)
	if err != nil {
		return fmt.Errorf("error updating profile photo file ID: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	id, err := r.CreateUser(ctx, user)
	if err != nil {
		return err
	}
	user.ID = id
	return nil
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.GetUserByEmail(ctx, email)
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	return r.GetUserByID(ctx, id)
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = $2, first_name = $3, last_name = $4, role_type = $5, 
		is_active = $6, department_id = $7, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, user.ID, user.Email, user.FirstName,
		user.LastName, user.RoleType, user.IsActive, user.DepartmentID)

	if err != nil {
		// Check for duplicate email error
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"users_email_key\" (SQLSTATE 23505)" {
			return apperrors.ErrEmailAlreadyExists
		}
		return fmt.Errorf("error updating user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return apperrors.ErrUserNotFound
	}

	return nil
}

// CreateInstructor creates a new instructor
func (r *UserRepository) CreateInstructor(ctx context.Context, instructor *models.Instructor) error {
	query := `
		INSERT INTO instructors (user_id, title)
		VALUES ($1, $2)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, instructor.UserID, instructor.Title).Scan(&id)
	if err != nil {
		return fmt.Errorf("error creating instructor: %w", err)
	}

	instructor.ID = id
	return nil
}

// GetInstructorByUserID retrieves an instructor by user ID
func (r *UserRepository) GetInstructorByUserID(ctx context.Context, userID int64) (*models.Instructor, error) {
	query := `
		SELECT id, user_id, title
		FROM instructors 
		WHERE user_id = $1
	`

	var instructor models.Instructor
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&instructor.ID, &instructor.UserID, &instructor.Title,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("instructor not found for user ID %d", userID)
		}
		return nil, fmt.Errorf("error retrieving instructor: %w", err)
	}

	// Get the user information
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user for instructor: %w", err)
	}

	instructor.User = user
	return &instructor, nil
}

// updateInstructorTitleLegacy updates an instructor's title (legacy method)
// DEPRECATED: Use UpdateInstructorTitle instead which updates the users table directly
func (r *UserRepository) updateInstructorTitleLegacy(ctx context.Context, userID int64, newTitle string) error {
	query := `
		UPDATE instructors
		SET title = $2
		WHERE user_id = $1
	`

	result, err := r.db.Exec(ctx, query, userID, newTitle)
	if err != nil {
		return fmt.Errorf("error updating instructor title in legacy table: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("instructor not found")
	}

	return nil
}

// CreateStudent creates a new student
func (r *UserRepository) CreateStudent(ctx context.Context, student *models.Student) error {
	query := `
		INSERT INTO students (user_id, identifier, graduation_year)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(ctx, query, student.UserID, student.Identifier, student.GraduationYear).Scan(&id)
	if err != nil {
		// Check for duplicate identifier
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"students_identifier_key\" (SQLSTATE 23505)" {
			return apperrors.ErrIdentifierExists
		}
		return fmt.Errorf("error creating student: %w", err)
	}

	student.ID = id
	return nil
}

// GetStudentByUserID retrieves a student by user ID
func (r *UserRepository) GetStudentByUserID(ctx context.Context, userID int64) (*models.Student, error) {
	query := `
		SELECT id, user_id, identifier, graduation_year
		FROM students
		WHERE user_id = $1
	`

	var student models.Student
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&student.ID, &student.UserID, &student.Identifier, &student.GraduationYear,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("student not found for user ID %d", userID)
		}
		return nil, fmt.Errorf("error retrieving student: %w", err)
	}

	// Get the user information
	user, err := r.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving user for student: %w", err)
	}

	student.User = user
	return &student, nil
}

// IdentifierExists checks if a student identifier already exists
func (r *UserRepository) IdentifierExists(ctx context.Context, identifier string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM students WHERE identifier = $1
		)
	`

	var exists bool
	err := r.db.QueryRow(ctx, query, identifier).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking identifier existence: %w", err)
	}

	return exists, nil
}

// FindByID retrieves a user by ID (alias for GetByID - for new user service)
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	return r.GetByID(ctx, id)
}

// FindByEmail retrieves a user by email (alias for GetByEmail - for new user service)
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.GetByEmail(ctx, email)
}

// FindByDepartment retrieves users by department
func (r *UserRepository) FindByDepartment(ctx context.Context, departmentID int64) ([]*models.User, error) {
	query := `
		SELECT id, email, first_name, last_name, role_type, created_at, updated_at, 
		last_login_at, department_id, profile_photo_file_id, is_active
		FROM users
		WHERE department_id = $1
	`

	rows, err := r.db.Query(ctx, query, departmentID)
	if err != nil {
		return nil, fmt.Errorf("error querying users by department: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.RoleType,
			&user.CreatedAt, &user.UpdatedAt, &lastLoginAt, &user.DepartmentID,
			&user.ProfilePhotoFileID, &user.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

// FindByDepartmentAndRole retrieves users by department and role
func (r *UserRepository) FindByDepartmentAndRole(ctx context.Context, departmentID int64, role models.RoleType) ([]*models.User, error) {
	query := `
		SELECT id, email, first_name, last_name, role_type, created_at, updated_at, 
		last_login_at, department_id, profile_photo_file_id, is_active
		FROM users
		WHERE department_id = $1 AND role_type = $2
	`

	rows, err := r.db.Query(ctx, query, departmentID, role)
	if err != nil {
		return nil, fmt.Errorf("error querying users by department and role: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.RoleType,
			&user.CreatedAt, &user.UpdatedAt, &lastLoginAt, &user.DepartmentID,
			&user.ProfilePhotoFileID, &user.IsActive,
		)

		if err != nil {
			return nil, fmt.Errorf("error scanning user row: %w", err)
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, nil
}

// FindByFilter retrieves users based on filter criteria with pagination
func (r *UserRepository) FindByFilter(ctx context.Context, departmentID *int64, roleType *models.RoleType, email *string, name *string, page, pageSize int) ([]*models.User, int64, error) {
	// Base SQL
	baseSQL := `
		FROM users
		WHERE 1=1
	`

	// Build where clauses and params
	whereClause := ""
	var params []interface{}
	paramIndex := 1

	if departmentID != nil {
		whereClause += fmt.Sprintf(" AND department_id = $%d", paramIndex)
		params = append(params, *departmentID)
		paramIndex++
	}

	if roleType != nil {
		whereClause += fmt.Sprintf(" AND role_type = $%d", paramIndex)
		params = append(params, *roleType)
		paramIndex++
	}

	if email != nil {
		whereClause += fmt.Sprintf(" AND email ILIKE $%d", paramIndex)
		params = append(params, "%"+*email+"%")
		paramIndex++
	}

	if name != nil {
		whereClause += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d)", paramIndex, paramIndex+1)
		params = append(params, "%"+*name+"%", "%"+*name+"%")
		paramIndex += 2
	}

	// Count total records first
	countQuery := "SELECT COUNT(*) " + baseSQL + whereClause
	var total int64
	err := r.db.QueryRow(ctx, countQuery, params...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting users: %w", err)
	}

	// Now get paginated results
	offset := (page - 1) * pageSize
	query := `
		SELECT id, email, first_name, last_name, role_type, created_at, updated_at, 
		last_login_at, department_id, profile_photo_file_id, is_active
	` + baseSQL + whereClause + `
		ORDER BY id
		LIMIT $` + fmt.Sprintf("%d", paramIndex) + ` OFFSET $` + fmt.Sprintf("%d", paramIndex+1)

	params = append(params, pageSize, offset)

	rows, err := r.db.Query(ctx, query, params...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		var lastLoginAt sql.NullTime

		err := rows.Scan(
			&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.RoleType,
			&user.CreatedAt, &user.UpdatedAt, &lastLoginAt, &user.DepartmentID,
			&user.ProfilePhotoFileID, &user.IsActive,
		)

		if err != nil {
			return nil, 0, fmt.Errorf("error scanning user row: %w", err)
		}

		if lastLoginAt.Valid {
			user.LastLoginAt = &lastLoginAt.Time
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, total, nil
}

// SetEmailVerified updates the email_verified field for a user
func (r *UserRepository) SetEmailVerified(ctx context.Context, userID int64, verified bool) error {
	query := `
		UPDATE users 
		SET email_verified = $1 
		WHERE id = $2
	`

	_, err := r.db.Exec(ctx, query, verified, userID)
	if err != nil {
		return fmt.Errorf("error updating email verification status: %w", err)
	}

	return nil
}

// IsEmailVerified checks if a user's email is verified
func (r *UserRepository) IsEmailVerified(ctx context.Context, userID int64) (bool, error) {
	query := `
		SELECT email_verified 
		FROM users 
		WHERE id = $1
	`

	var verified bool
	err := r.db.QueryRow(ctx, query, userID).Scan(&verified)
	if err != nil {
		if err == pgx.ErrNoRows {
			return false, apperrors.ErrUserNotFound
		}
		return false, fmt.Errorf("error checking email verification status: %w", err)
	}

	return verified, nil
}
