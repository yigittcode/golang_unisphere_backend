package filestorage

import (
	"context"
	"mime/multipart"
)

// FileInfo represents information about a stored file
type FileInfo struct {
	ID       int64  // Database ID of the file record (if applicable)
	Path     string // Relative or full path where the file is stored
	Filename string // Original filename
	FileSize int64  // Size in bytes
	MimeType string // MIME type of the file
}

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	// SaveFile saves a file and returns information about where it was stored
	SaveFile(fileHeader *multipart.FileHeader) (string, error)

	// SaveFileWithPath lets you specify a subdirectory for storing the file
	SaveFileWithPath(fileHeader *multipart.FileHeader, path string) (string, error)

	// DeleteFile removes a file from storage
	DeleteFile(filePath string) error

	// GetFullPath returns the full filesystem path for a given file URL
	GetFullPath(fileURL string) string
}

// FileStorageWithDB extends FileStorage with database operations
type FileStorageWithDB interface {
	FileStorage

	// SaveFileWithDB saves a file and creates a record in the database
	SaveFileWithDB(ctx context.Context, fileHeader *multipart.FileHeader, path string) (*FileInfo, error)

	// CreateFileRecord creates a file record in the database
	CreateFileRecord(ctx context.Context, filePath string, fileHeader *multipart.FileHeader) (int64, error)

	// DeleteFileRecord deletes a file record from the database and the corresponding file from storage
	DeleteFileRecord(ctx context.Context, fileID int64) error
}
