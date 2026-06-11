package core

import (
	"time"

	"github.com/luxeave/entropy-course/internal/course/domain"
)

type IDGenerator interface {
	NewCourseID() domain.CourseID
	NewLessonID() domain.LessonID
	NewBlockID() domain.BlockID
	NewQuizID() domain.QuizID
	NewQuestionID() domain.QuestionID
	NewPracticeID() domain.PracticeID
	NewTestCaseID() domain.TestCaseID
	NewTestID() domain.TestID
	NewTestItemID() domain.TestItemID
}

type Clock interface {
	Now() time.Time
}
