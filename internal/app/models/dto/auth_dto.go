package dto

// RegisterStudentRequest represents student registration request
type RegisterStudentRequest struct {
	Email          string `json:"email" binding:"required,email" example:"student@school.edu.tr"`       // Student's email address (must be unique)
	Password       string `json:"password" binding:"required,min=8" example:"Password123"`              // Student's password (min 8 characters, letter+number required by service)
	FirstName      string `json:"firstName" binding:"required" example:"John"`                          // Student's first name
	LastName       string `json:"lastName" binding:"required" example:"Doe"`                            // Student's last name
	DepartmentID   int64  `json:"departmentId" binding:"required,gt=0" example:"1"`                     // ID of the department the student belongs to
	StudentID      string `json:"studentId" binding:"required,len=8,numeric" example:"12345678"`        // Student's unique 8-digit ID
	GraduationYear *int   `json:"graduationYear,omitempty" binding:"omitempty,min=1900" example:"2025"` // Student's expected graduation year (optional)
}

// RegisterInstructorRequest represents instructor registration request
type RegisterInstructorRequest struct {
	Email        string `json:"email" binding:"required,email" example:"instructor@school.edu.tr"` // Instructor's email address (must be unique)
	Password     string `json:"password" binding:"required,min=8" example:"Password123"`           // Instructor's password (min 8 characters, letter+number required by service)
	FirstName    string `json:"firstName" binding:"required" example:"Jane"`                       // Instructor's first name
	LastName     string `json:"lastName" binding:"required" example:"Smith"`                       // Instructor's last name
	DepartmentID int64  `json:"departmentId" binding:"required,gt=0" example:"1"`                  // ID of the department the instructor belongs to
	Title        string `json:"title" binding:"required" example:"Professor"`                      // Instructor's academic title (e.g., Professor, Dr.)
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@school.edu.tr"` // User's registered email address
	Password string `json:"password" binding:"required" example:"Password123"`           // User's password
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required" example:"your_refresh_token_here"` // The refresh token obtained during login
}

// UpdateUserProfileRequest defines the parameters for updating user profile
type UpdateUserProfileRequest struct {
	FirstName string `json:"firstName" binding:"required" example:"John"`
	LastName  string `json:"lastName" binding:"required" example:"Doe"`
	Email     string `json:"email" binding:"required,email" example:"john.doe@school.edu.tr"`
}

// BaseUserProfile contains common user information
type BaseUserProfile struct {
	ID        int64        `json:"id"`
	Email     string       `json:"email"`
	FirstName string       `json:"firstName"`
	LastName  string       `json:"lastName"`
	RoleType  string       `json:"roleType"`
	Faculty   *FacultyInfo `json:"faculty,omitempty"`
	Photo     *PhotoInfo   `json:"photo,omitempty"`
}

// StudentProfile represents a student's profile
type StudentProfile struct {
	BaseUserProfile
	Identifier     string `json:"identifier"`
	GraduationYear *int   `json:"graduationYear,omitempty"`
}

// InstructorProfile represents an instructor's profile
type InstructorProfile struct {
	BaseUserProfile
	Title string `json:"title"`
}

// DepartmentInfo represents department information
type DepartmentInfo struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// FacultyInfo represents faculty information
type FacultyInfo struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	Department *DepartmentInfo `json:"department,omitempty"`
}

// PhotoInfo represents profile photo information
type PhotoInfo struct {
	ID       int64  `json:"id"`
	URL      string `json:"url"`
	FileType string `json:"fileType"`
}

// TokenResponse represents the response for token-based operations
type TokenResponse struct {
	AccessToken      string `json:"accessToken"`
	RefreshToken     string `json:"refreshToken"`
	TokenType        string `json:"tokenType"`
	ExpiresIn        int64  `json:"expiresIn"`
	RefreshExpiresIn int64  `json:"refreshExpiresIn"`
}

// AuthResponse represents the authentication response including tokens and user info
type AuthResponse struct {
	Tokens *TokenResponse   `json:"tokens"`
	User   *BaseUserProfile `json:"user"`
}
