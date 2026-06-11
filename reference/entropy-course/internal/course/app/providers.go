package app

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

type UUIDGenerator struct{}

func NewUUIDGenerator() UUIDGenerator {
	return UUIDGenerator{}
}

func (UUIDGenerator) NewCourseID() domain.CourseID {
	id, err := domain.NewCourseID(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate course id: %v", err))
	}

	return id
}

func (UUIDGenerator) NewLessonID() domain.LessonID {
	id, err := domain.NewLessonID(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate lesson id: %v", err))
	}

	return id
}

func (UUIDGenerator) NewBlockID() domain.BlockID {
	id, err := domain.NewBlockID(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate block id: %v", err))
	}

	return id
}

func (UUIDGenerator) NewQuizID() domain.QuizID {
	id, err := domain.NewQuizID(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate quiz id: %v", err))
	}

	return id
}

func (UUIDGenerator) NewQuestionID() domain.QuestionID {
	id, err := domain.NewQuestionID(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate question id: %v", err))
	}

	return id
}

func (UUIDGenerator) NewPracticeID() domain.PracticeID {
	id, err := domain.NewPracticeID(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate practice id: %v", err))
	}

	return id
}

func (UUIDGenerator) NewTestCaseID() domain.TestCaseID {
	id, err := domain.NewTestCaseID(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate test case id: %v", err))
	}

	return id
}

func (UUIDGenerator) NewTestID() domain.TestID {
	id, err := domain.NewTestID(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate test id: %v", err))
	}

	return id
}

func (UUIDGenerator) NewTestItemID() domain.TestItemID {
	id, err := domain.NewTestItemID(uuid.NewString())
	if err != nil {
		panic(fmt.Sprintf("generate test item id: %v", err))
	}

	return id
}

type SystemClock struct{}

func NewSystemClock() SystemClock {
	return SystemClock{}
}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}
