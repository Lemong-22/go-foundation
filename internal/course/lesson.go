package course

import "time"

// Lesson mewakili data bab/materi di dalam sebuah kursus
type Lesson struct {
	ID         string    `json:"id"`
	CourseID   string    `json:"course_id"`
	Title      string    `json:"title"`
	Slug       string    `json:"slug"`
	Content    string    `json:"content"`
	OrderIndex int       `json:"order_index"` // Urutan materi (Bab 1, Bab 2, dst)
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
