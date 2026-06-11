package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	ErrMissingCourseService  = errors.New("course service is required")
	ErrConfirmationDeclined  = errors.New("confirmation declined")
	ErrConfirmationRequired  = errors.New("confirmation required")
	ErrInstructorIDRequired  = errors.New("instructor id is required")
	ErrRequiredFlagMissing   = errors.New("required flag missing")
	ErrInvalidCommandOptions = errors.New("invalid command options")
)

type CourseRenderer interface {
	RenderCreatedCourse(id string) error
	RenderCourseList(format string, courses []core.CourseView) error
	RenderCourse(format string, course core.CourseView) error
	RenderUpdatedCourse(id string) error
	RenderConfirmation(message string) error
}

type CoursePrompter interface {
	Confirm(message string) (bool, error)
}

type CourseCommandOptions struct {
	Service  core.CourseService
	Renderer CourseRenderer
	Prompter CoursePrompter
	Config   *viper.Viper
}

type courseCommandContext struct {
	service  core.CourseService
	renderer CourseRenderer
	prompter CoursePrompter
	config   *viper.Viper
}

func NewCourseCommand(options CourseCommandOptions) *cobra.Command {
	context := courseCommandContext{
		service:  options.Service,
		renderer: options.Renderer,
		prompter: defaultCoursePrompter(options.Prompter),
		config:   configureCourseViper(options.Config),
	}

	command := &cobra.Command{
		Use:   "course",
		Short: "Manage courses",
	}

	command.AddCommand(
		newCourseCreateCommand(context),
		newCourseListCommand(context),
		newCourseGetCommand(context),
		newCourseUpdateCommand(context),
		newCourseDeleteCommand(context),
		newCoursePublishCommand(context),
		newCourseUnpublishCommand(context),
	)

	return command
}

func newCourseCreateCommand(context courseCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "Create a course",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			title, err := requiredFlag(command, "title")
			if err != nil {
				return err
			}
			slug, err := requiredFlag(command, "slug")
			if err != nil {
				return err
			}

			instructorID := instructorID(command, context.config)
			if instructorID == "" {
				return ErrInstructorIDRequired
			}

			out, err := context.service.CreateCourse(core.CreateCourseInput{
				Title:        title,
				Slug:         slug,
				Description:  stringFlag(command, "description"),
				InstructorID: instructorID,
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCreatedCourse(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("title", "", "course title")
	flags.String("slug", "", "course slug")
	flags.String("description", "", "course description")
	flags.String("instructor-id", "", "instructor id")
	bindFlag(context.config, command, "instructor-id")

	return command
}

func newCourseListCommand(context courseCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List courses",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.ListCourses(core.ListCoursesInput{Status: stringFlag(command, "status")})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCourseList(outputFormat(command), out.Courses)
		},
	}

	flags := command.Flags()
	flags.String("status", "", "course status filter")
	flags.StringP("output", "o", "table", "output format")

	return command
}

func newCourseGetCommand(context courseCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <course-id>",
		Short: "Get a course",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.GetCourse(core.GetCourseInput{ID: args[0]})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCourse(outputFormat(command), out.Course)
		},
	}

	command.Flags().StringP("output", "o", "table", "output format")

	return command
}

func newCourseUpdateCommand(context courseCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <course-id>",
		Short: "Update a course",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.UpdateCourse(core.UpdateCourseInput{
				ID:          args[0],
				Title:       changedStringFlag(command, "title"),
				Description: changedStringFlag(command, "description"),
				Slug:        changedStringFlag(command, "slug"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderUpdatedCourse(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("title", "", "course title")
	flags.String("description", "", "course description")
	flags.String("slug", "", "course slug")

	return command
}

func newCourseDeleteCommand(context courseCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <course-id>",
		Short: "Delete a course",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("delete course %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			if err := context.service.DeleteCourse(core.DeleteCourseInput{ID: args[0]}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("course deleted")
		},
	}

	command.Flags().Bool("force", false, "skip confirmation")

	return command
}

func newCoursePublishCommand(context courseCommandContext) *cobra.Command {
	return courseStateCommand(
		context,
		"publish <course-id>",
		"Publish a course",
		"course published",
		func(id string) error {
			return context.service.PublishCourse(core.PublishCourseInput{ID: id})
		},
	)
}

func newCourseUnpublishCommand(context courseCommandContext) *cobra.Command {
	return courseStateCommand(
		context,
		"unpublish <course-id>",
		"Unpublish a course",
		"course unpublished",
		func(id string) error {
			return context.service.UnpublishCourse(core.UnpublishCourseInput{ID: id})
		},
	)
}

func courseStateCommand(
	context courseCommandContext,
	use string,
	short string,
	confirmation string,
	run func(id string) error,
) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if err := run(args[0]); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation(confirmation)
		},
	}
}

func (context courseCommandContext) validate() error {
	if context.service == nil {
		return ErrMissingCourseService
	}
	if context.prompter == nil || context.config == nil {
		return ErrInvalidCommandOptions
	}

	return nil
}

func (context courseCommandContext) rendererFor(command *cobra.Command) CourseRenderer {
	if context.renderer != nil {
		return context.renderer
	}

	return newCourseOutputRenderer(command.OutOrStdout())
}

func configureCourseViper(config *viper.Viper) *viper.Viper {
	if config == nil {
		config = viper.New()
	}

	config.SetEnvPrefix("COURSE_CLI")
	config.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	config.AutomaticEnv()

	return config
}

func defaultCoursePrompter(prompter CoursePrompter) CoursePrompter {
	if prompter != nil {
		return prompter
	}

	return requiredCoursePrompter{}
}

func bindFlag(config *viper.Viper, command *cobra.Command, name string) {
	_ = config.BindPFlag(name, command.Flags().Lookup(name))
}

func requiredFlag(command *cobra.Command, name string) (string, error) {
	value := stringFlag(command, name)
	if value == "" {
		return "", fmt.Errorf("%w: --%s", ErrRequiredFlagMissing, name)
	}

	return value, nil
}

func stringFlag(command *cobra.Command, name string) string {
	value, _ := command.Flags().GetString(name)
	return value
}

func changedStringFlag(command *cobra.Command, name string) *string {
	flag := command.Flags().Lookup(name)
	if flag == nil || !flag.Changed {
		return nil
	}

	value := flag.Value.String()
	return &value
}

func requiredChangedStringFlag(command *cobra.Command, name string) (string, error) {
	value := changedStringFlag(command, name)
	if value == nil {
		return "", fmt.Errorf("%w: --%s", ErrRequiredFlagMissing, name)
	}

	return *value, nil
}

func boolFlag(command *cobra.Command, name string) bool {
	value, _ := command.Flags().GetBool(name)
	return value
}

func outputFormat(command *cobra.Command) string {
	return stringFlag(command, "output")
}

func instructorID(command *cobra.Command, config *viper.Viper) string {
	flag := command.Flags().Lookup("instructor-id")
	if flag != nil && flag.Changed {
		return flag.Value.String()
	}

	if value := config.GetString("instructor-id"); value != "" {
		return value
	}

	return config.GetString("instructor_id")
}

type requiredCoursePrompter struct{}

func (requiredCoursePrompter) Confirm(string) (bool, error) {
	return false, ErrConfirmationRequired
}
