package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
	"github.com/spf13/cobra"
)

var (
	ErrMissingTestService    = errors.New("test service is required")
	ErrInvalidTestItemOrder  = errors.New("invalid test item order")
	ErrInvalidCodingTestCase = errors.New("invalid coding test case")
)

type TestRenderer interface {
	RenderCreatedTest(id string) error
	RenderTestList(format string, tests []core.TestView) error
	RenderTest(format string, test core.TestDetailView) error
	RenderUpdatedTest(id string) error
	RenderCreatedTestItem(id string) error
	RenderTestItemList(format string, items []core.TestItemView) error
	RenderTestItem(format string, item core.TestItemView) error
	RenderUpdatedTestItem(id string) error
	RenderConfirmation(message string) error
}

type TestCommandOptions struct {
	Service  core.TestService
	Renderer TestRenderer
	Prompter CoursePrompter
}

type testCommandContext struct {
	service  core.TestService
	renderer TestRenderer
	prompter CoursePrompter
}

func NewTestCommand(options TestCommandOptions) *cobra.Command {
	context := testCommandContext{
		service:  options.Service,
		renderer: options.Renderer,
		prompter: defaultCoursePrompter(options.Prompter),
	}

	command := &cobra.Command{
		Use:   "test",
		Short: "Manage tests",
	}

	command.AddCommand(
		newTestCreateCommand(context),
		newTestListCommand(context),
		newTestGetCommand(context),
		newTestUpdateCommand(context),
		newTestDeleteCommand(context),
		newTestItemCommand(context),
	)

	return command
}

func newTestCreateCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "Create a test",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			courseID, err := requiredFlag(command, "course-id")
			if err != nil {
				return err
			}
			title, err := requiredFlag(command, "title")
			if err != nil {
				return err
			}

			out, err := context.service.CreateTest(core.CreateTestInput{
				CourseID:         courseID,
				Title:            title,
				TimeLimitMinutes: changedIntFlag(command, "time-limit-minutes"),
				PassThreshold:    changedFloatFlag(command, "pass-threshold"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCreatedTest(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("course-id", "", "course id")
	flags.String("title", "", "test title")
	flags.Int("time-limit-minutes", 0, "time limit in minutes")
	flags.Float64("pass-threshold", 0, "passing threshold")

	return command
}

func newTestListCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List tests",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			courseID, err := requiredFlag(command, "course-id")
			if err != nil {
				return err
			}

			out, err := context.service.ListTests(core.ListTestsInput{CourseID: courseID})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderTestList(outputFormat(command), out.Tests)
		},
	}

	flags := command.Flags()
	flags.String("course-id", "", "course id")
	flags.StringP("output", "o", "table", "output format")

	return command
}

func newTestGetCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <test-id>",
		Short: "Get a test",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.GetTest(core.GetTestInput{ID: args[0]})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderTest(outputFormat(command), out.Test)
		},
	}

	command.Flags().StringP("output", "o", "table", "output format")

	return command
}

func newTestUpdateCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <test-id>",
		Short: "Update a test",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			input, err := testUpdateInput(command, args[0])
			if err != nil {
				return err
			}

			out, err := context.service.UpdateTest(input)
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderUpdatedTest(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("title", "", "test title")
	flags.Int("time-limit-minutes", 0, "time limit in minutes; zero clears")
	flags.Float64("pass-threshold", 0, "passing threshold")
	flags.String("solution-zip-provider", "", "solution zip provider")
	flags.String("solution-zip-locator", "", "solution zip locator")
	flags.String("solution-video-provider", "", "solution video provider")
	flags.String("solution-video-locator", "", "solution video locator")
	flags.String("solution-video-caption", "", "solution video caption")

	return command
}

func testUpdateInput(command *cobra.Command, id string) (core.UpdateTestInput, error) {
	input := core.UpdateTestInput{
		ID:                    id,
		Title:                 changedStringFlag(command, "title"),
		TimeLimitMinutes:      changedIntFlag(command, "time-limit-minutes"),
		PassThreshold:         changedFloatFlag(command, "pass-threshold"),
		SolutionZipProvider:   changedStringFlag(command, "solution-zip-provider"),
		SolutionZipLocator:    changedStringFlag(command, "solution-zip-locator"),
		SolutionVideoProvider: changedStringFlag(command, "solution-video-provider"),
		SolutionVideoLocator:  changedStringFlag(command, "solution-video-locator"),
		SolutionVideoCaption:  changedStringFlag(command, "solution-video-caption"),
	}
	if hasAnyChangedSolutionFlag(input) && !hasRequiredChangedSolutionFlags(input) {
		return core.UpdateTestInput{}, domain.NewValidationError("solution", "must include zip provider, zip locator, video provider, and video locator together")
	}

	return input, nil
}

func hasAnyChangedSolutionFlag(in core.UpdateTestInput) bool {
	return in.SolutionZipProvider != nil ||
		in.SolutionZipLocator != nil ||
		in.SolutionVideoProvider != nil ||
		in.SolutionVideoLocator != nil ||
		in.SolutionVideoCaption != nil
}

