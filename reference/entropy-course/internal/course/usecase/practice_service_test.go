package usecase

import (
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	practiceIDValue = "550e8400-e29b-41d4-a716-446655440050"
	otherPracticeID = "550e8400-e29b-41d4-a716-446655440051"
	testCaseIDValue = "550e8400-e29b-41d4-a716-446655440060"
	otherTestCaseID = "550e8400-e29b-41d4-a716-446655440061"
	thirdTestCaseID = "550e8400-e29b-41d4-a716-446655440062"
)

func TestCreatePracticeSavesPractice(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 11, 0, 0, 0, time.UTC)}
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	practices := newPracticeRepositoryFake()
	service := newPracticeServiceFixture(courses, &lessonRepositoryFake{}, practices, clock)

	out, err := service.CreatePractice(core.CreatePracticeInput{
		CourseID:    courseIDValue,
		Title:       "  FizzBuzz  ",
		Language:    "golang",
		Prompt:      "  Print fizz buzz.  ",
		StarterCode: "package main",
		Solution:    "package main",
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	if out.ID != practiceIDValue {
		t.Fatalf("expected id %q, got %q", practiceIDValue, out.ID)
	}

	saved := practices.savedPractices[0]
	if saved.ID().String() != practiceIDValue || saved.CourseID().String() != courseIDValue {
		t.Fatalf("expected saved practice ids")
	}
	if saved.Title() != "FizzBuzz" || saved.Language() != domain.Golang() || saved.Prompt() != "Print fizz buzz." {
		t.Fatalf("expected normalized practice fields, got title=%q language=%q prompt=%q", saved.Title(), saved.Language().String(), saved.Prompt())
	}
	if saved.StarterCode() != "package main" || saved.Solution() != "package main" {
		t.Fatalf("expected starter code and solution to be saved")
	}
	if len(saved.TestCases()) != 0 {
		t.Fatalf("expected new practice to start without test cases")
	}
	if !saved.CreatedAt().Equal(clock.now) || !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected deterministic timestamps")
	}
}

func TestCreatePracticeRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name       string
		input      core.CreatePracticeInput
		seedCourse bool
		wantError  error
	}{
		{
			name:      "invalid course id",
			input:     core.CreatePracticeInput{CourseID: "bad-id", Title: "Practice", Language: "golang", Prompt: "Solve it"},
			wantError: domain.ErrValidation,
		},
		{
			name:      "course not found",
			input:     core.CreatePracticeInput{CourseID: courseIDValue, Title: "Practice", Language: "golang", Prompt: "Solve it"},
			wantError: domain.ErrNotFound,
		},
		{
			name:       "invalid language",
			input:      core.CreatePracticeInput{CourseID: courseIDValue, Title: "Practice", Language: "python", Prompt: "Solve it"},
			seedCourse: true,
			wantError:  domain.ErrValidation,
		},
		{
			name:       "missing title",
			input:      core.CreatePracticeInput{CourseID: courseIDValue, Title: "   ", Language: "golang", Prompt: "Solve it"},
			seedCourse: true,
			wantError:  domain.ErrValidation,
		},
		{
			name:       "missing prompt",
			input:      core.CreatePracticeInput{CourseID: courseIDValue, Title: "Practice", Language: "golang", Prompt: "   "},
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
			practices := newPracticeRepositoryFake()
			service := newPracticeServiceFixture(courses, &lessonRepositoryFake{}, practices, fixedClock{})

			_, err := service.CreatePractice(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(practices.savedPractices) != 0 {
				t.Fatalf("expected invalid practice not to be saved")
			}
		})
	}
}

