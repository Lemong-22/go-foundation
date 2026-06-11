package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
	"github.com/spf13/viper"
)

func TestCourseOutputRendererRendersTableJSONAndQuiet(t *testing.T) {
	course := courseViewFixture()

	t.Run("table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newCourseOutputRenderer(&output)

		err := renderer.RenderCourseList("table", []core.CourseView{course})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "ID") || !strings.Contains(got, "TITLE") || !strings.Contains(got, courseIDValue) {
			t.Fatalf("expected table output to include headers and course id, got %q", got)
		}
	})

	t.Run("json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newCourseOutputRenderer(&output)

		err := renderer.RenderCourseList("json", []core.CourseView{course})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var courses []core.CourseView
		if err := json.Unmarshal(output.Bytes(), &courses); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if len(courses) != 1 || courses[0].ID != courseIDValue {
			t.Fatalf("expected course id in json output, got %+v", courses)
		}
	})

	t.Run("quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newCourseOutputRenderer(&output)

		err := renderer.RenderCourseList("quiet", []core.CourseView{course})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != courseIDValue+"\n" {
			t.Fatalf("expected quiet id output, got %q", output.String())
		}
	})
}

func TestLessonOutputRendererRendersTableJSONAndQuiet(t *testing.T) {
	lesson := lessonViewFixture()

	t.Run("table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderLessonList("table", []core.LessonView{lesson})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "COURSE_ID") || !strings.Contains(got, "ORDER") || !strings.Contains(got, lessonIDValue) {
			t.Fatalf("expected table output to include headers and lesson id, got %q", got)
		}
	})

	t.Run("json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderLessonList("json", []core.LessonView{lesson})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var lessons []core.LessonView
		if err := json.Unmarshal(output.Bytes(), &lessons); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if len(lessons) != 1 || lessons[0].ID != lessonIDValue {
			t.Fatalf("expected lesson id in json output, got %+v", lessons)
		}

		var payload []map[string]any
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json object output, got %v", err)
		}
		if _, exists := payload[0]["Content"]; exists {
			t.Fatalf("expected lesson list json not to expose content, got %q", output.String())
		}
	})

	t.Run("quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderLessonList("quiet", []core.LessonView{lesson})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != lessonIDValue+"\n" {
			t.Fatalf("expected quiet id output, got %q", output.String())
		}
	})

	t.Run("detail table omits content", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderLesson("table", lesson)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if strings.Contains(got, "CONTENT") {
			t.Fatalf("expected lesson detail table not to expose content, got %q", got)
		}
		if !strings.Contains(got, "TITLE") || !strings.Contains(got, "First Lesson") {
			t.Fatalf("expected lesson detail table to include title, got %q", got)
		}
	})

	t.Run("detail json omits content", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderLesson("json", lesson)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json object output, got %v", err)
		}
		if _, exists := payload["Content"]; exists {
			t.Fatalf("expected lesson detail json not to expose content, got %q", output.String())
		}
	})
}

func TestBlockOutputRendererRendersTableJSONAndQuiet(t *testing.T) {
	block := blockViewFixture()

	t.Run("table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderBlockList("table", []core.BlockView{block})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "LESSON_ID") || !strings.Contains(got, "VIDEO_LOCATOR") || !strings.Contains(got, blockIDValue) {
			t.Fatalf("expected table output to include block headers and id, got %q", got)
		}
	})

	t.Run("json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderBlockList("json", []core.BlockView{block})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var blocks []core.BlockView
		if err := json.Unmarshal(output.Bytes(), &blocks); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if len(blocks) != 1 || blocks[0].ID != blockIDValue || blocks[0].VideoCaption != "Intro video" {
			t.Fatalf("expected block payload in json output, got %+v", blocks)
		}
	})

	t.Run("quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderBlockList("quiet", []core.BlockView{block})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != blockIDValue+"\n" {
			t.Fatalf("expected quiet block id output, got %q", output.String())
		}
	})

	t.Run("detail table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderBlock("table", block)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "KIND") || !strings.Contains(got, "youtube") || !strings.Contains(got, "Intro video") {
			t.Fatalf("expected block detail table to include payload fields, got %q", got)
		}
	})

	t.Run("detail json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderBlock("json", block)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload core.BlockView
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json object output, got %v", err)
		}
		if payload.ID != blockIDValue || payload.LessonID != lessonIDValue {
			t.Fatalf("expected block detail json to include ids, got %+v", payload)
		}
	})

	t.Run("detail quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newLessonOutputRenderer(&output)

		err := renderer.RenderBlock("quiet", block)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != blockIDValue+"\n" {
			t.Fatalf("expected quiet block id output, got %q", output.String())
		}
	})
}

