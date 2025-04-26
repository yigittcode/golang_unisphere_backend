package filestorage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

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

// SaveFile saves an uploaded file to the configured local storage path.
// It generates a unique filename and returns the accessible path/URL for the file.
// Returns an empty string and nil error if fileHeader is nil.
func (ls *LocalStorage) SaveFile(fileHeader *multipart.FileHeader) (string, error) {
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

	// Generate a unique filename to prevent collisions
	ext := filepath.Ext(fileHeader.Filename)
	uniqueFilename := uuid.New().String() + ext

	// Construct the full destination path
	dstPath := filepath.Join(ls.basePath, uniqueFilename)

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
	// If baseURL is configured, return a full URL, otherwise return relative path
	accessiblePath := uniqueFilename // Start with just the filename
	if ls.baseURL != "" {
		// Ensure no double slashes
		base := ls.baseURL
		if base[len(base)-1:] == "/" {
			base = base[:len(base)-1]
		}
		accessiblePath = base + "/" + uniqueFilename
	} else {
		// If no baseURL, maybe return path relative to storage base?
		// Or just the filename, assuming client knows the base?
		// Let's return path relative to uploads for now.
		accessiblePath = filepath.Join("uploads", uniqueFilename) // Assumes /uploads is served
	}

	logger.Info().Str("filename", fileHeader.Filename).Str("saved_as", uniqueFilename).Str("accessible_path", accessiblePath).Msg("File saved successfully")
	return accessiblePath, nil
}

// TODO: Add DeleteFile method if needed
// func (ls *LocalStorage) DeleteFile(filePath string) error { ... }
