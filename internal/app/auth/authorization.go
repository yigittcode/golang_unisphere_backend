package auth

import (
	"context"
	"fmt"

	"github.com/yigit/unisphere/internal/app/repositories"
	"github.com/yigit/unisphere/internal/pkg/apperrors"
)

// AuthorizationService handles authorization checks
type AuthorizationService struct {
	userRepo      *repositories.UserRepository
	classNoteRepo *repositories.ClassNoteRepository
	pastExamRepo  *repositories.PastExamRepository
}

// NewAuthorizationService creates a new AuthorizationService
func NewAuthorizationService(
	userRepo *repositories.UserRepository,
	classNoteRepo *repositories.ClassNoteRepository,
	pastExamRepo *repositories.PastExamRepository,
) *AuthorizationService {
	return &AuthorizationService{
		userRepo:      userRepo,
		classNoteRepo: classNoteRepo,
		pastExamRepo:  pastExamRepo,
	}
}

// ValidateClassNoteOwnership checks if a user has ownership of a class note
func (s *AuthorizationService) ValidateClassNoteOwnership(ctx context.Context, noteID, userID int64) error {
	// Get the class note
	note, err := s.classNoteRepo.GetByID(ctx, noteID)
	if err != nil {
		return fmt.Errorf("error getting class note: %w", err)
	}
	if note == nil {
		return apperrors.ErrClassNoteNotFound
}

	// Check if the user is the owner
	if note.UserID != userID {
		return apperrors.ErrPermissionDenied
	}

	return nil
}

// ValidatePastExamOwnership checks if a user has ownership of a past exam
func (s *AuthorizationService) ValidatePastExamOwnership(ctx context.Context, examID, userID int64) error {
	// Get the past exam
	exam, err := s.pastExamRepo.GetByID(ctx, examID)
	if err != nil {
		return fmt.Errorf("error getting past exam: %w", err)
	}
	if exam == nil {
		return apperrors.ErrPastExamNotFound
	}

	// Check if the user is the owner
	if exam.InstructorID != userID {
		return apperrors.ErrPermissionDenied
	}

	return nil
}

// CanAccessDepartmentContent checks if a user can access department content
func (s *AuthorizationService) CanAccessDepartmentContent(ctx context.Context, userID, departmentID int64) error {
	// Get the user
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("error getting user: %w", err)
	}
	if user == nil {
		return apperrors.ErrUserNotFound
}

	// Check if the user belongs to the department
	if user.DepartmentID == nil || *user.DepartmentID != departmentID {
		return apperrors.ErrPermissionDenied
	}

	return nil
}
