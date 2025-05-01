package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
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

// AuthService defines the interface for authentication-related operations
type AuthService interface {
	// User registration
	RegisterStudent(ctx context.Context, req *dto.RegisterStudentRequest) (*dto.TokenResponse, error)
	RegisterInstructor(ctx context.Context, req *dto.RegisterInstructorRequest) (*dto.TokenResponse, error)

	// Authentication
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error)
	RefreshToken(ctx context.Context, token string) (*dto.TokenResponse, error)

	// User profile
	GetProfile(ctx context.Context, userID int64) (*dto.UserResponse, error)
	GetUserByID(ctx context.Context, userID int64) (*models.User, error)
	UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) error
	UpdateProfilePhoto(ctx context.Context, userID int64, file *multipart.FileHeader) error
	DeleteProfilePhoto(ctx context.Context, userID int64) error

	// Validation
	ValidatePassword(password string) error
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
func (s *authServiceImpl) ValidatePassword(password string) error {
	// Password should be non-empty
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("%w: password cannot be empty", apperrors.ErrValidationFailed)
	}

	// Password should be at least 8 characters long
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters long", apperrors.ErrValidationFailed)
	}

	// Password should contain at least one uppercase letter
	hasUpper := false
	for _, c := range password {
		if unicode.IsUpper(c) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return fmt.Errorf("%w: password must contain at least one uppercase letter", apperrors.ErrValidationFailed)
	}

	// Password should contain at least one lowercase letter
	hasLower := false
	for _, c := range password {
		if unicode.IsLower(c) {
			hasLower = true
			break
		}
	}
	if !hasLower {
		return fmt.Errorf("%w: password must contain at least one lowercase letter", apperrors.ErrValidationFailed)
	}

	// Password should contain at least one digit
	hasDigit := false
	for _, c := range password {
		if unicode.IsDigit(c) {
			hasDigit = true
			break
		}
	}
	if !hasDigit {
		return fmt.Errorf("%w: password must contain at least one digit", apperrors.ErrValidationFailed)
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
		return apperrors.ErrValidationFailed
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
	if err := s.ValidatePassword(req.Password); err != nil {
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
	if err := s.ValidatePassword(req.Password); err != nil {
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

	// Create user in DB
	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	// Create instructor
	instructor := &models.Instructor{
		UserID: userID,
		Title:  req.Title,
	}

	// Create instructor in DB
	err = s.userRepo.CreateInstructor(ctx, instructor)
	if err != nil {
		return nil, fmt.Errorf("error creating instructor: %w", err)
	}

	// Generate token
	user.ID = userID // Set ID for token generation
	return s.generateTokenResponse(ctx, user)
}

// Login handles user login
func (s *authServiceImpl) Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Check if user is active
	if !user.IsActive {
		return nil, apperrors.ErrAccountDisabled
	}

	// Update last login time
	err = s.userRepo.UpdateLastLogin(ctx, user.ID)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Failed to update last login time")
		// Don't return error, as login was successful
	}

	// Generate and return token
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
func (s *authServiceImpl) GetProfile(ctx context.Context, userID int64) (*dto.UserResponse, error) {
	// Validate user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Get user from DB
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	// Get profile photo URL if exists
	var profilePhotoURL string
	if user.ProfilePhotoFileID != nil {
		file, err := s.fileRepo.GetByID(ctx, *user.ProfilePhotoFileID)
		if err == nil && file != nil { // Don't fail if photo not found
			profilePhotoURL = file.FileURL
		}
	}

	// Create base response
	response := &dto.UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Role:            string(user.RoleType),
		DepartmentID:    user.DepartmentID,
		ProfilePhotoURL: profilePhotoURL,
	}

	// Get role-specific information
	switch user.RoleType {
	case models.RoleStudent:
		// Get student details
		student, err := s.userRepo.GetStudentByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving student details: %w", err)
		}

		studentResponse := &dto.StudentResponse{
			UserResponse:   *response,
			StudentID:      student.Identifier,
			GraduationYear: student.GraduationYear,
		}
		return &studentResponse.UserResponse, nil

	case models.RoleInstructor:
		// Get instructor details
		instructor, err := s.userRepo.GetInstructorByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("error retrieving instructor details: %w", err)
		}

		instructorResponse := &dto.InstructorResponse{
			UserResponse: *response,
			Title:        instructor.Title,
		}
		return &instructorResponse.UserResponse, nil
	}

	return response, nil
}

