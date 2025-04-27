package models

import "time"

// Term definitions moved to models.go
// type Term string
// const (
// 	TermFall   Term = "FALL"
// 	TermSpring Term = "SPRING"
// )

// PastExam represents a past exam record in the database
type PastExam struct {
	ID           int64  `json:"id" db:"id"`
	Year         int    `json:"year" db:"year"`
	Term         Term   `json:"term" db:"term"` // Uses Term from models.go
	DepartmentID int64  `json:"department_id" db:"department_id"`
	CourseCode   string `json:"course_code" db:"course_code"`
	Title        string `json:"title" db:"title"`
	Content      string `json:"content" db:"content"`
	// FileURL alanı çoklu dosya desteği için kaldırıldı
	// FileURL      *string   `json:"file_url" db:"file_url"` // Changed to pointer for potential NULL
	InstructorID int64     `json:"instructor_id" db:"instructor_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`

	// Non-DB fields (potentially populated by service/repo)
	UploadedByName  string      `json:"uploaded_by_name,omitempty"`
	UploadedByEmail string      `json:"uploaded_by_email,omitempty"`
	Department      *Department `json:"department,omitempty"`
	Faculty         *Faculty    `json:"faculty,omitempty"`
	FacultyID       int64       `json:"faculty_id,omitempty"`

	// Çoklu dosya için yeni alan
	Files []*File `json:"files,omitempty"`
}
