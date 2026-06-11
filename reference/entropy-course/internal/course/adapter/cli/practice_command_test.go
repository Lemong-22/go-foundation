package cli

import (
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	testCaseIDValue      = "550e8400-e29b-41d4-a716-446655440070"
	otherTestCaseIDValue = "550e8400-e29b-41d4-a716-446655440071"
)

func TestPracticeCommandExposesRequiredSubcommands(t *testing.T) {
	command := NewPracticeCommand(PracticeCommandOptions{Service: &practiceServiceFake{}})

	wantCommands := [][]string{
		{"create"},
		{"list"},
		{"get"},
		{"update"},
		{"delete"},
		{"testcase", "add"},
		{"testcase", "list"},
		{"testcase", "get"},
		{"testcase", "update"},
		{"testcase", "remove"},
		{"testcase", "reorder"},
	}
	for _, path := range wantCommands {
		if _, _, err := command.Find(path); err != nil {
			t.Fatalf("expected practice command path %v to exist, got %v", path, err)
		}
	}
}

func TestPracticeCreateMapsFlagsToDTO(t *testing.T) {
	service := &practiceServiceFake{createOut: core.CreatePracticeOutput{ID: practiceIDValue}}
	renderer := &practiceRendererFake{}

	err := executeCourseCommand(
		NewPracticeCommand(PracticeCommandOptions{Service: service, Renderer: renderer}),
		"create",
		"--course-id", courseIDValue,
		"--title", "FizzBuzz",
		"--language", "golang",
		"--prompt", "Print fizz buzz",
		"--starter-code", "package main",
		"--solution", "fmt.Println()",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "create" || service.createIn.CourseID != courseIDValue || service.createIn.Title != "FizzBuzz" {
		t.Fatalf("expected practice create input, got called=%q input=%+v", service.called, service.createIn)
	}
	if service.createIn.Language != "golang" || service.createIn.Prompt != "Print fizz buzz" {
		t.Fatalf("expected language and prompt to map, got %+v", service.createIn)
	}
	if service.createIn.StarterCode != "package main" || service.createIn.Solution != "fmt.Println()" {
		t.Fatalf("expected source fields to map, got %+v", service.createIn)
	}
	if renderer.createdPracticeID != practiceIDValue {
		t.Fatalf("expected renderer to receive created practice id")
	}
}

func TestPracticeCreateRejectsInvalidLanguageBeforeServiceCall(t *testing.T) {
	service := &practiceServiceFake{}

	err := executeCourseCommand(
		NewPracticeCommand(PracticeCommandOptions{Service: service}),
		"create",
		"--course-id", courseIDValue,
		"--title", "FizzBuzz",
		"--language", "python",
		"--prompt", "Print fizz buzz",
	)
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

func TestPracticeCRUDCommandsMapInputsAndOutputs(t *testing.T) {
	practice := practiceViewFixture()
	detail := practiceDetailFixture()
	service := &practiceServiceFake{
		listOut:   core.ListPracticesOutput{Practices: []core.PracticeView{practice}},
		getOut:    core.GetPracticeOutput{Practice: detail},
		updateOut: core.UpdatePracticeOutput{ID: practiceIDValue},
	}

	tests := []struct {
		name     string
		args     []string
		wantCall string
		assert   func(t *testing.T, renderer *practiceRendererFake)
	}{
		{
			name:     "list",
			args:     []string{"list", "--course-id", courseIDValue, "--output", "json"},
			wantCall: "list",
			assert: func(t *testing.T, renderer *practiceRendererFake) {
				t.Helper()
				if renderer.practiceListFormat != "json" || len(renderer.practices) != 1 || renderer.practices[0] != practice {
					t.Fatalf("expected practice list renderer, got %+v", renderer)
				}
				if service.listIn.CourseID != courseIDValue {
					t.Fatalf("expected list course id, got %+v", service.listIn)
				}
			},
		},
		{
			name:     "get",
			args:     []string{"get", practiceIDValue, "-o", "quiet"},
			wantCall: "get",
			assert: func(t *testing.T, renderer *practiceRendererFake) {
				t.Helper()
				if renderer.practiceFormat != "quiet" || renderer.practice.ID != practiceIDValue {
					t.Fatalf("expected practice renderer, got %+v", renderer)
				}
				if service.getIn.ID != practiceIDValue {
					t.Fatalf("expected get id, got %+v", service.getIn)
				}
			},
		},
		{
			name: "update",
			args: []string{
				"update",
				practiceIDValue,
				"--title", "Advanced FizzBuzz",
				"--prompt", "Updated prompt",
				"--starter-code", "",
				"--solution", "updated solution",
			},
			wantCall: "update",
			assert: func(t *testing.T, renderer *practiceRendererFake) {
				t.Helper()
				if renderer.updatedPracticeID != practiceIDValue {
					t.Fatalf("expected updated practice id, got %+v", renderer)
				}
				if service.updateIn.ID != practiceIDValue || service.updateIn.Title == nil || *service.updateIn.Title != "Advanced FizzBuzz" {
					t.Fatalf("expected update title, got %+v", service.updateIn)
				}
				if service.updateIn.Prompt == nil || *service.updateIn.Prompt != "Updated prompt" {
					t.Fatalf("expected update prompt, got %+v", service.updateIn)
				}
				if service.updateIn.StarterCode == nil || *service.updateIn.StarterCode != "" {
					t.Fatalf("expected empty starter code pointer, got %+v", service.updateIn)
				}
				if service.updateIn.Solution == nil || *service.updateIn.Solution != "updated solution" {
					t.Fatalf("expected update solution, got %+v", service.updateIn)
				}
			},
		},
		{
			name:     "delete",
			args:     []string{"delete", practiceIDValue, "--force"},
			wantCall: "delete",
			assert: func(t *testing.T, renderer *practiceRendererFake) {
				t.Helper()
				if renderer.confirmation != "practice deleted" {
					t.Fatalf("expected delete confirmation, got %q", renderer.confirmation)
				}
				if service.deleteIn.ID != practiceIDValue {
					t.Fatalf("expected delete id, got %+v", service.deleteIn)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := &practiceRendererFake{}
			err := executeCourseCommand(
				NewPracticeCommand(PracticeCommandOptions{Service: service, Renderer: renderer}),
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

func TestPracticeDeleteRequiresConfirmation(t *testing.T) {
	service := &practiceServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}

	err := executeCourseCommand(
		NewPracticeCommand(PracticeCommandOptions{Service: service, Prompter: prompter}),
		"delete",
		practiceIDValue,
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

func TestPracticeDeletePrintsEmbeddingLessonIDs(t *testing.T) {
	service := &practiceServiceFake{err: practiceInUseError(t)}
	var stderr bytes.Buffer
	command := NewPracticeCommand(PracticeCommandOptions{Service: service})
	command.SetArgs([]string{"delete", practiceIDValue, "--force"})
	command.SetOut(io.Discard)
	command.SetErr(&stderr)
	command.SilenceUsage = true
	command.SilenceErrors = true

	err := command.Execute()
	if !errors.Is(err, domain.ErrPracticeInUse) {
		t.Fatalf("expected practice in use error, got %v", err)
	}
	if !strings.Contains(stderr.String(), lessonIDValue) || !strings.Contains(stderr.String(), otherLessonIDValue) {
		t.Fatalf("expected embedded lesson ids in stderr, got %q", stderr.String())
	}
}

func TestTestCaseAddMapsFlagsToDTOAndAllowsEmptyValues(t *testing.T) {
	position := 1
	service := &practiceServiceFake{addTestCaseOut: core.AddTestCaseOutput{ID: testCaseIDValue}}
	renderer := &practiceRendererFake{}

	err := executeCourseCommand(
		NewPracticeCommand(PracticeCommandOptions{Service: service, Renderer: renderer}),
		"testcase",
		"add",
		"--practice-id", practiceIDValue,
		"--stdin", "",
		"--expected-stdout", "",
		"--name", "",
		"--position", "1",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "add-testcase" || service.addTestCaseIn.PracticeID != practiceIDValue {
		t.Fatalf("expected add test case input, got called=%q input=%+v", service.called, service.addTestCaseIn)
	}
	if service.addTestCaseIn.Stdin != "" || service.addTestCaseIn.ExpectedStdout != "" || service.addTestCaseIn.Name != "" {
		t.Fatalf("expected empty test case strings to map, got %+v", service.addTestCaseIn)
	}
	if service.addTestCaseIn.Position == nil || *service.addTestCaseIn.Position != position {
		t.Fatalf("expected explicit position, got %v", service.addTestCaseIn.Position)
	}
	if renderer.createdTestCaseID != testCaseIDValue {
		t.Fatalf("expected renderer to receive created test case id")
	}
}

func TestTestCaseUpdateMapsChangedFlagsToDTO(t *testing.T) {
	service := &practiceServiceFake{updateTestCaseOut: core.UpdateTestCaseOutput{ID: testCaseIDValue}}
	renderer := &practiceRendererFake{}

	err := executeCourseCommand(
		NewPracticeCommand(PracticeCommandOptions{Service: service, Renderer: renderer}),
		"testcase",
		"update",
		testCaseIDValue,
		"--stdin", "updated input",
		"--expected-stdout", "",
		"--name", "",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "update-testcase" || service.updateTestCaseIn.ID != testCaseIDValue {
		t.Fatalf("expected update test case call, got called=%q input=%+v", service.called, service.updateTestCaseIn)
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
	if renderer.updatedTestCaseID != testCaseIDValue {
		t.Fatalf("expected renderer to receive updated test case id")
	}
}

func TestTestCaseRemoveRequiresConfirmation(t *testing.T) {
	service := &practiceServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}

	err := executeCourseCommand(
		NewPracticeCommand(PracticeCommandOptions{Service: service, Prompter: prompter}),
		"testcase",
		"remove",
		testCaseIDValue,
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

func TestTestCaseReadRemoveAndReorderCommandsMapDTOs(t *testing.T) {
	testCase := testCaseViewFixture()
	service := &practiceServiceFake{
		listTestCasesOut:  core.ListTestCasesOutput{TestCases: []core.TestCaseView{testCase}},
		getTestCaseOut:    core.GetTestCaseOutput{TestCase: testCase},
		updateTestCaseOut: core.UpdateTestCaseOutput{ID: testCaseIDValue},
	}

	tests := []struct {
		name     string
		args     []string
		wantCall string
		assert   func(t *testing.T, renderer *practiceRendererFake)
	}{
		{
			name:     "list",
			args:     []string{"testcase", "list", "--practice-id", practiceIDValue, "--output", "json"},
			wantCall: "list-testcases",
			assert: func(t *testing.T, renderer *practiceRendererFake) {
				t.Helper()
				if renderer.testCaseListFormat != "json" || len(renderer.testCases) != 1 {
					t.Fatalf("expected test case list renderer, got %+v", renderer)
				}
				if service.listTestCasesIn.PracticeID != practiceIDValue {
					t.Fatalf("expected list practice id, got %+v", service.listTestCasesIn)
				}
			},
		},
		{
			name:     "get",
			args:     []string{"testcase", "get", testCaseIDValue, "-o", "quiet"},
			wantCall: "get-testcase",
			assert: func(t *testing.T, renderer *practiceRendererFake) {
				t.Helper()
				if renderer.testCaseFormat != "quiet" || renderer.testCase.ID != testCaseIDValue {
					t.Fatalf("expected test case renderer, got %+v", renderer)
				}
				if service.getTestCaseIn.ID != testCaseIDValue {
					t.Fatalf("expected get test case id, got %+v", service.getTestCaseIn)
				}
			},
		},
		{
			name:     "remove",
			args:     []string{"testcase", "remove", testCaseIDValue, "--force"},
			wantCall: "remove-testcase",
			assert: func(t *testing.T, renderer *practiceRendererFake) {
				t.Helper()
				if renderer.confirmation != "practice test case removed" {
					t.Fatalf("expected remove confirmation, got %q", renderer.confirmation)
				}
				if service.removeTestCaseIn.ID != testCaseIDValue {
					t.Fatalf("expected remove test case id, got %+v", service.removeTestCaseIn)
				}
			},
		},
		{
			name: "reorder",
			args: []string{
				"testcase",
				"reorder",
				"--practice-id", practiceIDValue,
				"--order", testCaseIDValue + ":1," + otherTestCaseIDValue + ":0",
			},
			wantCall: "reorder-testcases",
			assert: func(t *testing.T, renderer *practiceRendererFake) {
				t.Helper()
				if renderer.confirmation != "practice test cases reordered" {
					t.Fatalf("expected reorder confirmation, got %q", renderer.confirmation)
				}
				want := []core.TestCasePlacementDTO{
					{TestCaseID: testCaseIDValue, Position: 1},
					{TestCaseID: otherTestCaseIDValue, Position: 0},
				}
				if service.reorderTestCasesIn.PracticeID != practiceIDValue || !reflect.DeepEqual(service.reorderTestCasesIn.Order, want) {
					t.Fatalf("expected reorder input %+v, got %+v", want, service.reorderTestCasesIn)
				}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			renderer := &practiceRendererFake{}
			err := executeCourseCommand(
				NewPracticeCommand(PracticeCommandOptions{Service: service, Renderer: renderer}),
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

func TestTestCaseReorderRejectsInvalidOrder(t *testing.T) {
	service := &practiceServiceFake{}

	err := executeCourseCommand(
		NewPracticeCommand(PracticeCommandOptions{Service: service}),
		"testcase",
		"reorder",
		"--practice-id", practiceIDValue,
		"--order", testCaseIDValue,
	)
	if !errors.Is(err, ErrInvalidTestCaseOrder) {
		t.Fatalf("expected invalid test case order, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

type practiceServiceFake struct {
	called    string
	callCount int
	err       error

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
	service.record("create")
	service.createIn = in
	if service.err != nil {
		return core.CreatePracticeOutput{}, service.err
	}

	return service.createOut, nil
}

func (service *practiceServiceFake) ListPractices(in core.ListPracticesInput) (core.ListPracticesOutput, error) {
	service.record("list")
	service.listIn = in
	if service.err != nil {
		return core.ListPracticesOutput{}, service.err
	}

	return service.listOut, nil
}

func (service *practiceServiceFake) GetPractice(in core.GetPracticeInput) (core.GetPracticeOutput, error) {
	service.record("get")
	service.getIn = in
	if service.err != nil {
		return core.GetPracticeOutput{}, service.err
	}

	return service.getOut, nil
}

func (service *practiceServiceFake) UpdatePractice(in core.UpdatePracticeInput) (core.UpdatePracticeOutput, error) {
	service.record("update")
	service.updateIn = in
	if service.err != nil {
		return core.UpdatePracticeOutput{}, service.err
	}

	return service.updateOut, nil
}

func (service *practiceServiceFake) DeletePractice(in core.DeletePracticeInput) error {
	service.record("delete")
	service.deleteIn = in
	return service.err
}

func (service *practiceServiceFake) AddTestCase(in core.AddTestCaseInput) (core.AddTestCaseOutput, error) {
	service.record("add-testcase")
	service.addTestCaseIn = in
	if service.err != nil {
		return core.AddTestCaseOutput{}, service.err
	}

	return service.addTestCaseOut, nil
}

func (service *practiceServiceFake) ListTestCases(in core.ListTestCasesInput) (core.ListTestCasesOutput, error) {
	service.record("list-testcases")
	service.listTestCasesIn = in
	if service.err != nil {
		return core.ListTestCasesOutput{}, service.err
	}

	return service.listTestCasesOut, nil
}

func (service *practiceServiceFake) GetTestCase(in core.GetTestCaseInput) (core.GetTestCaseOutput, error) {
	service.record("get-testcase")
	service.getTestCaseIn = in
	if service.err != nil {
		return core.GetTestCaseOutput{}, service.err
	}

	return service.getTestCaseOut, nil
}

func (service *practiceServiceFake) UpdateTestCase(in core.UpdateTestCaseInput) (core.UpdateTestCaseOutput, error) {
	service.record("update-testcase")
	service.updateTestCaseIn = in
	if service.err != nil {
		return core.UpdateTestCaseOutput{}, service.err
	}

	return service.updateTestCaseOut, nil
}

func (service *practiceServiceFake) RemoveTestCase(in core.RemoveTestCaseInput) error {
	service.record("remove-testcase")
	service.removeTestCaseIn = in
	return service.err
}

func (service *practiceServiceFake) ReorderTestCases(in core.ReorderTestCasesInput) error {
	service.record("reorder-testcases")
	service.reorderTestCasesIn = in
	return service.err
}

func (service *practiceServiceFake) record(called string) {
	service.called = called
	service.callCount++
}

type practiceRendererFake struct {
	createdPracticeID  string
	updatedPracticeID  string
	createdTestCaseID  string
	updatedTestCaseID  string
	practiceListFormat string
	practiceFormat     string
	testCaseListFormat string
	testCaseFormat     string
	practices          []core.PracticeView
	practice           core.PracticeDetailView
	testCases          []core.TestCaseView
	testCase           core.TestCaseView
	confirmation       string
}

func (renderer *practiceRendererFake) RenderCreatedPractice(id string) error {
	renderer.createdPracticeID = id
	return nil
}

func (renderer *practiceRendererFake) RenderPracticeList(format string, practices []core.PracticeView) error {
	renderer.practiceListFormat = format
	renderer.practices = practices
	return nil
}

func (renderer *practiceRendererFake) RenderPractice(format string, practice core.PracticeDetailView) error {
	renderer.practiceFormat = format
	renderer.practice = practice
	return nil
}

func (renderer *practiceRendererFake) RenderUpdatedPractice(id string) error {
	renderer.updatedPracticeID = id
	return nil
}

func (renderer *practiceRendererFake) RenderCreatedTestCase(id string) error {
	renderer.createdTestCaseID = id
	return nil
}

func (renderer *practiceRendererFake) RenderTestCaseList(format string, testCases []core.TestCaseView) error {
	renderer.testCaseListFormat = format
	renderer.testCases = testCases
	return nil
}

func (renderer *practiceRendererFake) RenderTestCase(format string, testCase core.TestCaseView) error {
	renderer.testCaseFormat = format
	renderer.testCase = testCase
	return nil
}

func (renderer *practiceRendererFake) RenderUpdatedTestCase(id string) error {
	renderer.updatedTestCaseID = id
	return nil
}

func (renderer *practiceRendererFake) RenderConfirmation(message string) error {
	renderer.confirmation = message
	return nil
}

func practiceViewFixture() core.PracticeView {
	now := time.Date(2026, 5, 27, 7, 0, 0, 0, time.UTC)
	return core.PracticeView{
		ID:            practiceIDValue,
		CourseID:      courseIDValue,
		Title:         "FizzBuzz",
		Language:      "golang",
		TestCaseCount: 1,
		HasSolution:   true,
		CreatedAt:     now,
		UpdatedAt:     now,
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
