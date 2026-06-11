package usecase

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/adapter/importsource"
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	importPlanZipHash = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	importBlockID     = "550e8400-e29b-41d4-a716-446655440070"
)

func TestPlanImportCreatesFullPlanWithPlaceholders(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 28, 16, 30, 0, 0, time.UTC)}
	service := newImportServiceForTest(
		fullParsedImportSource(),
		importSourceMetadata("1"),
		clock,
		newCourseRepositoryFake(),
		newLessonRepositoryFake(),
		newQuizRepositoryFake(),
		newPracticeRepositoryFake(),
		newTestRepositoryFake(),
	)

	out, err := service.PlanImport(core.PlanImportInput{ZipPath: "course.zip", InstructorID: instructorIDValue})
	if err != nil {
		t.Fatalf("expected import plan, got %v", err)
	}

	plan := out.Plan
	if plan.FormatVersion() != "1" || plan.ZipHash() != importPlanZipHash || !plan.GeneratedAt().Equal(clock.now) {
		t.Fatalf("expected plan metadata to come from source and clock")
	}
	if len(plan.Conflicts()) != 0 {
		t.Fatalf("expected no conflicts, got %d", len(plan.Conflicts()))
	}

	operations := plan.Operations()
	if len(operations) != 12 {
		t.Fatalf("expected 12 create operations, got %d", len(operations))
	}

	assertCreateOperation(t, importOperationByRef(t, operations, "course:intro-to-go"), domain.CourseEntity())

	quizPayload := decodeOperationPayload(t, importOperationByRef(t, operations, "quiz:foundations-quiz"))
	assertPayloadString(t, quizPayload, "course_id", "$course:intro-to-go")
	assertPayloadString(t, quizPayload, "import_local_id", "foundations-quiz")

	questionPayload := decodeOperationPayload(t, importOperationByRef(t, operations, "question:foundations-quiz:0"))
	assertPayloadString(t, questionPayload, "quiz_id", "$quiz:foundations-quiz")
	assertPayloadNumber(t, questionPayload, "position", 0)

	testCasePayload := decodeOperationPayload(t, importOperationByRef(t, operations, "test_case:fizzbuzz:0"))
	assertPayloadString(t, testCasePayload, "practice_id", "$practice:fizzbuzz")

	testItemPayload := decodeOperationPayload(t, importOperationByRef(t, operations, "test_item:final-test:0"))
	assertPayloadString(t, testItemPayload, "test_id", "$test:final-test")

	lessonPayload := decodeOperationPayload(t, importOperationByRef(t, operations, "lesson:Welcome"))
	assertPayloadString(t, lessonPayload, "course_id", "$course:intro-to-go")

	quizBlockPayload := decodeOperationPayload(t, importOperationByRef(t, operations, "block:Welcome:2"))
	assertPayloadString(t, quizBlockPayload, "lesson_id", "$lesson:Welcome")
	assertPayloadString(t, quizBlockPayload, "quiz_ref", "$quiz:foundations-quiz")

	practiceBlockPayload := decodeOperationPayload(t, importOperationByRef(t, operations, "block:Welcome:3"))
	assertPayloadString(t, practiceBlockPayload, "practice_ref", "$practice:fizzbuzz")
}

func TestPlanImportEmitsNoopsForExactMatches(t *testing.T) {
	courses := newCourseRepositoryFake()
	lessons := newLessonRepositoryFake()
	quizzes := newQuizRepositoryFake()
	practices := newPracticeRepositoryFake()
	tests := newTestRepositoryFake()
	storeExistingImportAggregates(t, courses, lessons, quizzes, practices, tests)

	service := newImportServiceForTest(
		parsedImportSourceWithTextLesson("Intro to Go", "Read this"),
		importSourceMetadata("1"),
		fixedClock{now: time.Date(2026, 5, 28, 17, 0, 0, 0, time.UTC)},
		courses,
		lessons,
		quizzes,
		practices,
		tests,
	)

	out, err := service.PlanImport(core.PlanImportInput{ZipPath: "course.zip", InstructorID: instructorIDValue})
	if err != nil {
		t.Fatalf("expected import plan, got %v", err)
	}
	if len(out.Plan.Conflicts()) != 0 {
		t.Fatalf("expected no conflicts, got %d", len(out.Plan.Conflicts()))
	}

	for _, ref := range []string{
		"course:intro-to-go",
		"quiz:foundations-quiz",
		"question:foundations-quiz:0",
		"practice:fizzbuzz",
		"test_case:fizzbuzz:0",
		"test:final-test",
		"test_item:final-test:0",
		"lesson:Welcome",
		"block:Welcome:0",
	} {
		operation := importOperationByRef(t, out.Plan.Operations(), ref)
		if !operation.Kind().IsNoop() {
			t.Fatalf("expected %s to be noop, got %s", ref, operation.Kind().String())
		}
		if operation.TargetID() == nil {
			t.Fatalf("expected %s noop to include target id", ref)
		}
	}
}

