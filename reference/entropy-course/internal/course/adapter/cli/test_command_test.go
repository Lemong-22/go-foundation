package cli

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	testIDValue          = "550e8400-e29b-41d4-a716-446655440080"
	otherTestIDValue     = "550e8400-e29b-41d4-a716-446655440081"
	testItemIDValue      = "550e8400-e29b-41d4-a716-446655440090"
	otherTestItemIDValue = "550e8400-e29b-41d4-a716-446655440091"
)

func TestTestCommandExposesRequiredSubcommands(t *testing.T) {
	command := NewTestCommand(TestCommandOptions{Service: &testServiceFake{}})

	wantCommands := [][]string{
		{"create"},
		{"list"},
		{"get"},
		{"update"},
		{"delete"},
		{"item", "add"},
		{"item", "list"},
		{"item", "get"},
		{"item", "update"},
		{"item", "remove"},
		{"item", "reorder"},
	}
	for _, path := range wantCommands {
		if _, _, err := command.Find(path); err != nil {
			t.Fatalf("expected test command path %v to exist, got %v", path, err)
		}
	}
}

func TestTestCreateMapsFlagsToDTO(t *testing.T) {
	service := &testServiceFake{createOut: core.CreateTestOutput{ID: testIDValue}}
	renderer := &testRendererFake{}

	err := executeCourseCommand(
		NewTestCommand(TestCommandOptions{Service: service, Renderer: renderer}),
		"create",
		"--course-id", courseIDValue,
		"--title", "Final Test",
		"--time-limit-minutes", "45",
		"--pass-threshold", "0.8",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "create" || service.createIn.CourseID != courseIDValue || service.createIn.Title != "Final Test" {
		t.Fatalf("expected test create input, got called=%q input=%+v", service.called, service.createIn)
	}
	if service.createIn.TimeLimitMinutes == nil || *service.createIn.TimeLimitMinutes != 45 {
		t.Fatalf("expected time limit pointer, got %v", service.createIn.TimeLimitMinutes)
	}
	if service.createIn.PassThreshold == nil || *service.createIn.PassThreshold != 0.8 {
		t.Fatalf("expected pass threshold pointer, got %v", service.createIn.PassThreshold)
	}
	if renderer.createdTestID != testIDValue {
		t.Fatalf("expected renderer to receive created test id")
	}
}

func TestTestCRUDCommandsMapInputsAndOutputs(t *testing.T) {
	testView := testViewFixture()
	detail := testDetailFixture()
	service := &testServiceFake{
		listOut:   core.ListTestsOutput{Tests: []core.TestView{testView}},
		getOut:    core.GetTestOutput{Test: detail},
		updateOut: core.UpdateTestOutput{ID: testIDValue},
	}

	tests := []struct {
		name     string
		args     []string
		wantCall string
		assert   func(t *testing.T, renderer *testRendererFake)
	}{
		{
			name:     "list",
			args:     []string{"list", "--course-id", courseIDValue, "--output", "json"},
			wantCall: "list",
			assert: func(t *testing.T, renderer *testRendererFake) {
				t.Helper()
				if renderer.testListFormat != "json" || len(renderer.tests) != 1 || renderer.tests[0] != testView {
					t.Fatalf("expected test list renderer, got %+v", renderer)
				}
				if service.listIn.CourseID != courseIDValue {
					t.Fatalf("expected list course id, got %+v", service.listIn)
				}
			},
		},
		{
			name:     "get",
			args:     []string{"get", testIDValue, "-o", "quiet"},
			wantCall: "get",
			assert: func(t *testing.T, renderer *testRendererFake) {
				t.Helper()
				if renderer.testFormat != "quiet" || renderer.test.ID != testIDValue {
					t.Fatalf("expected test renderer, got %+v", renderer)
				}
				if service.getIn.ID != testIDValue {
					t.Fatalf("expected get id, got %+v", service.getIn)
				}
			},
		},
		{
			name: "update",
			args: []string{
				"update",
				testIDValue,
				"--title", "Updated Test",
				"--time-limit-minutes", "0",
				"--pass-threshold", "0.9",
				"--solution-zip-provider", "url",
				"--solution-zip-locator", "https://example.com/solution.zip",
				"--solution-video-provider", "url",
				"--solution-video-locator", "https://example.com/video.mp4",
				"--solution-video-caption", "",
			},
			wantCall: "update",
			assert: func(t *testing.T, renderer *testRendererFake) {
				t.Helper()
				if renderer.updatedTestID != testIDValue {
					t.Fatalf("expected updated test id, got %+v", renderer)
				}
				if service.updateIn.ID != testIDValue || service.updateIn.Title == nil || *service.updateIn.Title != "Updated Test" {
					t.Fatalf("expected update title, got %+v", service.updateIn)
				}
				if service.updateIn.TimeLimitMinutes == nil || *service.updateIn.TimeLimitMinutes != 0 {
					t.Fatalf("expected zero time limit pointer, got %+v", service.updateIn)
				}
				if service.updateIn.SolutionVideoCaption == nil || *service.updateIn.SolutionVideoCaption != "" {
					t.Fatalf("expected empty solution caption pointer, got %+v", service.updateIn)
				}
			},
		},
		{
			name:     "delete",
			args:     []string{"delete", testIDValue, "--force"},
			wantCall: "delete",
			assert: func(t *testing.T, renderer *testRendererFake) {
				t.Helper()
				if renderer.confirmation != "test deleted" {
					t.Fatalf("expected delete confirmation, got %q", renderer.confirmation)
				}
				if service.deleteIn.ID != testIDValue {
					t.Fatalf("expected delete id, got %+v", service.deleteIn)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := &testRendererFake{}
			err := executeCourseCommand(
				NewTestCommand(TestCommandOptions{Service: service, Renderer: renderer}),
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

func TestTestUpdateRejectsPartialSolutionGroupBeforeServiceCall(t *testing.T) {
	service := &testServiceFake{}

	err := executeCourseCommand(
		NewTestCommand(TestCommandOptions{Service: service}),
		"update",
		testIDValue,
		"--solution-zip-provider", "url",
	)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

func TestTestDeleteRequiresConfirmation(t *testing.T) {
	service := &testServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}

	err := executeCourseCommand(
		NewTestCommand(TestCommandOptions{Service: service, Prompter: prompter}),
		"delete",
		testIDValue,
	)
	if !errors.Is(err, ErrConfirmationDeclined) {
		t.Fatalf("expected confirmation declined error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
	if prompter.message == "" {
		t.Fatalf("expected confirmation prompt")
	}
}

func TestTestItemAddMapsChoiceFlagsToDTO(t *testing.T) {
	service := &testServiceFake{addItemOut: core.AddTestItemOutput{ID: testItemIDValue}}
	renderer := &testRendererFake{}

	err := executeCourseCommand(
		NewTestCommand(TestCommandOptions{Service: service, Renderer: renderer}),
		"item",
		"add",
		"--test-id", testIDValue,
		"--kind", "choice",
		"--prompt", "Pick two",
		"--type", "multiple",
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
	if service.called != "add-item" || service.addItemIn.TestID != testIDValue || service.addItemIn.Kind != "choice" {
		t.Fatalf("expected add item input, got called=%q input=%+v", service.called, service.addItemIn)
	}
	if service.addItemIn.Prompt != "Pick two" || service.addItemIn.ChoiceType != "multiple" || service.addItemIn.Explanation != "A and B" {
		t.Fatalf("expected choice fields, got %+v", service.addItemIn)
	}
	if !reflect.DeepEqual(service.addItemIn.Options, wantOptions) || !reflect.DeepEqual(service.addItemIn.CorrectIndices, wantCorrect) {
		t.Fatalf("expected options/correct to map, got %+v", service.addItemIn)
	}
	if service.addItemIn.Position == nil || *service.addItemIn.Position != 1 {
		t.Fatalf("expected explicit position, got %+v", service.addItemIn)
	}
	if renderer.createdTestItemID != testItemIDValue {
		t.Fatalf("expected created item renderer")
	}
}

func TestTestItemAddMapsCodingFlagsAndTestCasesToDTO(t *testing.T) {
	service := &testServiceFake{addItemOut: core.AddTestItemOutput{ID: testItemIDValue}}

	err := executeCourseCommand(
		NewTestCommand(TestCommandOptions{Service: service, Renderer: &testRendererFake{}}),
		"item",
		"add",
		"--test-id", testIDValue,
		"--kind", "coding",
		"--prompt", "Write code",
		"--language", "golang",
		"--starter-code", "package main",
		"--solution", "func main() {}",
		"--testcase", "1::1::sample",
		"--testcase", "::ok",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "add-item" || service.addItemIn.CodingPrompt != "Write code" || service.addItemIn.Language != "golang" {
		t.Fatalf("expected coding fields, got called=%q input=%+v", service.called, service.addItemIn)
	}
	if service.addItemIn.StarterCode != "package main" || service.addItemIn.Solution != "func main() {}" {
		t.Fatalf("expected source fields, got %+v", service.addItemIn)
	}
	wantCases := []core.CodingTestCaseDTO{
		{Stdin: "1", ExpectedStdout: "1", Name: "sample"},
		{Stdin: "", ExpectedStdout: "ok"},
	}
	if !reflect.DeepEqual(service.addItemIn.TestCases, wantCases) {
		t.Fatalf("expected parsed test cases %+v, got %+v", wantCases, service.addItemIn.TestCases)
	}
}

func TestTestItemUpdateMapsChangedFlagsToDTO(t *testing.T) {
	service := &testServiceFake{updateItemOut: core.UpdateTestItemOutput{ID: testItemIDValue}}
	renderer := &testRendererFake{}

	err := executeCourseCommand(
		NewTestCommand(TestCommandOptions{Service: service, Renderer: renderer}),
		"item",
		"update",
		testItemIDValue,
		"--prompt", "Updated prompt",
		"--type", "single",
		"--option", "A",
		"--option", "B",
		"--correct", "1",
		"--explanation", "",
		"--language", "rust",
		"--starter-code", "",
		"--solution", "updated solution",
		"--testcase", "stdin::stdout::case",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "update-item" || service.updateItemIn.ID != testItemIDValue {
		t.Fatalf("expected update item call, got called=%q input=%+v", service.called, service.updateItemIn)
	}
	if service.updateItemIn.Prompt == nil || *service.updateItemIn.Prompt != "Updated prompt" {
		t.Fatalf("expected prompt pointer, got %+v", service.updateItemIn)
	}
	if service.updateItemIn.ChoiceType == nil || *service.updateItemIn.ChoiceType != "single" {
		t.Fatalf("expected choice type pointer, got %+v", service.updateItemIn)
	}
	if service.updateItemIn.Options == nil || !reflect.DeepEqual(*service.updateItemIn.Options, []string{"A", "B"}) {
		t.Fatalf("expected options pointer, got %+v", service.updateItemIn)
	}
	if service.updateItemIn.CorrectIndices == nil || !reflect.DeepEqual(*service.updateItemIn.CorrectIndices, []int{1}) {
		t.Fatalf("expected correct pointer, got %+v", service.updateItemIn)
	}
	if service.updateItemIn.Explanation == nil || *service.updateItemIn.Explanation != "" {
		t.Fatalf("expected empty explanation pointer, got %+v", service.updateItemIn)
	}
	if service.updateItemIn.Language == nil || *service.updateItemIn.Language != "rust" {
		t.Fatalf("expected language pointer, got %+v", service.updateItemIn)
	}
	if service.updateItemIn.StarterCode == nil || *service.updateItemIn.StarterCode != "" {
		t.Fatalf("expected empty starter code pointer, got %+v", service.updateItemIn)
	}
	if service.updateItemIn.TestCases == nil || len(*service.updateItemIn.TestCases) != 1 || (*service.updateItemIn.TestCases)[0].Name != "case" {
		t.Fatalf("expected testcase pointer, got %+v", service.updateItemIn)
	}
	if renderer.updatedTestItemID != testItemIDValue {
		t.Fatalf("expected updated item renderer")
	}
}

func TestTestItemReadRemoveAndReorderCommandsMapDTOs(t *testing.T) {
	item := testItemViewFixture()
	service := &testServiceFake{
		listItemsOut:  core.ListTestItemsOutput{Items: []core.TestItemView{item}},
		getItemOut:    core.GetTestItemOutput{Item: item},
		updateItemOut: core.UpdateTestItemOutput{ID: testItemIDValue},
	}

	tests := []struct {
		name     string
		args     []string
		wantCall string
		assert   func(t *testing.T, renderer *testRendererFake)
	}{
		{
			name:     "list",
			args:     []string{"item", "list", "--test-id", testIDValue, "--output", "json"},
			wantCall: "list-items",
			assert: func(t *testing.T, renderer *testRendererFake) {
				t.Helper()
				if renderer.itemListFormat != "json" || len(renderer.items) != 1 {
					t.Fatalf("expected item list renderer, got %+v", renderer)
				}
				if service.listItemsIn.TestID != testIDValue {
					t.Fatalf("expected list test id, got %+v", service.listItemsIn)
				}
			},
		},
		{
			name:     "get",
			args:     []string{"item", "get", testItemIDValue, "-o", "quiet"},
			wantCall: "get-item",
			assert: func(t *testing.T, renderer *testRendererFake) {
				t.Helper()
				if renderer.itemFormat != "quiet" || renderer.item.ID != testItemIDValue {
					t.Fatalf("expected item renderer, got %+v", renderer)
				}
				if service.getItemIn.ID != testItemIDValue {
					t.Fatalf("expected get item id, got %+v", service.getItemIn)
				}
			},
		},
		{
			name:     "remove",
			args:     []string{"item", "remove", testItemIDValue, "--force"},
			wantCall: "remove-item",
			assert: func(t *testing.T, renderer *testRendererFake) {
				t.Helper()
				if renderer.confirmation != "test item removed" {
					t.Fatalf("expected remove confirmation, got %q", renderer.confirmation)
				}
				if service.removeItemIn.ID != testItemIDValue {
					t.Fatalf("expected remove item id, got %+v", service.removeItemIn)
				}
			},
		},
		{
			name: "reorder",
			args: []string{
				"item",
				"reorder",
				"--test-id", testIDValue,
				"--order", testItemIDValue + ":1," + otherTestItemIDValue + ":0",
			},
			wantCall: "reorder-items",
			assert: func(t *testing.T, renderer *testRendererFake) {
				t.Helper()
				if renderer.confirmation != "test items reordered" {
					t.Fatalf("expected reorder confirmation, got %q", renderer.confirmation)
				}
				want := []core.TestItemPlacementDTO{
					{TestItemID: testItemIDValue, Position: 1},
					{TestItemID: otherTestItemIDValue, Position: 0},
				}
				if service.reorderItemsIn.TestID != testIDValue || !reflect.DeepEqual(service.reorderItemsIn.Order, want) {
					t.Fatalf("expected reorder input %+v, got %+v", want, service.reorderItemsIn)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := &testRendererFake{}
			err := executeCourseCommand(
				NewTestCommand(TestCommandOptions{Service: service, Renderer: renderer}),
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

func TestTestItemRemoveRequiresConfirmation(t *testing.T) {
	service := &testServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}

	err := executeCourseCommand(
		NewTestCommand(TestCommandOptions{Service: service, Prompter: prompter}),
		"item",
		"remove",
		testItemIDValue,
	)
	if !errors.Is(err, ErrConfirmationDeclined) {
		t.Fatalf("expected confirmation declined error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
	if prompter.message == "" {
		t.Fatalf("expected confirmation prompt")
	}
}

func TestTestItemAddRejectsInvalidTestCaseFormat(t *testing.T) {
	service := &testServiceFake{}

	err := executeCourseCommand(
		NewTestCommand(TestCommandOptions{Service: service}),
		"item",
		"add",
		"--test-id", testIDValue,
		"--kind", "coding",
		"--prompt", "Solve",
		"--language", "golang",
		"--testcase", "bad",
	)
	if !errors.Is(err, ErrInvalidCodingTestCase) {
		t.Fatalf("expected invalid testcase error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

func TestTestItemReorderRejectsInvalidOrder(t *testing.T) {
	service := &testServiceFake{}

	err := executeCourseCommand(
		NewTestCommand(TestCommandOptions{Service: service}),
		"item",
		"reorder",
		"--test-id", testIDValue,
		"--order", testItemIDValue,
	)
	if !errors.Is(err, ErrInvalidTestItemOrder) {
		t.Fatalf("expected invalid item order, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

type testServiceFake struct {
	called    string
	callCount int
	err       error

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
	service.record("create")
	service.createIn = in
	if service.err != nil {
		return core.CreateTestOutput{}, service.err
	}

	return service.createOut, nil
}

func (service *testServiceFake) ListTests(in core.ListTestsInput) (core.ListTestsOutput, error) {
	service.record("list")
	service.listIn = in
	if service.err != nil {
		return core.ListTestsOutput{}, service.err
	}

	return service.listOut, nil
}

func (service *testServiceFake) GetTest(in core.GetTestInput) (core.GetTestOutput, error) {
	service.record("get")
	service.getIn = in
	if service.err != nil {
		return core.GetTestOutput{}, service.err
	}

	return service.getOut, nil
}

func (service *testServiceFake) UpdateTest(in core.UpdateTestInput) (core.UpdateTestOutput, error) {
	service.record("update")
	service.updateIn = in
	if service.err != nil {
		return core.UpdateTestOutput{}, service.err
	}

	return service.updateOut, nil
}

func (service *testServiceFake) DeleteTest(in core.DeleteTestInput) error {
	service.record("delete")
	service.deleteIn = in
	return service.err
}

func (service *testServiceFake) AddTestItem(in core.AddTestItemInput) (core.AddTestItemOutput, error) {
	service.record("add-item")
	service.addItemIn = in
	if service.err != nil {
		return core.AddTestItemOutput{}, service.err
	}

	return service.addItemOut, nil
}

func (service *testServiceFake) ListTestItems(in core.ListTestItemsInput) (core.ListTestItemsOutput, error) {
	service.record("list-items")
	service.listItemsIn = in
	if service.err != nil {
		return core.ListTestItemsOutput{}, service.err
	}

	return service.listItemsOut, nil
}

func (service *testServiceFake) GetTestItem(in core.GetTestItemInput) (core.GetTestItemOutput, error) {
	service.record("get-item")
	service.getItemIn = in
	if service.err != nil {
		return core.GetTestItemOutput{}, service.err
	}

	return service.getItemOut, nil
}

func (service *testServiceFake) UpdateTestItem(in core.UpdateTestItemInput) (core.UpdateTestItemOutput, error) {
	service.record("update-item")
	service.updateItemIn = in
	if service.err != nil {
		return core.UpdateTestItemOutput{}, service.err
	}

	return service.updateItemOut, nil
}

func (service *testServiceFake) RemoveTestItem(in core.RemoveTestItemInput) error {
	service.record("remove-item")
	service.removeItemIn = in
	return service.err
}

func (service *testServiceFake) ReorderTestItems(in core.ReorderTestItemsInput) error {
	service.record("reorder-items")
	service.reorderItemsIn = in
	return service.err
}

func (service *testServiceFake) record(called string) {
	service.called = called
	service.callCount++
}

type testRendererFake struct {
	createdTestID     string
	updatedTestID     string
	createdTestItemID string
	updatedTestItemID string
	testListFormat    string
	testFormat        string
	itemListFormat    string
	itemFormat        string
	tests             []core.TestView
	test              core.TestDetailView
	items             []core.TestItemView
	item              core.TestItemView
	confirmation      string
}

func (renderer *testRendererFake) RenderCreatedTest(id string) error {
	renderer.createdTestID = id
	return nil
}

func (renderer *testRendererFake) RenderTestList(format string, tests []core.TestView) error {
	renderer.testListFormat = format
	renderer.tests = tests
	return nil
}

func (renderer *testRendererFake) RenderTest(format string, test core.TestDetailView) error {
	renderer.testFormat = format
	renderer.test = test
	return nil
}

func (renderer *testRendererFake) RenderUpdatedTest(id string) error {
	renderer.updatedTestID = id
	return nil
}

func (renderer *testRendererFake) RenderCreatedTestItem(id string) error {
	renderer.createdTestItemID = id
	return nil
}

func (renderer *testRendererFake) RenderTestItemList(format string, items []core.TestItemView) error {
	renderer.itemListFormat = format
	renderer.items = items
	return nil
}

func (renderer *testRendererFake) RenderTestItem(format string, item core.TestItemView) error {
	renderer.itemFormat = format
	renderer.item = item
	return nil
}

func (renderer *testRendererFake) RenderUpdatedTestItem(id string) error {
	renderer.updatedTestItemID = id
	return nil
}

func (renderer *testRendererFake) RenderConfirmation(message string) error {
	renderer.confirmation = message
	return nil
}

func testViewFixture() core.TestView {
	now := time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC)
	timeLimit := 45
	return core.TestView{
		ID:               testIDValue,
		CourseID:         courseIDValue,
		Title:            "Final Test",
		TimeLimitMinutes: &timeLimit,
		PassThreshold:    0.7,
		HasSolution:      true,
		ItemCount:        1,
		CreatedAt:        now,
		UpdatedAt:        now,
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
