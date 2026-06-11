package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

const (
	testIDValue      = "550e8400-e29b-41d4-a716-446655440014"
	testItemIDValue  = "550e8400-e29b-41d4-a716-446655440015"
	otherTestItemID  = "550e8400-e29b-41d4-a716-446655440016"
	thirdTestItemID  = "550e8400-e29b-41d4-a716-446655440017"
	fourthTestItemID = "550e8400-e29b-41d4-a716-446655440018"
)

func TestNewTestCreatesAggregate(t *testing.T) {
	now := time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC)
	timeLimit := mustTimeLimit(t, 90)
	solution := mustTestSolution(t)
	items := []TestItem{
		mustCodingTestItem(t, otherTestItemID, 1),
		mustChoiceTestItem(t, testItemIDValue, 0),
	}

	test, err := NewTest(
		mustTestID(t, testIDValue),
		mustCourseID(t, validUUID),
		"  Final Assessment  ",
		&timeLimit,
		DefaultPassThreshold(),
		&solution,
		items,
		now,
	)
	if err != nil {
		t.Fatalf("expected test, got error %v", err)
	}

	if test.ID().String() != testIDValue {
		t.Fatalf("expected test id %q, got %q", testIDValue, test.ID().String())
	}
	if test.CourseID().String() != validUUID {
		t.Fatalf("expected course id %q, got %q", validUUID, test.CourseID().String())
	}
	if test.Title() != "Final Assessment" {
		t.Fatalf("expected trimmed title, got %q", test.Title())
	}
	if got := test.TimeLimit(); got == nil || got.Minutes() != 90 {
		t.Fatalf("expected time limit 90, got %+v", got)
	}
	if test.PassThreshold() != DefaultPassThreshold() {
		t.Fatalf("expected default pass threshold")
	}
	if got := test.Solution(); got == nil || got.ExplanationCaption() != "Walkthrough" {
		t.Fatalf("expected solution package, got %+v", got)
	}
	if got := test.Items(); len(got) != 2 || got[0].ID().String() != testItemIDValue || got[1].ID().String() != otherTestItemID {
		t.Fatalf("expected items ordered by position, got %+v", got)
	}
	if !test.CreatedAt().Equal(now) || !test.UpdatedAt().Equal(now) {
		t.Fatalf("expected created and updated timestamps to equal %v", now)
	}
}

func TestNewTestAllowsEmptyItemsAndOptionalFields(t *testing.T) {
	test, err := NewTest(
		mustTestID(t, testIDValue),
		mustCourseID(t, validUUID),
		"Draft Assessment",
		nil,
		DefaultPassThreshold(),
		nil,
		nil,
		time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected empty draft test to be valid, got %v", err)
	}

	if test.TimeLimit() != nil {
		t.Fatalf("expected nil time limit for untimed test")
	}
	if test.Solution() != nil {
		t.Fatalf("expected nil solution during authoring")
	}
	if len(test.Items()) != 0 {
		t.Fatalf("expected no item-count invariant")
	}
}

