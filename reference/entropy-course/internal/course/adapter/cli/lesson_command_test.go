package cli

import (
	"errors"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
)

const (
	lessonIDValue      = "550e8400-e29b-41d4-a716-446655440020"
	otherLessonIDValue = "550e8400-e29b-41d4-a716-446655440021"
	blockIDValue       = "550e8400-e29b-41d4-a716-446655440030"
	otherBlockIDValue  = "550e8400-e29b-41d4-a716-446655440031"
	quizIDValue        = "550e8400-e29b-41d4-a716-446655440040"
	practiceIDValue    = "550e8400-e29b-41d4-a716-446655440060"
)

func TestLessonCommandExposesRequiredSubcommands(t *testing.T) {
	command := NewLessonCommand(LessonCommandOptions{Service: &lessonServiceFake{}})

	wantCommands := []string{"create", "list", "get", "update", "delete", "reorder", "block"}
	for _, name := range wantCommands {
		if _, _, err := command.Find([]string{name}); err != nil {
			t.Fatalf("expected lesson %s command to exist, got %v", name, err)
		}
	}

	wantBlockCommands := []string{"add", "list", "get", "update", "remove", "reorder"}
	for _, name := range wantBlockCommands {
		if _, _, err := command.Find([]string{"block", name}); err != nil {
			t.Fatalf("expected lesson block %s command to exist, got %v", name, err)
		}
	}
}

func TestLessonCreateMapsFlagsToDTOWithExplicitOrder(t *testing.T) {
	service := &lessonServiceFake{createOut: core.CreateLessonOutput{ID: lessonIDValue}}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"create",
		"--course-id", courseIDValue,
		"--title", "First Lesson",
		"--order", "3",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "create" || service.callCount != 1 {
		t.Fatalf("expected one create call, got called=%q count=%d", service.called, service.callCount)
	}
	if service.createIn.CourseID != courseIDValue || service.createIn.Title != "First Lesson" {
		t.Fatalf("expected course id and title to map, got %+v", service.createIn)
	}
	if service.createIn.Order == nil || *service.createIn.Order != 3 {
		t.Fatalf("expected explicit order pointer 3, got %v", service.createIn.Order)
	}
	if renderer.createdID != lessonIDValue {
		t.Fatalf("expected renderer to receive created id")
	}
}

func TestLessonCreateLeavesOrderNilWhenUnset(t *testing.T) {
	service := &lessonServiceFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service}),
		"create",
		"--course-id", courseIDValue,
		"--title", "First Lesson",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.createIn.Order != nil {
		t.Fatalf("expected nil order for append-by-default, got %v", *service.createIn.Order)
	}
}

func TestLessonCommandsDoNotExposeContentFlags(t *testing.T) {
	command := NewLessonCommand(LessonCommandOptions{Service: &lessonServiceFake{}})

	createCommand, _, err := command.Find([]string{"create"})
	if err != nil {
		t.Fatalf("expected lesson create command, got %v", err)
	}
	if createCommand.Flags().Lookup("content") != nil {
		t.Fatalf("expected lesson create to omit content flag")
	}

	updateCommand, _, err := command.Find([]string{"update"})
	if err != nil {
		t.Fatalf("expected lesson update command, got %v", err)
	}
	if updateCommand.Flags().Lookup("content") != nil {
		t.Fatalf("expected lesson update to omit content flag")
	}
}

func TestLessonCreateRequiresCourseIDAndTitle(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "course id", args: []string{"create", "--title", "First Lesson"}},
		{name: "title", args: []string{"create", "--course-id", courseIDValue}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &lessonServiceFake{}

			err := executeCourseCommand(NewLessonCommand(LessonCommandOptions{Service: service}), test.args...)
			if !errors.Is(err, ErrRequiredFlagMissing) {
				t.Fatalf("expected required flag error, got %v", err)
			}
			if service.called != "" {
				t.Fatalf("expected service not to be called, got %q", service.called)
			}
		})
	}
}

