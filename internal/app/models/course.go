package models

// Course represents a course offered by a department.
type Course struct {
	ID           int64   `json:"id" db:"id"`
	DepartmentID int64   `json:"departmentId" db:"department_id"`
	Code         string  `json:"code" db:"code"`
	Name         string  `json:"name" db:"name"`
	Description  *string `json:"description,omitempty" db:"description"` // Nullable
	Credits      int     `json:"credits" db:"credits"`

	// Relations (populated when needed)
	Department *Department `json:"department,omitempty"`
}
