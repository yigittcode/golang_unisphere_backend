package models

// Model tanımlamaları burada olacak
// Örnek:
// type User struct {
//     ID       string `json:"id"`
//     Username string `json:"username"`
//     Email    string `json:"email"`
// }

// RoleType defines the user role type
type RoleType string

const (
	RoleStudent    RoleType = "STUDENT"    // Original name
	RoleInstructor RoleType = "INSTRUCTOR" // Original name
)

// Term represents a semester term
type Term string

// Term constants
const (
	TermFall   Term = "FALL"   // Original name
	TermSpring Term = "SPRING" // Original name
)

// ClassNoteTerm type removed, use Term instead.
// const (
// 	ClassNoteTermFall   ClassNoteTerm = "FALL"
// 	ClassNoteTermSpring ClassNoteTerm = "SPRING"
// )

// // ClassNoteTerm is defined in class_note.go for now // Removed comment
