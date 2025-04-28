package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/rs/zerolog"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/auth"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
	"github.com/yigit/unisphere/internal/pkg/validation"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	RegisterStudent(ctx context.Context, req *dto.RegisterStudentRequest) (*dto.TokenResponse, error)
	RegisterInstructor(ctx context.Context, req *dto.RegisterInstructorRequest) (*dto.TokenResponse, error)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error)
	GetProfile(ctx context.Context, userID int64) (*dto.BaseUserProfile, error)
	UpdateProfilePhoto(ctx context.Context, userID int64, photo *multipart.FileHeader) (*dto.BaseUserProfile, error)
	DeleteProfilePhoto(ctx context.Context, userID int64) (*dto.BaseUserProfile, error)
	UpdateUserProfile(ctx context.Context, userID int64, req *dto.UpdateUserProfileRequest) (*dto.BaseUserProfile, error)
}

// authServiceImpl implements the AuthService interface
type authServiceImpl struct {
	userRepo       *repositories.UserRepository
	tokenRepo      *repositories.TokenRepository
	departmentRepo *repositories.DepartmentRepository
	facultyRepo    *repositories.FacultyRepository
	fileRepo       *repositories.FileRepository
	fileStorage    *filestorage.LocalStorage
	jwtService     *auth.JWTService
	logger         zerolog.Logger
}

// NewAuthService creates a new AuthService
func NewAuthService(
	userRepo *repositories.UserRepository,
	tokenRepo *repositories.TokenRepository,
	departmentRepo *repositories.DepartmentRepository,
	facultyRepo *repositories.FacultyRepository,
	fileRepo *repositories.FileRepository,
	fileStorage *filestorage.LocalStorage,
	jwtService *auth.JWTService,
	logger zerolog.Logger,
) AuthService {
	return &authServiceImpl{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		departmentRepo: departmentRepo,
		facultyRepo:    facultyRepo,
		fileRepo:       fileRepo,
		fileStorage:    fileStorage,
		jwtService:     jwtService,
		logger:         logger,
	}
}

// validateEmail validates an email address
func (s *authServiceImpl) validateEmail(email string) error {
	// Email should be non-empty
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("%w: email cannot be empty", apperrors.ErrValidationFailed)
	}

	// Email should have a valid format
	validator := validation.NewStringValidation(email).
		WithPattern(validation.CompiledPatterns.Email)

	if !validator.Validate() {
		return apperrors.ErrInvalidEmail
	}

	return nil
}

// validatePassword checks if password meets requirements
func (s *authServiceImpl) validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("%w: password cannot be empty", apperrors.ErrValidationFailed)
	}

	// Minimum uzunluk kontrolü
	validator := validation.NewStringValidation(password).
		WithMinLength(validation.PasswordMinLength)

	if !validator.Validate() {
		return fmt.Errorf("%w: password must be at least %d characters long",
			apperrors.ErrInvalidPassword, validation.PasswordMinLength)
	}

	// Check for at least one letter
	hasLetter := false
	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
			break
		}
	}
	if !hasLetter {
		return fmt.Errorf("%w: password must contain at least one letter", apperrors.ErrInvalidPassword)
	}

	// Check for at least one digit
	hasDigit := false
	for _, char := range password {
		if unicode.IsDigit(char) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return fmt.Errorf("%w: password must contain at least one digit", apperrors.ErrInvalidPassword)
	}

	return nil
}

// validateIdentifier validates a student identifier
func (s *authServiceImpl) validateIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("%w: student identifier cannot be empty", apperrors.ErrValidationFailed)
	}

	// Student identifier should match the pattern (8 digits)
	validator := validation.NewStringValidation(identifier).
		WithPattern(validation.CompiledPatterns.Identifier)

	if !validator.Validate() {
		return apperrors.ErrInvalidIdentifier
	}

	return nil
}

// validateUserID validates a user ID
func (s *authServiceImpl) validateUserID(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user ID must be positive", apperrors.ErrValidationFailed)
	}
	return nil
}

// validateToken validates a token string
func (s *authServiceImpl) validateToken(token string) error {
	// Token should be non-empty
	if strings.TrimSpace(token) == "" {
		return apperrors.ErrTokenInvalid
	}

	return nil
}

