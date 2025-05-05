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
	"github.com/yigit/unisphere/internal/pkg/email"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
	"github.com/yigit/unisphere/internal/pkg/validation"
	"golang.org/x/crypto/bcrypt"
)

// AuthService defines the interface for authentication-related operations
type AuthService interface {
	// User registration
	Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error)

	// Email verification
	VerifyEmail(ctx context.Context, token string) error
	ResendVerificationEmail(ctx context.Context, email string) error

	// Authentication
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.TokenResponse, error)
	RefreshToken(ctx context.Context, token string) (*dto.TokenResponse, error)

	// Password reset
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token string, newPassword string) error

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
	userRepo               *repositories.UserRepository
	tokenRepo              *repositories.TokenRepository
	departmentRepo         *repositories.DepartmentRepository
	facultyRepo            *repositories.FacultyRepository
	fileRepo               *repositories.FileRepository
	fileStorage            *filestorage.LocalStorage
	verificationTokenRepo  *repositories.VerificationTokenRepository
	passwordResetTokenRepo *repositories.PasswordResetTokenRepository
	emailService           email.EmailService
	jwtService             *auth.JWTService
	logger                 zerolog.Logger
}

// NewAuthService creates a new AuthService
func NewAuthService(
	userRepo *repositories.UserRepository,
	tokenRepo *repositories.TokenRepository,
	departmentRepo *repositories.DepartmentRepository,
	facultyRepo *repositories.FacultyRepository,
	fileRepo *repositories.FileRepository,
	fileStorage *filestorage.LocalStorage,
	verificationTokenRepo *repositories.VerificationTokenRepository,
	passwordResetTokenRepo *repositories.PasswordResetTokenRepository,
	emailService email.EmailService,
	jwtService *auth.JWTService,
	logger zerolog.Logger,
) AuthService {
	return &authServiceImpl{
		userRepo:               userRepo,
		tokenRepo:              tokenRepo,
		departmentRepo:         departmentRepo,
		facultyRepo:            facultyRepo,
		fileRepo:               fileRepo,
		fileStorage:            fileStorage,
		verificationTokenRepo:  verificationTokenRepo,
		passwordResetTokenRepo: passwordResetTokenRepo,
		emailService:           emailService,
		jwtService:             jwtService,
		logger:                 logger,
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
func (s *authServiceImpl) Register(ctx context.Context, req *dto.RegisterRequest) (*dto.RegisterResponse, error) {
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

	// Determine user role based on email pattern
	roleType := models.RoleInstructor // Default role is instructor

	// Extract username from email (part before @)
	emailParts := strings.Split(req.Email, "@")
	if len(emailParts) > 0 {
		username := emailParts[0]

		// Check if username starts with 's' followed by digits (e.g., s200201027)
		if len(username) > 1 && username[0] == 's' {
			// Check if remaining characters are digits
			remainingChars := username[1:]
			isStudent := true

			for _, char := range remainingChars {
				if !unicode.IsDigit(char) {
					isStudent = false
					break
				}
			}

			if isStudent && len(remainingChars) > 0 {
				roleType = models.RoleStudent
			}
		}
	}

	// Create user with department_id and email_verified=false
	user := &models.User{
		Email:         req.Email,
		Password:      string(hashedPassword),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		RoleType:      roleType,
		IsActive:      false, // Set to inactive until email is verified
		EmailVerified: false, // Email not verified yet
		DepartmentID:  &req.DepartmentID,
	}

	// Create user in DB
	userID, err := s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	user.ID = userID

	// Generate verification token
	verificationToken, err := GenerateTokenForVerification()
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to generate verification token")
		return nil, fmt.Errorf("error generating verification token: %w", err)
	}

	// Store verification token
	expiryTime := time.Now().Add(24 * time.Hour) // 24 hours expiry
	err = s.verificationTokenRepo.CreateToken(ctx, userID, verificationToken, expiryTime)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to store verification token")
		return nil, fmt.Errorf("error storing verification token: %w", err)
	}

	// Send verification email
	err = s.emailService.SendVerificationEmail(user.Email, user.FirstName, verificationToken)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to send verification email")
		return nil, fmt.Errorf("error sending verification email: %w", err)
	}

	// Return successful response
	return &dto.RegisterResponse{
		Message: "Registration successful. Please check your email to verify your account.",
		UserID:  userID,
	}, nil
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

	// Check if email is verified - temporarily bypassed
	// if !user.EmailVerified {
	// 	return nil, apperrors.ErrEmailNotVerified
	// }

	// Automatically mark email as verified if it's not
	if !user.EmailVerified {
		s.logger.Info().Int64("userID", user.ID).Msg("Auto-verifying email for login")
		err = s.userRepo.SetEmailVerified(ctx, user.ID, true)
		if err != nil {
			s.logger.Error().Err(err).Int64("userID", user.ID).Msg("Failed to auto-verify email")
			// Continue anyway, don't block login
		}
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

	return s.userRepo.UpdateProfile(ctx, userID, req.FirstName, req.LastName)
}

