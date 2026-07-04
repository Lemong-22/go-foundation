package course

import "context"

// CourseRepository adalah kontrak fungsi yang harus dipatuhi oleh database mana pun
type CourseRepository interface {
	Save(ctx context.Context, c *Course) error
	FindAll(ctx context.Context) ([]Course, error)
}