func TestListPracticesValidatesCourseAndReturnsViews(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	practices := newPracticeRepositoryFake()
	first := mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", "solution", mustTestCase(t, testCaseIDValue, 0))
	second := mustPractice(t, otherPracticeID, courseIDValue, "Hello", "javascript", "Prompt", "", "")
	practices.store(first)
	practices.store(second)
	service := newPracticeServiceFixture(courses, &lessonRepositoryFake{}, practices, fixedClock{})

	out, err := service.ListPractices(core.ListPracticesInput{CourseID: courseIDValue})
	if err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}

	if len(out.Practices) != 2 {
		t.Fatalf("expected two practices, got %d", len(out.Practices))
	}
	views := indexPracticeViewsByID(out.Practices)
	if views[practiceIDValue].TestCaseCount != 1 || !views[practiceIDValue].HasSolution {
		t.Fatalf("expected first practice view to include count and solution flag, got %+v", views[practiceIDValue])
	}
	if views[otherPracticeID].Language != "javascript" || views[otherPracticeID].HasSolution {
		t.Fatalf("expected second practice view, got %+v", views[otherPracticeID])
	}
}

func TestListPracticesRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.ListPracticesInput
		wantError error
	}{
		{name: "invalid course id", input: core.ListPracticesInput{CourseID: "bad-id"}, wantError: domain.ErrValidation},
		{name: "course not found", input: core.ListPracticesInput{CourseID: courseIDValue}, wantError: domain.ErrNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newPracticeRepositoryFake(), fixedClock{})

			_, err := service.ListPractices(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
		})
	}
}

func TestGetPracticeReturnsDetailWithOrderedTestCases(t *testing.T) {
	practices := newPracticeRepositoryFake()
	practice := mustPractice(
		t,
		practiceIDValue,
		courseIDValue,
		"FizzBuzz",
		"golang",
		"Print fizz buzz",
		"starter",
		"solution",
		mustTestCase(t, otherTestCaseID, 1),
		mustTestCase(t, testCaseIDValue, 0),
	)
	practices.store(practice)
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, fixedClock{})

	out, err := service.GetPractice(core.GetPracticeInput{ID: practiceIDValue})
	if err != nil {
		t.Fatalf("expected get to succeed, got %v", err)
	}

	if out.Practice.ID != practiceIDValue || out.Practice.TestCaseCount != 2 {
		t.Fatalf("expected practice detail, got %+v", out.Practice)
	}
	if out.Practice.Prompt != "Print fizz buzz" || out.Practice.StarterCode != "starter" || out.Practice.Solution != "solution" {
		t.Fatalf("expected practice content fields, got %+v", out.Practice)
	}
	if out.Practice.TestCases[0].ID != testCaseIDValue || out.Practice.TestCases[0].PracticeID != practiceIDValue {
		t.Fatalf("expected test cases mapped in position order, got %+v", out.Practice.TestCases)
	}
	if !reflect.DeepEqual(out.Practice.TestCases[0], testCaseView(practice.ID(), mustTestCase(t, testCaseIDValue, 0))) {
		t.Fatalf("expected full test case view to be mapped")
	}
}

func TestGetPracticeRejectsFailureModes(t *testing.T) {
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newPracticeRepositoryFake(), fixedClock{})

	if _, err := service.GetPractice(core.GetPracticeInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if _, err := service.GetPractice(core.GetPracticeInput{ID: practiceIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdatePracticeChangesProvidedFields(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC)}
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "starter", "solution"))
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, clock)
	title := "Advanced FizzBuzz"
	prompt := "Updated prompt"
	starterCode := "updated starter"
	solution := "updated solution"

	out, err := service.UpdatePractice(core.UpdatePracticeInput{
		ID:          practiceIDValue,
		Title:       &title,
		Prompt:      &prompt,
		StarterCode: &starterCode,
		Solution:    &solution,
	})
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if out.ID != practiceIDValue {
		t.Fatalf("expected output id %q, got %q", practiceIDValue, out.ID)
	}
	saved := practices.savedPractices[0]
	if saved.Title() != title || saved.Prompt() != prompt || saved.StarterCode() != starterCode || saved.Solution() != solution {
		t.Fatalf("expected practice fields to update")
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestUpdatePracticeRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.UpdatePracticeInput
		seed      bool
		wantError error
	}{
		{name: "invalid id", input: core.UpdatePracticeInput{ID: "bad-id", Title: stringPointer("Updated")}, wantError: domain.ErrValidation},
		{name: "nothing to update", input: core.UpdatePracticeInput{ID: practiceIDValue}, seed: true, wantError: domain.ErrValidation},
		{name: "not found", input: core.UpdatePracticeInput{ID: practiceIDValue, Title: stringPointer("Updated")}, wantError: domain.ErrNotFound},
		{name: "empty title", input: core.UpdatePracticeInput{ID: practiceIDValue, Title: stringPointer("   ")}, seed: true, wantError: domain.ErrValidation},
		{name: "empty prompt", input: core.UpdatePracticeInput{ID: practiceIDValue, Prompt: stringPointer("   ")}, seed: true, wantError: domain.ErrValidation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			practices := newPracticeRepositoryFake()
			if test.seed {
				practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", ""))
			}
			service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, fixedClock{})

			_, err := service.UpdatePractice(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(practices.savedPractices) != 0 {
				t.Fatalf("expected invalid update not to be saved")
			}
		})
	}
}

