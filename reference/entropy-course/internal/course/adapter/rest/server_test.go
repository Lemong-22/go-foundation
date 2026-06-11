package rest

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	apiTokenValue      = "test-token"
	courseIDValue      = "550e8400-e29b-41d4-a716-446655440000"
	lessonIDValue      = "550e8400-e29b-41d4-a716-446655440020"
	otherLessonIDValue = "550e8400-e29b-41d4-a716-446655440021"
	blockIDValue       = "550e8400-e29b-41d4-a716-446655440030"
	otherBlockIDValue  = "550e8400-e29b-41d4-a716-446655440031"
	quizIDValue        = "550e8400-e29b-41d4-a716-446655440040"
	questionIDValue    = "550e8400-e29b-41d4-a716-446655440050"
	otherQuestionID    = "550e8400-e29b-41d4-a716-446655440051"
	practiceIDValue    = "550e8400-e29b-41d4-a716-446655440060"
	testCaseIDValue    = "550e8400-e29b-41d4-a716-446655440070"
	otherTestCaseID    = "550e8400-e29b-41d4-a716-446655440071"
	testIDValue        = "550e8400-e29b-41d4-a716-446655440080"
	testItemIDValue    = "550e8400-e29b-41d4-a716-446655440090"
	otherTestItemID    = "550e8400-e29b-41d4-a716-446655440091"
)

func TestNewServerRequiresDependencies(t *testing.T) {
	if _, err := NewServer(Options{Lesson: &lessonServiceFake{}, Quiz: &quizServiceFake{}, Practice: &practiceServiceFake{}, Test: &testServiceFake{}, Token: apiTokenValue}); !errors.Is(err, ErrMissingCourseService) {
		t.Fatalf("expected missing course service error, got %v", err)
	}
	if _, err := NewServer(Options{Course: courseServiceFake{}, Quiz: &quizServiceFake{}, Practice: &practiceServiceFake{}, Test: &testServiceFake{}, Token: apiTokenValue}); !errors.Is(err, ErrMissingLessonService) {
		t.Fatalf("expected missing lesson service error, got %v", err)
	}
	if _, err := NewServer(Options{Course: courseServiceFake{}, Lesson: &lessonServiceFake{}, Practice: &practiceServiceFake{}, Test: &testServiceFake{}, Token: apiTokenValue}); !errors.Is(err, ErrMissingQuizService) {
		t.Fatalf("expected missing quiz service error, got %v", err)
	}
	if _, err := NewServer(Options{Course: courseServiceFake{}, Lesson: &lessonServiceFake{}, Quiz: &quizServiceFake{}, Test: &testServiceFake{}, Token: apiTokenValue}); !errors.Is(err, ErrMissingPracticeService) {
		t.Fatalf("expected missing practice service error, got %v", err)
	}
	if _, err := NewServer(Options{Course: courseServiceFake{}, Lesson: &lessonServiceFake{}, Quiz: &quizServiceFake{}, Practice: &practiceServiceFake{}, Token: apiTokenValue}); !errors.Is(err, ErrMissingTestService) {
		t.Fatalf("expected missing test service error, got %v", err)
	}
	if _, err := NewServer(Options{Course: courseServiceFake{}, Lesson: &lessonServiceFake{}, Quiz: &quizServiceFake{}, Practice: &practiceServiceFake{}, Test: &testServiceFake{}}); !errors.Is(err, ErrMissingAPIToken) {
		t.Fatalf("expected missing api token error, got %v", err)
	}
}

func TestServerRejectsUnauthorizedRequests(t *testing.T) {
	service := &lessonServiceFake{}
	server := newTestServer(t, service)
	request := httptest.NewRequest(http.MethodGet, "/v1/lessons/"+lessonIDValue+"/blocks", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized response, got %d", response.Code)
	}
	if service.called != "" {
		t.Fatalf("expected unauthorized request not to reach service, got %q", service.called)
	}
}

