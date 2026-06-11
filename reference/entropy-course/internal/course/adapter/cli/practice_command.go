package cli

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
	"github.com/spf13/cobra"
)

var (
	ErrMissingPracticeService = errors.New("practice service is required")
	ErrInvalidTestCaseOrder   = errors.New("invalid test case order")
)

type PracticeRenderer interface {
	RenderCreatedPractice(id string) error
	RenderPracticeList(format string, practices []core.PracticeView) error
	RenderPractice(format string, practice core.PracticeDetailView) error
	RenderUpdatedPractice(id string) error
	RenderCreatedTestCase(id string) error
	RenderTestCaseList(format string, testCases []core.TestCaseView) error
	RenderTestCase(format string, testCase core.TestCaseView) error
	RenderUpdatedTestCase(id string) error
	RenderConfirmation(message string) error
}

type PracticeCommandOptions struct {
	Service  core.PracticeService
	Renderer PracticeRenderer
	Prompter CoursePrompter
}

type practiceCommandContext struct {
	service  core.PracticeService
	renderer PracticeRenderer
	prompter CoursePrompter
}

func NewPracticeCommand(options PracticeCommandOptions) *cobra.Command {
	context := practiceCommandContext{
		service:  options.Service,
		renderer: options.Renderer,
		prompter: defaultCoursePrompter(options.Prompter),
	}

	command := &cobra.Command{
		Use:   "practice",
		Short: "Manage practices",
	}

	command.AddCommand(
		newPracticeCreateCommand(context),
		newPracticeListCommand(context),
		newPracticeGetCommand(context),
		newPracticeUpdateCommand(context),
		newPracticeDeleteCommand(context),
		newPracticeTestCaseCommand(context),
	)

	return command
}

func newPracticeCreateCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "Create a practice",
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
			language, err := requiredPracticeLanguage(command)
			if err != nil {
				return err
			}
			prompt, err := requiredFlag(command, "prompt")
			if err != nil {
				return err
			}

			out, err := context.service.CreatePractice(core.CreatePracticeInput{
				CourseID:    courseID,
				Title:       title,
				Language:    language,
				Prompt:      prompt,
				StarterCode: stringFlag(command, "starter-code"),
				Solution:    stringFlag(command, "solution"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCreatedPractice(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("course-id", "", "course id")
	flags.String("title", "", "practice title")
	flags.String("language", "", "practice language")
	flags.String("prompt", "", "practice prompt")
	flags.String("starter-code", "", "starter code")
	flags.String("solution", "", "reference solution")

	return command
}

func newPracticeListCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List practices",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			courseID, err := requiredFlag(command, "course-id")
			if err != nil {
				return err
			}

			out, err := context.service.ListPractices(core.ListPracticesInput{CourseID: courseID})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderPracticeList(outputFormat(command), out.Practices)
		},
	}

	flags := command.Flags()
	flags.String("course-id", "", "course id")
	flags.StringP("output", "o", "table", "output format")

	return command
}

func newPracticeGetCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <practice-id>",
		Short: "Get a practice",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.GetPractice(core.GetPracticeInput{ID: args[0]})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderPractice(outputFormat(command), out.Practice)
		},
	}

	command.Flags().StringP("output", "o", "table", "output format")

	return command
}

func newPracticeUpdateCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <practice-id>",
		Short: "Update a practice",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.UpdatePractice(core.UpdatePracticeInput{
				ID:          args[0],
				Title:       changedStringFlag(command, "title"),
				Prompt:      changedStringFlag(command, "prompt"),
				StarterCode: changedStringFlag(command, "starter-code"),
				Solution:    changedStringFlag(command, "solution"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderUpdatedPractice(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("title", "", "practice title")
	flags.String("prompt", "", "practice prompt")
	flags.String("starter-code", "", "starter code")
	flags.String("solution", "", "reference solution")

	return command
}

func newPracticeDeleteCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <practice-id>",
		Short: "Delete a practice",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("delete practice %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			if err := context.service.DeletePractice(core.DeletePracticeInput{ID: args[0]}); err != nil {
				_ = renderPracticeInUse(command.ErrOrStderr(), err)
				return err
			}

			return context.rendererFor(command).RenderConfirmation("practice deleted")
		},
	}

	command.Flags().Bool("force", false, "skip confirmation")

	return command
}

func newPracticeTestCaseCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "testcase",
		Short: "Manage practice test cases",
	}

	command.AddCommand(
		newTestCaseAddCommand(context),
		newTestCaseListCommand(context),
		newTestCaseGetCommand(context),
		newTestCaseUpdateCommand(context),
		newTestCaseRemoveCommand(context),
		newTestCaseReorderCommand(context),
	)

	return command
}

func newTestCaseAddCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "add",
		Short: "Add a practice test case",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			practiceID, err := requiredFlag(command, "practice-id")
			if err != nil {
				return err
			}

			out, err := context.service.AddTestCase(core.AddTestCaseInput{
				PracticeID:     practiceID,
				Stdin:          stringFlag(command, "stdin"),
				ExpectedStdout: stringFlag(command, "expected-stdout"),
				Name:           stringFlag(command, "name"),
				Position:       changedIntFlag(command, "position"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCreatedTestCase(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("practice-id", "", "practice id")
	flags.String("stdin", "", "stdin")
	flags.String("expected-stdout", "", "expected stdout")
	flags.String("name", "", "test case name")
	flags.Int("position", 0, "test case position")

	return command
}

func newTestCaseListCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List practice test cases",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			practiceID, err := requiredFlag(command, "practice-id")
			if err != nil {
				return err
			}

			out, err := context.service.ListTestCases(core.ListTestCasesInput{PracticeID: practiceID})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderTestCaseList(outputFormat(command), out.TestCases)
		},
	}

	flags := command.Flags()
	flags.String("practice-id", "", "practice id")
	flags.StringP("output", "o", "table", "output format")

	return command
}

func newTestCaseGetCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <testcase-id>",
		Short: "Get a practice test case",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.GetTestCase(core.GetTestCaseInput{ID: args[0]})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderTestCase(outputFormat(command), out.TestCase)
		},
	}

	command.Flags().StringP("output", "o", "table", "output format")

	return command
}

func newTestCaseUpdateCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <testcase-id>",
		Short: "Update a practice test case",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.UpdateTestCase(core.UpdateTestCaseInput{
				ID:             args[0],
				Stdin:          changedStringFlag(command, "stdin"),
				ExpectedStdout: changedStringFlag(command, "expected-stdout"),
				Name:           changedStringFlag(command, "name"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderUpdatedTestCase(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("stdin", "", "stdin")
	flags.String("expected-stdout", "", "expected stdout")
	flags.String("name", "", "test case name")

	return command
}

func newTestCaseRemoveCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "remove <testcase-id>",
		Short: "Remove a practice test case",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("remove practice test case %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			if err := context.service.RemoveTestCase(core.RemoveTestCaseInput{ID: args[0]}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("practice test case removed")
		},
	}

	command.Flags().Bool("force", false, "skip confirmation")

	return command
}

func newTestCaseReorderCommand(context practiceCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "reorder",
		Short: "Reorder practice test cases",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			practiceID, err := requiredFlag(command, "practice-id")
			if err != nil {
				return err
			}
			order, err := requiredFlag(command, "order")
			if err != nil {
				return err
			}
			placements, err := parseTestCasePlacements(order)
			if err != nil {
				return err
			}

			if err := context.service.ReorderTestCases(core.ReorderTestCasesInput{
				PracticeID: practiceID,
				Order:      placements,
			}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("practice test cases reordered")
		},
	}

	flags := command.Flags()
	flags.String("practice-id", "", "practice id")
	flags.String("order", "", "testcase-id:position pairs")

	return command
}

func (context practiceCommandContext) validate() error {
	if context.service == nil {
		return ErrMissingPracticeService
	}
	if context.prompter == nil {
		return ErrInvalidCommandOptions
	}

	return nil
}

func (context practiceCommandContext) rendererFor(command *cobra.Command) PracticeRenderer {
	if context.renderer != nil {
		return context.renderer
	}

	return newPracticeOutputRenderer(command.OutOrStdout())
}

func requiredPracticeLanguage(command *cobra.Command) (string, error) {
	value, err := requiredFlag(command, "language")
	if err != nil {
		return "", err
	}
	if _, err := domain.NewLanguage(value); err != nil {
		return "", err
	}

	return value, nil
}

func parseTestCasePlacements(value string) ([]core.TestCasePlacementDTO, error) {
	parts := strings.Split(value, ",")
	placements := make([]core.TestCasePlacementDTO, 0, len(parts))

	for _, part := range parts {
		placement, err := parseTestCasePlacement(part)
		if err != nil {
			return nil, err
		}

		placements = append(placements, placement)
	}

	return placements, nil
}

func parseTestCasePlacement(value string) (core.TestCasePlacementDTO, error) {
	testCaseID, positionText, ok := strings.Cut(value, ":")
	if !ok {
		return core.TestCasePlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidTestCaseOrder, value)
	}

	testCaseID = strings.TrimSpace(testCaseID)
	positionText = strings.TrimSpace(positionText)
	if testCaseID == "" || positionText == "" {
		return core.TestCasePlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidTestCaseOrder, value)
	}

	position, err := strconv.Atoi(positionText)
	if err != nil {
		return core.TestCasePlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidTestCaseOrder, value)
	}

	return core.TestCasePlacementDTO{TestCaseID: testCaseID, Position: position}, nil
}

func renderPracticeInUse(writer io.Writer, err error) error {
	var practiceInUse domain.PracticeInUseError
	if !errors.As(err, &practiceInUse) {
		return nil
	}

	if _, writeErr := fmt.Fprintln(writer, "practice is embedded in lessons:"); writeErr != nil {
		return writeErr
	}
	for _, lessonID := range practiceInUse.LessonIDs {
		if _, writeErr := fmt.Fprintf(writer, "- %s\n", lessonID.String()); writeErr != nil {
			return writeErr
		}
	}

	return nil
}