func TestDeletePracticeDeletesWhenNotEmbedded(t *testing.T) {
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", ""))
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, fixedClock{})

	if err := service.DeletePractice(core.DeletePracticeInput{ID: practiceIDValue}); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	if _, exists := practices.practices[practiceIDValue]; exists {
		t.Fatalf("expected practice to be deleted")
	}
	if len(practices.deletedPracticeIDs) != 1 || practices.deletedPracticeIDs[0].String() != practiceIDValue {
		t.Fatalf("expected practice delete to be recorded")
	}
}

func TestDeletePracticeReturnsPracticeInUseWhenEmbedded(t *testing.T) {
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", ""))
	lessons := newLessonRepositoryFake()
	lessons.store(mustLessonFixture(
		t,
		lessonIDValue,
		courseIDValue,
		"Practice Lesson",
		[]domain.ContentBlock{mustPracticeBlock(t, thirdLessonIDValue, 0, practiceIDValue)},
		0,
	))
	service := newPracticeServiceFixture(newCourseRepositoryFake(), lessons, practices, fixedClock{})

	err := service.DeletePractice(core.DeletePracticeInput{ID: practiceIDValue})
	if !errors.Is(err, domain.ErrPracticeInUse) {
		t.Fatalf("expected practice in use error, got %v", err)
	}

	var inUse domain.PracticeInUseError
	if !errors.As(err, &inUse) {
		t.Fatalf("expected practice in use error details, got %v", err)
	}
	if len(inUse.LessonIDs) != 1 || inUse.LessonIDs[0].String() != lessonIDValue {
		t.Fatalf("expected embedding lesson id, got %+v", inUse.LessonIDs)
	}
	if len(practices.deletedPracticeIDs) != 0 {
		t.Fatalf("expected embedded practice not to be deleted")
	}
}

