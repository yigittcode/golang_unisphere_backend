package models

import (
	"time"
)

// User defines the user model based on the 'users' table
type User struct {
	ID                 int64      `json:"id" db:"id" example:"1"`                                                  // Unique identifier for the user
	Email              string     `json:"email" db:"email" example:"user@school.edu.tr"`                           // User's email address
	Password           string     `json:"-" db:"password"`                                                         // User's hashed password (excluded from JSON)
	FirstName          string     `json:"firstName" db:"first_name" example:"John"`                                // User's first name
	LastName           string     `json:"lastName" db:"last_name" example:"Doe"`                                   // User's last name
	CreatedAt          time.Time  `json:"createdAt" db:"created_at" example:"2024-01-01T10:00:00Z"`                // Timestamp when the user was created
	UpdatedAt          time.Time  `json:"updatedAt" db:"updated_at" example:"2024-01-02T15:30:00Z"`                // Timestamp when the user was last updated
	RoleType           RoleType   `json:"roleType" db:"role_type" example:"STUDENT"`                               // User's role (STUDENT or INSTRUCTOR)
	IsActive           bool       `json:"isActive" db:"is_active" example:"true"`                                  // Whether the user account is active
	EmailVerified      bool       `json:"emailVerified" db:"email_verified" example:"false"`                       // Whether the email has been verified
	LastLoginAt        *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at" example:"2024-04-20T18:00:00Z"` // Timestamp of the last login (nullable)
	DepartmentID       *int64     `json:"departmentId,omitempty" db:"department_id" example:"1"`                   // User's department (nullable)
	ProfilePhotoFileID *int64     `json:"profilePhotoFileId,omitempty" db:"profile_photo_file_id"`                 // Profile photo file ID (nullable)

	// Relations (populated when needed)
	Department *Department `json:"department,omitempty"` // Relation, no db tag
}
