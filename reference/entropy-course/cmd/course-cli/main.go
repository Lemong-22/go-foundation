package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	coursecli "github.com/luxeave/entropy-course/internal/course/adapter/cli"
	"github.com/luxeave/entropy-course/internal/course/adapter/playground"
	"github.com/luxeave/entropy-course/internal/course/app"
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	command := newRootCommand(ctx)
	if err := command.ExecuteContext(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(coursecli.ExitCode(err))
	}
}

func newRootCommand(ctx context.Context) *cobra.Command {
	config := app.ConfigureViper(viper.New())
	scope := &containerScope{ctx: ctx, config: config}

	command := &cobra.Command{
		Use:           "course-cli",
		Short:         "Manage courses, lessons, quizzes, practices, and tests",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	command.PersistentFlags().String("db-url", "", "PostgreSQL connection string")
	command.PersistentFlags().String("api-token", "", "REST API bearer token")
	_ = config.BindPFlag("db-url", command.PersistentFlags().Lookup("db-url"))
	_ = config.BindPFlag("api-token", command.PersistentFlags().Lookup("api-token"))

	command.PersistentPostRun = func(*cobra.Command, []string) {
		scope.Close()
	}

	command.AddCommand(
		coursecli.NewCourseCommand(coursecli.CourseCommandOptions{
			Service: deferredCourseService{scope: scope},
			Config:  config,
		}),
		coursecli.NewLessonCommand(coursecli.LessonCommandOptions{
			Service: deferredLessonService{scope: scope},
		}),
		coursecli.NewQuizCommand(coursecli.QuizCommandOptions{
			Service: deferredQuizService{scope: scope},
		}),
		coursecli.NewPracticeCommand(coursecli.PracticeCommandOptions{
			Service: deferredPracticeService{scope: scope},
		}),
		coursecli.NewTestCommand(coursecli.TestCommandOptions{
			Service: deferredTestService{scope: scope},
		}),
		coursecli.NewImportCommand(coursecli.ImportCommandOptions{
			Service: deferredImportService{scope: scope},
			Config:  config,
		}),
		newMigrateCommand(ctx, config),
		newPlaygroundCommand(ctx, scope, config),
		newRestCommand(ctx, scope),
	)

	return command
}

func newPlaygroundCommand(ctx context.Context, scope *containerScope, config *viper.Viper) *cobra.Command {
	address := playground.DefaultAddress
	command := &cobra.Command{
		Use:   "playground",
		Short: "Serve the loopback browser playground",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			container, err := scope.Container()
			if err != nil {
				return err
			}

			server, err := playground.NewServer(playground.ServerOptions{
				Runner: playground.NewRunner(playground.Services{
					Course:   container.Course,
					Lesson:   container.Lesson,
					Quiz:     container.Quiz,
					Practice: container.Practice,
					Test:     container.Test,
					Import:   container.Import,
					Config:   config,
				}),
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(os.Stdout, "Course CLI playground listening on http://%s\n", address)
			return server.ListenAndServe(ctx, address)
		},
	}
	command.Flags().StringVar(&address, "addr", playground.DefaultAddress, "loopback address to bind")

	return command
}

type containerScope struct {
	ctx    context.Context
	config *viper.Viper

	mu        sync.Mutex
	container *app.CLI
}

func (scope *containerScope) Container() (*app.CLI, error) {
	scope.mu.Lock()
	defer scope.mu.Unlock()

	if scope.container != nil {
		return scope.container, nil
	}

	cfg, err := app.LoadConfig(scope.config)
	if err != nil {
		return nil, err
	}

	container, err := app.BuildContainer(scope.ctx, cfg)
	if err != nil {
		return nil, err
	}

	scope.container = container
	return container, nil
}

func (scope *containerScope) Close() {
	scope.mu.Lock()
	defer scope.mu.Unlock()

	if scope.container != nil {
		scope.container.Close()
		scope.container = nil
	}
}

type deferredCourseService struct {
	scope *containerScope
}

func (service deferredCourseService) CreateCourse(in core.CreateCourseInput) (core.CreateCourseOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.CreateCourseOutput{}, err
	}

	return container.Course.CreateCourse(in)
}

func (service deferredCourseService) ListCourses(in core.ListCoursesInput) (core.ListCoursesOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ListCoursesOutput{}, err
	}

	return container.Course.ListCourses(in)
}

func (service deferredCourseService) GetCourse(in core.GetCourseInput) (core.GetCourseOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.GetCourseOutput{}, err
	}

	return container.Course.GetCourse(in)
}

func (service deferredCourseService) UpdateCourse(in core.UpdateCourseInput) (core.UpdateCourseOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.UpdateCourseOutput{}, err
	}

	return container.Course.UpdateCourse(in)
}

func (service deferredCourseService) DeleteCourse(in core.DeleteCourseInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Course.DeleteCourse(in)
}

