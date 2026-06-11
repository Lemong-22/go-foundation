package playground

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	coursecli "github.com/luxeave/entropy-course/internal/course/adapter/cli"
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Runner struct {
	courseService core.CourseService
	lessonService core.LessonService
	quizService   core.QuizService
	practice      core.PracticeService
	test          core.TestService
	imports       core.ImportService
	config        *viper.Viper
	commands      map[string]Command
}

type Services struct {
	Course   core.CourseService
	Lesson   core.LessonService
	Quiz     core.QuizService
	Practice core.PracticeService
	Test     core.TestService
	Import   core.ImportService
	Config   *viper.Viper
}

type RunResult struct {
	OK        bool   `json:"ok"`
	Action    string `json:"action"`
	Command   string `json:"command"`
	Output    string `json:"output"`
	ExitCode  int    `json:"exitCode"`
	ElapsedMS int64  `json:"elapsedMs"`
}

func NewRunner(services Services) Runner {
	commands := make(map[string]Command)
	for _, command := range Catalog() {
		commands[command.ID] = command
	}

	return Runner{
		courseService: services.Course,
		lessonService: services.Lesson,
		quizService:   services.Quiz,
		practice:      services.Practice,
		test:          services.Test,
		imports:       services.Import,
		config:        services.Config,
		commands:      commands,
	}
}

func (runner Runner) Run(ctx context.Context, action string, values map[string]string) RunResult {
	startedAt := time.Now()
	command, ok := runner.commands[action]
	if !ok {
		return runner.failure(action, "", fmt.Errorf("unknown action: %s", action), startedAt)
	}

	args := commandArgs(command, values)
	output, err := runner.execute(ctx, args)
	if err != nil {
		if output != "" {
			output += "\n"
		}
		output += "error: " + err.Error()

		return RunResult{
			OK:        false,
			Action:    action,
			Command:   preview(command, values),
			Output:    output,
			ExitCode:  coursecli.ExitCode(err),
			ElapsedMS: elapsedMilliseconds(startedAt),
		}
	}

	return RunResult{
		OK:        true,
		Action:    action,
		Command:   preview(command, values),
		Output:    strings.TrimRight(output, "\n"),
		ExitCode:  coursecli.ExitOK,
		ElapsedMS: elapsedMilliseconds(startedAt),
	}
}

func (runner Runner) execute(ctx context.Context, args []string) (string, error) {
	var output bytes.Buffer
	command := runner.rootCommand()
	command.SetContext(ctx)
	command.SetArgs(args)
	command.SetOut(&output)
	command.SetErr(&output)
	command.SilenceUsage = true
	command.SilenceErrors = true

	err := command.Execute()
	return output.String(), err
}

func (runner Runner) rootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "course-cli",
		Short: "Manage courses, lessons, quizzes, practices, and tests",
	}
	root.AddCommand(
		coursecli.NewCourseCommand(coursecli.CourseCommandOptions{
			Service: runner.courseService,
			Config:  runner.config,
		}),
		coursecli.NewLessonCommand(coursecli.LessonCommandOptions{
			Service: runner.lessonService,
		}),
		coursecli.NewQuizCommand(coursecli.QuizCommandOptions{
			Service: runner.quizService,
		}),
		coursecli.NewPracticeCommand(coursecli.PracticeCommandOptions{
			Service: runner.practice,
		}),
		coursecli.NewTestCommand(coursecli.TestCommandOptions{
			Service: runner.test,
		}),
		coursecli.NewImportCommand(coursecli.ImportCommandOptions{
			Service: runner.imports,
			Config:  runner.config,
		}),
	)

	return root
}

func (runner Runner) instructorID() string {
	if runner.config == nil {
		return ""
	}
	if value := runner.config.GetString("instructor-id"); value != "" {
		return value
	}

	return runner.config.GetString("instructor_id")
}

func (runner Runner) failure(action string, command string, err error, startedAt time.Time) RunResult {
	return RunResult{
		OK:        false,
		Action:    action,
		Command:   command,
		Output:    "error: " + err.Error(),
		ExitCode:  coursecli.ExitValidation,
		ElapsedMS: elapsedMilliseconds(startedAt),
	}
}

func commandArgs(command Command, values map[string]string) []string {
	args := strings.Fields(command.Command)
	for _, field := range command.Fields {
		value, exists := values[field.Key]
		value = strings.TrimSpace(value)
		if field.Binding.BoolFlag {
			if isTruthy(value) {
				args = append(args, field.Binding.Flag)
			}
			continue
		}
		if value == "" && (!exists || (!field.Binding.AllowZero && !field.Binding.AllowEmpty)) {
			continue
		}
		if field.Binding.Argument {
			args = append(args, value)
			continue
		}
		if field.Binding.Flag != "" {
			if field.Binding.MultiValue {
				for _, item := range splitMultiValue(value, field.Binding.Separator) {
					args = append(args, field.Binding.Flag, item)
				}
				continue
			}
			args = append(args, field.Binding.Flag, value)
		}
	}

	return args
}

func splitMultiValue(value string, separator string) []string {
	if separator == "" {
		separator = ","
	}

	parts := strings.Split(value, separator)
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}

	return values
}

func preview(command Command, values map[string]string) string {
	args := append([]string{"course-cli"}, commandArgs(command, values)...)
	return shellQuote(args)
}

func shellQuote(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "" {
			quoted = append(quoted, `""`)
			continue
		}
		if strings.ContainsAny(arg, " \t\n\"'") {
			quoted = append(quoted, `"`+strings.ReplaceAll(arg, `"`, `\"`)+`"`)
			continue
		}
		quoted = append(quoted, arg)
	}

	return strings.Join(quoted, " ")
}

func isTruthy(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func elapsedMilliseconds(startedAt time.Time) int64 {
	elapsed := time.Since(startedAt).Milliseconds()
	if elapsed < 1 {
		return 1
	}

	return elapsed
}
