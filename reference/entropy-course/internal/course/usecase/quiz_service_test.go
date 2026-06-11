package usecase

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	quizIDValue     = "550e8400-e29b-41d4-a716-446655440030"
	otherQuizID     = "550e8400-e29b-41d4-a716-446655440031"
	questionIDValue = "550e8400-e29b-41d4-a716-446655440040"
	otherQuestionID = "550e8400-e29b-41d4-a716-446655440041"
)

func TestCreateQuizSavesQuizWithDefaultThreshold(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 8, 0, 0, 0, time.UTC)}
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	quizzes := newQuizRepositoryFake()
	service := newQuizServiceFixture(courses, &lessonRepositoryFake{}, quizzes, clock)

	out, err := service.CreateQuiz(core.CreateQuizInput{
		CourseID: courseIDValue,
		Title:    "  Basics Quiz  ",
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	if out.ID != quizIDValue {
		t.Fatalf("expected id %q, got %q", quizIDValue, out.ID)
	}

	saved := quizzes.savedQuizzes[0]
	if saved.ID().String() != quizIDValue || saved.CourseID().String() != courseIDValue {
		t.Fatalf("expected saved quiz ids")
	}
	if saved.Title() != "Basics Quiz" || saved.PassThreshold() != domain.DefaultPassThreshold() {
		t.Fatalf("expected title and default threshold, got title=%q threshold=%f", saved.Title(), saved.PassThreshold().Float64())
	}
	if len(saved.Questions()) != 0 {
		t.Fatalf("expected new quiz to start without questions")
	}
	if !saved.CreatedAt().Equal(clock.now) || !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected deterministic timestamps")
	}
}

func TestCreateQuizUsesCustomThreshold(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	quizzes := newQuizRepositoryFake()
	service := newQuizServiceFixture(courses, &lessonRepositoryFake{}, quizzes, fixedClock{})
	threshold := 0.9

	if _, err := service.CreateQuiz(core.CreateQuizInput{
		CourseID:      courseIDValue,
		Title:         "Basics Quiz",
		PassThreshold: &threshold,
	}); err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	if got := quizzes.savedQuizzes[0].PassThreshold().Float64(); got != threshold {
		t.Fatalf("expected custom threshold %f, got %f", threshold, got)
	}
}

func TestCreateQuizRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name       string
		input      core.CreateQuizInput
		seedCourse bool
		wantError  error
	}{
		{
			name:      "invalid course id",
			input:     core.CreateQuizInput{CourseID: "bad-id", Title: "Quiz"},
			wantError: domain.ErrValidation,
		},
		{
			name:      "course not found",
			input:     core.CreateQuizInput{CourseID: courseIDValue, Title: "Quiz"},
			wantError: domain.ErrNotFound,
		},
		{
			name:       "invalid threshold",
			input:      core.CreateQuizInput{CourseID: courseIDValue, Title: "Quiz", PassThreshold: floatPointer(1.2)},
			seedCourse: true,
			wantError:  domain.ErrValidation,
		},
		{
			name:       "missing title",
			input:      core.CreateQuizInput{CourseID: courseIDValue, Title: "   "},
			seedCourse: true,
			wantError:  domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			courses := newCourseRepositoryFake()
			if test.seedCourse {
				courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
			}
			quizzes := newQuizRepositoryFake()
			service := newQuizServiceFixture(courses, &lessonRepositoryFake{}, quizzes, fixedClock{})

			_, err := service.CreateQuiz(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(quizzes.savedQuizzes) != 0 {
				t.Fatalf("expected invalid quiz not to be saved")
			}
		})
	}
}

func TestListQuizzesValidatesCourseAndReturnsViews(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	quizzes := newQuizRepositoryFake()
	first := mustQuiz(t, quizIDValue, courseIDValue, "First Quiz", 0.7, mustChoiceQuestion(t, questionIDValue, 0))
	second := mustQuiz(t, otherQuizID, courseIDValue, "Second Quiz", 0.8)
	quizzes.store(first)
	quizzes.store(second)
	service := newQuizServiceFixture(courses, &lessonRepositoryFake{}, quizzes, fixedClock{})

	out, err := service.ListQuizzes(core.ListQuizzesInput{CourseID: courseIDValue})
	if err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}

	if len(out.Quizzes) != 2 {
		t.Fatalf("expected two quizzes, got %d", len(out.Quizzes))
	}
	if out.Quizzes[0].ID != quizIDValue || out.Quizzes[0].QuestionCount != 1 {
		t.Fatalf("expected first quiz view, got %+v", out.Quizzes[0])
	}
	if out.Quizzes[1].ID != otherQuizID || out.Quizzes[1].PassThreshold != 0.8 {
		t.Fatalf("expected second quiz view, got %+v", out.Quizzes[1])
	}
}

func TestListQuizzesRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.ListQuizzesInput
		wantError error
	}{
		{name: "invalid course id", input: core.ListQuizzesInput{CourseID: "bad-id"}, wantError: domain.ErrValidation},
		{name: "course not found", input: core.ListQuizzesInput{CourseID: courseIDValue}, wantError: domain.ErrNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newQuizRepositoryFake(), fixedClock{})

			_, err := service.ListQuizzes(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
		})
	}
}

func TestGetQuizReturnsDetailWithOrderedQuestions(t *testing.T) {
	quizzes := newQuizRepositoryFake()
	quiz := mustQuiz(
		t,
		quizIDValue,
		courseIDValue,
		"Basics Quiz",
		0.7,
		mustChoiceQuestion(t, otherQuestionID, 1),
		mustChoiceQuestion(t, questionIDValue, 0),
	)
	quizzes.store(quiz)
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})

	out, err := service.GetQuiz(core.GetQuizInput{ID: quizIDValue})
	if err != nil {
		t.Fatalf("expected get to succeed, got %v", err)
	}

	if out.Quiz.ID != quizIDValue || out.Quiz.QuestionCount != 2 {
		t.Fatalf("expected quiz detail, got %+v", out.Quiz)
	}
	if out.Quiz.Questions[0].ID != questionIDValue || out.Quiz.Questions[0].QuizID != quizIDValue {
		t.Fatalf("expected questions mapped in position order, got %+v", out.Quiz.Questions)
	}
	if !reflect.DeepEqual(out.Quiz.Questions[0].CorrectIndices, []int{0}) {
		t.Fatalf("expected correct indices to be mapped")
	}
}

func TestGetQuizRejectsFailureModes(t *testing.T) {
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newQuizRepositoryFake(), fixedClock{})

	if _, err := service.GetQuiz(core.GetQuizInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if _, err := service.GetQuiz(core.GetQuizInput{ID: quizIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateQuizChangesProvidedFields(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 9, 0, 0, 0, time.UTC)}
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, clock)
	title := "Advanced Quiz"
	threshold := 0.85

	out, err := service.UpdateQuiz(core.UpdateQuizInput{
		ID:            quizIDValue,
		Title:         &title,
		PassThreshold: &threshold,
	})
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if out.ID != quizIDValue {
		t.Fatalf("expected output id %q, got %q", quizIDValue, out.ID)
	}
	saved := quizzes.savedQuizzes[0]
	if saved.Title() != title || saved.PassThreshold().Float64() != threshold {
		t.Fatalf("expected quiz fields to update")
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestUpdateQuizRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.UpdateQuizInput
		seed      bool
		wantError error
	}{
		{name: "invalid id", input: core.UpdateQuizInput{ID: "bad-id", Title: stringPointer("Updated")}, wantError: domain.ErrValidation},
		{name: "nothing to update", input: core.UpdateQuizInput{ID: quizIDValue}, seed: true, wantError: domain.ErrValidation},
		{name: "not found", input: core.UpdateQuizInput{ID: quizIDValue, Title: stringPointer("Updated")}, wantError: domain.ErrNotFound},
		{name: "empty title", input: core.UpdateQuizInput{ID: quizIDValue, Title: stringPointer("   ")}, seed: true, wantError: domain.ErrValidation},
		{name: "invalid threshold", input: core.UpdateQuizInput{ID: quizIDValue, PassThreshold: floatPointer(-0.1)}, seed: true, wantError: domain.ErrValidation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quizzes := newQuizRepositoryFake()
			if test.seed {
				quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7))
			}
			service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})

			_, err := service.UpdateQuiz(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(quizzes.savedQuizzes) != 0 {
				t.Fatalf("expected invalid update not to be saved")
			}
		})
	}
}

