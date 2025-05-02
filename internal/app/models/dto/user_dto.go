package dto


// NOTE: UserResponse is already defined in auth_dto.go
// We'll create an extended version with additional fields

// ExtendedUserResponse represents detailed user information 
// Extends the basic UserResponse with additional fields
type ExtendedUserResponse struct {
	ID                 int64      `json:"id"`
	Email              string     `json:"email"`
	FirstName          string     `json:"firstName"`
	LastName           string     `json:"lastName"`
	Role               string     `json:"role"`
	DepartmentID       *int64     `json:"departmentId,omitempty"`
	ProfilePhotoFileID *int64     `json:"profilePhotoFileId,omitempty"`
	ProfilePhotoURL    string     `json:"profilePhotoUrl,omitempty"`
	IsActive           bool       `json:"isActive"`
}

// UserFilterRequest represents user filtering parameters
type UserFilterRequest struct {
	DepartmentID *int64  `form:"departmentId,omitempty"`
	Role         *string `form:"role,omitempty"`
	Email        *string `form:"email,omitempty"`
	Name         *string `form:"name,omitempty"` // For searching by first or last name
	Page         int     `form:"page,default=1" binding:"min=1"`
	PageSize     int     `form:"pageSize,default=10" binding:"min=1,max=100"`
}

// UserListResponse represents a list of users with pagination
type UserListResponse struct {
	Users []ExtendedUserResponse `json:"users"`
	PaginationInfo
}

// UpdateUserRequest represents user update data
type UpdateUserRequest struct {
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	// DepartmentID removed - users shouldn't be able to change their department
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=8"`
}

// UpdateProfilePhotoResponse represents a successful profile photo update
type UpdateProfilePhotoResponse struct {
	ProfilePhotoFileID int64 `json:"profilePhotoFileId"`
}

