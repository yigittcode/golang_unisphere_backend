package models

// PastExam represents a past exam in the database
type PastExam struct {
	ID           int64  `db:"id"`
	CourseCode   string `db:"course_code"`
	Year         int    `db:"year"`
	Term         string `db:"term"`
	FileID       string `db:"file_id"`
	DepartmentID int64  `db:"department_id"`
	InstructorID int64  `db:"instructor_id"`
}