func TestDeleteQuizDeletesWhenNotEmbedded(t *testing.T) {
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})

	if err := service.DeleteQuiz(core.DeleteQuizInput{ID: quizIDValue}); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	if _, exists := quizzes.quizzes[quizIDValue]; exists {
		t.Fatalf("expected quiz to be deleted")
	}
	if len(quizzes.deletedQuizIDs) != 1 || quizzes.deletedQuizIDs[0].String() != quizIDValue {
		t.Fatalf("expected quiz delete to be recorded")
	}
}

func TestDeleteQuizReturnsQuizInUseWhenEmbedded(t *testing.T) {
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7))
	lessons := newLessonRepositoryFake()
	lessons.store(mustLessonFixture(
		t,
		lessonIDValue,
		courseIDValue,
		"Quiz Lesson",
		[]domain.ContentBlock{mustQuizBlock(t, thirdLessonIDValue, 0, quizIDValue)},
		0,
	))
	service := newQuizServiceFixture(newCourseRepositoryFake(), lessons, quizzes, fixedClock{})

	err := service.DeleteQuiz(core.DeleteQuizInput{ID: quizIDValue})
	if !errors.Is(err, domain.ErrQuizInUse) {
		t.Fatalf("expected quiz in use error, got %v", err)
	}

	var inUse domain.QuizInUseError
	if !errors.As(err, &inUse) {
		t.Fatalf("expected quiz in use error details, got %v", err)
	}
	if len(inUse.LessonIDs) != 1 || inUse.LessonIDs[0].String() != lessonIDValue {
		t.Fatalf("expected embedding lesson id, got %+v", inUse.LessonIDs)
	}
	if len(quizzes.deletedQuizIDs) != 0 {
		t.Fatalf("expected embedded quiz not to be deleted")
	}
}

func TestDeleteQuizRejectsFailureModes(t *testing.T) {
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newQuizRepositoryFake(), fixedClock{})

	if err := service.DeleteQuiz(core.DeleteQuizInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if err := service.DeleteQuiz(core.DeleteQuizInput{ID: quizIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestAddQuestionAppendsByDefault(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC)}
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7, mustChoiceQuestion(t, otherQuestionID, 0)))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, clock)

	out, err := service.AddQuestion(core.AddQuestionInput{
		QuizID:         quizIDValue,
		Type:           "single",
		Prompt:         "  Pick A  ",
		Options:        []string{"A", "B"},
		CorrectIndices: []int{0},
		Explanation:    "Because A",
	})
	if err != nil {
		t.Fatalf("expected add question to succeed, got %v", err)
	}

	if out.ID != questionIDValue {
		t.Fatalf("expected question id %q, got %q", questionIDValue, out.ID)
	}
	saved := quizzes.savedQuizzes[0]
	questions := saved.Questions()
	if len(questions) != 2 {
		t.Fatalf("expected two questions, got %d", len(questions))
	}
	added := questions[1]
	if added.ID().String() != questionIDValue || added.Position().Int() != 1 {
		t.Fatalf("expected appended question at position 1, got %+v", added)
	}
	if added.Prompt() != "Pick A" || added.Explanation() != "Because A" {
		t.Fatalf("expected prompt and explanation to be saved")
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestAddQuestionInsertsAtExplicitPosition(t *testing.T) {
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7, mustChoiceQuestion(t, otherQuestionID, 0)))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})
	position := 0

	if _, err := service.AddQuestion(core.AddQuestionInput{
		QuizID:         quizIDValue,
		Type:           "multiple",
		Prompt:         "Pick both",
		Options:        []string{"A", "B", "C"},
		CorrectIndices: []int{0, 1},
		Position:       &position,
	}); err != nil {
		t.Fatalf("expected positioned add to succeed, got %v", err)
	}

	questions := quizzes.savedQuizzes[0].Questions()
	if questions[0].ID().String() != questionIDValue || questions[0].Position().Int() != 0 || !questions[0].Type().IsMultiple() {
		t.Fatalf("expected inserted multiple-choice question at position 0, got %+v", questions[0])
	}
	if questions[1].ID().String() != otherQuestionID || questions[1].Position().Int() != 1 {
		t.Fatalf("expected existing question to shift to position 1, got %+v", questions[1])
	}
}

func TestAddQuestionRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.AddQuestionInput
		seed      bool
		wantError error
	}{
		{
			name:      "invalid quiz id",
			input:     core.AddQuestionInput{QuizID: "bad-id", Type: "single", Prompt: "Pick", Options: []string{"A", "B"}, CorrectIndices: []int{0}},
			wantError: domain.ErrValidation,
		},
		{
			name:      "quiz not found",
			input:     core.AddQuestionInput{QuizID: quizIDValue, Type: "single", Prompt: "Pick", Options: []string{"A", "B"}, CorrectIndices: []int{0}},
			wantError: domain.ErrNotFound,
		},
		{
			name:      "invalid type",
			input:     core.AddQuestionInput{QuizID: quizIDValue, Type: "short", Prompt: "Pick", Options: []string{"A", "B"}, CorrectIndices: []int{0}},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name:      "negative position",
			input:     core.AddQuestionInput{QuizID: quizIDValue, Type: "single", Prompt: "Pick", Options: []string{"A", "B"}, CorrectIndices: []int{0}, Position: intPointer(-1)},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name:      "single choice with multiple correct indices",
			input:     core.AddQuestionInput{QuizID: quizIDValue, Type: "single", Prompt: "Pick", Options: []string{"A", "B"}, CorrectIndices: []int{0, 1}},
			seed:      true,
			wantError: domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quizzes := newQuizRepositoryFake()
			if test.seed {
				quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7))
			}
			service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})

			_, err := service.AddQuestion(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(quizzes.savedQuizzes) != 0 {
				t.Fatalf("expected invalid add not to be saved")
			}
		})
	}
}

func TestListQuestionsReturnsOrderedViews(t *testing.T) {
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(
		t,
		quizIDValue,
		courseIDValue,
		"Basics Quiz",
		0.7,
		mustChoiceQuestion(t, otherQuestionID, 1),
		mustChoiceQuestion(t, questionIDValue, 0),
	))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})

	out, err := service.ListQuestions(core.ListQuestionsInput{QuizID: quizIDValue})
	if err != nil {
		t.Fatalf("expected list questions to succeed, got %v", err)
	}

	if len(out.Questions) != 2 {
		t.Fatalf("expected two questions, got %d", len(out.Questions))
	}
	if out.Questions[0].ID != questionIDValue || out.Questions[0].Position != 0 {
		t.Fatalf("expected questions in position order, got %+v", out.Questions)
	}
	if out.Questions[1].ID != otherQuestionID || out.Questions[1].QuizID != quizIDValue {
		t.Fatalf("expected owning quiz id in question view, got %+v", out.Questions[1])
	}
}

func TestListQuestionsRejectsFailureModes(t *testing.T) {
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newQuizRepositoryFake(), fixedClock{})

	if _, err := service.ListQuestions(core.ListQuestionsInput{QuizID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.ListQuestions(core.ListQuestionsInput{QuizID: quizIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGetQuestionFindsOwningQuiz(t *testing.T) {
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7, mustChoiceQuestion(t, questionIDValue, 0)))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})

	out, err := service.GetQuestion(core.GetQuestionInput{ID: questionIDValue})
	if err != nil {
		t.Fatalf("expected get question to succeed, got %v", err)
	}

	if out.Question.ID != questionIDValue || out.Question.QuizID != quizIDValue {
		t.Fatalf("expected question view from owning quiz, got %+v", out.Question)
	}
	if out.Question.Type != "single" || out.Question.Prompt != "Pick one" {
		t.Fatalf("expected question content to be mapped, got %+v", out.Question)
	}
}

