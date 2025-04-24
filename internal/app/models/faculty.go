package models

// Faculty represents a faculty at the university
type Faculty struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}
 