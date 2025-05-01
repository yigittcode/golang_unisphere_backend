package models

import "time"

// PastExam represents a past exam in the database
type PastExam struct {
	ID           int64     `db:"id"`
	Year         int       `db:"year"`
	Term         Term      `db:"term"`
	CourseCode   string    `db:"course_code"`
	Title        string    `db:"title"`
	Content      string    `db:"content"`
	DepartmentID int64     `db:"department_id"`
	InstructorID int64     `db:"instructor_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	// İlişkisel alanlar
	Files []*File `json:"files,omitempty"` // İlişkili dosyalar
}
