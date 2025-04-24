package models

import "time"

// Term represents a semester term
type Term string

// Term constants
const (
	TermFall   Term = "FALL"
	TermSpring Term = "SPRING"
)

// PastExam represents a past exam record in the database
type PastExam struct {
	ID              int64     `json:"id"`
	Year            int       `json:"year"`
	Term            Term      `json:"term"`
	DepartmentID    int64     `json:"department_id"`
	CourseCode      string    `json:"course_code"`
	Title           string    `json:"title"`
	Content         string    `json:"content"`
	FileURL         string    `json:"file_url"`
	InstructorID    int64     `json:"instructor_id"`
	UploadedByName  string    `json:"uploaded_by_name,omitempty"`
	UploadedByEmail string    `json:"uploaded_by_email,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Relations
	Department *Department `json:"department,omitempty"`
	Faculty    *Faculty    `json:"faculty,omitempty"`
	FacultyID  int64       `json:"faculty_id,omitempty"`
}