func TestLessonListMapsCourseIDAndOutputFormat(t *testing.T) {
	lesson := lessonViewFixture()
	service := &lessonServiceFake{listOut: core.ListLessonsOutput{Lessons: []core.LessonView{lesson}}}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"list",
		"--course-id", courseIDValue,
		"-o", "json",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "list" || service.callCount != 1 || service.listIn.CourseID != courseIDValue {
		t.Fatalf("expected one list call, got called=%q count=%d input=%+v", service.called, service.callCount, service.listIn)
	}
	if renderer.listFormat != "json" || len(renderer.lessons) != 1 || renderer.lessons[0] != lesson {
		t.Fatalf("expected renderer to receive json lesson list")
	}
}

func TestLessonGetMapsIDAndOutputFormat(t *testing.T) {
	lesson := lessonViewFixture()
	service := &lessonServiceFake{getOut: core.GetLessonOutput{Lesson: lesson}}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"get",
		lessonIDValue,
		"-o", "quiet",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "get" || service.callCount != 1 || service.getIn.ID != lessonIDValue {
		t.Fatalf("expected one get call, got called=%q count=%d input=%+v", service.called, service.callCount, service.getIn)
	}
	if renderer.lessonFormat != "quiet" || renderer.lesson != lesson {
		t.Fatalf("expected renderer to receive quiet lesson view")
	}
}

func TestLessonUpdateMapsOnlyChangedFlags(t *testing.T) {
	service := &lessonServiceFake{updateOut: core.UpdateLessonOutput{ID: lessonIDValue}}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"update",
		lessonIDValue,
		"--title", "Updated Lesson",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "update" || service.callCount != 1 || service.updateIn.ID != lessonIDValue {
		t.Fatalf("expected one update call, got called=%q count=%d input=%+v", service.called, service.callCount, service.updateIn)
	}
	if service.updateIn.Title == nil || *service.updateIn.Title != "Updated Lesson" {
		t.Fatalf("expected title pointer to be set")
	}
	if renderer.updatedID != lessonIDValue {
		t.Fatalf("expected renderer to receive updated id")
	}
}

func TestLessonDeleteRequiresConfirmation(t *testing.T) {
	service := &lessonServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Prompter: prompter}),
		"delete",
		lessonIDValue,
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

func TestLessonDeleteForceSkipsConfirmation(t *testing.T) {
	service := &lessonServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{
			Service:  service,
			Renderer: renderer,
			Prompter: prompter,
		}),
		"delete",
		lessonIDValue,
		"--force",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "delete" || service.callCount != 1 || service.deleteIn.ID != lessonIDValue {
		t.Fatalf("expected one delete call, got called=%q count=%d input=%+v", service.called, service.callCount, service.deleteIn)
	}
	if prompter.message != "" {
		t.Fatalf("expected force delete not to prompt")
	}
	if renderer.confirmation != "lesson deleted" {
		t.Fatalf("expected delete confirmation to render")
	}
}

func TestLessonReorderParsesOrderPairs(t *testing.T) {
	service := &lessonServiceFake{}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"reorder",
		"--course-id", courseIDValue,
		"--order", lessonIDValue+":2, "+otherLessonIDValue+":0",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "reorder" || service.callCount != 1 || service.reorderIn.CourseID != courseIDValue {
		t.Fatalf("expected one reorder call, got called=%q count=%d input=%+v", service.called, service.callCount, service.reorderIn)
	}

	want := []core.LessonPosition{
		{LessonID: lessonIDValue, Position: 2},
		{LessonID: otherLessonIDValue, Position: 0},
	}
	if len(service.reorderIn.Order) != len(want) {
		t.Fatalf("expected %d positions, got %d", len(want), len(service.reorderIn.Order))
	}
	for index := range want {
		if service.reorderIn.Order[index] != want[index] {
			t.Fatalf("expected position %+v at index %d, got %+v", want[index], index, service.reorderIn.Order[index])
		}
	}
	if renderer.confirmation != "lessons reordered" {
		t.Fatalf("expected reorder confirmation to render")
	}
}

