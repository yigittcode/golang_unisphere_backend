package models

// Instructor defines the instructor model based on the 'instructors' table
type Instructor struct {
	ID    int64  `json:"id" db:"id" example:"1"`           // Unique identifier for the instructor record
	UserID int64  `json:"userId" db:"user_id" example:"5"` // ID of the associated user account
	Title  string `json:"title" db:"title" example:"Associate Professor"` // Academic title of the instructor

	// Relations (populated when needed)
	User       *User       `json:"user,omitempty"`       // Associated user information
	Department *Department `json:"department,omitempty"` // Associated department
}