func TestRestoreTestRejectsInvalidState(t *testing.T) {
	now := time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC)
	invalidTimeLimit := TimeLimit{}

	tests := []struct {
		name      string
		id        TestID
		courseID  CourseID
		title     string
		timeLimit *TimeLimit
		items     []TestItem
		createdAt time.Time
		updatedAt time.Time
	}{
		{
			name:      "empty test id",
			id:        TestID{},
			courseID:  mustCourseID(t, validUUID),
			title:     "Test",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "empty course id",
			id:        mustTestID(t, testIDValue),
			courseID:  CourseID{},
			title:     "Test",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "empty title",
			id:        mustTestID(t, testIDValue),
			courseID:  mustCourseID(t, validUUID),
			title:     "   ",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "invalid time limit",
			id:        mustTestID(t, testIDValue),
			courseID:  mustCourseID(t, validUUID),
			title:     "Test",
			timeLimit: &invalidTimeLimit,
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "updated before created",
			id:        mustTestID(t, testIDValue),
			courseID:  mustCourseID(t, validUUID),
			title:     "Test",
			createdAt: now,
			updatedAt: now.Add(-time.Minute),
		},
		{
			name:     "duplicate item ids",
			id:       mustTestID(t, testIDValue),
			courseID: mustCourseID(t, validUUID),
			title:    "Test",
			items: []TestItem{
				mustChoiceTestItem(t, testItemIDValue, 0),
				mustChoiceTestItem(t, testItemIDValue, 1),
			},
			createdAt: now,
			updatedAt: now,
		},
		{
			name:     "duplicate item positions",
			id:       mustTestID(t, testIDValue),
			courseID: mustCourseID(t, validUUID),
			title:    "Test",
			items: []TestItem{
				mustChoiceTestItem(t, testItemIDValue, 0),
				mustChoiceTestItem(t, otherTestItemID, 0),
			},
			createdAt: now,
			updatedAt: now,
		},
		{
			name:     "item position gap",
			id:       mustTestID(t, testIDValue),
			courseID: mustCourseID(t, validUUID),
			title:    "Test",
			items: []TestItem{
				mustChoiceTestItem(t, testItemIDValue, 0),
				mustChoiceTestItem(t, otherTestItemID, 2),
			},
			createdAt: now,
			updatedAt: now,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := RestoreTest(
				test.id,
				test.courseID,
				test.title,
				test.timeLimit,
				DefaultPassThreshold(),
				nil,
				test.items,
				test.createdAt,
				test.updatedAt,
			)
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestTestValueObjects(t *testing.T) {
	kind, err := NewTestItemKind("choice")
	if err != nil {
		t.Fatalf("expected choice kind, got error %v", err)
	}
	if kind != ChoiceKind() || !kind.IsChoice() || kind.IsCoding() || kind.String() != "choice" {
		t.Fatalf("expected choice kind, got %q", kind.String())
	}

	coding := CodingKind()
	if !coding.IsCoding() || coding.IsChoice() || coding.String() != "coding" {
		t.Fatalf("expected coding kind, got %q", coding.String())
	}
	if _, err := NewTestItemKind("essay"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid kind validation error, got %v", err)
	}

	position, err := NewTestItemPosition(0)
	if err != nil {
		t.Fatalf("expected position, got error %v", err)
	}
	if position.Int() != 0 {
		t.Fatalf("expected position 0, got %d", position.Int())
	}
	if _, err := NewTestItemPosition(-1); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected negative position validation error, got %v", err)
	}

	limit, err := NewTimeLimit(30)
	if err != nil {
		t.Fatalf("expected time limit, got error %v", err)
	}
	if limit.Minutes() != 30 {
		t.Fatalf("expected 30 minutes, got %d", limit.Minutes())
	}
	if _, err := NewTimeLimit(0); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected zero time limit validation error, got %v", err)
	}
}

func TestTestSolutionValidatesRequiredMediaRefs(t *testing.T) {
	zip := mustMediaRef(t, URLProvider(), "https://example.com/solution.zip")
	video := mustMediaRef(t, YouTubeProvider(), "dQw4w9WgXcQ")

	solution, err := NewTestSolution(zip, video, "Caption")
	if err != nil {
		t.Fatalf("expected solution, got error %v", err)
	}
	if solution.SolutionZip() != zip || solution.ExplanationVideo() != video || solution.ExplanationCaption() != "Caption" {
		t.Fatalf("expected solution refs and caption to be retained")
	}

	if _, err := NewTestSolution(MediaRef{}, video, ""); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected missing zip validation error, got %v", err)
	}
	if _, err := NewTestSolution(zip, MediaRef{}, ""); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected missing video validation error, got %v", err)
	}
}