func TestAddLessonBlockParsesRequestAndEncodesID(t *testing.T) {
	position := 2
	service := &lessonServiceFake{addOut: core.AddLessonBlockOutput{ID: blockIDValue}}
	response := authedRequest(t, service, http.MethodPost, "/v1/lessons/"+lessonIDValue+"/blocks", `{
		"kind": "text",
		"markdown": "## Intro",
		"position": 2
	}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "add" {
		t.Fatalf("expected add call, got %q", service.called)
	}
	want := core.AddLessonBlockInput{
		LessonID: lessonIDValue,
		Kind:     "text",
		Markdown: "## Intro",
		Position: &position,
	}
	if service.addIn.LessonID != want.LessonID || service.addIn.Kind != want.Kind || service.addIn.Markdown != want.Markdown ||
		service.addIn.Position == nil || *service.addIn.Position != *want.Position {
		t.Fatalf("expected add input %+v, got %+v", want, service.addIn)
	}

	var out core.AddLessonBlockOutput
	decodeResponse(t, response, &out)
	if out.ID != blockIDValue {
		t.Fatalf("expected block id response, got %+v", out)
	}
}

func TestAddLessonBlockParsesQuizRef(t *testing.T) {
	service := &lessonServiceFake{addOut: core.AddLessonBlockOutput{ID: blockIDValue}}
	response := authedRequest(t, service, http.MethodPost, "/v1/lessons/"+lessonIDValue+"/blocks", `{
		"kind": "quiz",
		"quizRef": "`+quizIDValue+`"
	}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.addIn.LessonID != lessonIDValue || service.addIn.Kind != "quiz" || service.addIn.QuizRef != quizIDValue {
		t.Fatalf("expected quiz add input, got %+v", service.addIn)
	}
}

func TestAddLessonBlockParsesPracticeRef(t *testing.T) {
	service := &lessonServiceFake{addOut: core.AddLessonBlockOutput{ID: blockIDValue}}
	response := authedRequest(t, service, http.MethodPost, "/v1/lessons/"+lessonIDValue+"/blocks", `{
		"kind": "practice",
		"practiceRef": "`+practiceIDValue+`"
	}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.addIn.LessonID != lessonIDValue || service.addIn.Kind != "practice" || service.addIn.PracticeRef != practiceIDValue {
		t.Fatalf("expected practice add input, got %+v", service.addIn)
	}
}

func TestListLessonBlocksEncodesOutput(t *testing.T) {
	service := &lessonServiceFake{
		listOut: core.ListLessonBlocksOutput{Blocks: []core.BlockView{blockViewFixture()}},
	}
	response := authedRequest(t, service, http.MethodGet, "/v1/lessons/"+lessonIDValue+"/blocks", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "list" || service.listIn.LessonID != lessonIDValue {
		t.Fatalf("expected list lesson blocks input, got called=%q input=%+v", service.called, service.listIn)
	}

	var out core.ListLessonBlocksOutput
	decodeResponse(t, response, &out)
	if len(out.Blocks) != 1 || out.Blocks[0].ID != blockIDValue {
		t.Fatalf("expected block list response, got %+v", out)
	}
}

func TestGetLessonBlockRoutesByBlockID(t *testing.T) {
	service := &lessonServiceFake{
		getOut: core.GetLessonBlockOutput{Block: blockViewFixture()},
	}
	response := authedRequest(t, service, http.MethodGet, "/v1/blocks/"+blockIDValue, "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "get" || service.getIn.ID != blockIDValue {
		t.Fatalf("expected get block input, got called=%q input=%+v", service.called, service.getIn)
	}

	var out core.GetLessonBlockOutput
	decodeResponse(t, response, &out)
	if out.Block.ID != blockIDValue || out.Block.VideoCaption != "Intro video" {
		t.Fatalf("expected block response, got %+v", out)
	}
}

func TestUpdateLessonBlockParsesPointerFields(t *testing.T) {
	service := &lessonServiceFake{updateOut: core.UpdateLessonBlockOutput{ID: blockIDValue}}
	response := authedRequest(t, service, http.MethodPatch, "/v1/blocks/"+blockIDValue, `{
		"markdown": "Updated",
		"videoCaption": ""
	}`)

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "update" || service.updateIn.ID != blockIDValue {
		t.Fatalf("expected update block input, got called=%q input=%+v", service.called, service.updateIn)
	}
	if service.updateIn.Markdown == nil || *service.updateIn.Markdown != "Updated" {
		t.Fatalf("expected markdown pointer to be set, got %+v", service.updateIn)
	}
	if service.updateIn.VideoCaption == nil || *service.updateIn.VideoCaption != "" {
		t.Fatalf("expected empty caption pointer to be set, got %+v", service.updateIn)
	}
	if service.updateIn.VideoProvider != nil || service.updateIn.VideoLocator != nil {
		t.Fatalf("expected unchanged video fields to remain nil, got %+v", service.updateIn)
	}

	var out core.UpdateLessonBlockOutput
	decodeResponse(t, response, &out)
	if out.ID != blockIDValue {
		t.Fatalf("expected update id response, got %+v", out)
	}
}

func TestDeleteLessonBlockMapsToRemove(t *testing.T) {
	service := &lessonServiceFake{}
	response := authedRequest(t, service, http.MethodDelete, "/v1/blocks/"+blockIDValue, "")

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "remove" || service.removeIn.ID != blockIDValue {
		t.Fatalf("expected remove block input, got called=%q input=%+v", service.called, service.removeIn)
	}
}

func TestReorderLessonBlocksParsesOrder(t *testing.T) {
	service := &lessonServiceFake{}
	response := authedRequest(t, service, http.MethodPost, "/v1/lessons/"+lessonIDValue+"/blocks/reorder", `{
		"order": [
			{"blockID": "`+blockIDValue+`", "position": 1},
			{"blockID": "`+otherBlockIDValue+`", "position": 0}
		]
	}`)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "reorder" || service.reorderIn.LessonID != lessonIDValue {
		t.Fatalf("expected reorder input, got called=%q input=%+v", service.called, service.reorderIn)
	}
	if len(service.reorderIn.Order) != 2 || service.reorderIn.Order[0].BlockID != blockIDValue || service.reorderIn.Order[1].Position != 0 {
		t.Fatalf("expected reorder placements, got %+v", service.reorderIn.Order)
	}
}

func TestCreateQuizParsesRequestAndEncodesID(t *testing.T) {
	threshold := 0.8
	service := &quizServiceFake{createOut: core.CreateQuizOutput{ID: quizIDValue}}
	response := authedQuizRequest(t, service, http.MethodPost, "/v1/quizzes", `{
		"courseID": "`+courseIDValue+`",
		"title": "Basics Quiz",
		"passThreshold": 0.8
	}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "create" {
		t.Fatalf("expected create call, got %q", service.called)
	}
	want := core.CreateQuizInput{CourseID: courseIDValue, Title: "Basics Quiz", PassThreshold: &threshold}
	if service.createIn.CourseID != want.CourseID || service.createIn.Title != want.Title ||
		service.createIn.PassThreshold == nil || *service.createIn.PassThreshold != *want.PassThreshold {
		t.Fatalf("expected create quiz input %+v, got %+v", want, service.createIn)
	}

	var out core.CreateQuizOutput
	decodeResponse(t, response, &out)
	if out.ID != quizIDValue {
		t.Fatalf("expected quiz id response, got %+v", out)
	}
}

func TestListQuizzesRoutesByCourseID(t *testing.T) {
	service := &quizServiceFake{
		listOut: core.ListQuizzesOutput{Quizzes: []core.QuizView{quizViewFixture()}},
	}
	response := authedQuizRequest(t, service, http.MethodGet, "/v1/courses/"+courseIDValue+"/quizzes", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "list" || service.listIn.CourseID != courseIDValue {
		t.Fatalf("expected list quizzes input, got called=%q input=%+v", service.called, service.listIn)
	}

	var out core.ListQuizzesOutput
	decodeResponse(t, response, &out)
	if len(out.Quizzes) != 1 || out.Quizzes[0].ID != quizIDValue {
		t.Fatalf("expected quiz list response, got %+v", out)
	}
}

func TestQuizRoutesByQuizID(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		service := &quizServiceFake{getOut: core.GetQuizOutput{Quiz: quizDetailFixture()}}
		response := authedQuizRequest(t, service, http.MethodGet, "/v1/quizzes/"+quizIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "get" || service.getIn.ID != quizIDValue {
			t.Fatalf("expected get quiz input, got called=%q input=%+v", service.called, service.getIn)
		}

		var out core.GetQuizOutput
		decodeResponse(t, response, &out)
		if out.Quiz.ID != quizIDValue || len(out.Quiz.Questions) != 1 {
			t.Fatalf("expected quiz detail response, got %+v", out)
		}
	})

	t.Run("patch", func(t *testing.T) {
		service := &quizServiceFake{updateOut: core.UpdateQuizOutput{ID: quizIDValue}}
		response := authedQuizRequest(t, service, http.MethodPatch, "/v1/quizzes/"+quizIDValue, `{
			"title": "Advanced Quiz",
			"passThreshold": 0.9
		}`)

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "update" || service.updateIn.ID != quizIDValue {
			t.Fatalf("expected update quiz input, got called=%q input=%+v", service.called, service.updateIn)
		}
		if service.updateIn.Title == nil || *service.updateIn.Title != "Advanced Quiz" {
			t.Fatalf("expected title pointer to be set, got %+v", service.updateIn)
		}
		if service.updateIn.PassThreshold == nil || *service.updateIn.PassThreshold != 0.9 {
			t.Fatalf("expected pass threshold pointer to be set, got %+v", service.updateIn)
		}
	})

	t.Run("delete", func(t *testing.T) {
		service := &quizServiceFake{}
		response := authedQuizRequest(t, service, http.MethodDelete, "/v1/quizzes/"+quizIDValue, "")

		if response.Code != http.StatusNoContent {
			t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "delete" || service.deleteIn.ID != quizIDValue {
			t.Fatalf("expected delete quiz input, got called=%q input=%+v", service.called, service.deleteIn)
		}
	})
}

func TestAddAndListQuestionsRouteByQuizID(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		position := 1
		service := &quizServiceFake{addQuestionOut: core.AddQuestionOutput{ID: questionIDValue}}
		response := authedQuizRequest(t, service, http.MethodPost, "/v1/quizzes/"+quizIDValue+"/questions", `{
			"type": "multiple",
			"prompt": "Pick two",
			"options": ["A", "B"],
			"correctIndices": [0, 1],
			"explanation": "A and B",
			"position": 1
		}`)

		if response.Code != http.StatusCreated {
			t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "add-question" || service.addQuestionIn.QuizID != quizIDValue {
			t.Fatalf("expected add question input, got called=%q input=%+v", service.called, service.addQuestionIn)
		}
		if len(service.addQuestionIn.Options) != 2 || len(service.addQuestionIn.CorrectIndices) != 2 {
			t.Fatalf("expected question content to map, got %+v", service.addQuestionIn)
		}
		if service.addQuestionIn.Position == nil || *service.addQuestionIn.Position != position {
			t.Fatalf("expected position pointer, got %+v", service.addQuestionIn)
		}
	})

	t.Run("list", func(t *testing.T) {
		service := &quizServiceFake{
			listQuestionsOut: core.ListQuestionsOutput{Questions: []core.QuestionView{questionViewFixture()}},
		}
		response := authedQuizRequest(t, service, http.MethodGet, "/v1/quizzes/"+quizIDValue+"/questions", "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "list-questions" || service.listQuestionsIn.QuizID != quizIDValue {
			t.Fatalf("expected list questions input, got called=%q input=%+v", service.called, service.listQuestionsIn)
		}

		var out core.ListQuestionsOutput
		decodeResponse(t, response, &out)
		if len(out.Questions) != 1 || out.Questions[0].ID != questionIDValue {
			t.Fatalf("expected question list response, got %+v", out)
		}
	})
}

func TestQuestionRoutesByQuestionID(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		service := &quizServiceFake{getQuestionOut: core.GetQuestionOutput{Question: questionViewFixture()}}
		response := authedQuizRequest(t, service, http.MethodGet, "/v1/questions/"+questionIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "get-question" || service.getQuestionIn.ID != questionIDValue {
			t.Fatalf("expected get question input, got called=%q input=%+v", service.called, service.getQuestionIn)
		}
	})

	t.Run("patch preserves atomic options and correct indices", func(t *testing.T) {
		service := &quizServiceFake{updateQuestionOut: core.UpdateQuestionOutput{ID: questionIDValue}}
		response := authedQuizRequest(t, service, http.MethodPatch, "/v1/questions/"+questionIDValue, `{
			"prompt": "Updated prompt",
			"options": ["A", "C"],
			"correctIndices": [1],
			"explanation": ""
		}`)

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "update-question" || service.updateQuestionIn.ID != questionIDValue {
			t.Fatalf("expected update question input, got called=%q input=%+v", service.called, service.updateQuestionIn)
		}
		if service.updateQuestionIn.Prompt == nil || *service.updateQuestionIn.Prompt != "Updated prompt" {
			t.Fatalf("expected prompt pointer, got %+v", service.updateQuestionIn)
		}
		if service.updateQuestionIn.Options == nil || len(*service.updateQuestionIn.Options) != 2 {
			t.Fatalf("expected options pointer, got %+v", service.updateQuestionIn)
		}
		if service.updateQuestionIn.CorrectIndices == nil || len(*service.updateQuestionIn.CorrectIndices) != 1 || (*service.updateQuestionIn.CorrectIndices)[0] != 1 {
			t.Fatalf("expected correct indices pointer, got %+v", service.updateQuestionIn)
		}
		if service.updateQuestionIn.Explanation == nil || *service.updateQuestionIn.Explanation != "" {
			t.Fatalf("expected empty explanation pointer, got %+v", service.updateQuestionIn)
		}
	})

	t.Run("delete", func(t *testing.T) {
		service := &quizServiceFake{}
		response := authedQuizRequest(t, service, http.MethodDelete, "/v1/questions/"+questionIDValue, "")

		if response.Code != http.StatusNoContent {
			t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "remove-question" || service.removeQuestionIn.ID != questionIDValue {
			t.Fatalf("expected remove question input, got called=%q input=%+v", service.called, service.removeQuestionIn)
		}
	})
}

func TestReorderQuestionsParsesOrder(t *testing.T) {
	service := &quizServiceFake{}
	response := authedQuizRequest(t, service, http.MethodPost, "/v1/quizzes/"+quizIDValue+"/questions/reorder", `{
		"order": [
			{"questionID": "`+questionIDValue+`", "position": 1},
			{"questionID": "`+otherQuestionID+`", "position": 0}
		]
	}`)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "reorder-questions" || service.reorderQuestionsIn.QuizID != quizIDValue {
		t.Fatalf("expected reorder questions input, got called=%q input=%+v", service.called, service.reorderQuestionsIn)
	}
	if len(service.reorderQuestionsIn.Order) != 2 || service.reorderQuestionsIn.Order[0].QuestionID != questionIDValue || service.reorderQuestionsIn.Order[1].Position != 0 {
		t.Fatalf("expected question reorder placements, got %+v", service.reorderQuestionsIn.Order)
	}
}

func TestCreatePracticeParsesRequestAndEncodesID(t *testing.T) {
	service := &practiceServiceFake{createOut: core.CreatePracticeOutput{ID: practiceIDValue}}
	response := authedPracticeRequest(t, service, http.MethodPost, "/v1/practices", `{
		"courseID": "`+courseIDValue+`",
		"title": "FizzBuzz",
		"language": "golang",
		"prompt": "Print fizz buzz",
		"starterCode": "package main",
		"solution": "fmt.Println()"
	}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "create" {
		t.Fatalf("expected create call, got %q", service.called)
	}
	want := core.CreatePracticeInput{
		CourseID:    courseIDValue,
		Title:       "FizzBuzz",
		Language:    "golang",
		Prompt:      "Print fizz buzz",
		StarterCode: "package main",
		Solution:    "fmt.Println()",
	}
	if service.createIn != want {
		t.Fatalf("expected create practice input %+v, got %+v", want, service.createIn)
	}

	var out core.CreatePracticeOutput
	decodeResponse(t, response, &out)
	if out.ID != practiceIDValue {
		t.Fatalf("expected practice id response, got %+v", out)
	}
}

func TestListPracticesRoutesByCourseID(t *testing.T) {
	service := &practiceServiceFake{
		listOut: core.ListPracticesOutput{Practices: []core.PracticeView{practiceViewFixture()}},
	}
	response := authedPracticeRequest(t, service, http.MethodGet, "/v1/courses/"+courseIDValue+"/practices", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "list" || service.listIn.CourseID != courseIDValue {
		t.Fatalf("expected list practices input, got called=%q input=%+v", service.called, service.listIn)
	}

	var out core.ListPracticesOutput
	decodeResponse(t, response, &out)
	if len(out.Practices) != 1 || out.Practices[0].ID != practiceIDValue {
		t.Fatalf("expected practice list response, got %+v", out)
	}
}

func TestPracticeRoutesByPracticeID(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		service := &practiceServiceFake{getOut: core.GetPracticeOutput{Practice: practiceDetailFixture()}}
		response := authedPracticeRequest(t, service, http.MethodGet, "/v1/practices/"+practiceIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "get" || service.getIn.ID != practiceIDValue {
			t.Fatalf("expected get practice input, got called=%q input=%+v", service.called, service.getIn)
		}

		var out core.GetPracticeOutput
		decodeResponse(t, response, &out)
		if out.Practice.ID != practiceIDValue || len(out.Practice.TestCases) != 1 {
			t.Fatalf("expected practice detail response, got %+v", out)
		}
	})

	t.Run("patch", func(t *testing.T) {
		service := &practiceServiceFake{updateOut: core.UpdatePracticeOutput{ID: practiceIDValue}}
		response := authedPracticeRequest(t, service, http.MethodPatch, "/v1/practices/"+practiceIDValue, `{
			"title": "Updated FizzBuzz",
			"prompt": "Updated prompt",
			"starterCode": "",
			"solution": ""
		}`)

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "update" || service.updateIn.ID != practiceIDValue {
			t.Fatalf("expected update practice input, got called=%q input=%+v", service.called, service.updateIn)
		}
		if service.updateIn.Title == nil || *service.updateIn.Title != "Updated FizzBuzz" {
			t.Fatalf("expected title pointer, got %+v", service.updateIn)
		}
		if service.updateIn.Prompt == nil || *service.updateIn.Prompt != "Updated prompt" {
			t.Fatalf("expected prompt pointer, got %+v", service.updateIn)
		}
		if service.updateIn.StarterCode == nil || *service.updateIn.StarterCode != "" {
			t.Fatalf("expected empty starter code pointer, got %+v", service.updateIn)
		}
		if service.updateIn.Solution == nil || *service.updateIn.Solution != "" {
			t.Fatalf("expected empty solution pointer, got %+v", service.updateIn)
		}
	})

	t.Run("delete", func(t *testing.T) {
		service := &practiceServiceFake{}
		response := authedPracticeRequest(t, service, http.MethodDelete, "/v1/practices/"+practiceIDValue, "")

		if response.Code != http.StatusNoContent {
			t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "delete" || service.deleteIn.ID != practiceIDValue {
			t.Fatalf("expected delete practice input, got called=%q input=%+v", service.called, service.deleteIn)
		}
	})
}

func TestAddAndListTestCasesRouteByPracticeID(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		position := 1
		service := &practiceServiceFake{addTestCaseOut: core.AddTestCaseOutput{ID: testCaseIDValue}}
		response := authedPracticeRequest(t, service, http.MethodPost, "/v1/practices/"+practiceIDValue+"/testcases", `{
			"stdin": "",
			"expectedStdout": "",
			"name": "",
			"position": 1
		}`)

		if response.Code != http.StatusCreated {
			t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "add-testcase" || service.addTestCaseIn.PracticeID != practiceIDValue {
			t.Fatalf("expected add test case input, got called=%q input=%+v", service.called, service.addTestCaseIn)
		}
		if service.addTestCaseIn.Stdin != "" || service.addTestCaseIn.ExpectedStdout != "" || service.addTestCaseIn.Name != "" {
			t.Fatalf("expected empty test case strings to map, got %+v", service.addTestCaseIn)
		}
		if service.addTestCaseIn.Position == nil || *service.addTestCaseIn.Position != position {
			t.Fatalf("expected position pointer, got %+v", service.addTestCaseIn)
		}
	})

	t.Run("list", func(t *testing.T) {
		service := &practiceServiceFake{
			listTestCasesOut: core.ListTestCasesOutput{TestCases: []core.TestCaseView{testCaseViewFixture()}},
		}
		response := authedPracticeRequest(t, service, http.MethodGet, "/v1/practices/"+practiceIDValue+"/testcases", "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "list-testcases" || service.listTestCasesIn.PracticeID != practiceIDValue {
			t.Fatalf("expected list test cases input, got called=%q input=%+v", service.called, service.listTestCasesIn)
		}

		var out core.ListTestCasesOutput
		decodeResponse(t, response, &out)
		if len(out.TestCases) != 1 || out.TestCases[0].ID != testCaseIDValue {
			t.Fatalf("expected test case list response, got %+v", out)
		}
	})
}

func TestTestCaseRoutesByTestCaseID(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		service := &practiceServiceFake{getTestCaseOut: core.GetTestCaseOutput{TestCase: testCaseViewFixture()}}
		response := authedPracticeRequest(t, service, http.MethodGet, "/v1/testcases/"+testCaseIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "get-testcase" || service.getTestCaseIn.ID != testCaseIDValue {
			t.Fatalf("expected get test case input, got called=%q input=%+v", service.called, service.getTestCaseIn)
		}
	})

	t.Run("patch preserves empty fields", func(t *testing.T) {
		service := &practiceServiceFake{updateTestCaseOut: core.UpdateTestCaseOutput{ID: testCaseIDValue}}
		response := authedPracticeRequest(t, service, http.MethodPatch, "/v1/testcases/"+testCaseIDValue, `{
			"stdin": "updated input",
			"expectedStdout": "",
			"name": ""
		}`)

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "update-testcase" || service.updateTestCaseIn.ID != testCaseIDValue {
			t.Fatalf("expected update test case input, got called=%q input=%+v", service.called, service.updateTestCaseIn)
		}
		if service.updateTestCaseIn.Stdin == nil || *service.updateTestCaseIn.Stdin != "updated input" {
			t.Fatalf("expected stdin pointer, got %+v", service.updateTestCaseIn)
		}
		if service.updateTestCaseIn.ExpectedStdout == nil || *service.updateTestCaseIn.ExpectedStdout != "" {
			t.Fatalf("expected empty expected stdout pointer, got %+v", service.updateTestCaseIn)
		}
		if service.updateTestCaseIn.Name == nil || *service.updateTestCaseIn.Name != "" {
			t.Fatalf("expected empty name pointer, got %+v", service.updateTestCaseIn)
		}
	})

	t.Run("delete", func(t *testing.T) {
		service := &practiceServiceFake{}
		response := authedPracticeRequest(t, service, http.MethodDelete, "/v1/testcases/"+testCaseIDValue, "")

		if response.Code != http.StatusNoContent {
			t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "remove-testcase" || service.removeTestCaseIn.ID != testCaseIDValue {
			t.Fatalf("expected remove test case input, got called=%q input=%+v", service.called, service.removeTestCaseIn)
		}
	})
}

func TestReorderTestCasesParsesOrder(t *testing.T) {
	service := &practiceServiceFake{}
	response := authedPracticeRequest(t, service, http.MethodPost, "/v1/practices/"+practiceIDValue+"/testcases/reorder", `{
		"order": [
			{"testCaseID": "`+testCaseIDValue+`", "position": 1},
			{"testCaseID": "`+otherTestCaseID+`", "position": 0}
		]
	}`)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "reorder-testcases" || service.reorderTestCasesIn.PracticeID != practiceIDValue {
		t.Fatalf("expected reorder test cases input, got called=%q input=%+v", service.called, service.reorderTestCasesIn)
	}
	if len(service.reorderTestCasesIn.Order) != 2 || service.reorderTestCasesIn.Order[0].TestCaseID != testCaseIDValue || service.reorderTestCasesIn.Order[1].Position != 0 {
		t.Fatalf("expected test case reorder placements, got %+v", service.reorderTestCasesIn.Order)
	}
}

func TestCreateTestParsesRequestAndEncodesID(t *testing.T) {
	timeLimit := 45
	threshold := 0.85
	service := &testServiceFake{createOut: core.CreateTestOutput{ID: testIDValue}}
	response := authedTestRequest(t, service, http.MethodPost, "/v1/tests", `{
		"courseID": "`+courseIDValue+`",
		"title": "Final Test",
		"timeLimitMinutes": 45,
		"passThreshold": 0.85
	}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "create" {
		t.Fatalf("expected create call, got %q", service.called)
	}
	want := core.CreateTestInput{
		CourseID:         courseIDValue,
		Title:            "Final Test",
		TimeLimitMinutes: &timeLimit,
		PassThreshold:    &threshold,
	}
	if service.createIn.CourseID != want.CourseID || service.createIn.Title != want.Title ||
		service.createIn.TimeLimitMinutes == nil || *service.createIn.TimeLimitMinutes != *want.TimeLimitMinutes ||
		service.createIn.PassThreshold == nil || *service.createIn.PassThreshold != *want.PassThreshold {
		t.Fatalf("expected create test input %+v, got %+v", want, service.createIn)
	}

	var out core.CreateTestOutput
	decodeResponse(t, response, &out)
	if out.ID != testIDValue {
		t.Fatalf("expected test id response, got %+v", out)
	}
}

func TestListTestsRoutesByCourseID(t *testing.T) {
	service := &testServiceFake{
		listOut: core.ListTestsOutput{Tests: []core.TestView{testViewFixture()}},
	}
	response := authedTestRequest(t, service, http.MethodGet, "/v1/courses/"+courseIDValue+"/tests", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "list" || service.listIn.CourseID != courseIDValue {
		t.Fatalf("expected list tests input, got called=%q input=%+v", service.called, service.listIn)
	}

	var out core.ListTestsOutput
	decodeResponse(t, response, &out)
	if len(out.Tests) != 1 || out.Tests[0].ID != testIDValue || out.Tests[0].TimeLimitMinutes == nil {
		t.Fatalf("expected test list response, got %+v", out)
	}
}

func TestTestRoutesByTestID(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		service := &testServiceFake{getOut: core.GetTestOutput{Test: testDetailFixture()}}
		response := authedTestRequest(t, service, http.MethodGet, "/v1/tests/"+testIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "get" || service.getIn.ID != testIDValue {
			t.Fatalf("expected get test input, got called=%q input=%+v", service.called, service.getIn)
		}

		var out core.GetTestOutput
		decodeResponse(t, response, &out)
		if out.Test.ID != testIDValue || out.Test.Solution == nil || len(out.Test.Items) != 1 {
			t.Fatalf("expected test detail response, got %+v", out)
		}
	})

	t.Run("patch", func(t *testing.T) {
		service := &testServiceFake{updateOut: core.UpdateTestOutput{ID: testIDValue}}
		response := authedTestRequest(t, service, http.MethodPatch, "/v1/tests/"+testIDValue, `{
			"title": "Updated Final",
			"timeLimitMinutes": 0,
			"passThreshold": 0.9,
			"solutionZipProvider": "url",
			"solutionZipLocator": "https://example.com/updated.zip",
			"solutionVideoProvider": "url",
			"solutionVideoLocator": "https://example.com/video.mp4",
			"solutionVideoCaption": ""
		}`)

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "update" || service.updateIn.ID != testIDValue {
			t.Fatalf("expected update test input, got called=%q input=%+v", service.called, service.updateIn)
		}
		if service.updateIn.Title == nil || *service.updateIn.Title != "Updated Final" {
			t.Fatalf("expected title pointer, got %+v", service.updateIn)
		}
		if service.updateIn.TimeLimitMinutes == nil || *service.updateIn.TimeLimitMinutes != 0 {
			t.Fatalf("expected zero time limit pointer, got %+v", service.updateIn)
		}
		if service.updateIn.PassThreshold == nil || *service.updateIn.PassThreshold != 0.9 {
			t.Fatalf("expected pass threshold pointer, got %+v", service.updateIn)
		}
		if service.updateIn.SolutionVideoCaption == nil || *service.updateIn.SolutionVideoCaption != "" {
			t.Fatalf("expected empty solution caption pointer, got %+v", service.updateIn)
		}
		if service.updateIn.SolutionZipProvider == nil || *service.updateIn.SolutionZipProvider != "url" ||
			service.updateIn.SolutionZipLocator == nil || *service.updateIn.SolutionZipLocator != "https://example.com/updated.zip" ||
			service.updateIn.SolutionVideoProvider == nil || *service.updateIn.SolutionVideoProvider != "url" ||
			service.updateIn.SolutionVideoLocator == nil || *service.updateIn.SolutionVideoLocator != "https://example.com/video.mp4" {
			t.Fatalf("expected all solution fields to map, got %+v", service.updateIn)
		}

		var out core.UpdateTestOutput
		decodeResponse(t, response, &out)
		if out.ID != testIDValue {
			t.Fatalf("expected update id response, got %+v", out)
		}
	})

	t.Run("delete", func(t *testing.T) {
		service := &testServiceFake{}
		response := authedTestRequest(t, service, http.MethodDelete, "/v1/tests/"+testIDValue, "")

		if response.Code != http.StatusNoContent {
			t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "delete" || service.deleteIn.ID != testIDValue {
			t.Fatalf("expected delete test input, got called=%q input=%+v", service.called, service.deleteIn)
		}
	})
}

func TestAddAndListTestItemsRouteByTestID(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		position := 1
		service := &testServiceFake{addItemOut: core.AddTestItemOutput{ID: testItemIDValue}}
		response := authedTestRequest(t, service, http.MethodPost, "/v1/tests/"+testIDValue+"/items", `{
			"kind": "coding",
			"codingPrompt": "Write code",
			"language": "golang",
			"starterCode": "package main",
			"solution": "func main() {}",
			"testCases": [
				{"stdin": "1", "expectedStdout": "1", "name": "sample"}
			],
			"position": 1
		}`)

		if response.Code != http.StatusCreated {
			t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "add-item" || service.addItemIn.TestID != testIDValue {
			t.Fatalf("expected add test item input, got called=%q input=%+v", service.called, service.addItemIn)
		}
		if service.addItemIn.Kind != "coding" || service.addItemIn.CodingPrompt != "Write code" ||
			service.addItemIn.Language != "golang" || service.addItemIn.StarterCode != "package main" ||
			service.addItemIn.Solution != "func main() {}" {
			t.Fatalf("expected coding item payload to map, got %+v", service.addItemIn)
		}
		if len(service.addItemIn.TestCases) != 1 || service.addItemIn.TestCases[0].ExpectedStdout != "1" {
			t.Fatalf("expected coding test cases to map, got %+v", service.addItemIn.TestCases)
		}
		if service.addItemIn.Position == nil || *service.addItemIn.Position != position {
			t.Fatalf("expected position pointer, got %+v", service.addItemIn)
		}
	})

	t.Run("list", func(t *testing.T) {
		service := &testServiceFake{
			listItemsOut: core.ListTestItemsOutput{Items: []core.TestItemView{testItemViewFixture()}},
		}
		response := authedTestRequest(t, service, http.MethodGet, "/v1/tests/"+testIDValue+"/items", "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "list-items" || service.listItemsIn.TestID != testIDValue {
			t.Fatalf("expected list test items input, got called=%q input=%+v", service.called, service.listItemsIn)
		}

		var out core.ListTestItemsOutput
		decodeResponse(t, response, &out)
		if len(out.Items) != 1 || out.Items[0].ID != testItemIDValue {
			t.Fatalf("expected test item list response, got %+v", out)
		}
	})
}

func TestTestItemRoutesByItemID(t *testing.T) {
	t.Run("get", func(t *testing.T) {
		service := &testServiceFake{getItemOut: core.GetTestItemOutput{Item: testItemViewFixture()}}
		response := authedTestRequest(t, service, http.MethodGet, "/v1/test-items/"+testItemIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "get-item" || service.getItemIn.ID != testItemIDValue {
			t.Fatalf("expected get test item input, got called=%q input=%+v", service.called, service.getItemIn)
		}
	})

	t.Run("patch preserves optional item payloads", func(t *testing.T) {
		service := &testServiceFake{updateItemOut: core.UpdateTestItemOutput{ID: testItemIDValue}}
		response := authedTestRequest(t, service, http.MethodPatch, "/v1/test-items/"+testItemIDValue, `{
			"prompt": "Updated prompt",
			"choiceType": "single",
			"options": ["A", "B"],
			"correctIndices": [1],
			"explanation": "",
			"codingPrompt": "Updated coding prompt",
			"language": "rust",
			"starterCode": "",
			"solution": "updated",
			"testCases": [
				{"stdin": "stdin", "expectedStdout": "stdout", "name": "case"}
			]
		}`)

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "update-item" || service.updateItemIn.ID != testItemIDValue {
			t.Fatalf("expected update test item input, got called=%q input=%+v", service.called, service.updateItemIn)
		}
		if service.updateItemIn.Prompt == nil || *service.updateItemIn.Prompt != "Updated prompt" {
			t.Fatalf("expected prompt pointer, got %+v", service.updateItemIn)
		}
		if service.updateItemIn.ChoiceType == nil || *service.updateItemIn.ChoiceType != "single" {
			t.Fatalf("expected choice type pointer, got %+v", service.updateItemIn)
		}
		if service.updateItemIn.Options == nil || len(*service.updateItemIn.Options) != 2 {
			t.Fatalf("expected options pointer, got %+v", service.updateItemIn)
		}
		if service.updateItemIn.CorrectIndices == nil || len(*service.updateItemIn.CorrectIndices) != 1 || (*service.updateItemIn.CorrectIndices)[0] != 1 {
			t.Fatalf("expected correct indices pointer, got %+v", service.updateItemIn)
		}
		if service.updateItemIn.Explanation == nil || *service.updateItemIn.Explanation != "" {
			t.Fatalf("expected empty explanation pointer, got %+v", service.updateItemIn)
		}
		if service.updateItemIn.CodingPrompt == nil || *service.updateItemIn.CodingPrompt != "Updated coding prompt" {
			t.Fatalf("expected coding prompt pointer, got %+v", service.updateItemIn)
		}
		if service.updateItemIn.Language == nil || *service.updateItemIn.Language != "rust" {
			t.Fatalf("expected language pointer, got %+v", service.updateItemIn)
		}
		if service.updateItemIn.StarterCode == nil || *service.updateItemIn.StarterCode != "" {
			t.Fatalf("expected empty starter code pointer, got %+v", service.updateItemIn)
		}
		if service.updateItemIn.Solution == nil || *service.updateItemIn.Solution != "updated" {
			t.Fatalf("expected solution pointer, got %+v", service.updateItemIn)
		}
		if service.updateItemIn.TestCases == nil || len(*service.updateItemIn.TestCases) != 1 || (*service.updateItemIn.TestCases)[0].Name != "case" {
			t.Fatalf("expected coding test cases pointer, got %+v", service.updateItemIn)
		}
	})

	t.Run("delete", func(t *testing.T) {
		service := &testServiceFake{}
		response := authedTestRequest(t, service, http.MethodDelete, "/v1/test-items/"+testItemIDValue, "")

		if response.Code != http.StatusNoContent {
			t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "remove-item" || service.removeItemIn.ID != testItemIDValue {
			t.Fatalf("expected remove test item input, got called=%q input=%+v", service.called, service.removeItemIn)
		}
	})
}

func TestLearnerReadViewProjectsQuizWithoutAnswers(t *testing.T) {
	service := &quizServiceFake{getOut: core.GetQuizOutput{Quiz: quizDetailFixture()}}
	response := authedQuizRequest(t, service, http.MethodGet, "/v1/quizzes/"+quizIDValue+"?view=learner", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "get" || service.getIn.ID != quizIDValue {
		t.Fatalf("expected get quiz input, got called=%q input=%+v", service.called, service.getIn)
	}
	assertJSONOmitsKeys(t, response.Body.Bytes(), "CorrectIndices", "Explanation")

	var out learnerGetQuizOutput
	decodeResponseBytes(t, response.Body.Bytes(), &out)
	if out.Quiz.ID != quizIDValue || len(out.Quiz.Questions) != 1 {
		t.Fatalf("expected learner quiz detail, got %+v", out)
	}
	if out.Quiz.Questions[0].Prompt != "Pick one" || len(out.Quiz.Questions[0].Options) != 2 {
		t.Fatalf("expected learner-visible question prompt/options, got %+v", out.Quiz.Questions[0])
	}
}

func TestLearnerReadViewProjectsPracticeWithoutSolutionOrTestCases(t *testing.T) {
	service := &practiceServiceFake{getOut: core.GetPracticeOutput{Practice: practiceDetailFixture()}}
	response := authedPracticeRequest(t, service, http.MethodGet, "/v1/practices/"+practiceIDValue+"?view=learner", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "get" || service.getIn.ID != practiceIDValue {
		t.Fatalf("expected get practice input, got called=%q input=%+v", service.called, service.getIn)
	}
	assertJSONOmitsKeys(t, response.Body.Bytes(), "Solution", "TestCases", "ExpectedStdout")

	var out learnerGetPracticeOutput
	decodeResponseBytes(t, response.Body.Bytes(), &out)
	if out.Practice.ID != practiceIDValue || out.Practice.Prompt != "Print fizz buzz" || out.Practice.StarterCode != "package main" {
		t.Fatalf("expected learner practice prompt/starter code, got %+v", out)
	}
}

func TestLearnerReadViewProjectsTestWithoutSolutionsOrAnswerKeys(t *testing.T) {
	service := &testServiceFake{getOut: core.GetTestOutput{Test: testDetailWithChoiceAndCodingFixture()}}
	response := authedTestRequest(t, service, http.MethodGet, "/v1/tests/"+testIDValue+"?view=learner", "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "get" || service.getIn.ID != testIDValue {
		t.Fatalf("expected get test input, got called=%q input=%+v", service.called, service.getIn)
	}
	assertJSONOmitsKeys(t, response.Body.Bytes(), "Solution", "ChoiceCorrectIndices", "ChoiceExplanation", "CodingSolution", "TestCases", "ExpectedStdout")

	var out learnerGetTestOutput
	decodeResponseBytes(t, response.Body.Bytes(), &out)
	if out.Test.ID != testIDValue || len(out.Test.Items) != 2 {
		t.Fatalf("expected learner test detail with items, got %+v", out)
	}
	if out.Test.Items[0].ChoicePrompt != "Pick one" || len(out.Test.Items[0].ChoiceOptions) != 2 {
		t.Fatalf("expected learner choice item prompt/options, got %+v", out.Test.Items[0])
	}
	if out.Test.Items[1].CodingPrompt != "Write code" || out.Test.Items[1].StarterCode != "package main" {
		t.Fatalf("expected learner coding item prompt/starter code, got %+v", out.Test.Items[1])
	}
}

func TestLearnerReadViewOmitsAllAnswerBearingFieldNames(t *testing.T) {
	answerFields := []string{
		"Answer",
		"Answers",
		"CorrectAnswer",
		"CorrectAnswers",
		"CorrectIndices",
		"ExpectedStdout",
		"Explanation",
		"Solution",
		"TestCases",
		"TestSolution",
		"ChoiceCorrectIndices",
		"ChoiceExplanation",
		"CodingSolution",
	}

	tests := []struct {
		name     string
		response *httptest.ResponseRecorder
	}{
		{
			name:     "quiz",
			response: authedQuizRequest(t, &quizServiceFake{getOut: core.GetQuizOutput{Quiz: quizDetailFixture()}}, http.MethodGet, "/v1/quizzes/"+quizIDValue+"?view=learner", ""),
		},
		{
			name:     "practice",
			response: authedPracticeRequest(t, &practiceServiceFake{getOut: core.GetPracticeOutput{Practice: practiceDetailFixture()}}, http.MethodGet, "/v1/practices/"+practiceIDValue+"?view=learner", ""),
		},
		{
			name:     "test",
			response: authedTestRequest(t, &testServiceFake{getOut: core.GetTestOutput{Test: testDetailWithChoiceAndCodingFixture()}}, http.MethodGet, "/v1/tests/"+testIDValue+"?view=learner", ""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.response.Code != http.StatusOK {
				t.Fatalf("expected ok response, got %d body=%s", test.response.Code, test.response.Body.String())
			}
			assertJSONOmitsKeys(t, test.response.Body.Bytes(), answerFields...)
		})
	}
}

func TestInstructorReadViewRemainsFullFidelity(t *testing.T) {
	t.Run("default quiz read", func(t *testing.T) {
		service := &quizServiceFake{getOut: core.GetQuizOutput{Quiz: quizDetailFixture()}}
		response := authedQuizRequest(t, service, http.MethodGet, "/v1/quizzes/"+quizIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		assertJSONContainsKeys(t, response.Body.Bytes(), "CorrectIndices", "Explanation")
	})

	t.Run("explicit instructor practice read", func(t *testing.T) {
		service := &practiceServiceFake{getOut: core.GetPracticeOutput{Practice: practiceDetailFixture()}}
		response := authedPracticeRequest(t, service, http.MethodGet, "/v1/practices/"+practiceIDValue+"?view=instructor", "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		assertJSONContainsKeys(t, response.Body.Bytes(), "Solution", "TestCases", "ExpectedStdout")
	})

	t.Run("default test read", func(t *testing.T) {
		service := &testServiceFake{getOut: core.GetTestOutput{Test: testDetailWithChoiceAndCodingFixture()}}
		response := authedTestRequest(t, service, http.MethodGet, "/v1/tests/"+testIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
		}
		assertJSONContainsKeys(t, response.Body.Bytes(), "Solution", "ChoiceCorrectIndices", "ChoiceExplanation", "CodingSolution", "TestCases", "ExpectedStdout")
	})
}

func TestReadViewRejectsUnknownValue(t *testing.T) {
	service := &quizServiceFake{getOut: core.GetQuizOutput{Quiz: quizDetailFixture()}}
	response := authedQuizRequest(t, service, http.MethodGet, "/v1/quizzes/"+quizIDValue+"?view=public", "")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "" {
		t.Fatalf("expected invalid read view not to reach service, got %q", service.called)
	}
}

func TestLearnerReadViewRequiresPublishedCourse(t *testing.T) {
	draftCourse := courseServiceFake{getOut: core.GetCourseOutput{Course: courseViewWithStatus("draft")}}

	t.Run("quiz draft hidden", func(t *testing.T) {
		service := &quizServiceFake{getOut: core.GetQuizOutput{Quiz: quizDetailFixture()}}
		response := authedQuizRequestWithCourse(t, draftCourse, service, http.MethodGet, "/v1/quizzes/"+quizIDValue+"?view=learner", "")

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected not found response for draft learner quiz, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "get" {
			t.Fatalf("expected quiz detail to be read before course status check, got %q", service.called)
		}
	})

	t.Run("practice draft hidden", func(t *testing.T) {
		service := &practiceServiceFake{getOut: core.GetPracticeOutput{Practice: practiceDetailFixture()}}
		response := authedPracticeRequestWithCourse(t, draftCourse, service, http.MethodGet, "/v1/practices/"+practiceIDValue+"?view=learner", "")

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected not found response for draft learner practice, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "get" {
			t.Fatalf("expected practice detail to be read before course status check, got %q", service.called)
		}
	})

	t.Run("test draft hidden", func(t *testing.T) {
		service := &testServiceFake{getOut: core.GetTestOutput{Test: testDetailWithChoiceAndCodingFixture()}}
		response := authedTestRequestWithCourse(t, draftCourse, service, http.MethodGet, "/v1/tests/"+testIDValue+"?view=learner", "")

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected not found response for draft learner test, got %d body=%s", response.Code, response.Body.String())
		}
		if service.called != "get" {
			t.Fatalf("expected test detail to be read before course status check, got %q", service.called)
		}
	})
}

func TestInstructorReadViewStillServesDraftCourseContent(t *testing.T) {
	draftCourse := courseServiceFake{getOut: core.GetCourseOutput{Course: courseViewWithStatus("draft")}}

	t.Run("quiz", func(t *testing.T) {
		service := &quizServiceFake{getOut: core.GetQuizOutput{Quiz: quizDetailFixture()}}
		response := authedQuizRequestWithCourse(t, draftCourse, service, http.MethodGet, "/v1/quizzes/"+quizIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response for instructor quiz, got %d body=%s", response.Code, response.Body.String())
		}
		assertJSONContainsKeys(t, response.Body.Bytes(), "CorrectIndices", "Explanation")
	})

	t.Run("practice", func(t *testing.T) {
		service := &practiceServiceFake{getOut: core.GetPracticeOutput{Practice: practiceDetailFixture()}}
		response := authedPracticeRequestWithCourse(t, draftCourse, service, http.MethodGet, "/v1/practices/"+practiceIDValue+"?view=instructor", "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response for instructor practice, got %d body=%s", response.Code, response.Body.String())
		}
		assertJSONContainsKeys(t, response.Body.Bytes(), "Solution", "TestCases")
	})

	t.Run("test", func(t *testing.T) {
		service := &testServiceFake{getOut: core.GetTestOutput{Test: testDetailWithChoiceAndCodingFixture()}}
		response := authedTestRequestWithCourse(t, draftCourse, service, http.MethodGet, "/v1/tests/"+testIDValue, "")

		if response.Code != http.StatusOK {
			t.Fatalf("expected ok response for instructor test, got %d body=%s", response.Code, response.Body.String())
		}
		assertJSONContainsKeys(t, response.Body.Bytes(), "Solution", "ChoiceCorrectIndices", "CodingSolution", "TestCases")
	})
}

func TestReorderTestItemsParsesOrder(t *testing.T) {
	service := &testServiceFake{}
	response := authedTestRequest(t, service, http.MethodPost, "/v1/tests/"+testIDValue+"/items/reorder", `{
		"order": [
			{"testItemID": "`+testItemIDValue+`", "position": 1},
			{"testItemID": "`+otherTestItemID+`", "position": 0}
		]
	}`)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "reorder-items" || service.reorderItemsIn.TestID != testIDValue {
		t.Fatalf("expected reorder test items input, got called=%q input=%+v", service.called, service.reorderItemsIn)
	}
	if len(service.reorderItemsIn.Order) != 2 || service.reorderItemsIn.Order[0].TestItemID != testItemIDValue || service.reorderItemsIn.Order[1].Position != 0 {
		t.Fatalf("expected test item reorder placements, got %+v", service.reorderItemsIn.Order)
	}
}

func TestTestErrorsMapToHTTPStatus(t *testing.T) {
	t.Run("validation", func(t *testing.T) {
		service := &testServiceFake{err: domain.NewValidationError("title", "is required")}
		response := authedTestRequest(t, service, http.MethodPost, "/v1/tests", `{
			"courseID": "`+courseIDValue+`",
			"title": ""
		}`)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected bad request response, got %d body=%s", response.Code, response.Body.String())
		}
	})

	t.Run("not found", func(t *testing.T) {
		service := &testServiceFake{err: domain.ErrNotFound}
		response := authedTestRequest(t, service, http.MethodGet, "/v1/tests/"+testIDValue, "")

		if response.Code != http.StatusNotFound {
			t.Fatalf("expected not found response, got %d body=%s", response.Code, response.Body.String())
		}
	})
}

func TestPracticeInUseMapsToConflictWithLessonIDs(t *testing.T) {
	service := &practiceServiceFake{err: practiceInUseError(t)}
	response := authedPracticeRequest(t, service, http.MethodDelete, "/v1/practices/"+practiceIDValue, "")

	if response.Code != http.StatusConflict {
		t.Fatalf("expected conflict response, got %d body=%s", response.Code, response.Body.String())
	}

	var out practiceInUseResponse
	decodeResponse(t, response, &out)
	if out.Error == "" || len(out.LessonIDs) != 2 || out.LessonIDs[0] != lessonIDValue || out.LessonIDs[1] != otherLessonIDValue {
		t.Fatalf("expected practice in use response with lesson ids, got %+v", out)
	}
}

func TestQuizInUseMapsToConflictWithLessonIDs(t *testing.T) {
	service := &quizServiceFake{err: quizInUseError(t)}
	response := authedQuizRequest(t, service, http.MethodDelete, "/v1/quizzes/"+quizIDValue, "")

	if response.Code != http.StatusConflict {
		t.Fatalf("expected conflict response, got %d body=%s", response.Code, response.Body.String())
	}

	var out quizInUseResponse
	decodeResponse(t, response, &out)
	if out.Error == "" || len(out.LessonIDs) != 2 || out.LessonIDs[0] != lessonIDValue || out.LessonIDs[1] != otherLessonIDValue {
		t.Fatalf("expected quiz in use response with lesson ids, got %+v", out)
	}
}

func TestServerMapsErrorsToHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "validation", err: domain.NewValidationError("kind", "must be text or video"), want: http.StatusBadRequest},
		{name: "not found", err: domain.ErrNotFound, want: http.StatusNotFound},
		{name: "internal", err: errors.New("database unavailable"), want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &lessonServiceFake{err: test.err}
			response := authedRequest(t, service, http.MethodGet, "/v1/blocks/"+blockIDValue, "")

			if response.Code != test.want {
				t.Fatalf("expected status %d, got %d body=%s", test.want, response.Code, response.Body.String())
			}
		})
	}
}

func TestServerRejectsMalformedJSON(t *testing.T) {
	service := &lessonServiceFake{}
	response := authedRequest(t, service, http.MethodPost, "/v1/lessons/"+lessonIDValue+"/blocks", `{`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "" {
		t.Fatalf("expected malformed json not to reach service, got %q", service.called)
	}
}

func newTestServer(t *testing.T, lesson core.LessonService) *Server {
	t.Helper()

	server, err := NewServer(Options{
		Course:   courseServiceFake{},
		Lesson:   lesson,
		Quiz:     &quizServiceFake{},
		Practice: &practiceServiceFake{},
		Test:     &testServiceFake{},
		Token:    apiTokenValue,
	})
	if err != nil {
		t.Fatalf("expected test server, got %v", err)
	}

	return server
}

func authedRequest(t *testing.T, service *lessonServiceFake, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	server := newTestServer(t, service)
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+apiTokenValue)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	return response
}

func authedQuizRequest(t *testing.T, service *quizServiceFake, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	return authedQuizRequestWithCourse(t, courseServiceFake{}, service, method, path, body)
}

func authedQuizRequestWithCourse(t *testing.T, course core.CourseService, service *quizServiceFake, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	server, err := NewServer(Options{
		Course:   course,
		Lesson:   &lessonServiceFake{},
		Quiz:     service,
		Practice: &practiceServiceFake{},
		Test:     &testServiceFake{},
		Token:    apiTokenValue,
	})
	if err != nil {
		t.Fatalf("expected test server, got %v", err)
	}

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+apiTokenValue)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	return response
}

func authedPracticeRequest(t *testing.T, service *practiceServiceFake, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	return authedPracticeRequestWithCourse(t, courseServiceFake{}, service, method, path, body)
}

func authedPracticeRequestWithCourse(t *testing.T, course core.CourseService, service *practiceServiceFake, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	server, err := NewServer(Options{
		Course:   course,
		Lesson:   &lessonServiceFake{},
		Quiz:     &quizServiceFake{},
		Practice: service,
		Test:     &testServiceFake{},
		Token:    apiTokenValue,
	})
	if err != nil {
		t.Fatalf("expected test server, got %v", err)
	}

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+apiTokenValue)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	return response
}