func hasRequiredChangedSolutionFlags(in core.UpdateTestInput) bool {
	return in.SolutionZipProvider != nil &&
		in.SolutionZipLocator != nil &&
		in.SolutionVideoProvider != nil &&
		in.SolutionVideoLocator != nil
}

func newTestDeleteCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <test-id>",
		Short: "Delete a test",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("delete test %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			if err := context.service.DeleteTest(core.DeleteTestInput{ID: args[0]}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("test deleted")
		},
	}

	command.Flags().Bool("force", false, "skip confirmation")

	return command
}

func newTestItemCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "item",
		Short: "Manage test items",
	}

	command.AddCommand(
		newTestItemAddCommand(context),
		newTestItemListCommand(context),
		newTestItemGetCommand(context),
		newTestItemUpdateCommand(context),
		newTestItemRemoveCommand(context),
		newTestItemReorderCommand(context),
	)

	return command
}

func newTestItemAddCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "add",
		Short: "Add a test item",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			input, err := testItemAddInput(command)
			if err != nil {
				return err
			}

			out, err := context.service.AddTestItem(input)
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCreatedTestItem(out.ID)
		},
	}

	addTestItemFlags(command)
	return command
}

func addTestItemFlags(command *cobra.Command) {
	flags := command.Flags()
	flags.String("test-id", "", "test id")
	flags.String("kind", "", "item kind")
	flags.Int("position", 0, "item position")
	flags.String("prompt", "", "item prompt")
	flags.String("type", "", "choice item type")
	flags.StringArray("option", nil, "choice option")
	flags.IntSlice("correct", nil, "correct choice index")
	flags.String("explanation", "", "choice explanation")
	flags.String("language", "", "coding language")
	flags.String("starter-code", "", "starter code")
	flags.String("solution", "", "reference solution")
	flags.StringArray("testcase", nil, "coding test case stdin::expected[::name]")
}

func testItemAddInput(command *cobra.Command) (core.AddTestItemInput, error) {
	testID, err := requiredFlag(command, "test-id")
	if err != nil {
		return core.AddTestItemInput{}, err
	}
	kind, err := requiredTestItemKind(command)
	if err != nil {
		return core.AddTestItemInput{}, err
	}
	prompt, err := requiredFlag(command, "prompt")
	if err != nil {
		return core.AddTestItemInput{}, err
	}

	input := core.AddTestItemInput{
		TestID:   testID,
		Kind:     kind.String(),
		Position: changedIntFlag(command, "position"),
	}
	if kind.IsChoice() {
		choiceType, err := requiredFlag(command, "type")
		if err != nil {
			return core.AddTestItemInput{}, err
		}
		options, err := requiredStringArrayFlag(command, "option")
		if err != nil {
			return core.AddTestItemInput{}, err
		}
		correct, err := requiredIntSliceFlag(command, "correct")
		if err != nil {
			return core.AddTestItemInput{}, err
		}

		input.Prompt = prompt
		input.ChoiceType = choiceType
		input.Options = options
		input.CorrectIndices = correct
		input.Explanation = stringFlag(command, "explanation")
		return input, nil
	}

	language, err := requiredTestItemLanguage(command)
	if err != nil {
		return core.AddTestItemInput{}, err
	}
	testCases, err := requiredCodingTestCases(command, "testcase")
	if err != nil {
		return core.AddTestItemInput{}, err
	}

	input.CodingPrompt = prompt
	input.Language = language
	input.StarterCode = stringFlag(command, "starter-code")
	input.Solution = stringFlag(command, "solution")
	input.TestCases = testCases
	return input, nil
}

func newTestItemListCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List test items",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			testID, err := requiredFlag(command, "test-id")
			if err != nil {
				return err
			}

			out, err := context.service.ListTestItems(core.ListTestItemsInput{TestID: testID})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderTestItemList(outputFormat(command), out.Items)
		},
	}

	flags := command.Flags()
	flags.String("test-id", "", "test id")
	flags.StringP("output", "o", "table", "output format")

	return command
}

func newTestItemGetCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <item-id>",
		Short: "Get a test item",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.GetTestItem(core.GetTestItemInput{ID: args[0]})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderTestItem(outputFormat(command), out.Item)
		},
	}

	command.Flags().StringP("output", "o", "table", "output format")

	return command
}

func newTestItemUpdateCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <item-id>",
		Short: "Update a test item",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			testCases, err := changedCodingTestCases(command, "testcase")
			if err != nil {
				return err
			}

			out, err := context.service.UpdateTestItem(core.UpdateTestItemInput{
				ID:             args[0],
				Prompt:         changedStringFlag(command, "prompt"),
				ChoiceType:     changedStringFlag(command, "type"),
				Options:        changedStringArrayFlag(command, "option"),
				CorrectIndices: changedIntSliceFlag(command, "correct"),
				Explanation:    changedStringFlag(command, "explanation"),
				Language:       changedStringFlag(command, "language"),
				StarterCode:    changedStringFlag(command, "starter-code"),
				Solution:       changedStringFlag(command, "solution"),
				TestCases:      testCases,
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderUpdatedTestItem(out.ID)
		},
	}

	updateTestItemFlags(command)
	return command
}

