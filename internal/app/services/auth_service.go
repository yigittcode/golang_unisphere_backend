package services

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/rs/zerolog"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/auth"
	"golang.org/x/crypto/bcrypt"
)

// Define custom error types for auth service
var (
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrInvalidPassword        = errors.New("invalid password format")
	ErrInvalidStudentID       = errors.New("invalid student ID format")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrStudentIDAlreadyExists = errors.New("student ID already in use")
	ErrTokenNotFound          = errors.New("token not found")
	ErrTokenExpired           = errors.New("token has expired")
	ErrTokenRevoked           = errors.New("token has been revoked")
	ErrTokenInvalid           = errors.New("invalid token")
	ErrUserNotFound           = errors.New("user not found")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrAuthValidation         = errors.New("auth validation failed")
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo       *repositories.UserRepository
	tokenRepo      *repositories.TokenRepository
	departmentRepo *repositories.DepartmentRepository
	facultyRepo    *repositories.FacultyRepository
	jwtService     *auth.JWTService
	logger         zerolog.Logger
}

// NewAuthService creates a new AuthService
func NewAuthService(
	userRepo *repositories.UserRepository,
	tokenRepo *repositories.TokenRepository,
	departmentRepo *repositories.DepartmentRepository,
	facultyRepo *repositories.FacultyRepository,
	jwtService *auth.JWTService,
	logger zerolog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		departmentRepo: departmentRepo,
		facultyRepo:    facultyRepo,
		jwtService:     jwtService,
		logger:         logger,
	}
}

// validateEmail validates an email address
func (s *AuthService) validateEmail(email string) error {
	// Email should be non-empty
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("%w: email cannot be empty", ErrAuthValidation)
	}

	// Email should have a valid format
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}

	return nil
}

// validatePassword checks if password meets requirements
func (s *AuthService) validatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("%w: password cannot be empty", ErrAuthValidation)
	}

	// Check length
	if len(password) < 8 {
		return fmt.Errorf("%w: password must be at least 8 characters long", ErrInvalidPassword)
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
		return fmt.Errorf("%w: password must contain at least one letter", ErrInvalidPassword)
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
		return fmt.Errorf("%w: password must contain at least one digit", ErrInvalidPassword)
	}

	return nil
}

// validateStudentID validates a student ID
func (s *AuthService) validateStudentID(studentID string) error {
	if studentID == "" {
		return fmt.Errorf("%w: student ID cannot be empty", ErrAuthValidation)
	}

	// Student ID should match the pattern (8 digits)
	studentIDRegex := regexp.MustCompile(`^\d{8}$`)
	if !studentIDRegex.MatchString(studentID) {
		return ErrInvalidStudentID
	}

	return nil
}

// validateUserID validates a user ID
func (s *AuthService) validateUserID(userID int64) error {
	if userID <= 0 {
		return fmt.Errorf("%w: user ID must be positive", ErrAuthValidation)
	}
	return nil
}

