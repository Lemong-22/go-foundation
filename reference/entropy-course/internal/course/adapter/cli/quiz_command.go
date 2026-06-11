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
	ErrMissingQuizService   = errors.New("quiz service is required")
	ErrInvalidQuestionOrder = errors.New("invalid question order")
)

type QuizRenderer interface {
	RenderCreatedQuiz(id string) error
	RenderQuizList(format string, quizzes []core.QuizView) error
	RenderQuiz(format string, quiz core.QuizDetailView) error
	RenderUpdatedQuiz(id string) error
	RenderCreatedQuestion(id string) error
	RenderQuestionList(format string, questions []core.QuestionView) error
	RenderQuestion(format string, question core.QuestionView) error
	RenderUpdatedQuestion(id string) error
	RenderConfirmation(message string) error
}

type QuizCommandOptions struct {
	Service  core.QuizService
	Renderer QuizRenderer
	Prompter CoursePrompter
}

type quizCommandContext struct {
	service  core.QuizService
	renderer QuizRenderer
	prompter CoursePrompter
}

func NewQuizCommand(options QuizCommandOptions) *cobra.Command {
	context := quizCommandContext{
		service:  options.Service,
		renderer: options.Renderer,
		prompter: defaultCoursePrompter(options.Prompter),
	}

	command := &cobra.Command{
		Use:   "quiz",
		Short: "Manage quizzes",
	}

	command.AddCommand(
		newQuizCreateCommand(context),
		newQuizListCommand(context),
		newQuizGetCommand(context),
		newQuizUpdateCommand(context),
		newQuizDeleteCommand(context),
		newQuizQuestionCommand(context),
	)

	return command
}

func newQuizCreateCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "create",
		Short: "Create a quiz",
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

			out, err := context.service.CreateQuiz(core.CreateQuizInput{
				CourseID:      courseID,
				Title:         title,
				PassThreshold: changedFloatFlag(command, "pass-threshold"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCreatedQuiz(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("course-id", "", "course id")
	flags.String("title", "", "quiz title")
	flags.Float64("pass-threshold", 0, "passing threshold")

	return command
}

func newQuizListCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List quizzes",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			courseID, err := requiredFlag(command, "course-id")
			if err != nil {
				return err
			}

			out, err := context.service.ListQuizzes(core.ListQuizzesInput{CourseID: courseID})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderQuizList(outputFormat(command), out.Quizzes)
		},
	}

	flags := command.Flags()
	flags.String("course-id", "", "course id")
	flags.StringP("output", "o", "table", "output format")

	return command
}

func newQuizGetCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <quiz-id>",
		Short: "Get a quiz",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.GetQuiz(core.GetQuizInput{ID: args[0]})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderQuiz(outputFormat(command), out.Quiz)
		},
	}

	command.Flags().StringP("output", "o", "table", "output format")

	return command
}

func newQuizUpdateCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <quiz-id>",
		Short: "Update a quiz",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.UpdateQuiz(core.UpdateQuizInput{
				ID:            args[0],
				Title:         changedStringFlag(command, "title"),
				PassThreshold: changedFloatFlag(command, "pass-threshold"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderUpdatedQuiz(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("title", "", "quiz title")
	flags.Float64("pass-threshold", 0, "passing threshold")

	return command
}

func newQuizDeleteCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete <quiz-id>",
		Short: "Delete a quiz",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("delete quiz %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			if err := context.service.DeleteQuiz(core.DeleteQuizInput{ID: args[0]}); err != nil {
				_ = renderQuizInUse(command.ErrOrStderr(), err)
				return err
			}

			return context.rendererFor(command).RenderConfirmation("quiz deleted")
		},
	}

	command.Flags().Bool("force", false, "skip confirmation")

	return command
}

func newQuizQuestionCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "question",
		Short: "Manage quiz questions",
	}

	command.AddCommand(
		newQuestionAddCommand(context),
		newQuestionListCommand(context),
		newQuestionGetCommand(context),
		newQuestionUpdateCommand(context),
		newQuestionRemoveCommand(context),
		newQuestionReorderCommand(context),
	)

	return command
}

func newQuestionAddCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "add",
		Short: "Add a quiz question",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			input, err := questionAddInput(command)
			if err != nil {
				return err
			}

			out, err := context.service.AddQuestion(input)
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderCreatedQuestion(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("quiz-id", "", "quiz id")
	flags.String("type", "", "question type")
	flags.String("prompt", "", "question prompt")
	flags.StringArray("option", nil, "answer option")
	flags.IntSlice("correct", nil, "correct option index")
	flags.String("explanation", "", "question explanation")
	flags.Int("position", 0, "question position")

	return command
}

func questionAddInput(command *cobra.Command) (core.AddQuestionInput, error) {
	quizID, err := requiredFlag(command, "quiz-id")
	if err != nil {
		return core.AddQuestionInput{}, err
	}
	questionType, err := requiredFlag(command, "type")
	if err != nil {
		return core.AddQuestionInput{}, err
	}
	prompt, err := requiredFlag(command, "prompt")
	if err != nil {
		return core.AddQuestionInput{}, err
	}
	options, err := requiredStringArrayFlag(command, "option")
	if err != nil {
		return core.AddQuestionInput{}, err
	}
	correct, err := requiredIntSliceFlag(command, "correct")
	if err != nil {
		return core.AddQuestionInput{}, err
	}

	return core.AddQuestionInput{
		QuizID:         quizID,
		Type:           questionType,
		Prompt:         prompt,
		Options:        options,
		CorrectIndices: correct,
		Explanation:    stringFlag(command, "explanation"),
		Position:       changedIntFlag(command, "position"),
	}, nil
}

func newQuestionListCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "list",
		Short: "List quiz questions",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			quizID, err := requiredFlag(command, "quiz-id")
			if err != nil {
				return err
			}

			out, err := context.service.ListQuestions(core.ListQuestionsInput{QuizID: quizID})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderQuestionList(outputFormat(command), out.Questions)
		},
	}

	flags := command.Flags()
	flags.String("quiz-id", "", "quiz id")
	flags.StringP("output", "o", "table", "output format")

	return command
}

func newQuestionGetCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "get <question-id>",
		Short: "Get a quiz question",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.GetQuestion(core.GetQuestionInput{ID: args[0]})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderQuestion(outputFormat(command), out.Question)
		},
	}

	command.Flags().StringP("output", "o", "table", "output format")

	return command
}

func newQuestionUpdateCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "update <question-id>",
		Short: "Update a quiz question",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			out, err := context.service.UpdateQuestion(core.UpdateQuestionInput{
				ID:             args[0],
				Prompt:         changedStringFlag(command, "prompt"),
				Options:        changedStringArrayFlag(command, "option"),
				CorrectIndices: changedIntSliceFlag(command, "correct"),
				Explanation:    changedStringFlag(command, "explanation"),
			})
			if err != nil {
				return err
			}

			return context.rendererFor(command).RenderUpdatedQuestion(out.ID)
		},
	}

	flags := command.Flags()
	flags.String("prompt", "", "question prompt")
	flags.StringArray("option", nil, "answer option")
	flags.IntSlice("correct", nil, "correct option index")
	flags.String("explanation", "", "question explanation")

	return command
}

func newQuestionRemoveCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "remove <question-id>",
		Short: "Remove a quiz question",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			if !boolFlag(command, "force") {
				confirmed, err := context.prompter.Confirm(fmt.Sprintf("remove quiz question %s", args[0]))
				if err != nil {
					return err
				}
				if !confirmed {
					return ErrConfirmationDeclined
				}
			}

			if err := context.service.RemoveQuestion(core.RemoveQuestionInput{ID: args[0]}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("quiz question removed")
		},
	}

	command.Flags().Bool("force", false, "skip confirmation")

	return command
}

