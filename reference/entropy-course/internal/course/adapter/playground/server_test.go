package playground

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	courseIDValue        = "550e8400-e29b-41d4-a716-446655440000"
	instructorIDValue    = "550e8400-e29b-41d4-a716-446655440010"
	lessonIDValue        = "550e8400-e29b-41d4-a716-446655440020"
	blockIDValue         = "550e8400-e29b-41d4-a716-446655440030"
	otherBlockIDValue    = "550e8400-e29b-41d4-a716-446655440031"
	quizIDValue          = "550e8400-e29b-41d4-a716-446655440040"
	questionIDValue      = "550e8400-e29b-41d4-a716-446655440050"
	otherQuestionIDValue = "550e8400-e29b-41d4-a716-446655440051"
	practiceIDValue      = "550e8400-e29b-41d4-a716-446655440060"
	testCaseIDValue      = "550e8400-e29b-41d4-a716-446655440070"
	otherTestCaseIDValue = "550e8400-e29b-41d4-a716-446655440071"
	testIDValue          = "550e8400-e29b-41d4-a716-446655440080"
	testItemIDValue      = "550e8400-e29b-41d4-a716-446655440090"
	otherTestItemIDValue = "550e8400-e29b-41d4-a716-446655440091"
)

func TestCatalogContainsCourseLessonQuizPracticeAndBlockCommands(t *testing.T) {
	commands := Catalog()
	ids := make(map[string]bool)
	for _, command := range commands {
		ids[command.ID] = true
	}

	want := []string{
		"course-create",
		"course-list",
		"course-get",
		"course-update",
		"course-delete",
		"course-publish",
		"course-unpublish",
		"import-plan",
		"import-apply",
		"quiz-create",
		"quiz-list",
		"quiz-get",
		"quiz-update",
		"quiz-delete",
		"quiz-question-add",
		"quiz-question-list",
		"quiz-question-get",
		"quiz-question-update",
		"quiz-question-remove",
		"quiz-question-reorder",
		"practice-create",
		"practice-list",
		"practice-get",
		"practice-update",
		"practice-delete",
		"practice-testcase-add",
		"practice-testcase-list",
		"practice-testcase-get",
		"practice-testcase-update",
		"practice-testcase-remove",
		"practice-testcase-reorder",
		"test-create",
		"test-list",
		"test-get",
		"test-update",
		"test-delete",
		"test-item-add",
		"test-item-list",
		"test-item-get",
		"test-item-update",
		"test-item-remove",
		"test-item-reorder",
		"lesson-create",
		"lesson-list",
		"lesson-get",
		"lesson-update",
		"lesson-delete",
		"lesson-reorder",
		"lesson-block-add",
		"lesson-block-list",
		"lesson-block-get",
		"lesson-block-update",
		"lesson-block-remove",
		"lesson-block-reorder",
	}
	for _, id := range want {
		if !ids[id] {
			t.Fatalf("expected catalog to include %s", id)
		}
	}
}

func TestCatalogMatchesRunnableCobraCommands(t *testing.T) {
	runner := NewRunner(Services{Config: viper.New()})
	actualCommands := runnableCobraCommands(runner.rootCommand())
	catalogCommands := Catalog()
	catalogByPath := make(map[string]Command, len(catalogCommands))

	for _, command := range catalogCommands {
		if command.ID != strings.ReplaceAll(command.Command, " ", "-") {
			t.Fatalf("expected %s id to match command path %q", command.ID, command.Command)
		}
		catalogByPath[command.Command] = command

		cobraCommand, ok := actualCommands[command.Command]
		if !ok {
			t.Fatalf("expected catalog command %q to exist in Cobra command tree", command.Command)
		}
		assertCatalogFieldsMatchCobra(t, command, cobraCommand)
	}

	for path := range actualCommands {
		if _, ok := catalogByPath[path]; !ok {
			t.Fatalf("expected Cobra command %q to be represented in playground catalog", path)
		}
	}
}

func TestCatalogSupportsPickersRepeatableFieldsAndEmptyPracticeFields(t *testing.T) {
	commands := Catalog()

	blockAdd := commandByID(t, commands, "lesson-block-add")
	kind := fieldByKey(t, blockAdd, "kind")
	if !slices.Contains(kind.Options, "quiz") {
		t.Fatalf("expected lesson block kind options to include quiz, got %#v", kind.Options)
	}
	if !slices.Contains(kind.Options, "practice") {
		t.Fatalf("expected lesson block kind options to include practice, got %#v", kind.Options)
	}
	quizID := fieldByKey(t, blockAdd, "quiz-id")
	if quizID.RequiredWhen != "kind=quiz" || quizID.Binding.Flag != "--quiz-id" {
		t.Fatalf("expected quiz picker field for lesson blocks, got %+v", quizID)
	}
	practiceID := fieldByKey(t, blockAdd, "practice-id")
	if practiceID.RequiredWhen != "kind=practice" || practiceID.Binding.Flag != "--practice-id" {
		t.Fatalf("expected practice picker field for lesson blocks, got %+v", practiceID)
	}
	text := fieldByKey(t, blockAdd, "text")
	if text.RequiredWhen != "kind=text" || text.Binding.Flag != "--text" {
		t.Fatalf("expected text field to be conditional for lesson blocks, got %+v", text)
	}
	videoProvider := fieldByKey(t, blockAdd, "video-provider")
	if videoProvider.RequiredWhen != "kind=video" || videoProvider.Binding.Flag != "--video-provider" {
		t.Fatalf("expected video provider field to be conditional for lesson blocks, got %+v", videoProvider)
	}
	videoCaption := fieldByKey(t, blockAdd, "video-caption")
	if videoCaption.VisibleWhen != "kind=video" || videoCaption.RequiredWhen != "" || videoCaption.Binding.Flag != "--video-caption" {
		t.Fatalf("expected video caption field to be optional and visible only for video blocks, got %+v", videoCaption)
	}

	questionAdd := commandByID(t, commands, "quiz-question-add")
	options := fieldByKey(t, questionAdd, "option")
	if !options.Binding.MultiValue || options.Binding.Separator != "\n" {
		t.Fatalf("expected option field to split one option per line, got %+v", options.Binding)
	}
	correct := fieldByKey(t, questionAdd, "correct")
	if !correct.Binding.MultiValue || correct.Binding.Separator != "," {
		t.Fatalf("expected correct field to split comma-separated indices, got %+v", correct.Binding)
	}

	testCaseUpdate := commandByID(t, commands, "practice-testcase-update")
	for _, key := range []string{"stdin", "expected-stdout", "name"} {
		field := fieldByKey(t, testCaseUpdate, key)
		if !field.Binding.AllowEmpty {
			t.Fatalf("expected practice test case %s field to allow deliberate empty values", key)
		}
	}

	blockList := commandByID(t, commands, "lesson-block-list")
	if !strings.Contains(blockList.Description, "quiz") || !strings.Contains(blockList.Description, "practice") {
		t.Fatalf("expected lesson block list docs to mention quiz and practice blocks, got %q", blockList.Description)
	}

	testItemAdd := commandByID(t, commands, "test-item-add")
	kind = fieldByKey(t, testItemAdd, "kind")
	if !slices.Contains(kind.Options, "choice") || !slices.Contains(kind.Options, "coding") {
		t.Fatalf("expected test item kind options to include choice and coding, got %#v", kind.Options)
	}
	choiceOptions := fieldByKey(t, testItemAdd, "option")
	if choiceOptions.RequiredWhen != "kind=choice" || !choiceOptions.Binding.MultiValue || choiceOptions.Binding.Separator != "\n" {
		t.Fatalf("expected choice options to be one per line and conditionally required, got %+v", choiceOptions)
	}
	codingCases := fieldByKey(t, testItemAdd, "testcase")
	if codingCases.RequiredWhen != "kind=coding" || !codingCases.Binding.MultiValue || codingCases.Binding.Separator != "\n" {
		t.Fatalf("expected coding test cases to be one per line and conditionally required, got %+v", codingCases)
	}
	for _, key := range []string{"explanation", "starter-code", "solution"} {
		field := fieldByKey(t, commandByID(t, commands, "test-item-update"), key)
		if !field.Binding.AllowEmpty {
			t.Fatalf("expected test item update %s field to allow deliberate empty values", key)
		}
	}
	testItemUpdate := commandByID(t, commands, "test-item-update")
	payloadKind := fieldByKey(t, testItemUpdate, "payload-kind")
	if payloadKind.Binding.Flag != "" || payloadKind.Binding.Argument || !slices.Contains(payloadKind.Options, "choice") || !slices.Contains(payloadKind.Options, "coding") {
		t.Fatalf("expected payload kind to be a playground-only choice/coding selector, got %+v", payloadKind)
	}
	if fieldByKey(t, testItemUpdate, "type").VisibleWhen != "payload-kind=choice" {
		t.Fatalf("expected choice fields to be scoped to choice payloads")
	}
	if fieldByKey(t, testItemUpdate, "language").VisibleWhen != "payload-kind=coding" {
		t.Fatalf("expected coding fields to be scoped to coding payloads")
	}
}

