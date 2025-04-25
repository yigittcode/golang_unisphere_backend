package models

import (
	"time"
)

// User defines the user model based on the 'users' table
type User struct {
	ID          int64      `json:"id" db:"id"`
	Email       string     `json:"email" db:"email"`
	Password    string     `json:"-" db:"password"` // Exclude from JSON, but needed for DB mapping
	FirstName   string     `json:"firstName" db:"first_name"`
	LastName    string     `json:"lastName" db:"last_name"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time  `json:"updatedAt" db:"updated_at"`
	RoleType    RoleType   `json:"roleType" db:"role_type"` // Uses RoleType from models.go
	IsActive    bool       `json:"isActive" db:"is_active"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty" db:"last_login_at"` // Added LastLoginAt (nullable)
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
