package usecase

import (
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	testIDValue     = "550e8400-e29b-41d4-a716-446655440050"
	otherTestID     = "550e8400-e29b-41d4-a716-446655440051"
	testItemIDValue = "550e8400-e29b-41d4-a716-446655440052"
	otherTestItemID = "550e8400-e29b-41d4-a716-446655440053"
)

func TestCreateTestSavesTestWithDefaults(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)}
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	tests := newTestRepositoryFake()
	service := newTestServiceFixture(courses, tests, clock)

	out, err := service.CreateTest(core.CreateTestInput{
		CourseID: courseIDValue,
		Title:    "  Final Test  ",
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	if out.ID != testIDValue {
		t.Fatalf("expected id %q, got %q", testIDValue, out.ID)
	}

	saved := tests.savedTests[0]
	if saved.ID().String() != testIDValue || saved.CourseID().String() != courseIDValue {
		t.Fatalf("expected saved test ids")
	}
	if saved.Title() != "Final Test" || saved.PassThreshold() != domain.DefaultPassThreshold() {
		t.Fatalf("expected title and default threshold, got title=%q threshold=%f", saved.Title(), saved.PassThreshold().Float64())
	}
	if saved.TimeLimit() != nil || saved.Solution() != nil || len(saved.Items()) != 0 {
		t.Fatalf("expected new test to start untimed, without solution, and without items")
	}
	if !saved.CreatedAt().Equal(clock.now) || !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected deterministic timestamps")
	}
}

func TestCreateTestUsesCustomTimeLimitAndThreshold(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	tests := newTestRepositoryFake()
	service := newTestServiceFixture(courses, tests, fixedClock{})
	minutes := 90
	threshold := 0.85

	if _, err := service.CreateTest(core.CreateTestInput{
		CourseID:         courseIDValue,
		Title:            "Final Test",
		TimeLimitMinutes: &minutes,
		PassThreshold:    &threshold,
	}); err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	saved := tests.savedTests[0]
	if got := saved.TimeLimit(); got == nil || got.Minutes() != minutes {
		t.Fatalf("expected custom time limit %d, got %+v", minutes, got)
	}
	if got := saved.PassThreshold().Float64(); got != threshold {
		t.Fatalf("expected custom threshold %f, got %f", threshold, got)
	}
}

func TestCreateTestRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name       string
		input      core.CreateTestInput
		seedCourse bool
		wantError  error
	}{
		{name: "invalid course id", input: core.CreateTestInput{CourseID: "bad-id", Title: "Test"}, wantError: domain.ErrValidation},
		{name: "course not found", input: core.CreateTestInput{CourseID: courseIDValue, Title: "Test"}, wantError: domain.ErrNotFound},
		{name: "invalid time limit", input: core.CreateTestInput{CourseID: courseIDValue, Title: "Test", TimeLimitMinutes: intPointer(0)}, seedCourse: true, wantError: domain.ErrValidation},
		{name: "invalid threshold", input: core.CreateTestInput{CourseID: courseIDValue, Title: "Test", PassThreshold: floatPointer(1.2)}, seedCourse: true, wantError: domain.ErrValidation},
		{name: "missing title", input: core.CreateTestInput{CourseID: courseIDValue, Title: "   "}, seedCourse: true, wantError: domain.ErrValidation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			courses := newCourseRepositoryFake()
			if test.seedCourse {
				courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
			}
			repo := newTestRepositoryFake()
			service := newTestServiceFixture(courses, repo, fixedClock{})

			_, err := service.CreateTest(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(repo.savedTests) != 0 {
				t.Fatalf("expected invalid test not to be saved")
			}
		})
	}
}

func TestListTestsValidatesCourseAndReturnsViews(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	repo := newTestRepositoryFake()
	firstLimit := mustTestTimeLimit(45)
	first := mustDomainTest(t, testIDValue, courseIDValue, "First Test", 0.7, &firstLimit, nil, mustChoiceTestItem(testItemIDValue, 0))
	second := mustDomainTest(t, otherTestID, courseIDValue, "Second Test", 0.8, nil, mustTestSolution(), mustCodingTestItem(otherTestItemID, 0))
	repo.store(first)
	repo.store(second)
	service := newTestServiceFixture(courses, repo, fixedClock{})

	out, err := service.ListTests(core.ListTestsInput{CourseID: courseIDValue})
	if err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}

	if len(out.Tests) != 2 {
		t.Fatalf("expected two tests, got %d", len(out.Tests))
	}
	if out.Tests[0].ID != testIDValue || out.Tests[0].ItemCount != 1 || out.Tests[0].TimeLimitMinutes == nil || *out.Tests[0].TimeLimitMinutes != 45 {
		t.Fatalf("expected first test view, got %+v", out.Tests[0])
	}
	if out.Tests[1].ID != otherTestID || !out.Tests[1].HasSolution || out.Tests[1].PassThreshold != 0.8 {
		t.Fatalf("expected second test view, got %+v", out.Tests[1])
	}
}

func TestListTestsRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.ListTestsInput
		wantError error
	}{
		{name: "invalid course id", input: core.ListTestsInput{CourseID: "bad-id"}, wantError: domain.ErrValidation},
		{name: "course not found", input: core.ListTestsInput{CourseID: courseIDValue}, wantError: domain.ErrNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newTestServiceFixture(newCourseRepositoryFake(), newTestRepositoryFake(), fixedClock{})

			_, err := service.ListTests(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
		})
	}
}

func TestGetTestReturnsDetailWithSolutionAndOrderedItems(t *testing.T) {
	repo := newTestRepositoryFake()
	test := mustDomainTest(
		t,
		testIDValue,
		courseIDValue,
		"Final Test",
		0.7,
		nil,
		mustTestSolution(),
		mustCodingTestItem(otherTestItemID, 1),
		mustChoiceTestItem(testItemIDValue, 0),
	)
	repo.store(test)
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})

	out, err := service.GetTest(core.GetTestInput{ID: testIDValue})
	if err != nil {
		t.Fatalf("expected get to succeed, got %v", err)
	}

	if out.Test.ID != testIDValue || out.Test.ItemCount != 2 || out.Test.Solution == nil {
		t.Fatalf("expected test detail, got %+v", out.Test)
	}
	if out.Test.Solution.ZipProvider != "url" || out.Test.Solution.VideoCaption != "Walkthrough" {
		t.Fatalf("expected solution fields, got %+v", out.Test.Solution)
	}
	if out.Test.Items[0].ID != testItemIDValue || out.Test.Items[0].ChoicePrompt != "Pick one" || out.Test.Items[0].ChoiceCorrectIndices[0] != 0 {
		t.Fatalf("expected choice item mapped first, got %+v", out.Test.Items)
	}
	if out.Test.Items[1].ID != otherTestItemID || out.Test.Items[1].CodingPrompt != "Write a program." || out.Test.Items[1].TestCases[0].ExpectedStdout != "stdout" {
		t.Fatalf("expected coding item mapped second, got %+v", out.Test.Items)
	}
}

func TestGetTestRejectsFailureModes(t *testing.T) {
	service := newTestServiceFixture(newCourseRepositoryFake(), newTestRepositoryFake(), fixedClock{})

	if _, err := service.GetTest(core.GetTestInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.GetTest(core.GetTestInput{ID: testIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateTestChangesMetadataClearsTimeLimitAndSetsSolution(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 27, 11, 0, 0, 0, time.UTC)}
	timeLimit := mustTestTimeLimit(60)
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, &timeLimit, nil))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, clock)
	title := "Updated Test"
	clearLimit := 0
	threshold := 0.9
	zipProvider := "url"
	zipLocator := "https://example.com/solution.zip"
	videoProvider := "youtube"
	videoLocator := "dQw4w9WgXcQ"
	caption := "Walkthrough"

	out, err := service.UpdateTest(core.UpdateTestInput{
		ID:                    testIDValue,
		Title:                 &title,
		TimeLimitMinutes:      &clearLimit,
		PassThreshold:         &threshold,
		SolutionZipProvider:   &zipProvider,
		SolutionZipLocator:    &zipLocator,
		SolutionVideoProvider: &videoProvider,
		SolutionVideoLocator:  &videoLocator,
		SolutionVideoCaption:  &caption,
	})
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if out.ID != testIDValue {
		t.Fatalf("expected output id %q, got %q", testIDValue, out.ID)
	}
	saved := repo.savedTests[0]
	if saved.Title() != title || saved.TimeLimit() != nil || saved.PassThreshold().Float64() != threshold {
		t.Fatalf("expected metadata updates, got title=%q limit=%+v threshold=%f", saved.Title(), saved.TimeLimit(), saved.PassThreshold().Float64())
	}
	if saved.Solution() == nil || saved.Solution().ExplanationCaption() != caption {
		t.Fatalf("expected solution to be set, got %+v", saved.Solution())
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestUpdateTestSetsPositiveTimeLimit(t *testing.T) {
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})
	minutes := 75

	if _, err := service.UpdateTest(core.UpdateTestInput{ID: testIDValue, TimeLimitMinutes: &minutes}); err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if got := repo.savedTests[0].TimeLimit(); got == nil || got.Minutes() != minutes {
		t.Fatalf("expected time limit %d, got %+v", minutes, got)
	}
}

func TestUpdateTestRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.UpdateTestInput
		seed      bool
		wantError error
	}{
		{name: "invalid id", input: core.UpdateTestInput{ID: "bad-id", Title: stringPointer("Updated")}, wantError: domain.ErrValidation},
		{name: "nothing to update", input: core.UpdateTestInput{ID: testIDValue}, seed: true, wantError: domain.ErrValidation},
		{name: "not found", input: core.UpdateTestInput{ID: testIDValue, Title: stringPointer("Updated")}, wantError: domain.ErrNotFound},
		{name: "empty title", input: core.UpdateTestInput{ID: testIDValue, Title: stringPointer("   ")}, seed: true, wantError: domain.ErrValidation},
		{name: "invalid threshold", input: core.UpdateTestInput{ID: testIDValue, PassThreshold: floatPointer(-0.1)}, seed: true, wantError: domain.ErrValidation},
		{name: "negative time limit", input: core.UpdateTestInput{ID: testIDValue, TimeLimitMinutes: intPointer(-1)}, seed: true, wantError: domain.ErrValidation},
		{
			name: "partial solution group",
			input: core.UpdateTestInput{
				ID:                  testIDValue,
				SolutionZipProvider: stringPointer("url"),
				SolutionZipLocator:  stringPointer("https://example.com/solution.zip"),
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "caption without solution group",
			input: core.UpdateTestInput{
				ID:                   testIDValue,
				SolutionVideoCaption: stringPointer("Walkthrough"),
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "invalid solution provider",
			input: core.UpdateTestInput{
				ID:                    testIDValue,
				SolutionZipProvider:   stringPointer("s3"),
				SolutionZipLocator:    stringPointer("https://example.com/solution.zip"),
				SolutionVideoProvider: stringPointer("youtube"),
				SolutionVideoLocator:  stringPointer("dQw4w9WgXcQ"),
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "invalid solution locator",
			input: core.UpdateTestInput{
				ID:                    testIDValue,
				SolutionZipProvider:   stringPointer("url"),
				SolutionZipLocator:    stringPointer("/relative.zip"),
				SolutionVideoProvider: stringPointer("youtube"),
				SolutionVideoLocator:  stringPointer("dQw4w9WgXcQ"),
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newTestRepositoryFake()
			if test.seed {
				repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil))
			}
			service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})

			_, err := service.UpdateTest(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(repo.savedTests) != 0 {
				t.Fatalf("expected invalid update not to be saved")
			}
		})
	}
}

func TestDeleteTestDeletesUnconditionally(t *testing.T) {
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})

	if err := service.DeleteTest(core.DeleteTestInput{ID: testIDValue}); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	if _, exists := repo.tests[testIDValue]; exists {
		t.Fatalf("expected test to be deleted")
	}
	if len(repo.deletedTestIDs) != 1 || repo.deletedTestIDs[0].String() != testIDValue {
		t.Fatalf("expected test delete to be recorded")
	}
}