func TestIndexServesCommandInventory(t *testing.T) {
	server := newTestServer(t, &courseServiceFake{}, &lessonServiceFake{})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok index, got %d", response.Code)
	}
	body := response.Body.String()
	for _, text := range []string{"Course CLI Playground", "REST API runs separately on 127.0.0.1:8788", "course-create", "importConsoleButton", "quiz-create", "practice-create", "practice-testcase-add", "test-create", "test-item-add", "lesson-block-reorder"} {
		if !strings.Contains(body, text) {
			t.Fatalf("expected index to contain %q", text)
		}
	}
}

func TestIndexDocumentsResolvedImportWorkflow(t *testing.T) {
	server := newTestServer(t, &courseServiceFake{}, &lessonServiceFake{})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok index, got %d", response.Code)
	}
	body := response.Body.String()
	for _, text := range []string{
		"Normalize source material into a v1 course zip",
		"resolve conflicts by choosing update or skip against existing candidates",
		"change the zip identity and replan",
		"This playground runs bounded-context CLI actions only",
		"Start the REST API in another terminal with course-cli rest",
		"data-conflict-field=\"target_id\"",
		"[\"update\", \"skip\"]",
		"POST /import/plan",
	} {
		if !strings.Contains(body, text) {
			t.Fatalf("expected index to contain %q", text)
		}
	}
	if strings.Contains(body, "[\"update\", \"skip\", \"create\"]") {
		t.Fatalf("expected import console to avoid create conflict resolutions")
	}
}

func TestImportPlanEndpointMapsUploadToService(t *testing.T) {
	imports := &importServiceFake{
		planOut: core.PlanImportOutput{Plan: importPlanFixture(t)},
	}
	server := newTestServerWithImport(t, imports)

	response := postImportMultipart(t, server, "/import/plan", importMultipartRequest{zip: []byte("zip-bytes")})

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if imports.called != "plan" {
		t.Fatalf("expected plan call, got %q", imports.called)
	}
	if imports.planIn.InstructorID != instructorIDValue {
		t.Fatalf("expected instructor id, got %+v", imports.planIn)
	}
	if imports.planZipContent != "zip-bytes" {
		t.Fatalf("expected uploaded zip bytes during service call, got %q", imports.planZipContent)
	}
	if _, err := os.Stat(imports.planIn.ZipPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected temp zip to be deleted after request, got %v", err)
	}

	var out map[string]any
	if err := json.NewDecoder(response.Body).Decode(&out); err != nil {
		t.Fatalf("expected import plan json, got %v", err)
	}
	if out["format_version"] != "1" || out["zip_hash"] == "" {
		t.Fatalf("expected import plan response, got %+v", out)
	}
}

func TestImportApplyEndpointMapsUploadResolvedPlanAndStrategy(t *testing.T) {
	imports := &importServiceFake{
		applyOut: core.ApplyPlanOutput{Result: applyResultFixture(t, nil)},
	}
	server := newTestServerWithImport(t, imports)

	response := postImportMultipart(t, server, "/import/apply?conflict_strategy=update", importMultipartRequest{
		zip:              []byte("zip-bytes"),
		resolvedPlanJSON: []byte(`{"format_version":"1"}`),
	})

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if imports.called != "apply" {
		t.Fatalf("expected apply call, got %q", imports.called)
	}
	if imports.applyIn.InstructorID != instructorIDValue || imports.applyIn.ConflictStrategy != "update" {
		t.Fatalf("expected instructor and strategy, got %+v", imports.applyIn)
	}
	if string(imports.applyIn.ResolvedPlanJSON) != `{"format_version":"1"}` {
		t.Fatalf("expected resolved plan bytes, got %s", string(imports.applyIn.ResolvedPlanJSON))
	}
	if imports.applyZipContent != "zip-bytes" {
		t.Fatalf("expected uploaded zip bytes during service call, got %q", imports.applyZipContent)
	}
	if _, err := os.Stat(imports.applyIn.ZipPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected temp zip to be deleted after request, got %v", err)
	}

	var out applyResultResponse
	if err := json.NewDecoder(response.Body).Decode(&out); err != nil {
		t.Fatalf("expected apply result json, got %v", err)
	}
	if out.AggregatesSucceeded != 1 || out.AggregatesFailed != 0 {
		t.Fatalf("expected successful apply response, got %+v", out)
	}
}

func TestImportApplyEndpointReturnsPartialFailures(t *testing.T) {
	failed := failedOperationFixture(t)
	imports := &importServiceFake{
		applyOut: core.ApplyPlanOutput{Result: applyResultFixture(t, []domain.FailedOperation{failed})},
	}
	server := newTestServerWithImport(t, imports)

	response := postImportMultipart(t, server, "/import/apply", importMultipartRequest{zip: []byte("zip-bytes")})

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal server error response, got %d body=%s", response.Code, response.Body.String())
	}
	var out applyResultResponse
	if err := json.NewDecoder(response.Body).Decode(&out); err != nil {
		t.Fatalf("expected apply result json, got %v", err)
	}
	if out.AggregatesFailed != 1 || len(out.Failed) != 1 || out.Failed[0].Error != "course create failed" {
		t.Fatalf("expected failed operation details, got %+v", out)
	}
}

func TestImportEndpointMapsErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "missing zip", want: http.StatusBadRequest},
		{name: "parse", err: domain.NewImportSourceParseError("course.zip", "invalid yaml", nil), want: http.StatusBadRequest},
		{name: "hash mismatch", err: domain.NewImportPlanHashMismatchError(strings.Repeat("a", 64), strings.Repeat("b", 64)), want: http.StatusBadRequest},
		{name: "unresolved", err: domain.NewUnresolvedImportConflictsError([]domain.ImportConflict{importConflictFixture(t)}), want: http.StatusConflict},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			imports := &importServiceFake{planErr: test.err}
			server := newTestServerWithImport(t, imports)
			request := importMultipartRequest{zip: []byte("zip-bytes")}
			if test.name == "missing zip" {
				request.zip = nil
			}

			response := postImportMultipart(t, server, "/import/plan", request)

			if response.Code != test.want {
				t.Fatalf("expected status %d, got %d body=%s", test.want, response.Code, response.Body.String())
			}
			if test.name == "missing zip" && imports.called != "" {
				t.Fatalf("expected missing zip not to reach service, got %q", imports.called)
			}
		})
	}
}

