package domain

import (
	"strings"

	"github.com/google/uuid"
)

type CourseID struct {
	value string
}

func NewCourseID(value string) (CourseID, error) {
	parsed, err := parseUUID("course_id", value)
	if err != nil {
		return CourseID{}, err
	}

	return CourseID{value: parsed}, nil
}

func (id CourseID) String() string {
	return id.value
}

type LessonID struct {
	value string
}

func NewLessonID(value string) (LessonID, error) {
	parsed, err := parseUUID("lesson_id", value)
	if err != nil {
		return LessonID{}, err
	}

	return LessonID{value: parsed}, nil
}

func (id LessonID) String() string {
	return id.value
}

type BlockID struct {
	value string
}

func NewBlockID(value string) (BlockID, error) {
	parsed, err := parseUUID("block_id", value)
	if err != nil {
		return BlockID{}, err
	}

	return BlockID{value: parsed}, nil
}

func (id BlockID) String() string {
	return id.value
}

type QuizID struct {
	value string
}

func NewQuizID(value string) (QuizID, error) {
	parsed, err := parseUUID("quiz_id", value)
	if err != nil {
		return QuizID{}, err
	}

	return QuizID{value: parsed}, nil
}

func (id QuizID) String() string {
	return id.value
}

type PracticeID struct {
	value string
}

func NewPracticeID(value string) (PracticeID, error) {
	parsed, err := parseUUID("practice_id", value)
	if err != nil {
		return PracticeID{}, err
	}

	return PracticeID{value: parsed}, nil
}

func (id PracticeID) String() string {
	return id.value
}

type TestID struct {
	value string
}

func NewTestID(value string) (TestID, error) {
	parsed, err := parseUUID("test_id", value)
	if err != nil {
		return TestID{}, err
	}

	return TestID{value: parsed}, nil
}

func (id TestID) String() string {
	return id.value
}

type QuestionID struct {
	value string
}

func NewQuestionID(value string) (QuestionID, error) {
	parsed, err := parseUUID("question_id", value)
	if err != nil {
		return QuestionID{}, err
	}

	return QuestionID{value: parsed}, nil
}

func (id QuestionID) String() string {
	return id.value
}

type TestCaseID struct {
	value string
}

func NewTestCaseID(value string) (TestCaseID, error) {
	parsed, err := parseUUID("test_case_id", value)
	if err != nil {
		return TestCaseID{}, err
	}

	return TestCaseID{value: parsed}, nil
}

func (id TestCaseID) String() string {
	return id.value
}

type TestItemID struct {
	value string
}

func NewTestItemID(value string) (TestItemID, error) {
	parsed, err := parseUUID("test_item_id", value)
	if err != nil {
		return TestItemID{}, err
	}

	return TestItemID{value: parsed}, nil
}

func (id TestItemID) String() string {
	return id.value
}

type InstructorID struct {
	value string
}

func NewInstructorID(value string) (InstructorID, error) {
	parsed, err := parseUUID("instructor_id", value)
	if err != nil {
		return InstructorID{}, err
	}

	return InstructorID{value: parsed}, nil
}

func (id InstructorID) String() string {
	return id.value
}

func parseUUID(field string, value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", NewValidationError(field, "must not be empty")
	}

	parsed, err := uuid.Parse(trimmed)
	if err != nil {
		return "", NewValidationError(field, "must be a valid UUID")
	}

	return parsed.String(), nil
}
