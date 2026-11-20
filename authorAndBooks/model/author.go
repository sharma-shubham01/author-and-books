package model

type Author struct {
	ID        int    `json:"ID"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Dob       string `json:"dob"`
	PenName   string `json:"pen_name"`
}