func TestRunExecutesCourseCreateThroughCobra(t *testing.T) {
	courses := &courseServiceFake{createOut: core.CreateCourseOutput{ID: courseIDValue}}
	server := newTestServer(t, courses, &lessonServiceFake{})

	result, status := postRun(t, server, url.Values{
		"action":        {"course-create"},
		"title":         {"Intro to Go"},
		"slug":          {"intro-to-go"},
		"description":   {"Learn Go"},
		"instructor-id": {instructorIDValue},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok course-create, status=%d result=%+v", status, result)
	}
	if result.Output != courseIDValue {
		t.Fatalf("expected created course id output, got %q", result.Output)
	}
	want := core.CreateCourseInput{
		Title:        "Intro to Go",
		Slug:         "intro-to-go",
		Description:  "Learn Go",
		InstructorID: instructorIDValue,
	}
	if courses.createIn != want || courses.called != "create" {
		t.Fatalf("expected create input %+v, got called=%q input=%+v", want, courses.called, courses.createIn)
	}
}

func TestRunSurfacesValidationFailures(t *testing.T) {
	courses := &courseServiceFake{}
	server := newTestServer(t, courses, &lessonServiceFake{})

	result, status := postRun(t, server, url.Values{
		"action": {"course-create"},
		"slug":   {"intro-to-go"},
	})

	if status != http.StatusBadRequest || result.OK {
		t.Fatalf("expected validation failure, status=%d result=%+v", status, result)
	}
	if !strings.Contains(result.Output, "--title") {
		t.Fatalf("expected missing title in output, got %q", result.Output)
	}
	if courses.called != "" {
		t.Fatalf("expected service not to be called, got %q", courses.called)
	}
}

func TestRunExecutesLessonListThroughCobra(t *testing.T) {
	lessons := &lessonServiceFake{
		listOut: core.ListLessonsOutput{Lessons: []core.LessonView{lessonViewFixture()}},
	}
	server := newTestServer(t, &courseServiceFake{}, lessons)

	result, status := postRun(t, server, url.Values{
		"action":    {"lesson-list"},
		"course-id": {courseIDValue},
		"output":    {"quiet"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok lesson-list, status=%d result=%+v", status, result)
	}
	if result.Output != lessonIDValue {
		t.Fatalf("expected quiet lesson id output, got %q", result.Output)
	}
	if lessons.called != "list" || lessons.listIn.CourseID != courseIDValue {
		t.Fatalf("expected lesson list input, got called=%q input=%+v", lessons.called, lessons.listIn)
	}
}

func TestRunExecutesLessonBlockAddThroughCobra(t *testing.T) {
	lessons := &lessonServiceFake{addBlockOut: core.AddLessonBlockOutput{ID: blockIDValue}}
	server := newTestServer(t, &courseServiceFake{}, lessons)

	result, status := postRun(t, server, url.Values{
		"action":    {"lesson-block-add"},
		"lesson-id": {lessonIDValue},
		"kind":      {"text"},
		"text":      {"## Intro"},
		"position":  {"2"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok lesson-block-add, status=%d result=%+v", status, result)
	}
	if result.Output != blockIDValue {
		t.Fatalf("expected created block id output, got %q", result.Output)
	}
	if lessons.called != "add-block" {
		t.Fatalf("expected add block call, got %q", lessons.called)
	}
	if lessons.addBlockIn.LessonID != lessonIDValue || lessons.addBlockIn.Kind != "text" || lessons.addBlockIn.Markdown != "## Intro" {
		t.Fatalf("expected add block input to map, got %+v", lessons.addBlockIn)
	}
	if lessons.addBlockIn.Position == nil || *lessons.addBlockIn.Position != 2 {
		t.Fatalf("expected position 2, got %v", lessons.addBlockIn.Position)
	}
}

func TestRunEmbedsQuizThroughLessonBlockAdd(t *testing.T) {
	lessons := &lessonServiceFake{addBlockOut: core.AddLessonBlockOutput{ID: blockIDValue}}
	server := newTestServer(t, &courseServiceFake{}, lessons)

	result, status := postRun(t, server, url.Values{
		"action":    {"lesson-block-add"},
		"lesson-id": {lessonIDValue},
		"kind":      {"quiz"},
		"quiz-id":   {quizIDValue},
		"position":  {"0"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok quiz lesson-block-add, status=%d result=%+v", status, result)
	}
	if result.Output != blockIDValue {
		t.Fatalf("expected created block id output, got %q", result.Output)
	}
	if lessons.called != "add-block" {
		t.Fatalf("expected add block call, got %q", lessons.called)
	}
	if lessons.addBlockIn.LessonID != lessonIDValue || lessons.addBlockIn.Kind != "quiz" || lessons.addBlockIn.QuizRef != quizIDValue {
		t.Fatalf("expected quiz block input to map, got %+v", lessons.addBlockIn)
	}
	if lessons.addBlockIn.Position == nil || *lessons.addBlockIn.Position != 0 {
		t.Fatalf("expected position 0, got %v", lessons.addBlockIn.Position)
	}
}

func TestRunSurfacesCrossCourseQuizEmbedValidation(t *testing.T) {
	lessons := &lessonServiceFake{addBlockErr: domain.NewValidationError("quiz_ref", "must belong to the lesson course")}
	server := newTestServer(t, &courseServiceFake{}, lessons)

	result, status := postRun(t, server, url.Values{
		"action":    {"lesson-block-add"},
		"lesson-id": {lessonIDValue},
		"kind":      {"quiz"},
		"quiz-id":   {quizIDValue},
	})

	if status != http.StatusBadRequest || result.OK {
		t.Fatalf("expected validation failure, status=%d result=%+v", status, result)
	}
	if !strings.Contains(result.Output, "quiz_ref") || !strings.Contains(result.Output, "must belong to the lesson course") {
		t.Fatalf("expected cross-course validation output, got %q", result.Output)
	}
}

func TestRunExecutesQuizCreateThroughCobra(t *testing.T) {
	quizzes := &quizServiceFake{createOut: core.CreateQuizOutput{ID: quizIDValue}}
	server := newTestServer(t, &courseServiceFake{}, &lessonServiceFake{}, quizzes)

	result, status := postRun(t, server, url.Values{
		"action":         {"quiz-create"},
		"course-id":      {courseIDValue},
		"title":          {"Go Basics Quiz"},
		"pass-threshold": {"0.8"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok quiz-create, status=%d result=%+v", status, result)
	}
	if result.Output != quizIDValue {
		t.Fatalf("expected created quiz id output, got %q", result.Output)
	}
	if quizzes.called != "create" || quizzes.createIn.CourseID != courseIDValue || quizzes.createIn.Title != "Go Basics Quiz" {
		t.Fatalf("expected quiz create input, got called=%q input=%+v", quizzes.called, quizzes.createIn)
	}
	if quizzes.createIn.PassThreshold == nil || *quizzes.createIn.PassThreshold != 0.8 {
		t.Fatalf("expected pass threshold 0.8, got %+v", quizzes.createIn.PassThreshold)
	}
}

func TestRunExecutesQuizQuestionAddThroughCobra(t *testing.T) {
	quizzes := &quizServiceFake{addQuestionOut: core.AddQuestionOutput{ID: questionIDValue}}
	server := newTestServer(t, &courseServiceFake{}, &lessonServiceFake{}, quizzes)

	result, status := postRun(t, server, url.Values{
		"action":      {"quiz-question-add"},
		"quiz-id":     {quizIDValue},
		"type":        {"multiple"},
		"prompt":      {"Pick two"},
		"option":      {"Option A\nOption B\nOption C"},
		"correct":     {"0,2"},
		"explanation": {"A and C"},
		"position":    {"1"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok quiz-question-add, status=%d result=%+v", status, result)
	}
	if result.Output != questionIDValue {
		t.Fatalf("expected created question id output, got %q", result.Output)
	}
	if quizzes.called != "add-question" || quizzes.addQuestionIn.QuizID != quizIDValue || quizzes.addQuestionIn.Type != "multiple" {
		t.Fatalf("expected question add input, got called=%q input=%+v", quizzes.called, quizzes.addQuestionIn)
	}
	if !slices.Equal(quizzes.addQuestionIn.Options, []string{"Option A", "Option B", "Option C"}) {
		t.Fatalf("expected options to split by line, got %#v", quizzes.addQuestionIn.Options)
	}
	if !slices.Equal(quizzes.addQuestionIn.CorrectIndices, []int{0, 2}) {
		t.Fatalf("expected correct indices to split by comma, got %#v", quizzes.addQuestionIn.CorrectIndices)
	}
	if quizzes.addQuestionIn.Explanation != "A and C" {
		t.Fatalf("expected explanation to map, got %q", quizzes.addQuestionIn.Explanation)
	}
	if quizzes.addQuestionIn.Position == nil || *quizzes.addQuestionIn.Position != 1 {
		t.Fatalf("expected position 1, got %v", quizzes.addQuestionIn.Position)
	}
}

func TestRunExecutesQuizQuestionUpdateThroughCobra(t *testing.T) {
	quizzes := &quizServiceFake{updateQuestionOut: core.UpdateQuestionOutput{ID: questionIDValue}}
	server := newTestServer(t, &courseServiceFake{}, &lessonServiceFake{}, quizzes)

	result, status := postRun(t, server, url.Values{
		"action":      {"quiz-question-update"},
		"question-id": {questionIDValue},
		"prompt":      {"Updated prompt"},
		"option":      {"Option B\nOption C"},
		"correct":     {"1"},
		"explanation": {"Updated explanation"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok quiz-question-update, status=%d result=%+v", status, result)
	}
	if quizzes.called != "update-question" || quizzes.updateQuestionIn.ID != questionIDValue {
		t.Fatalf("expected question update input, got called=%q input=%+v", quizzes.called, quizzes.updateQuestionIn)
	}
	if quizzes.updateQuestionIn.Prompt == nil || *quizzes.updateQuestionIn.Prompt != "Updated prompt" {
		t.Fatalf("expected prompt pointer, got %+v", quizzes.updateQuestionIn)
	}
	if quizzes.updateQuestionIn.Options == nil || !slices.Equal(*quizzes.updateQuestionIn.Options, []string{"Option B", "Option C"}) {
		t.Fatalf("expected options pointer, got %+v", quizzes.updateQuestionIn)
	}
	if quizzes.updateQuestionIn.CorrectIndices == nil || !slices.Equal(*quizzes.updateQuestionIn.CorrectIndices, []int{1}) {
		t.Fatalf("expected correct indices pointer, got %+v", quizzes.updateQuestionIn)
	}
	if quizzes.updateQuestionIn.Explanation == nil || *quizzes.updateQuestionIn.Explanation != "Updated explanation" {
		t.Fatalf("expected explanation pointer, got %+v", quizzes.updateQuestionIn)
	}
}

func TestRunExecutesQuizQuestionReorderThroughCobra(t *testing.T) {
	quizzes := &quizServiceFake{}
	server := newTestServer(t, &courseServiceFake{}, &lessonServiceFake{}, quizzes)

	result, status := postRun(t, server, url.Values{
		"action":  {"quiz-question-reorder"},
		"quiz-id": {quizIDValue},
		"order":   {questionIDValue + ":1," + otherQuestionIDValue + ":0"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok quiz-question-reorder, status=%d result=%+v", status, result)
	}
	if result.Output != "quiz questions reordered" {
		t.Fatalf("expected reorder confirmation output, got %q", result.Output)
	}
	if quizzes.called != "reorder-questions" || quizzes.reorderQuestionsIn.QuizID != quizIDValue {
		t.Fatalf("expected reorder questions input, got called=%q input=%+v", quizzes.called, quizzes.reorderQuestionsIn)
	}

	want := []core.QuestionPlacementDTO{
		{QuestionID: questionIDValue, Position: 1},
		{QuestionID: otherQuestionIDValue, Position: 0},
	}
	if !slices.Equal(quizzes.reorderQuestionsIn.Order, want) {
		t.Fatalf("expected question placements %+v, got %+v", want, quizzes.reorderQuestionsIn.Order)
	}
}

func TestRunSurfacesQuizInUseDeleteConflict(t *testing.T) {
	quizzes := &quizServiceFake{deleteErr: domain.NewQuizInUseError([]domain.LessonID{mustLessonID(t, lessonIDValue)})}
	server := newTestServer(t, &courseServiceFake{}, &lessonServiceFake{}, quizzes)

	result, status := postRun(t, server, url.Values{
		"action":  {"quiz-delete"},
		"quiz-id": {quizIDValue},
		"force":   {"true"},
	})

	if status != http.StatusBadRequest || result.OK {
		t.Fatalf("expected quiz in use failure, status=%d result=%+v", status, result)
	}
	if !strings.Contains(result.Output, "quiz is embedded in lessons") || !strings.Contains(result.Output, lessonIDValue) {
		t.Fatalf("expected quiz in use details, got %q", result.Output)
	}
}

func TestRunExecutesPracticeCreateThroughCobra(t *testing.T) {
	practices := &practiceServiceFake{createOut: core.CreatePracticeOutput{ID: practiceIDValue}}
	server := newTestServerWithPractice(t, &courseServiceFake{}, &lessonServiceFake{}, practices)

	result, status := postRun(t, server, url.Values{
		"action":       {"practice-create"},
		"course-id":    {courseIDValue},
		"title":        {"FizzBuzz"},
		"language":     {"golang"},
		"prompt":       {"Print fizz buzz"},
		"starter-code": {"package main"},
		"solution":     {"fmt.Println()"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok practice-create, status=%d result=%+v", status, result)
	}
	if result.Output != practiceIDValue {
		t.Fatalf("expected created practice id output, got %q", result.Output)
	}
	want := core.CreatePracticeInput{
		CourseID:    courseIDValue,
		Title:       "FizzBuzz",
		Language:    "golang",
		Prompt:      "Print fizz buzz",
		StarterCode: "package main",
		Solution:    "fmt.Println()",
	}
	if practices.called != "create" || practices.createIn != want {
		t.Fatalf("expected practice create input %+v, got called=%q input=%+v", want, practices.called, practices.createIn)
	}
}

func TestRunExecutesPracticeTestCaseAddUpdateAndReorderThroughCobra(t *testing.T) {
	t.Run("add", func(t *testing.T) {
		practices := &practiceServiceFake{addTestCaseOut: core.AddTestCaseOutput{ID: testCaseIDValue}}
		server := newTestServerWithPractice(t, &courseServiceFake{}, &lessonServiceFake{}, practices)

		result, status := postRun(t, server, url.Values{
			"action":          {"practice-testcase-add"},
			"practice-id":     {practiceIDValue},
			"stdin":           {"3"},
			"expected-stdout": {"Fizz"},
			"name":            {"multiple of three"},
			"position":        {"1"},
		})

		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok practice-testcase-add, status=%d result=%+v", status, result)
		}
		if result.Output != testCaseIDValue {
			t.Fatalf("expected created test case id output, got %q", result.Output)
		}
		if practices.called != "add-testcase" || practices.addTestCaseIn.PracticeID != practiceIDValue {
			t.Fatalf("expected add test case input, got called=%q input=%+v", practices.called, practices.addTestCaseIn)
		}
		if practices.addTestCaseIn.Stdin != "3" || practices.addTestCaseIn.ExpectedStdout != "Fizz" || practices.addTestCaseIn.Name != "multiple of three" {
			t.Fatalf("expected test case content to map, got %+v", practices.addTestCaseIn)
		}
		if practices.addTestCaseIn.Position == nil || *practices.addTestCaseIn.Position != 1 {
			t.Fatalf("expected position 1, got %+v", practices.addTestCaseIn.Position)
		}
	})

	t.Run("update allows empty values", func(t *testing.T) {
		practices := &practiceServiceFake{updateTestCaseOut: core.UpdateTestCaseOutput{ID: testCaseIDValue}}
		server := newTestServerWithPractice(t, &courseServiceFake{}, &lessonServiceFake{}, practices)

		result, status := postRun(t, server, url.Values{
			"action":          {"practice-testcase-update"},
			"testcase-id":     {testCaseIDValue},
			"stdin":           {"updated input"},
			"expected-stdout": {""},
			"name":            {""},
		})

		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok practice-testcase-update, status=%d result=%+v", status, result)
		}
		if practices.called != "update-testcase" || practices.updateTestCaseIn.ID != testCaseIDValue {
			t.Fatalf("expected update test case input, got called=%q input=%+v", practices.called, practices.updateTestCaseIn)
		}
		if practices.updateTestCaseIn.Stdin == nil || *practices.updateTestCaseIn.Stdin != "updated input" {
			t.Fatalf("expected stdin pointer, got %+v", practices.updateTestCaseIn)
		}
		if practices.updateTestCaseIn.ExpectedStdout == nil || *practices.updateTestCaseIn.ExpectedStdout != "" {
			t.Fatalf("expected empty expected stdout pointer, got %+v", practices.updateTestCaseIn)
		}
		if practices.updateTestCaseIn.Name == nil || *practices.updateTestCaseIn.Name != "" {
			t.Fatalf("expected empty name pointer, got %+v", practices.updateTestCaseIn)
		}
	})

	t.Run("reorder", func(t *testing.T) {
		practices := &practiceServiceFake{}
		server := newTestServerWithPractice(t, &courseServiceFake{}, &lessonServiceFake{}, practices)

		result, status := postRun(t, server, url.Values{
			"action":      {"practice-testcase-reorder"},
			"practice-id": {practiceIDValue},
			"order":       {testCaseIDValue + ":1," + otherTestCaseIDValue + ":0"},
		})

		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok practice-testcase-reorder, status=%d result=%+v", status, result)
		}
		if result.Output != "practice test cases reordered" {
			t.Fatalf("expected reorder confirmation output, got %q", result.Output)
		}
		if practices.called != "reorder-testcases" || practices.reorderTestCasesIn.PracticeID != practiceIDValue {
			t.Fatalf("expected reorder test cases input, got called=%q input=%+v", practices.called, practices.reorderTestCasesIn)
		}

		want := []core.TestCasePlacementDTO{
			{TestCaseID: testCaseIDValue, Position: 1},
			{TestCaseID: otherTestCaseIDValue, Position: 0},
		}
		if !slices.Equal(practices.reorderTestCasesIn.Order, want) {
			t.Fatalf("expected test case placements %+v, got %+v", want, practices.reorderTestCasesIn.Order)
		}
	})
}

func TestRunEmbedsPracticeThroughLessonBlockAdd(t *testing.T) {
	lessons := &lessonServiceFake{addBlockOut: core.AddLessonBlockOutput{ID: blockIDValue}}
	server := newTestServer(t, &courseServiceFake{}, lessons)

	result, status := postRun(t, server, url.Values{
		"action":      {"lesson-block-add"},
		"lesson-id":   {lessonIDValue},
		"kind":        {"practice"},
		"practice-id": {practiceIDValue},
		"position":    {"0"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok practice lesson-block-add, status=%d result=%+v", status, result)
	}
	if result.Output != blockIDValue {
		t.Fatalf("expected created block id output, got %q", result.Output)
	}
	if lessons.called != "add-block" {
		t.Fatalf("expected add block call, got %q", lessons.called)
	}
	if lessons.addBlockIn.LessonID != lessonIDValue || lessons.addBlockIn.Kind != "practice" || lessons.addBlockIn.PracticeRef != practiceIDValue {
		t.Fatalf("expected practice block input to map, got %+v", lessons.addBlockIn)
	}
	if lessons.addBlockIn.Position == nil || *lessons.addBlockIn.Position != 0 {
		t.Fatalf("expected position 0, got %v", lessons.addBlockIn.Position)
	}
}

func TestRunSurfacesPracticeFailureStates(t *testing.T) {
	t.Run("cross-course practice embed", func(t *testing.T) {
		lessons := &lessonServiceFake{addBlockErr: domain.NewValidationError("practice_ref", "must belong to the lesson course")}
		server := newTestServer(t, &courseServiceFake{}, lessons)

		result, status := postRun(t, server, url.Values{
			"action":      {"lesson-block-add"},
			"lesson-id":   {lessonIDValue},
			"kind":        {"practice"},
			"practice-id": {practiceIDValue},
		})

		if status != http.StatusBadRequest || result.OK {
			t.Fatalf("expected validation failure, status=%d result=%+v", status, result)
		}
		if !strings.Contains(result.Output, "practice_ref") || !strings.Contains(result.Output, "must belong to the lesson course") {
			t.Fatalf("expected cross-course validation output, got %q", result.Output)
		}
	})

	t.Run("practice in use delete", func(t *testing.T) {
		practices := &practiceServiceFake{deleteErr: domain.NewPracticeInUseError([]domain.LessonID{mustLessonID(t, lessonIDValue)})}
		server := newTestServerWithPractice(t, &courseServiceFake{}, &lessonServiceFake{}, practices)

		result, status := postRun(t, server, url.Values{
			"action":      {"practice-delete"},
			"practice-id": {practiceIDValue},
			"force":       {"true"},
		})

		if status != http.StatusBadRequest || result.OK {
			t.Fatalf("expected practice in use failure, status=%d result=%+v", status, result)
		}
		if !strings.Contains(result.Output, "practice is embedded in lessons") || !strings.Contains(result.Output, lessonIDValue) {
			t.Fatalf("expected practice in use details, got %q", result.Output)
		}
	})
}

func TestRunExecutesTestCreateAndUpdateThroughCobra(t *testing.T) {
	t.Run("create metadata", func(t *testing.T) {
		tests := &testServiceFake{createOut: core.CreateTestOutput{ID: testIDValue}}
		server := newTestServerWithTest(t, &courseServiceFake{}, &lessonServiceFake{}, tests)

		result, status := postRun(t, server, url.Values{
			"action":             {"test-create"},
			"course-id":          {courseIDValue},
			"title":              {"Final Test"},
			"time-limit-minutes": {"45"},
			"pass-threshold":     {"0.8"},
		})

		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok test-create, status=%d result=%+v", status, result)
		}
		if result.Output != testIDValue {
			t.Fatalf("expected created test id output, got %q", result.Output)
		}
		if tests.called != "create" || tests.createIn.CourseID != courseIDValue || tests.createIn.Title != "Final Test" {
			t.Fatalf("expected test create input, got called=%q input=%+v", tests.called, tests.createIn)
		}
		if tests.createIn.TimeLimitMinutes == nil || *tests.createIn.TimeLimitMinutes != 45 {
			t.Fatalf("expected time limit 45, got %+v", tests.createIn.TimeLimitMinutes)
		}
		if tests.createIn.PassThreshold == nil || *tests.createIn.PassThreshold != 0.8 {
			t.Fatalf("expected pass threshold 0.8, got %+v", tests.createIn.PassThreshold)
		}
	})

	t.Run("update solution group", func(t *testing.T) {
		tests := &testServiceFake{updateOut: core.UpdateTestOutput{ID: testIDValue}}
		server := newTestServerWithTest(t, &courseServiceFake{}, &lessonServiceFake{}, tests)

		result, status := postRun(t, server, url.Values{
			"action":                  {"test-update"},
			"test-id":                 {testIDValue},
			"title":                   {"Updated Test"},
			"time-limit-minutes":      {"0"},
			"pass-threshold":          {"0.9"},
			"solution-zip-provider":   {"url"},
			"solution-zip-locator":    {"https://example.com/solution.zip"},
			"solution-video-provider": {"url"},
			"solution-video-locator":  {"https://example.com/video.mp4"},
			"solution-video-caption":  {""},
		})

		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok test-update, status=%d result=%+v", status, result)
		}
		if tests.called != "update" || tests.updateIn.ID != testIDValue {
			t.Fatalf("expected update test input, got called=%q input=%+v", tests.called, tests.updateIn)
		}
		if tests.updateIn.Title == nil || *tests.updateIn.Title != "Updated Test" {
			t.Fatalf("expected title pointer, got %+v", tests.updateIn)
		}
		if tests.updateIn.TimeLimitMinutes == nil || *tests.updateIn.TimeLimitMinutes != 0 {
			t.Fatalf("expected zero time limit pointer, got %+v", tests.updateIn)
		}
		if tests.updateIn.SolutionZipProvider == nil || *tests.updateIn.SolutionZipProvider != "url" ||
			tests.updateIn.SolutionZipLocator == nil || *tests.updateIn.SolutionZipLocator != "https://example.com/solution.zip" ||
			tests.updateIn.SolutionVideoProvider == nil || *tests.updateIn.SolutionVideoProvider != "url" ||
			tests.updateIn.SolutionVideoLocator == nil || *tests.updateIn.SolutionVideoLocator != "https://example.com/video.mp4" ||
			tests.updateIn.SolutionVideoCaption == nil || *tests.updateIn.SolutionVideoCaption != "" {
			t.Fatalf("expected solution package fields to map atomically, got %+v", tests.updateIn)
		}
	})
}

func TestRunExecutesTestItemWorkflowsThroughCobra(t *testing.T) {
	t.Run("add choice item", func(t *testing.T) {
		tests := &testServiceFake{addItemOut: core.AddTestItemOutput{ID: testItemIDValue}}
		server := newTestServerWithTest(t, &courseServiceFake{}, &lessonServiceFake{}, tests)

		result, status := postRun(t, server, url.Values{
			"action":      {"test-item-add"},
			"test-id":     {testIDValue},
			"kind":        {"choice"},
			"prompt":      {"Pick two"},
			"type":        {"multiple"},
			"option":      {"A\nB"},
			"correct":     {"0,1"},
			"explanation": {"A and B"},
			"position":    {"1"},
		})

		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok test-item-add choice, status=%d result=%+v", status, result)
		}
		if result.Output != testItemIDValue {
			t.Fatalf("expected created item id output, got %q", result.Output)
		}
		if tests.called != "add-item" || tests.addItemIn.TestID != testIDValue || tests.addItemIn.Kind != "choice" {
			t.Fatalf("expected add choice input, got called=%q input=%+v", tests.called, tests.addItemIn)
		}
		if tests.addItemIn.Prompt != "Pick two" || tests.addItemIn.ChoiceType != "multiple" || tests.addItemIn.Explanation != "A and B" {
			t.Fatalf("expected choice item payload, got %+v", tests.addItemIn)
		}
		if !slices.Equal(tests.addItemIn.Options, []string{"A", "B"}) || !slices.Equal(tests.addItemIn.CorrectIndices, []int{0, 1}) {
			t.Fatalf("expected options and correct indices to map, got %+v", tests.addItemIn)
		}
		if tests.addItemIn.Position == nil || *tests.addItemIn.Position != 1 {
			t.Fatalf("expected item position 1, got %+v", tests.addItemIn.Position)
		}
	})

	t.Run("add coding item", func(t *testing.T) {
		tests := &testServiceFake{addItemOut: core.AddTestItemOutput{ID: testItemIDValue}}
		server := newTestServerWithTest(t, &courseServiceFake{}, &lessonServiceFake{}, tests)

		result, status := postRun(t, server, url.Values{
			"action":       {"test-item-add"},
			"test-id":      {testIDValue},
			"kind":         {"coding"},
			"prompt":       {"Write code"},
			"language":     {"golang"},
			"starter-code": {"package main"},
			"solution":     {"func main() {}"},
			"testcase":     {"1::1::sample\n::ok"},
		})

		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok test-item-add coding, status=%d result=%+v", status, result)
		}
		if tests.called != "add-item" || tests.addItemIn.CodingPrompt != "Write code" || tests.addItemIn.Language != "golang" {
			t.Fatalf("expected coding item input, got called=%q input=%+v", tests.called, tests.addItemIn)
		}
		if tests.addItemIn.StarterCode != "package main" || tests.addItemIn.Solution != "func main() {}" {
			t.Fatalf("expected coding source fields, got %+v", tests.addItemIn)
		}
		wantCases := []core.CodingTestCaseDTO{
			{Stdin: "1", ExpectedStdout: "1", Name: "sample"},
			{Stdin: "", ExpectedStdout: "ok"},
		}
		if !slices.Equal(tests.addItemIn.TestCases, wantCases) {
			t.Fatalf("expected coding cases %+v, got %+v", wantCases, tests.addItemIn.TestCases)
		}
	})

	t.Run("read selected item detail", func(t *testing.T) {
		item := testItemViewFixture()
		tests := &testServiceFake{
			listItemsOut: core.ListTestItemsOutput{Items: []core.TestItemView{item}},
			getItemOut:   core.GetTestItemOutput{Item: item},
		}
		server := newTestServerWithTest(t, &courseServiceFake{}, &lessonServiceFake{}, tests)

		result, status := postRun(t, server, url.Values{
			"action":  {"test-item-list"},
			"test-id": {testIDValue},
			"output":  {"quiet"},
		})
		if status != http.StatusOK || !result.OK || result.Output != testItemIDValue {
			t.Fatalf("expected quiet test item list output, status=%d result=%+v", status, result)
		}
		if tests.called != "list-items" || tests.listItemsIn.TestID != testIDValue {
			t.Fatalf("expected list items input, got called=%q input=%+v", tests.called, tests.listItemsIn)
		}

		result, status = postRun(t, server, url.Values{
			"action":  {"test-item-get"},
			"item-id": {testItemIDValue},
			"output":  {"table"},
		})
		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok test-item-get, status=%d result=%+v", status, result)
		}
		if !strings.Contains(result.Output, "CHOICE_OPTIONS") || !strings.Contains(result.Output, "A | B") {
			t.Fatalf("expected full item detail output, got %q", result.Output)
		}
		if tests.called != "get-item" || tests.getItemIn.ID != testItemIDValue {
			t.Fatalf("expected get item input, got called=%q input=%+v", tests.called, tests.getItemIn)
		}
	})

	t.Run("update item and reorder", func(t *testing.T) {
		tests := &testServiceFake{updateItemOut: core.UpdateTestItemOutput{ID: testItemIDValue}}
		server := newTestServerWithTest(t, &courseServiceFake{}, &lessonServiceFake{}, tests)

		result, status := postRun(t, server, url.Values{
			"action":       {"test-item-update"},
			"item-id":      {testItemIDValue},
			"prompt":       {"Updated prompt"},
			"type":         {"single"},
			"option":       {"A\nB"},
			"correct":      {"1"},
			"explanation":  {""},
			"language":     {"rust"},
			"starter-code": {""},
			"solution":     {"updated solution"},
			"testcase":     {"stdin::stdout::case"},
		})
		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok test-item-update, status=%d result=%+v", status, result)
		}
		if tests.called != "update-item" || tests.updateItemIn.ID != testItemIDValue {
			t.Fatalf("expected update item input, got called=%q input=%+v", tests.called, tests.updateItemIn)
		}
		if tests.updateItemIn.Explanation == nil || *tests.updateItemIn.Explanation != "" ||
			tests.updateItemIn.StarterCode == nil || *tests.updateItemIn.StarterCode != "" {
			t.Fatalf("expected deliberate empty fields to map, got %+v", tests.updateItemIn)
		}
		if tests.updateItemIn.Options == nil || !slices.Equal(*tests.updateItemIn.Options, []string{"A", "B"}) ||
			tests.updateItemIn.CorrectIndices == nil || !slices.Equal(*tests.updateItemIn.CorrectIndices, []int{1}) {
			t.Fatalf("expected updated choice payload to map, got %+v", tests.updateItemIn)
		}
		if tests.updateItemIn.TestCases == nil || !slices.Equal(*tests.updateItemIn.TestCases, []core.CodingTestCaseDTO{{Stdin: "stdin", ExpectedStdout: "stdout", Name: "case"}}) {
			t.Fatalf("expected updated coding cases to map, got %+v", tests.updateItemIn)
		}

		result, status = postRun(t, server, url.Values{
			"action":  {"test-item-reorder"},
			"test-id": {testIDValue},
			"order":   {testItemIDValue + ":1," + otherTestItemIDValue + ":0"},
		})
		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok test-item-reorder, status=%d result=%+v", status, result)
		}
		if result.Output != "test items reordered" {
			t.Fatalf("expected reorder confirmation, got %q", result.Output)
		}
		want := []core.TestItemPlacementDTO{
			{TestItemID: testItemIDValue, Position: 1},
			{TestItemID: otherTestItemIDValue, Position: 0},
		}
		if tests.called != "reorder-items" || tests.reorderItemsIn.TestID != testIDValue || !slices.Equal(tests.reorderItemsIn.Order, want) {
			t.Fatalf("expected reorder input %+v, got called=%q input=%+v", want, tests.called, tests.reorderItemsIn)
		}
	})

	t.Run("remove item", func(t *testing.T) {
		tests := &testServiceFake{}
		server := newTestServerWithTest(t, &courseServiceFake{}, &lessonServiceFake{}, tests)

		result, status := postRun(t, server, url.Values{
			"action":  {"test-item-remove"},
			"item-id": {testItemIDValue},
			"force":   {"true"},
		})

		if status != http.StatusOK || !result.OK {
			t.Fatalf("expected ok test-item-remove, status=%d result=%+v", status, result)
		}
		if result.Output != "test item removed" {
			t.Fatalf("expected remove confirmation, got %q", result.Output)
		}
		if tests.called != "remove-item" || tests.removeItemIn.ID != testItemIDValue {
			t.Fatalf("expected remove item input, got called=%q input=%+v", tests.called, tests.removeItemIn)
		}
	})
}

func TestRunExecutesLessonBlockRemoveThroughCobra(t *testing.T) {
	lessons := &lessonServiceFake{}
	server := newTestServer(t, &courseServiceFake{}, lessons)

	result, status := postRun(t, server, url.Values{
		"action":   {"lesson-block-remove"},
		"block-id": {blockIDValue},
		"force":    {"true"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok lesson-block-remove, status=%d result=%+v", status, result)
	}
	if result.Command != "course-cli lesson block remove "+blockIDValue+" --force" {
		t.Fatalf("expected positional block id preview, got %q", result.Command)
	}
	if result.Output != "lesson block removed" {
		t.Fatalf("expected remove confirmation output, got %q", result.Output)
	}
	if lessons.called != "remove-block" || lessons.removeBlockIn.ID != blockIDValue {
		t.Fatalf("expected remove block input, got called=%q input=%+v", lessons.called, lessons.removeBlockIn)
	}
}

func TestRunExecutesLessonBlockReorderThroughCobra(t *testing.T) {
	lessons := &lessonServiceFake{}
	server := newTestServer(t, &courseServiceFake{}, lessons)

	result, status := postRun(t, server, url.Values{
		"action":    {"lesson-block-reorder"},
		"lesson-id": {lessonIDValue},
		"order":     {blockIDValue + ":1," + otherBlockIDValue + ":0"},
	})

	if status != http.StatusOK || !result.OK {
		t.Fatalf("expected ok lesson-block-reorder, status=%d result=%+v", status, result)
	}
	if result.Output != "lesson blocks reordered" {
		t.Fatalf("expected reorder confirmation output, got %q", result.Output)
	}
	if lessons.called != "reorder-blocks" || lessons.reorderBlocksIn.LessonID != lessonIDValue {
		t.Fatalf("expected reorder block input, got called=%q input=%+v", lessons.called, lessons.reorderBlocksIn)
	}

	want := []core.BlockPlacementDTO{
		{BlockID: blockIDValue, Position: 1},
		{BlockID: otherBlockIDValue, Position: 0},
	}
	if len(lessons.reorderBlocksIn.Order) != len(want) {
		t.Fatalf("expected %d block placements, got %d", len(want), len(lessons.reorderBlocksIn.Order))
	}
	for index := range want {
		if lessons.reorderBlocksIn.Order[index] != want[index] {
			t.Fatalf("expected placement %+v at index %d, got %+v", want[index], index, lessons.reorderBlocksIn.Order[index])
		}
	}
}

func TestValidateLoopbackAddressRejectsPublicBind(t *testing.T) {
	if err := validateLoopbackAddress("127.0.0.1:8787"); err != nil {
		t.Fatalf("expected loopback address to pass, got %v", err)
	}
	if err := validateLoopbackAddress("0.0.0.0:8787"); err == nil {
		t.Fatalf("expected public bind address to be rejected")
	}
}

func newTestServer(t *testing.T, courses core.CourseService, lessons core.LessonService, quizzes ...core.QuizService) *Server {
	t.Helper()

	var quizService core.QuizService = &quizServiceFake{}
	if len(quizzes) > 0 {
		quizService = quizzes[0]
	}

	return newTestServerWithServices(t, courses, lessons, quizService, &practiceServiceFake{}, &testServiceFake{})
}

func newTestServerWithPractice(t *testing.T, courses core.CourseService, lessons core.LessonService, practices core.PracticeService) *Server {
	t.Helper()

	return newTestServerWithServices(t, courses, lessons, &quizServiceFake{}, practices, &testServiceFake{})
}

func newTestServerWithTest(t *testing.T, courses core.CourseService, lessons core.LessonService, tests core.TestService) *Server {
	t.Helper()

	return newTestServerWithServices(t, courses, lessons, &quizServiceFake{}, &practiceServiceFake{}, tests)
}

func newTestServerWithImport(t *testing.T, imports core.ImportService) *Server {
	t.Helper()

	return newTestServerWithServicesAndImport(t, &courseServiceFake{}, &lessonServiceFake{}, &quizServiceFake{}, &practiceServiceFake{}, &testServiceFake{}, imports)
}

func newTestServerWithServices(t *testing.T, courses core.CourseService, lessons core.LessonService, quizzes core.QuizService, practices core.PracticeService, tests core.TestService) *Server {
	t.Helper()

	return newTestServerWithServicesAndImport(t, courses, lessons, quizzes, practices, tests, &importServiceFake{})
}

func newTestServerWithServicesAndImport(t *testing.T, courses core.CourseService, lessons core.LessonService, quizzes core.QuizService, practices core.PracticeService, tests core.TestService, imports core.ImportService) *Server {
	t.Helper()

	config := viper.New()
	config.Set("instructor-id", instructorIDValue)
	server, err := NewServer(ServerOptions{
		Runner: NewRunner(Services{
			Course:   courses,
			Lesson:   lessons,
			Quiz:     quizzes,
			Practice: practices,
			Test:     tests,
			Import:   imports,
			Config:   config,
		}),
	})
	if err != nil {
		t.Fatalf("expected test server, got %v", err)
	}

	return server
}

func commandByID(t *testing.T, commands []Command, id string) Command {
	t.Helper()

	for _, command := range commands {
		if command.ID == id {
			return command
		}
	}

	t.Fatalf("expected catalog command %s", id)
	return Command{}
}

func runnableCobraCommands(root *cobra.Command) map[string]*cobra.Command {
	commands := make(map[string]*cobra.Command)

	var walk func(prefix []string, command *cobra.Command)
	walk = func(prefix []string, command *cobra.Command) {
		name := command.Name()
		if name == "" || name == "help" {
			return
		}

		path := append(prefix, name)
		children := command.Commands()
		if len(children) == 0 {
			if len(path) > 1 {
				commands[strings.Join(path[1:], " ")] = command
			}
			return
		}

		for _, child := range children {
			if child.Hidden {
				continue
			}
			walk(path, child)
		}
	}

	walk(nil, root)
	return commands
}

func assertCatalogFieldsMatchCobra(t *testing.T, command Command, cobraCommand *cobra.Command) {
	t.Helper()

	flagNames := cobraFlagNames(cobraCommand)
	seenFlags := make(map[string]bool, len(flagNames))
	positionals := 0

	for _, field := range command.Fields {
		if field.Binding.Argument {
			positionals++
		}
		if field.Binding.Flag == "" {
			continue
		}

		name := strings.TrimPrefix(field.Binding.Flag, "--")
		if !flagNames[name] {
			t.Fatalf("expected catalog command %s field %s to map to Cobra flag --%s", command.ID, field.Key, name)
		}
		seenFlags[name] = true
	}

	for name := range flagNames {
		if !seenFlags[name] {
			t.Fatalf("expected catalog command %s to expose Cobra flag --%s", command.ID, name)
		}
	}

	hasPositionalArg := strings.Contains(cobraCommand.Use, "<")
	if hasPositionalArg && positionals == 0 {
		t.Fatalf("expected catalog command %s to expose positional argument from %q", command.ID, cobraCommand.Use)
	}
	if !hasPositionalArg && positionals > 0 {
		t.Fatalf("expected catalog command %s not to expose positional arguments for %q", command.ID, cobraCommand.Use)
	}
}

func cobraFlagNames(command *cobra.Command) map[string]bool {
	names := make(map[string]bool)
	command.NonInheritedFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Name != "help" {
			names[flag.Name] = true
		}
	})

	return names
}

func fieldByKey(t *testing.T, command Command, key string) Field {
	t.Helper()

	for _, field := range command.Fields {
		if field.Key == key {
			return field
		}
	}

	t.Fatalf("expected command %s to include field %s", command.ID, key)
	return Field{}
}

func mustLessonID(t *testing.T, value string) domain.LessonID {
	t.Helper()

	id, err := domain.NewLessonID(value)
	if err != nil {
		t.Fatalf("expected lesson id, got %v", err)
	}

	return id
}

func postRun(t *testing.T, server *Server, values url.Values) (RunResult, int) {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/run", strings.NewReader(values.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	var result RunResult
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		t.Fatalf("expected run json, got %v", err)
	}

	return result, response.Code
}

type importMultipartRequest struct {
	zip              []byte
	resolvedPlanJSON []byte
}

func postImportMultipart(t *testing.T, server *Server, path string, request importMultipartRequest) *httptest.ResponseRecorder {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if request.zip != nil {
		writeMultipartFile(t, writer, "zip", "course.zip", request.zip)
	}
	if request.resolvedPlanJSON != nil {
		writeMultipartFile(t, writer, "resolved_plan", "resolved-plan.json", request.resolvedPlanJSON)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("expected multipart writer to close, got %v", err)
	}

	httpRequest := httptest.NewRequest(http.MethodPost, path, body)
	httpRequest.Header.Set("Content-Type", writer.FormDataContentType())
	response := httptest.NewRecorder()
	server.Handler().ServeHTTP(response, httpRequest)
	return response
}

func writeMultipartFile(t *testing.T, writer *multipart.Writer, field string, filename string, content []byte) {
	t.Helper()

	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("expected multipart file part, got %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("expected multipart content, got %v", err)
	}
}

type courseServiceFake struct {
	called string

	createIn  core.CreateCourseInput
	createOut core.CreateCourseOutput
	listIn    core.ListCoursesInput
	listOut   core.ListCoursesOutput
	getIn     core.GetCourseInput
	getOut    core.GetCourseOutput
	updateIn  core.UpdateCourseInput
	updateOut core.UpdateCourseOutput
	deleteIn  core.DeleteCourseInput
	publishIn core.PublishCourseInput
	unpubIn   core.UnpublishCourseInput
}

func (service *courseServiceFake) CreateCourse(in core.CreateCourseInput) (core.CreateCourseOutput, error) {
	service.called = "create"
	service.createIn = in
	return service.createOut, nil
}

func (service *courseServiceFake) ListCourses(in core.ListCoursesInput) (core.ListCoursesOutput, error) {
	service.called = "list"
	service.listIn = in
	return service.listOut, nil
}

func (service *courseServiceFake) GetCourse(in core.GetCourseInput) (core.GetCourseOutput, error) {
	service.called = "get"
	service.getIn = in
	return service.getOut, nil
}

func (service *courseServiceFake) UpdateCourse(in core.UpdateCourseInput) (core.UpdateCourseOutput, error) {
	service.called = "update"
	service.updateIn = in
	return service.updateOut, nil
}

func (service *courseServiceFake) DeleteCourse(in core.DeleteCourseInput) error {
	service.called = "delete"
	service.deleteIn = in
	return nil
}

func (service *courseServiceFake) PublishCourse(in core.PublishCourseInput) error {
	service.called = "publish"
	service.publishIn = in
	return nil
}

func (service *courseServiceFake) UnpublishCourse(in core.UnpublishCourseInput) error {
	service.called = "unpublish"
	service.unpubIn = in
	return nil
}

type importServiceFake struct {
	called string

	planIn         core.PlanImportInput
	planOut        core.PlanImportOutput
	planErr        error
	planZipContent string

	applyIn         core.ApplyPlanInput
	applyOut        core.ApplyPlanOutput
	applyErr        error
	applyZipContent string
}

func (service *importServiceFake) PlanImport(in core.PlanImportInput) (core.PlanImportOutput, error) {
	service.called = "plan"
	service.planIn = in
	service.planZipContent = readImportZipForTest(in.ZipPath)
	if service.planErr != nil {
		return core.PlanImportOutput{}, service.planErr
	}

	return service.planOut, nil
}

func (service *importServiceFake) ApplyPlan(in core.ApplyPlanInput) (core.ApplyPlanOutput, error) {
	service.called = "apply"
	service.applyIn = in
	service.applyZipContent = readImportZipForTest(in.ZipPath)
	if service.applyErr != nil {
		return core.ApplyPlanOutput{}, service.applyErr
	}

	return service.applyOut, nil
}

func readImportZipForTest(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return ""
	}

	return string(data)
}

func importPlanFixture(t *testing.T) domain.ImportPlan {
	t.Helper()

	plan, err := domain.NewImportPlan("1", strings.Repeat("a", 64), time.Unix(1, 0).UTC(), nil, nil)
	if err != nil {
		t.Fatalf("expected import plan fixture, got %v", err)
	}

	return plan
}

func applyResultFixture(t *testing.T, failed []domain.FailedOperation) domain.ApplyResult {
	t.Helper()

	aggregatesSucceeded := 1
	aggregatesFailed := 0
	if len(failed) > 0 {
		aggregatesSucceeded = 0
		aggregatesFailed = 1
	}
	result, err := domain.NewApplyResult(nil, failed, nil, aggregatesSucceeded, aggregatesFailed)
	if err != nil {
		t.Fatalf("expected apply result fixture, got %v", err)
	}

	return result
}

func failedOperationFixture(t *testing.T) domain.FailedOperation {
	t.Helper()

	failed, err := domain.NewFailedOperation(importOperationFixture(t), errors.New("course create failed"))
	if err != nil {
		t.Fatalf("expected failed operation fixture, got %v", err)
	}

	return failed
}

func importOperationFixture(t *testing.T) domain.ImportOperation {
	t.Helper()

	operation, err := domain.NewImportOperation(
		domain.CreateOperation(),
		domain.CourseEntity(),
		"course:intro",
		nil,
		json.RawMessage(`{"title":"Intro"}`),
	)
	if err != nil {
		t.Fatalf("expected import operation fixture, got %v", err)
	}

	return operation
}

func importConflictFixture(t *testing.T) domain.ImportConflict {
	t.Helper()

	candidate, err := domain.NewConflictCandidate(courseIDValue, "Existing course")
	if err != nil {
		t.Fatalf("expected conflict candidate fixture, got %v", err)
	}
	conflict, err := domain.NewImportConflict(
		domain.CourseEntity(),
		"course:intro",
		domain.SlugCollision(),
		[]domain.ConflictCandidate{candidate},
		domain.SkipOperation(),
		json.RawMessage(`{"title":"Intro"}`),
	)
	if err != nil {
		t.Fatalf("expected import conflict fixture, got %v", err)
	}

	return conflict
}

type lessonServiceFake struct {
	called string

	createIn        core.CreateLessonInput
	createOut       core.CreateLessonOutput
	listIn          core.ListLessonsInput
	listOut         core.ListLessonsOutput
	getIn           core.GetLessonInput
	getOut          core.GetLessonOutput
	updateIn        core.UpdateLessonInput
	updateOut       core.UpdateLessonOutput
	deleteIn        core.DeleteLessonInput
	reorderIn       core.ReorderLessonsInput
	addBlockIn      core.AddLessonBlockInput
	addBlockOut     core.AddLessonBlockOutput
	addBlockErr     error
	listBlocksIn    core.ListLessonBlocksInput
	listBlocksOut   core.ListLessonBlocksOutput
	getBlockIn      core.GetLessonBlockInput
	getBlockOut     core.GetLessonBlockOutput
	updateBlockIn   core.UpdateLessonBlockInput
	updateBlockOut  core.UpdateLessonBlockOutput
	removeBlockIn   core.RemoveLessonBlockInput
	reorderBlocksIn core.ReorderLessonBlocksInput
}

func (service *lessonServiceFake) CreateLesson(in core.CreateLessonInput) (core.CreateLessonOutput, error) {
	service.called = "create"
	service.createIn = in
	return service.createOut, nil
}

func (service *lessonServiceFake) ListLessons(in core.ListLessonsInput) (core.ListLessonsOutput, error) {
	service.called = "list"
	service.listIn = in
	return service.listOut, nil
}

func (service *lessonServiceFake) GetLesson(in core.GetLessonInput) (core.GetLessonOutput, error) {
	service.called = "get"
	service.getIn = in
	return service.getOut, nil
}

func (service *lessonServiceFake) UpdateLesson(in core.UpdateLessonInput) (core.UpdateLessonOutput, error) {
	service.called = "update"
	service.updateIn = in
	return service.updateOut, nil
}

func (service *lessonServiceFake) DeleteLesson(in core.DeleteLessonInput) error {
	service.called = "delete"
	service.deleteIn = in
	return nil
}

func (service *lessonServiceFake) ReorderLessons(in core.ReorderLessonsInput) error {
	service.called = "reorder"
	service.reorderIn = in
	return nil
}

func (service *lessonServiceFake) AddLessonBlock(in core.AddLessonBlockInput) (core.AddLessonBlockOutput, error) {
	service.called = "add-block"
	service.addBlockIn = in
	if service.addBlockErr != nil {
		return core.AddLessonBlockOutput{}, service.addBlockErr
	}
	return service.addBlockOut, nil
}

func (service *lessonServiceFake) ListLessonBlocks(in core.ListLessonBlocksInput) (core.ListLessonBlocksOutput, error) {
	service.called = "list-blocks"
	service.listBlocksIn = in
	return service.listBlocksOut, nil
}

func (service *lessonServiceFake) GetLessonBlock(in core.GetLessonBlockInput) (core.GetLessonBlockOutput, error) {
	service.called = "get-block"
	service.getBlockIn = in
	return service.getBlockOut, nil
}

func (service *lessonServiceFake) UpdateLessonBlock(in core.UpdateLessonBlockInput) (core.UpdateLessonBlockOutput, error) {
	service.called = "update-block"
	service.updateBlockIn = in
	return service.updateBlockOut, nil
}

func (service *lessonServiceFake) RemoveLessonBlock(in core.RemoveLessonBlockInput) error {
	service.called = "remove-block"
	service.removeBlockIn = in
	return nil
}

func (service *lessonServiceFake) ReorderLessonBlocks(in core.ReorderLessonBlocksInput) error {
	service.called = "reorder-blocks"
	service.reorderBlocksIn = in
	return nil
}

type quizServiceFake struct {
	called string

	createIn           core.CreateQuizInput
	createOut          core.CreateQuizOutput
	listIn             core.ListQuizzesInput
	listOut            core.ListQuizzesOutput
	getIn              core.GetQuizInput
	getOut             core.GetQuizOutput
	updateIn           core.UpdateQuizInput
	updateOut          core.UpdateQuizOutput
	deleteIn           core.DeleteQuizInput
	deleteErr          error
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
	return service.createOut, nil
}

func (service *quizServiceFake) ListQuizzes(in core.ListQuizzesInput) (core.ListQuizzesOutput, error) {
	service.called = "list"
	service.listIn = in
	return service.listOut, nil
}

func (service *quizServiceFake) GetQuiz(in core.GetQuizInput) (core.GetQuizOutput, error) {
	service.called = "get"
	service.getIn = in
	return service.getOut, nil
}

func (service *quizServiceFake) UpdateQuiz(in core.UpdateQuizInput) (core.UpdateQuizOutput, error) {
	service.called = "update"
	service.updateIn = in
	return service.updateOut, nil
}

func (service *quizServiceFake) DeleteQuiz(in core.DeleteQuizInput) error {
	service.called = "delete"
	service.deleteIn = in
	return service.deleteErr
}

func (service *quizServiceFake) AddQuestion(in core.AddQuestionInput) (core.AddQuestionOutput, error) {
	service.called = "add-question"
	service.addQuestionIn = in
	return service.addQuestionOut, nil
}

func (service *quizServiceFake) ListQuestions(in core.ListQuestionsInput) (core.ListQuestionsOutput, error) {
	service.called = "list-questions"
	service.listQuestionsIn = in
	return service.listQuestionsOut, nil
}

func (service *quizServiceFake) GetQuestion(in core.GetQuestionInput) (core.GetQuestionOutput, error) {
	service.called = "get-question"
	service.getQuestionIn = in
	return service.getQuestionOut, nil
}

func (service *quizServiceFake) UpdateQuestion(in core.UpdateQuestionInput) (core.UpdateQuestionOutput, error) {
	service.called = "update-question"
	service.updateQuestionIn = in
	return service.updateQuestionOut, nil
}

func (service *quizServiceFake) RemoveQuestion(in core.RemoveQuestionInput) error {
	service.called = "remove-question"
	service.removeQuestionIn = in
	return nil
}

func (service *quizServiceFake) ReorderQuestions(in core.ReorderQuestionsInput) error {
	service.called = "reorder-questions"
	service.reorderQuestionsIn = in
	return nil
}

type practiceServiceFake struct {
	called string

	createIn           core.CreatePracticeInput
	createOut          core.CreatePracticeOutput
	listIn             core.ListPracticesInput
	listOut            core.ListPracticesOutput
	getIn              core.GetPracticeInput
	getOut             core.GetPracticeOutput
	updateIn           core.UpdatePracticeInput
	updateOut          core.UpdatePracticeOutput
	deleteIn           core.DeletePracticeInput
	deleteErr          error
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
	return service.createOut, nil
}

func (service *practiceServiceFake) ListPractices(in core.ListPracticesInput) (core.ListPracticesOutput, error) {
	service.called = "list"
	service.listIn = in
	return service.listOut, nil
}

func (service *practiceServiceFake) GetPractice(in core.GetPracticeInput) (core.GetPracticeOutput, error) {
	service.called = "get"
	service.getIn = in
	return service.getOut, nil
}

func (service *practiceServiceFake) UpdatePractice(in core.UpdatePracticeInput) (core.UpdatePracticeOutput, error) {
	service.called = "update"
	service.updateIn = in
	return service.updateOut, nil
}

func (service *practiceServiceFake) DeletePractice(in core.DeletePracticeInput) error {
	service.called = "delete"
	service.deleteIn = in
	return service.deleteErr
}

func (service *practiceServiceFake) AddTestCase(in core.AddTestCaseInput) (core.AddTestCaseOutput, error) {
	service.called = "add-testcase"
	service.addTestCaseIn = in
	return service.addTestCaseOut, nil
}

func (service *practiceServiceFake) ListTestCases(in core.ListTestCasesInput) (core.ListTestCasesOutput, error) {
	service.called = "list-testcases"
	service.listTestCasesIn = in
	return service.listTestCasesOut, nil
}

func (service *practiceServiceFake) GetTestCase(in core.GetTestCaseInput) (core.GetTestCaseOutput, error) {
	service.called = "get-testcase"
	service.getTestCaseIn = in
	return service.getTestCaseOut, nil
}

func (service *practiceServiceFake) UpdateTestCase(in core.UpdateTestCaseInput) (core.UpdateTestCaseOutput, error) {
	service.called = "update-testcase"
	service.updateTestCaseIn = in
	return service.updateTestCaseOut, nil
}

func (service *practiceServiceFake) RemoveTestCase(in core.RemoveTestCaseInput) error {
	service.called = "remove-testcase"
	service.removeTestCaseIn = in
	return nil
}

func (service *practiceServiceFake) ReorderTestCases(in core.ReorderTestCasesInput) error {
	service.called = "reorder-testcases"
	service.reorderTestCasesIn = in
	return nil
}

type testServiceFake struct {
	called string

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
	return service.createOut, nil
}

func (service *testServiceFake) ListTests(in core.ListTestsInput) (core.ListTestsOutput, error) {
	service.called = "list"
	service.listIn = in
	return service.listOut, nil
}

func (service *testServiceFake) GetTest(in core.GetTestInput) (core.GetTestOutput, error) {
	service.called = "get"
	service.getIn = in
	return service.getOut, nil
}

func (service *testServiceFake) UpdateTest(in core.UpdateTestInput) (core.UpdateTestOutput, error) {
	service.called = "update"
	service.updateIn = in
	return service.updateOut, nil
}

func (service *testServiceFake) DeleteTest(in core.DeleteTestInput) error {
	service.called = "delete"
	service.deleteIn = in
	return nil
}

func (service *testServiceFake) AddTestItem(in core.AddTestItemInput) (core.AddTestItemOutput, error) {
	service.called = "add-item"
	service.addItemIn = in
	return service.addItemOut, nil
}

func (service *testServiceFake) ListTestItems(in core.ListTestItemsInput) (core.ListTestItemsOutput, error) {
	service.called = "list-items"
	service.listItemsIn = in
	return service.listItemsOut, nil
}

func (service *testServiceFake) GetTestItem(in core.GetTestItemInput) (core.GetTestItemOutput, error) {
	service.called = "get-item"
	service.getItemIn = in
	return service.getItemOut, nil
}

func (service *testServiceFake) UpdateTestItem(in core.UpdateTestItemInput) (core.UpdateTestItemOutput, error) {
	service.called = "update-item"
	service.updateItemIn = in
	return service.updateItemOut, nil
}

func (service *testServiceFake) RemoveTestItem(in core.RemoveTestItemInput) error {
	service.called = "remove-item"
	service.removeItemIn = in
	return nil
}

func (service *testServiceFake) ReorderTestItems(in core.ReorderTestItemsInput) error {
	service.called = "reorder-items"
	service.reorderItemsIn = in
	return nil
}

func lessonViewFixture() core.LessonView {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	return core.LessonView{
		ID:        lessonIDValue,
		CourseID:  courseIDValue,
		Title:     "First Lesson",
		Order:     0,
		CreatedAt: now,
		UpdatedAt: now,
	}
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
