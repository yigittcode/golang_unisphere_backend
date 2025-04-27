package filestorage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// LocalStorage handles saving files to the local filesystem.
type LocalStorage struct {
	basePath string // The root directory where files will be stored
	baseURL  string // The base URL to access the stored files (optional, for generating full URLs)
}

// NewLocalStorage creates a new LocalStorage instance.
// basePath is the required directory path on the server.
// baseURL is optional; if provided, it will be prepended to returned file paths.
func NewLocalStorage(basePath, baseURL string) (*LocalStorage, error) {
	// Ensure the base path exists
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		logger.Error().Err(err).Str("path", basePath).Msg("Failed to create storage directory")
		return nil, fmt.Errorf("failed to create storage directory %s: %w", basePath, err)
	}
	logger.Info().Str("path", basePath).Msg("Local storage directory ensured")

	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}, nil
}

// SaveFileWithPath saves a file to a specified subdirectory
func (ls *LocalStorage) SaveFileWithPath(fileHeader *multipart.FileHeader, subPath string) (string, error) {
	if fileHeader == nil {
		return "", nil // No file uploaded
	}

	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		logger.Error().Err(err).Str("filename", fileHeader.Filename).Msg("Failed to open uploaded file")
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Ensure the subdirectory exists
	fullDirPath := ls.basePath
	if subPath != "" {
		fullDirPath = filepath.Join(ls.basePath, subPath)
		if err := os.MkdirAll(fullDirPath, os.ModePerm); err != nil {
			logger.Error().Err(err).Str("path", fullDirPath).Msg("Failed to create subdirectory")
			return "", fmt.Errorf("failed to create subdirectory: %w", err)
		}
	}

	// Generate a unique filename to prevent collisions
	ext := filepath.Ext(fileHeader.Filename)
	uniqueFilename := uuid.New().String() + ext

	// Construct the full destination path
	dstPath := filepath.Join(fullDirPath, uniqueFilename)

	// Create the destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		logger.Error().Err(err).Str("path", dstPath).Msg("Failed to create destination file")
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy the uploaded file content to the destination file
	if _, err = io.Copy(dst, file); err != nil {
		logger.Error().Err(err).Str("path", dstPath).Msg("Failed to copy uploaded file content")
		// Attempt to remove the partially created file
		_ = os.Remove(dstPath)
		return "", fmt.Errorf("failed to save file content: %w", err)
	}

	// Construct the accessible path/URL
	var accessiblePath string

	if ls.baseURL != "" {
		// If baseURL is provided, use it to construct a URL
		// Make sure we don't have double slashes
		if subPath != "" {
			accessiblePath = strings.TrimRight(ls.baseURL, "/") + "/" + subPath + "/" + uniqueFilename
		} else {
			accessiblePath = strings.TrimRight(ls.baseURL, "/") + "/" + uniqueFilename
		}
	} else {
		// If no baseURL, use the relative path to uploads directory
		if subPath != "" {
			accessiblePath = filepath.Join("uploads", subPath, uniqueFilename)
		} else {
			accessiblePath = filepath.Join("uploads", uniqueFilename)
		}
	}

	logger.Info().Str("filename", fileHeader.Filename).Str("saved_as", uniqueFilename).Str("accessible_path", accessiblePath).Msg("File saved successfully")
	return accessiblePath, nil
}

// SaveFile saves an uploaded file using the default path
func (ls *LocalStorage) SaveFile(fileHeader *multipart.FileHeader) (string, error) {
	return ls.SaveFileWithPath(fileHeader, "")
}

// DeleteFile removes a file from the storage filesystem.
// It accepts the file path as stored in the database (e.g., uploads/filename.jpg).
// Returns nil if deletion is successful or if the file doesn't exist.
func (ls *LocalStorage) DeleteFile(filePath string) error {
	if filePath == "" {
		return nil // Nothing to delete
	}

	// Extract the filename from the path
	// The stored path is typically in the format: "uploads/filename.ext"
	filename := filepath.Base(filePath)

	// Ensure we're only getting the filename portion
	if filename == "" || filename == "." || filename == "/" || filename == "uploads" {
		return fmt.Errorf("invalid file path: %s", filePath)
	}

	// Construct the full physical path to the file
	physicalPath := filepath.Join(ls.basePath, filename)

	// Check if the file exists first
	if _, err := os.Stat(physicalPath); os.IsNotExist(err) {
		logger.Warn().Str("path", physicalPath).Msg("File to delete does not exist")
		return nil // Consider this a successful delete (idempotent operation)
	}

	// Remove the file
	if err := os.Remove(physicalPath); err != nil {
		logger.Error().Err(err).Str("path", physicalPath).Msg("Failed to delete file")
		return fmt.Errorf("failed to delete file: %w", err)
	}

	logger.Info().Str("path", physicalPath).Msg("File deleted successfully")
	return nil
}

// GetFullPath returns the full filesystem path for a given file URL.
// This is useful for getting the actual path for deletion.
func (ls *LocalStorage) GetFullPath(fileURL string) string {
	// Extract filename from URL or path
	filename := filepath.Base(fileURL)
	if filename == "" || filename == "." || filename == "/" {
		return ""
	}

	return filepath.Join(ls.basePath, filename)
}

// TODO: Add DeleteFile method if needed
// func (ls *LocalStorage) DeleteFile(filePath string) error { ... }