// UpdateProfilePhoto updates a user's profile photo
func (s *authServiceImpl) UpdateProfilePhoto(ctx context.Context, userID int64, file *multipart.FileHeader) error {
	// Delegate to the user service for a consistent implementation
	userService := NewUserService(s.userRepo, s.departmentRepo, s.fileRepo, s.fileStorage, s, s.logger)
	_, err := userService.UpdateProfilePhoto(ctx, userID, file)
	return err
}

// DeleteProfilePhoto deletes a user's profile photo
func (s *authServiceImpl) DeleteProfilePhoto(ctx context.Context, userID int64) error {
	// Delegate to the user service for a consistent implementation
	userService := NewUserService(s.userRepo, s.departmentRepo, s.fileRepo, s.fileStorage, s, s.logger)
	return userService.DeleteProfilePhoto(ctx, userID)
}

// VerifyEmail verifies a user's email using the verification token
func (s *authServiceImpl) VerifyEmail(ctx context.Context, token string) error {
	// Validate token
	if strings.TrimSpace(token) == "" {
		return apperrors.ErrInvalidEmailToken
	}

	// Get token info
	userID, expiryDate, err := s.verificationTokenRepo.GetTokenInfo(ctx, token)
	if err != nil {
		s.logger.Error().Err(err).Str("token", token).Msg("Failed to get verification token info")
		return apperrors.ErrInvalidEmailToken
	}

	// Check if token is expired - more lenient now (tokens still work for a bit after expiration)
	tokenGracePeriod := 72 * time.Hour // 3 days grace period
	if expiryDate.Add(tokenGracePeriod).Before(time.Now()) {
		s.logger.Warn().Str("token", token).Time("expiryDate", expiryDate).Msg("Verification token expired")
		// Delete expired token
		_ = s.verificationTokenRepo.DeleteToken(ctx, token)
		return apperrors.ErrInvalidEmailToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to get user for email verification")
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if email is already verified - not returning an error, just log it
	if user.EmailVerified {
		s.logger.Info().Int64("userID", userID).Msg("Email already verified")
		// Delete token as it's no longer needed
		_ = s.verificationTokenRepo.DeleteToken(ctx, token)
		// Don't return error, just let them know it was already verified
		return nil
	}

	// Mark email as verified and activate the account
	err = s.userRepo.SetEmailVerified(ctx, userID, true)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to set email verified")
		return fmt.Errorf("error updating email verification status: %w", err)
	}

	// Activate user account
	user.IsActive = true
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to activate user account")
		return fmt.Errorf("error activating user account: %w", err)
	}

	// Delete verification token after successful verification
	err = s.verificationTokenRepo.DeleteToken(ctx, token)
	if err != nil {
		s.logger.Warn().Err(err).Str("token", token).Msg("Failed to delete verification token")
		// Don't return error, as verification was successful
	}

	// Send welcome email
	err = s.emailService.SendWelcomeEmail(user.Email, user.FirstName)
	if err != nil {
		s.logger.Warn().Err(err).Int64("userID", userID).Msg("Failed to send welcome email")
		// Don't return error, as verification was successful
	}

	return nil
}

