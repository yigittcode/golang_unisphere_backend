package models

import (
	"time"
)

// RoleType defines the user role type
type RoleType string

const (
	// RoleStudent represents a student role
	RoleStudent RoleType = "STUDENT"
	// RoleInstructor represents an instructor role
	RoleInstructor RoleType = "INSTRUCTOR"
)

// User defines the user model
type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Exclude from JSON responses
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	RoleType  RoleType  `json:"roleType"`
	IsActive  bool      `json:"isActive"`
}

// Student defines the student model (extends User)
type Student struct {
	ID             int64       `json:"id"`
	UserID         int64       `json:"userId"`
	StudentID      string      `json:"studentId"`
	DepartmentID   int64       `json:"departmentId"`
	GraduationYear *int        `json:"graduationYear,omitempty"`
	User           *User       `json:"user,omitempty"`
	Department     *Department `json:"department,omitempty"`
}

// Instructor defines the instructor model (extends User)
type Instructor struct {
	ID           int64       `json:"id"`
	UserID       int64       `json:"userId"`
	DepartmentID int64       `json:"departmentId"`
	Title        string      `json:"title"`
	User         *User       `json:"user,omitempty"`
	Department   *Department `json:"department,omitempty"`
}
 