package dto

// RegisterStudentRequest represents student registration data
type RegisterStudentRequest struct {
	Email          string `json:"email" binding:"required,email"`
	Password       string `json:"password" binding:"required,min=8"`
	FirstName      string `json:"firstName" binding:"required"`
	LastName       string `json:"lastName" binding:"required"`
	DepartmentID   int64  `json:"departmentId" binding:"required,gt=0"`
	StudentID      string `json:"studentId" binding:"required,len=8,numeric"`
	GraduationYear *int   `json:"graduationYear,omitempty" binding:"omitempty,min=1900"`
}

// RegisterInstructorRequest represents instructor registration data
type RegisterInstructorRequest struct {
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=8"`
	FirstName    string `json:"firstName" binding:"required"`
	LastName     string `json:"lastName" binding:"required"`
	DepartmentID int64  `json:"departmentId" binding:"required,gt=0"`
	Title        string `json:"title" binding:"required"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse represents JWT token information
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	TokenType    string `json:"tokenType" example:"Bearer"`
	ExpiresIn    int64  `json:"expiresIn"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

// UpdateUserProfileRequest defines the parameters for updating user profile
type UpdateUserProfileRequest struct {
	FirstName string `json:"firstName" binding:"required" example:"John"`
	LastName  string `json:"lastName" binding:"required" example:"Doe"`
	Email     string `json:"email" binding:"required,email" example:"john.doe@school.edu.tr"`
}

// BaseUserProfile represents the base user profile information
type BaseUserProfile struct {
	ID        int64  `json:"id" example:"1"`
	Email     string `json:"email" example:"user@example.com"`
	FirstName string `json:"firstName" example:"John"`
	LastName  string `json:"lastName" example:"Doe"`
	Role      string `json:"role" example:"student"`
}

// UserResponse represents basic user information
type UserResponse struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	Role         string `json:"role"`
	DepartmentID *int64 `json:"departmentId,omitempty"`
}

// StudentResponse extends UserResponse with student-specific fields
type StudentResponse struct {
	UserResponse
	StudentID      string `json:"studentId"`
	GraduationYear *int   `json:"graduationYear,omitempty"`
}

// TeacherResponse extends UserResponse with teacher-specific fields
type TeacherResponse struct {
	UserResponse
	Title string `json:"title"`
}

// UpdateProfileRequest represents profile update data
type UpdateProfileRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
}

// AuthResponse represents successful authentication response
type AuthResponse struct {
	Token TokenResponse `json:"token"`
	User  UserResponse  `json:"user"`
}
