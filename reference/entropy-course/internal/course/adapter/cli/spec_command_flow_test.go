package cli

import (
	"testing"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/spf13/viper"
)

func TestSpecCourseCommandFlowsCallExactlyOneInboundPort(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "course create",
			args: []string{
				"create",
				"--title", "Intro to Go",
				"--slug", "intro-to-go",
				"--description", "Learn Go",
				"--instructor-id", instructorIDValue,
			},
			want: "create",
		},
		{name: "course list", args: []string{"list", "--status", "draft", "--output", "json"}, want: "list"},
		{name: "course get", args: []string{"get", courseIDValue, "--output", "quiet"}, want: "get"},
		{name: "course update", args: []string{"update", courseIDValue, "--title", "Advanced Go"}, want: "update"},
		{name: "course delete", args: []string{"delete", courseIDValue, "--force"}, want: "delete"},
		{name: "course publish", args: []string{"publish", courseIDValue}, want: "publish"},
		{name: "course unpublish", args: []string{"unpublish", courseIDValue}, want: "unpublish"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &courseServiceFake{
				createOut: core.CreateCourseOutput{ID: courseIDValue},
				listOut:   core.ListCoursesOutput{Courses: []core.CourseView{courseViewFixture()}},
				getOut:    core.GetCourseOutput{Course: courseViewFixture()},
				updateOut: core.UpdateCourseOutput{ID: courseIDValue},
			}

			err := executeCourseCommand(
				NewCourseCommand(CourseCommandOptions{
					Service:  service,
					Renderer: &courseRendererFake{},
					Config:   viper.New(),
				}),
				test.args...,
			)
			if err != nil {
				t.Fatalf("expected command to succeed, got %v", err)
			}

			if service.called != test.want || service.callCount != 1 {
				t.Fatalf("expected exactly one %q call, got called=%q count=%d", test.want, service.called, service.callCount)
			}
		})
	}
}

func TestSpecLessonCommandFlowsCallExactlyOneInboundPort(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "lesson create",
			args: []string{
				"create",
				"--course-id", courseIDValue,
				"--title", "First Lesson",
				"--order", "0",
			},
			want: "create",
		},
		{name: "lesson list", args: []string{"list", "--course-id", courseIDValue, "--output", "json"}, want: "list"},
		{name: "lesson get", args: []string{"get", lessonIDValue, "--output", "quiet"}, want: "get"},
		{name: "lesson update", args: []string{"update", lessonIDValue, "--title", "Updated Lesson"}, want: "update"},
		{name: "lesson delete", args: []string{"delete", lessonIDValue, "--force"}, want: "delete"},
		{
			name: "lesson reorder",
			args: []string{
				"reorder",
				"--course-id", courseIDValue,
				"--order", lessonIDValue + ":0," + otherLessonIDValue + ":1",
			},
			want: "reorder",
		},
		{
			name: "lesson block add",
			args: []string{
				"block",
				"add",
				"--lesson-id", lessonIDValue,
				"--kind", "text",
				"--text", "Content",
			},
			want: "add-block",
		},
		{name: "lesson block list", args: []string{"block", "list", "--lesson-id", lessonIDValue, "--output", "json"}, want: "list-blocks"},
		{name: "lesson block get", args: []string{"block", "get", blockIDValue, "--output", "quiet"}, want: "get-block"},
		{name: "lesson block update", args: []string{"block", "update", blockIDValue, "--text", "Updated"}, want: "update-block"},
		{name: "lesson block remove", args: []string{"block", "remove", blockIDValue, "--force"}, want: "remove-block"},
		{
			name: "lesson block reorder",
			args: []string{
				"block",
				"reorder",
				"--lesson-id", lessonIDValue,
				"--order", blockIDValue + ":0," + otherBlockIDValue + ":1",
			},
			want: "reorder-blocks",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &lessonServiceFake{
				createOut:      core.CreateLessonOutput{ID: lessonIDValue},
				listOut:        core.ListLessonsOutput{Lessons: []core.LessonView{lessonViewFixture()}},
				getOut:         core.GetLessonOutput{Lesson: lessonViewFixture()},
				updateOut:      core.UpdateLessonOutput{ID: lessonIDValue},
				addBlockOut:    core.AddLessonBlockOutput{ID: blockIDValue},
				listBlocksOut:  core.ListLessonBlocksOutput{Blocks: []core.BlockView{blockViewFixture()}},
				getBlockOut:    core.GetLessonBlockOutput{Block: blockViewFixture()},
				updateBlockOut: core.UpdateLessonBlockOutput{ID: blockIDValue},
			}

			err := executeCourseCommand(
				NewLessonCommand(LessonCommandOptions{
					Service:  service,
					Renderer: &lessonRendererFake{},
				}),
				test.args...,
			)
			if err != nil {
				t.Fatalf("expected command to succeed, got %v", err)
			}

			if service.called != test.want || service.callCount != 1 {
				t.Fatalf("expected exactly one %q call, got called=%q count=%d", test.want, service.called, service.callCount)
			}
		})
	}
}