func TestDeletePracticeRejectsFailureModes(t *testing.T) {
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newPracticeRepositoryFake(), fixedClock{})

	if err := service.DeletePractice(core.DeletePracticeInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if err := service.DeletePractice(core.DeletePracticeInput{ID: practiceIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestAddTestCaseAppendsByDefaultAndAllowsEmptyStrings(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 13, 0, 0, 0, time.UTC)}
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", "", mustTestCase(t, otherTestCaseID, 0)))
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, clock)

	out, err := service.AddTestCase(core.AddTestCaseInput{PracticeID: practiceIDValue})
	if err != nil {
		t.Fatalf("expected add test case to succeed, got %v", err)
	}

	if out.ID != testCaseIDValue {
		t.Fatalf("expected test case id %q, got %q", testCaseIDValue, out.ID)
	}
	saved := practices.savedPractices[0]
	testCases := saved.TestCases()
	if len(testCases) != 2 {
		t.Fatalf("expected two test cases, got %d", len(testCases))
	}
	added := testCases[1]
	if added.ID().String() != testCaseIDValue || added.Position().Int() != 1 {
		t.Fatalf("expected appended test case at position 1, got %+v", added)
	}
	if added.Stdin() != "" || added.ExpectedStdout() != "" || added.Name() != "" {
		t.Fatalf("expected empty stdin, expected stdout, and name to remain valid")
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestAddTestCaseInsertsAtExplicitPosition(t *testing.T) {
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", "", mustTestCase(t, otherTestCaseID, 0)))
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, fixedClock{})
	position := 0

	_, err := service.AddTestCase(core.AddTestCaseInput{
		PracticeID:     practiceIDValue,
		Stdin:          "1\n",
		ExpectedStdout: "1\n",
		Name:           "identity",
		Position:       &position,
	})
	if err != nil {
		t.Fatalf("expected insert to succeed, got %v", err)
	}

	testCases := practices.savedPractices[0].TestCases()
	if testCases[0].ID().String() != testCaseIDValue || testCases[0].Position().Int() != 0 {
		t.Fatalf("expected inserted test case at position 0, got %+v", testCases[0])
	}
	if testCases[1].ID().String() != otherTestCaseID || testCases[1].Position().Int() != 1 {
		t.Fatalf("expected existing test case to shift, got %+v", testCases[1])
	}
}

func TestAddTestCaseRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.AddTestCaseInput
		seed      bool
		wantError error
	}{
		{name: "invalid practice id", input: core.AddTestCaseInput{PracticeID: "bad-id"}, wantError: domain.ErrValidation},
		{name: "practice not found", input: core.AddTestCaseInput{PracticeID: practiceIDValue}, wantError: domain.ErrNotFound},
		{name: "negative position", input: core.AddTestCaseInput{PracticeID: practiceIDValue, Position: intPointer(-1)}, seed: true, wantError: domain.ErrValidation},
		{name: "out of range position", input: core.AddTestCaseInput{PracticeID: practiceIDValue, Position: intPointer(2)}, seed: true, wantError: domain.ErrValidation},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			practices := newPracticeRepositoryFake()
			if test.seed {
				practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", ""))
			}
			service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, fixedClock{})

			_, err := service.AddTestCase(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(practices.savedPractices) != 0 {
				t.Fatalf("expected invalid add not to be saved")
			}
		})
	}
}

func TestListTestCasesReturnsOrderedViews(t *testing.T) {
	practices := newPracticeRepositoryFake()
	practice := mustPractice(
		t,
		practiceIDValue,
		courseIDValue,
		"FizzBuzz",
		"golang",
		"Prompt",
		"",
		"",
		mustTestCase(t, otherTestCaseID, 1),
		mustTestCase(t, testCaseIDValue, 0),
	)
	practices.store(practice)
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, fixedClock{})

	out, err := service.ListTestCases(core.ListTestCasesInput{PracticeID: practiceIDValue})
	if err != nil {
		t.Fatalf("expected list test cases to succeed, got %v", err)
	}

	want := []core.TestCaseView{
		testCaseView(practice.ID(), mustTestCase(t, testCaseIDValue, 0)),
		testCaseView(practice.ID(), mustTestCase(t, otherTestCaseID, 1)),
	}
	if !reflect.DeepEqual(out.TestCases, want) {
		t.Fatalf("expected test case views %+v, got %+v", want, out.TestCases)
	}
}

