package models

import "github.com/yigit/unisphere/internal/app/models/dto/enums"

// Model tanımlamaları burada olacak
// Örnek:
// type User struct {
//     ID       string `json:"id"`
//     Username string `json:"username"`
//     Email    string `json:"email"`
// }

// Type aliases for backward compatibility
type RoleType = enums.RoleType
type Term = enums.Term

// Constants for backward compatibility
var (
	RoleStudent    = enums.RoleStudent
	RoleInstructor = enums.RoleInstructor
	RoleAdmin      = enums.RoleAdmin
	TermFall       = enums.TermFall
	TermSpring     = enums.TermSpring
)