func TestSpecQuizCommandFlowsCallExactlyOneInboundPort(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "quiz create",
			args: []string{
				"create",
				"--course-id", courseIDValue,
				"--title", "Basics Quiz",
				"--pass-threshold", "0.8",
			},
			want: "create",
		},
		{name: "quiz list", args: []string{"list", "--course-id", courseIDValue, "--output", "json"}, want: "list"},
		{name: "quiz get", args: []string{"get", quizIDValue, "--output", "quiet"}, want: "get"},
		{name: "quiz update", args: []string{"update", quizIDValue, "--title", "Advanced Quiz"}, want: "update"},
		{name: "quiz delete", args: []string{"delete", quizIDValue, "--force"}, want: "delete"},
		{
			name: "quiz question add",
			args: []string{
				"question",
				"add",
				"--quiz-id", quizIDValue,
				"--type", "single",
				"--prompt", "Pick one",
				"--option", "A",
				"--option", "B",
				"--correct", "0",
			},
			want: "add-question",
		},
		{name: "quiz question list", args: []string{"question", "list", "--quiz-id", quizIDValue, "--output", "json"}, want: "list-questions"},
		{name: "quiz question get", args: []string{"question", "get", questionIDValue, "--output", "quiet"}, want: "get-question"},
		{name: "quiz question update", args: []string{"question", "update", questionIDValue, "--prompt", "Updated"}, want: "update-question"},
		{name: "quiz question remove", args: []string{"question", "remove", questionIDValue, "--force"}, want: "remove-question"},
		{
			name: "quiz question reorder",
			args: []string{
				"question",
				"reorder",
				"--quiz-id", quizIDValue,
				"--order", questionIDValue + ":0," + otherQuestionIDValue + ":1",
			},
			want: "reorder-questions",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &quizServiceFake{
				createOut:         core.CreateQuizOutput{ID: quizIDValue},
				listOut:           core.ListQuizzesOutput{Quizzes: []core.QuizView{quizViewFixture()}},
				getOut:            core.GetQuizOutput{Quiz: quizDetailFixture()},
				updateOut:         core.UpdateQuizOutput{ID: quizIDValue},
				addQuestionOut:    core.AddQuestionOutput{ID: questionIDValue},
				listQuestionsOut:  core.ListQuestionsOutput{Questions: []core.QuestionView{questionViewFixture()}},
				getQuestionOut:    core.GetQuestionOutput{Question: questionViewFixture()},
				updateQuestionOut: core.UpdateQuestionOutput{ID: questionIDValue},
			}

			err := executeCourseCommand(
				NewQuizCommand(QuizCommandOptions{
					Service:  service,
					Renderer: &quizRendererFake{},
				}),
				test.args...,
			)
			if err != nil {
				t.Fatalf("expected command to succeed, got %v", err)
			}

			if service.called != test.want || service.callCount != 1 {
				t.Fatalf("expected exactly one %q call, got called=%q count=%d", test.want, service.called, service.callCount)
			}
		})
	}
}