// ResendVerificationEmail sends a new verification email to the user
func (s *authServiceImpl) ResendVerificationEmail(ctx context.Context, email string) error {
	// Validate email - more lenient to help users
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("%w: email cannot be empty", apperrors.ErrValidationFailed)
	}

	// Convert email to lowercase to ensure consistency
	email = strings.ToLower(email)

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		s.logger.Error().Err(err).Str("email", email).Msg("Failed to find user for resending verification email")
		return apperrors.ErrUserNotFound
	}

	// If email is already verified, just set user as active and return success
	if user.EmailVerified {
		s.logger.Info().Int64("userID", user.ID).Msg("Email already verified, ensuring user is active")
		// Make sure user is active
		if !user.IsActive {
			user.IsActive = true
			err = s.userRepo.Update(ctx, user)
			if err != nil {
				s.logger.Error().Err(err).Int64("userID", user.ID).Msg("Failed to activate user account")
				// Continue anyway
			}
		}
		// Return success rather than error
		return nil
	}

	// Delete any existing verification tokens for the user
	err = s.verificationTokenRepo.DeleteTokensByUserID(ctx, user.ID)
	if err != nil {
		s.logger.Warn().Err(err).Int64("userID", user.ID).Msg("Failed to delete existing verification tokens")
		// Continue anyway
	}

	// Generate new verification token
	verificationToken, err := GenerateTokenForVerification()
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", user.ID).Msg("Failed to generate verification token")
		return fmt.Errorf("error generating verification token: %w", err)
	}

	// Store verification token with longer expiry
	expiryTime := time.Now().Add(7 * 24 * time.Hour) // 7 days expiry instead of 24 hours
	err = s.verificationTokenRepo.CreateToken(ctx, user.ID, verificationToken, expiryTime)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", user.ID).Msg("Failed to store verification token")
		return fmt.Errorf("error storing verification token: %w", err)
	}

	// Send verification email
	err = s.emailService.SendVerificationEmail(user.Email, user.FirstName, verificationToken)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", user.ID).Msg("Failed to send verification email")
		return fmt.Errorf("error sending verification email: %w", err)
	}

	return nil
}

// Helper functions

// GenerateTokenForVerification generates a random token for email verification
func GenerateTokenForVerification() (string, error) {
	return email.GenerateVerificationToken()
}

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

// GenerateTokenForPasswordReset generates a new random token for password reset
func GenerateTokenForPasswordReset() (string, error) {
	return GenerateTokenForVerification() // Use the same token generation
}

// ForgotPassword initiates the password reset process
func (s *authServiceImpl) ForgotPassword(ctx context.Context, email string) error {
	// Validate email
	if err := s.validateEmail(email); err != nil {
		return err
	}

	// Convert email to lowercase
	email = strings.ToLower(email)

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			// Don't reveal that the user doesn't exist for security reasons
			s.logger.Info().Str("email", email).Msg("Password reset requested for non-existent user")
			return nil
		}
		return fmt.Errorf("error retrieving user: %w", err)
	}

	// Delete any existing password reset tokens for user
	err = s.passwordResetTokenRepo.DeleteTokensByUserID(ctx, user.ID)
	if err != nil {
		s.logger.Warn().Err(err).Int64("userID", user.ID).Msg("Failed to delete existing password reset tokens")
		// Continue anyway
	}

	// Generate password reset token
	resetToken, err := GenerateTokenForPasswordReset()
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", user.ID).Msg("Failed to generate password reset token")
		return fmt.Errorf("error generating reset token: %w", err)
	}

	// Store password reset token (valid for 24 hours)
	expiryTime := time.Now().Add(24 * time.Hour)
	err = s.passwordResetTokenRepo.CreateToken(ctx, user.ID, resetToken, expiryTime)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", user.ID).Msg("Failed to store password reset token")
		return fmt.Errorf("error storing reset token: %w", err)
	}

	// Send password reset email
	err = s.emailService.SendPasswordResetEmail(user.Email, user.FirstName, resetToken)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", user.ID).Msg("Failed to send password reset email")
		return fmt.Errorf("error sending password reset email: %w", err)
	}

	return nil
}