// GetUserByID retrieves a user by their ID
func (s *authServiceImpl) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateProfile updates a user's profile information
func (s *authServiceImpl) UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.FirstName = req.FirstName
	user.LastName = req.LastName
	user.Email = req.Email

	return s.userRepo.UpdateProfile(ctx, userID, req.FirstName, req.LastName, req.Email)
}

// UpdateProfilePhoto updates a user's profile photo
func (s *authServiceImpl) UpdateProfilePhoto(ctx context.Context, userID int64, file *multipart.FileHeader) error {
	// Validate file
	if file == nil {
		return fmt.Errorf("%w: no file provided", apperrors.ErrValidationFailed)
	}

	// Check file size (max 5MB)
	if file.Size > 5*1024*1024 {
		return fmt.Errorf("%w: file size exceeds 5MB limit", apperrors.ErrValidationFailed)
	}

	// Check file type
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("%w: invalid file type. Only images are allowed", apperrors.ErrValidationFailed)
	}

	s.logger.Debug().
		Int64("userID", userID).
		Str("fileName", file.Filename).
		Int64("fileSize", file.Size).
		Str("contentType", contentType).
		Msg("Starting profile photo upload")

	// Generate unique filename
	filename := fmt.Sprintf("profile_photo_%d_%d%s", userID, time.Now().Unix(), filepath.Ext(file.Filename))

	// Save file using fileStorage
	fileURL, err := s.fileStorage.SaveFileWithPath(file, "profile_photos")
	if err != nil {
		s.logger.Error().Err(err).
			Str("filename", filename).
			Msg("Failed to save file to storage")
		return fmt.Errorf("failed to save profile photo: %w", err)
	}

	s.logger.Debug().
		Str("fileURL", fileURL).
		Str("filename", filename).
		Msg("File saved successfully")

	// Extract relative path from URL
	relativeFilePath := strings.TrimPrefix(fileURL, s.fileStorage.GetBaseURL())
	relativeFilePath = strings.TrimPrefix(relativeFilePath, "/uploads/")

	// Create file record in database
	fileRecord := &models.File{
		FileName:     filename,
		FilePath:     relativeFilePath,
		FileURL:      fileURL,
		FileSize:     file.Size,
		FileType:     contentType,
		ResourceType: models.FileTypeProfilePhoto,
		ResourceID:   userID,
		UploadedBy:   userID,
	}

	s.logger.Debug().
		Interface("fileRecord", fileRecord).
		Msg("Attempting to create file record in database")

	fileID, err := s.fileRepo.Create(ctx, fileRecord)
	if err != nil {
		s.logger.Error().Err(err).
			Interface("fileRecord", fileRecord).
			Msg("Failed to create file record in database")
		// If DB save fails, try to delete the physical file
		if delErr := s.fileStorage.DeleteFile(fileRecord.FilePath); delErr != nil {
			s.logger.Error().Err(delErr).
				Str("filePath", fileRecord.FilePath).
				Msg("Failed to delete file after database error")
		}
		return fmt.Errorf("failed to save file record: %w", err)
	}

	s.logger.Debug().
		Int64("fileID", fileID).
		Msg("File record created successfully")

	// Get old profile photo ID if exists
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("userID", userID).
			Msg("Failed to get user details")
		// If we can't get the user, clean up the new file and return error
		if delErr := s.fileStorage.DeleteFile(fileRecord.FilePath); delErr != nil {
			s.logger.Error().Err(delErr).
				Str("filePath", fileRecord.FilePath).
				Msg("Failed to delete file after user fetch error")
		}
		if delErr := s.fileRepo.Delete(ctx, fileID); delErr != nil {
			s.logger.Error().Err(delErr).
				Int64("fileID", fileID).
				Msg("Failed to delete file record after user fetch error")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update user's profile photo ID
	err = s.userRepo.UpdateProfilePhotoFileID(ctx, userID, &fileID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("userID", userID).
			Int64("fileID", fileID).
			Msg("Failed to update user's profile photo ID")
		// If user update fails, try to delete both the physical file and the file record
		if delErr := s.fileStorage.DeleteFile(fileRecord.FilePath); delErr != nil {
			s.logger.Error().Err(delErr).
				Str("filePath", fileRecord.FilePath).
				Msg("Failed to delete file after profile update error")
		}
		if delErr := s.fileRepo.Delete(ctx, fileID); delErr != nil {
			s.logger.Error().Err(delErr).
				Int64("fileID", fileID).
				Msg("Failed to delete file record after profile update error")
		}
		return fmt.Errorf("failed to update user's profile photo: %w", err)
	}

	s.logger.Debug().
		Int64("userID", userID).
		Int64("fileID", fileID).
		Msg("Profile photo ID updated successfully")

	// If user had an old profile photo, delete it
	if user.ProfilePhotoFileID != nil {
		oldFile, err := s.fileRepo.GetByID(ctx, *user.ProfilePhotoFileID)
		if err != nil {
			s.logger.Warn().Err(err).
				Int64("oldFileID", *user.ProfilePhotoFileID).
				Msg("Failed to get old profile photo details")
		} else if oldFile != nil {
			if delErr := s.fileStorage.DeleteFile(oldFile.FilePath); delErr != nil {
				s.logger.Warn().Err(delErr).
					Str("oldFilePath", oldFile.FilePath).
					Msg("Failed to delete old profile photo file")
			}
			if delErr := s.fileRepo.Delete(ctx, *user.ProfilePhotoFileID); delErr != nil {
				s.logger.Warn().Err(delErr).
					Int64("oldFileID", *user.ProfilePhotoFileID).
					Msg("Failed to delete old profile photo record")
			}
			s.logger.Info().
				Int64("oldFileID", *user.ProfilePhotoFileID).
				Msg("Old profile photo deleted successfully")
		}
	}

	s.logger.Info().
		Int64("userID", userID).
		Int64("fileID", fileID).
		Str("fileURL", fileURL).
		Msg("Profile photo updated successfully")

	return nil
}

