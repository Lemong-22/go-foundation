package cli

import (
	"testing"

	"github.com/spf13/viper"
)

func TestDeferredCourseArchiveCommandRemainsAbsent(t *testing.T) {
	command := NewCourseCommand(CourseCommandOptions{Service: &courseServiceFake{}, Config: viper.New()})

	if _, _, err := command.Find([]string{"archive"}); err == nil {
		t.Fatalf("expected deferred course archive command to remain absent")
	}
}

func TestDeferredLessonPublishCommandsRemainAbsent(t *testing.T) {
	command := NewLessonCommand(LessonCommandOptions{Service: &lessonServiceFake{}})

	deferredCommands := []string{"publish", "unpublish"}
	for _, name := range deferredCommands {
		if _, _, err := command.Find([]string{name}); err == nil {
			t.Fatalf("expected deferred lesson %s command to remain absent", name)
		}
	}
}
