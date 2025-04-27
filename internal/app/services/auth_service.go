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

// AuthService handles authentication operations
type AuthService struct {
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
) *AuthService {
	return &AuthService{
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
func (s *AuthService) validateEmail(email string) error {
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
func (s *AuthService) validatePassword(password string) error {
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
func (s *AuthService) validateIdentifier(identifier string) error {
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
func (s *AuthService) validateUserID(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user ID must be positive", apperrors.ErrValidationFailed)
	}
	return nil
}

// validateToken validates a token string
func (s *AuthService) validateToken(token string) error {
	// Token should be non-empty
	if strings.TrimSpace(token) == "" {
		return apperrors.ErrTokenInvalid
	}

	return nil
}

// RegisterStudent registers a new student
func (s *AuthService) RegisterStudent(ctx context.Context, req *dto.RegisterStudentRequest) (*dto.TokenResponse, error) {
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

	// Create user
	user := &models.User{
		Email:        req.Email,
		Password:     string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		RoleType:     models.RoleStudent,
		IsActive:     true,
		DepartmentID: &req.DepartmentID,
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
		DepartmentID:   req.DepartmentID,
		GraduationYear: req.GraduationYear,
	}

	// Create student in DB
	err = s.userRepo.CreateStudent(ctx, student)
	if err != nil {
		return nil, fmt.Errorf("error creating student: %w", err)
	}

	// Generate token
	user.ID = userID // Set ID for token generation
	tokenResponse, err := s.generateTokenResponse(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	return tokenResponse, nil
}

// RegisterInstructor registers a new instructor
func (s *AuthService) RegisterInstructor(ctx context.Context, req *dto.RegisterInstructorRequest) (*dto.TokenResponse, error) {
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

	// Create user
	user := &models.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		RoleType:  models.RoleInstructor,
		IsActive:  true,
	}

	// Add user to database
	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("user creation error: %w", err)
	}

	// Create instructor
	instructor := &models.Instructor{
		UserID:       userID,
		DepartmentID: req.DepartmentID,
		Title:        req.Title,
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
func (s *AuthService) Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error) {
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
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*dto.TokenResponse, error) {
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
func (s *AuthService) GetProfile(ctx context.Context, userID int64) (*dto.UserProfile, error) {
	// Validate user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Get user from DB
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	// Create profile response
	profile := &dto.UserProfile{
		ID:                 user.ID,
		Email:              user.Email,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		RoleType:           string(user.RoleType),
		ProfilePhotoFileId: user.ProfilePhotoFileID,
	}

	var departmentID int64

	// Get role-specific information
	switch user.RoleType {
	case models.RoleStudent:
		// Get student details
		student, err := s.userRepo.GetStudentByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving student details: %w", err)
		}

		profile.Identifier = &student.Identifier
		profile.GraduationYear = student.GraduationYear
		departmentID = student.DepartmentID // Store department ID

	case models.RoleInstructor:
		// Get instructor details
		instructor, err := s.userRepo.GetInstructorByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving instructor details: %w", err)
		}

		profile.Title = &instructor.Title
		departmentID = instructor.DepartmentID // Store department ID
	}

	// Get department details (both roles have a department)
	if departmentID > 0 {
		// If user has department_id set, prioritize that
		if user.DepartmentID != nil {
			departmentID = *user.DepartmentID
		}

		profile.DepartmentID = departmentID

		// Get department name
		departmentName, err := s.userRepo.GetDepartmentNameByID(ctx, departmentID)
		if err == nil {
			profile.DepartmentName = departmentName
		}

		// Get faculty details through repository
		facultyDetails, err := s.departmentRepo.GetFacultyByDepartmentID(ctx, departmentID)
		if err == nil && facultyDetails != nil {
			profile.FacultyID = facultyDetails.ID
			profile.FacultyName = facultyDetails.Name
		}
	}

	return profile, nil
}

// UpdateProfilePhoto updates a user's profile photo
func (s *AuthService) UpdateProfilePhoto(ctx context.Context, userID int64, fileHeader *multipart.FileHeader) (*dto.UserProfile, error) {
	// Validate user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Verify user exists
	if _, err := s.userRepo.GetUserByID(ctx, userID); err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user information: %w", err)
	}

	// Validate file type - only allow image formats for profile photos
	if !s.isValidImageFile(fileHeader.Filename) {
		return nil, fmt.Errorf("%w: only image files (jpg, jpeg, png, gif) are allowed for profile photos", apperrors.ErrValidationFailed)
	}

	// We don't need to open the file manually, the SaveFile method will handle that
	newFilePath, err := s.fileStorage.SaveFile(fileHeader)
	if err != nil {
		return nil, fmt.Errorf("error saving file: %w", err)
	}

	// Create file record in database
	fileID, err := s.createFileRecord(ctx, fileHeader.Filename, newFilePath, "profile_photo")
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
func (s *AuthService) isValidImageFile(filename string) bool {
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
func (s *AuthService) createFileRecord(ctx context.Context, originalName, filePath, fileType string) (int64, error) {
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

	// Create the file record in the database
	file := &models.File{
		FileName: originalName,
		FilePath: filePath,
		FileURL:  s.fileStorage.GetFileURL(filePath),
		FileSize: fileSize,
		FileType: mimeType,
		// For profile photos, we're not setting ResourceType or ResourceID as they're referenced directly from user table
	}

	return s.fileRepo.CreateFile(ctx, file)
}

// DeleteProfilePhoto deletes a user's profile photo
func (s *AuthService) DeleteProfilePhoto(ctx context.Context, userID int64) (*dto.UserProfile, error) {
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

		// First set the user's profile photo file ID to NULL to remove the reference
		if err := s.userRepo.UpdateUserProfilePhotoFileID(ctx, userID, nil); err != nil {
			return nil, fmt.Errorf("error updating profile photo: %w", err)
		}

		// Then delete the file from the files table and filesystem
		if err := s.fileRepo.DeleteFile(ctx, fileID); err != nil {
			// Log the error but continue, as we've already removed the reference
			s.logger.Error().Err(err).Int64("fileID", fileID).Int64("userID", userID).Msg("Error deleting profile photo file")
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
func (s *AuthService) UpdateUserProfile(ctx context.Context, userID int64, req *dto.UpdateUserProfileRequest) (*dto.UserProfile, error) {
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
func (s *AuthService) generateTokenResponse(ctx context.Context, user *models.User) (*dto.TokenResponse, error) {
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
	return &dto.TokenResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresIn:        int64(expiresIn),        // Convert int to int64
		RefreshExpiresIn: int64(refreshExpiresIn), // Convert int to int64
	}, nil
}
