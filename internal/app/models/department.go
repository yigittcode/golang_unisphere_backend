package models

// Department represents a department within a specific faculty
type Department struct {
	ID        int64    `json:"id" db:"id" example:"1"`                                           // Unique identifier for the department
	FacultyID int64    `json:"faculty_id" db:"faculty_id" binding:"required,gt=0" example:"1"`   // ID of the faculty this department belongs to (required)
	Name      string   `json:"name" db:"name" binding:"required" example:"Computer Engineering"` // Name of the department (required)
	Code      string   `json:"code" db:"code" binding:"required" example:"CENG"`                 // Unique code for the department (e.g., CENG, EEE, MATH)
	Faculty   *Faculty `json:"faculty,omitempty"`                                                // Associated faculty details (populated in some responses)
}
