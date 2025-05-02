package models

// Student defines the student model based on the 'students' table
type Student struct {
	ID             int64  `json:"id" db:"id" example:"1"`                     // Unique identifier for the student record
	UserID         int64  `json:"userId" db:"user_id" example:"5"`            // ID of the associated user account
	Identifier     string `json:"identifier" db:"identifier" example:"12345"` // Student's unique identifier/student number
	GraduationYear int    `json:"graduationYear" db:"graduation_year" example:"2026"` // Expected graduation year

	// Relations (populated when needed)
	User       *User       `json:"user,omitempty"`       // Associated user information
	Department *Department `json:"department,omitempty"` // Associated department
}