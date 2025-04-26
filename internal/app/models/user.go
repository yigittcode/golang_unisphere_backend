package models

import (
	"time"
)

// User defines the user model based on the 'users' table
type User struct {
	ID              int64      `json:"id" db:"id" example:"1"`                                                         // Unique identifier for the user
	Email           string     `json:"email" db:"email" example:"user@school.edu.tr"`                                  // User's email address
	Password        string     `json:"-" db:"password"`                                                                // User's hashed password (excluded from JSON)
	FirstName       string     `json:"firstName" db:"first_name" example:"John"`                                       // User's first name
	LastName        string     `json:"lastName" db:"last_name" example:"Doe"`                                          // User's last name
	CreatedAt       time.Time  `json:"createdAt" db:"created_at" example:"2024-01-01T10:00:00Z"`                       // Timestamp when the user was created
	UpdatedAt       time.Time  `json:"updatedAt" db:"updated_at" example:"2024-01-02T15:30:00Z"`                       // Timestamp when the user was last updated
	RoleType        RoleType   `json:"roleType" db:"role_type" example:"STUDENT"`                                      // User's role (STUDENT or INSTRUCTOR)
	IsActive        bool       `json:"isActive" db:"is_active" example:"true"`                                         // Whether the user account is active
	LastLoginAt     *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at" example:"2024-04-20T18:00:00Z"`        // Timestamp of the last login (nullable)
	ProfilePhotoURL *string    `json:"profilePhotoUrl,omitempty" db:"profile_photo_url" example:"uploads/profile.jpg"` // URL of the user's profile photo (nullable)
	// last_login_at from DB is missing, add if needed
}

// Student defines the student model based on the 'students' table
type Student struct {
	ID             int64       `json:"id" db:"id"`
	UserID         int64       `json:"userId" db:"user_id"`
	StudentID      string      `json:"studentId" db:"student_id"`
	DepartmentID   int64       `json:"departmentId" db:"department_id"`
	GraduationYear *int        `json:"graduationYear,omitempty" db:"graduation_year"` // Pointer for potential NULL
	User           *User       `json:"user,omitempty"`                                // Relation, no db tag
	Department     *Department `json:"department,omitempty"`                          // Relation, no db tag
}

// Instructor defines the instructor model based on the 'instructors' table
type Instructor struct {
	ID           int64       `json:"id" db:"id"`
	UserID       int64       `json:"userId" db:"user_id"`
	DepartmentID int64       `json:"departmentId" db:"department_id"`
	Title        string      `json:"title" db:"title"`
	User         *User       `json:"user,omitempty"`       // Relation, no db tag
	Department   *Department `json:"department,omitempty"` // Relation, no db tag
}