func (service deferredCourseService) PublishCourse(in core.PublishCourseInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Course.PublishCourse(in)
}

func (service deferredCourseService) UnpublishCourse(in core.UnpublishCourseInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Course.UnpublishCourse(in)
}

type deferredLessonService struct {
	scope *containerScope
}

func (service deferredLessonService) CreateLesson(in core.CreateLessonInput) (core.CreateLessonOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.CreateLessonOutput{}, err
	}

	return container.Lesson.CreateLesson(in)
}

func (service deferredLessonService) ListLessons(in core.ListLessonsInput) (core.ListLessonsOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ListLessonsOutput{}, err
	}

	return container.Lesson.ListLessons(in)
}

func (service deferredLessonService) GetLesson(in core.GetLessonInput) (core.GetLessonOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.GetLessonOutput{}, err
	}

	return container.Lesson.GetLesson(in)
}

func (service deferredLessonService) UpdateLesson(in core.UpdateLessonInput) (core.UpdateLessonOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.UpdateLessonOutput{}, err
	}

	return container.Lesson.UpdateLesson(in)
}

func (service deferredLessonService) DeleteLesson(in core.DeleteLessonInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Lesson.DeleteLesson(in)
}

func (service deferredLessonService) ReorderLessons(in core.ReorderLessonsInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Lesson.ReorderLessons(in)
}

func (service deferredLessonService) AddLessonBlock(in core.AddLessonBlockInput) (core.AddLessonBlockOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.AddLessonBlockOutput{}, err
	}

	return container.Lesson.AddLessonBlock(in)
}

func (service deferredLessonService) ListLessonBlocks(in core.ListLessonBlocksInput) (core.ListLessonBlocksOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ListLessonBlocksOutput{}, err
	}

	return container.Lesson.ListLessonBlocks(in)
}

func (service deferredLessonService) GetLessonBlock(in core.GetLessonBlockInput) (core.GetLessonBlockOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.GetLessonBlockOutput{}, err
	}

	return container.Lesson.GetLessonBlock(in)
}

func (service deferredLessonService) UpdateLessonBlock(in core.UpdateLessonBlockInput) (core.UpdateLessonBlockOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.UpdateLessonBlockOutput{}, err
	}

	return container.Lesson.UpdateLessonBlock(in)
}

func (service deferredLessonService) RemoveLessonBlock(in core.RemoveLessonBlockInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Lesson.RemoveLessonBlock(in)
}

func (service deferredLessonService) ReorderLessonBlocks(in core.ReorderLessonBlocksInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Lesson.ReorderLessonBlocks(in)
}

type deferredQuizService struct {
	scope *containerScope
}

func (service deferredQuizService) CreateQuiz(in core.CreateQuizInput) (core.CreateQuizOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.CreateQuizOutput{}, err
	}

	return container.Quiz.CreateQuiz(in)
}

func (service deferredQuizService) ListQuizzes(in core.ListQuizzesInput) (core.ListQuizzesOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ListQuizzesOutput{}, err
	}

	return container.Quiz.ListQuizzes(in)
}

func (service deferredQuizService) GetQuiz(in core.GetQuizInput) (core.GetQuizOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.GetQuizOutput{}, err
	}

	return container.Quiz.GetQuiz(in)
}

func (service deferredQuizService) UpdateQuiz(in core.UpdateQuizInput) (core.UpdateQuizOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.UpdateQuizOutput{}, err
	}

	return container.Quiz.UpdateQuiz(in)
}

func (service deferredQuizService) DeleteQuiz(in core.DeleteQuizInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Quiz.DeleteQuiz(in)
}

func (service deferredQuizService) AddQuestion(in core.AddQuestionInput) (core.AddQuestionOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.AddQuestionOutput{}, err
	}

	return container.Quiz.AddQuestion(in)
}

func (service deferredQuizService) ListQuestions(in core.ListQuestionsInput) (core.ListQuestionsOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ListQuestionsOutput{}, err
	}

	return container.Quiz.ListQuestions(in)
}

func (service deferredQuizService) GetQuestion(in core.GetQuestionInput) (core.GetQuestionOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.GetQuestionOutput{}, err
	}

	return container.Quiz.GetQuestion(in)
}

func (service deferredQuizService) UpdateQuestion(in core.UpdateQuestionInput) (core.UpdateQuestionOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.UpdateQuestionOutput{}, err
	}

	return container.Quiz.UpdateQuestion(in)
}

func (service deferredQuizService) RemoveQuestion(in core.RemoveQuestionInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Quiz.RemoveQuestion(in)
}

func (service deferredQuizService) ReorderQuestions(in core.ReorderQuestionsInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Quiz.ReorderQuestions(in)
}

type deferredPracticeService struct {
	scope *containerScope
}

