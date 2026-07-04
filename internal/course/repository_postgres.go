package course

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresCourseRepository adalah struct nyata yang memegang kolam koneksi DB
type PostgresCourseRepository struct {
	db *pgxpool.Pool
}

// NewPostgresCourseRepository bertindak sebagai fungsi pembuat (constructor)
func NewPostgresCourseRepository(db *pgxpool.Pool) *PostgresCourseRepository {
	return &PostgresCourseRepository{db: db}
}

// Save bertugas memasukkan data Course asli ke dalam tabel Postgres
func (r *PostgresCourseRepository) Save(ctx context.Context, c *Course) error {
	query := `
		INSERT INTO courses (id, title, slug, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE 
		SET title = $2, slug = $3, description = $4, status = $5, updated_at = $6`

	// Eksekusi query pakai SQL Parameterized ($1, $2, dst) biar aman dari SQL Injection
	_, err := r.db.Exec(ctx, query, c.ID, c.Title, c.Slug, c.Description, c.Status, c.CreatedAt, c.UpdatedAt)
	return err
}

// FindAll bertugas menarik seluruh baris data course dari Postgres
func (r *PostgresCourseRepository) FindAll(ctx context.Context) ([]Course, error) {
	query := `SELECT id, title, slug, description, status, created_at, updated_at FROM courses`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // Jangan lupa ditutup biar CS Call Center-nya ga gantung

	var courses []Course
	for rows.Next() {
		var c Course
		// Suntikkan data dari baris database langsung ke alamat memori struct Course
		err := rows.Scan(&c.ID, &c.Title, &c.Slug, &c.Description, &c.Status, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			return nil, err
		}
		courses = append(courses, c)
	}
	return courses, nil
}
