package enums

// RoleType defines the user role type
type RoleType string

const (
	RoleStudent    RoleType = "STUDENT"
	RoleInstructor RoleType = "INSTRUCTOR"
)

// Term represents a semester term
type Term string

// Term constants
const (
	TermFall   Term = "FALL"
	TermSpring Term = "SPRING"
)