func TestDeleteTestRejectsFailureModes(t *testing.T) {
	service := newTestServiceFixture(newCourseRepositoryFake(), newTestRepositoryFake(), fixedClock{})

	if err := service.DeleteTest(core.DeleteTestInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if err := service.DeleteTest(core.DeleteTestInput{ID: testIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestAddTestItemAppendsChoiceItem(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 27, 12, 0, 0, 0, time.UTC)}
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil, mustCodingTestItem(otherTestItemID, 0)))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, clock)

	out, err := service.AddTestItem(core.AddTestItemInput{
		TestID:         testIDValue,
		Kind:           "choice",
		Prompt:         "  Pick one  ",
		ChoiceType:     "single",
		Options:        []string{"A", "B"},
		CorrectIndices: []int{0},
		Explanation:    "Because A",
	})
	if err != nil {
		t.Fatalf("expected add choice item to succeed, got %v", err)
	}

	if out.ID != testItemIDValue {
		t.Fatalf("expected item id %q, got %q", testItemIDValue, out.ID)
	}
	saved := repo.savedTests[0]
	items := saved.Items()
	if len(items) != 2 {
		t.Fatalf("expected two items, got %d", len(items))
	}
	added := items[1]
	if added.ID().String() != testItemIDValue || added.Position().Int() != 1 || !added.Kind().IsChoice() {
		t.Fatalf("expected appended choice item at position 1, got %+v", added)
	}
	body := added.Body().(domain.ChoiceItemBody)
	if body.Prompt() != "Pick one" || body.Explanation() != "Because A" {
		t.Fatalf("expected choice item content to be saved")
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestAddTestItemInsertsCodingItemAtExplicitPosition(t *testing.T) {
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil, mustChoiceTestItem(otherTestItemID, 0)))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})
	position := 0

	if _, err := service.AddTestItem(core.AddTestItemInput{
		TestID:       testIDValue,
		Kind:         "coding",
		Position:     &position,
		CodingPrompt: "  Write code  ",
		Language:     "golang",
		StarterCode:  "package main",
		Solution:     "package main",
		TestCases:    []core.CodingTestCaseDTO{{Stdin: "stdin", ExpectedStdout: "stdout", Name: "sample"}},
	}); err != nil {
		t.Fatalf("expected add coding item to succeed, got %v", err)
	}

	items := repo.savedTests[0].Items()
	if items[0].ID().String() != testItemIDValue || items[0].Position().Int() != 0 || !items[0].Kind().IsCoding() {
		t.Fatalf("expected inserted coding item at position 0, got %+v", items[0])
	}
	body := items[0].Body().(domain.CodingItemBody)
	if body.Prompt() != "Write code" || body.Language() != domain.Golang() || body.TestCases()[0].ExpectedStdout() != "stdout" {
		t.Fatalf("expected coding item body to be saved, got %+v", body)
	}
	if items[1].ID().String() != otherTestItemID || items[1].Position().Int() != 1 {
		t.Fatalf("expected existing item to shift to position 1, got %+v", items[1])
	}
}

func TestAddTestItemRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.AddTestItemInput
		seed      bool
		wantError error
	}{
		{name: "invalid test id", input: core.AddTestItemInput{TestID: "bad-id", Kind: "choice", Prompt: "Pick", ChoiceType: "single", Options: []string{"A", "B"}, CorrectIndices: []int{0}}, wantError: domain.ErrValidation},
		{name: "test not found", input: core.AddTestItemInput{TestID: testIDValue, Kind: "choice", Prompt: "Pick", ChoiceType: "single", Options: []string{"A", "B"}, CorrectIndices: []int{0}}, wantError: domain.ErrNotFound},
		{name: "invalid kind", input: core.AddTestItemInput{TestID: testIDValue, Kind: "essay"}, seed: true, wantError: domain.ErrValidation},
		{name: "negative position", input: core.AddTestItemInput{TestID: testIDValue, Kind: "choice", Position: intPointer(-1), Prompt: "Pick", ChoiceType: "single", Options: []string{"A", "B"}, CorrectIndices: []int{0}}, seed: true, wantError: domain.ErrValidation},
		{name: "duplicate correct index", input: core.AddTestItemInput{TestID: testIDValue, Kind: "choice", Prompt: "Pick", ChoiceType: "multiple", Options: []string{"A", "B"}, CorrectIndices: []int{0, 0}}, seed: true, wantError: domain.ErrValidation},
		{name: "out of range correct index", input: core.AddTestItemInput{TestID: testIDValue, Kind: "choice", Prompt: "Pick", ChoiceType: "single", Options: []string{"A", "B"}, CorrectIndices: []int{2}}, seed: true, wantError: domain.ErrValidation},
		{name: "single choice cardinality", input: core.AddTestItemInput{TestID: testIDValue, Kind: "choice", Prompt: "Pick", ChoiceType: "single", Options: []string{"A", "B"}, CorrectIndices: []int{0, 1}}, seed: true, wantError: domain.ErrValidation},
		{name: "invalid language", input: core.AddTestItemInput{TestID: testIDValue, Kind: "coding", CodingPrompt: "Solve", Language: "python", TestCases: []core.CodingTestCaseDTO{{ExpectedStdout: "ok"}}}, seed: true, wantError: domain.ErrValidation},
		{name: "no coding test cases", input: core.AddTestItemInput{TestID: testIDValue, Kind: "coding", CodingPrompt: "Solve", Language: "golang"}, seed: true, wantError: domain.ErrValidation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newTestRepositoryFake()
			if test.seed {
				repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil))
			}
			service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})

			_, err := service.AddTestItem(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(repo.savedTests) != 0 {
				t.Fatalf("expected invalid add not to be saved")
			}
		})
	}
}

