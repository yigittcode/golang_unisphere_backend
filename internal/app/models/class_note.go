package models

// ClassNoteTerm definition removed (now in models.go as Term)

// ClassNote represents a class note in the database
type ClassNote struct {
	ID           int64  `db:"id"`
	CourseCode   string `db:"course_code"`
	Title        string `db:"title"`
	Description  string `db:"description"`
	FileID       string `db:"file_id"`
	DepartmentID int64  `db:"department_id"`
	InstructorID int64  `db:"instructor_id"`
}
