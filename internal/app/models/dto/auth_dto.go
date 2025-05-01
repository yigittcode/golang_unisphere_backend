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

// UpdateProfileRequest represents profile update data
type UpdateProfileRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

// UserResponse represents basic user information
type UserResponse struct {
	ID              int64  `json:"id"`
	Email           string `json:"email"`
	FirstName       string `json:"firstName"`
	LastName        string `json:"lastName"`
	Role            string `json:"role"`
	DepartmentID    *int64 `json:"departmentId,omitempty"`
	ProfilePhotoURL string `json:"profilePhotoUrl,omitempty"`
}

// AuthResponse represents successful authentication response
type AuthResponse struct {
	Token TokenResponse `json:"token"`
	User  interface{}   `json:"user"`
}
