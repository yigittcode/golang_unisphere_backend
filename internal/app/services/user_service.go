package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog"
	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/models/dto"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
	"github.com/yigit/unisphere/internal/pkg/filestorage"
)

// UserService defines the interface for user operations
type UserService interface {
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	GetUsersByDepartment(ctx context.Context, departmentID int64, roleType *string) ([]*models.User, error)
	GetUserProfile(ctx context.Context, userID int64) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID int64, req *dto.UpdateUserRequest) (*models.User, error)
	UpdateProfilePhoto(ctx context.Context, userID int64, file *multipart.FileHeader) (*models.File, error)
	DeleteProfilePhoto(ctx context.Context, userID int64) error
	GetUsersByFilter(ctx context.Context, filter *dto.UserFilterRequest) ([]*models.User, int64, error)
	GetFileByID(ctx context.Context, fileID int64) (*models.File, error)
}

// userServiceImpl implements UserService
type userServiceImpl struct {
	userRepo       *repositories.UserRepository
	departmentRepo *repositories.DepartmentRepository
	fileRepo       *repositories.FileRepository
	fileStorage    *filestorage.LocalStorage
	logger         zerolog.Logger
}

// NewUserService creates a new UserService
func NewUserService(
	userRepo *repositories.UserRepository,
	departmentRepo *repositories.DepartmentRepository,
	fileRepo *repositories.FileRepository,
	fileStorage *filestorage.LocalStorage,
	logger zerolog.Logger,
) UserService {
	return &userServiceImpl{
		userRepo:       userRepo,
		departmentRepo: departmentRepo,
		fileRepo:       fileRepo,
		fileStorage:    fileStorage,
		logger:         logger,
	}
}

// GetUserByID retrieves a user by ID
func (s *userServiceImpl) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	// Get user by ID
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}
	
	// Get department information if available
	if user.DepartmentID != nil {
		department, err := s.departmentRepo.GetByID(ctx, *user.DepartmentID)
		if err == nil && department != nil {
			user.Department = department
		}
	}
	
	return user, nil
}

// GetUsersByDepartment retrieves users by department ID
func (s *userServiceImpl) GetUsersByDepartment(ctx context.Context, departmentID int64, roleType *string) ([]*models.User, error) {
	// Check if department exists
	department, err := s.departmentRepo.GetByID(ctx, departmentID)
	if err != nil {
		return nil, fmt.Errorf("error finding department: %w", err)
	}
	if department == nil {
		return nil, apperrors.ErrDepartmentNotFound
	}
	
	// Get users by department
	var users []*models.User
	if roleType != nil {
		// Convert string to RoleType
		role := models.RoleType(*roleType)
		users, err = s.userRepo.FindByDepartmentAndRole(ctx, departmentID, role)
	} else {
		users, err = s.userRepo.FindByDepartment(ctx, departmentID)
	}
	
	if err != nil {
		return nil, fmt.Errorf("error finding users by department: %w", err)
	}
	
	return users, nil
}

// GetUserProfile retrieves the profile of a user
func (s *userServiceImpl) GetUserProfile(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	
	// If user has a profile photo, get the file information
	if user.ProfilePhotoFileID != nil {
		file, err := s.fileRepo.GetByID(ctx, *user.ProfilePhotoFileID)
		if err == nil && file != nil {
			// Add file information to user (you might need to extend the user model to include this)
			// This is just a placeholder for the implementation
			s.logger.Debug().Int64("fileID", file.ID).Str("fileURL", file.FileURL).Msg("Found profile photo for user")
		}
	}
	
	return user, nil
}