func TestPlanImportEmitsContentDiffConflicts(t *testing.T) {
	courses := newCourseRepositoryFake()
	lessons := newLessonRepositoryFake()
	quizzes := newQuizRepositoryFake()
	practices := newPracticeRepositoryFake()
	tests := newTestRepositoryFake()
	storeExistingImportAggregates(t, courses, lessons, quizzes, practices, tests)

	parsed := parsedImportSourceWithTextLesson("Renamed Course", "Changed text")
	parsed.Quizzes[0].Questions[0].Prompt = "Pick a different answer"
	parsed.Practices[0].TestCases[0].ExpectedStdout = "different stdout"
	parsed.Tests[0].Items[0].Prompt = "Pick a different item"

	service := newImportServiceForTest(
		parsed,
		importSourceMetadata("1"),
		fixedClock{now: time.Date(2026, 5, 28, 17, 30, 0, 0, time.UTC)},
		courses,
		lessons,
		quizzes,
		practices,
		tests,
	)

	out, err := service.PlanImport(core.PlanImportInput{ZipPath: "course.zip", InstructorID: instructorIDValue})
	if err != nil {
		t.Fatalf("expected import plan with conflicts, got %v", err)
	}

	conflicts := out.Plan.Conflicts()
	if len(conflicts) != 5 {
		t.Fatalf("expected 5 conflicts, got %d", len(conflicts))
	}

	courseConflict := importConflictByRef(t, conflicts, "course:intro-to-go")
	if !courseConflict.Reason().IsSlugCollision() || !courseConflict.Recommended().IsUpdate() {
		t.Fatalf("expected course slug collision with update recommendation")
	}

	for _, ref := range []string{
		"question:foundations-quiz:0",
		"test_case:fizzbuzz:0",
		"test_item:final-test:0",
		"block:Welcome:0",
	} {
		conflict := importConflictByRef(t, conflicts, ref)
		if !conflict.Reason().IsPositionCollision() || !conflict.Recommended().IsUpdate() {
			t.Fatalf("expected %s position collision with update recommendation", ref)
		}
	}
}

func TestPlanImportResolvesExistingAndPlaceholderLessonRefs(t *testing.T) {
	courses := newCourseRepositoryFake()
	lessons := newLessonRepositoryFake()
	quizzes := newQuizRepositoryFake()
	practices := newPracticeRepositoryFake()
	tests := newTestRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro to Go", "intro-to-go", domain.Draft()))
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Foundations Quiz", domain.DefaultPassThreshold().Float64()))

	parsed := core.ParsedImportSource{
		FormatVersion: "1",
		Course:        parsedCourse("Intro to Go"),
		Quizzes:       []core.ParsedQuiz{parsedQuiz("Pick one")},
		Practices:     []core.ParsedPractice{parsedPractice("stdout")},
		Lessons: []core.ParsedLesson{{
			Title: "Welcome",
			Blocks: []core.ParsedLessonBlock{
				{Kind: domain.QuizKind().String(), QuizRef: "foundations-quiz"},
				{Kind: domain.PracticeKind().String(), PracticeRef: "fizzbuzz"},
			},
		}},
	}
	service := newImportServiceForTest(
		parsed,
		importSourceMetadata("1"),
		fixedClock{now: time.Date(2026, 5, 28, 18, 0, 0, 0, time.UTC)},
		courses,
		lessons,
		quizzes,
		practices,
		tests,
	)

	out, err := service.PlanImport(core.PlanImportInput{ZipPath: "course.zip", InstructorID: instructorIDValue})
	if err != nil {
		t.Fatalf("expected import plan, got %v", err)
	}

	quizBlockPayload := decodeOperationPayload(t, importOperationByRef(t, out.Plan.Operations(), "block:Welcome:0"))
	assertPayloadString(t, quizBlockPayload, "quiz_ref", quizIDValue)

	practiceBlockPayload := decodeOperationPayload(t, importOperationByRef(t, out.Plan.Operations(), "block:Welcome:1"))
	assertPayloadString(t, practiceBlockPayload, "practice_ref", "$practice:fizzbuzz")
}