func TestLessonReorderRejectsInvalidOrder(t *testing.T) {
	service := &lessonServiceFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service}),
		"reorder",
		"--course-id", courseIDValue,
		"--order", lessonIDValue,
	)
	if !errors.Is(err, ErrInvalidLessonOrder) {
		t.Fatalf("expected invalid lesson order error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

func TestLessonBlockAddMapsTextFlagsToDTO(t *testing.T) {
	service := &lessonServiceFake{addBlockOut: core.AddLessonBlockOutput{ID: blockIDValue}}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"block",
		"add",
		"--lesson-id", lessonIDValue,
		"--kind", "text",
		"--text", "## Intro",
		"--position", "2",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "add-block" || service.callCount != 1 {
		t.Fatalf("expected one add block call, got called=%q count=%d", service.called, service.callCount)
	}
	if service.addBlockIn.LessonID != lessonIDValue || service.addBlockIn.Kind != "text" || service.addBlockIn.Markdown != "## Intro" {
		t.Fatalf("expected text block input to map, got %+v", service.addBlockIn)
	}
	if service.addBlockIn.Position == nil || *service.addBlockIn.Position != 2 {
		t.Fatalf("expected explicit block position 2, got %v", service.addBlockIn.Position)
	}
	if renderer.createdBlockID != blockIDValue {
		t.Fatalf("expected renderer to receive created block id")
	}
}

func TestLessonBlockAddMapsVideoFlagsToDTO(t *testing.T) {
	service := &lessonServiceFake{addBlockOut: core.AddLessonBlockOutput{ID: blockIDValue}}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service}),
		"block",
		"add",
		"--lesson-id", lessonIDValue,
		"--kind", "video",
		"--video-provider", "youtube",
		"--video-locator", "dQw4w9WgXcQ",
		"--video-caption", "Intro video",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.addBlockIn.LessonID != lessonIDValue || service.addBlockIn.Kind != "video" {
		t.Fatalf("expected video block identifiers to map, got %+v", service.addBlockIn)
	}
	if service.addBlockIn.VideoProvider != "youtube" || service.addBlockIn.VideoLocator != "dQw4w9WgXcQ" || service.addBlockIn.VideoCaption != "Intro video" {
		t.Fatalf("expected video block payload to map, got %+v", service.addBlockIn)
	}
	if service.addBlockIn.Position != nil {
		t.Fatalf("expected nil block position when unset, got %v", service.addBlockIn.Position)
	}
}

func TestLessonBlockAddMapsQuizFlagsToDTO(t *testing.T) {
	service := &lessonServiceFake{addBlockOut: core.AddLessonBlockOutput{ID: blockIDValue}}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service}),
		"block",
		"add",
		"--lesson-id", lessonIDValue,
		"--kind", "quiz",
		"--quiz-id", quizIDValue,
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.addBlockIn.LessonID != lessonIDValue || service.addBlockIn.Kind != "quiz" {
		t.Fatalf("expected quiz block identifiers to map, got %+v", service.addBlockIn)
	}
	if service.addBlockIn.QuizRef != quizIDValue {
		t.Fatalf("expected quiz ref %q, got %+v", quizIDValue, service.addBlockIn)
	}
	if service.addBlockIn.Position != nil {
		t.Fatalf("expected nil block position when unset, got %v", service.addBlockIn.Position)
	}
}

func TestLessonBlockAddMapsPracticeFlagsToDTO(t *testing.T) {
	service := &lessonServiceFake{addBlockOut: core.AddLessonBlockOutput{ID: blockIDValue}}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service}),
		"block",
		"add",
		"--lesson-id", lessonIDValue,
		"--kind", "practice",
		"--practice-id", practiceIDValue,
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.addBlockIn.LessonID != lessonIDValue || service.addBlockIn.Kind != "practice" {
		t.Fatalf("expected practice block identifiers to map, got %+v", service.addBlockIn)
	}
	if service.addBlockIn.PracticeRef != practiceIDValue {
		t.Fatalf("expected practice ref %q, got %+v", practiceIDValue, service.addBlockIn)
	}
	if service.addBlockIn.Position != nil {
		t.Fatalf("expected nil block position when unset, got %v", service.addBlockIn.Position)
	}
}

