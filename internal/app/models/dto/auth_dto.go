package dto

import "time"

// RegisterStudentRequest represents student registration request
type RegisterStudentRequest struct {
	Email          string `json:"email" binding:"required,email" example:"student@school.edu.tr"`
	Password       string `json:"password" binding:"required,min=6" example:"Password123"`
	FirstName      string `json:"firstName" binding:"required" example:"John"`
	LastName       string `json:"lastName" binding:"required" example:"Doe"`
	DepartmentID   int64  `json:"departmentId" binding:"required" example:"1"`
	StudentID      string `json:"studentId" binding:"required" example:"12345678"`
	GraduationYear *int   `json:"graduationYear" binding:"omitempty" example:"2025"`
}

// RegisterInstructorRequest represents instructor registration request
type RegisterInstructorRequest struct {
	Email        string `json:"email" binding:"required,email" example:"instructor@school.edu.tr"`
	Password     string `json:"password" binding:"required,min=6" example:"Password123"`
	FirstName    string `json:"firstName" binding:"required" example:"Jane"`
	LastName     string `json:"lastName" binding:"required" example:"Smith"`
	DepartmentID int64  `json:"departmentId" binding:"required" example:"1"`
	Title        string `json:"title" binding:"required" example:"Professor"`
}

// LoginRequest represents login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@school.edu.tr"`
	Password string `json:"password" binding:"required" example:"Password123"`
}

// RefreshTokenRequest represents token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required" example:"62a71580-9d9e-4884-a000-5dc497a3d1d8"`
}

// TokenResponse represents token response
type TokenResponse struct {
	AccessToken      string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken     string `json:"refresh_token" example:"62a71580-9d9e-4884-a000-5dc497a3d1d8"`
	TokenType        string `json:"token_type" example:"Bearer"`
	ExpiresIn        int    `json:"expires_in" example:"86400"`          // seconds
	RefreshExpiresIn int    `json:"refresh_expires_in" example:"604800"` // seconds
}

// APIResponse represents standard API response
type APIResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message" example:"Operation successful"`
	Data      interface{} `json:"data,omitempty"`
	Error     interface{} `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp" example:"2025-04-23T12:01:05.123Z"`
}

// UserProfile represents user profile
type UserProfile struct {
	ID        int64  `json:"id" example:"1"`
	Email     string `json:"email" example:"user@school.edu.tr"`
	FirstName string `json:"firstName" example:"John"`
	LastName  string `json:"lastName" example:"Doe"`
	RoleType  string `json:"roleType" example:"STUDENT"`
	// Student or instructor specific fields
	StudentID      string `json:"studentId,omitempty" example:"1234567"`
	GraduationYear *int   `json:"graduationYear,omitempty" example:"2025"`
	Title          string `json:"title,omitempty" example:"Professor"`
	DepartmentID   int64  `json:"departmentId" example:"1"`
	DepartmentName string `json:"departmentName,omitempty" example:"Computer Science"`
}