func newQuestionReorderCommand(context quizCommandContext) *cobra.Command {
	command := &cobra.Command{
		Use:   "reorder",
		Short: "Reorder quiz questions",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := context.validate(); err != nil {
				return err
			}

			quizID, err := requiredFlag(command, "quiz-id")
			if err != nil {
				return err
			}
			order, err := requiredFlag(command, "order")
			if err != nil {
				return err
			}
			placements, err := parseQuestionPlacements(order)
			if err != nil {
				return err
			}

			if err := context.service.ReorderQuestions(core.ReorderQuestionsInput{
				QuizID: quizID,
				Order:  placements,
			}); err != nil {
				return err
			}

			return context.rendererFor(command).RenderConfirmation("quiz questions reordered")
		},
	}

	flags := command.Flags()
	flags.String("quiz-id", "", "quiz id")
	flags.String("order", "", "question-id:position pairs")

	return command
}

func (context quizCommandContext) validate() error {
	if context.service == nil {
		return ErrMissingQuizService
	}
	if context.prompter == nil {
		return ErrInvalidCommandOptions
	}

	return nil
}

func (context quizCommandContext) rendererFor(command *cobra.Command) QuizRenderer {
	if context.renderer != nil {
		return context.renderer
	}

	return newQuizOutputRenderer(command.OutOrStdout())
}

func changedFloatFlag(command *cobra.Command, name string) *float64 {
	flag := command.Flags().Lookup(name)
	if flag == nil || !flag.Changed {
		return nil
	}

	value, _ := command.Flags().GetFloat64(name)
	return &value
}

func requiredStringArrayFlag(command *cobra.Command, name string) ([]string, error) {
	values := stringArrayFlag(command, name)
	if len(values) == 0 {
		return nil, fmt.Errorf("%w: --%s", ErrRequiredFlagMissing, name)
	}

	return values, nil
}

func changedStringArrayFlag(command *cobra.Command, name string) *[]string {
	flag := command.Flags().Lookup(name)
	if flag == nil || !flag.Changed {
		return nil
	}

	values := stringArrayFlag(command, name)
	return &values
}

func stringArrayFlag(command *cobra.Command, name string) []string {
	values, _ := command.Flags().GetStringArray(name)
	return values
}

func requiredIntSliceFlag(command *cobra.Command, name string) ([]int, error) {
	values := intSliceFlag(command, name)
	if len(values) == 0 {
		return nil, fmt.Errorf("%w: --%s", ErrRequiredFlagMissing, name)
	}

	return values, nil
}

func changedIntSliceFlag(command *cobra.Command, name string) *[]int {
	flag := command.Flags().Lookup(name)
	if flag == nil || !flag.Changed {
		return nil
	}

	values := intSliceFlag(command, name)
	return &values
}

func intSliceFlag(command *cobra.Command, name string) []int {
	values, _ := command.Flags().GetIntSlice(name)
	return values
}

func parseQuestionPlacements(value string) ([]core.QuestionPlacementDTO, error) {
	parts := strings.Split(value, ",")
	placements := make([]core.QuestionPlacementDTO, 0, len(parts))

	for _, part := range parts {
		placement, err := parseQuestionPlacement(part)
		if err != nil {
			return nil, err
		}

		placements = append(placements, placement)
	}

	return placements, nil
}

func parseQuestionPlacement(value string) (core.QuestionPlacementDTO, error) {
	questionID, positionText, ok := strings.Cut(value, ":")
	if !ok {
		return core.QuestionPlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidQuestionOrder, value)
	}

	questionID = strings.TrimSpace(questionID)
	positionText = strings.TrimSpace(positionText)
	if questionID == "" || positionText == "" {
		return core.QuestionPlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidQuestionOrder, value)
	}

	position, err := strconv.Atoi(positionText)
	if err != nil {
		return core.QuestionPlacementDTO{}, fmt.Errorf("%w: %s", ErrInvalidQuestionOrder, value)
	}

	return core.QuestionPlacementDTO{QuestionID: questionID, Position: position}, nil
}

func renderQuizInUse(writer io.Writer, err error) error {
	var quizInUse domain.QuizInUseError
	if !errors.As(err, &quizInUse) {
		return nil
	}

	if _, writeErr := fmt.Fprintln(writer, "quiz is embedded in lessons:"); writeErr != nil {
		return writeErr
	}
	for _, lessonID := range quizInUse.LessonIDs {
		if _, writeErr := fmt.Fprintf(writer, "- %s\n", lessonID.String()); writeErr != nil {
			return writeErr
		}
	}

	return nil
}