func TestLessonBlockAddRequiresKindPayloadFlags(t *testing.T) {
	service := &lessonServiceFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service}),
		"block",
		"add",
		"--lesson-id", lessonIDValue,
		"--kind", "text",
	)
	if !errors.Is(err, ErrRequiredFlagMissing) {
		t.Fatalf("expected required text flag error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

func TestLessonBlockListMapsLessonIDAndOutputFormat(t *testing.T) {
	block := blockViewFixture()
	service := &lessonServiceFake{listBlocksOut: core.ListLessonBlocksOutput{Blocks: []core.BlockView{block}}}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"block",
		"list",
		"--lesson-id", lessonIDValue,
		"--output", "json",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "list-blocks" || service.callCount != 1 || service.listBlocksIn.LessonID != lessonIDValue {
		t.Fatalf("expected one list blocks call, got called=%q count=%d input=%+v", service.called, service.callCount, service.listBlocksIn)
	}
	if renderer.blockListFormat != "json" || len(renderer.blocks) != 1 || renderer.blocks[0] != block {
		t.Fatalf("expected renderer to receive json block list")
	}
}

func TestLessonBlockGetMapsIDAndOutputFormat(t *testing.T) {
	block := blockViewFixture()
	service := &lessonServiceFake{getBlockOut: core.GetLessonBlockOutput{Block: block}}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"block",
		"get",
		blockIDValue,
		"--output", "quiet",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "get-block" || service.callCount != 1 || service.getBlockIn.ID != blockIDValue {
		t.Fatalf("expected one get block call, got called=%q count=%d input=%+v", service.called, service.callCount, service.getBlockIn)
	}
	if renderer.blockFormat != "quiet" || renderer.block != block {
		t.Fatalf("expected renderer to receive quiet block view")
	}
}

func TestLessonBlockUpdateMapsOnlyChangedFlags(t *testing.T) {
	service := &lessonServiceFake{updateBlockOut: core.UpdateLessonBlockOutput{ID: blockIDValue}}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"block",
		"update",
		blockIDValue,
		"--text", "Updated markdown",
		"--video-caption", "",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "update-block" || service.callCount != 1 || service.updateBlockIn.ID != blockIDValue {
		t.Fatalf("expected one update block call, got called=%q count=%d input=%+v", service.called, service.callCount, service.updateBlockIn)
	}
	if service.updateBlockIn.Markdown == nil || *service.updateBlockIn.Markdown != "Updated markdown" {
		t.Fatalf("expected text pointer to be set")
	}
	if service.updateBlockIn.VideoCaption == nil || *service.updateBlockIn.VideoCaption != "" {
		t.Fatalf("expected empty video caption pointer to be set")
	}
	if service.updateBlockIn.VideoProvider != nil || service.updateBlockIn.VideoLocator != nil {
		t.Fatalf("expected unchanged video fields to stay nil")
	}
	if renderer.updatedBlockID != blockIDValue {
		t.Fatalf("expected renderer to receive updated block id")
	}
}

func TestLessonBlockRemoveRequiresConfirmation(t *testing.T) {
	service := &lessonServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Prompter: prompter}),
		"block",
		"remove",
		blockIDValue,
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

func TestLessonBlockRemoveForceSkipsConfirmation(t *testing.T) {
	service := &lessonServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{
			Service:  service,
			Renderer: renderer,
			Prompter: prompter,
		}),
		"block",
		"remove",
		blockIDValue,
		"--force",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "remove-block" || service.callCount != 1 || service.removeBlockIn.ID != blockIDValue {
		t.Fatalf("expected one remove block call, got called=%q count=%d input=%+v", service.called, service.callCount, service.removeBlockIn)
	}
	if prompter.message != "" {
		t.Fatalf("expected force remove not to prompt")
	}
	if renderer.confirmation != "lesson block removed" {
		t.Fatalf("expected block remove confirmation to render")
	}
}