func TestListTestItemsReturnsOrderedViews(t *testing.T) {
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(
		t,
		testIDValue,
		courseIDValue,
		"Final Test",
		0.7,
		nil,
		nil,
		mustCodingTestItem(otherTestItemID, 1),
		mustChoiceTestItem(testItemIDValue, 0),
	))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})

	out, err := service.ListTestItems(core.ListTestItemsInput{TestID: testIDValue})
	if err != nil {
		t.Fatalf("expected list items to succeed, got %v", err)
	}

	if len(out.Items) != 2 {
		t.Fatalf("expected two items, got %d", len(out.Items))
	}
	if out.Items[0].ID != testItemIDValue || out.Items[0].Position != 0 || out.Items[0].ChoicePrompt != "Pick one" {
		t.Fatalf("expected choice item first, got %+v", out.Items)
	}
	if out.Items[1].ID != otherTestItemID || out.Items[1].TestCases[0].Name != "sample" {
		t.Fatalf("expected coding item second, got %+v", out.Items[1])
	}
}

func TestListTestItemsRejectsFailureModes(t *testing.T) {
	service := newTestServiceFixture(newCourseRepositoryFake(), newTestRepositoryFake(), fixedClock{})

	if _, err := service.ListTestItems(core.ListTestItemsInput{TestID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.ListTestItems(core.ListTestItemsInput{TestID: testIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGetTestItemFindsOwningTest(t *testing.T) {
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil, mustChoiceTestItem(testItemIDValue, 0)))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})

	out, err := service.GetTestItem(core.GetTestItemInput{ID: testItemIDValue})
	if err != nil {
		t.Fatalf("expected get item to succeed, got %v", err)
	}

	if out.Item.ID != testItemIDValue || out.Item.TestID != testIDValue || out.Item.Kind != "choice" {
		t.Fatalf("expected item view from owning test, got %+v", out.Item)
	}
}

func TestGetTestItemRejectsFailureModes(t *testing.T) {
	service := newTestServiceFixture(newCourseRepositoryFake(), newTestRepositoryFake(), fixedClock{})

	if _, err := service.GetTestItem(core.GetTestItemInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.GetTestItem(core.GetTestItemInput{ID: testItemIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateTestItemChangesChoiceBody(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 27, 13, 0, 0, 0, time.UTC)}
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil, mustChoiceTestItem(testItemIDValue, 0)))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, clock)
	prompt := "Pick many"
	choiceType := "multiple"
	options := []string{"A", "B", "C"}
	correct := []int{0, 2}
	explanation := "Because A and C"

	out, err := service.UpdateTestItem(core.UpdateTestItemInput{
		ID:             testItemIDValue,
		Prompt:         &prompt,
		ChoiceType:     &choiceType,
		Options:        &options,
		CorrectIndices: &correct,
		Explanation:    &explanation,
	})
	if err != nil {
		t.Fatalf("expected update choice item to succeed, got %v", err)
	}

	if out.ID != testItemIDValue {
		t.Fatalf("expected output id %q, got %q", testItemIDValue, out.ID)
	}
	saved := repo.savedTests[0]
	item, err := saved.Item(mustTestItemID(testItemIDValue))
	if err != nil {
		t.Fatalf("expected saved item, got %v", err)
	}
	body := item.Body().(domain.ChoiceItemBody)
	if body.Prompt() != prompt || !body.Type().IsMultiple() || body.Explanation() != explanation {
		t.Fatalf("expected choice body updates, got %+v", body)
	}
	if body.CorrectIndices()[1] != 2 || !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected content and timestamp to update")
	}
}