// DeleteProfilePhoto deletes a user's profile photo
func (s *authServiceImpl) DeleteProfilePhoto(ctx context.Context, userID int64) error {
	// Get user to check if they have a profile photo
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("userID", userID).
			Msg("Failed to get user details")
		return fmt.Errorf("failed to get user: %w", err)
	}

	// If user has no profile photo, return success
	if user.ProfilePhotoFileID == nil {
		s.logger.Info().
			Int64("userID", userID).
			Msg("User has no profile photo to delete")
		return nil
	}

	// Get file details
	file, err := s.fileRepo.GetByID(ctx, *user.ProfilePhotoFileID)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("fileID", *user.ProfilePhotoFileID).
			Msg("Failed to get file details")
		return fmt.Errorf("failed to get file details: %w", err)
	}

	// Delete physical file
	if err := s.fileStorage.DeleteFile(file.FilePath); err != nil {
		s.logger.Error().Err(err).
			Str("filePath", file.FilePath).
			Msg("Failed to delete physical file")
		// Continue with database cleanup even if file deletion fails
	}

	// Delete file record from database
	if err := s.fileRepo.Delete(ctx, *user.ProfilePhotoFileID); err != nil {
		s.logger.Error().Err(err).
			Int64("fileID", *user.ProfilePhotoFileID).
			Msg("Failed to delete file record")
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	// Update user's profile photo ID to null
	err = s.userRepo.UpdateProfilePhotoFileID(ctx, userID, nil)
	if err != nil {
		s.logger.Error().Err(err).
			Int64("userID", userID).
			Msg("Failed to update user's profile photo ID")
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	s.logger.Info().
		Int64("userID", userID).
		Int64("fileID", *user.ProfilePhotoFileID).
		Msg("Profile photo deleted successfully")

	return nil
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
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		ExpiresIn:             int64(expiresIn),
		RefreshTokenExpiresIn: int64(refreshExpiresIn),
	}

	return tokenResponse, nil
}