// RegisterStudent registers a new student
func (s *authServiceImpl) RegisterStudent(ctx context.Context, req *dto.RegisterStudentRequest) (*dto.TokenResponse, error) {
	// Validate email
	if err := s.validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Validate password
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Validate student identifier
	if err := s.validateIdentifier(req.StudentID); err != nil {
		return nil, err
	}

	// Check if student identifier already exists
	exists, err := s.userRepo.IdentifierExists(ctx, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("error checking if student identifier exists: %w", err)
	}
	if exists {
		return nil, apperrors.ErrIdentifierExists
	}

	// Check if email already exists
	exists, err = s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking if email exists: %w", err)
	}
	if exists {
		return nil, apperrors.ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user with department_id
	user := &models.User{
		Email:        req.Email,
		Password:     string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		RoleType:     models.RoleStudent,
		IsActive:     true,
		DepartmentID: &req.DepartmentID, // Department bilgisi sadece user tablosunda tutulacak
	}

	// Create user in DB
	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Create student
	student := &models.Student{
		UserID:         userID,
		Identifier:     req.StudentID,
		GraduationYear: req.GraduationYear,
	}

	// Create student in DB
	err = s.userRepo.CreateStudent(ctx, student)
	if err != nil {
		return nil, fmt.Errorf("error creating student: %w", err)
	}

	// Generate token
	user.ID = userID // Set ID for token generation
	return s.generateTokenResponse(ctx, user)
}

// RegisterInstructor registers a new instructor
func (s *authServiceImpl) RegisterInstructor(ctx context.Context, req *dto.RegisterInstructorRequest) (*dto.TokenResponse, error) {
	// Validate email
	if err := s.validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Validate password
	if err := s.validatePassword(req.Password); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking if email exists: %w", err)
	}
	if exists {
		return nil, apperrors.ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	// Create user with department_id
	user := &models.User{
		Email:        req.Email,
		Password:     string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		RoleType:     models.RoleInstructor,
		IsActive:     true,
		DepartmentID: &req.DepartmentID,
	}

	// Add user to database
	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("user creation error: %w", err)
	}

	// Create instructor
	instructor := &models.Instructor{
		UserID: userID,
		Title:  req.Title,
	}

	// Add instructor to database
	if err := s.userRepo.CreateInstructor(ctx, instructor); err != nil {
		return nil, fmt.Errorf("instructor creation error: %w", err)
	}

	// Add User ID
	user.ID = userID

	// Generate token
	return s.generateTokenResponse(ctx, user)
}

// Login authenticates a user
func (s *authServiceImpl) Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error) {
	// Validate email
	if err := s.validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Validate password format (not content)
	if req.Password == "" {
		return nil, fmt.Errorf("%w: password cannot be empty", apperrors.ErrValidationFailed)
	}

	// Find user by email
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Password validation
	if !auth.CheckPassword(user.Password, req.Password) {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Generate token
	return s.generateTokenResponse(ctx, user)
}

// RefreshToken creates a new access token using a refresh token
func (s *authServiceImpl) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
	// Validate refresh token
	if err := s.validateToken(refreshToken); err != nil {
		return nil, err
	}

	// Get token information with additional validation
	userID, expiryDate, isRevoked, err := s.tokenRepo.GetTokenByValue(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, apperrors.ErrTokenNotFound) {
			return nil, apperrors.ErrTokenNotFound
		}
		if errors.Is(err, apperrors.ErrTokenExpired) {
			return nil, apperrors.ErrTokenExpired
		}
		if errors.Is(err, apperrors.ErrTokenRevoked) {
			return nil, apperrors.ErrTokenRevoked
		}
		return nil, fmt.Errorf("token validation error: %w", err)
	}

	// Additional security checks
	// 1. Check expiry date explicitly
	if expiryDate.Before(time.Now()) {
		// Also revoke expired token
		_ = s.tokenRepo.RevokeToken(ctx, refreshToken)
		return nil, apperrors.ErrTokenExpired
	}

	// 2. Check revocation status explicitly
	if isRevoked {
		return nil, apperrors.ErrTokenRevoked
	}

	// Get user information
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Revoke old token (important for security - prevents token reuse)
	if err := s.tokenRepo.RevokeToken(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Generate new token
	return s.generateTokenResponse(ctx, user)
}

