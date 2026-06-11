package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

const (
	practiceIDValue = "550e8400-e29b-41d4-a716-446655440010"
	testCaseIDValue = "550e8400-e29b-41d4-a716-446655440011"
	otherTestCaseID = "550e8400-e29b-41d4-a716-446655440012"
	thirdTestCaseID = "550e8400-e29b-41d4-a716-446655440013"
)

func TestNewPracticeCreatesAggregate(t *testing.T) {
	now := time.Date(2026, 5, 26, 8, 0, 0, 0, time.UTC)
	testCases := []TestCase{mustTestCase(t, testCaseIDValue, 0)}

	practice, err := NewPractice(
		mustPracticeID(t, practiceIDValue),
		mustCourseID(t, validUUID),
		"  Print Hello  ",
		JavaScript(),
		"  Write a program that prints hello.  ",
		"console.log('')",
		"console.log('hello')",
		testCases,
		now,
	)
	if err != nil {
		t.Fatalf("expected practice, got error %v", err)
	}

	if practice.ID().String() != practiceIDValue {
		t.Fatalf("expected practice id %q, got %q", practiceIDValue, practice.ID().String())
	}
	if practice.CourseID().String() != validUUID {
		t.Fatalf("expected course id %q, got %q", validUUID, practice.CourseID().String())
	}
	if practice.Title() != "Print Hello" {
		t.Fatalf("expected trimmed title, got %q", practice.Title())
	}
	if practice.Language() != JavaScript() {
		t.Fatalf("expected javascript language, got %q", practice.Language().String())
	}
	if practice.Prompt() != "Write a program that prints hello." {
		t.Fatalf("expected trimmed prompt, got %q", practice.Prompt())
	}
	if practice.StarterCode() != "console.log('')" || practice.Solution() != "console.log('hello')" {
		t.Fatalf("expected starter code and solution to be retained")
	}
	if got := practice.TestCases(); !reflect.DeepEqual(got, testCases) {
		t.Fatalf("expected test cases %+v, got %+v", testCases, got)
	}
	if !practice.CreatedAt().Equal(now) || !practice.UpdatedAt().Equal(now) {
		t.Fatalf("expected created and updated timestamps to equal %v", now)
	}
}

