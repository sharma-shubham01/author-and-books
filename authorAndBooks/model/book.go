package model

type Book struct {
	ID            int    `json:"ID"`
	Title         string `json:"title"`
	Author        Author `json:"author"`
	Publication   string `json:"publication"`
	PublishedDate string `json:"published_date"`
}