func TestLessonBlockReorderParsesOrderPairs(t *testing.T) {
	service := &lessonServiceFake{}
	renderer := &lessonRendererFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service, Renderer: renderer}),
		"block",
		"reorder",
		"--lesson-id", lessonIDValue,
		"--order", blockIDValue+":2, "+otherBlockIDValue+":0",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "reorder-blocks" || service.callCount != 1 || service.reorderBlocksIn.LessonID != lessonIDValue {
		t.Fatalf("expected one reorder blocks call, got called=%q count=%d input=%+v", service.called, service.callCount, service.reorderBlocksIn)
	}

	want := []core.BlockPlacementDTO{
		{BlockID: blockIDValue, Position: 2},
		{BlockID: otherBlockIDValue, Position: 0},
	}
	if len(service.reorderBlocksIn.Order) != len(want) {
		t.Fatalf("expected %d positions, got %d", len(want), len(service.reorderBlocksIn.Order))
	}
	for index := range want {
		if service.reorderBlocksIn.Order[index] != want[index] {
			t.Fatalf("expected placement %+v at index %d, got %+v", want[index], index, service.reorderBlocksIn.Order[index])
		}
	}
	if renderer.confirmation != "lesson blocks reordered" {
		t.Fatalf("expected reorder blocks confirmation to render")
	}
}