func TestNewPracticeRejectsInvalidState(t *testing.T) {
	now := time.Date(2026, 5, 26, 8, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		id        PracticeID
		courseID  CourseID
		title     string
		language  Language
		prompt    string
		testCases []TestCase
		createdAt time.Time
		updatedAt time.Time
	}{
		{
			name:      "empty practice id",
			id:        PracticeID{},
			courseID:  mustCourseID(t, validUUID),
			title:     "Practice",
			language:  JavaScript(),
			prompt:    "Solve it",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "empty course id",
			id:        mustPracticeID(t, practiceIDValue),
			courseID:  CourseID{},
			title:     "Practice",
			language:  JavaScript(),
			prompt:    "Solve it",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "empty title",
			id:        mustPracticeID(t, practiceIDValue),
			courseID:  mustCourseID(t, validUUID),
			title:     "   ",
			language:  JavaScript(),
			prompt:    "Solve it",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "invalid language",
			id:        mustPracticeID(t, practiceIDValue),
			courseID:  mustCourseID(t, validUUID),
			title:     "Practice",
			language:  Language{},
			prompt:    "Solve it",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "empty prompt",
			id:        mustPracticeID(t, practiceIDValue),
			courseID:  mustCourseID(t, validUUID),
			title:     "Practice",
			language:  JavaScript(),
			prompt:    "   ",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "updated before created",
			id:        mustPracticeID(t, practiceIDValue),
			courseID:  mustCourseID(t, validUUID),
			title:     "Practice",
			language:  JavaScript(),
			prompt:    "Solve it",
			createdAt: now,
			updatedAt: now.Add(-time.Minute),
		},
		{
			name:     "duplicate test case ids",
			id:       mustPracticeID(t, practiceIDValue),
			courseID: mustCourseID(t, validUUID),
			title:    "Practice",
			language: JavaScript(),
			prompt:   "Solve it",
			testCases: []TestCase{
				mustTestCase(t, testCaseIDValue, 0),
				mustTestCase(t, testCaseIDValue, 1),
			},
			createdAt: now,
			updatedAt: now,
		},
		{
			name:     "duplicate test case positions",
			id:       mustPracticeID(t, practiceIDValue),
			courseID: mustCourseID(t, validUUID),
			title:    "Practice",
			language: JavaScript(),
			prompt:   "Solve it",
			testCases: []TestCase{
				mustTestCase(t, testCaseIDValue, 0),
				mustTestCase(t, otherTestCaseID, 0),
			},
			createdAt: now,
			updatedAt: now,
		},
		{
			name:     "test case position gap",
			id:       mustPracticeID(t, practiceIDValue),
			courseID: mustCourseID(t, validUUID),
			title:    "Practice",
			language: JavaScript(),
			prompt:   "Solve it",
			testCases: []TestCase{
				mustTestCase(t, testCaseIDValue, 0),
				mustTestCase(t, otherTestCaseID, 2),
			},
			createdAt: now,
			updatedAt: now,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := RestorePractice(
				test.id,
				test.courseID,
				test.title,
				test.language,
				test.prompt,
				"",
				"",
				test.testCases,
				test.createdAt,
				test.updatedAt,
			)
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestNewTestCaseAllowsEmptyIOAndName(t *testing.T) {
	testCase, err := NewTestCase(
		mustTestCaseID(t, testCaseIDValue),
		"",
		"",
		"",
		mustTestCasePosition(t, 0),
	)
	if err != nil {
		t.Fatalf("expected empty input/output test case to be valid, got %v", err)
	}

	if testCase.Stdin() != "" || testCase.ExpectedStdout() != "" || testCase.Name() != "" {
		t.Fatalf("expected empty strings to be retained")
	}

	if _, err := NewTestCase(TestCaseID{}, "", "", "", mustTestCasePosition(t, 0)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected empty id validation error, got %v", err)
	}
}

func TestPracticeAddTestCaseInsertsAndShiftsPositions(t *testing.T) {
	practice := mustPracticeWithTestCases(t,
		mustTestCase(t, testCaseIDValue, 0),
		mustTestCase(t, otherTestCaseID, 1),
	)
	inserted := mustTestCase(t, thirdTestCaseID, 1)
	changedAt := practice.CreatedAt().Add(time.Hour)

	if err := practice.AddTestCase(inserted, changedAt); err != nil {
		t.Fatalf("expected add test case to succeed, got %v", err)
	}

	testCases := practice.TestCases()
	if len(testCases) != 3 {
		t.Fatalf("expected three test cases, got %d", len(testCases))
	}
	if testCases[0].ID().String() != testCaseIDValue || testCases[0].Position().Int() != 0 {
		t.Fatalf("expected original first test case to stay at position 0")
	}
	if testCases[1].ID().String() != thirdTestCaseID || testCases[1].Position().Int() != 1 {
		t.Fatalf("expected inserted test case at position 1, got %+v", testCases[1])
	}
	if testCases[2].ID().String() != otherTestCaseID || testCases[2].Position().Int() != 2 {
		t.Fatalf("expected original second test case to shift to position 2")
	}
	if !practice.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}
}

func TestPracticeAddTestCaseRejectsDuplicateIDAndOutOfRangePosition(t *testing.T) {
	practice := mustPracticeWithTestCases(t, mustTestCase(t, testCaseIDValue, 0))

	if err := practice.AddTestCase(mustTestCase(t, testCaseIDValue, 1), time.Now()); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected duplicate id validation error, got %v", err)
	}

	if err := practice.AddTestCase(mustTestCase(t, otherTestCaseID, 2), time.Now()); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected out-of-range position validation error, got %v", err)
	}
}

func TestPracticeRemoveTestCaseCompactsPositions(t *testing.T) {
	practice := mustPracticeWithTestCases(t,
		mustTestCase(t, testCaseIDValue, 0),
		mustTestCase(t, otherTestCaseID, 1),
		mustTestCase(t, thirdTestCaseID, 2),
	)
	changedAt := practice.CreatedAt().Add(time.Hour)

	if err := practice.RemoveTestCase(mustTestCaseID(t, otherTestCaseID), changedAt); err != nil {
		t.Fatalf("expected remove to succeed, got %v", err)
	}

	testCases := practice.TestCases()
	if len(testCases) != 2 {
		t.Fatalf("expected two test cases, got %d", len(testCases))
	}
	if testCases[0].ID().String() != testCaseIDValue || testCases[0].Position().Int() != 0 {
		t.Fatalf("expected first test case at position 0")
	}
	if testCases[1].ID().String() != thirdTestCaseID || testCases[1].Position().Int() != 1 {
		t.Fatalf("expected third test case compacted to position 1")
	}
	if !practice.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}
}

