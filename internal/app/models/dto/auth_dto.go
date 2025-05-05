package dto

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse represents JWT token information
type TokenResponse struct {
	AccessToken           string `json:"accessToken"`
	TokenType             string `json:"tokenType" example:"Bearer"`
	ExpiresIn             int64  `json:"expiresIn"`
	RefreshToken          string `json:"refreshToken,omitempty"`
	RefreshTokenExpiresIn int64  `json:"refreshTokenExpiresIn,omitempty"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// RegisterRequest represents a generic user registration request
type RegisterRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	FirstName    string `json:"firstName" binding:"required"`
	LastName     string `json:"lastName" binding:"required"`
	DepartmentID int64  `json:"departmentId" binding:"required,min=1"`
}

// UpdateProfileRequest represents profile update data
type UpdateProfileRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

// UserResponse represents basic user information
type UserResponse struct {
	ID                 int64  `json:"id"`
	Email              string `json:"email"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Role               string `json:"role"`
	DepartmentID       *int64 `json:"departmentId,omitempty"`
	ProfilePhotoFileID *int64 `json:"profilePhotoFileId,omitempty"`
	ProfilePhotoURL    string `json:"profilePhotoUrl,omitempty"`
}

// AuthResponse represents successful authentication response
type AuthResponse struct {
	Token TokenResponse `json:"token"`
	User  interface{}   `json:"user"`
}

// RegisterResponse defines the response model for user registration
type RegisterResponse struct {
	Message string `json:"message" example:"Verification email sent. Please check your inbox to complete registration."`
	UserID  int64  `json:"userId" example:"123"`
}

// VerifyEmailRequest defines the request model for email verification
type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required" example:"abc123"`
}

// VerifyEmailResponse defines the response model for email verification
type VerifyEmailResponse struct {
	Message string `json:"message" example:"Email verified successfully"`
}

// ForgotPasswordRequest represents a password reset request
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents a password reset operation
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message"`
}
