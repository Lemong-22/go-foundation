package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/spf13/cobra"
)

var (
	ErrMissingLessonService = errors.New("lesson service is required")
	ErrInvalidLessonOrder   = errors.New("invalid lesson order")
	ErrInvalidBlockOrder    = errors.New("invalid block order")
)

type LessonRenderer interface {
	RenderCreatedLesson(id string) error
	RenderLessonList(format string, lessons []core.LessonView) error
	RenderLesson(format string, lesson core.LessonView) error
	RenderUpdatedLesson(id string) error
	RenderCreatedBlock(id string) error
	RenderBlockList(format string, blocks []core.BlockView) error
	RenderBlock(format string, block core.BlockView) error
	RenderUpdatedBlock(id string) error
	RenderConfirmation(message string) error
}

type LessonCommandOptions struct {
	Service  core.LessonService
	Renderer LessonRenderer
	Prompter CoursePrompter
}

type lessonCommandContext struct {
	service  core.LessonService
	renderer LessonRenderer
	prompter CoursePrompter
}

func NewLessonCommand(options LessonCommandOptions) *cobra.Command {
	context := lessonCommandContext{
		service:  options.Service,
		renderer: options.Renderer,
		prompter: defaultCoursePrompter(options.Prompter),
	}

	command := &cobra.Command{
		Use:   "lesson",
		Short: "Manage lessons",
	}

	command.AddCommand(
		newLessonCreateCommand(context),
		newLessonListCommand(context),
		newLessonGetCommand(context),
		newLessonUpdateCommand(context),
		newLessonDeleteCommand(context),
		newLessonReorderCommand(context),
		newLessonBlockCommand(context),
	)

	return command
}

func newLessonCreateCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "Create a lesson",
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

			out, err := context.service.CreateLesson(core.CreateLessonInput{
				CourseID: courseID,
				Title:    title,
				Order:    changedIntFlag(command, "order"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCreatedLesson(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("course-id", "", "course id")
	flags.String("title", "", "lesson title")
	flags.Int("order", 0, "lesson order")

	return command
}

func newLessonListCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List lessons",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			courseID, err := requiredFlag(command, "course-id")
			if err != nil {
				return err
			}

			out, err := context.service.ListLessons(core.ListLessonsInput{CourseID: courseID})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderLessonList(outputFormat(command), out.Lessons)
		},
	}

	flags := command.Flags()
	flags.String("course-id", "", "course id")
	flags.StringP("output", "o", "table", "output format")

	return command
}

func newLessonGetCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <lesson-id>",
		Short: "Get a lesson",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.GetLesson(core.GetLessonInput{ID: args[0]})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderLesson(outputFormat(command), out.Lesson)
		},
	}

	command.Flags().StringP("output", "o", "table", "output format")

	return command
}

func newLessonUpdateCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <lesson-id>",
		Short: "Update a lesson",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.UpdateLesson(core.UpdateLessonInput{
				ID:    args[0],
				Title: changedStringFlag(command, "title"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderUpdatedLesson(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("title", "", "lesson title")

	return command
}

func newLessonDeleteCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <lesson-id>",
		Short: "Delete a lesson",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("delete lesson %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			if err := context.service.DeleteLesson(core.DeleteLessonInput{ID: args[0]}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("lesson deleted")
		},
	}

	command.Flags().Bool("force", false, "skip confirmation")

	return command
}

func newLessonReorderCommand(context lessonCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "reorder",
		Short: "Reorder lessons",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			courseID, err := requiredFlag(command, "course-id")
			if err != nil {
				return err
			}

			order, err := requiredFlag(command, "order")
			if err != nil {
				return err
			}
			positions, err := parseLessonPositions(order)
			if err != nil {
				return err
			}

			if err := context.service.ReorderLessons(core.ReorderLessonsInput{
				CourseID: courseID,
				Order:    positions,
			}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("lessons reordered")
		},
	}

	flags := command.Flags()
	flags.String("course-id", "", "course id")
	flags.String("order", "", "lesson-id:position pairs")

	return command
}

func parseLessonPositions(value string) ([]core.LessonPosition, error) {
	parts := strings.Split(value, ",")
	positions := make([]core.LessonPosition, 0, len(parts))

	for _, part := range parts {
		position, err := parseLessonPosition(part)
		if err != nil {
			return nil, err
		}

		positions = append(positions, position)
	}

	return positions, nil
}

func parseLessonPosition(value string) (core.LessonPosition, error) {
	lessonID, positionText, ok := strings.Cut(value, ":")
	if !ok {
		return core.LessonPosition{}, fmt.Errorf("%w: %s", ErrInvalidLessonOrder, value)
	}

	lessonID = strings.TrimSpace(lessonID)
	positionText = strings.TrimSpace(positionText)
	if lessonID == "" || positionText == "" {
		return core.LessonPosition{}, fmt.Errorf("%w: %s", ErrInvalidLessonOrder, value)
	}

	position, err := strconv.Atoi(positionText)
	if err != nil {
		return core.LessonPosition{}, fmt.Errorf("%w: %s", ErrInvalidLessonOrder, value)
	}

	return core.LessonPosition{LessonID: lessonID, Position: position}, nil
}

func (context lessonCommandContext) validate() error {
	if context.service == nil {
		return ErrMissingLessonService
	}
	if context.prompter == nil {
		return ErrInvalidCommandOptions
	}

	return nil
}

func (context lessonCommandContext) rendererFor(command *cobra.Command) LessonRenderer {
	if context.renderer != nil {
		return context.renderer
	}

	return newLessonOutputRenderer(command.OutOrStdout())
}

func changedIntFlag(command *cobra.Command, name string) *int {
	flag := command.Flags().Lookup(name)
	if flag == nil || !flag.Changed {
		return nil
	}

	value, _ := command.Flags().GetInt(name)
	return &value
}