func TestPracticeReorderTestCasesRequiresPermutation(t *testing.T) {
	practice := mustPracticeWithTestCases(t,
		mustTestCase(t, testCaseIDValue, 0),
		mustTestCase(t, otherTestCaseID, 1),
	)
	changedAt := practice.CreatedAt().Add(time.Hour)

	err := practice.ReorderTestCases([]TestCasePlacement{
		{TestCaseID: mustTestCaseID(t, otherTestCaseID), Position: mustTestCasePosition(t, 0)},
		{TestCaseID: mustTestCaseID(t, testCaseIDValue), Position: mustTestCasePosition(t, 1)},
	}, changedAt)
	if err != nil {
		t.Fatalf("expected reorder to succeed, got %v", err)
	}

	testCases := practice.TestCases()
	if testCases[0].ID().String() != otherTestCaseID || testCases[1].ID().String() != testCaseIDValue {
		t.Fatalf("expected test cases to reorder by position")
	}
	if !practice.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}

	tests := []struct {
		name  string
		order []TestCasePlacement
	}{
		{
			name:  "missing test case",
			order: []TestCasePlacement{{TestCaseID: mustTestCaseID(t, testCaseIDValue), Position: mustTestCasePosition(t, 0)}},
		},
		{
			name: "unknown test case",
			order: []TestCasePlacement{
				{TestCaseID: mustTestCaseID(t, testCaseIDValue), Position: mustTestCasePosition(t, 0)},
				{TestCaseID: mustTestCaseID(t, thirdTestCaseID), Position: mustTestCasePosition(t, 1)},
			},
		},
		{
			name: "duplicate test case",
			order: []TestCasePlacement{
				{TestCaseID: mustTestCaseID(t, testCaseIDValue), Position: mustTestCasePosition(t, 0)},
				{TestCaseID: mustTestCaseID(t, testCaseIDValue), Position: mustTestCasePosition(t, 1)},
			},
		},
		{
			name: "duplicate position",
			order: []TestCasePlacement{
				{TestCaseID: mustTestCaseID(t, testCaseIDValue), Position: mustTestCasePosition(t, 0)},
				{TestCaseID: mustTestCaseID(t, otherTestCaseID), Position: mustTestCasePosition(t, 0)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			practice := mustPracticeWithTestCases(t,
				mustTestCase(t, testCaseIDValue, 0),
				mustTestCase(t, otherTestCaseID, 1),
			)
			if err := practice.ReorderTestCases(test.order, changedAt); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestPracticeMutationsUpdateThroughAggregate(t *testing.T) {
	practice := mustPracticeWithTestCases(t, mustTestCase(t, testCaseIDValue, 0))
	changedAt := practice.CreatedAt().Add(time.Hour)
	testCaseID := mustTestCaseID(t, testCaseIDValue)

	if err := practice.Rename("  Updated Practice  ", changedAt); err != nil {
		t.Fatalf("expected rename to succeed, got %v", err)
	}
	if practice.Title() != "Updated Practice" || !practice.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected title and timestamp to update")
	}

	promptChangedAt := changedAt.Add(time.Hour)
	if err := practice.ChangePrompt("  Updated prompt  ", promptChangedAt); err != nil {
		t.Fatalf("expected prompt change to succeed, got %v", err)
	}
	if practice.Prompt() != "Updated prompt" || !practice.UpdatedAt().Equal(promptChangedAt) {
		t.Fatalf("expected prompt and timestamp to update")
	}

	updatedAt := practice.UpdatedAt()
	if err := practice.ChangePrompt("   ", updatedAt.Add(time.Hour)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected empty prompt validation error, got %v", err)
	}
	if practice.Prompt() != "Updated prompt" || !practice.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected invalid prompt to leave state unchanged")
	}

	starterChangedAt := updatedAt.Add(2 * time.Hour)
	practice.ChangeStarterCode("console.log('start')", starterChangedAt)
	if practice.StarterCode() != "console.log('start')" || !practice.UpdatedAt().Equal(starterChangedAt) {
		t.Fatalf("expected starter code and timestamp to update")
	}

	solutionChangedAt := starterChangedAt.Add(time.Hour)
	practice.ChangeSolution("console.log('done')", solutionChangedAt)
	if practice.Solution() != "console.log('done')" || !practice.UpdatedAt().Equal(solutionChangedAt) {
		t.Fatalf("expected solution and timestamp to update")
	}

	stdinChangedAt := solutionChangedAt.Add(time.Hour)
	if err := practice.ChangeTestCaseStdin(testCaseID, "input", stdinChangedAt); err != nil {
		t.Fatalf("expected stdin change to succeed, got %v", err)
	}
	testCase, err := practice.TestCase(testCaseID)
	if err != nil {
		t.Fatalf("expected test case to exist, got %v", err)
	}
	if testCase.Stdin() != "input" || !practice.UpdatedAt().Equal(stdinChangedAt) {
		t.Fatalf("expected stdin and timestamp to update")
	}

	expectedChangedAt := stdinChangedAt.Add(time.Hour)
	if err := practice.ChangeTestCaseExpectedStdout(testCaseID, "output", expectedChangedAt); err != nil {
		t.Fatalf("expected expected stdout change to succeed, got %v", err)
	}
	testCase, _ = practice.TestCase(testCaseID)
	if testCase.ExpectedStdout() != "output" || !practice.UpdatedAt().Equal(expectedChangedAt) {
		t.Fatalf("expected expected stdout and timestamp to update")
	}

	nameChangedAt := expectedChangedAt.Add(time.Hour)
	if err := practice.ChangeTestCaseName(testCaseID, "sample", nameChangedAt); err != nil {
		t.Fatalf("expected name change to succeed, got %v", err)
	}
	testCase, _ = practice.TestCase(testCaseID)
	if testCase.Name() != "sample" || !practice.UpdatedAt().Equal(nameChangedAt) {
		t.Fatalf("expected name and timestamp to update")
	}

	if err := practice.ChangeTestCaseName(mustTestCaseID(t, otherTestCaseID), "missing", nameChangedAt.Add(time.Hour)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing test case error, got %v", err)
	}
}

func TestLanguageValueObject(t *testing.T) {
	tests := []struct {
		value string
		want  Language
	}{
		{value: "javascript", want: JavaScript()},
		{value: "typescript", want: TypeScript()},
		{value: "golang", want: Golang()},
		{value: "rust", want: Rust()},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			language, err := NewLanguage(test.value)
			if err != nil {
				t.Fatalf("expected language, got error %v", err)
			}
			if language != test.want || language.String() != test.value {
				t.Fatalf("expected language %q, got %q", test.want.String(), language.String())
			}
		})
	}

	if !JavaScript().IsJavaScript() || !TypeScript().IsTypeScript() || !Golang().IsGolang() || !Rust().IsRust() {
		t.Fatalf("expected per-language predicates to identify each language")
	}

	if _, err := NewLanguage("python"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected unsupported language validation error, got %v", err)
	}
}

