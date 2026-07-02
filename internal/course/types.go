package course

import "time"

// Course ini jadi cetakan kayak interface buat sebuah kursus
type Course struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Status      string    `json:"status"` // misal: "draft" atau "published"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