// GetProfile retrieves a user's profile
func (s *authServiceImpl) GetProfile(ctx context.Context, userID int64) (*dto.BaseUserProfile, error) {
	// Validate user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Get user from DB
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	// Create base profile
	baseProfile := &dto.BaseUserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		RoleType:  string(user.RoleType),
	}

	// Get faculty and department information if available
	if user.DepartmentID != nil {
		departmentID := *user.DepartmentID

		// Get department details
		departmentName, err := s.userRepo.GetDepartmentNameByID(ctx, departmentID)
		if err == nil {
			// Get faculty details through repository
			facultyDetails, err := s.departmentRepo.GetFacultyByDepartmentID(ctx, departmentID)
			if err == nil && facultyDetails != nil {
				// Create department info
				department := &dto.DepartmentInfo{
					ID:   departmentID,
					Name: departmentName,
				}

				// Create faculty info with department
				baseProfile.Faculty = &dto.FacultyInfo{
					ID:         facultyDetails.ID,
					Name:       facultyDetails.Name,
					Department: department,
				}
			}
		}
	}

	// Add profile photo information if available
	if user.ProfilePhotoFileID != nil {
		fileID := *user.ProfilePhotoFileID

		file, err := s.fileRepo.GetFileByID(ctx, fileID)
		if err == nil {
			baseProfile.Photo = &dto.PhotoInfo{
				ID:       fileID,
				URL:      file.FileURL,
				FileType: file.FileType,
			}
		}
	}

	// Get role-specific information
	switch user.RoleType {
	case models.RoleStudent:
		// Get student details
		student, err := s.userRepo.GetStudentByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving student details: %w", err)
		}

		studentProfile := &dto.StudentProfile{
			BaseUserProfile: *baseProfile,
			Identifier:      student.Identifier,
			GraduationYear:  student.GraduationYear,
		}
		return &studentProfile.BaseUserProfile, nil

	case models.RoleInstructor:
		// Get instructor details
		instructor, err := s.userRepo.GetInstructorByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving instructor details: %w", err)
		}

		instructorProfile := &dto.InstructorProfile{
			BaseUserProfile: *baseProfile,
			Title:           instructor.Title,
		}
		return &instructorProfile.BaseUserProfile, nil
	}

	return baseProfile, nil
}

// UpdateProfilePhoto updates a user's profile photo
func (s *authServiceImpl) UpdateProfilePhoto(ctx context.Context, userID int64, fileHeader *multipart.FileHeader) (*dto.BaseUserProfile, error) {
	// Validate user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Get the user to check if they exist and to find any existing profile photo
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user information: %w", err)
	}

	// Validate file type - only allow image formats for profile photos
	if !s.isValidImageFile(fileHeader.Filename) {
		return nil, fmt.Errorf("%w: only image files (jpg, jpeg, png, gif) are allowed for profile photos", apperrors.ErrValidationFailed)
	}

	// Check if user already has a profile photo and delete it
	if user.ProfilePhotoFileID != nil {
		// Get old file ID
		oldFileID := *user.ProfilePhotoFileID

		// Get file details to have the file path for physical deletion
		oldFile, err := s.fileRepo.GetFileByID(ctx, oldFileID)
		if err == nil && oldFile != nil {
			// Store the file path for deletion
			oldFilePath := oldFile.FilePath

			// First set the user's profile photo file ID to NULL to remove the reference
			if err := s.userRepo.UpdateUserProfilePhotoFileID(ctx, userID, nil); err != nil {
				return nil, fmt.Errorf("error removing old profile photo reference: %w", err)
			}

			// Delete the file from the database
			if err := s.fileRepo.DeleteFile(ctx, oldFileID); err != nil {
				// Log the error but continue
				s.logger.Error().Err(err).Int64("fileID", oldFileID).Int64("userID", userID).Msg("Error deleting old profile photo from database")
			}

			// Delete the physical file
			if err := s.fileStorage.DeleteFile(oldFilePath); err != nil {
				s.logger.Error().Err(err).Str("filePath", oldFilePath).Int64("userID", userID).Msg("Error deleting old profile photo file from filesystem")
			}
		} else {
			// If we couldn't get file details, still try to remove DB reference
			if err := s.userRepo.UpdateUserProfilePhotoFileID(ctx, userID, nil); err != nil {
				return nil, fmt.Errorf("error removing old profile photo reference: %w", err)
			}

			// And try to delete the DB record
			if err := s.fileRepo.DeleteFile(ctx, oldFileID); err != nil {
				s.logger.Error().Err(err).Int64("fileID", oldFileID).Int64("userID", userID).Msg("Error deleting old profile photo record")
			}
		}
	}

	// We don't need to open the file manually, the SaveFile method will handle that
	newFilePath, err := s.fileStorage.SaveFile(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("error saving file: %w", err)
	}

	// Create file record in database
	fileID, err := s.createFileRecord(ctx, fileHeader.Filename, newFilePath, "profile_photo", userID)
	if err != nil {
		// Try to delete the file if the database insertion fails
		_ = s.fileStorage.DeleteFile(newFilePath)
		return nil, fmt.Errorf("error creating file record: %w", err)
	}

	// Update user's profile photo file ID with the new file ID
	if err := s.userRepo.UpdateUserProfilePhotoFileID(ctx, userID, &fileID); err != nil {
		// Try to delete the file if the update fails
		_ = s.fileStorage.DeleteFile(newFilePath)
		return nil, fmt.Errorf("error updating profile photo: %w", err)
	}

	// Return updated profile
	return s.GetProfile(ctx, userID)
}

