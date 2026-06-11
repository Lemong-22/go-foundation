package cli

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	questionIDValue      = "550e8400-e29b-41d4-a716-446655440050"
	otherQuestionIDValue = "550e8400-e29b-41d4-a716-446655440051"
)

func TestQuizCommandExposesRequiredSubcommands(t *testing.T) {
	command := NewQuizCommand(QuizCommandOptions{Service: &quizServiceFake{}})

	wantCommands := [][]string{
		{"create"},
		{"list"},
		{"get"},
		{"update"},
		{"delete"},
		{"question", "add"},
		{"question", "list"},
		{"question", "get"},
		{"question", "update"},
		{"question", "remove"},
		{"question", "reorder"},
	}
	for _, path := range wantCommands {
		if _, _, err := command.Find(path); err != nil {
			t.Fatalf("expected quiz command path %v to exist, got %v", path, err)
		}
	}
}

func TestQuizCreateMapsFlagsToDTO(t *testing.T) {
	service := &quizServiceFake{createOut: core.CreateQuizOutput{ID: quizIDValue}}
	renderer := &quizRendererFake{}

	err := executeCourseCommand(
		NewQuizCommand(QuizCommandOptions{Service: service, Renderer: renderer}),
		"create",
		"--course-id", courseIDValue,
		"--title", "Basics Quiz",
		"--pass-threshold", "0.8",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "create" || service.createIn.CourseID != courseIDValue || service.createIn.Title != "Basics Quiz" {
		t.Fatalf("expected quiz create input, got called=%q input=%+v", service.called, service.createIn)
	}
	if service.createIn.PassThreshold == nil || *service.createIn.PassThreshold != 0.8 {
		t.Fatalf("expected pass threshold pointer, got %v", service.createIn.PassThreshold)
	}
	if renderer.createdQuizID != quizIDValue {
		t.Fatalf("expected renderer to receive created quiz id")
	}
}

func TestQuizCRUDCommandsMapInputsAndOutputs(t *testing.T) {
	quiz := quizViewFixture()
	detail := quizDetailFixture()
	service := &quizServiceFake{
		listOut:   core.ListQuizzesOutput{Quizzes: []core.QuizView{quiz}},
		getOut:    core.GetQuizOutput{Quiz: detail},
		updateOut: core.UpdateQuizOutput{ID: quizIDValue},
	}

	tests := []struct {
		name     string
		args     []string
		wantCall string
		assert   func(t *testing.T, renderer *quizRendererFake)
	}{
		{
			name:     "list",
			args:     []string{"list", "--course-id", courseIDValue, "--output", "json"},
			wantCall: "list",
			assert: func(t *testing.T, renderer *quizRendererFake) {
				t.Helper()
				if renderer.quizListFormat != "json" || len(renderer.quizzes) != 1 || renderer.quizzes[0] != quiz {
					t.Fatalf("expected quiz list renderer, got %+v", renderer)
				}
				if service.listIn.CourseID != courseIDValue {
					t.Fatalf("expected list course id, got %+v", service.listIn)
				}
			},
		},
		{
			name:     "get",
			args:     []string{"get", quizIDValue, "-o", "quiet"},
			wantCall: "get",
			assert: func(t *testing.T, renderer *quizRendererFake) {
				t.Helper()
				if renderer.quizFormat != "quiet" || renderer.quiz.ID != quizIDValue {
					t.Fatalf("expected quiz renderer, got %+v", renderer)
				}
				if service.getIn.ID != quizIDValue {
					t.Fatalf("expected get id, got %+v", service.getIn)
				}
			},
		},
		{
			name:     "update",
			args:     []string{"update", quizIDValue, "--title", "Advanced Quiz", "--pass-threshold", "0.9"},
			wantCall: "update",
			assert: func(t *testing.T, renderer *quizRendererFake) {
				t.Helper()
				if renderer.updatedQuizID != quizIDValue {
					t.Fatalf("expected updated quiz id, got %+v", renderer)
				}
				if service.updateIn.ID != quizIDValue || service.updateIn.Title == nil || *service.updateIn.Title != "Advanced Quiz" {
					t.Fatalf("expected update title, got %+v", service.updateIn)
				}
				if service.updateIn.PassThreshold == nil || *service.updateIn.PassThreshold != 0.9 {
					t.Fatalf("expected update pass threshold, got %+v", service.updateIn)
				}
			},
		},
		{
			name:     "delete",
			args:     []string{"delete", quizIDValue, "--force"},
			wantCall: "delete",
			assert: func(t *testing.T, renderer *quizRendererFake) {
				t.Helper()
				if renderer.confirmation != "quiz deleted" {
					t.Fatalf("expected delete confirmation, got %q", renderer.confirmation)
				}
				if service.deleteIn.ID != quizIDValue {
					t.Fatalf("expected delete id, got %+v", service.deleteIn)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := &quizRendererFake{}
			err := executeCourseCommand(
				NewQuizCommand(QuizCommandOptions{Service: service, Renderer: renderer}),
				test.args...,
			)
			if err != nil {
				t.Fatalf("expected command to succeed, got %v", err)
			}
			if service.called != test.wantCall {
				t.Fatalf("expected %q call, got %q", test.wantCall, service.called)
			}
			test.assert(t, renderer)
		})
	}
}

func TestQuizDeletePrintsEmbeddingLessonIDs(t *testing.T) {
	service := &quizServiceFake{err: quizInUseError(t)}
	var stderr bytes.Buffer
	command := NewQuizCommand(QuizCommandOptions{Service: service})
	command.SetArgs([]string{"delete", quizIDValue, "--force"})
	command.SetOut(io.Discard)
	command.SetErr(&stderr)
	command.SilenceUsage = true
	command.SilenceErrors = true

	err := command.Execute()
	if !errors.Is(err, domain.ErrQuizInUse) {
		t.Fatalf("expected quiz in use error, got %v", err)
	}
	if !strings.Contains(stderr.String(), lessonIDValue) || !strings.Contains(stderr.String(), otherLessonIDValue) {
		t.Fatalf("expected embedded lesson ids in stderr, got %q", stderr.String())
	}
}

func TestQuestionAddMapsFlagsToDTO(t *testing.T) {
	position := 1
	service := &quizServiceFake{addQuestionOut: core.AddQuestionOutput{ID: questionIDValue}}
	renderer := &quizRendererFake{}

	err := executeCourseCommand(
		NewQuizCommand(QuizCommandOptions{Service: service, Renderer: renderer}),
		"question",
		"add",
		"--quiz-id", quizIDValue,
		"--type", "multiple",
		"--prompt", "Pick two",
		"--option", "A",
		"--option", "B",
		"--correct", "0",
		"--correct", "1",
		"--explanation", "A and B",
		"--position", "1",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	wantOptions := []string{"A", "B"}
	wantCorrect := []int{0, 1}
	if service.called != "add-question" || service.addQuestionIn.QuizID != quizIDValue || service.addQuestionIn.Type != "multiple" {
		t.Fatalf("expected add question input, got called=%q input=%+v", service.called, service.addQuestionIn)
	}
	if !reflect.DeepEqual(service.addQuestionIn.Options, wantOptions) || !reflect.DeepEqual(service.addQuestionIn.CorrectIndices, wantCorrect) {
		t.Fatalf("expected options/correct to map, got %+v", service.addQuestionIn)
	}
	if service.addQuestionIn.Position == nil || *service.addQuestionIn.Position != position {
		t.Fatalf("expected explicit position, got %v", service.addQuestionIn.Position)
	}
	if renderer.createdQuestionID != questionIDValue {
		t.Fatalf("expected renderer to receive created question id")
	}
}

func TestQuestionAddRequiresOptionsAndCorrectIndices(t *testing.T) {
	service := &quizServiceFake{}

	err := executeCourseCommand(
		NewQuizCommand(QuizCommandOptions{Service: service}),
		"question",
		"add",
		"--quiz-id", quizIDValue,
		"--type", "single",
		"--prompt", "Pick one",
	)
	if !errors.Is(err, ErrRequiredFlagMissing) {
		t.Fatalf("expected required flag error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

func TestQuestionUpdateMapsChangedFlagsToDTO(t *testing.T) {
	service := &quizServiceFake{updateQuestionOut: core.UpdateQuestionOutput{ID: questionIDValue}}
	renderer := &quizRendererFake{}

	err := executeCourseCommand(
		NewQuizCommand(QuizCommandOptions{Service: service, Renderer: renderer}),
		"question",
		"update",
		questionIDValue,
		"--prompt", "Updated prompt",
		"--option", "A",
		"--option", "C",
		"--correct", "1",
		"--explanation", "",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "update-question" || service.updateQuestionIn.ID != questionIDValue {
		t.Fatalf("expected update question call, got called=%q input=%+v", service.called, service.updateQuestionIn)
	}
	if service.updateQuestionIn.Prompt == nil || *service.updateQuestionIn.Prompt != "Updated prompt" {
		t.Fatalf("expected prompt pointer, got %+v", service.updateQuestionIn)
	}
	if service.updateQuestionIn.Options == nil || !reflect.DeepEqual(*service.updateQuestionIn.Options, []string{"A", "C"}) {
		t.Fatalf("expected options pointer, got %+v", service.updateQuestionIn)
	}
	if service.updateQuestionIn.CorrectIndices == nil || !reflect.DeepEqual(*service.updateQuestionIn.CorrectIndices, []int{1}) {
		t.Fatalf("expected correct indices pointer, got %+v", service.updateQuestionIn)
	}
	if service.updateQuestionIn.Explanation == nil || *service.updateQuestionIn.Explanation != "" {
		t.Fatalf("expected empty explanation pointer, got %+v", service.updateQuestionIn)
	}
	if renderer.updatedQuestionID != questionIDValue {
		t.Fatalf("expected renderer to receive updated question id")
	}
}

func TestQuestionReadRemoveAndReorderCommandsMapDTOs(t *testing.T) {
	question := questionViewFixture()
	service := &quizServiceFake{
		listQuestionsOut:  core.ListQuestionsOutput{Questions: []core.QuestionView{question}},
		getQuestionOut:    core.GetQuestionOutput{Question: question},
		updateQuestionOut: core.UpdateQuestionOutput{ID: questionIDValue},
	}

	tests := []struct {
		name     string
		args     []string
		wantCall string
		assert   func(t *testing.T, renderer *quizRendererFake)
	}{
		{
			name:     "list",
			args:     []string{"question", "list", "--quiz-id", quizIDValue, "--output", "json"},
			wantCall: "list-questions",
			assert: func(t *testing.T, renderer *quizRendererFake) {
				t.Helper()
				if renderer.questionListFormat != "json" || len(renderer.questions) != 1 {
					t.Fatalf("expected question list renderer, got %+v", renderer)
				}
				if service.listQuestionsIn.QuizID != quizIDValue {
					t.Fatalf("expected list quiz id, got %+v", service.listQuestionsIn)
				}
			},
		},
		{
			name:     "get",
			args:     []string{"question", "get", questionIDValue, "-o", "quiet"},
			wantCall: "get-question",
			assert: func(t *testing.T, renderer *quizRendererFake) {
				t.Helper()
				if renderer.questionFormat != "quiet" || renderer.question.ID != questionIDValue {
					t.Fatalf("expected question renderer, got %+v", renderer)
				}
				if service.getQuestionIn.ID != questionIDValue {
					t.Fatalf("expected get question id, got %+v", service.getQuestionIn)
				}
			},
		},
		{
			name:     "remove",
			args:     []string{"question", "remove", questionIDValue, "--force"},
			wantCall: "remove-question",
			assert: func(t *testing.T, renderer *quizRendererFake) {
				t.Helper()
				if renderer.confirmation != "quiz question removed" {
					t.Fatalf("expected remove confirmation, got %q", renderer.confirmation)
				}
				if service.removeQuestionIn.ID != questionIDValue {
					t.Fatalf("expected remove question id, got %+v", service.removeQuestionIn)
				}
			},
		},
		{
			name: "reorder",
			args: []string{
				"question",
				"reorder",
				"--quiz-id", quizIDValue,
				"--order", questionIDValue + ":1," + otherQuestionIDValue + ":0",
			},
			wantCall: "reorder-questions",
			assert: func(t *testing.T, renderer *quizRendererFake) {
				t.Helper()
				if renderer.confirmation != "quiz questions reordered" {
					t.Fatalf("expected reorder confirmation, got %q", renderer.confirmation)
				}
				want := []core.QuestionPlacementDTO{
					{QuestionID: questionIDValue, Position: 1},
					{QuestionID: otherQuestionIDValue, Position: 0},
				}
				if service.reorderQuestionsIn.QuizID != quizIDValue || !reflect.DeepEqual(service.reorderQuestionsIn.Order, want) {
					t.Fatalf("expected reorder input %+v, got %+v", want, service.reorderQuestionsIn)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := &quizRendererFake{}
			err := executeCourseCommand(
				NewQuizCommand(QuizCommandOptions{Service: service, Renderer: renderer}),
				test.args...,
			)
			if err != nil {
				t.Fatalf("expected command to succeed, got %v", err)
			}
			if service.called != test.wantCall {
				t.Fatalf("expected %q call, got %q", test.wantCall, service.called)
			}
			test.assert(t, renderer)
		})
	}
}

func TestQuestionReorderRejectsInvalidOrder(t *testing.T) {
	service := &quizServiceFake{}

	err := executeCourseCommand(
		NewQuizCommand(QuizCommandOptions{Service: service}),
		"question",
		"reorder",
		"--quiz-id", quizIDValue,
		"--order", questionIDValue,
	)
	if !errors.Is(err, ErrInvalidQuestionOrder) {
		t.Fatalf("expected invalid question order, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

type quizServiceFake struct {
	called    string
	callCount int
	err       error

	createIn           core.CreateQuizInput
	createOut          core.CreateQuizOutput
	listIn             core.ListQuizzesInput
	listOut            core.ListQuizzesOutput
	getIn              core.GetQuizInput
	getOut             core.GetQuizOutput
	updateIn           core.UpdateQuizInput
	updateOut          core.UpdateQuizOutput
	deleteIn           core.DeleteQuizInput
	addQuestionIn      core.AddQuestionInput
	addQuestionOut     core.AddQuestionOutput
	listQuestionsIn    core.ListQuestionsInput
	listQuestionsOut   core.ListQuestionsOutput
	getQuestionIn      core.GetQuestionInput
	getQuestionOut     core.GetQuestionOutput
	updateQuestionIn   core.UpdateQuestionInput
	updateQuestionOut  core.UpdateQuestionOutput
	removeQuestionIn   core.RemoveQuestionInput
	reorderQuestionsIn core.ReorderQuestionsInput
}

func (service *quizServiceFake) CreateQuiz(in core.CreateQuizInput) (core.CreateQuizOutput, error) {
	service.record("create")
	service.createIn = in
	if service.err != nil {
		return core.CreateQuizOutput{}, service.err
	}

	return service.createOut, nil
}

func (service *quizServiceFake) ListQuizzes(in core.ListQuizzesInput) (core.ListQuizzesOutput, error) {
	service.record("list")
	service.listIn = in
	if service.err != nil {
		return core.ListQuizzesOutput{}, service.err
	}

	return service.listOut, nil
}

func (service *quizServiceFake) GetQuiz(in core.GetQuizInput) (core.GetQuizOutput, error) {
	service.record("get")
	service.getIn = in
	if service.err != nil {
		return core.GetQuizOutput{}, service.err
	}

	return service.getOut, nil
}

func (service *quizServiceFake) UpdateQuiz(in core.UpdateQuizInput) (core.UpdateQuizOutput, error) {
	service.record("update")
	service.updateIn = in
	if service.err != nil {
		return core.UpdateQuizOutput{}, service.err
	}

	return service.updateOut, nil
}

func (service *quizServiceFake) DeleteQuiz(in core.DeleteQuizInput) error {
	service.record("delete")
	service.deleteIn = in
	return service.err
}

func (service *quizServiceFake) AddQuestion(in core.AddQuestionInput) (core.AddQuestionOutput, error) {
	service.record("add-question")
	service.addQuestionIn = in
	if service.err != nil {
		return core.AddQuestionOutput{}, service.err
	}

	return service.addQuestionOut, nil
}

func (service *quizServiceFake) ListQuestions(in core.ListQuestionsInput) (core.ListQuestionsOutput, error) {
	service.record("list-questions")
	service.listQuestionsIn = in
	if service.err != nil {
		return core.ListQuestionsOutput{}, service.err
	}

	return service.listQuestionsOut, nil
}

func (service *quizServiceFake) GetQuestion(in core.GetQuestionInput) (core.GetQuestionOutput, error) {
	service.record("get-question")
	service.getQuestionIn = in
	if service.err != nil {
		return core.GetQuestionOutput{}, service.err
	}

	return service.getQuestionOut, nil
}

func (service *quizServiceFake) UpdateQuestion(in core.UpdateQuestionInput) (core.UpdateQuestionOutput, error) {
	service.record("update-question")
	service.updateQuestionIn = in
	if service.err != nil {
		return core.UpdateQuestionOutput{}, service.err
	}

	return service.updateQuestionOut, nil
}

func (service *quizServiceFake) RemoveQuestion(in core.RemoveQuestionInput) error {
	service.record("remove-question")
	service.removeQuestionIn = in
	return service.err
}

func (service *quizServiceFake) ReorderQuestions(in core.ReorderQuestionsInput) error {
	service.record("reorder-questions")
	service.reorderQuestionsIn = in
	return service.err
}

func (service *quizServiceFake) record(called string) {
	service.called = called
	service.callCount++
}

type quizRendererFake struct {
	createdQuizID      string
	updatedQuizID      string
	createdQuestionID  string
	updatedQuestionID  string
	quizListFormat     string
	quizFormat         string
	questionListFormat string
	questionFormat     string
	quizzes            []core.QuizView
	quiz               core.QuizDetailView
	questions          []core.QuestionView
	question           core.QuestionView
	confirmation       string
}

func (renderer *quizRendererFake) RenderCreatedQuiz(id string) error {
	renderer.createdQuizID = id
	return nil
}

func (renderer *quizRendererFake) RenderQuizList(format string, quizzes []core.QuizView) error {
	renderer.quizListFormat = format
	renderer.quizzes = quizzes
	return nil
}

func (renderer *quizRendererFake) RenderQuiz(format string, quiz core.QuizDetailView) error {
	renderer.quizFormat = format
	renderer.quiz = quiz
	return nil
}

func (renderer *quizRendererFake) RenderUpdatedQuiz(id string) error {
	renderer.updatedQuizID = id
	return nil
}

func (renderer *quizRendererFake) RenderCreatedQuestion(id string) error {
	renderer.createdQuestionID = id
	return nil
}

func (renderer *quizRendererFake) RenderQuestionList(format string, questions []core.QuestionView) error {
	renderer.questionListFormat = format
	renderer.questions = questions
	return nil
}

func (renderer *quizRendererFake) RenderQuestion(format string, question core.QuestionView) error {
	renderer.questionFormat = format
	renderer.question = question
	return nil
}

func (renderer *quizRendererFake) RenderUpdatedQuestion(id string) error {
	renderer.updatedQuestionID = id
	return nil
}

func (renderer *quizRendererFake) RenderConfirmation(message string) error {
	renderer.confirmation = message
	return nil
}

func quizViewFixture() core.QuizView {
	now := time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)
	return core.QuizView{
		ID:            quizIDValue,
		CourseID:      courseIDValue,
		Title:         "Basics Quiz",
		PassThreshold: 0.7,
		QuestionCount: 1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

func quizDetailFixture() core.QuizDetailView {
	return core.QuizDetailView{
		QuizView:  quizViewFixture(),
		Questions: []core.QuestionView{questionViewFixture()},
	}
}

func questionViewFixture() core.QuestionView {
	return core.QuestionView{
		ID:             questionIDValue,
		QuizID:         quizIDValue,
		Type:           "single",
		Prompt:         "Pick one",
		Explanation:    "Because A",
		Options:        []string{"A", "B"},
		CorrectIndices: []int{0},
		Position:       0,
	}
}

func quizInUseError(t *testing.T) error {
	t.Helper()

	lessonID, err := domain.NewLessonID(lessonIDValue)
	if err != nil {
		t.Fatalf("expected lesson id, got %v", err)
	}
	otherLessonID, err := domain.NewLessonID(otherLessonIDValue)
	if err != nil {
		t.Fatalf("expected lesson id, got %v", err)
	}

	return domain.NewQuizInUseError([]domain.LessonID{lessonID, otherLessonID})
}