// UpdateUserProfile updates a user's profile information
func (s *userServiceImpl) UpdateUserProfile(ctx context.Context, userID int64, req *dto.UpdateUserRequest) (*models.User, error) {
	// Get current user information
	currentUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Error().Err(err).Int64("userID", userID).Msg("Error finding user for profile update")
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		// For any other errors
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	if currentUser == nil {
		s.logger.Error().Int64("userID", userID).Msg("User not found for profile update")
		return nil, apperrors.ErrUserNotFound
	}
	
	// Check if email is changing and if it's already in use
	if currentUser.Email != req.Email {
		// Check if email is already in use by another user
		existingUser, err := s.userRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			// If it's "user not found", that's good (email not taken)
			if errors.Is(err, apperrors.ErrUserNotFound) {
				// Email is available - continue with the update
				s.logger.Debug().Str("email", req.Email).Msg("Email is available for use")
			} else {
				// Any other error is a real problem
				s.logger.Error().Err(err).Str("email", req.Email).Msg("Error checking email availability")
				return nil, fmt.Errorf("error checking email availability: %w", err)
			}
		} else if existingUser != nil && existingUser.ID != userID {
			// Email exists and belongs to someone else
			return nil, apperrors.ErrEmailAlreadyExists
		}
	}
	
	// Update user information
	currentUser.FirstName = req.FirstName
	currentUser.LastName = req.LastName
	currentUser.Email = req.Email
	
	// Department ID is no longer part of the update request
	// Users can't change their department through the profile update
	
	// Save updated user
	err = s.userRepo.Update(ctx, currentUser)
	if err != nil {
		return nil, fmt.Errorf("error updating user: %w", err)
	}
	
	// Return updated user
	return currentUser, nil
}

// UpdateProfilePhoto updates a user's profile photo
func (s *userServiceImpl) UpdateProfilePhoto(ctx context.Context, userID int64, fileHeader *multipart.FileHeader) (*models.File, error) {
	// Get user information
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("error finding user: %w", err)
	}
	if user == nil {
		return nil, apperrors.ErrUserNotFound
	}
	
	// Get the content type
	contentType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return nil, fmt.Errorf("file is not an image: %s", contentType)
	}
	
	// Generate a storage path based on user ID
	subPath := fmt.Sprintf("profile_photos/user_%d", userID)
	
	// Upload to storage - LocalStorage will generate a unique filename using UUID
	fileURL, err := s.fileStorage.SaveFileWithPath(fileHeader, subPath)
	if err != nil {
		return nil, fmt.Errorf("error uploading file: %w", err)
	}
	
	// Get just the filename from the URL
	filename := filepath.Base(fileURL)
	
	// Extract relative path from URL
	relativeFilePath := strings.TrimPrefix(fileURL, s.fileStorage.GetBaseURL())
	relativeFilePath = strings.TrimPrefix(relativeFilePath, "/uploads/")
	
	// Create file record
	file := &models.File{
		FileName:     filename,
		FilePath:     relativeFilePath,
		FileURL:      fileURL,
		FileSize:     fileHeader.Size,
		FileType:     contentType,
		ResourceType: models.FileTypeProfilePhoto,
		ResourceID:   userID,
		UploadedBy:   userID,
	}
	
	// Save file metadata first
	fileID, err := s.fileRepo.Create(ctx, file)
	if err != nil {
		// Clean up the uploaded file if metadata save fails
		_ = s.fileStorage.DeleteFile(relativeFilePath)
		return nil, fmt.Errorf("error saving file metadata: %w", err)
	}
	file.ID = fileID
	
	// Delete old profile photo if exists
	if user.ProfilePhotoFileID != nil && *user.ProfilePhotoFileID != fileID {
		oldFile, err := s.fileRepo.GetByID(ctx, *user.ProfilePhotoFileID)
		if err == nil && oldFile != nil {
			// Store the old file path for deletion after user update succeeds
			oldFilePath := oldFile.FilePath
			oldFileID := oldFile.ID
			
			// Update user's profile photo ID first
			err = s.userRepo.UpdateProfilePhotoFileID(ctx, userID, &fileID)
			if err != nil {
				// If user update fails, clean up the new file
				_ = s.fileStorage.DeleteFile(relativeFilePath)
				_ = s.fileRepo.Delete(ctx, fileID)
				return nil, fmt.Errorf("error updating user's profile photo ID: %w", err)
			}
			
			// Now that user update succeeded, delete old file
			_ = s.fileStorage.DeleteFile(oldFilePath)
			_ = s.fileRepo.Delete(ctx, oldFileID)
		}
	} else {
		// No old photo or same photo ID (which shouldn't happen), just update user
		err = s.userRepo.UpdateProfilePhotoFileID(ctx, userID, &fileID)
		if err != nil {
			// If user update fails, clean up the new file
			_ = s.fileStorage.DeleteFile(relativeFilePath)
			_ = s.fileRepo.Delete(ctx, fileID)
			return nil, fmt.Errorf("error updating user's profile photo ID: %w", err)
		}
	}
	
	// Return the file information
	return file, nil
}

