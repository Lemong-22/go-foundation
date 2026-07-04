package course

import (
	"context"
	"errors"
)

// ErrNotFound dikembalikan saat entri tidak ditemukan di storage.
var ErrNotFound = errors.New("course not found")

// CourseRepository adalah kontrak fungsi yang harus dipatuhi oleh database mana pun.
type CourseRepository interface {
	Save(ctx context.Context, c *Course) error
	FindByID(ctx context.Context, id string) (*Course, error)
	FindAll(ctx context.Context) ([]Course, error)
	Delete(ctx context.Context, id string) error
}
