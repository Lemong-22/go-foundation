package cli

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	courseIDValue     = "550e8400-e29b-41d4-a716-446655440000"
	instructorIDValue = "550e8400-e29b-41d4-a716-446655440010"
)

func TestCourseCommandExposesRequiredSubcommands(t *testing.T) {
	command := NewCourseCommand(CourseCommandOptions{Service: &courseServiceFake{}, Config: viper.New()})

	wantCommands := []string{"create", "list", "get", "update", "delete", "publish", "unpublish"}
	for _, name := range wantCommands {
		if _, _, err := command.Find([]string{name}); err != nil {
			t.Fatalf("expected course %s command to exist, got %v", name, err)
		}
	}
}

func TestCourseCreateMapsFlagsAndConfigInstructorToDTO(t *testing.T) {
	service := &courseServiceFake{createOut: core.CreateCourseOutput{ID: courseIDValue}}
	renderer := &courseRendererFake{}
	config := viper.New()
	config.Set("instructor-id", instructorIDValue)

	err := executeCourseCommand(
		NewCourseCommand(CourseCommandOptions{Service: service, Renderer: renderer, Config: config}),
		"create",
		"--title", "Intro to Go",
		"--slug", "intro-to-go",
		"--description", "Learn Go",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "create" {
		t.Fatalf("expected create to be called, got %q", service.called)
	}
	want := core.CreateCourseInput{
		Title:        "Intro to Go",
		Slug:         "intro-to-go",
		Description:  "Learn Go",
		InstructorID: instructorIDValue,
	}
	if service.createIn != want {
		t.Fatalf("expected create input %+v, got %+v", want, service.createIn)
	}
	if renderer.createdID != courseIDValue {
		t.Fatalf("expected renderer to receive created id")
	}
}

func TestCourseCreateMapsEnvInstructorToDTO(t *testing.T) {
	t.Setenv("COURSE_CLI_INSTRUCTOR_ID", instructorIDValue)
	service := &courseServiceFake{createOut: core.CreateCourseOutput{ID: courseIDValue}}

	err := executeCourseCommand(
		NewCourseCommand(CourseCommandOptions{Service: service, Config: viper.New()}),
		"create",
		"--title", "Intro to Go",
		"--slug", "intro-to-go",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.createIn.InstructorID != instructorIDValue {
		t.Fatalf("expected instructor id from environment, got %q", service.createIn.InstructorID)
	}
}

func TestCourseCreateRequiresTitle(t *testing.T) {
	service := &courseServiceFake{}

	err := executeCourseCommand(
		NewCourseCommand(CourseCommandOptions{Service: service, Config: viper.New()}),
		"create",
		"--slug", "intro-to-go",
		"--instructor-id", instructorIDValue,
	)
	if !errors.Is(err, ErrRequiredFlagMissing) {
		t.Fatalf("expected required flag error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

func TestCourseListMapsStatusAndOutputFormat(t *testing.T) {
	course := courseViewFixture()
	service := &courseServiceFake{listOut: core.ListCoursesOutput{Courses: []core.CourseView{course}}}
	renderer := &courseRendererFake{}

	err := executeCourseCommand(
		NewCourseCommand(CourseCommandOptions{Service: service, Renderer: renderer, Config: viper.New()}),
		"list",
		"--status", "published",
		"--output", "json",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "list" || service.listIn.Status != "published" {
		t.Fatalf("expected list with status filter, got called=%q input=%+v", service.called, service.listIn)
	}
	if renderer.listFormat != "json" || len(renderer.courses) != 1 || renderer.courses[0] != course {
		t.Fatalf("expected renderer to receive json course list")
	}
}

func TestCourseGetMapsIDAndOutputFormat(t *testing.T) {
	course := courseViewFixture()
	service := &courseServiceFake{getOut: core.GetCourseOutput{Course: course}}
	renderer := &courseRendererFake{}

	err := executeCourseCommand(
		NewCourseCommand(CourseCommandOptions{Service: service, Renderer: renderer, Config: viper.New()}),
		"get",
		courseIDValue,
		"-o", "quiet",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "get" || service.getIn.ID != courseIDValue {
		t.Fatalf("expected get with id, got called=%q input=%+v", service.called, service.getIn)
	}
	if renderer.courseFormat != "quiet" || renderer.course != course {
		t.Fatalf("expected renderer to receive quiet course view")
	}
}

func TestCourseUpdateMapsOnlyChangedFlags(t *testing.T) {
	service := &courseServiceFake{updateOut: core.UpdateCourseOutput{ID: courseIDValue}}
	renderer := &courseRendererFake{}

	err := executeCourseCommand(
		NewCourseCommand(CourseCommandOptions{Service: service, Renderer: renderer, Config: viper.New()}),
		"update",
		courseIDValue,
		"--title", "Advanced Go",
		"--description", "",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "update" || service.updateIn.ID != courseIDValue {
		t.Fatalf("expected update call, got called=%q input=%+v", service.called, service.updateIn)
	}
	if service.updateIn.Title == nil || *service.updateIn.Title != "Advanced Go" {
		t.Fatalf("expected title pointer to be set")
	}
	if service.updateIn.Description == nil || *service.updateIn.Description != "" {
		t.Fatalf("expected changed empty description to be set")
	}
	if service.updateIn.Slug != nil {
		t.Fatalf("expected unchanged slug to remain nil")
	}
	if renderer.updatedID != courseIDValue {
		t.Fatalf("expected renderer to receive updated id")
	}
}

func TestCourseDeleteRequiresConfirmation(t *testing.T) {
	service := &courseServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}

	err := executeCourseCommand(
		NewCourseCommand(CourseCommandOptions{Service: service, Prompter: prompter, Config: viper.New()}),
		"delete",
		courseIDValue,
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

func TestCourseDeleteForceSkipsConfirmation(t *testing.T) {
	service := &courseServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}
	renderer := &courseRendererFake{}

	err := executeCourseCommand(
		NewCourseCommand(CourseCommandOptions{
			Service:  service,
			Renderer: renderer,
			Prompter: prompter,
			Config:   viper.New(),
		}),
		"delete",
		courseIDValue,
		"--force",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "delete" || service.deleteIn.ID != courseIDValue {
		t.Fatalf("expected delete with id, got called=%q input=%+v", service.called, service.deleteIn)
	}
	if prompter.message != "" {
		t.Fatalf("expected force delete not to prompt")
	}
	if renderer.confirmation != "course deleted" {
		t.Fatalf("expected delete confirmation to render")
	}
}

func TestCoursePublishAndUnpublishMapID(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "publish", args: []string{"publish", courseIDValue}, want: "publish"},
		{name: "unpublish", args: []string{"unpublish", courseIDValue}, want: "unpublish"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &courseServiceFake{}

			err := executeCourseCommand(
				NewCourseCommand(CourseCommandOptions{Service: service, Config: viper.New()}),
				test.args...,
			)
			if err != nil {
				t.Fatalf("expected command to succeed, got %v", err)
			}

			if service.called != test.want {
				t.Fatalf("expected %q call, got %q", test.want, service.called)
			}
		})
	}
}

func executeCourseCommand(command *cobra.Command, args ...string) error {
	command.SetArgs(args)
	command.SetOut(io.Discard)
	command.SetErr(io.Discard)
	return command.Execute()
}

type courseServiceFake struct {
	called    string
	callCount int

	createIn    core.CreateCourseInput
	createOut   core.CreateCourseOutput
	listIn      core.ListCoursesInput
	listOut     core.ListCoursesOutput
	getIn       core.GetCourseInput
	getOut      core.GetCourseOutput
	updateIn    core.UpdateCourseInput
	updateOut   core.UpdateCourseOutput
	deleteIn    core.DeleteCourseInput
	publishIn   core.PublishCourseInput
	unpublishIn core.UnpublishCourseInput
}

func (service *courseServiceFake) CreateCourse(in core.CreateCourseInput) (core.CreateCourseOutput, error) {
	service.called = "create"
	service.callCount++
	service.createIn = in
	return service.createOut, nil
}

func (service *courseServiceFake) ListCourses(in core.ListCoursesInput) (core.ListCoursesOutput, error) {
	service.called = "list"
	service.callCount++
	service.listIn = in
	return service.listOut, nil
}

func (service *courseServiceFake) GetCourse(in core.GetCourseInput) (core.GetCourseOutput, error) {
	service.called = "get"
	service.callCount++
	service.getIn = in
	return service.getOut, nil
}

func (service *courseServiceFake) UpdateCourse(in core.UpdateCourseInput) (core.UpdateCourseOutput, error) {
	service.called = "update"
	service.callCount++
	service.updateIn = in
	return service.updateOut, nil
}

func (service *courseServiceFake) DeleteCourse(in core.DeleteCourseInput) error {
	service.called = "delete"
	service.callCount++
	service.deleteIn = in
	return nil
}

func (service *courseServiceFake) PublishCourse(in core.PublishCourseInput) error {
	service.called = "publish"
	service.callCount++
	service.publishIn = in
	return nil
}

func (service *courseServiceFake) UnpublishCourse(in core.UnpublishCourseInput) error {
	service.called = "unpublish"
	service.callCount++
	service.unpublishIn = in
	return nil
}

type courseRendererFake struct {
	createdID    string
	updatedID    string
	listFormat   string
	courseFormat string
	courses      []core.CourseView
	course       core.CourseView
	confirmation string
}

func (renderer *courseRendererFake) RenderCreatedCourse(id string) error {
	renderer.createdID = id
	return nil
}

func (renderer *courseRendererFake) RenderCourseList(format string, courses []core.CourseView) error {
	renderer.listFormat = format
	renderer.courses = courses
	return nil
}

func (renderer *courseRendererFake) RenderCourse(format string, course core.CourseView) error {
	renderer.courseFormat = format
	renderer.course = course
	return nil
}

func (renderer *courseRendererFake) RenderUpdatedCourse(id string) error {
	renderer.updatedID = id
	return nil
}

func (renderer *courseRendererFake) RenderConfirmation(message string) error {
	renderer.confirmation = message
	return nil
}

type coursePrompterFake struct {
	confirmed bool
	message   string
}

func (prompter *coursePrompterFake) Confirm(message string) (bool, error) {
	prompter.message = message
	return prompter.confirmed, nil
}

func courseViewFixture() core.CourseView {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	return core.CourseView{
		ID:           courseIDValue,
		Title:        "Intro to Go",
		Slug:         "intro-to-go",
		Description:  "Learn Go",
		InstructorID: instructorIDValue,
		Status:       "published",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