// ResetPassword resets a user's password using a valid token
func (s *authServiceImpl) ResetPassword(ctx context.Context, token string, newPassword string) error {
	// Validate token
	if strings.TrimSpace(token) == "" {
		return apperrors.ErrInvalidPasswordResetToken
	}

	// Validate new password
	if err := s.ValidatePassword(newPassword); err != nil {
		return err
	}

	// Get token info
	userID, expiryDate, used, err := s.passwordResetTokenRepo.GetTokenInfo(ctx, token)
	if err != nil {
		s.logger.Error().Err(err).Str("token", token).Msg("Failed to get password reset token info")
		return apperrors.ErrInvalidPasswordResetToken
	}

	// Check if token has already been used
	if used {
		s.logger.Warn().Str("token", token).Msg("Password reset token already used")
		return apperrors.ErrPasswordResetTokenUsed
	}

	// Check if token is expired
	if expiryDate.Before(time.Now()) {
		s.logger.Warn().Str("token", token).Time("expiryDate", expiryDate).Msg("Password reset token expired")
		// Delete expired token
		_ = s.passwordResetTokenRepo.DeleteToken(ctx, token)
		return apperrors.ErrInvalidPasswordResetToken
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to get user for password reset")
		return fmt.Errorf("user not found: %w", err)
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error hashing password: %w", err)
	}

	// Update user's password - Execute a direct query rather than using Update
	query := `
		UPDATE users 
		SET password = $1, updated_at = NOW() 
		WHERE id = $2
	`

	_, err = s.userRepo.GetDB().Exec(ctx, query, string(hashedPassword), userID)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to update user password with direct query")
		// Fall back to using the user object if direct query fails

		// Update user's password
		user.Password = string(hashedPassword)
		err = s.userRepo.Update(ctx, user)
		if err != nil {
			s.logger.Error().Err(err).Int64("userID", userID).Msg("Failed to update user password with both methods")
			return fmt.Errorf("error updating user password: %w", err)
		}
	} else {
		s.logger.Info().Int64("userID", userID).Msg("User password updated successfully with direct query")
	}

	// Mark token as used
	err = s.passwordResetTokenRepo.MarkTokenAsUsed(ctx, token)
	if err != nil {
		s.logger.Warn().Err(err).Str("token", token).Msg("Failed to mark password reset token as used")
		// Don't return error since password was updated successfully
	}

	// Ensure that the user account is active and email is verified
	if !user.IsActive || !user.EmailVerified {
		// Do a direct update for activation too
		activateQuery := `
			UPDATE users 
			SET is_active = true, email_verified = true, updated_at = NOW() 
			WHERE id = $1
		`

		_, err = s.userRepo.GetDB().Exec(ctx, activateQuery, userID)
		if err != nil {
			s.logger.Warn().Err(err).Int64("userID", userID).Msg("Failed to activate user with direct query")

			// Fall back to the original method
			user.IsActive = true
			user.EmailVerified = true
			err = s.userRepo.Update(ctx, user)
			if err != nil {
				s.logger.Warn().Err(err).Int64("userID", userID).Msg("Failed to activate user account after password reset")
				// Don't return error since password was updated successfully
			}
		} else {
			s.logger.Info().Int64("userID", userID).Msg("User activated successfully with direct query")
		}
	}

	// Send password changed notification email
	err = s.emailService.SendPasswordChangedEmail(user.Email, user.FirstName)
	if err != nil {
		s.logger.Warn().Err(err).Int64("userID", userID).Msg("Failed to send password changed notification")
		// Don't return error since password was updated successfully
	}

	return nil
}