func TestGetQuestionRejectsFailureModes(t *testing.T) {
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newQuizRepositoryFake(), fixedClock{})

	if _, err := service.GetQuestion(core.GetQuestionInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.GetQuestion(core.GetQuestionInput{ID: questionIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateQuestionChangesProvidedFields(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 11, 0, 0, 0, time.UTC)}
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7, mustChoiceQuestion(t, questionIDValue, 0)))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, clock)
	prompt := "Updated?"
	options := []string{"A", "B", "C"}
	correct := []int{2}
	explanation := "Because C"

	out, err := service.UpdateQuestion(core.UpdateQuestionInput{
		ID:             questionIDValue,
		Prompt:         &prompt,
		Options:        &options,
		CorrectIndices: &correct,
		Explanation:    &explanation,
	})
	if err != nil {
		t.Fatalf("expected update question to succeed, got %v", err)
	}

	if out.ID != questionIDValue {
		t.Fatalf("expected output id %q, got %q", questionIDValue, out.ID)
	}
	if len(quizzes.savedQuizzes) != 1 {
		t.Fatalf("expected one save, got %d", len(quizzes.savedQuizzes))
	}
	saved := quizzes.savedQuizzes[0]
	question, err := saved.Question(mustQuestionID(questionIDValue))
	if err != nil {
		t.Fatalf("expected saved question, got %v", err)
	}
	if question.Prompt() != prompt || question.Explanation() != explanation {
		t.Fatalf("expected prompt and explanation to update")
	}
	if !reflect.DeepEqual(question.Options(), options) || !reflect.DeepEqual(question.CorrectIndices(), correct) {
		t.Fatalf("expected options and correct indices to update atomically")
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestUpdateQuestionSupportsMultipleChoiceContent(t *testing.T) {
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7, mustMultipleChoiceQuestion(t, questionIDValue, 0)))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})
	options := []string{"A", "B", "C"}
	correct := []int{0, 2}

	if _, err := service.UpdateQuestion(core.UpdateQuestionInput{
		ID:             questionIDValue,
		Options:        &options,
		CorrectIndices: &correct,
	}); err != nil {
		t.Fatalf("expected multiple-choice content update to succeed, got %v", err)
	}

	question, err := quizzes.savedQuizzes[0].Question(mustQuestionID(questionIDValue))
	if err != nil {
		t.Fatalf("expected saved question, got %v", err)
	}
	if !reflect.DeepEqual(question.CorrectIndices(), correct) {
		t.Fatalf("expected multiple correct indices, got %+v", question.CorrectIndices())
	}
}

func TestUpdateQuestionRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.UpdateQuestionInput
		seed      bool
		wantError error
	}{
		{name: "invalid id", input: core.UpdateQuestionInput{ID: "bad-id", Prompt: stringPointer("Updated")}, wantError: domain.ErrValidation},
		{name: "nothing to update", input: core.UpdateQuestionInput{ID: questionIDValue}, seed: true, wantError: domain.ErrValidation},
		{name: "options without correct indices", input: core.UpdateQuestionInput{ID: questionIDValue, Options: stringSlicePointer([]string{"A", "B"})}, seed: true, wantError: domain.ErrValidation},
		{name: "correct indices without options", input: core.UpdateQuestionInput{ID: questionIDValue, CorrectIndices: intSlicePointer([]int{0})}, seed: true, wantError: domain.ErrValidation},
		{name: "not found", input: core.UpdateQuestionInput{ID: questionIDValue, Prompt: stringPointer("Updated")}, wantError: domain.ErrNotFound},
		{name: "empty prompt", input: core.UpdateQuestionInput{ID: questionIDValue, Prompt: stringPointer("   ")}, seed: true, wantError: domain.ErrValidation},
		{name: "invalid content", input: core.UpdateQuestionInput{ID: questionIDValue, Options: stringSlicePointer([]string{"A"}), CorrectIndices: intSlicePointer([]int{0})}, seed: true, wantError: domain.ErrValidation},
		{name: "single choice with multiple correct indices", input: core.UpdateQuestionInput{ID: questionIDValue, Options: stringSlicePointer([]string{"A", "B"}), CorrectIndices: intSlicePointer([]int{0, 1})}, seed: true, wantError: domain.ErrValidation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quizzes := newQuizRepositoryFake()
			if test.seed {
				quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7, mustChoiceQuestion(t, questionIDValue, 0)))
			}
			service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})

			_, err := service.UpdateQuestion(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(quizzes.savedQuizzes) != 0 {
				t.Fatalf("expected invalid update not to be saved")
			}
		})
	}
}