func TestChoiceItemBodyValidatesQuestionContent(t *testing.T) {
	valid, err := NewChoiceItemBody(
		SingleChoice(),
		"  Pick one  ",
		[]string{"A", "B"},
		[]int{0},
		"Because A",
	)
	if err != nil {
		t.Fatalf("expected choice body, got error %v", err)
	}
	if valid.Kind() != ChoiceKind() || valid.Prompt() != "Pick one" || valid.Explanation() != "Because A" {
		t.Fatalf("expected choice body to normalize prompt and report choice kind")
	}
	if !reflect.DeepEqual(valid.Options(), []string{"A", "B"}) || !reflect.DeepEqual(valid.CorrectIndices(), []int{0}) {
		t.Fatalf("expected options and correct indices to be retained")
	}

	tests := []struct {
		name           string
		questionType   ChoiceQuestionType
		prompt         string
		options        []string
		correctIndices []int
	}{
		{name: "empty prompt", questionType: SingleChoice(), prompt: "   ", options: []string{"A", "B"}, correctIndices: []int{0}},
		{name: "too few options", questionType: SingleChoice(), prompt: "Pick one", options: []string{"A"}, correctIndices: []int{0}},
		{name: "no correct indices", questionType: SingleChoice(), prompt: "Pick one", options: []string{"A", "B"}},
		{name: "out of range correct index", questionType: SingleChoice(), prompt: "Pick one", options: []string{"A", "B"}, correctIndices: []int{2}},
		{name: "duplicate correct index", questionType: MultipleChoice(), prompt: "Pick many", options: []string{"A", "B"}, correctIndices: []int{0, 0}},
		{name: "single choice with multiple correct indices", questionType: SingleChoice(), prompt: "Pick one", options: []string{"A", "B"}, correctIndices: []int{0, 1}},
		{name: "invalid type", questionType: ChoiceQuestionType{}, prompt: "Pick one", options: []string{"A", "B"}, correctIndices: []int{0}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewChoiceItemBody(test.questionType, test.prompt, test.options, test.correctIndices, "")
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestCodingItemBodyValidatesPracticeContentAndBundledCases(t *testing.T) {
	emptyCase := NewCodingTestCase("", "", "")
	if emptyCase.Stdin() != "" || emptyCase.ExpectedStdout() != "" || emptyCase.Name() != "" {
		t.Fatalf("expected coding test case to allow empty strings")
	}

	body, err := NewCodingItemBody(
		JavaScript(),
		"  Print hello  ",
		"console.log('')",
		"console.log('hello')",
		[]CodingTestCase{NewCodingTestCase("stdin", "stdout", "sample")},
	)
	if err != nil {
		t.Fatalf("expected coding body, got error %v", err)
	}
	if body.Kind() != CodingKind() || body.Language() != JavaScript() || body.Prompt() != "Print hello" {
		t.Fatalf("expected coding body to normalize prompt and report coding kind")
	}
	if body.StarterCode() != "console.log('')" || body.Solution() != "console.log('hello')" {
		t.Fatalf("expected starter code and solution to be retained")
	}
	if cases := body.TestCases(); len(cases) != 1 || cases[0].Name() != "sample" {
		t.Fatalf("expected bundled test cases, got %+v", cases)
	}

	tests := []struct {
		name      string
		language  Language
		prompt    string
		testCases []CodingTestCase
	}{
		{name: "invalid language", language: Language{}, prompt: "Solve", testCases: []CodingTestCase{emptyCase}},
		{name: "empty prompt", language: JavaScript(), prompt: "   ", testCases: []CodingTestCase{emptyCase}},
		{name: "no test cases", language: JavaScript(), prompt: "Solve"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewCodingItemBody(test.language, test.prompt, "", "", test.testCases)
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestTestItemRequiresMatchingKindAndBody(t *testing.T) {
	choiceBody := mustChoiceItemBody(t)

	item, err := NewTestItem(
		mustTestItemID(t, testItemIDValue),
		ChoiceKind(),
		choiceBody,
		mustTestItemPosition(t, 0),
	)
	if err != nil {
		t.Fatalf("expected test item, got error %v", err)
	}
	if item.ID().String() != testItemIDValue || item.Kind() != ChoiceKind() || item.Position().Int() != 0 {
		t.Fatalf("expected item identity, kind, and position to be retained")
	}

	if _, err := NewTestItem(TestItemID{}, ChoiceKind(), choiceBody, mustTestItemPosition(t, 0)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected empty id validation error, got %v", err)
	}
	if _, err := NewTestItem(mustTestItemID(t, otherTestItemID), CodingKind(), choiceBody, mustTestItemPosition(t, 0)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected body kind mismatch validation error, got %v", err)
	}
	if _, err := NewTestItem(mustTestItemID(t, otherTestItemID), ChoiceKind(), nil, mustTestItemPosition(t, 0)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected nil body validation error, got %v", err)
	}
}

func TestTestMutationsUpdateMetadataAndTimestamp(t *testing.T) {
	test := mustTestWithItems(t)
	renamedAt := test.CreatedAt().Add(time.Hour)

	if err := test.Rename("  Updated Assessment  ", renamedAt); err != nil {
		t.Fatalf("expected rename to succeed, got %v", err)
	}
	if test.Title() != "Updated Assessment" || !test.UpdatedAt().Equal(renamedAt) {
		t.Fatalf("expected title and timestamp to update")
	}

	updatedAt := test.UpdatedAt()
	if err := test.Rename("   ", updatedAt.Add(time.Hour)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected empty title validation error, got %v", err)
	}
	if test.Title() != "Updated Assessment" || !test.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected invalid rename to leave state unchanged")
	}

	limit := mustTimeLimit(t, 45)
	limitChangedAt := updatedAt.Add(2 * time.Hour)
	test.ChangeTimeLimit(&limit, limitChangedAt)
	if got := test.TimeLimit(); got == nil || got.Minutes() != 45 || !test.UpdatedAt().Equal(limitChangedAt) {
		t.Fatalf("expected time limit and timestamp to update")
	}

	clearedAt := limitChangedAt.Add(time.Hour)
	test.ChangeTimeLimit(nil, clearedAt)
	if test.TimeLimit() != nil || !test.UpdatedAt().Equal(clearedAt) {
		t.Fatalf("expected time limit to clear and timestamp to update")
	}

	threshold := mustPassThreshold(t, 0.85)
	thresholdChangedAt := clearedAt.Add(time.Hour)
	test.ChangePassThreshold(threshold, thresholdChangedAt)
	if test.PassThreshold() != threshold || !test.UpdatedAt().Equal(thresholdChangedAt) {
		t.Fatalf("expected pass threshold and timestamp to update")
	}

	solution := mustTestSolution(t)
	solutionChangedAt := thresholdChangedAt.Add(time.Hour)
	test.SetSolution(solution, solutionChangedAt)
	if got := test.Solution(); got == nil || got.ExplanationCaption() != "Walkthrough" || !test.UpdatedAt().Equal(solutionChangedAt) {
		t.Fatalf("expected solution and timestamp to update")
	}
}

func TestTestAddItemInsertsAndShiftsPositions(t *testing.T) {
	test := mustTestWithItems(t,
		mustChoiceTestItem(t, testItemIDValue, 0),
		mustCodingTestItem(t, otherTestItemID, 1),
	)
	inserted := mustChoiceTestItem(t, thirdTestItemID, 1)
	changedAt := test.CreatedAt().Add(time.Hour)

	if err := test.AddItem(inserted, changedAt); err != nil {
		t.Fatalf("expected add item to succeed, got %v", err)
	}

	items := test.Items()
	if len(items) != 3 {
		t.Fatalf("expected three items, got %d", len(items))
	}
	if items[0].ID().String() != testItemIDValue || items[0].Position().Int() != 0 {
		t.Fatalf("expected original first item to stay at position 0")
	}
	if items[1].ID().String() != thirdTestItemID || items[1].Position().Int() != 1 {
		t.Fatalf("expected inserted item at position 1, got %+v", items[1])
	}
	if items[2].ID().String() != otherTestItemID || items[2].Position().Int() != 2 {
		t.Fatalf("expected original second item to shift to position 2")
	}
	if !test.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}
}

func TestTestAddItemRejectsDuplicateIDAndOutOfRangePosition(t *testing.T) {
	test := mustTestWithItems(t, mustChoiceTestItem(t, testItemIDValue, 0))

	if err := test.AddItem(mustChoiceTestItem(t, testItemIDValue, 1), time.Now()); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected duplicate id validation error, got %v", err)
	}

	if err := test.AddItem(mustChoiceTestItem(t, otherTestItemID, 2), time.Now()); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected out-of-range position validation error, got %v", err)
	}
}

func TestTestRemoveItemCompactsPositions(t *testing.T) {
	test := mustTestWithItems(t,
		mustChoiceTestItem(t, testItemIDValue, 0),
		mustCodingTestItem(t, otherTestItemID, 1),
		mustChoiceTestItem(t, thirdTestItemID, 2),
	)
	changedAt := test.CreatedAt().Add(time.Hour)

	if err := test.RemoveItem(mustTestItemID(t, otherTestItemID), changedAt); err != nil {
		t.Fatalf("expected remove to succeed, got %v", err)
	}

	items := test.Items()
	if len(items) != 2 {
		t.Fatalf("expected two items, got %d", len(items))
	}
	if items[0].ID().String() != testItemIDValue || items[0].Position().Int() != 0 {
		t.Fatalf("expected first item at position 0")
	}
	if items[1].ID().String() != thirdTestItemID || items[1].Position().Int() != 1 {
		t.Fatalf("expected third item compacted to position 1")
	}
	if !test.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}
}

func TestTestReorderItemsRequiresPermutation(t *testing.T) {
	test := mustTestWithItems(t,
		mustChoiceTestItem(t, testItemIDValue, 0),
		mustCodingTestItem(t, otherTestItemID, 1),
	)
	changedAt := test.CreatedAt().Add(time.Hour)

	err := test.ReorderItems([]TestItemPlacement{
		{TestItemID: mustTestItemID(t, otherTestItemID), Position: mustTestItemPosition(t, 0)},
		{TestItemID: mustTestItemID(t, testItemIDValue), Position: mustTestItemPosition(t, 1)},
	}, changedAt)
	if err != nil {
		t.Fatalf("expected reorder to succeed, got %v", err)
	}

	items := test.Items()
	if items[0].ID().String() != otherTestItemID || items[1].ID().String() != testItemIDValue {
		t.Fatalf("expected items to reorder by position")
	}
	if !test.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}

	tests := []struct {
		name  string
		order []TestItemPlacement
	}{
		{
			name:  "missing item",
			order: []TestItemPlacement{{TestItemID: mustTestItemID(t, testItemIDValue), Position: mustTestItemPosition(t, 0)}},
		},
		{
			name: "unknown item",
			order: []TestItemPlacement{
				{TestItemID: mustTestItemID(t, testItemIDValue), Position: mustTestItemPosition(t, 0)},
				{TestItemID: mustTestItemID(t, thirdTestItemID), Position: mustTestItemPosition(t, 1)},
			},
		},
		{
			name: "duplicate item",
			order: []TestItemPlacement{
				{TestItemID: mustTestItemID(t, testItemIDValue), Position: mustTestItemPosition(t, 0)},
				{TestItemID: mustTestItemID(t, testItemIDValue), Position: mustTestItemPosition(t, 1)},
			},
		},
		{
			name: "duplicate position",
			order: []TestItemPlacement{
				{TestItemID: mustTestItemID(t, testItemIDValue), Position: mustTestItemPosition(t, 0)},
				{TestItemID: mustTestItemID(t, otherTestItemID), Position: mustTestItemPosition(t, 0)},
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			test := mustTestWithItems(t,
				mustChoiceTestItem(t, testItemIDValue, 0),
				mustCodingTestItem(t, otherTestItemID, 1),
			)
			if err := test.ReorderItems(testCase.order, changedAt); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestTestReplaceItemBodyPreservesKindAndTimestamp(t *testing.T) {
	test := mustTestWithItems(t, mustChoiceTestItem(t, testItemIDValue, 0))
	changedAt := test.CreatedAt().Add(time.Hour)
	itemID := mustTestItemID(t, testItemIDValue)
	body, err := NewChoiceItemBody(MultipleChoice(), "Pick many", []string{"A", "B"}, []int{0, 1}, "Both")
	if err != nil {
		t.Fatalf("expected replacement body, got error %v", err)
	}

	if err := test.ReplaceItemBody(itemID, body, changedAt); err != nil {
		t.Fatalf("expected replace body to succeed, got %v", err)
	}
	item, err := test.Item(itemID)
	if err != nil {
		t.Fatalf("expected item to exist, got %v", err)
	}
	got := item.Body().(ChoiceItemBody)
	if got.Type() != MultipleChoice() || got.Prompt() != "Pick many" || !test.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected body and timestamp to update")
	}

	updatedAt := test.UpdatedAt()
	if err := test.ReplaceItemBody(itemID, mustCodingItemBody(t), updatedAt.Add(time.Hour)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected kind mismatch validation error, got %v", err)
	}
	item, err = test.Item(itemID)
	if err != nil {
		t.Fatalf("expected item to exist, got %v", err)
	}
	if item.Body().(ChoiceItemBody).Prompt() != "Pick many" || !test.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected invalid body replacement to leave state unchanged")
	}

	if err := test.ReplaceItemBody(mustTestItemID(t, fourthTestItemID), body, updatedAt.Add(time.Hour)); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing item error, got %v", err)
	}
}

func mustTestWithItems(t *testing.T, items ...TestItem) Test {
	t.Helper()

	test, err := NewTest(
		mustTestID(t, testIDValue),
		mustCourseID(t, validUUID),
		"Final Assessment",
		nil,
		DefaultPassThreshold(),
		nil,
		items,
		time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected test fixture, got error %v", err)
	}

	return test
}

func mustChoiceTestItem(t *testing.T, id string, position int) TestItem {
	t.Helper()

	item, err := NewTestItem(
		mustTestItemID(t, id),
		ChoiceKind(),
		mustChoiceItemBody(t),
		mustTestItemPosition(t, position),
	)
	if err != nil {
		t.Fatalf("expected choice item fixture, got error %v", err)
	}

	return item
}

func mustCodingTestItem(t *testing.T, id string, position int) TestItem {
	t.Helper()

	item, err := NewTestItem(
		mustTestItemID(t, id),
		CodingKind(),
		mustCodingItemBody(t),
		mustTestItemPosition(t, position),
	)
	if err != nil {
		t.Fatalf("expected coding item fixture, got error %v", err)
	}

	return item
}

func mustChoiceItemBody(t *testing.T) ChoiceItemBody {
	t.Helper()

	body, err := NewChoiceItemBody(
		SingleChoice(),
		"Pick one",
		[]string{"A", "B"},
		[]int{0},
		"",
	)
	if err != nil {
		t.Fatalf("expected choice item body fixture, got error %v", err)
	}

	return body
}

func mustCodingItemBody(t *testing.T) CodingItemBody {
	t.Helper()

	body, err := NewCodingItemBody(
		JavaScript(),
		"Write a program that prints hello.",
		"",
		"",
		[]CodingTestCase{NewCodingTestCase("stdin", "stdout", "sample")},
	)
	if err != nil {
		t.Fatalf("expected coding item body fixture, got error %v", err)
	}

	return body
}

func mustTestID(t *testing.T, value string) TestID {
	t.Helper()

	id, err := NewTestID(value)
	if err != nil {
		t.Fatalf("expected test id fixture, got error %v", err)
	}

	return id
}

func mustTestItemID(t *testing.T, value string) TestItemID {
	t.Helper()

	id, err := NewTestItemID(value)
	if err != nil {
		t.Fatalf("expected test item id fixture, got error %v", err)
	}

	return id
}

func mustTestItemPosition(t *testing.T, value int) TestItemPosition {
	t.Helper()

	position, err := NewTestItemPosition(value)
	if err != nil {
		t.Fatalf("expected test item position fixture, got error %v", err)
	}

	return position
}

func mustTimeLimit(t *testing.T, minutes int) TimeLimit {
	t.Helper()

	limit, err := NewTimeLimit(minutes)
	if err != nil {
		t.Fatalf("expected time limit fixture, got error %v", err)
	}

	return limit
}

func mustPassThreshold(t *testing.T, value float64) PassThreshold {
	t.Helper()

	threshold, err := NewPassThreshold(value)
	if err != nil {
		t.Fatalf("expected pass threshold fixture, got error %v", err)
	}

	return threshold
}

func mustTestSolution(t *testing.T) TestSolution {
	t.Helper()

	solution, err := NewTestSolution(
		mustMediaRef(t, URLProvider(), "https://example.com/solution.zip"),
		mustMediaRef(t, YouTubeProvider(), "dQw4w9WgXcQ"),
		"Walkthrough",
	)
	if err != nil {
		t.Fatalf("expected test solution fixture, got error %v", err)
	}

	return solution
}
