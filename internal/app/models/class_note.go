package models

import "time"

// ClassNoteTerm definition removed (now in models.go as Term)

// ClassNote represents a class note in the database
type ClassNote struct {
	ID           int64     `db:"id"`
	CourseCode   string    `db:"course_code"`
	Title        string    `db:"title"`
	Description  string    `db:"description"`
	Content      string    `db:"content"`
	DepartmentID int64     `db:"department_id"`
	UserID       int64     `db:"user_id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	// İlişkisel alanlar
	Files []*File `json:"files,omitempty"` // İlişkili dosyalar
}

// ClassNoteFile represents the relationship between class notes and files
type ClassNoteFile struct {
	ID          int64     `db:"id"`
	ClassNoteID int64     `db:"class_note_id"`
	FileID      int64     `db:"file_id"`
	CreatedAt   time.Time `db:"created_at"`
}