func updateTestItemFlags(command *cobra.Command) {
	flags := command.Flags()
	flags.String("prompt", "", "item prompt")
	flags.String("type", "", "choice item type")
	flags.StringArray("option", nil, "choice option")
	flags.IntSlice("correct", nil, "correct choice index")
	flags.String("explanation", "", "choice explanation")
	flags.String("language", "", "coding language")
	flags.String("starter-code", "", "starter code")
	flags.String("solution", "", "reference solution")
	flags.StringArray("testcase", nil, "coding test case stdin::expected[::name]")
}

func newTestItemRemoveCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "remove <item-id>",
		Short: "Remove a test item",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("remove test item %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			if err := context.service.RemoveTestItem(core.RemoveTestItemInput{ID: args[0]}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("test item removed")
		},
	}

	command.Flags().Bool("force", false, "skip confirmation")

	return command
}

func newTestItemReorderCommand(context testCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "reorder",
		Short: "Reorder test items",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			testID, err := requiredFlag(command, "test-id")
			if err != nil {
				return err
			}
			order, err := requiredFlag(command, "order")
			if err != nil {
				return err
			}
			placements, err := parseTestItemPlacements(order)
			if err != nil {
				return err
			}

			if err := context.service.ReorderTestItems(core.ReorderTestItemsInput{
				TestID: testID,
				Order:  placements,
			}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("test items reordered")
		},
	}

	flags := command.Flags()
	flags.String("test-id", "", "test id")
	flags.String("order", "", "item-id:position pairs")

	return command
}

func (context testCommandContext) validate() error {
	if context.service == nil {
		return ErrMissingTestService
	}
	if context.prompter == nil {
		return ErrInvalidCommandOptions
	}

	return nil
}

func (context testCommandContext) rendererFor(command *cobra.Command) TestRenderer {
	if context.renderer != nil {
		return context.renderer
	}

	return newTestOutputRenderer(command.OutOrStdout())
}

func requiredTestItemKind(command *cobra.Command) (domain.TestItemKind, error) {
	value, err := requiredFlag(command, "kind")
	if err != nil {
		return domain.TestItemKind{}, err
	}

	return domain.NewTestItemKind(value)
}

func requiredTestItemLanguage(command *cobra.Command) (string, error) {
	value, err := requiredFlag(command, "language")
	if err != nil {
		return "", err
	}
	if _, err := domain.NewLanguage(value); err != nil {
		return "", err
	}

	return value, nil
}

func requiredCodingTestCases(command *cobra.Command, name string) ([]core.CodingTestCaseDTO, error) {
	values, err := requiredStringArrayFlag(command, name)
	if err != nil {
		return nil, err
	}

	return parseCodingTestCases(values)
}

func changedCodingTestCases(command *cobra.Command, name string) (*[]core.CodingTestCaseDTO, error) {
	flag := command.Flags().Lookup(name)
	if flag == nil || !flag.Changed {
		return nil, nil
	}

	testCases, err := parseCodingTestCases(stringArrayFlag(command, name))
	if err != nil {
		return nil, err
	}

	return &testCases, nil
}

func parseCodingTestCases(values []string) ([]core.CodingTestCaseDTO, error) {
	testCases := make([]core.CodingTestCaseDTO, 0, len(values))
	for _, value := range values {
		testCase, err := parseCodingTestCase(value)
		if err != nil {
			return nil, err
		}
		testCases = append(testCases, testCase)
	}

	return testCases, nil
}

func parseCodingTestCase(value string) (core.CodingTestCaseDTO, error) {
	parts := strings.Split(value, "::")
	if len(parts) != 2 && len(parts) != 3 {
		return core.CodingTestCaseDTO{}, fmt.Errorf("%w: %s", ErrInvalidCodingTestCase, value)
	}

	testCase := core.CodingTestCaseDTO{
		Stdin:          parts[0],
		ExpectedStdout: parts[1],
	}
	if len(parts) == 3 {
		testCase.Name = parts[2]
	}

	return testCase, nil
}

func parseTestItemPlacements(value string) ([]core.TestItemPlacementDTO, error) {
	parts := strings.Split(value, ",")
	placements := make([]core.TestItemPlacementDTO, 0, len(parts))

	for _, part := range parts {
		placement, err := parseTestItemPlacement(part)
		if err != nil {
			return nil, err
		}

		placements = append(placements, placement)
	}

	return placements, nil
}

func parseTestItemPlacement(value string) (core.TestItemPlacementDTO, error) {
	itemID, positionText, ok := strings.Cut(value, ":")
	if !ok {
		return core.TestItemPlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidTestItemOrder, value)
	}

	itemID = strings.TrimSpace(itemID)
	positionText = strings.TrimSpace(positionText)
	if itemID == "" || positionText == "" {
		return core.TestItemPlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidTestItemOrder, value)
	}

	position, err := strconv.Atoi(positionText)
	if err != nil {
		return core.TestItemPlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidTestItemOrder, value)
	}

	return core.TestItemPlacementDTO{TestItemID: itemID, Position: position}, nil
}
