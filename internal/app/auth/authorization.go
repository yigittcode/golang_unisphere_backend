package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/yigit/unisphere/internal/app/models"
	"github.com/yigit/unisphere/internal/app/repositories"
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
	userRepo     *repositories.UserRepository
	pastExamRepo *repositories.PastExamRepository
}

// NewAuthorizationService creates a new AuthorizationService
func NewAuthorizationService(userRepo *repositories.UserRepository, pastExamRepo *repositories.PastExamRepository) *AuthorizationService {
	return &AuthorizationService{
		userRepo:     userRepo,
		pastExamRepo: pastExamRepo,
	}
}

// IsInstructor checks if the user is an instructor
func (s *AuthorizationService) IsInstructor(ctx context.Context, userID int64) (bool, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
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
		if errors.Is(err, repositories.ErrPastExamNotFound) {
			return false, ErrResourceNotFound
		}
		return false, err
	}

	// Get instructor ID for the current user
	instructor, err := s.userRepo.GetInstructorByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("error getting instructor: %w", err)
	}

	// Check if the user is the owner of this exam by comparing instructor IDs
	return pastExam.InstructorID == instructor.ID, nil
}

// ValidatePastExamOwnership validates if the user owns the past exam or returns an error
func (s *AuthorizationService) ValidatePastExamOwnership(ctx context.Context, pastExamID, userID int64) error {
	canModify, err := s.CanModifyPastExam(ctx, pastExamID, userID)
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
		return nil, fmt.Errorf("failed to get instructor information: %w", err)
	}

	if instructor == nil {
		return nil, ErrUserNotFound
	}

	return instructor, nil
}
