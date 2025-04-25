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

// TokenResponse represents the structure of access and refresh tokens returned upon successful authentication
type TokenResponse struct {
	AccessToken      string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIs..."`  // The JWT access token for accessing protected resources
	RefreshToken     string `json:"refresh_token" example:"your_refresh_token_here"` // The refresh token used to obtain new access tokens
	TokenType        string `json:"token_type" example:"Bearer"`                     // Type of the token (always Bearer)
	ExpiresIn        int64  `json:"expires_in" example:"3600"`                       // Duration in seconds until the access token expires
	RefreshExpiresIn int64  `json:"refresh_expires_in" example:"2592000"`            // Duration in seconds until the refresh token expires (e.g., 30 days)
}

// UserProfile represents user profile information returned by the API
type UserProfile struct {
	ID        int64  `json:"id" example:"1"`                                        // Unique identifier for the user
	Email     string `json:"email" example:"user@school.edu.tr"`                    // User's email address
	FirstName string `json:"firstName" example:"John"`                              // User's first name
	LastName  string `json:"lastName" example:"Doe"`                                // User's last name
	RoleType  string `json:"roleType" example:"STUDENT" enums:"STUDENT,INSTRUCTOR"` // User's role (STUDENT or INSTRUCTOR)
	// Student or instructor specific fields
	StudentID      *string `json:"studentId,omitempty" example:"12345678"`                  // Student's unique 8-digit ID (only for students)
	GraduationYear *int    `json:"graduationYear,omitempty" example:"2025"`                 // Student's expected graduation year (optional, only for students)
	Title          *string `json:"title,omitempty" example:"Professor"`                     // Instructor's academic title (only for instructors)
	DepartmentID   int64   `json:"departmentId" example:"1"`                                // ID of the user's department
	DepartmentName string  `json:"departmentName,omitempty" example:"Computer Engineering"` // Name of the user's department
	FacultyID      int64   `json:"facultyId,omitempty" example:"1"`                         // ID of the user's faculty
	FacultyName    string  `json:"facultyName,omitempty" example:"Engineering Faculty"`     // Name of the user's faculty
}
