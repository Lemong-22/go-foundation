package core

import (
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/domain"
)

const importPortZipHash = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

func TestImportPortInterfacesCompile(t *testing.T) {
	var _ ImportService = importServiceStub{}
	var _ ImportSource = importSourceStub{}
}

func TestImportDTOsKeepAdapterInputsPrimitiveAndOpaque(t *testing.T) {
	resolvedPlanJSON := []byte(`{"format_version":"1"}`)
	input := ApplyPlanInput{
		ZipPath:          "/tmp/course.zip",
		InstructorID:     "550e8400-e29b-41d4-a716-446655440000",
		ResolvedPlanJSON: resolvedPlanJSON,
		ConflictStrategy: "fail",
	}

	if input.ZipPath == "" || input.InstructorID == "" || string(input.ResolvedPlanJSON) == "" || input.ConflictStrategy != "fail" {
		t.Fatalf("expected import apply input to carry primitive and opaque boundary values")
	}

	plan := mustCoreImportPlan(t)
	planOutput := PlanImportOutput{Plan: plan}
	if planOutput.Plan.ZipHash() != importPortZipHash {
		t.Fatalf("expected plan output to carry import plan")
	}

	result := mustCoreApplyResult(t)
	applyOutput := ApplyPlanOutput{Result: result}
	if applyOutput.Result.AggregatesSucceeded() != 1 {
		t.Fatalf("expected apply output to carry apply result")
	}
}

func TestParsedImportSourceCarriesServiceDTOShape(t *testing.T) {
	order := 0
	passThreshold := 0.8
	timeLimit := 30

	source := ParsedImportSource{
		FormatVersion: "1",
		Course: ParsedCourse{
			Title:       "Intro to Go",
			Slug:        "intro-to-go",
			Description: "Learn Go",
			Status:      "draft",
		},
		Lessons: []ParsedLesson{
			{
				Title: "Foundations",
				Order: &order,
				Blocks: []ParsedLessonBlock{
					{Kind: "quiz", QuizRef: "foundations-quiz", Position: &order},
				},
			},
		},
		Quizzes: []ParsedQuiz{
			{
				Slug:          "foundations-quiz",
				Title:         "Foundations Quiz",
				PassThreshold: &passThreshold,
				Questions: []ParsedQuestion{
					{Type: "single", Prompt: "Pick one", Options: []string{"A", "B"}, CorrectIndices: []int{0}, Position: &order},
				},
			},
		},
		Practices: []ParsedPractice{
			{
				Slug:     "fizzbuzz",
				Title:    "FizzBuzz",
				Language: "golang",
				TestCases: []ParsedPracticeTestCase{
					{Stdin: "3", ExpectedStdout: "Fizz", Name: "multiple of three", Position: &order},
				},
			},
		},
		Tests: []ParsedTest{
			{
				Slug:             "midterm",
				Title:            "Midterm",
				TimeLimitMinutes: &timeLimit,
				PassThreshold:    &passThreshold,
				Solution:         &ParsedTestSolution{ZipProvider: "url", ZipLocator: "https://example.com/solution.zip"},
				Items: []ParsedTestItem{
					{Kind: "coding", CodingPrompt: "Write code", Language: "golang", TestCases: []ParsedCodingTestCase{{Stdin: "3", ExpectedStdout: "Fizz"}}},
				},
			},
		},
	}

	if source.Course.Slug != "intro-to-go" || source.Lessons[0].Blocks[0].QuizRef != "foundations-quiz" {
		t.Fatalf("expected parsed source to retain course and lesson block data")
	}
	if *source.Quizzes[0].PassThreshold != passThreshold || source.Practices[0].TestCases[0].ExpectedStdout != "Fizz" {
		t.Fatalf("expected parsed source to retain quiz and practice data")
	}
	if *source.Tests[0].TimeLimitMinutes != timeLimit || source.Tests[0].Items[0].TestCases[0].ExpectedStdout != "Fizz" {
		t.Fatalf("expected parsed source to retain test data")
	}
}

type importServiceStub struct{}

func (importServiceStub) PlanImport(in PlanImportInput) (PlanImportOutput, error) {
	return PlanImportOutput{}, nil
}

func (importServiceStub) ApplyPlan(in ApplyPlanInput) (ApplyPlanOutput, error) {
	return ApplyPlanOutput{}, nil
}

type importSourceStub struct{}

func (importSourceStub) Open(zipPath string) (ParsedImportSource, ImportSourceMetadata, error) {
	return ParsedImportSource{}, ImportSourceMetadata{}, nil
}

func mustCoreImportPlan(t *testing.T) domain.ImportPlan {
	t.Helper()

	plan, err := domain.NewImportPlan("1", importPortZipHash, time.Date(2026, 5, 28, 16, 0, 0, 0, time.UTC), nil, nil)
	if err != nil {
		t.Fatalf("expected import plan, got error %v", err)
	}

	return plan
}

func mustCoreApplyResult(t *testing.T) domain.ApplyResult {
	t.Helper()

	result, err := domain.NewApplyResult(nil, nil, nil, 1, 0)
	if err != nil {
		t.Fatalf("expected apply result, got error %v", err)
	}

	return result
}