func (service deferredPracticeService) CreatePractice(in core.CreatePracticeInput) (core.CreatePracticeOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.CreatePracticeOutput{}, err
	}

	return container.Practice.CreatePractice(in)
}

func (service deferredPracticeService) ListPractices(in core.ListPracticesInput) (core.ListPracticesOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ListPracticesOutput{}, err
	}

	return container.Practice.ListPractices(in)
}

func (service deferredPracticeService) GetPractice(in core.GetPracticeInput) (core.GetPracticeOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.GetPracticeOutput{}, err
	}

	return container.Practice.GetPractice(in)
}

func (service deferredPracticeService) UpdatePractice(in core.UpdatePracticeInput) (core.UpdatePracticeOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.UpdatePracticeOutput{}, err
	}

	return container.Practice.UpdatePractice(in)
}

func (service deferredPracticeService) DeletePractice(in core.DeletePracticeInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Practice.DeletePractice(in)
}

func (service deferredPracticeService) AddTestCase(in core.AddTestCaseInput) (core.AddTestCaseOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.AddTestCaseOutput{}, err
	}

	return container.Practice.AddTestCase(in)
}

func (service deferredPracticeService) ListTestCases(in core.ListTestCasesInput) (core.ListTestCasesOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ListTestCasesOutput{}, err
	}

	return container.Practice.ListTestCases(in)
}

func (service deferredPracticeService) GetTestCase(in core.GetTestCaseInput) (core.GetTestCaseOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.GetTestCaseOutput{}, err
	}

	return container.Practice.GetTestCase(in)
}

func (service deferredPracticeService) UpdateTestCase(in core.UpdateTestCaseInput) (core.UpdateTestCaseOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.UpdateTestCaseOutput{}, err
	}

	return container.Practice.UpdateTestCase(in)
}

func (service deferredPracticeService) RemoveTestCase(in core.RemoveTestCaseInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Practice.RemoveTestCase(in)
}

func (service deferredPracticeService) ReorderTestCases(in core.ReorderTestCasesInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Practice.ReorderTestCases(in)
}

type deferredTestService struct {
	scope *containerScope
}

func (service deferredTestService) CreateTest(in core.CreateTestInput) (core.CreateTestOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.CreateTestOutput{}, err
	}

	return container.Test.CreateTest(in)
}

func (service deferredTestService) ListTests(in core.ListTestsInput) (core.ListTestsOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ListTestsOutput{}, err
	}

	return container.Test.ListTests(in)
}

func (service deferredTestService) GetTest(in core.GetTestInput) (core.GetTestOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.GetTestOutput{}, err
	}

	return container.Test.GetTest(in)
}

func (service deferredTestService) UpdateTest(in core.UpdateTestInput) (core.UpdateTestOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.UpdateTestOutput{}, err
	}

	return container.Test.UpdateTest(in)
}

func (service deferredTestService) DeleteTest(in core.DeleteTestInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Test.DeleteTest(in)
}

func (service deferredTestService) AddTestItem(in core.AddTestItemInput) (core.AddTestItemOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.AddTestItemOutput{}, err
	}

	return container.Test.AddTestItem(in)
}

func (service deferredTestService) ListTestItems(in core.ListTestItemsInput) (core.ListTestItemsOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ListTestItemsOutput{}, err
	}

	return container.Test.ListTestItems(in)
}

func (service deferredTestService) GetTestItem(in core.GetTestItemInput) (core.GetTestItemOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.GetTestItemOutput{}, err
	}

	return container.Test.GetTestItem(in)
}

func (service deferredTestService) UpdateTestItem(in core.UpdateTestItemInput) (core.UpdateTestItemOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.UpdateTestItemOutput{}, err
	}

	return container.Test.UpdateTestItem(in)
}

func (service deferredTestService) RemoveTestItem(in core.RemoveTestItemInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Test.RemoveTestItem(in)
}

func (service deferredTestService) ReorderTestItems(in core.ReorderTestItemsInput) error {
	container, err := service.scope.Container()
	if err != nil {
		return err
	}

	return container.Test.ReorderTestItems(in)
}

type deferredImportService struct {
	scope *containerScope
}

func (service deferredImportService) PlanImport(in core.PlanImportInput) (core.PlanImportOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.PlanImportOutput{}, err
	}
	if container.Import == nil {
		return core.PlanImportOutput{}, coursecli.ErrMissingImportService
	}

	return container.Import.PlanImport(in)
}

func (service deferredImportService) ApplyPlan(in core.ApplyPlanInput) (core.ApplyPlanOutput, error) {
	container, err := service.scope.Container()
	if err != nil {
		return core.ApplyPlanOutput{}, err
	}
	if container.Import == nil {
		return core.ApplyPlanOutput{}, coursecli.ErrMissingImportService
	}

	return container.Import.ApplyPlan(in)
}
