package models

type Executive struct {
	ID        int    `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Position  string `json:"position,omitempty"`
	Email     string `json:"email,omitempty"`
}
