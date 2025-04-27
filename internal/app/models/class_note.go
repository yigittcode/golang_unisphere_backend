package models

import "time"

// ClassNoteTerm definition removed (now in models.go as Term)

// ClassNote represents the structure for a class note in the database.
type ClassNote struct {
	ID           int64  `db:"id" json:"id"`
	Year         int    `db:"year" json:"year"`
	Term         Term   `db:"term" json:"term"` // Uses Term from models.go
	DepartmentID int64  `db:"department_id" json:"departmentId"`
	CourseCode   string `db:"course_code" json:"courseCode"`
	Title        string `db:"title" json:"title"`
	Content      string `db:"content" json:"content"`
	// Image alanı çoklu dosya desteği için kaldırıldı
	// Image        *string   `db:"image" json:"image"` // Pointer to handle NULL
	UserID    int64     `db:"user_id" json:"userId"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`

	// Çoklu dosya için yeni alan
	Files []*File `json:"files,omitempty"`
}