func authedTestRequest(t *testing.T, service *testServiceFake, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	return authedTestRequestWithCourse(t, courseServiceFake{}, service, method, path, body)
}

func authedTestRequestWithCourse(t *testing.T, course core.CourseService, service *testServiceFake, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	server, err := NewServer(Options{
		Course:   course,
		Lesson:   &lessonServiceFake{},
		Quiz:     &quizServiceFake{},
		Practice: &practiceServiceFake{},
		Test:     service,
		Token:    apiTokenValue,
	})
	if err != nil {
		t.Fatalf("expected test server, got %v", err)
	}

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+apiTokenValue)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	return response
}

func decodeResponse(t *testing.T, response *httptest.ResponseRecorder, value any) {
	t.Helper()

	if err := json.NewDecoder(response.Body).Decode(value); err != nil {
		t.Fatalf("expected json response, got %v", err)
	}
}

func decodeResponseBytes(t *testing.T, body []byte, value any) {
	t.Helper()

	if err := json.Unmarshal(body, value); err != nil {
		t.Fatalf("expected json response, got %v body=%s", err, string(body))
	}
}

func assertJSONOmitsKeys(t *testing.T, body []byte, keys ...string) {
	t.Helper()

	for _, key := range keys {
		if jsonHasKey(t, body, key) {
			t.Fatalf("expected JSON to omit key %q, got body=%s", key, string(body))
		}
	}
}