func TestListTestCasesRejectsFailureModes(t *testing.T) {
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newPracticeRepositoryFake(), fixedClock{})

	if _, err := service.ListTestCases(core.ListTestCasesInput{PracticeID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.ListTestCases(core.ListTestCasesInput{PracticeID: practiceIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGetTestCaseLoadsOwningPractice(t *testing.T) {
	practices := newPracticeRepositoryFake()
	practice := mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", "", mustTestCase(t, testCaseIDValue, 0))
	practices.store(practice)
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, fixedClock{})

	out, err := service.GetTestCase(core.GetTestCaseInput{ID: testCaseIDValue})
	if err != nil {
		t.Fatalf("expected get test case to succeed, got %v", err)
	}

	want := testCaseView(practice.ID(), mustTestCase(t, testCaseIDValue, 0))
	if out.TestCase != want {
		t.Fatalf("expected test case view %+v, got %+v", want, out.TestCase)
	}
}

func TestGetTestCaseRejectsFailureModes(t *testing.T) {
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newPracticeRepositoryFake(), fixedClock{})

	if _, err := service.GetTestCase(core.GetTestCaseInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.GetTestCase(core.GetTestCaseInput{ID: testCaseIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateTestCaseChangesProvidedFieldsAndAllowsEmptyStrings(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 14, 0, 0, 0, time.UTC)}
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", "", mustTestCase(t, testCaseIDValue, 0)))
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, clock)
	empty := ""

	out, err := service.UpdateTestCase(core.UpdateTestCaseInput{
		ID:             testCaseIDValue,
		Stdin:          &empty,
		ExpectedStdout: &empty,
		Name:           &empty,
	})
	if err != nil {
		t.Fatalf("expected update test case to succeed, got %v", err)
	}

	if out.ID != testCaseIDValue {
		t.Fatalf("expected output id %q, got %q", testCaseIDValue, out.ID)
	}
	saved := practices.savedPractices[0]
	testCase, err := saved.TestCase(mustTestCaseID(testCaseIDValue))
	if err != nil {
		t.Fatalf("expected saved test case, got %v", err)
	}
	if testCase.Stdin() != "" || testCase.ExpectedStdout() != "" || testCase.Name() != "" {
		t.Fatalf("expected empty test case strings to be saved")
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestUpdateTestCaseRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.UpdateTestCaseInput
		seed      bool
		wantError error
	}{
		{name: "invalid id", input: core.UpdateTestCaseInput{ID: "bad-id", Name: stringPointer("Updated")}, wantError: domain.ErrValidation},
		{name: "nothing to update", input: core.UpdateTestCaseInput{ID: testCaseIDValue}, seed: true, wantError: domain.ErrValidation},
		{name: "not found", input: core.UpdateTestCaseInput{ID: testCaseIDValue, Name: stringPointer("Updated")}, wantError: domain.ErrNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			practices := newPracticeRepositoryFake()
			if test.seed {
				practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", "", mustTestCase(t, testCaseIDValue, 0)))
			}
			service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, fixedClock{})

			_, err := service.UpdateTestCase(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(practices.savedPractices) != 0 {
				t.Fatalf("expected invalid update not to be saved")
			}
		})
	}
}

func TestRemoveTestCaseCompactsPositions(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 15, 0, 0, 0, time.UTC)}
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(
		t,
		practiceIDValue,
		courseIDValue,
		"FizzBuzz",
		"golang",
		"Prompt",
		"",
		"",
		mustTestCase(t, testCaseIDValue, 0),
		mustTestCase(t, otherTestCaseID, 1),
		mustTestCase(t, thirdTestCaseID, 2),
	))
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, clock)

	if err := service.RemoveTestCase(core.RemoveTestCaseInput{ID: otherTestCaseID}); err != nil {
		t.Fatalf("expected remove test case to succeed, got %v", err)
	}

	saved := practices.savedPractices[0]
	testCases := saved.TestCases()
	if len(testCases) != 2 {
		t.Fatalf("expected two test cases, got %d", len(testCases))
	}
	if testCases[0].ID().String() != testCaseIDValue || testCases[0].Position().Int() != 0 {
		t.Fatalf("expected first test case at position 0, got %+v", testCases[0])
	}
	if testCases[1].ID().String() != thirdTestCaseID || testCases[1].Position().Int() != 1 {
		t.Fatalf("expected third test case compacted to position 1, got %+v", testCases[1])
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestRemoveTestCaseRejectsFailureModes(t *testing.T) {
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, newPracticeRepositoryFake(), fixedClock{})

	if err := service.RemoveTestCase(core.RemoveTestCaseInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if err := service.RemoveTestCase(core.RemoveTestCaseInput{ID: testCaseIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestReorderTestCasesRequiresPermutation(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 16, 0, 0, 0, time.UTC)}
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(
		t,
		practiceIDValue,
		courseIDValue,
		"FizzBuzz",
		"golang",
		"Prompt",
		"",
		"",
		mustTestCase(t, testCaseIDValue, 0),
		mustTestCase(t, otherTestCaseID, 1),
	))
	service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, clock)

	err := service.ReorderTestCases(core.ReorderTestCasesInput{
		PracticeID: practiceIDValue,
		Order: []core.TestCasePlacementDTO{
			{TestCaseID: otherTestCaseID, Position: 0},
			{TestCaseID: testCaseIDValue, Position: 1},
		},
	})
	if err != nil {
		t.Fatalf("expected reorder to succeed, got %v", err)
	}

	saved := practices.savedPractices[0]
	testCases := saved.TestCases()
	if testCases[0].ID().String() != otherTestCaseID || testCases[1].ID().String() != testCaseIDValue {
		t.Fatalf("expected test cases to reorder by position, got %+v", testCases)
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestReorderTestCasesRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.ReorderTestCasesInput
		seed      bool
		wantError error
	}{
		{name: "invalid practice id", input: core.ReorderTestCasesInput{PracticeID: "bad-id"}, wantError: domain.ErrValidation},
		{name: "practice not found", input: core.ReorderTestCasesInput{PracticeID: practiceIDValue}, wantError: domain.ErrNotFound},
		{
			name: "unknown test case",
			input: core.ReorderTestCasesInput{
				PracticeID: practiceIDValue,
				Order: []core.TestCasePlacementDTO{
					{TestCaseID: testCaseIDValue, Position: 0},
					{TestCaseID: thirdTestCaseID, Position: 1},
				},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "negative position",
			input: core.ReorderTestCasesInput{
				PracticeID: practiceIDValue,
				Order:      []core.TestCasePlacementDTO{{TestCaseID: testCaseIDValue, Position: -1}},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "missing test case",
			input: core.ReorderTestCasesInput{
				PracticeID: practiceIDValue,
				Order:      []core.TestCasePlacementDTO{{TestCaseID: testCaseIDValue, Position: 0}},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "duplicate position",
			input: core.ReorderTestCasesInput{
				PracticeID: practiceIDValue,
				Order: []core.TestCasePlacementDTO{
					{TestCaseID: testCaseIDValue, Position: 0},
					{TestCaseID: otherTestCaseID, Position: 0},
				},
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			practices := newPracticeRepositoryFake()
			if test.seed {
				practices.store(mustPractice(
					t,
					practiceIDValue,
					courseIDValue,
					"FizzBuzz",
					"golang",
					"Prompt",
					"",
					"",
					mustTestCase(t, testCaseIDValue, 0),
					mustTestCase(t, otherTestCaseID, 1),
				))
			}
			service := newPracticeServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, practices, fixedClock{})

			err := service.ReorderTestCases(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(practices.savedPractices) != 0 {
				t.Fatalf("expected invalid reorder not to be saved")
			}
		})
	}
}

func newPracticeServiceFixture(
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	practices *practiceRepositoryFake,
	clock fixedClock,
) *PracticeService {
	ids := fixedIDGenerator{
		practiceID: mustPracticeID(practiceIDValue),
		testCaseID: mustTestCaseID(testCaseIDValue),
	}
	return NewPracticeService(courses, lessons, practices, ids, clock)
}

type practiceRepositoryFake struct {
	practices          map[string]domain.Practice
	operations         *[]string
	savedPractices     []domain.Practice
	deletedPracticeIDs []domain.PracticeID
}

func newPracticeRepositoryFake() *practiceRepositoryFake {
	return &practiceRepositoryFake{practices: make(map[string]domain.Practice)}
}

func (repo *practiceRepositoryFake) Save(practice domain.Practice) error {
	repo.savedPractices = append(repo.savedPractices, practice)
	repo.store(practice)

	return nil
}

func (repo *practiceRepositoryFake) FindByID(id domain.PracticeID) (domain.Practice, error) {
	practice, exists := repo.practices[id.String()]
	if !exists {
		return domain.Practice{}, domain.ErrNotFound
	}

	return practice, nil
}

func (repo *practiceRepositoryFake) FindByCourse(courseID domain.CourseID) ([]domain.Practice, error) {
	practices := make([]domain.Practice, 0, len(repo.practices))
	for _, practice := range repo.practices {
		if practice.CourseID() == courseID {
			practices = append(practices, practice)
		}
	}

	sort.Slice(practices, func(i, j int) bool {
		return practices[i].ID().String() < practices[j].ID().String()
	})

	return practices, nil
}

func (repo *practiceRepositoryFake) FindByTestCaseID(id domain.TestCaseID) (domain.Practice, error) {
	for _, practice := range repo.practices {
		if _, err := practice.TestCase(id); err == nil {
			return practice, nil
		}
	}

	return domain.Practice{}, domain.ErrNotFound
}

func (repo *practiceRepositoryFake) Delete(id domain.PracticeID) error {
	if _, exists := repo.practices[id.String()]; !exists {
		return domain.ErrNotFound
	}

	repo.deletedPracticeIDs = append(repo.deletedPracticeIDs, id)
	delete(repo.practices, id.String())

	return nil
}

func (repo *practiceRepositoryFake) DeleteByCourse(courseID domain.CourseID) error {
	if repo.operations != nil {
		*repo.operations = append(*repo.operations, "practices:"+courseID.String())
	}

	for id, practice := range repo.practices {
		if practice.CourseID() == courseID {
			delete(repo.practices, id)
		}
	}

	return nil
}

func (repo *practiceRepositoryFake) store(practice domain.Practice) {
	if repo.practices == nil {
		repo.practices = make(map[string]domain.Practice)
	}

	repo.practices[practice.ID().String()] = practice
}

func mustPractice(
	t *testing.T,
	idValue string,
	courseIDValue string,
	title string,
	languageValue string,
	prompt string,
	starterCode string,
	solution string,
	testCases ...domain.TestCase,
) domain.Practice {
	t.Helper()

	language, err := domain.NewLanguage(languageValue)
	if err != nil {
		t.Fatalf("expected language fixture, got %v", err)
	}

	practice, err := domain.NewPractice(
		mustPracticeID(idValue),
		mustCourseID(courseIDValue),
		title,
		language,
		prompt,
		starterCode,
		solution,
		testCases,
		time.Date(2026, 5, 26, 11, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected practice fixture, got %v", err)
	}

	return practice
}

func mustTestCase(t *testing.T, idValue string, positionValue int) domain.TestCase {
	t.Helper()

	position, err := domain.NewTestCasePosition(positionValue)
	if err != nil {
		t.Fatalf("expected test case position fixture, got %v", err)
	}

	testCase, err := domain.NewTestCase(mustTestCaseID(idValue), "stdin", "stdout", "case", position)
	if err != nil {
		t.Fatalf("expected test case fixture, got %v", err)
	}

	return testCase
}

func mustPracticeBlock(t *testing.T, idValue string, positionValue int, practiceIDValue string) domain.ContentBlock {
	t.Helper()

	block, err := domain.NewPracticeBlock(mustBlockID(idValue), mustBlockPosition(positionValue), mustPracticeID(practiceIDValue))
	if err != nil {
		t.Fatalf("expected practice block fixture, got %v", err)
	}

	return block
}

func mustPracticeID(value string) domain.PracticeID {
	id, err := domain.NewPracticeID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustTestCaseID(value string) domain.TestCaseID {
	id, err := domain.NewTestCaseID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func indexPracticeViewsByID(views []core.PracticeView) map[string]core.PracticeView {
	indexed := make(map[string]core.PracticeView, len(views))
	for _, view := range views {
		indexed[view.ID] = view
	}

	return indexed
}