func TestQuizOutputRendererRendersTableJSONAndQuiet(t *testing.T) {
	quiz := quizViewFixture()
	detail := quizDetailFixture()
	question := questionViewFixture()

	t.Run("quiz table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newQuizOutputRenderer(&output)

		err := renderer.RenderQuizList("table", []core.QuizView{quiz})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "PASS_THRESHOLD") || !strings.Contains(got, quizIDValue) {
			t.Fatalf("expected quiz table output, got %q", got)
		}
	})

	t.Run("quiz json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newQuizOutputRenderer(&output)

		err := renderer.RenderQuiz("json", detail)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload core.QuizDetailView
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if payload.ID != quizIDValue || len(payload.Questions) != 1 {
			t.Fatalf("expected quiz detail json, got %+v", payload)
		}
	})

	t.Run("quiz quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newQuizOutputRenderer(&output)

		err := renderer.RenderQuizList("quiet", []core.QuizView{quiz})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != quizIDValue+"\n" {
			t.Fatalf("expected quiet quiz id, got %q", output.String())
		}
	})

	t.Run("question table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newQuizOutputRenderer(&output)

		err := renderer.RenderQuestionList("table", []core.QuestionView{question})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "CORRECT_INDICES") || !strings.Contains(got, questionIDValue) {
			t.Fatalf("expected question table output, got %q", got)
		}
	})

	t.Run("question json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newQuizOutputRenderer(&output)

		err := renderer.RenderQuestion("json", question)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload core.QuestionView
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if payload.ID != questionIDValue || len(payload.CorrectIndices) != 1 {
			t.Fatalf("expected question detail json, got %+v", payload)
		}
	})

	t.Run("question quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newQuizOutputRenderer(&output)

		err := renderer.RenderQuestionList("quiet", []core.QuestionView{question})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != questionIDValue+"\n" {
			t.Fatalf("expected quiet question id, got %q", output.String())
		}
	})
}

func TestPracticeOutputRendererRendersTableJSONAndQuiet(t *testing.T) {
	practice := practiceViewFixture()
	detail := practiceDetailFixture()
	testCase := testCaseViewFixture()

	t.Run("practice table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newPracticeOutputRenderer(&output)

		err := renderer.RenderPracticeList("table", []core.PracticeView{practice})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "LANGUAGE") || !strings.Contains(got, practiceIDValue) {
			t.Fatalf("expected practice table output, got %q", got)
		}
	})

	t.Run("practice json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newPracticeOutputRenderer(&output)

		err := renderer.RenderPractice("json", detail)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload core.PracticeDetailView
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if payload.ID != practiceIDValue || len(payload.TestCases) != 1 {
			t.Fatalf("expected practice detail json, got %+v", payload)
		}
	})

	t.Run("practice quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newPracticeOutputRenderer(&output)

		err := renderer.RenderPracticeList("quiet", []core.PracticeView{practice})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != practiceIDValue+"\n" {
			t.Fatalf("expected quiet practice id, got %q", output.String())
		}
	})

	t.Run("test case table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newPracticeOutputRenderer(&output)

		err := renderer.RenderTestCaseList("table", []core.TestCaseView{testCase})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "EXPECTED_STDOUT") || !strings.Contains(got, testCaseIDValue) {
			t.Fatalf("expected test case table output, got %q", got)
		}
	})

	t.Run("test case json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newPracticeOutputRenderer(&output)

		err := renderer.RenderTestCase("json", testCase)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload core.TestCaseView
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if payload.ID != testCaseIDValue || payload.PracticeID != practiceIDValue {
			t.Fatalf("expected test case detail json, got %+v", payload)
		}
	})

	t.Run("test case quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newPracticeOutputRenderer(&output)

		err := renderer.RenderTestCaseList("quiet", []core.TestCaseView{testCase})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != testCaseIDValue+"\n" {
			t.Fatalf("expected quiet test case id, got %q", output.String())
		}
	})
}

func TestTestOutputRendererRendersTableJSONAndQuiet(t *testing.T) {
	testView := testViewFixture()
	detail := testDetailFixture()
	item := testItemViewFixture()

	t.Run("test table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newTestOutputRenderer(&output)

		err := renderer.RenderTestList("table", []core.TestView{testView})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "TIME_LIMIT_MINUTES") || !strings.Contains(got, testIDValue) {
			t.Fatalf("expected test table output, got %q", got)
		}
	})

	t.Run("test json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newTestOutputRenderer(&output)

		err := renderer.RenderTest("json", detail)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload core.TestDetailView
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if payload.ID != testIDValue || len(payload.Items) != 1 || payload.Solution == nil {
			t.Fatalf("expected test detail json, got %+v", payload)
		}
	})

	t.Run("test quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newTestOutputRenderer(&output)

		err := renderer.RenderTestList("quiet", []core.TestView{testView})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != testIDValue+"\n" {
			t.Fatalf("expected quiet test id, got %q", output.String())
		}
	})

	t.Run("item table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newTestOutputRenderer(&output)

		err := renderer.RenderTestItemList("table", []core.TestItemView{item})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "TEST_CASE_COUNT") || !strings.Contains(got, testItemIDValue) {
			t.Fatalf("expected item table output, got %q", got)
		}
	})

	t.Run("item json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newTestOutputRenderer(&output)

		err := renderer.RenderTestItem("json", item)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload core.TestItemView
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if payload.ID != testItemIDValue || payload.TestID != testIDValue {
			t.Fatalf("expected item detail json, got %+v", payload)
		}
	})

	t.Run("item quiet", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newTestOutputRenderer(&output)

		err := renderer.RenderTestItemList("quiet", []core.TestItemView{item})
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		if output.String() != testItemIDValue+"\n" {
			t.Fatalf("expected quiet item id, got %q", output.String())
		}
	})
}

