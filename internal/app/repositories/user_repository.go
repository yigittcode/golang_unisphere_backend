package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories/user"
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
}

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

// UpdateProfile updates a user's basic profile information
func (r *UserRepository) UpdateProfile(ctx context.Context, userID int64, firstName, lastName, email string) error {
	return r.common.UpdateUserProfile(ctx, userID, firstName, lastName, email)
}

// UpdateProfilePhotoFileID updates the profile photo file ID for a given user
func (r *UserRepository) UpdateProfilePhotoFileID(ctx context.Context, userID int64, fileID *int64) error {
	return r.common.UpdateUserProfilePhotoFileID(ctx, userID, fileID)
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
	return r.common.DeleteUser(ctx, id)
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
	return r.common.UpdateUser(ctx, user)
}
