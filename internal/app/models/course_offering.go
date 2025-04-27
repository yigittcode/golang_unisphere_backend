package models

// CourseOffering represents a specific offering of a course by an instructor in a given year and term.
type CourseOffering struct {
	ID           int64 `json:"id" db:"id"`
	CourseID     int64 `json:"courseId" db:"course_id"`
	InstructorID int64 `json:"instructorId" db:"instructor_id"`
	Year         int   `json:"year" db:"year"`
	Term         Term  `json:"term" db:"term"` // Uses Term from models.go

	// Relations (populated when needed)
	Course     *Course     `json:"course,omitempty"`
	Instructor *Instructor `json:"instructor,omitempty"`
}