func assertJSONContainsKeys(t *testing.T, body []byte, keys ...string) {
	t.Helper()

	for _, key := range keys {
		if !jsonHasKey(t, body, key) {
			t.Fatalf("expected JSON to contain key %q, got body=%s", key, string(body))
		}
	}
}

func jsonHasKey(t *testing.T, body []byte, key string) bool {
	t.Helper()

	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatalf("expected json response, got %v body=%s", err, string(body))
	}

	return jsonValueHasKey(value, key)
}

func jsonValueHasKey(value any, key string) bool {
	switch typed := value.(type) {
	case map[string]any:
		if _, ok := typed[key]; ok {
			return true
		}
		for _, child := range typed {
			if jsonValueHasKey(child, key) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if jsonValueHasKey(child, key) {
				return true
			}
		}
	}

	return false
}

func blockViewFixture() core.BlockView {
	return core.BlockView{
		ID:            blockIDValue,
		LessonID:      lessonIDValue,
		Kind:          "video",
		Position:      1,
		VideoProvider: "youtube",
		VideoLocator:  "dQw4w9WgXcQ",
		VideoCaption:  "Intro video",
	}
}

func courseViewWithStatus(status string) core.CourseView {
	course := courseViewFixture()
	course.Status = status
	return course
}

func quizViewFixture() core.QuizView {
	return core.QuizView{
		ID:            quizIDValue,
		CourseID:      courseIDValue,
		Title:         "Basics Quiz",
		PassThreshold: 0.7,
		QuestionCount: 1,
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

func practiceViewFixture() core.PracticeView {
	return core.PracticeView{
		ID:            practiceIDValue,
		CourseID:      courseIDValue,
		Title:         "FizzBuzz",
		Language:      "golang",
		TestCaseCount: 1,
		HasSolution:   true,
	}
}

func practiceDetailFixture() core.PracticeDetailView {
	return core.PracticeDetailView{
		PracticeView: practiceViewFixture(),
		Prompt:       "Print fizz buzz",
		StarterCode:  "package main",
		Solution:     "fmt.Println()",
		TestCases:    []core.TestCaseView{testCaseViewFixture()},
	}
}

func testCaseViewFixture() core.TestCaseView {
	return core.TestCaseView{
		ID:             testCaseIDValue,
		PracticeID:     practiceIDValue,
		Stdin:          "1\n",
		ExpectedStdout: "1\n",
		Name:           "identity",
		Position:       0,
	}
}

func testViewFixture() core.TestView {
	timeLimit := 45
	return core.TestView{
		ID:               testIDValue,
		CourseID:         courseIDValue,
		Title:            "Final Test",
		TimeLimitMinutes: &timeLimit,
		PassThreshold:    0.7,
		HasSolution:      true,
		ItemCount:        1,
	}
}

func testDetailFixture() core.TestDetailView {
	return core.TestDetailView{
		TestView: testViewFixture(),
		Solution: &core.TestSolutionView{
			ZipProvider:   "url",
			ZipLocator:    "https://example.com/solution.zip",
			VideoProvider: "url",
			VideoLocator:  "https://example.com/video.mp4",
			VideoCaption:  "Walkthrough",
		},
		Items: []core.TestItemView{testItemViewFixture()},
	}
}

func testDetailWithChoiceAndCodingFixture() core.TestDetailView {
	detail := testDetailFixture()
	detail.ItemCount = 2
	detail.Items = []core.TestItemView{
		testItemViewFixture(),
		{
			ID:             otherTestItemID,
			TestID:         testIDValue,
			Kind:           "coding",
			Position:       1,
			CodingPrompt:   "Write code",
			Language:       "golang",
			StarterCode:    "package main",
			CodingSolution: "func main() {}",
			TestCases: []core.CodingTestCaseDTO{
				{
					Stdin:          "1\n",
					ExpectedStdout: "1\n",
					Name:           "hidden",
				},
			},
		},
	}

	return detail
}

func testItemViewFixture() core.TestItemView {
	return core.TestItemView{
		ID:                   testItemIDValue,
		TestID:               testIDValue,
		Kind:                 "choice",
		Position:             0,
		ChoicePrompt:         "Pick one",
		ChoiceType:           "single",
		ChoiceOptions:        []string{"A", "B"},
		ChoiceCorrectIndices: []int{0},
		ChoiceExplanation:    "Because A",
	}
}

func practiceInUseError(t *testing.T) error {
	t.Helper()

	lessonID, err := domain.NewLessonID(lessonIDValue)
	if err != nil {
		t.Fatalf("expected lesson id, got %v", err)
	}
	otherLessonID, err := domain.NewLessonID(otherLessonIDValue)
	if err != nil {
		t.Fatalf("expected lesson id, got %v", err)
	}

	return domain.NewPracticeInUseError([]domain.LessonID{lessonID, otherLessonID})
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

type lessonServiceFake struct {
	called string
	err    error

	addIn     core.AddLessonBlockInput
	addOut    core.AddLessonBlockOutput
	listIn    core.ListLessonBlocksInput
	listOut   core.ListLessonBlocksOutput
	getIn     core.GetLessonBlockInput
	getOut    core.GetLessonBlockOutput
	updateIn  core.UpdateLessonBlockInput
	updateOut core.UpdateLessonBlockOutput
	removeIn  core.RemoveLessonBlockInput
	reorderIn core.ReorderLessonBlocksInput
}

func (service *lessonServiceFake) CreateLesson(core.CreateLessonInput) (core.CreateLessonOutput, error) {
	return core.CreateLessonOutput{}, nil
}

func (service *lessonServiceFake) ListLessons(core.ListLessonsInput) (core.ListLessonsOutput, error) {
	return core.ListLessonsOutput{}, nil
}

func (service *lessonServiceFake) GetLesson(core.GetLessonInput) (core.GetLessonOutput, error) {
	return core.GetLessonOutput{}, nil
}

func (service *lessonServiceFake) UpdateLesson(core.UpdateLessonInput) (core.UpdateLessonOutput, error) {
	return core.UpdateLessonOutput{}, nil
}

func (service *lessonServiceFake) DeleteLesson(core.DeleteLessonInput) error {
	return nil
}

func (service *lessonServiceFake) ReorderLessons(core.ReorderLessonsInput) error {
	return nil
}

func (service *lessonServiceFake) AddLessonBlock(in core.AddLessonBlockInput) (core.AddLessonBlockOutput, error) {
	service.called = "add"
	service.addIn = in
	if service.err != nil {
		return core.AddLessonBlockOutput{}, service.err
	}

	return service.addOut, nil
}

func (service *lessonServiceFake) ListLessonBlocks(in core.ListLessonBlocksInput) (core.ListLessonBlocksOutput, error) {
	service.called = "list"
	service.listIn = in
	if service.err != nil {
		return core.ListLessonBlocksOutput{}, service.err
	}

	return service.listOut, nil
}

func (service *lessonServiceFake) GetLessonBlock(in core.GetLessonBlockInput) (core.GetLessonBlockOutput, error) {
	service.called = "get"
	service.getIn = in
	if service.err != nil {
		return core.GetLessonBlockOutput{}, service.err
	}

	return service.getOut, nil
}

func (service *lessonServiceFake) UpdateLessonBlock(in core.UpdateLessonBlockInput) (core.UpdateLessonBlockOutput, error) {
	service.called = "update"
	service.updateIn = in
	if service.err != nil {
		return core.UpdateLessonBlockOutput{}, service.err
	}

	return service.updateOut, nil
}

func (service *lessonServiceFake) RemoveLessonBlock(in core.RemoveLessonBlockInput) error {
	service.called = "remove"
	service.removeIn = in
	return service.err
}

func (service *lessonServiceFake) ReorderLessonBlocks(in core.ReorderLessonBlocksInput) error {
	service.called = "reorder"
	service.reorderIn = in
	return service.err
}

type courseServiceFake struct {
	err    error
	getOut core.GetCourseOutput
}

func (courseServiceFake) CreateCourse(core.CreateCourseInput) (core.CreateCourseOutput, error) {
	return core.CreateCourseOutput{}, nil
}

func (courseServiceFake) ListCourses(core.ListCoursesInput) (core.ListCoursesOutput, error) {
	return core.ListCoursesOutput{}, nil
}

func (service courseServiceFake) GetCourse(core.GetCourseInput) (core.GetCourseOutput, error) {
	if service.err != nil {
		return core.GetCourseOutput{}, service.err
	}
	if service.getOut.Course.ID != "" {
		return service.getOut, nil
	}

	return core.GetCourseOutput{Course: courseViewFixture()}, nil
}

func (courseServiceFake) UpdateCourse(core.UpdateCourseInput) (core.UpdateCourseOutput, error) {
	return core.UpdateCourseOutput{}, nil
}

func (courseServiceFake) DeleteCourse(core.DeleteCourseInput) error {
	return nil
}

func (courseServiceFake) PublishCourse(core.PublishCourseInput) error {
	return nil
}

func (courseServiceFake) UnpublishCourse(core.UnpublishCourseInput) error {
	return nil
}

type quizServiceFake struct {
	called string
	err    error

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
	service.called = "create"
	service.createIn = in
	if service.err != nil {
		return core.CreateQuizOutput{}, service.err
	}

	return service.createOut, nil
}

func (service *quizServiceFake) ListQuizzes(in core.ListQuizzesInput) (core.ListQuizzesOutput, error) {
	service.called = "list"
	service.listIn = in
	if service.err != nil {
		return core.ListQuizzesOutput{}, service.err
	}

	return service.listOut, nil
}

func (service *quizServiceFake) GetQuiz(in core.GetQuizInput) (core.GetQuizOutput, error) {
	service.called = "get"
	service.getIn = in
	if service.err != nil {
		return core.GetQuizOutput{}, service.err
	}

	return service.getOut, nil
}

func (service *quizServiceFake) UpdateQuiz(in core.UpdateQuizInput) (core.UpdateQuizOutput, error) {
	service.called = "update"
	service.updateIn = in
	if service.err != nil {
		return core.UpdateQuizOutput{}, service.err
	}

	return service.updateOut, nil
}

func (service *quizServiceFake) DeleteQuiz(in core.DeleteQuizInput) error {
	service.called = "delete"
	service.deleteIn = in
	return service.err
}

func (service *quizServiceFake) AddQuestion(in core.AddQuestionInput) (core.AddQuestionOutput, error) {
	service.called = "add-question"
	service.addQuestionIn = in
	if service.err != nil {
		return core.AddQuestionOutput{}, service.err
	}

	return service.addQuestionOut, nil
}

func (service *quizServiceFake) ListQuestions(in core.ListQuestionsInput) (core.ListQuestionsOutput, error) {
	service.called = "list-questions"
	service.listQuestionsIn = in
	if service.err != nil {
		return core.ListQuestionsOutput{}, service.err
	}

	return service.listQuestionsOut, nil
}

func (service *quizServiceFake) GetQuestion(in core.GetQuestionInput) (core.GetQuestionOutput, error) {
	service.called = "get-question"
	service.getQuestionIn = in
	if service.err != nil {
		return core.GetQuestionOutput{}, service.err
	}

	return service.getQuestionOut, nil
}

func (service *quizServiceFake) UpdateQuestion(in core.UpdateQuestionInput) (core.UpdateQuestionOutput, error) {
	service.called = "update-question"
	service.updateQuestionIn = in
	if service.err != nil {
		return core.UpdateQuestionOutput{}, service.err
	}

	return service.updateQuestionOut, nil
}

func (service *quizServiceFake) RemoveQuestion(in core.RemoveQuestionInput) error {
	service.called = "remove-question"
	service.removeQuestionIn = in
	return service.err
}

func (service *quizServiceFake) ReorderQuestions(in core.ReorderQuestionsInput) error {
	service.called = "reorder-questions"
	service.reorderQuestionsIn = in
	return service.err
}

type practiceServiceFake struct {
	called string
	err    error

	createIn           core.CreatePracticeInput
	createOut          core.CreatePracticeOutput
	listIn             core.ListPracticesInput
	listOut            core.ListPracticesOutput
	getIn              core.GetPracticeInput
	getOut             core.GetPracticeOutput
	updateIn           core.UpdatePracticeInput
	updateOut          core.UpdatePracticeOutput
	deleteIn           core.DeletePracticeInput
	addTestCaseIn      core.AddTestCaseInput
	addTestCaseOut     core.AddTestCaseOutput
	listTestCasesIn    core.ListTestCasesInput
	listTestCasesOut   core.ListTestCasesOutput
	getTestCaseIn      core.GetTestCaseInput
	getTestCaseOut     core.GetTestCaseOutput
	updateTestCaseIn   core.UpdateTestCaseInput
	updateTestCaseOut  core.UpdateTestCaseOutput
	removeTestCaseIn   core.RemoveTestCaseInput
	reorderTestCasesIn core.ReorderTestCasesInput
}

func (service *practiceServiceFake) CreatePractice(in core.CreatePracticeInput) (core.CreatePracticeOutput, error) {
	service.called = "create"
	service.createIn = in
	if service.err != nil {
		return core.CreatePracticeOutput{}, service.err
	}

	return service.createOut, nil
}

func (service *practiceServiceFake) ListPractices(in core.ListPracticesInput) (core.ListPracticesOutput, error) {
	service.called = "list"
	service.listIn = in
	if service.err != nil {
		return core.ListPracticesOutput{}, service.err
	}

	return service.listOut, nil
}

func (service *practiceServiceFake) GetPractice(in core.GetPracticeInput) (core.GetPracticeOutput, error) {
	service.called = "get"
	service.getIn = in
	if service.err != nil {
		return core.GetPracticeOutput{}, service.err
	}

	return service.getOut, nil
}

func (service *practiceServiceFake) UpdatePractice(in core.UpdatePracticeInput) (core.UpdatePracticeOutput, error) {
	service.called = "update"
	service.updateIn = in
	if service.err != nil {
		return core.UpdatePracticeOutput{}, service.err
	}

	return service.updateOut, nil
}

func (service *practiceServiceFake) DeletePractice(in core.DeletePracticeInput) error {
	service.called = "delete"
	service.deleteIn = in
	return service.err
}

func (service *practiceServiceFake) AddTestCase(in core.AddTestCaseInput) (core.AddTestCaseOutput, error) {
	service.called = "add-testcase"
	service.addTestCaseIn = in
	if service.err != nil {
		return core.AddTestCaseOutput{}, service.err
	}

	return service.addTestCaseOut, nil
}

func (service *practiceServiceFake) ListTestCases(in core.ListTestCasesInput) (core.ListTestCasesOutput, error) {
	service.called = "list-testcases"
	service.listTestCasesIn = in
	if service.err != nil {
		return core.ListTestCasesOutput{}, service.err
	}

	return service.listTestCasesOut, nil
}

func (service *practiceServiceFake) GetTestCase(in core.GetTestCaseInput) (core.GetTestCaseOutput, error) {
	service.called = "get-testcase"
	service.getTestCaseIn = in
	if service.err != nil {
		return core.GetTestCaseOutput{}, service.err
	}

	return service.getTestCaseOut, nil
}

func (service *practiceServiceFake) UpdateTestCase(in core.UpdateTestCaseInput) (core.UpdateTestCaseOutput, error) {
	service.called = "update-testcase"
	service.updateTestCaseIn = in
	if service.err != nil {
		return core.UpdateTestCaseOutput{}, service.err
	}

	return service.updateTestCaseOut, nil
}

func (service *practiceServiceFake) RemoveTestCase(in core.RemoveTestCaseInput) error {
	service.called = "remove-testcase"
	service.removeTestCaseIn = in
	return service.err
}

func (service *practiceServiceFake) ReorderTestCases(in core.ReorderTestCasesInput) error {
	service.called = "reorder-testcases"
	service.reorderTestCasesIn = in
	return service.err
}

type testServiceFake struct {
	called string
	err    error

	createIn       core.CreateTestInput
	createOut      core.CreateTestOutput
	listIn         core.ListTestsInput
	listOut        core.ListTestsOutput
	getIn          core.GetTestInput
	getOut         core.GetTestOutput
	updateIn       core.UpdateTestInput
	updateOut      core.UpdateTestOutput
	deleteIn       core.DeleteTestInput
	addItemIn      core.AddTestItemInput
	addItemOut     core.AddTestItemOutput
	listItemsIn    core.ListTestItemsInput
	listItemsOut   core.ListTestItemsOutput
	getItemIn      core.GetTestItemInput
	getItemOut     core.GetTestItemOutput
	updateItemIn   core.UpdateTestItemInput
	updateItemOut  core.UpdateTestItemOutput
	removeItemIn   core.RemoveTestItemInput
	reorderItemsIn core.ReorderTestItemsInput
}

func (service *testServiceFake) CreateTest(in core.CreateTestInput) (core.CreateTestOutput, error) {
	service.called = "create"
	service.createIn = in
	if service.err != nil {
		return core.CreateTestOutput{}, service.err
	}

	return service.createOut, nil
}

func (service *testServiceFake) ListTests(in core.ListTestsInput) (core.ListTestsOutput, error) {
	service.called = "list"
	service.listIn = in
	if service.err != nil {
		return core.ListTestsOutput{}, service.err
	}

	return service.listOut, nil
}

func (service *testServiceFake) GetTest(in core.GetTestInput) (core.GetTestOutput, error) {
	service.called = "get"
	service.getIn = in
	if service.err != nil {
		return core.GetTestOutput{}, service.err
	}

	return service.getOut, nil
}

func (service *testServiceFake) UpdateTest(in core.UpdateTestInput) (core.UpdateTestOutput, error) {
	service.called = "update"
	service.updateIn = in
	if service.err != nil {
		return core.UpdateTestOutput{}, service.err
	}

	return service.updateOut, nil
}

func (service *testServiceFake) DeleteTest(in core.DeleteTestInput) error {
	service.called = "delete"
	service.deleteIn = in
	return service.err
}

func (service *testServiceFake) AddTestItem(in core.AddTestItemInput) (core.AddTestItemOutput, error) {
	service.called = "add-item"
	service.addItemIn = in
	if service.err != nil {
		return core.AddTestItemOutput{}, service.err
	}

	return service.addItemOut, nil
}

func (service *testServiceFake) ListTestItems(in core.ListTestItemsInput) (core.ListTestItemsOutput, error) {
	service.called = "list-items"
	service.listItemsIn = in
	if service.err != nil {
		return core.ListTestItemsOutput{}, service.err
	}

	return service.listItemsOut, nil
}

func (service *testServiceFake) GetTestItem(in core.GetTestItemInput) (core.GetTestItemOutput, error) {
	service.called = "get-item"
	service.getItemIn = in
	if service.err != nil {
		return core.GetTestItemOutput{}, service.err
	}

	return service.getItemOut, nil
}

func (service *testServiceFake) UpdateTestItem(in core.UpdateTestItemInput) (core.UpdateTestItemOutput, error) {
	service.called = "update-item"
	service.updateItemIn = in
	if service.err != nil {
		return core.UpdateTestItemOutput{}, service.err
	}

	return service.updateItemOut, nil
}

func (service *testServiceFake) RemoveTestItem(in core.RemoveTestItemInput) error {
	service.called = "remove-item"
	service.removeItemIn = in
	return service.err
}

func (service *testServiceFake) ReorderTestItems(in core.ReorderTestItemsInput) error {
	service.called = "reorder-items"
	service.reorderItemsIn = in
	return service.err
}
