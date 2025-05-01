package models

import "time"

// FileType represents the type of file
type FileType string

const (
	FileTypePastExam     FileType = "PAST_EXAM"
	FileTypeClassNote    FileType = "CLASS_NOTE"
	FileTypeProfilePhoto FileType = "PROFILE_PHOTO"
)

// File represents a file in the system
type File struct {
	ID           int64     `json:"id" db:"id"`
	FileName     string    `json:"fileName" db:"file_name"`
	FilePath     string    `json:"filePath" db:"file_path"`
	FileURL      string    `json:"fileUrl" db:"file_url"`
	FileSize     int64     `json:"fileSize" db:"file_size"`
	FileType     string    `json:"fileType" db:"file_type"` // MIME type
	ResourceType FileType  `json:"resourceType" db:"resource_type"`
	ResourceID   int64     `json:"resourceId" db:"resource_id"`
	UploadedBy   int64     `json:"uploadedBy" db:"uploaded_by"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// PastExamFile represents the association between past exams and files
type PastExamFile struct {
	ID         int64     `json:"id" db:"id"`
	PastExamID int64     `json:"pastExamId" db:"past_exam_id"`
	FileID     int64     `json:"fileId" db:"file_id"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
}