func TestRemoveQuestionCompactsPositions(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)}
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(
		t,
		quizIDValue,
		courseIDValue,
		"Basics Quiz",
		0.7,
		mustChoiceQuestion(t, questionIDValue, 0),
		mustChoiceQuestion(t, otherQuestionID, 1),
	))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, clock)

	if err := service.RemoveQuestion(core.RemoveQuestionInput{ID: questionIDValue}); err != nil {
		t.Fatalf("expected remove question to succeed, got %v", err)
	}

	saved := quizzes.savedQuizzes[0]
	questions := saved.Questions()
	if len(questions) != 1 {
		t.Fatalf("expected one remaining question, got %d", len(questions))
	}
	if questions[0].ID().String() != otherQuestionID || questions[0].Position().Int() != 0 {
		t.Fatalf("expected remaining question compacted to position 0, got %+v", questions[0])
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestRemoveQuestionRejectsFailureModes(t *testing.T) {
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newQuizRepositoryFake(), fixedClock{})

	if err := service.RemoveQuestion(core.RemoveQuestionInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if err := service.RemoveQuestion(core.RemoveQuestionInput{ID: questionIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestReorderQuestionsSavesPermutation(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 13, 0, 0, 0, time.UTC)}
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(
		t,
		quizIDValue,
		courseIDValue,
		"Basics Quiz",
		0.7,
		mustChoiceQuestion(t, questionIDValue, 0),
		mustChoiceQuestion(t, otherQuestionID, 1),
	))
	service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, clock)

	err := service.ReorderQuestions(core.ReorderQuestionsInput{
		QuizID: quizIDValue,
		Order: []core.QuestionPlacementDTO{
			{QuestionID: otherQuestionID, Position: 0},
			{QuestionID: questionIDValue, Position: 1},
		},
	})
	if err != nil {
		t.Fatalf("expected reorder to succeed, got %v", err)
	}

	saved := quizzes.savedQuizzes[0]
	questions := saved.Questions()
	if questions[0].ID().String() != otherQuestionID || questions[1].ID().String() != questionIDValue {
		t.Fatalf("expected questions reordered by position, got %+v", questions)
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestReorderQuestionsRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.ReorderQuestionsInput
		seed      bool
		wantError error
	}{
		{name: "invalid quiz id", input: core.ReorderQuestionsInput{QuizID: "bad-id"}, wantError: domain.ErrValidation},
		{name: "quiz not found", input: core.ReorderQuestionsInput{QuizID: quizIDValue}, wantError: domain.ErrNotFound},
		{
			name: "unknown question",
			input: core.ReorderQuestionsInput{
				QuizID: quizIDValue,
				Order: []core.QuestionPlacementDTO{
					{QuestionID: questionIDValue, Position: 0},
					{QuestionID: "550e8400-e29b-41d4-a716-446655440099", Position: 1},
				},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "duplicate position",
			input: core.ReorderQuestionsInput{
				QuizID: quizIDValue,
				Order: []core.QuestionPlacementDTO{
					{QuestionID: questionIDValue, Position: 0},
					{QuestionID: otherQuestionID, Position: 0},
				},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "missing question",
			input: core.ReorderQuestionsInput{
				QuizID: quizIDValue,
				Order:  []core.QuestionPlacementDTO{{QuestionID: questionIDValue, Position: 0}},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "position gap",
			input: core.ReorderQuestionsInput{
				QuizID: quizIDValue,
				Order: []core.QuestionPlacementDTO{
					{QuestionID: questionIDValue, Position: 0},
					{QuestionID: otherQuestionID, Position: 2},
				},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quizzes := newQuizRepositoryFake()
			if test.seed {
				quizzes.store(mustQuiz(
					t,
					quizIDValue,
					courseIDValue,
					"Basics Quiz",
					0.7,
					mustChoiceQuestion(t, questionIDValue, 0),
					mustChoiceQuestion(t, otherQuestionID, 1),
				))
			}
			service := newQuizServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, quizzes, fixedClock{})

			err := service.ReorderQuestions(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(quizzes.savedQuizzes) != 0 {
				t.Fatalf("expected invalid reorder not to be saved")
			}
		})
	}
}

func newQuizServiceFixture(
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	quizzes *quizRepositoryFake,
	clock fixedClock,
) *QuizService {
	ids := fixedIDGenerator{
		quizID:     mustQuizID(quizIDValue),
		questionID: mustQuestionID(questionIDValue),
	}
	return NewQuizService(courses, lessons, quizzes, ids, clock)
}

type quizRepositoryFake struct {
	quizzes        map[string]domain.Quiz
	savedQuizzes   []domain.Quiz
	deletedQuizIDs []domain.QuizID
	operations     *[]string
}

func newQuizRepositoryFake() *quizRepositoryFake {
	return &quizRepositoryFake{quizzes: make(map[string]domain.Quiz)}
}

func (repo *quizRepositoryFake) Save(quiz domain.Quiz) error {
	repo.savedQuizzes = append(repo.savedQuizzes, quiz)
	repo.store(quiz)

	return nil
}

func (repo *quizRepositoryFake) FindByID(id domain.QuizID) (domain.Quiz, error) {
	quiz, exists := repo.quizzes[id.String()]
	if !exists {
		return domain.Quiz{}, domain.ErrNotFound
	}

	return quiz, nil
}

func (repo *quizRepositoryFake) FindByCourse(courseID domain.CourseID) ([]domain.Quiz, error) {
	quizzes := make([]domain.Quiz, 0, len(repo.quizzes))
	for _, quiz := range repo.quizzes {
		if quiz.CourseID() == courseID {
			quizzes = append(quizzes, quiz)
		}
	}

	return quizzes, nil
}

func (repo *quizRepositoryFake) FindByQuestionID(id domain.QuestionID) (domain.Quiz, error) {
	for _, quiz := range repo.quizzes {
		if _, err := quiz.Question(id); err == nil {
			return quiz, nil
		}
	}

	return domain.Quiz{}, domain.ErrNotFound
}

func (repo *quizRepositoryFake) Delete(id domain.QuizID) error {
	if _, exists := repo.quizzes[id.String()]; !exists {
		return domain.ErrNotFound
	}

	repo.deletedQuizIDs = append(repo.deletedQuizIDs, id)
	delete(repo.quizzes, id.String())

	return nil
}

func (repo *quizRepositoryFake) DeleteByCourse(courseID domain.CourseID) error {
	if repo.operations != nil {
		*repo.operations = append(*repo.operations, "quizzes:"+courseID.String())
	}

	for id, quiz := range repo.quizzes {
		if quiz.CourseID() == courseID {
			delete(repo.quizzes, id)
		}
	}

	return nil
}

func (repo *quizRepositoryFake) store(quiz domain.Quiz) {
	if repo.quizzes == nil {
		repo.quizzes = make(map[string]domain.Quiz)
	}

	repo.quizzes[quiz.ID().String()] = quiz
}

func mustQuiz(t *testing.T, idValue string, courseIDValue string, title string, thresholdValue float64, questions ...domain.ChoiceQuestion) domain.Quiz {
	t.Helper()

	threshold, err := domain.NewPassThreshold(thresholdValue)
	if err != nil {
		t.Fatalf("expected pass threshold fixture, got %v", err)
	}

	quiz, err := domain.NewQuiz(
		mustQuizID(idValue),
		mustCourseID(courseIDValue),
		title,
		threshold,
		questions,
		time.Date(2026, 5, 26, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected quiz fixture, got %v", err)
	}

	return quiz
}

func mustChoiceQuestion(t *testing.T, idValue string, positionValue int) domain.ChoiceQuestion {
	t.Helper()

	question, err := domain.NewChoiceQuestion(
		mustQuestionID(idValue),
		domain.SingleChoice(),
		"Pick one",
		[]string{"A", "B"},
		[]int{0},
		"Because A",
		mustQuestionPosition(positionValue),
	)
	if err != nil {
		t.Fatalf("expected choice question fixture, got %v", err)
	}

	return question
}

func mustMultipleChoiceQuestion(t *testing.T, idValue string, positionValue int) domain.ChoiceQuestion {
	t.Helper()

	question, err := domain.NewChoiceQuestion(
		mustQuestionID(idValue),
		domain.MultipleChoice(),
		"Pick many",
		[]string{"A", "B", "C"},
		[]int{0, 1},
		"Because A and B",
		mustQuestionPosition(positionValue),
	)
	if err != nil {
		t.Fatalf("expected multiple-choice question fixture, got %v", err)
	}

	return question
}

func mustQuizBlock(t *testing.T, idValue string, positionValue int, quizIDValue string) domain.ContentBlock {
	t.Helper()

	block, err := domain.NewQuizBlock(mustBlockID(idValue), mustBlockPosition(positionValue), mustQuizID(quizIDValue))
	if err != nil {
		t.Fatalf("expected quiz block fixture, got %v", err)
	}

	return block
}

func mustQuizID(value string) domain.QuizID {
	id, err := domain.NewQuizID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustQuestionID(value string) domain.QuestionID {
	id, err := domain.NewQuestionID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustQuestionPosition(value int) domain.QuestionPosition {
	position, err := domain.NewQuestionPosition(value)
	if err != nil {
		panic(err)
	}

	return position
}

func mustBlockPosition(value int) domain.BlockPosition {
	position, err := domain.NewBlockPosition(value)
	if err != nil {
		panic(err)
	}

	return position
}

func floatPointer(value float64) *float64 {
	return &value
}

func stringSlicePointer(value []string) *[]string {
	return &value
}

func intSlicePointer(value []int) *[]int {
	return &value
}
