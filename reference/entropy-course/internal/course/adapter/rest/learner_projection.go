package rest

import (
	"net/http"
	"strings"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	readViewQueryParam = "view"
	instructorReadView = "instructor"
	learnerReadView    = "learner"
)

// learnerReadRequested defines the L2 read contract: omitted view and
// ?view=instructor preserve full-fidelity instructor responses, while
// ?view=learner serializes only the DTOs below at the REST boundary.
func learnerReadRequested(request *http.Request) (bool, error) {
	view := strings.TrimSpace(request.URL.Query().Get(readViewQueryParam))
	switch view {
	case "", instructorReadView:
		return false, nil
	case learnerReadView:
		return true, nil
	default:
		return false, domain.NewValidationError(readViewQueryParam, "must be learner or instructor")
	}
}

func (server *Server) ensureLearnerCoursePublished(courseID string) error {
	out, err := server.course.GetCourse(core.GetCourseInput{ID: courseID})
	if err != nil {
		return err
	}

	if out.Course.Status != domain.Published().String() {
		return domain.ErrNotFound
	}

	return nil
}

type learnerGetQuizOutput struct {
	Quiz learnerQuizDetailView
}

type learnerQuizDetailView struct {
	ID            string
	CourseID      string
	Title         string
	PassThreshold float64
	QuestionCount int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Questions     []learnerQuestionView
}

type learnerQuestionView struct {
	ID       string
	QuizID   string
	Type     string
	Prompt   string
	Options  []string
	Position int
}

type learnerGetPracticeOutput struct {
	Practice learnerPracticeDetailView
}

type learnerPracticeDetailView struct {
	ID          string
	CourseID    string
	Title       string
	Language    string
	Prompt      string
	StarterCode string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type learnerGetTestOutput struct {
	Test learnerTestDetailView
}

type learnerTestDetailView struct {
	ID               string
	CourseID         string
	Title            string
	TimeLimitMinutes *int
	PassThreshold    float64
	ItemCount        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Items            []learnerTestItemView
}

type learnerTestItemView struct {
	ID       string
	TestID   string
	Kind     string
	Position int

	ChoicePrompt  string
	ChoiceType    string
	ChoiceOptions []string

	CodingPrompt string
	Language     string
	StarterCode  string
}

func learnerQuizOutput(out core.GetQuizOutput) learnerGetQuizOutput {
	return learnerGetQuizOutput{
		Quiz: learnerQuizDetail(out.Quiz),
	}
}

func learnerQuizDetail(quiz core.QuizDetailView) learnerQuizDetailView {
	questions := make([]learnerQuestionView, 0, len(quiz.Questions))
	for _, question := range quiz.Questions {
		questions = append(questions, learnerQuestionView{
			ID:       question.ID,
			QuizID:   question.QuizID,
			Type:     question.Type,
			Prompt:   question.Prompt,
			Options:  question.Options,
			Position: question.Position,
		})
	}

	return learnerQuizDetailView{
		ID:            quiz.ID,
		CourseID:      quiz.CourseID,
		Title:         quiz.Title,
		PassThreshold: quiz.PassThreshold,
		QuestionCount: quiz.QuestionCount,
		CreatedAt:     quiz.CreatedAt,
		UpdatedAt:     quiz.UpdatedAt,
		Questions:     questions,
	}
}

func learnerPracticeOutput(out core.GetPracticeOutput) learnerGetPracticeOutput {
	practice := out.Practice
	return learnerGetPracticeOutput{
		Practice: learnerPracticeDetailView{
			ID:          practice.ID,
			CourseID:    practice.CourseID,
			Title:       practice.Title,
			Language:    practice.Language,
			Prompt:      practice.Prompt,
			StarterCode: practice.StarterCode,
			CreatedAt:   practice.CreatedAt,
			UpdatedAt:   practice.UpdatedAt,
		},
	}
}

func learnerTestOutput(out core.GetTestOutput) learnerGetTestOutput {
	return learnerGetTestOutput{
		Test: learnerTestDetail(out.Test),
	}
}

func learnerTestDetail(test core.TestDetailView) learnerTestDetailView {
	items := make([]learnerTestItemView, 0, len(test.Items))
	for _, item := range test.Items {
		items = append(items, learnerTestItemView{
			ID:            item.ID,
			TestID:        item.TestID,
			Kind:          item.Kind,
			Position:      item.Position,
			ChoicePrompt:  item.ChoicePrompt,
			ChoiceType:    item.ChoiceType,
			ChoiceOptions: item.ChoiceOptions,
			CodingPrompt:  item.CodingPrompt,
			Language:      item.Language,
			StarterCode:   item.StarterCode,
		})
	}

	return learnerTestDetailView{
		ID:               test.ID,
		CourseID:         test.CourseID,
		Title:            test.Title,
		TimeLimitMinutes: test.TimeLimitMinutes,
		PassThreshold:    test.PassThreshold,
		ItemCount:        test.ItemCount,
		CreatedAt:        test.CreatedAt,
		UpdatedAt:        test.UpdatedAt,
		Items:            items,
	}
}