// Helper method to check if file is a valid image
func (s *authServiceImpl) isValidImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
	}
	return validExtensions[ext]
}

// Helper method to create a file record in the database
func (s *authServiceImpl) createFileRecord(ctx context.Context, originalName, filePath, fileType string, userID int64) (int64, error) {
	// Determine MIME type based on file extension
	ext := strings.ToLower(filepath.Ext(originalName))
	var mimeType string

	switch ext {
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".png":
		mimeType = "image/png"
	case ".gif":
		mimeType = "image/gif"
	case ".pdf":
		mimeType = "application/pdf"
	default:
		mimeType = "application/octet-stream" // Default
	}

	// For file size, we need to get the actual file size
	// This assumes filePath contains the path relative to uploads directory
	baseDir := "./uploads" // This should match your storage configuration
	fullPath := filepath.Join(baseDir, filepath.Base(filePath))

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return 0, fmt.Errorf("error getting file info: %w", err)
	}

	fileSize := fileInfo.Size()

	// Set resource type based on the fileType parameter
	var resourceType models.FileType
	var resourceID int64

	if fileType == "profile_photo" {
		resourceType = models.FileTypeProfilePhoto
		resourceID = userID // For profile photos, resourceID is the user's ID
	}

	// Create the file record in the database
	file := &models.File{
		FileName:     originalName,
		FilePath:     filePath,
		FileURL:      s.fileStorage.GetFileURL(filePath),
		FileSize:     fileSize,
		FileType:     mimeType,
		UploadedBy:   userID,
		ResourceType: resourceType,
		ResourceID:   resourceID,
	}

	return s.fileRepo.CreateFile(ctx, file)
}

// DeleteProfilePhoto deletes a user's profile photo
func (s *authServiceImpl) DeleteProfilePhoto(ctx context.Context, userID int64) (*dto.BaseUserProfile, error) {
	// Validate user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Get the user to access the profilePhotoFileID
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user information: %w", err)
	}

	// If there's a profile photo file, delete it from the files table and filesystem
	if user.ProfilePhotoFileID != nil {
		fileID := *user.ProfilePhotoFileID

		// Get file details to have the file path
		file, err := s.fileRepo.GetFileByID(ctx, fileID)
		if err == nil && file != nil {
			// Get the file path to delete the physical file
			filePath := file.FilePath

			// First set the user's profile photo file ID to NULL to remove the reference
			if err := s.userRepo.UpdateUserProfilePhotoFileID(ctx, userID, nil); err != nil {
				return nil, fmt.Errorf("error updating profile photo: %w", err)
			}

			// Delete the file from the database
			if err := s.fileRepo.DeleteFile(ctx, fileID); err != nil {
				// Log the error but continue, as we've already removed the reference
				s.logger.Error().Err(err).Int64("fileID", fileID).Int64("userID", userID).Msg("Error deleting profile photo from database")
			}

			// Also delete from filesystem using fileStorage
			if err := s.fileStorage.DeleteFile(filePath); err != nil {
				s.logger.Error().Err(err).Str("filePath", filePath).Int64("userID", userID).Msg("Error deleting profile photo file from filesystem")
			}
		} else {
			// If we couldn't get file details, still try to remove DB reference
			if err := s.userRepo.UpdateUserProfilePhotoFileID(ctx, userID, nil); err != nil {
				return nil, fmt.Errorf("error updating profile photo: %w", err)
			}

			// And try to delete the DB record
			if err := s.fileRepo.DeleteFile(ctx, fileID); err != nil {
				s.logger.Error().Err(err).Int64("fileID", fileID).Int64("userID", userID).Msg("Error deleting profile photo record")
			}
		}
	} else {
		// If there's no profile photo file ID, just update to make sure it's NULL
		if err := s.userRepo.UpdateUserProfilePhotoFileID(ctx, userID, nil); err != nil {
			return nil, fmt.Errorf("error updating profile photo: %w", err)
		}
	}

	// Return updated profile
	return s.GetProfile(ctx, userID)
}

