package course

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresCourseRepository adalah struct nyata yang memegang kolam koneksi DB.
type PostgresCourseRepository struct {
	db *pgxpool.Pool
}

// NewPostgresCourseRepository bertindak sebagai fungsi pembuat (constructor).
func NewPostgresCourseRepository(db *pgxpool.Pool) *PostgresCourseRepository {
	return &PostgresCourseRepository{db: db}
}

// Save bertugas memasukkan data Course ke dalam tabel Postgres.
// Kalau id sudah ada, dia akan update title/slug/description/status/updated_at.
// updated_at dikunci ke c.UpdatedAt (bukan c.CreatedAt) supaya caller bebas atur timestamp.
func (r *PostgresCourseRepository) Save(ctx context.Context, c *Course) error {
	const query = `
		INSERT INTO courses (id, title, slug, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE
		SET title = EXCLUDED.title,
		    slug = EXCLUDED.slug,
		    description = EXCLUDED.description,
		    status = EXCLUDED.status,
		    updated_at = EXCLUDED.updated_at`

	_, err := r.db.Exec(ctx, query,
		c.ID, c.Title, c.Slug, c.Description, c.Status, c.CreatedAt, c.UpdatedAt,
	)
	return err
}

// FindByID menarik satu baris berdasarkan id.
// Mengembalikan ErrNotFound kalau row tidak ada.
func (r *PostgresCourseRepository) FindByID(ctx context.Context, id string) (*Course, error) {
	const query = `
		SELECT id, title, slug, description, status, created_at, updated_at
		FROM courses
		WHERE id = $1`

	var c Course
	err := r.db.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Title, &c.Slug, &c.Description, &c.Status, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query FindByID: %w", err)
	}
	return &c, nil
}

// FindAll bertugas menarik seluruh baris data course dari Postgres.
func (r *PostgresCourseRepository) FindAll(ctx context.Context) ([]Course, error) {
	const query = `
		SELECT id, title, slug, description, status, created_at, updated_at
		FROM courses
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := make([]Course, 0)
	for rows.Next() {
		var c Course
		if err := rows.Scan(
			&c.ID, &c.Title, &c.Slug, &c.Description, &c.Status, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return courses, nil
}

// Delete menghapus course berdasarkan id.
// Mengembalikan ErrNotFound kalau id tidak ada (RowsAffected == 0).
func (r *PostgresCourseRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM courses WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("exec Delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// SaveLesson memasukkan Lesson baru ke tabel Postgres.
// Kalau id sudah ada, akan upsert via ON CONFLICT pakai EXCLUDED.*
// (sama pola dengan Save(Course)) — caller bebas atur CreatedAt/UpdatedAt.
// order_index disimpan apa adanya dari struct (default 0 kalau gak di-set).
func (r *PostgresCourseRepository) SaveLesson(ctx context.Context, l *Lesson) error {
	const query = `
		INSERT INTO lessons (id, course_id, title, slug, content, order_index, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE
		SET course_id   = EXCLUDED.course_id,
		    title       = EXCLUDED.title,
		    slug        = EXCLUDED.slug,
		    content     = EXCLUDED.content,
		    order_index = EXCLUDED.order_index,
		    updated_at  = EXCLUDED.updated_at`

	_, err := r.db.Exec(ctx, query,
		l.ID, l.CourseID, l.Title, l.Slug, l.Content, l.OrderIndex, l.CreatedAt, l.UpdatedAt,
	)
	return err
}

// FindLessonsByCourseID menarik semua lesson milik course tertentu.
// Diurutkan by order_index ASC, lalu created_at ASC sebagai tie-breaker
// supaya output stabil kalau ada dua lesson dengan order_index sama.
// Return slice kosong (bukan nil error) kalau course ada tapi belum punya lesson.
func (r *PostgresCourseRepository) FindLessonsByCourseID(ctx context.Context, courseID string) ([]Lesson, error) {
	const query = `
		SELECT id, course_id, title, slug, content, order_index, created_at, updated_at
		FROM lessons
		WHERE course_id = $1
		ORDER BY order_index ASC, created_at ASC`

	rows, err := r.db.Query(ctx, query, courseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessons := make([]Lesson, 0)
	for rows.Next() {
		var l Lesson
		if err := rows.Scan(
			&l.ID, &l.CourseID, &l.Title, &l.Slug, &l.Content, &l.OrderIndex, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		lessons = append(lessons, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return lessons, nil
}

// DeleteLesson menghapus lesson berdasarkan id.
// Mengembalikan ErrNotFound kalau id tidak ada (RowsAffected == 0).
// Catatan: ini tidak akan kena FK cascade — lessons yang dihapus langsung
// lewat sini tidak memicu delete di courses (FK ada di arah sebaliknya).
func (r *PostgresCourseRepository) DeleteLesson(ctx context.Context, id string) error {
	const query = `DELETE FROM lessons WHERE id = $1`

	tag, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("exec DeleteLesson: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
