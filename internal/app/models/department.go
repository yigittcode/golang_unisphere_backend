package models

// Department represents a department in a faculty
type Department struct {
	ID        int64    `json:"id"`
	FacultyID int64    `json:"faculty_id"`
	Name      string   `json:"name"`
	Code      string   `json:"code"`
	Faculty   *Faculty `json:"faculty,omitempty"`
}
