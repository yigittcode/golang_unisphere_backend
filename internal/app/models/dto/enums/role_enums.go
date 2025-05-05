package enums

// RoleType defines the user role type
type RoleType string

const (
	RoleStudent    RoleType = "STUDENT"
	RoleInstructor RoleType = "INSTRUCTOR"
	RoleAdmin      RoleType = "ADMIN"
)

// Term represents a semester term
type Term string

// Term constants
const (
	TermFall   Term = "FALL"
	TermSpring Term = "SPRING"
)