func TestLessonBlockReorderRejectsInvalidOrder(t *testing.T) {
	service := &lessonServiceFake{}

	err := executeCourseCommand(
		NewLessonCommand(LessonCommandOptions{Service: service}),
		"block",
		"reorder",
		"--lesson-id", lessonIDValue,
		"--order", blockIDValue,
	)
	if !errors.Is(err, ErrInvalidBlockOrder) {
		t.Fatalf("expected invalid block order error, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

type lessonServiceFake struct {
	called    string
	callCount int

	createIn        core.CreateLessonInput
	createOut       core.CreateLessonOutput
	listIn          core.ListLessonsInput
	listOut         core.ListLessonsOutput
	getIn           core.GetLessonInput
	getOut          core.GetLessonOutput
	updateIn        core.UpdateLessonInput
	updateOut       core.UpdateLessonOutput
	deleteIn        core.DeleteLessonInput
	reorderIn       core.ReorderLessonsInput
	addBlockIn      core.AddLessonBlockInput
	addBlockOut     core.AddLessonBlockOutput
	listBlocksIn    core.ListLessonBlocksInput
	listBlocksOut   core.ListLessonBlocksOutput
	getBlockIn      core.GetLessonBlockInput
	getBlockOut     core.GetLessonBlockOutput
	updateBlockIn   core.UpdateLessonBlockInput
	updateBlockOut  core.UpdateLessonBlockOutput
	removeBlockIn   core.RemoveLessonBlockInput
	reorderBlocksIn core.ReorderLessonBlocksInput
}

func (service *lessonServiceFake) CreateLesson(in core.CreateLessonInput) (core.CreateLessonOutput, error) {
	service.called = "create"
	service.callCount++
	service.createIn = in
	return service.createOut, nil
}

func (service *lessonServiceFake) ListLessons(in core.ListLessonsInput) (core.ListLessonsOutput, error) {
	service.called = "list"
	service.callCount++
	service.listIn = in
	return service.listOut, nil
}

func (service *lessonServiceFake) GetLesson(in core.GetLessonInput) (core.GetLessonOutput, error) {
	service.called = "get"
	service.callCount++
	service.getIn = in
	return service.getOut, nil
}

func (service *lessonServiceFake) UpdateLesson(in core.UpdateLessonInput) (core.UpdateLessonOutput, error) {
	service.called = "update"
	service.callCount++
	service.updateIn = in
	return service.updateOut, nil
}

func (service *lessonServiceFake) DeleteLesson(in core.DeleteLessonInput) error {
	service.called = "delete"
	service.callCount++
	service.deleteIn = in
	return nil
}

func (service *lessonServiceFake) ReorderLessons(in core.ReorderLessonsInput) error {
	service.called = "reorder"
	service.callCount++
	service.reorderIn = in
	return nil
}

func (service *lessonServiceFake) AddLessonBlock(in core.AddLessonBlockInput) (core.AddLessonBlockOutput, error) {
	service.called = "add-block"
	service.callCount++
	service.addBlockIn = in
	return service.addBlockOut, nil
}

func (service *lessonServiceFake) ListLessonBlocks(in core.ListLessonBlocksInput) (core.ListLessonBlocksOutput, error) {
	service.called = "list-blocks"
	service.callCount++
	service.listBlocksIn = in
	return service.listBlocksOut, nil
}

func (service *lessonServiceFake) GetLessonBlock(in core.GetLessonBlockInput) (core.GetLessonBlockOutput, error) {
	service.called = "get-block"
	service.callCount++
	service.getBlockIn = in
	return service.getBlockOut, nil
}

func (service *lessonServiceFake) UpdateLessonBlock(in core.UpdateLessonBlockInput) (core.UpdateLessonBlockOutput, error) {
	service.called = "update-block"
	service.callCount++
	service.updateBlockIn = in
	return service.updateBlockOut, nil
}

func (service *lessonServiceFake) RemoveLessonBlock(in core.RemoveLessonBlockInput) error {
	service.called = "remove-block"
	service.callCount++
	service.removeBlockIn = in
	return nil
}

func (service *lessonServiceFake) ReorderLessonBlocks(in core.ReorderLessonBlocksInput) error {
	service.called = "reorder-blocks"
	service.callCount++
	service.reorderBlocksIn = in
	return nil
}

type lessonRendererFake struct {
	createdID       string
	updatedID       string
	listFormat      string
	lessonFormat    string
	lessons         []core.LessonView
	lesson          core.LessonView
	createdBlockID  string
	updatedBlockID  string
	blockListFormat string
	blockFormat     string
	blocks          []core.BlockView
	block           core.BlockView
	confirmation    string
}

func (renderer *lessonRendererFake) RenderCreatedLesson(id string) error {
	renderer.createdID = id
	return nil
}

func (renderer *lessonRendererFake) RenderLessonList(format string, lessons []core.LessonView) error {
	renderer.listFormat = format
	renderer.lessons = lessons
	return nil
}

func (renderer *lessonRendererFake) RenderLesson(format string, lesson core.LessonView) error {
	renderer.lessonFormat = format
	renderer.lesson = lesson
	return nil
}

func (renderer *lessonRendererFake) RenderUpdatedLesson(id string) error {
	renderer.updatedID = id
	return nil
}

func (renderer *lessonRendererFake) RenderCreatedBlock(id string) error {
	renderer.createdBlockID = id
	return nil
}

func (renderer *lessonRendererFake) RenderBlockList(format string, blocks []core.BlockView) error {
	renderer.blockListFormat = format
	renderer.blocks = blocks
	return nil
}

func (renderer *lessonRendererFake) RenderBlock(format string, block core.BlockView) error {
	renderer.blockFormat = format
	renderer.block = block
	return nil
}

func (renderer *lessonRendererFake) RenderUpdatedBlock(id string) error {
	renderer.updatedBlockID = id
	return nil
}

func (renderer *lessonRendererFake) RenderConfirmation(message string) error {
	renderer.confirmation = message
	return nil
}

func lessonViewFixture() core.LessonView {
	now := time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)
	return core.LessonView{
		ID:        lessonIDValue,
		CourseID:  courseIDValue,
		Title:     "First Lesson",
		Order:     0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func blockViewFixture() core.BlockView {
	return core.BlockView{
		ID:            blockIDValue,
		LessonID:      lessonIDValue,
		Kind:          "video",
		Position:      1,
		VideoProvider: "youtube",
		VideoLocator:  "dQw4w9WgXcQ",
		VideoCaption:  "Intro video",
	}
}
