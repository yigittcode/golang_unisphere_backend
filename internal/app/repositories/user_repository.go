package repositories

import (
	"context"
	
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories/user"
)

// These errors have been moved to apperrors package
// Use apperrors.ErrUserNotFound, apperrors.ErrEmailAlreadyExists, and apperrors.ErrIdentifierExists instead

// UserRepository combines all user-related repositories
type UserRepository struct {
	common     *user.CommonRepository
	student    *user.StudentRepository
	instructor *user.InstructorRepository
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		common:     user.NewRepository(db),
		student:    user.NewStudentRepository(db),
		instructor: user.NewInstructorRepository(db),
	}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) (int64, error) {
	return r.common.CreateUser(ctx, user)
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return r.common.GetUserByEmail(ctx, email)
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return r.common.GetUserByID(ctx, id)
}

// EmailExists checks if an email already exists
func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.common.EmailExists(ctx, email)
}

// CreateStudent creates a new student
func (r *UserRepository) CreateStudent(ctx context.Context, student *models.Student) error {
	return r.student.CreateStudent(ctx, student)
}

// GetStudentByUserID retrieves a student by user ID
func (r *UserRepository) GetStudentByUserID(ctx context.Context, userID int64) (*models.Student, error) {
	return r.student.GetStudentByUserID(ctx, userID)
}

// IdentifierExists checks if a student identifier already exists
func (r *UserRepository) IdentifierExists(ctx context.Context, identifier string) (bool, error) {
	return r.student.IdentifierExists(ctx, identifier)
}

// CreateInstructor creates a new instructor
func (r *UserRepository) CreateInstructor(ctx context.Context, instructor *models.Instructor) error {
	return r.instructor.CreateInstructor(ctx, instructor)
}

// GetInstructorByUserID retrieves an instructor by user ID
func (r *UserRepository) GetInstructorByUserID(ctx context.Context, userID int64) (*models.Instructor, error) {
	return r.instructor.GetInstructorByUserID(ctx, userID)
}

// GetInstructorsByDepartmentID retrieves all instructors in a department
func (r *UserRepository) GetInstructorsByDepartmentID(ctx context.Context, departmentID int64) ([]*models.Instructor, error) {
	return r.instructor.GetInstructorsByDepartmentID(ctx, departmentID)
}

// UpdateInstructorTitle updates an instructor's title
func (r *UserRepository) UpdateInstructorTitle(ctx context.Context, userID int64, newTitle string) error {
	return r.instructor.UpdateInstructorTitle(ctx, userID, newTitle)
}

// GetDepartmentNameByID retrieves department name by ID
func (r *UserRepository) GetDepartmentNameByID(ctx context.Context, departmentID int64) (string, error) {
	return r.common.GetDepartmentNameByID(ctx, departmentID)
}

// UpdateLastLogin updates the last login time
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID int64) error {
	return r.common.UpdateLastLogin(ctx, userID)
}

// UpdateUserProfilePhotoURL updates only the profile photo URL for a given user.
func (r *UserRepository) UpdateUserProfilePhotoURL(ctx context.Context, userID int64, photoURL *string) error {
	// Delegate to the common repository
	return r.common.UpdateUserProfilePhotoURL(ctx, userID, photoURL)
}

// UpdateUserProfilePhotoFileID updates the profile photo file ID for a given user.
func (r *UserRepository) UpdateUserProfilePhotoFileID(ctx context.Context, userID int64, fileID *int64) error {
	// Delegate to the common repository
	return r.common.UpdateUserProfilePhotoFileID(ctx, userID, fileID)
}

// UpdateUserProfile updates a user's basic profile information
func (r *UserRepository) UpdateUserProfile(ctx context.Context, userID int64, firstName, lastName, email string) error {
	// Delegate to the common repository
	return r.common.UpdateUserProfile(ctx, userID, firstName, lastName, email)
}
