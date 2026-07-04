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