func TestUpdateTestItemChangesCodingBodyAndReplacesCases(t *testing.T) {
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil, mustCodingTestItem(testItemIDValue, 0)))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})
	prompt := "Solve it"
	language := "rust"
	starter := "package main\nfunc main() {}"
	solution := "package main\nfunc main() { println(\"ok\") }"
	testCases := []core.CodingTestCaseDTO{
		{Stdin: "1", ExpectedStdout: "one", Name: "one"},
		{Stdin: "2", ExpectedStdout: "two", Name: "two"},
	}

	if _, err := service.UpdateTestItem(core.UpdateTestItemInput{
		ID:          testItemIDValue,
		Prompt:      &prompt,
		Language:    &language,
		StarterCode: &starter,
		Solution:    &solution,
		TestCases:   &testCases,
	}); err != nil {
		t.Fatalf("expected update coding item to succeed, got %v", err)
	}

	item, err := repo.savedTests[0].Item(mustTestItemID(testItemIDValue))
	if err != nil {
		t.Fatalf("expected saved item, got %v", err)
	}
	body := item.Body().(domain.CodingItemBody)
	if body.Prompt() != prompt || body.Language() != domain.Rust() || body.StarterCode() != starter || body.Solution() != solution {
		t.Fatalf("expected coding body updates, got %+v", body)
	}
	if cases := body.TestCases(); len(cases) != 2 || cases[1].ExpectedStdout() != "two" {
		t.Fatalf("expected coding test cases to be replaced, got %+v", cases)
	}
}

func TestUpdateTestItemRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.UpdateTestItemInput
		item      domain.TestItem
		wantError error
	}{
		{name: "invalid id", input: core.UpdateTestItemInput{ID: "bad-id", Prompt: stringPointer("Updated")}, wantError: domain.ErrValidation},
		{name: "empty update", input: core.UpdateTestItemInput{ID: testItemIDValue}, item: mustChoiceTestItem(testItemIDValue, 0), wantError: domain.ErrValidation},
		{name: "not found", input: core.UpdateTestItemInput{ID: testItemIDValue, Prompt: stringPointer("Updated")}, wantError: domain.ErrNotFound},
		{name: "choice item with coding fields", input: core.UpdateTestItemInput{ID: testItemIDValue, CodingPrompt: stringPointer("Solve")}, item: mustChoiceTestItem(testItemIDValue, 0), wantError: domain.ErrValidation},
		{name: "coding item with choice fields", input: core.UpdateTestItemInput{ID: testItemIDValue, ChoiceType: stringPointer("single")}, item: mustCodingTestItem(testItemIDValue, 0), wantError: domain.ErrValidation},
		{name: "options without correct indices", input: core.UpdateTestItemInput{ID: testItemIDValue, Options: stringSlicePointer([]string{"A", "B"})}, item: mustChoiceTestItem(testItemIDValue, 0), wantError: domain.ErrValidation},
		{name: "correct indices without options", input: core.UpdateTestItemInput{ID: testItemIDValue, CorrectIndices: intSlicePointer([]int{0})}, item: mustChoiceTestItem(testItemIDValue, 0), wantError: domain.ErrValidation},
		{name: "invalid choice content", input: core.UpdateTestItemInput{ID: testItemIDValue, Options: stringSlicePointer([]string{"A"}), CorrectIndices: intSlicePointer([]int{0})}, item: mustChoiceTestItem(testItemIDValue, 0), wantError: domain.ErrValidation},
		{name: "empty coding test cases", input: core.UpdateTestItemInput{ID: testItemIDValue, TestCases: codingTestCaseSlicePointer(nil)}, item: mustCodingTestItem(testItemIDValue, 0), wantError: domain.ErrValidation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newTestRepositoryFake()
			if test.item.ID().String() != "" {
				repo.store(mustDomainTest(t, testIDValue, courseIDValue, "Final Test", 0.7, nil, nil, test.item))
			}
			service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})

			_, err := service.UpdateTestItem(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(repo.savedTests) != 0 {
				t.Fatalf("expected invalid update not to be saved")
			}
		})
	}
}

