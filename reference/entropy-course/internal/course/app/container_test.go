package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/luxeave/entropy-course/internal/course/core"
)

const validDBURL = "postgres://postgres:postgres@localhost:5432/entropy_course?sslmode=disable"

func TestBuildContainerWiresCourseServicesAndPool(t *testing.T) {
	cfg := Config{
		DBURL:        validDBURL,
		InstructorID: configInstructorID,
		APIToken:     configAPIToken,
	}

	container, err := BuildContainer(context.Background(), cfg)
	if err != nil {
		t.Fatalf("expected container to build, got %v", err)
	}
	defer container.Close()

	if container.Course == nil || container.Lesson == nil || container.Quiz == nil || container.Practice == nil || container.Test == nil || container.Import == nil {
		t.Fatalf("expected course, lesson, quiz, practice, test, and import services to be wired")
	}
	if container.pool == nil {
		t.Fatalf("expected postgres pool to be constructed")
	}
	if container.Config != cfg {
		t.Fatalf("expected config to be retained, got %+v", container.Config)
	}

	var _ core.CourseService = container.Course
	var _ core.LessonService = container.Lesson
	var _ core.QuizService = container.Quiz
	var _ core.PracticeService = container.Practice
	var _ core.TestService = container.Test
	var _ core.ImportService = container.Import
}

func TestBuildContainerRejectsMissingDatabaseURL(t *testing.T) {
	_, err := BuildContainer(context.Background(), Config{})
	if !errors.Is(err, ErrMissingDatabaseURL) {
		t.Fatalf("expected missing database url error, got %v", err)
	}
	if !strings.Contains(err.Error(), "connect db") {
		t.Fatalf("expected connect db context, got %v", err)
	}
}

func TestBuildContainerWrapsPoolConstructionError(t *testing.T) {
	_, err := BuildContainer(context.Background(), Config{DBURL: "postgres://%"})
	if err == nil {
		t.Fatalf("expected invalid pool config to fail")
	}
	if !strings.Contains(err.Error(), "connect db") {
		t.Fatalf("expected connect db context, got %v", err)
	}
}
