package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const uniqueViolationCode = "23505"

var _ core.CourseRepository = (*PostgresCourseRepository)(nil)

type PostgresCourseRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresCourseRepository(pool *pgxpool.Pool) *PostgresCourseRepository {
	return &PostgresCourseRepository{pool: pool}
}

func (repo *PostgresCourseRepository) Save(course domain.Course) error {
	_, err := repo.pool.Exec(
		context.Background(),
		`INSERT INTO courses (
			id,
			title,
			slug,
			description,
			instructor_id,
			status,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			title = EXCLUDED.title,
			slug = EXCLUDED.slug,
			description = EXCLUDED.description,
			instructor_id = EXCLUDED.instructor_id,
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at`,
		course.ID().String(),
		course.Title(),
		course.Slug().String(),
		course.Description(),
		course.InstructorID().String(),
		course.Status().String(),
		course.CreatedAt(),
		course.UpdatedAt(),
	)

	return mapCourseRepositoryError(err)
}

func (repo *PostgresCourseRepository) FindByID(id domain.CourseID) (domain.Course, error) {
	course, err := scanCourse(repo.pool.QueryRow(
		context.Background(),
		selectCourseSQL+` WHERE id = $1`,
		id.String(),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Course{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Course{}, err
	}

	return course, nil
}

func (repo *PostgresCourseRepository) FindBySlug(slug domain.Slug) (domain.Course, error) {
	course, err := scanCourse(repo.pool.QueryRow(
		context.Background(),
		selectCourseSQL+` WHERE slug = $1`,
		slug.String(),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Course{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Course{}, err
	}

	return course, nil
}

func (repo *PostgresCourseRepository) FindAll(filter core.CourseFilter) ([]domain.Course, error) {
	ctx := context.Background()
	var (
		rows pgx.Rows
		err  error
	)

	if filter.Status != nil {
		rows, err = repo.pool.Query(ctx, selectCourseSQL+` WHERE status = $1 ORDER BY created_at DESC`, filter.Status.String())
	} else {
		rows, err = repo.pool.Query(ctx, selectCourseSQL+` ORDER BY created_at DESC`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	courses := []domain.Course{}
	for rows.Next() {
		course, err := scanCourse(rows)
		if err != nil {
			return nil, err
		}

		courses = append(courses, course)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}

func (repo *PostgresCourseRepository) Delete(id domain.CourseID) error {
	tag, err := repo.pool.Exec(context.Background(), `DELETE FROM courses WHERE id = $1`, id.String())
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

const selectCourseSQL = `SELECT
	id,
	title,
	slug,
	description,
	instructor_id,
	status,
	created_at,
	updated_at
FROM courses`

type courseScanner interface {
	Scan(dest ...any) error
}

func scanCourse(scanner courseScanner) (domain.Course, error) {
	var (
		idValue           string
		title             string
		slugValue         string
		description       string
		instructorIDValue string
		statusValue       string
		createdAt         time.Time
		updatedAt         time.Time
	)

	if err := scanner.Scan(
		&idValue,
		&title,
		&slugValue,
		&description,
		&instructorIDValue,
		&statusValue,
		&createdAt,
		&updatedAt,
	); err != nil {
		return domain.Course{}, err
	}

	id, err := domain.NewCourseID(idValue)
	if err != nil {
		return domain.Course{}, err
	}

	slug, err := domain.NewSlug(slugValue)
	if err != nil {
		return domain.Course{}, err
	}

	instructorID, err := domain.NewInstructorID(instructorIDValue)
	if err != nil {
		return domain.Course{}, err
	}

	status, err := domain.NewCourseStatus(statusValue)
	if err != nil {
		return domain.Course{}, err
	}

	return domain.RestoreCourse(id, title, slug, description, instructorID, status, createdAt, updatedAt)
}

func mapCourseRepositoryError(err error) error {
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && isSlugUniqueViolation(pgErr) {
		return domain.ErrSlugTaken
	}

	return err
}

func isSlugUniqueViolation(err *pgconn.PgError) bool {
	if err.Code != uniqueViolationCode {
		return false
	}

	constraintName := strings.ToLower(err.ConstraintName)
	return constraintName == "" || strings.Contains(constraintName, "slug")
}