func TestRemoveTestItemCompactsPositions(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 27, 14, 0, 0, 0, time.UTC)}
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(
		t,
		testIDValue,
		courseIDValue,
		"Final Test",
		0.7,
		nil,
		nil,
		mustChoiceTestItem(testItemIDValue, 0),
		mustCodingTestItem(otherTestItemID, 1),
	))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, clock)

	if err := service.RemoveTestItem(core.RemoveTestItemInput{ID: testItemIDValue}); err != nil {
		t.Fatalf("expected remove item to succeed, got %v", err)
	}

	saved := repo.savedTests[0]
	items := saved.Items()
	if len(items) != 1 {
		t.Fatalf("expected one remaining item, got %d", len(items))
	}
	if items[0].ID().String() != otherTestItemID || items[0].Position().Int() != 0 {
		t.Fatalf("expected remaining item compacted to position 0, got %+v", items[0])
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestRemoveTestItemRejectsFailureModes(t *testing.T) {
	service := newTestServiceFixture(newCourseRepositoryFake(), newTestRepositoryFake(), fixedClock{})

	if err := service.RemoveTestItem(core.RemoveTestItemInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if err := service.RemoveTestItem(core.RemoveTestItemInput{ID: testItemIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestReorderTestItemsSavesPermutation(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 27, 15, 0, 0, 0, time.UTC)}
	repo := newTestRepositoryFake()
	repo.store(mustDomainTest(
		t,
		testIDValue,
		courseIDValue,
		"Final Test",
		0.7,
		nil,
		nil,
		mustChoiceTestItem(testItemIDValue, 0),
		mustCodingTestItem(otherTestItemID, 1),
	))
	service := newTestServiceFixture(newCourseRepositoryFake(), repo, clock)

	err := service.ReorderTestItems(core.ReorderTestItemsInput{
		TestID: testIDValue,
		Order: []core.TestItemPlacementDTO{
			{TestItemID: otherTestItemID, Position: 0},
			{TestItemID: testItemIDValue, Position: 1},
		},
	})
	if err != nil {
		t.Fatalf("expected reorder to succeed, got %v", err)
	}

	saved := repo.savedTests[0]
	items := saved.Items()
	if items[0].ID().String() != otherTestItemID || items[1].ID().String() != testItemIDValue {
		t.Fatalf("expected items reordered by position, got %+v", items)
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestReorderTestItemsRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.ReorderTestItemsInput
		seed      bool
		wantError error
	}{
		{name: "invalid test id", input: core.ReorderTestItemsInput{TestID: "bad-id"}, wantError: domain.ErrValidation},
		{name: "test not found", input: core.ReorderTestItemsInput{TestID: testIDValue}, wantError: domain.ErrNotFound},
		{
			name: "unknown item",
			input: core.ReorderTestItemsInput{
				TestID: testIDValue,
				Order: []core.TestItemPlacementDTO{
					{TestItemID: testItemIDValue, Position: 0},
					{TestItemID: "550e8400-e29b-41d4-a716-446655440099", Position: 1},
				},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "duplicate position",
			input: core.ReorderTestItemsInput{
				TestID: testIDValue,
				Order: []core.TestItemPlacementDTO{
					{TestItemID: testItemIDValue, Position: 0},
					{TestItemID: otherTestItemID, Position: 0},
				},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "missing item",
			input: core.ReorderTestItemsInput{
				TestID: testIDValue,
				Order:  []core.TestItemPlacementDTO{{TestItemID: testItemIDValue, Position: 0}},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "position gap",
			input: core.ReorderTestItemsInput{
				TestID: testIDValue,
				Order: []core.TestItemPlacementDTO{
					{TestItemID: testItemIDValue, Position: 0},
					{TestItemID: otherTestItemID, Position: 2},
				},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := newTestRepositoryFake()
			if test.seed {
				repo.store(mustDomainTest(
					t,
					testIDValue,
					courseIDValue,
					"Final Test",
					0.7,
					nil,
					nil,
					mustChoiceTestItem(testItemIDValue, 0),
					mustCodingTestItem(otherTestItemID, 1),
				))
			}
			service := newTestServiceFixture(newCourseRepositoryFake(), repo, fixedClock{})

			err := service.ReorderTestItems(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(repo.savedTests) != 0 {
				t.Fatalf("expected invalid reorder not to be saved")
			}
		})
	}
}

func newTestServiceFixture(
	courses *courseRepositoryFake,
	tests *testRepositoryFake,
	clock fixedClock,
) *TestService {
	ids := fixedIDGenerator{
		testID:     mustTestID(testIDValue),
		testItemID: mustTestItemID(testItemIDValue),
	}
	return NewTestService(courses, tests, ids, clock)
}

type testRepositoryFake struct {
	tests          map[string]domain.Test
	savedTests     []domain.Test
	deletedTestIDs []domain.TestID
}

func newTestRepositoryFake() *testRepositoryFake {
	return &testRepositoryFake{tests: make(map[string]domain.Test)}
}

func (repo *testRepositoryFake) Save(test domain.Test) error {
	repo.savedTests = append(repo.savedTests, test)
	repo.store(test)

	return nil
}

func (repo *testRepositoryFake) FindByID(id domain.TestID) (domain.Test, error) {
	test, exists := repo.tests[id.String()]
	if !exists {
		return domain.Test{}, domain.ErrNotFound
	}

	return test, nil
}

func (repo *testRepositoryFake) FindByCourse(courseID domain.CourseID) ([]domain.Test, error) {
	tests := make([]domain.Test, 0, len(repo.tests))
	for _, test := range repo.tests {
		if test.CourseID() == courseID {
			tests = append(tests, test)
		}
	}

	sort.Slice(tests, func(i, j int) bool {
		return tests[i].CreatedAt().Before(tests[j].CreatedAt())
	})

	return tests, nil
}

func (repo *testRepositoryFake) FindByItemID(id domain.TestItemID) (domain.Test, error) {
	for _, test := range repo.tests {
		if _, err := test.Item(id); err == nil {
			return test, nil
		}
	}

	return domain.Test{}, domain.ErrNotFound
}

func (repo *testRepositoryFake) Delete(id domain.TestID) error {
	if _, exists := repo.tests[id.String()]; !exists {
		return domain.ErrNotFound
	}

	repo.deletedTestIDs = append(repo.deletedTestIDs, id)
	delete(repo.tests, id.String())

	return nil
}

func (repo *testRepositoryFake) DeleteByCourse(courseID domain.CourseID) error {
	for id, test := range repo.tests {
		if test.CourseID() == courseID {
			delete(repo.tests, id)
		}
	}

	return nil
}

func (repo *testRepositoryFake) store(test domain.Test) {
	if repo.tests == nil {
		repo.tests = make(map[string]domain.Test)
	}

	repo.tests[test.ID().String()] = test
}

func mustDomainTest(
	t *testing.T,
	idValue string,
	courseIDValue string,
	title string,
	thresholdValue float64,
	timeLimit *domain.TimeLimit,
	solution *domain.TestSolution,
	items ...domain.TestItem,
) domain.Test {
	t.Helper()

	threshold, err := domain.NewPassThreshold(thresholdValue)
	if err != nil {
		t.Fatalf("expected pass threshold fixture, got %v", err)
	}

	test, err := domain.NewTest(
		mustTestID(idValue),
		mustCourseID(courseIDValue),
		title,
		timeLimit,
		threshold,
		solution,
		items,
		time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected test fixture, got %v", err)
	}

	return test
}

func mustChoiceTestItem(idValue string, positionValue int) domain.TestItem {
	body, err := domain.NewChoiceItemBody(
		domain.SingleChoice(),
		"Pick one",
		[]string{"A", "B"},
		[]int{0},
		"Because A",
	)
	if err != nil {
		panic(err)
	}

	item, err := domain.NewTestItem(mustTestItemID(idValue), domain.ChoiceKind(), body, mustTestItemPosition(positionValue))
	if err != nil {
		panic(err)
	}

	return item
}

func mustCodingTestItem(idValue string, positionValue int) domain.TestItem {
	body, err := domain.NewCodingItemBody(
		domain.Golang(),
		"Write a program.",
		"package main",
		"package main",
		[]domain.CodingTestCase{domain.NewCodingTestCase("stdin", "stdout", "sample")},
	)
	if err != nil {
		panic(err)
	}

	item, err := domain.NewTestItem(mustTestItemID(idValue), domain.CodingKind(), body, mustTestItemPosition(positionValue))
	if err != nil {
		panic(err)
	}

	return item
}

func mustTestSolution() *domain.TestSolution {
	solution, err := domain.NewTestSolution(
		mustTestMediaRef(domain.URLProvider(), "https://example.com/solution.zip"),
		mustTestMediaRef(domain.YouTubeProvider(), "dQw4w9WgXcQ"),
		"Walkthrough",
	)
	if err != nil {
		panic(err)
	}

	return &solution
}

func mustTestMediaRef(provider domain.MediaProvider, locator string) domain.MediaRef {
	ref, err := domain.NewMediaRef(provider, locator)
	if err != nil {
		panic(err)
	}

	return ref
}

func mustTestID(value string) domain.TestID {
	id, err := domain.NewTestID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustTestItemID(value string) domain.TestItemID {
	id, err := domain.NewTestItemID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustTestItemPosition(value int) domain.TestItemPosition {
	position, err := domain.NewTestItemPosition(value)
	if err != nil {
		panic(err)
	}

	return position
}

func mustTestTimeLimit(value int) domain.TimeLimit {
	timeLimit, err := domain.NewTimeLimit(value)
	if err != nil {
		panic(err)
	}

	return timeLimit
}

func codingTestCaseSlicePointer(value []core.CodingTestCaseDTO) *[]core.CodingTestCaseDTO {
	return &value
}