// validateToken validates a token string
func (s *AuthService) validateToken(token string) error {
	// Token should be non-empty
	if strings.TrimSpace(token) == "" {
		return ErrTokenInvalid
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

	// Validate student ID
	if err := s.validateStudentID(req.StudentID); err != nil {
		return nil, err
	}

	// Check if student ID already exists
	exists, err := s.userRepo.StudentIDExists(ctx, req.StudentID)
	if err != nil {
		return nil, fmt.Errorf("error checking if student ID exists: %w", err)
	}
	if exists {
		return nil, ErrStudentIDAlreadyExists
	}

	// Check if email already exists
	exists, err = s.userRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("error checking if email exists: %w", err)
	}
	if exists {
		return nil, ErrEmailAlreadyExists
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
		RoleType:  models.RoleStudent,
		IsActive:  true,
	}

	// Add user to database
	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("user creation error: %w", err)
	}

	// Create student
	student := &models.Student{
		UserID:         userID,
		StudentID:      req.StudentID,
		DepartmentID:   req.DepartmentID,
		GraduationYear: req.GraduationYear,
	}

	// Add student to database
	if err := s.userRepo.CreateStudent(ctx, student); err != nil {
		return nil, fmt.Errorf("student creation error: %w", err)
	}

	// Add User ID
	user.ID = userID

	// Generate token
	return s.generateTokenResponse(ctx, user)
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
		return nil, ErrEmailAlreadyExists
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
		return nil, fmt.Errorf("%w: password cannot be empty", ErrAuthValidation)
	}

	// Find user by email
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Password validation
	if !auth.CheckPassword(user.Password, req.Password) {
		return nil, ErrInvalidCredentials
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
		if errors.Is(err, repositories.ErrTokenNotFound) {
			return nil, ErrTokenNotFound
		}
		if errors.Is(err, repositories.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, repositories.ErrTokenRevoked) {
			return nil, ErrTokenRevoked
		}
		return nil, fmt.Errorf("token validation error: %w", err)
	}

	// Additional security checks
	// 1. Check expiry date explicitly
	if expiryDate.Before(time.Now()) {
		// Also revoke expired token
		_ = s.tokenRepo.RevokeToken(ctx, refreshToken)
		return nil, ErrTokenExpired
	}

	// 2. Check revocation status explicitly
	if isRevoked {
		return nil, ErrTokenRevoked
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

// GetProfile retrieves user profile
func (s *AuthService) GetProfile(ctx context.Context, userID int64) (*dto.UserProfile, error) {
	// Validate user ID
	if err := s.validateUserID(userID); err != nil {
		return nil, err
	}

	// Get user information
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		// Check if the error is user not found from repository
		if errors.Is(err, repositories.ErrUserNotFound) { // Check specific repo error
			return nil, ErrUserNotFound // Return the service's ErrUserNotFound
		}
		// For other errors, wrap and return
		return nil, fmt.Errorf("failed to get user information: %w", err)
	}

	// Basic profile information
	profile := &dto.UserProfile{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		RoleType:  string(user.RoleType),
	}

	var departmentID int64
	// Get additional information based on user type
	if user.RoleType == models.RoleStudent {
		student, err := s.userRepo.GetStudentByUserID(ctx, user.ID)
		if err != nil {
			// Check specific repo error
			if errors.Is(err, repositories.ErrUserNotFound) {
				return nil, ErrUserNotFound // Return service's UserNotFound
			}
			return nil, fmt.Errorf("failed to get student information: %w", err)
		}

		profile.StudentID = &student.StudentID
		profile.GraduationYear = student.GraduationYear
		departmentID = student.DepartmentID // Store department ID

	} else if user.RoleType == models.RoleInstructor {
		instructor, err := s.userRepo.GetInstructorByUserID(ctx, user.ID)
		if err != nil {
			// Check specific repo error
			if errors.Is(err, repositories.ErrUserNotFound) {
				return nil, ErrUserNotFound // Return service's UserNotFound
			}
			return nil, fmt.Errorf("failed to get instructor information: %w", err)
		}

		profile.Title = &instructor.Title
		departmentID = instructor.DepartmentID // Store department ID
	}

	// Fetch Department and Faculty info if DepartmentID is available
	if departmentID > 0 {
		profile.DepartmentID = departmentID
		department, err := s.departmentRepo.GetByID(ctx, departmentID)
		if err != nil {
			s.logger.Warn().Err(err).Int64("deptId", departmentID).Msg("Could not find department info for profile")
		} else if department != nil {
			profile.DepartmentName = department.Name
			profile.FacultyID = department.FacultyID // Get FacultyID from department

			// Fetch Faculty Name using FacultyID
			faculty, err := s.facultyRepo.GetFacultyByID(ctx, department.FacultyID) // Use facultyRepo
			if err != nil {
				s.logger.Warn().Err(err).Int64("facultyId", department.FacultyID).Msg("Could not find faculty name for profile")
			} else if faculty != nil {
				profile.FacultyName = faculty.Name
			}
		}
	}

	return profile, nil
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
