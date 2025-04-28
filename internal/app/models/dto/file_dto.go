package dto

// FileResponse represents basic file information
type FileResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// FileUploadResponse represents file upload result
type FileUploadResponse struct {
	FileID string `json:"fileId"`
}