func TestSpecPracticeCommandFlowsCallExactlyOneInboundPort(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "practice create",
			args: []string{
				"create",
				"--course-id", courseIDValue,
				"--title", "FizzBuzz",
				"--language", "golang",
				"--prompt", "Print fizz buzz",
			},
			want: "create",
		},
		{name: "practice list", args: []string{"list", "--course-id", courseIDValue, "--output", "json"}, want: "list"},
		{name: "practice get", args: []string{"get", practiceIDValue, "--output", "quiet"}, want: "get"},
		{name: "practice update", args: []string{"update", practiceIDValue, "--title", "Advanced FizzBuzz"}, want: "update"},
		{name: "practice delete", args: []string{"delete", practiceIDValue, "--force"}, want: "delete"},
		{
			name: "practice testcase add",
			args: []string{
				"testcase",
				"add",
				"--practice-id", practiceIDValue,
				"--stdin", "1",
				"--expected-stdout", "1",
			},
			want: "add-testcase",
		},
		{name: "practice testcase list", args: []string{"testcase", "list", "--practice-id", practiceIDValue, "--output", "json"}, want: "list-testcases"},
		{name: "practice testcase get", args: []string{"testcase", "get", testCaseIDValue, "--output", "quiet"}, want: "get-testcase"},
		{name: "practice testcase update", args: []string{"testcase", "update", testCaseIDValue, "--stdin", "updated"}, want: "update-testcase"},
		{name: "practice testcase remove", args: []string{"testcase", "remove", testCaseIDValue, "--force"}, want: "remove-testcase"},
		{
			name: "practice testcase reorder",
			args: []string{
				"testcase",
				"reorder",
				"--practice-id", practiceIDValue,
				"--order", testCaseIDValue + ":0," + otherTestCaseIDValue + ":1",
			},
			want: "reorder-testcases",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &practiceServiceFake{
				createOut:         core.CreatePracticeOutput{ID: practiceIDValue},
				listOut:           core.ListPracticesOutput{Practices: []core.PracticeView{practiceViewFixture()}},
				getOut:            core.GetPracticeOutput{Practice: practiceDetailFixture()},
				updateOut:         core.UpdatePracticeOutput{ID: practiceIDValue},
				addTestCaseOut:    core.AddTestCaseOutput{ID: testCaseIDValue},
				listTestCasesOut:  core.ListTestCasesOutput{TestCases: []core.TestCaseView{testCaseViewFixture()}},
				getTestCaseOut:    core.GetTestCaseOutput{TestCase: testCaseViewFixture()},
				updateTestCaseOut: core.UpdateTestCaseOutput{ID: testCaseIDValue},
			}

			err := executeCourseCommand(
				NewPracticeCommand(PracticeCommandOptions{
					Service:  service,
					Renderer: &practiceRendererFake{},
				}),
				test.args...,
			)
			if err != nil {
				t.Fatalf("expected command to succeed, got %v", err)
			}

			if service.called != test.want || service.callCount != 1 {
				t.Fatalf("expected exactly one %q call, got called=%q count=%d", test.want, service.called, service.callCount)
			}
		})
	}
}

func TestSpecTestCommandFlowsCallExactlyOneInboundPort(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "test create",
			args: []string{
				"create",
				"--course-id", courseIDValue,
				"--title", "Final Test",
				"--time-limit-minutes", "45",
			},
			want: "create",
		},
		{name: "test list", args: []string{"list", "--course-id", courseIDValue, "--output", "json"}, want: "list"},
		{name: "test get", args: []string{"get", testIDValue, "--output", "quiet"}, want: "get"},
		{name: "test update", args: []string{"update", testIDValue, "--title", "Updated Test"}, want: "update"},
		{name: "test delete", args: []string{"delete", testIDValue, "--force"}, want: "delete"},
		{
			name: "test item add choice",
			args: []string{
				"item",
				"add",
				"--test-id", testIDValue,
				"--kind", "choice",
				"--prompt", "Pick one",
				"--type", "single",
				"--option", "A",
				"--option", "B",
				"--correct", "0",
			},
			want: "add-item",
		},
		{name: "test item list", args: []string{"item", "list", "--test-id", testIDValue, "--output", "json"}, want: "list-items"},
		{name: "test item get", args: []string{"item", "get", testItemIDValue, "--output", "quiet"}, want: "get-item"},
		{name: "test item update", args: []string{"item", "update", testItemIDValue, "--prompt", "Updated"}, want: "update-item"},
		{name: "test item remove", args: []string{"item", "remove", testItemIDValue, "--force"}, want: "remove-item"},
		{
			name: "test item reorder",
			args: []string{
				"item",
				"reorder",
				"--test-id", testIDValue,
				"--order", testItemIDValue + ":0," + otherTestItemIDValue + ":1",
			},
			want: "reorder-items",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &testServiceFake{
				createOut:     core.CreateTestOutput{ID: testIDValue},
				listOut:       core.ListTestsOutput{Tests: []core.TestView{testViewFixture()}},
				getOut:        core.GetTestOutput{Test: testDetailFixture()},
				updateOut:     core.UpdateTestOutput{ID: testIDValue},
				addItemOut:    core.AddTestItemOutput{ID: testItemIDValue},
				listItemsOut:  core.ListTestItemsOutput{Items: []core.TestItemView{testItemViewFixture()}},
				getItemOut:    core.GetTestItemOutput{Item: testItemViewFixture()},
				updateItemOut: core.UpdateTestItemOutput{ID: testItemIDValue},
			}

			err := executeCourseCommand(
				NewTestCommand(TestCommandOptions{
					Service:  service,
					Renderer: &testRendererFake{},
				}),
				test.args...,
			)
			if err != nil {
				t.Fatalf("expected command to succeed, got %v", err)
			}

			if service.called != test.want || service.callCount != 1 {
				t.Fatalf("expected exactly one %q call, got called=%q count=%d", test.want, service.called, service.callCount)
			}
		})
	}
}