// UpdateUserProfile updates a user's profile information
func (s *authServiceImpl) UpdateUserProfile(ctx context.Context, userID int64, req *dto.UpdateUserProfileRequest) (*dto.BaseUserProfile, error) {
	// Validate user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Validate email
	if err := s.validateEmail(req.Email); err != nil {
		return nil, err
	}

	// Validate first name
	firstName := strings.TrimSpace(req.FirstName)
	firstNameValidator := validation.NewStringValidation(firstName).
		WithMinLength(validation.NameMinLength).
		WithMaxLength(validation.NameMaxLength)

	if !firstNameValidator.Validate() {
		if firstName == "" {
			return nil, fmt.Errorf("%w: first name cannot be empty", apperrors.ErrValidationFailed)
		} else if len(firstName) < validation.NameMinLength {
			return nil, fmt.Errorf("%w: first name must be at least %d characters",
				apperrors.ErrValidationFailed, validation.NameMinLength)
		} else {
			return nil, fmt.Errorf("%w: first name cannot exceed %d characters",
				apperrors.ErrValidationFailed, validation.NameMaxLength)
		}
	}

	// Validate last name
	lastName := strings.TrimSpace(req.LastName)
	lastNameValidator := validation.NewStringValidation(lastName).
		WithMinLength(validation.NameMinLength).
		WithMaxLength(validation.NameMaxLength)

	if !lastNameValidator.Validate() {
		if lastName == "" {
			return nil, fmt.Errorf("%w: last name cannot be empty", apperrors.ErrValidationFailed)
		} else if len(lastName) < validation.NameMinLength {
			return nil, fmt.Errorf("%w: last name must be at least %d characters",
				apperrors.ErrValidationFailed, validation.NameMinLength)
		} else {
			return nil, fmt.Errorf("%w: last name cannot exceed %d characters",
				apperrors.ErrValidationFailed, validation.NameMaxLength)
		}
	}

	// Check if the email is different from the current one and already exists
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get current user information: %w", err)
	}

	// Check if new email already exists (only if email is being changed)
	if user.Email != req.Email {
		exists, err := s.userRepo.EmailExists(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("error checking if email exists: %w", err)
		}
		if exists {
			return nil, apperrors.ErrEmailAlreadyExists
		}
	}

	// Update the user profile
	err = s.userRepo.UpdateUserProfile(ctx, userID, firstName, lastName, req.Email)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		if errors.Is(err, apperrors.ErrEmailAlreadyExists) {
			return nil, apperrors.ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	// Fetch the updated profile to return
	return s.GetProfile(ctx, userID)
}

// Helper functions

// generateTokenResponse creates token response
func (s *authServiceImpl) generateTokenResponse(ctx context.Context, user *models.User) (*dto.TokenResponse, error) {
	// Create access and refresh token pair
	accessToken, refreshToken, expiresIn, refreshExpiresIn, err := s.jwtService.GenerateTokenPair(user)
	if err != nil {
		return nil, fmt.Errorf("token generation error: %w", err)
	}

	// Refresh token expiry
	tokenExpiry := s.jwtService.GetRefreshTokenExpiry()

	// Save refresh token to database
	if err := s.tokenRepo.CreateToken(ctx, refreshToken, user.ID, tokenExpiry); err != nil {
		return nil, fmt.Errorf("token saving error: %w", err)
	}

	// Create token response
	tokenResponse := &dto.TokenResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        int64(expiresIn),
		RefreshExpiresIn: int64(refreshExpiresIn),
	}

	return tokenResponse, nil
}
