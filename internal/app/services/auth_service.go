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
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.TokenResponse, error)

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

// Register registers a new user
func (s *authServiceImpl) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.TokenResponse, error) {
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
		RoleType:     req.RoleType,
		IsActive:     true,
		DepartmentID: &req.DepartmentID,
	}

	// Create user in DB
	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
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
		ID:                 user.ID,
		Email:              user.Email,
		FirstName:          user.FirstName,
		LastName:           user.LastName,
		Role:               string(user.RoleType),
		DepartmentID:       user.DepartmentID,
		ProfilePhotoFileID: user.ProfilePhotoFileID,
		ProfilePhotoURL:    profilePhotoURL,
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
	// Delegate to the user service for a consistent implementation
	userService := NewUserService(s.userRepo, s.departmentRepo, s.fileRepo, s.fileStorage, s.logger)
	_, err := userService.UpdateProfilePhoto(ctx, userID, file)
	return err
}

// DeleteProfilePhoto deletes a user's profile photo
func (s *authServiceImpl) DeleteProfilePhoto(ctx context.Context, userID int64) error {
	// Delegate to the user service for a consistent implementation
	userService := NewUserService(s.userRepo, s.departmentRepo, s.fileRepo, s.fileStorage, s.logger)
	return userService.DeleteProfilePhoto(ctx, userID)
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