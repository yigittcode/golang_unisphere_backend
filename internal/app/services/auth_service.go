package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
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
	GetUserByID(ctx context.Context, userID int64) (*models.User, error)
	UpdateProfile(ctx context.Context, userID int64, req *dto.UpdateProfileRequest) error
	UpdateProfilePhoto(ctx context.Context, userID int64, file *multipart.FileHeader) error
	DeleteProfilePhoto(ctx context.Context, userID int64) error
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

	// Create base response
	response := &dto.UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Role:         string(user.RoleType),
		DepartmentID: user.DepartmentID,
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

		teacherResponse := &dto.TeacherResponse{
			UserResponse: *response,
			Title:        instructor.Title,
		}
		return &teacherResponse.UserResponse, nil
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
	// TODO: Implement photo upload logic
	return nil
}

// DeleteProfilePhoto deletes a user's profile photo
func (s *authServiceImpl) DeleteProfilePhoto(ctx context.Context, userID int64) error {
	// TODO: Implement photo deletion logic
	return nil
}

// Helper functions

// generateTokenResponse creates token response
func (s *authServiceImpl) generateTokenResponse(ctx context.Context, user *models.User) (*dto.TokenResponse, error) {
	// Create access and refresh token pair
	accessToken, refreshToken, expiresIn, _, err := s.jwtService.GenerateTokenPair(user)
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
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(expiresIn),
	}

	return tokenResponse, nil
}
