package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/logger"
)

// Common errors
var (
	ErrUserNotFound     = errors.New("user not found")
	ErrNotInstructor    = errors.New("only instructors can perform this action")
	ErrPermissionDenied = errors.New("you don't have permission for this action")
	ErrResourceNotFound = errors.New("resource not found")
)

// AuthorizationService handles authorization operations
type AuthorizationService struct {
	userRepo      *repositories.UserRepository
	pastExamRepo  *repositories.PastExamRepository
	classNoteRepo *repositories.ClassNoteRepository
}

// NewAuthorizationService creates a new AuthorizationService
func NewAuthorizationService(userRepo *repositories.UserRepository, pastExamRepo *repositories.PastExamRepository, classNoteRepo *repositories.ClassNoteRepository) *AuthorizationService {
	return &AuthorizationService{
		userRepo:      userRepo,
		pastExamRepo:  pastExamRepo,
		classNoteRepo: classNoteRepo,
	}
}

// IsInstructor checks if the user is an instructor
func (s *AuthorizationService) IsInstructor(ctx context.Context, userID int64) (bool, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return false, ErrUserNotFound
		}
		logger.Error().Err(err).Int64("userID", userID).Msg("Error getting user by ID in IsInstructor")
		return false, err
	}
	if user == nil {
		return false, ErrUserNotFound
	}
	return user.RoleType == models.RoleInstructor, nil
}

// ValidateInstructor validates if the user is an instructor or returns an error
func (s *AuthorizationService) ValidateInstructor(ctx context.Context, userID int64) error {
	isInstructor, err := s.IsInstructor(ctx, userID)
	if err != nil {
		return err
	}

	if !isInstructor {
		return ErrNotInstructor
	}

	return nil
}

// CanModifyPastExam checks if the user can modify a past exam
func (s *AuthorizationService) CanModifyPastExam(ctx context.Context, pastExamID, userID int64) (bool, error) {
	// First check if the user is an instructor
	isInstructor, err := s.IsInstructor(ctx, userID)
	if err != nil {
		return false, err
	}

	if !isInstructor {
		return false, nil
	}

	// Get the past exam
	pastExam, err := s.pastExamRepo.GetPastExamByID(ctx, pastExamID)
	if err != nil {
		if errors.Is(err, repositories.ErrPastExamNotFound) || errors.Is(err, repositories.ErrNotFound) {
			return false, ErrResourceNotFound
		}
		logger.Error().Err(err).Int64("pastExamID", pastExamID).Msg("Error getting past exam by ID")
		return false, err
	}

	// Get instructor ID for the current user
	instructor, err := s.userRepo.GetInstructorByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			logger.Warn().Int64("userID", userID).Msg("Instructor record not found for user marked as instructor")
			return false, ErrUserNotFound
		}
		logger.Error().Err(err).Int64("userID", userID).Msg("Error getting instructor by user ID")
		return false, fmt.Errorf("error getting instructor: %w", err)
	}
	if instructor == nil {
		logger.Warn().Int64("userID", userID).Msg("Instructor record is nil for user marked as instructor")
		return false, ErrUserNotFound
	}

	// Check if the user is the owner of this exam by comparing instructor IDs
	return pastExam.InstructorID == instructor.ID, nil
}

// ValidatePastExamOwnership validates if the user owns the past exam or returns an error
func (s *AuthorizationService) ValidatePastExamOwnership(ctx context.Context, pastExamID, userID int64) error {
	canModify, err := s.CanModifyPastExam(ctx, pastExamID, userID)
	if err != nil {
		if errors.Is(err, ErrResourceNotFound) || errors.Is(err, ErrUserNotFound) {
			return err
		}
		logger.Error().Err(err).Int64("pastExamID", pastExamID).Int64("userID", userID).Msg("Unexpected error during past exam ownership validation")
		return fmt.Errorf("failed to check past exam ownership: %w", err)
	}

	if !canModify {
		return ErrPermissionDenied
	}

	return nil
}

// CanModifyClassNote checks if the user can modify (update/delete) a class note.
func (s *AuthorizationService) CanModifyClassNote(ctx context.Context, noteID int64, userID int64) (bool, error) {
	isOwner, err := s.classNoteRepo.CheckNoteOwner(ctx, noteID, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return false, ErrResourceNotFound
		}
		logger.Error().Err(err).Int64("noteID", noteID).Int64("userID", userID).Msg("Error checking class note owner")
		return false, fmt.Errorf("failed to check class note ownership: %w", err)
	}
	return isOwner, nil
}

// ValidateClassNoteOwnership validates if the user owns the class note or returns an error.
func (s *AuthorizationService) ValidateClassNoteOwnership(ctx context.Context, noteID int64, userID int64) error {
	canModify, err := s.CanModifyClassNote(ctx, noteID, userID)
	if err != nil {
		return err
	}
	if !canModify {
		return ErrPermissionDenied
	}
	return nil
}

// GetUserInfo returns user information
func (s *AuthorizationService) GetUserInfo(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		logger.Error().Err(err).Int64("userID", userID).Msg("Error getting user by ID in GetUserInfo")
		return nil, fmt.Errorf("failed to get user information: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return user, nil
}

// GetInstructorByUserID returns instructor information for a user
func (s *AuthorizationService) GetInstructorByUserID(ctx context.Context, userID int64) (*models.Instructor, error) {
	// First validate that the user is an instructor
	isInstructor, err := s.IsInstructor(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !isInstructor {
		return nil, ErrNotInstructor
	}

	// Get instructor information
	instructor, err := s.userRepo.GetInstructorByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			logger.Warn().Int64("userID", userID).Msg("Instructor record not found for user marked as instructor in GetInstructorByUserID")
			return nil, ErrUserNotFound
		}
		logger.Error().Err(err).Int64("userID", userID).Msg("Error getting instructor by user ID in GetInstructorByUserID")
		return nil, fmt.Errorf("failed to get instructor information: %w", err)
	}

	if instructor == nil {
		return nil, ErrUserNotFound
	}

	return instructor, nil
}