// DeleteProfilePhoto deletes a user's profile photo
func (s *userServiceImpl) DeleteProfilePhoto(ctx context.Context, userID int64) error {
	// Get user information
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error finding user: %w", err)
	}
	if user == nil {
		return apperrors.ErrUserNotFound
	}
	
	// If no profile photo, return specific error
	if user.ProfilePhotoFileID == nil {
		return apperrors.NewResourceNotFoundError("Profile photo does not exist for this user")
	}
	
	// Get the file details
	oldFile, err := s.fileRepo.GetByID(ctx, *user.ProfilePhotoFileID)
	if err != nil {
		s.logger.Warn().Err(err).Int64("fileID", *user.ProfilePhotoFileID).
			Msg("Could not find profile photo file record to delete")
		// Even if we can't find the file record, we should still update the user
		// to clear their profile photo ID to fix inconsistent state
	} 
	
	// First, update the user's profile photo ID to null
	err = s.userRepo.UpdateProfilePhotoFileID(ctx, userID, nil)
	if err != nil {
		return fmt.Errorf("error updating user profile: %w", err)
	}
	
	// If we found the file, delete it
	if oldFile != nil {
		// Now that the user record is updated, delete the file
		if delErr := s.fileStorage.DeleteFile(oldFile.FilePath); delErr != nil {
			s.logger.Warn().Err(delErr).Str("filePath", oldFile.FilePath).
				Msg("Failed to delete profile photo file from storage")
		}
		
		// Delete the file record
		if delErr := s.fileRepo.Delete(ctx, *user.ProfilePhotoFileID); delErr != nil {
			s.logger.Warn().Err(delErr).Int64("fileID", *user.ProfilePhotoFileID).
				Msg("Failed to delete profile photo record from database")
		}
	}
	
	return nil
}

// GetFileByID retrieves a file by ID
func (s *userServiceImpl) GetFileByID(ctx context.Context, fileID int64) (*models.File, error) {
	file, err := s.fileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving file: %w", err)
	}
	return file, nil
}

// GetUsersByFilter retrieves users based on filter criteria with pagination
func (s *userServiceImpl) GetUsersByFilter(ctx context.Context, filter *dto.UserFilterRequest) ([]*models.User, int64, error) {
	// Convert role string to RoleType if provided
	var roleType *models.RoleType
	if filter.Role != nil {
		role := models.RoleType(*filter.Role)
		roleType = &role
	}
	
	// Get users by filter
	users, total, err := s.userRepo.FindByFilter(ctx, 
		filter.DepartmentID, 
		roleType, 
		filter.Email, 
		filter.Name,
		filter.Page, 
		filter.PageSize)
	
	if err != nil {
		return nil, 0, fmt.Errorf("error finding users by filter: %w", err)
	}
	
	// Load departments for users
	for _, user := range users {
		if user.DepartmentID != nil {
			department, err := s.departmentRepo.GetByID(ctx, *user.DepartmentID)
			if err == nil && department != nil {
				user.Department = department
			}
		}
	}
	
	return users, total, nil
}

