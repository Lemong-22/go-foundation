package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ErrMissingImportService = errors.New("import service is required")

type ImportRenderer interface {
	RenderImportPlan(format string, plan domain.ImportPlan) error
	RenderApplyResult(format string, result domain.ApplyResult) error
}

type ImportCommandOptions struct {
	Service  core.ImportService
	Renderer ImportRenderer
	Prompter CoursePrompter
	Config   *viper.Viper
}

type importCommandContext struct {
	service  core.ImportService
	renderer ImportRenderer
	prompter CoursePrompter
	config   *viper.Viper
}

func NewImportCommand(options ImportCommandOptions) *cobra.Command {
	context := importCommandContext{
		service:  options.Service,
		renderer: options.Renderer,
		prompter: defaultCoursePrompter(options.Prompter),
		config:   configureCourseViper(options.Config),
	}

	command := &cobra.Command{
		Use:   "import",
		Short: "Plan and apply course imports",
	}
	command.AddCommand(
		newImportPlanCommand(context),
		newImportApplyCommand(context),
	)

	return command
}

func newImportPlanCommand(context importCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "plan <zip-path>",
		Short: "Compute an import plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			instructorID := instructorID(command, context.config)
			if instructorID == "" {
				return ErrInstructorIDRequired
			}

			out, err := context.service.PlanImport(core.PlanImportInput{
				ZipPath:      args[0],
				InstructorID: instructorID,
			})
			if err != nil {
				return formatImportCommandError(err)
			}

			writer, closeWriter, err := importOutputWriter(command)
			if err != nil {
				return err
			}
			defer closeWriter()

			return context.rendererFor(writerOrCommandOutput(writer, command)).RenderImportPlan(importOutputFormat(command), out.Plan)
		},
	}

	flags := command.Flags()
	flags.StringP("format", "o", outputJSON, "output format")
	flags.String("output", "", "output file")
	flags.String("instructor-id", "", "instructor id")
	bindFlag(context.config, command, "instructor-id")

	return command
}

func newImportApplyCommand(context importCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "apply <zip-path>",
		Short: "Apply an import plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			instructorID := instructorID(command, context.config)
			if instructorID == "" {
				return ErrInstructorIDRequired
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("apply import %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			resolvedPlanJSON, err := resolvedPlanJSON(command)
			if err != nil {
				return err
			}

			out, err := context.service.ApplyPlan(core.ApplyPlanInput{
				ZipPath:          args[0],
				InstructorID:     instructorID,
				ResolvedPlanJSON: resolvedPlanJSON,
				ConflictStrategy: stringFlag(command, "conflict-strategy"),
			})
			if err != nil {
				return formatImportCommandError(err)
			}

			return context.rendererFor(command.OutOrStdout()).RenderApplyResult(importOutputFormat(command), out.Result)
		},
	}

	flags := command.Flags()
	flags.StringP("format", "o", outputTable, "output format")
	flags.String("resolved-plan", "", "resolved plan JSON file")
	flags.String("conflict-strategy", domain.FailConflicts().String(), "conflict strategy")
	flags.Bool("force", false, "skip confirmation")
	flags.String("instructor-id", "", "instructor id")
	bindFlag(context.config, command, "instructor-id")

	return command
}

func (context importCommandContext) validate() error {
	if context.service == nil {
		return ErrMissingImportService
	}
	if context.prompter == nil || context.config == nil {
		return ErrInvalidCommandOptions
	}

	return nil
}

func (context importCommandContext) rendererFor(writer io.Writer) ImportRenderer {
	if context.renderer != nil {
		return context.renderer
	}

	return newImportOutputRenderer(writer)
}

func importOutputWriter(command *cobra.Command) (*os.File, func(), error) {
	path := strings.TrimSpace(stringFlag(command, "output"))
	if path == "" {
		return nil, func() {}, nil
	}

	file, err := os.Create(path)
	if err != nil {
		return nil, func() {}, err
	}

	return file, func() { _ = file.Close() }, nil
}

func writerOrCommandOutput(file *os.File, command *cobra.Command) io.Writer {
	if file != nil {
		return file
	}

	return command.OutOrStdout()
}

func importOutputFormat(command *cobra.Command) string {
	return stringFlag(command, "format")
}

func resolvedPlanJSON(command *cobra.Command) ([]byte, error) {
	path := strings.TrimSpace(stringFlag(command, "resolved-plan"))
	if path == "" {
		return nil, nil
	}

	return os.ReadFile(path)
}

func formatImportCommandError(err error) error {
	var unresolved domain.UnresolvedImportConflictsError
	if errors.As(err, &unresolved) {
		refs := make([]string, 0, len(unresolved.Conflicts()))
		for _, conflict := range unresolved.Conflicts() {
			refs = append(refs, conflict.EntityRef())
		}
		if len(refs) > 0 {
			return fmt.Errorf("%w: %s", err, strings.Join(refs, ", "))
		}
	}

	return err
}
