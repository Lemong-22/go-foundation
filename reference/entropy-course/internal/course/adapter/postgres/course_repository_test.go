package postgres

import (
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	courseIDValue     = "550e8400-e29b-41d4-a716-446655440000"
	instructorIDValue = "550e8400-e29b-41d4-a716-446655440010"
)

func TestMapCourseRepositoryErrorMapsSlugUniqueViolation(t *testing.T) {
	err := mapCourseRepositoryError(&pgconn.PgError{
		Code:           uniqueViolationCode,
		ConstraintName: "courses_slug_key",
	})

	if !errors.Is(err, domain.ErrSlugTaken) {
		t.Fatalf("expected slug taken, got %v", err)
	}
}

func TestMapCourseRepositoryErrorPreservesOtherErrors(t *testing.T) {
	errBoom := errors.New("boom")

	if err := mapCourseRepositoryError(errBoom); !errors.Is(err, errBoom) {
		t.Fatalf("expected original error, got %v", err)
	}
}

func TestScanCourseRestoresDomainCourse(t *testing.T) {
	createdAt := time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	course, err := scanCourse(courseScannerFake{
		id:           courseIDValue,
		title:        "Intro to Go",
		slug:         "intro-to-go",
		description:  "Learn Go",
		instructorID: instructorIDValue,
		status:       "published",
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if course.ID().String() != courseIDValue || course.Status() != domain.Published() {
		t.Fatalf("expected persisted identity and status to be restored")
	}
	if course.Slug().String() != "intro-to-go" || course.InstructorID().String() != instructorIDValue {
		t.Fatalf("expected persisted slug and instructor to be restored")
	}
	if !course.CreatedAt().Equal(createdAt) || !course.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected persisted timestamps to be restored")
	}
}

type courseScannerFake struct {
	id           string
	title        string
	slug         string
	description  string
	instructorID string
	status       string
	createdAt    time.Time
	updatedAt    time.Time
	err          error
}

func (scanner courseScannerFake) Scan(dest ...any) error {
	if scanner.err != nil {
		return scanner.err
	}

	values := []any{
		scanner.id,
		scanner.title,
		scanner.slug,
		scanner.description,
		scanner.instructorID,
		scanner.status,
		scanner.createdAt,
		scanner.updatedAt,
	}

	for i, value := range values {
		switch target := dest[i].(type) {
		case *string:
			*target = value.(string)
		case *time.Time:
			*target = value.(time.Time)
		}
	}

	return nil
}