func TestPracticeBodyKind(t *testing.T) {
	practiceID := mustPracticeID(t, practiceIDValue)
	body := PracticeBody{PracticeRef: practiceID}

	if body.Kind() != PracticeKind() || body.PracticeRef != practiceID {
		t.Fatalf("expected practice body to keep practice ref and report practice kind")
	}

	block, err := NewPracticeBlock(mustBlockID(t, blockIDValue), mustBlockPosition(t, 0), practiceID)
	if err != nil {
		t.Fatalf("expected practice block, got error %v", err)
	}
	if block.Kind() != PracticeKind() {
		t.Fatalf("expected practice block kind, got %q", block.Kind().String())
	}
}

func mustPracticeWithTestCases(t *testing.T, testCases ...TestCase) Practice {
	t.Helper()

	practice, err := NewPractice(
		mustPracticeID(t, practiceIDValue),
		mustCourseID(t, validUUID),
		"Print Hello",
		JavaScript(),
		"Write a program that prints hello.",
		"",
		"",
		testCases,
		time.Date(2026, 5, 26, 8, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected practice fixture, got error %v", err)
	}

	return practice
}

func mustTestCase(t *testing.T, id string, position int) TestCase {
	t.Helper()

	testCase, err := NewTestCase(
		mustTestCaseID(t, id),
		"stdin",
		"stdout",
		"case",
		mustTestCasePosition(t, position),
	)
	if err != nil {
		t.Fatalf("expected test case fixture, got error %v", err)
	}

	return testCase
}

func mustPracticeID(t *testing.T, value string) PracticeID {
	t.Helper()

	id, err := NewPracticeID(value)
	if err != nil {
		t.Fatalf("expected practice id fixture, got error %v", err)
	}

	return id
}

func mustTestCaseID(t *testing.T, value string) TestCaseID {
	t.Helper()

	id, err := NewTestCaseID(value)
	if err != nil {
		t.Fatalf("expected test case id fixture, got error %v", err)
	}

	return id
}

func mustTestCasePosition(t *testing.T, value int) TestCasePosition {
	t.Helper()

	position, err := NewTestCasePosition(value)
	if err != nil {
		t.Fatalf("expected test case position fixture, got error %v", err)
	}

	return position
}