func TestImportOutputRendererRendersPlanTableAndJSON(t *testing.T) {
	targetID := courseIDValue
	operation := importOperationFixture(t, domain.NoopOperation(), domain.CourseEntity(), "course:intro", &targetID)
	conflict := importConflictFixture(t, "course:advanced")
	plan := importPlanFixture(t, []domain.ImportOperation{operation}, []domain.ImportConflict{conflict})

	t.Run("table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newImportOutputRenderer(&output)

		err := renderer.RenderImportPlan("table", plan)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "OPERATIONS") || !strings.Contains(got, "course:intro") || !strings.Contains(got, "CONFLICTS") || !strings.Contains(got, "course:advanced") {
			t.Fatalf("expected plan table output, got %q", got)
		}
	})

	t.Run("json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newImportOutputRenderer(&output)

		err := renderer.RenderImportPlan("json", plan)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if payload["zip_hash"] != importZipHash {
			t.Fatalf("expected plan json payload, got %+v", payload)
		}
	})
}

func TestImportOutputRendererRendersApplyResultTableAndJSON(t *testing.T) {
	targetID := courseIDValue
	appliedOperation := importOperationFixture(t, domain.UpdateOperation(), domain.CourseEntity(), "course:intro", &targetID)
	skippedOperation := importOperationFixture(t, domain.SkipOperation(), domain.CourseEntity(), "course:skip", &targetID)
	failedOperation := importOperationFixture(t, domain.UpdateOperation(), domain.CourseEntity(), "course:failed", &targetID)
	applied, err := domain.NewAppliedOperation(appliedOperation, "updated course")
	if err != nil {
		t.Fatalf("expected applied fixture, got %v", err)
	}
	failed, err := domain.NewFailedOperation(failedOperation, errors.New("boom"))
	if err != nil {
		t.Fatalf("expected failed fixture, got %v", err)
	}
	result, err := domain.NewApplyResult([]domain.AppliedOperation{applied}, []domain.FailedOperation{failed}, []domain.ImportOperation{skippedOperation}, 1, 1)
	if err != nil {
		t.Fatalf("expected apply result fixture, got %v", err)
	}

	t.Run("table", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newImportOutputRenderer(&output)

		err := renderer.RenderApplyResult("table", result)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		got := output.String()
		if !strings.Contains(got, "AGGREGATES_SUCCEEDED") || !strings.Contains(got, "FAILED") || !strings.Contains(got, "course:failed") {
			t.Fatalf("expected apply table output, got %q", got)
		}
	})

	t.Run("json", func(t *testing.T) {
		var output bytes.Buffer
		renderer := newImportOutputRenderer(&output)

		err := renderer.RenderApplyResult("json", result)
		if err != nil {
			t.Fatalf("expected render to succeed, got %v", err)
		}

		var payload map[string]any
		if err := json.Unmarshal(output.Bytes(), &payload); err != nil {
			t.Fatalf("expected json output, got %v", err)
		}
		if payload["aggregates_failed"] != float64(1) {
			t.Fatalf("expected apply result json, got %+v", payload)
		}
	})
}

func TestOutputRendererRejectsUnknownFormat(t *testing.T) {
	var output bytes.Buffer
	renderer := newCourseOutputRenderer(&output)

	err := renderer.RenderCourse("xml", courseViewFixture())
	if !errors.Is(err, ErrUnsupportedOutputFormat) {
		t.Fatalf("expected unsupported output format error, got %v", err)
	}
}

func TestCourseCommandUsesDefaultRendererOutput(t *testing.T) {
	service := &courseServiceFake{createOut: core.CreateCourseOutput{ID: courseIDValue}}
	var output bytes.Buffer

	command := NewCourseCommand(CourseCommandOptions{Service: service, Config: viper.New()})
	command.SetArgs([]string{
		"create",
		"--title", "Intro to Go",
		"--slug", "intro-to-go",
		"--instructor-id", instructorIDValue,
	})
	command.SetOut(&output)
	command.SetErr(io.Discard)

	if err := command.Execute(); err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}
	if output.String() != courseIDValue+"\n" {
		t.Fatalf("expected default renderer to write created id, got %q", output.String())
	}
}
