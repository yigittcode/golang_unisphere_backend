package domain

import "time"

// ResourceType tanımı - dosyanın hangi entity'ye ait olduğunu belirtir
type ResourceType string

const (
	ResourceTypePastExam  ResourceType = "PAST_EXAM"
	ResourceTypeClassNote ResourceType = "CLASS_NOTE"
	ResourceTypeUser      ResourceType = "USER"
)

// File dosyalara ait bilgileri içeren struct
type File struct {
	ID           int64        `db:"id" json:"id"`
	FileName     string       `db:"file_name" json:"fileName"`
	FilePath     string       `db:"file_path" json:"filePath"`
	FileURL      string       `db:"file_url" json:"fileUrl"`
	FileSize     int64        `db:"file_size" json:"fileSize"`
	FileType     string       `db:"file_type" json:"fileType"` // MIME type
	ResourceType ResourceType `db:"resource_type" json:"resourceType"`
	ResourceID   int64        `db:"resource_id" json:"resourceId"`
	UploadedBy   int64        `db:"uploaded_by" json:"uploadedBy"`
	CreatedAt    time.Time    `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updatedAt"`
}

// FileResponse - API yanıtları için kullanılacak File DTO
type FileResponse struct {
	ID        int64     `json:"id"`
	FileName  string    `json:"fileName"`
	FileURL   string    `json:"fileUrl"`
	FileSize  int64     `json:"fileSize"`
	FileType  string    `json:"fileType"` // MIME type
	CreatedAt time.Time `json:"createdAt"`
}
