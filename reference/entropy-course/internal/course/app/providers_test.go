package app

import (
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
)

func TestUUIDGeneratorProducesValidDistinctIDs(t *testing.T) {
	var _ core.IDGenerator = UUIDGenerator{}

	generator := NewUUIDGenerator()
	courseID := generator.NewCourseID()
	lessonID := generator.NewLessonID()
	blockID := generator.NewBlockID()
	quizID := generator.NewQuizID()
	questionID := generator.NewQuestionID()
	practiceID := generator.NewPracticeID()
	testCaseID := generator.NewTestCaseID()
	testID := generator.NewTestID()
	testItemID := generator.NewTestItemID()

	if courseID.String() == "" ||
		lessonID.String() == "" ||
		blockID.String() == "" ||
		quizID.String() == "" ||
		questionID.String() == "" ||
		practiceID.String() == "" ||
		testCaseID.String() == "" ||
		testID.String() == "" ||
		testItemID.String() == "" {
		t.Fatalf("expected generated ids to be non-empty")
	}
	values := []string{
		courseID.String(),
		lessonID.String(),
		blockID.String(),
		quizID.String(),
		questionID.String(),
		practiceID.String(),
		testCaseID.String(),
		testID.String(),
		testItemID.String(),
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			t.Fatalf("expected generated ids to differ")
		}
		seen[value] = struct{}{}
	}
}

func TestSystemClockReturnsCurrentUTC(t *testing.T) {
	var _ core.Clock = SystemClock{}

	clock := NewSystemClock()
	before := time.Now().UTC().Add(-time.Second)
	now := clock.Now()
	after := time.Now().UTC().Add(time.Second)

	if now.Before(before) || now.After(after) {
		t.Fatalf("expected current time, got %v", now)
	}
	if now.Location() != time.UTC {
		t.Fatalf("expected UTC time, got %v", now.Location())
	}
}
