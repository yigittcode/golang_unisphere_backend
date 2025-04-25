package models

// Faculty represents a faculty at the university
type Faculty struct {
	ID          int64   `json:"id" db:"id" example:"1"`                                                                       // Unique identifier for the faculty
	Name        string  `json:"name" db:"name" binding:"required" example:"Engineering Faculty"`                              // Name of the faculty (required)
	Code        string  `json:"code" db:"code" binding:"required" example:"ENG"`                                              // Unique code for the faculty (e.g., ENG, SCI)
	Description *string `json:"description,omitempty" db:"description" example:"Faculty of Engineering and Natural Sciences"` // Optional description of the faculty
	// Description field from DB is missing in model, add if needed
}