func TestPlanImportRejectsUnknownLessonRefs(t *testing.T) {
	parsed := core.ParsedImportSource{
		FormatVersion: "1",
		Course:        parsedCourse("Intro to Go"),
		Lessons: []core.ParsedLesson{{
			Title:  "Welcome",
			Blocks: []core.ParsedLessonBlock{{Kind: domain.QuizKind().String(), QuizRef: "missing-quiz"}},
		}},
	}
	service := newImportServiceForTest(
		parsed,
		importSourceMetadata("1"),
		fixedClock{now: time.Date(2026, 5, 28, 18, 30, 0, 0, time.UTC)},
		newCourseRepositoryFake(),
		newLessonRepositoryFake(),
		newQuizRepositoryFake(),
		newPracticeRepositoryFake(),
		newTestRepositoryFake(),
	)

	if _, err := service.PlanImport(core.PlanImportInput{ZipPath: "course.zip", InstructorID: instructorIDValue}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestPlanImportRejectsUnsupportedFormat(t *testing.T) {
	service := newImportServiceForTest(
		core.ParsedImportSource{},
		importSourceMetadata("2"),
		fixedClock{now: time.Date(2026, 5, 28, 19, 0, 0, 0, time.UTC)},
		newCourseRepositoryFake(),
		newLessonRepositoryFake(),
		newQuizRepositoryFake(),
		newPracticeRepositoryFake(),
		newTestRepositoryFake(),
	)

	if _, err := service.PlanImport(core.PlanImportInput{ZipPath: "course.zip", InstructorID: instructorIDValue}); !errors.Is(err, domain.ErrUnsupportedImportFormat) {
		t.Fatalf("expected unsupported format error, got %v", err)
	}
}

func TestApplyPlanRejectsResolvedPlanHashMismatch(t *testing.T) {
	plan, err := domain.NewImportPlan(
		"1",
		"dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
		time.Date(2026, 5, 28, 19, 30, 0, 0, time.UTC),
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("expected plan fixture, got %v", err)
	}
	planJSON, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("expected plan JSON, got %v", err)
	}

	service := newImportApplyServiceForTest(
		parsedCourseOnly("Intro to Go"),
		importSourceMetadata("1"),
		fixedClock{now: time.Date(2026, 5, 28, 19, 30, 0, 0, time.UTC)},
		newCourseRepositoryFake(),
		newLessonRepositoryFake(),
		newQuizRepositoryFake(),
		newPracticeRepositoryFake(),
		newTestRepositoryFake(),
	)

	_, err = service.ApplyPlan(core.ApplyPlanInput{
		ZipPath:          "course.zip",
		InstructorID:     instructorIDValue,
		ResolvedPlanJSON: planJSON,
	})
	if !errors.Is(err, domain.ErrImportPlanHashMismatch) {
		t.Fatalf("expected hash mismatch, got %v", err)
	}
}

func TestApplyPlanRejectsUnresolvedConflictsWithFailStrategy(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro to Go", "intro-to-go", domain.Draft()))
	service := newImportApplyServiceForTest(
		parsedCourseOnly("Renamed Course"),
		importSourceMetadata("1"),
		fixedClock{now: time.Date(2026, 5, 28, 20, 0, 0, 0, time.UTC)},
		courses,
		newLessonRepositoryFake(),
		newQuizRepositoryFake(),
		newPracticeRepositoryFake(),
		newTestRepositoryFake(),
	)

	_, err := service.ApplyPlan(core.ApplyPlanInput{ZipPath: "course.zip", InstructorID: instructorIDValue})
	if !errors.Is(err, domain.ErrUnresolvedImportConflicts) {
		t.Fatalf("expected unresolved conflicts error, got %v", err)
	}
}

func TestApplyPlanAppliesSkipAndUpdateConflictStrategies(t *testing.T) {
	t.Run("skip", func(t *testing.T) {
		courses := newCourseRepositoryFake()
		courses.store(mustCourse(t, courseIDValue, "Intro to Go", "intro-to-go", domain.Draft()))
		service := newImportApplyServiceForTest(
			parsedCourseOnly("Renamed Course"),
			importSourceMetadata("1"),
			fixedClock{now: time.Date(2026, 5, 28, 20, 30, 0, 0, time.UTC)},
			courses,
			newLessonRepositoryFake(),
			newQuizRepositoryFake(),
			newPracticeRepositoryFake(),
			newTestRepositoryFake(),
		)

		out, err := service.ApplyPlan(core.ApplyPlanInput{
			ZipPath:          "course.zip",
			InstructorID:     instructorIDValue,
			ConflictStrategy: domain.SkipConflicts().String(),
		})
		if err != nil {
			t.Fatalf("expected skip strategy to succeed, got %v", err)
		}

		if len(out.Result.Skipped()) != 1 || !out.Result.Skipped()[0].Kind().IsSkip() {
			t.Fatalf("expected one skipped conflict operation")
		}
		if out.Result.AggregatesSucceeded() != 1 || out.Result.AggregatesFailed() != 0 {
			t.Fatalf("expected one successful aggregate, got success=%d failed=%d", out.Result.AggregatesSucceeded(), out.Result.AggregatesFailed())
		}
		course, err := courses.FindByID(mustCourseID(courseIDValue))
		if err != nil {
			t.Fatalf("expected course fixture, got %v", err)
		}
		if course.Title() != "Intro to Go" {
			t.Fatalf("expected skipped course to stay unchanged, got %q", course.Title())
		}
	})

	t.Run("update", func(t *testing.T) {
		courses := newCourseRepositoryFake()
		courses.store(mustCourse(t, courseIDValue, "Intro to Go", "intro-to-go", domain.Draft()))
		service := newImportApplyServiceForTest(
			parsedCourseOnly("Renamed Course"),
			importSourceMetadata("1"),
			fixedClock{now: time.Date(2026, 5, 28, 21, 0, 0, 0, time.UTC)},
			courses,
			newLessonRepositoryFake(),
			newQuizRepositoryFake(),
			newPracticeRepositoryFake(),
			newTestRepositoryFake(),
		)

		out, err := service.ApplyPlan(core.ApplyPlanInput{
			ZipPath:          "course.zip",
			InstructorID:     instructorIDValue,
			ConflictStrategy: domain.UpdateConflicts().String(),
		})
		if err != nil {
			t.Fatalf("expected update strategy to succeed, got %v", err)
		}

		if len(out.Result.Applied()) != 1 || !out.Result.Applied()[0].Operation().Kind().IsUpdate() {
			t.Fatalf("expected one applied update operation")
		}
		course, err := courses.FindByID(mustCourseID(courseIDValue))
		if err != nil {
			t.Fatalf("expected course fixture, got %v", err)
		}
		if course.Title() != "Renamed Course" {
			t.Fatalf("expected course title to update, got %q", course.Title())
		}
	})
}

func TestApplyPlanRecordsPartialAggregateFailure(t *testing.T) {
	parsed := parsedCourseWithQuizLesson()
	parsed.Lessons = nil
	parsed.Quizzes[0].Questions[0].Type = "bad"

	service := newImportApplyServiceForTest(
		parsed,
		importSourceMetadata("1"),
		fixedClock{now: time.Date(2026, 5, 28, 21, 30, 0, 0, time.UTC)},
		newCourseRepositoryFake(),
		newLessonRepositoryFake(),
		newQuizRepositoryFake(),
		newPracticeRepositoryFake(),
		newTestRepositoryFake(),
	)

	out, err := service.ApplyPlan(core.ApplyPlanInput{ZipPath: "course.zip", InstructorID: instructorIDValue})
	if err != nil {
		t.Fatalf("expected partial failure result, got %v", err)
	}

	if out.Result.AggregatesSucceeded() != 1 || out.Result.AggregatesFailed() != 1 {
		t.Fatalf("expected one successful and one failed aggregate, got success=%d failed=%d", out.Result.AggregatesSucceeded(), out.Result.AggregatesFailed())
	}
	if len(out.Result.Failed()) != 1 {
		t.Fatalf("expected one failed operation, got %d", len(out.Result.Failed()))
	}
	if out.Result.Failed()[0].Operation().EntityRef() != "question:foundations-quiz:0" {
		t.Fatalf("expected failed question operation, got %s", out.Result.Failed()[0].Operation().EntityRef())
	}
}

func TestApplyPlanCreatesMultiAggregateImportAndResolvesPlaceholders(t *testing.T) {
	courses := newCourseRepositoryFake()
	lessons := newLessonRepositoryFake()
	quizzes := newQuizRepositoryFake()
	practices := newPracticeRepositoryFake()
	tests := newTestRepositoryFake()
	service := newImportApplyServiceForTest(
		parsedCourseWithQuizLesson(),
		importSourceMetadata("1"),
		fixedClock{now: time.Date(2026, 5, 28, 22, 0, 0, 0, time.UTC)},
		courses,
		lessons,
		quizzes,
		practices,
		tests,
	)

	out, err := service.ApplyPlan(core.ApplyPlanInput{ZipPath: "course.zip", InstructorID: instructorIDValue})
	if err != nil {
		t.Fatalf("expected apply to succeed, got %v", err)
	}

	if len(out.Result.Applied()) != 5 || len(out.Result.Failed()) != 0 || len(out.Result.Skipped()) != 0 {
		t.Fatalf("expected five applied operations, got applied=%d failed=%d skipped=%d", len(out.Result.Applied()), len(out.Result.Failed()), len(out.Result.Skipped()))
	}
	if out.Result.AggregatesSucceeded() != 3 || out.Result.AggregatesFailed() != 0 {
		t.Fatalf("expected three successful aggregates, got success=%d failed=%d", out.Result.AggregatesSucceeded(), out.Result.AggregatesFailed())
	}

	lesson, err := lessons.FindByID(mustLessonID(lessonIDValue))
	if err != nil {
		t.Fatalf("expected created lesson, got %v", err)
	}
	blocks := lesson.Blocks()
	if len(blocks) != 1 {
		t.Fatalf("expected one lesson block, got %d", len(blocks))
	}
	body, ok := blocks[0].Body().(domain.QuizBody)
	if !ok {
		t.Fatalf("expected quiz block, got %T", blocks[0].Body())
	}
	if body.QuizRef.String() != quizIDValue {
		t.Fatalf("expected quiz placeholder to resolve to %s, got %s", quizIDValue, body.QuizRef.String())
	}
}

func TestImportWorkflowWithFullZipFixturePlansAndAppliesAllAggregates(t *testing.T) {
	courses := newCourseRepositoryFake()
	lessons := newLessonRepositoryFake()
	quizzes := newQuizRepositoryFake()
	practices := newPracticeRepositoryFake()
	tests := newTestRepositoryFake()
	service := newZipImportApplyServiceForTest(
		fixedClock{now: time.Date(2026, 5, 29, 1, 30, 0, 0, time.UTC)},
		courses,
		lessons,
		quizzes,
		practices,
		tests,
	)
	zipPath := writeImportZipFixture(t, fullZipFixtureEntries())

	planOut, err := service.PlanImport(core.PlanImportInput{ZipPath: zipPath, InstructorID: instructorIDValue})
	if err != nil {
		t.Fatalf("expected fixture import plan, got %v", err)
	}
	if planOut.Plan.FormatVersion() != "1" || planOut.Plan.ZipHash() == "" {
		t.Fatalf("expected fixture plan metadata, got version=%q hash=%q", planOut.Plan.FormatVersion(), planOut.Plan.ZipHash())
	}
	if len(planOut.Plan.Conflicts()) != 0 {
		t.Fatalf("expected no fixture conflicts, got %d", len(planOut.Plan.Conflicts()))
	}
	if len(planOut.Plan.Operations()) != 13 {
		t.Fatalf("expected full fixture to produce 13 operations, got %d", len(planOut.Plan.Operations()))
	}

	applyOut, err := service.ApplyPlan(core.ApplyPlanInput{ZipPath: zipPath, InstructorID: instructorIDValue})
	if err != nil {
		t.Fatalf("expected fixture import apply to succeed, got %v", err)
	}
	if len(applyOut.Result.Failed()) != 0 || len(applyOut.Result.Skipped()) != 0 {
		t.Fatalf("expected all fixture operations applied, got failed=%d skipped=%d", len(applyOut.Result.Failed()), len(applyOut.Result.Skipped()))
	}
	if applyOut.Result.AggregatesSucceeded() != 5 || applyOut.Result.AggregatesFailed() != 0 {
		t.Fatalf("expected five successful aggregates, got success=%d failed=%d", applyOut.Result.AggregatesSucceeded(), applyOut.Result.AggregatesFailed())
	}

	course, err := courses.FindBySlug(mustSlug("intro-to-go"))
	if err != nil {
		t.Fatalf("expected created course, got %v", err)
	}
	if course.Title() != "Intro to Go" || course.Status() != domain.Draft() {
		t.Fatalf("expected created draft course, got title=%q status=%q", course.Title(), course.Status().String())
	}

	quizList, err := quizzes.FindByCourse(course.ID())
	if err != nil {
		t.Fatalf("expected created quiz, got %v", err)
	}
	if len(quizList) != 1 || len(quizList[0].Questions()) != 1 {
		t.Fatalf("expected one quiz with one question, got %+v", quizList)
	}

	practiceList, err := practices.FindByCourse(course.ID())
	if err != nil {
		t.Fatalf("expected created practice, got %v", err)
	}
	if len(practiceList) != 1 || len(practiceList[0].TestCases()) != 1 {
		t.Fatalf("expected one practice with one test case, got %+v", practiceList)
	}

	testList, err := tests.FindByCourse(course.ID())
	if err != nil {
		t.Fatalf("expected created test, got %v", err)
	}
	if len(testList) != 1 || len(testList[0].Items()) != 2 || testList[0].Solution() == nil {
		t.Fatalf("expected one test with two items and a solution, got %+v", testList)
	}

	lessonList, err := lessons.FindByCourse(course.ID())
	if err != nil {
		t.Fatalf("expected created lesson, got %v", err)
	}
	if len(lessonList) != 1 || len(lessonList[0].Blocks()) != 4 {
		t.Fatalf("expected one lesson with four blocks, got %+v", lessonList)
	}
	blocks := lessonList[0].Blocks()
	if _, ok := blocks[0].Body().(domain.TextBody); !ok {
		t.Fatalf("expected text block, got %T", blocks[0].Body())
	}
	if _, ok := blocks[1].Body().(domain.VideoBody); !ok {
		t.Fatalf("expected video block, got %T", blocks[1].Body())
	}
	if quizBody, ok := blocks[2].Body().(domain.QuizBody); !ok || quizBody.QuizRef != quizList[0].ID() {
		t.Fatalf("expected quiz block to resolve imported quiz id, got %#v", blocks[2].Body())
	}
	if practiceBody, ok := blocks[3].Body().(domain.PracticeBody); !ok || practiceBody.PracticeRef != practiceList[0].ID() {
		t.Fatalf("expected practice block to resolve imported practice id, got %#v", blocks[3].Body())
	}
}

func TestPlanImportWithConflictingZipFixtureEmitsSlugCollision(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro to Go", "intro-to-go", domain.Draft()))
	service := newZipImportPlanServiceForTest(
		fixedClock{now: time.Date(2026, 5, 29, 2, 0, 0, 0, time.UTC)},
		courses,
		newLessonRepositoryFake(),
		newQuizRepositoryFake(),
		newPracticeRepositoryFake(),
		newTestRepositoryFake(),
	)

	out, err := service.PlanImport(core.PlanImportInput{
		ZipPath:      writeImportZipFixture(t, conflictingZipFixtureEntries()),
		InstructorID: instructorIDValue,
	})
	if err != nil {
		t.Fatalf("expected conflicting fixture plan, got %v", err)
	}

	conflict := importConflictByRef(t, out.Plan.Conflicts(), "course:intro-to-go")
	if !conflict.Reason().IsSlugCollision() || !conflict.Recommended().IsUpdate() || len(conflict.Candidates()) != 1 {
		t.Fatalf("expected course slug collision with update recommendation, got %+v", conflict)
	}
}

func TestPlanImportWithMalformedZipFixtureReturnsParseError(t *testing.T) {
	service := newZipImportPlanServiceForTest(
		fixedClock{now: time.Date(2026, 5, 29, 2, 30, 0, 0, time.UTC)},
		newCourseRepositoryFake(),
		newLessonRepositoryFake(),
		newQuizRepositoryFake(),
		newPracticeRepositoryFake(),
		newTestRepositoryFake(),
	)

	_, err := service.PlanImport(core.PlanImportInput{
		ZipPath:      writeImportZipFixture(t, malformedZipFixtureEntries()),
		InstructorID: instructorIDValue,
	})
	if !errors.Is(err, domain.ErrImportSourceParse) {
		t.Fatalf("expected malformed fixture parse error, got %v", err)
	}
}

func TestApplyPlanRejectsResolvedPlanFromDifferentZipFixture(t *testing.T) {
	service := newZipImportApplyServiceForTest(
		fixedClock{now: time.Date(2026, 5, 29, 3, 0, 0, 0, time.UTC)},
		newCourseRepositoryFake(),
		newLessonRepositoryFake(),
		newQuizRepositoryFake(),
		newPracticeRepositoryFake(),
		newTestRepositoryFake(),
	)
	minimalZip := writeImportZipFixture(t, minimalZipFixtureEntries())
	fullZip := writeImportZipFixture(t, fullZipFixtureEntries())

	planOut, err := service.PlanImport(core.PlanImportInput{ZipPath: minimalZip, InstructorID: instructorIDValue})
	if err != nil {
		t.Fatalf("expected minimal fixture plan, got %v", err)
	}
	planJSON, err := json.Marshal(planOut.Plan)
	if err != nil {
		t.Fatalf("expected plan json, got %v", err)
	}

	_, err = service.ApplyPlan(core.ApplyPlanInput{
		ZipPath:          fullZip,
		InstructorID:     instructorIDValue,
		ResolvedPlanJSON: planJSON,
	})
	if !errors.Is(err, domain.ErrImportPlanHashMismatch) {
		t.Fatalf("expected hash mismatch for different zip fixture, got %v", err)
	}
}

type importSourceFake struct {
	parsed   core.ParsedImportSource
	metadata core.ImportSourceMetadata
	err      error
}

func (source importSourceFake) Open(string) (core.ParsedImportSource, core.ImportSourceMetadata, error) {
	if source.err != nil {
		return core.ParsedImportSource{}, core.ImportSourceMetadata{}, source.err
	}

	return source.parsed, source.metadata, nil
}

func newImportServiceForTest(
	parsed core.ParsedImportSource,
	metadata core.ImportSourceMetadata,
	clock fixedClock,
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	quizzes *quizRepositoryFake,
	practices *practiceRepositoryFake,
	tests *testRepositoryFake,
) *ImportService {
	return NewImportService(
		importSourceFake{parsed: parsed, metadata: metadata},
		clock,
		courses,
		lessons,
		quizzes,
		practices,
		tests,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func newImportApplyServiceForTest(
	parsed core.ParsedImportSource,
	metadata core.ImportSourceMetadata,
	clock fixedClock,
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	quizzes *quizRepositoryFake,
	practices *practiceRepositoryFake,
	tests *testRepositoryFake,
) *ImportService {
	ids := fixedIDGenerator{
		courseID:   mustCourseID(courseIDValue),
		lessonID:   mustLessonID(lessonIDValue),
		blockID:    mustBlockID(importBlockID),
		quizID:     mustQuizID(quizIDValue),
		questionID: mustQuestionID(questionIDValue),
		practiceID: mustPracticeID(practiceIDValue),
		testCaseID: mustTestCaseID(testCaseIDValue),
		testID:     mustTestID(testIDValue),
		testItemID: mustTestItemID(testItemIDValue),
	}
	courseService := NewCourseService(courses, lessons, quizzes, ids, clock, practices, tests)
	lessonService := NewLessonService(courses, lessons, quizzes, ids, clock, practices)
	quizService := NewQuizService(courses, lessons, quizzes, ids, clock)
	practiceService := NewPracticeService(courses, lessons, practices, ids, clock)
	testService := NewTestService(courses, tests, ids, clock)

	return NewImportService(
		importSourceFake{parsed: parsed, metadata: metadata},
		clock,
		courses,
		lessons,
		quizzes,
		practices,
		tests,
		courseService,
		lessonService,
		quizService,
		practiceService,
		testService,
	)
}

func importSourceMetadata(formatVersion string) core.ImportSourceMetadata {
	return core.ImportSourceMetadata{
		ZipHash:       importPlanZipHash,
		FormatVersion: formatVersion,
	}
}

func newZipImportPlanServiceForTest(
	clock fixedClock,
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	quizzes *quizRepositoryFake,
	practices *practiceRepositoryFake,
	tests *testRepositoryFake,
) *ImportService {
	return NewImportService(
		importsource.NewZipImportSource(),
		clock,
		courses,
		lessons,
		quizzes,
		practices,
		tests,
		nil,
		nil,
		nil,
		nil,
		nil,
	)
}

func newZipImportApplyServiceForTest(
	clock fixedClock,
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	quizzes *quizRepositoryFake,
	practices *practiceRepositoryFake,
	tests *testRepositoryFake,
) *ImportService {
	ids := &sequentialIDGenerator{}
	courseService := NewCourseService(courses, lessons, quizzes, ids, clock, practices, tests)
	lessonService := NewLessonService(courses, lessons, quizzes, ids, clock, practices)
	quizService := NewQuizService(courses, lessons, quizzes, ids, clock)
	practiceService := NewPracticeService(courses, lessons, practices, ids, clock)
	testService := NewTestService(courses, tests, ids, clock)

	return NewImportService(
		importsource.NewZipImportSource(),
		clock,
		courses,
		lessons,
		quizzes,
		practices,
		tests,
		courseService,
		lessonService,
		quizService,
		practiceService,
		testService,
	)
}

type sequentialIDGenerator struct {
	next int
}

func (generator *sequentialIDGenerator) NewCourseID() domain.CourseID {
	return mustCourseID(generator.nextUUID())
}

func (generator *sequentialIDGenerator) NewLessonID() domain.LessonID {
	return mustLessonID(generator.nextUUID())
}

func (generator *sequentialIDGenerator) NewBlockID() domain.BlockID {
	return mustBlockID(generator.nextUUID())
}

func (generator *sequentialIDGenerator) NewQuizID() domain.QuizID {
	return mustQuizID(generator.nextUUID())
}

func (generator *sequentialIDGenerator) NewQuestionID() domain.QuestionID {
	return mustQuestionID(generator.nextUUID())
}

func (generator *sequentialIDGenerator) NewPracticeID() domain.PracticeID {
	return mustPracticeID(generator.nextUUID())
}

func (generator *sequentialIDGenerator) NewTestCaseID() domain.TestCaseID {
	return mustTestCaseID(generator.nextUUID())
}

func (generator *sequentialIDGenerator) NewTestID() domain.TestID {
	return mustTestID(generator.nextUUID())
}

func (generator *sequentialIDGenerator) NewTestItemID() domain.TestItemID {
	return mustTestItemID(generator.nextUUID())
}

func (generator *sequentialIDGenerator) nextUUID() string {
	generator.next++
	return fmt.Sprintf("550e8400-e29b-41d4-a716-%012d", generator.next)
}

type zipFixtureEntry struct {
	name string
	body string
}

func writeImportZipFixture(t *testing.T, entries []zipFixtureEntry) string {
	t.Helper()

	zipPath := filepath.Join(t.TempDir(), "course.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create fixture zip: %v", err)
	}

	writer := zip.NewWriter(file)
	for _, entry := range entries {
		entryWriter, err := writer.Create(entry.name)
		if err != nil {
			t.Fatalf("create fixture zip entry %s: %v", entry.name, err)
		}
		if _, err := entryWriter.Write([]byte(entry.body)); err != nil {
			t.Fatalf("write fixture zip entry %s: %v", entry.name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close fixture zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close fixture zip file: %v", err)
	}

	return zipPath
}

func minimalZipFixtureEntries() []zipFixtureEntry {
	return []zipFixtureEntry{
		{name: "format_version.txt", body: "1\n"},
		{name: "course.yaml", body: `
title: Intro to Go
slug: intro-to-go
description: Learn Go
status: draft
`},
	}
}

func conflictingZipFixtureEntries() []zipFixtureEntry {
	entries := minimalZipFixtureEntries()
	entries[1].body = `
title: Renamed Go Course
slug: intro-to-go
description: Different content
status: draft
`
	return entries
}

func malformedZipFixtureEntries() []zipFixtureEntry {
	return []zipFixtureEntry{
		{name: "format_version.txt", body: "1\n"},
		{name: "course.yaml", body: "title: ["},
	}
}

func fullZipFixtureEntries() []zipFixtureEntry {
	return []zipFixtureEntry{
		{name: "format_version.txt", body: "1\n"},
		{name: "course.yaml", body: `
title: Intro to Go
slug: intro-to-go
description: Learn Go end to end
status: draft
`},
		{name: "quizzes/foundations-quiz.yaml", body: `
slug: foundations-quiz
title: Foundations Quiz
pass_threshold: 0.7
questions:
  - type: single
    prompt: Pick one
    options:
      - A
      - B
    correct_indices:
      - 0
    explanation: Because A.
    position: 0
`},
		{name: "practices/fizzbuzz.yaml", body: `
slug: fizzbuzz
title: FizzBuzz
language: golang
prompt: Print Fizz for multiples of three.
starter_code: package main
solution: package main
test_cases:
  - stdin: "3"
    expected_stdout: Fizz
    name: multiple of three
    position: 0
`},
		{name: "tests/final-test.yaml", body: `
slug: final-test
title: Final Test
time_limit_minutes: 30
pass_threshold: 0.8
solution:
  zip_provider: url
  zip_locator: https://example.com/solution.zip
  video_provider: youtube
  video_locator: dQw4w9WgXcQ
  video_caption: Walkthrough
items:
  - kind: choice
    prompt: Pick one
    choice_type: single
    options:
      - A
      - B
    correct_indices:
      - 0
    explanation: Because A.
    position: 0
  - kind: coding
    coding_prompt: Write FizzBuzz
    language: golang
    starter_code: package main
    solution: package main
    test_cases:
      - stdin: "3"
        expected_stdout: Fizz
        name: multiple of three
    position: 1
`},
		{name: "lessons/01-foundations.md", body: `---
title: Welcome
order: 0
blocks:
  - kind: text
    markdown: Read this
    position: 0
  - kind: video
    video_provider: youtube
    video_locator: dQw4w9WgXcQ
    video_caption: Setup
    position: 1
  - kind: quiz
    quiz_ref: foundations-quiz
    position: 2
  - kind: practice
    practice_ref: fizzbuzz
    position: 3
---
Lesson body.
`},
	}
}

func parsedCourseOnly(title string) core.ParsedImportSource {
	return core.ParsedImportSource{
		FormatVersion: "1",
		Course:        parsedCourse(title),
	}
}

func parsedCourseWithQuizLesson() core.ParsedImportSource {
	return core.ParsedImportSource{
		FormatVersion: "1",
		Course:        parsedCourse("Intro to Go"),
		Quizzes:       []core.ParsedQuiz{parsedQuiz("Pick one")},
		Lessons: []core.ParsedLesson{{
			Title: "Welcome",
			Blocks: []core.ParsedLessonBlock{{
				Kind:    domain.QuizKind().String(),
				QuizRef: "foundations-quiz",
			}},
		}},
	}
}

func fullParsedImportSource() core.ParsedImportSource {
	parsed := parsedImportSourceWithTextLesson("Intro to Go", "Read this")
	parsed.Lessons[0].Blocks = append(parsed.Lessons[0].Blocks,
		core.ParsedLessonBlock{
			Kind:          domain.VideoKind().String(),
			VideoProvider: domain.YouTubeProvider().String(),
			VideoLocator:  "dQw4w9WgXcQ",
			VideoCaption:  "Walkthrough",
		},
		core.ParsedLessonBlock{Kind: domain.QuizKind().String(), QuizRef: "foundations-quiz"},
		core.ParsedLessonBlock{Kind: domain.PracticeKind().String(), PracticeRef: "fizzbuzz"},
	)

	return parsed
}

func parsedImportSourceWithTextLesson(courseTitle string, markdown string) core.ParsedImportSource {
	return core.ParsedImportSource{
		FormatVersion: "1",
		Course:        parsedCourse(courseTitle),
		Quizzes:       []core.ParsedQuiz{parsedQuiz("Pick one")},
		Practices:     []core.ParsedPractice{parsedPractice("stdout")},
		Tests:         []core.ParsedTest{parsedTest("Pick one")},
		Lessons: []core.ParsedLesson{{
			Title:  "Welcome",
			Blocks: []core.ParsedLessonBlock{{Kind: domain.TextKind().String(), Markdown: markdown}},
		}},
	}
}

func parsedCourse(title string) core.ParsedCourse {
	return core.ParsedCourse{
		Title:       title,
		Slug:        "intro-to-go",
		Description: "",
		Status:      domain.Draft().String(),
	}
}

func parsedQuiz(prompt string) core.ParsedQuiz {
	return core.ParsedQuiz{
		Slug:  "foundations-quiz",
		Title: "Foundations Quiz",
		Questions: []core.ParsedQuestion{{
			Type:           domain.SingleChoice().String(),
			Prompt:         prompt,
			Options:        []string{"A", "B"},
			CorrectIndices: []int{0},
			Explanation:    "Because A",
		}},
	}
}

func parsedPractice(expectedStdout string) core.ParsedPractice {
	return core.ParsedPractice{
		Slug:        "fizzbuzz",
		Title:       "FizzBuzz",
		Language:    domain.Golang().String(),
		Prompt:      "Print fizz buzz.",
		StarterCode: "package main",
		Solution:    "package main",
		TestCases: []core.ParsedPracticeTestCase{{
			Stdin:          "stdin",
			ExpectedStdout: expectedStdout,
			Name:           "case",
		}},
	}
}

func parsedTest(prompt string) core.ParsedTest {
	return core.ParsedTest{
		Slug:  "final-test",
		Title: "Final Test",
		Items: []core.ParsedTestItem{{
			Kind:           domain.ChoiceKind().String(),
			Prompt:         prompt,
			ChoiceType:     domain.SingleChoice().String(),
			Options:        []string{"A", "B"},
			CorrectIndices: []int{0},
			Explanation:    "Because A",
		}},
	}
}

func storeExistingImportAggregates(
	t *testing.T,
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	quizzes *quizRepositoryFake,
	practices *practiceRepositoryFake,
	tests *testRepositoryFake,
) {
	t.Helper()

	courses.store(mustCourse(t, courseIDValue, "Intro to Go", "intro-to-go", domain.Draft()))
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Foundations Quiz", domain.DefaultPassThreshold().Float64(), mustChoiceQuestion(t, questionIDValue, 0)))
	practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", domain.Golang().String(), "Print fizz buzz.", "package main", "package main", mustTestCase(t, testCaseIDValue, 0)))
	tests.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", domain.DefaultPassThreshold().Float64(), nil, nil, mustChoiceTestItem(testItemIDValue, 0)))
	lessons.store(mustLessonFixture(t, lessonIDValue, courseIDValue, "Welcome", []domain.ContentBlock{mustTextBlock(t, importBlockID, 0, "Read this")}, 0))
}

func importOperationByRef(t *testing.T, operations []domain.ImportOperation, ref string) domain.ImportOperation {
	t.Helper()

	for _, operation := range operations {
		if operation.EntityRef() == ref {
			return operation
		}
	}

	t.Fatalf("expected operation %q", ref)
	return domain.ImportOperation{}
}

func importConflictByRef(t *testing.T, conflicts []domain.ImportConflict, ref string) domain.ImportConflict {
	t.Helper()

	for _, conflict := range conflicts {
		if conflict.EntityRef() == ref {
			return conflict
		}
	}

	t.Fatalf("expected conflict %q", ref)
	return domain.ImportConflict{}
}

func assertCreateOperation(t *testing.T, operation domain.ImportOperation, entityType domain.EntityType) {
	t.Helper()

	if !operation.Kind().IsCreate() {
		t.Fatalf("expected create operation, got %s", operation.Kind().String())
	}
	if operation.EntityType() != entityType {
		t.Fatalf("expected entity type %s, got %s", entityType.String(), operation.EntityType().String())
	}
	if operation.TargetID() != nil {
		t.Fatalf("expected create operation target id to be nil")
	}
}

func decodeOperationPayload(t *testing.T, operation domain.ImportOperation) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(operation.Payload(), &payload); err != nil {
		t.Fatalf("expected JSON payload, got %v", err)
	}

	return payload
}

func assertPayloadString(t *testing.T, payload map[string]any, field string, want string) {
	t.Helper()

	if got, _ := payload[field].(string); got != want {
		t.Fatalf("expected payload %s=%q, got %#v", field, want, payload[field])
	}
}

func assertPayloadNumber(t *testing.T, payload map[string]any, field string, want float64) {
	t.Helper()

	if got, _ := payload[field].(float64); got != want {
		t.Fatalf("expected payload %s=%v, got %#v", field, want, payload[field])
	}
}
