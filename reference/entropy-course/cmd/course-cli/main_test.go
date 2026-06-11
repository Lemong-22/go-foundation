package main

import (
	"context"
	"testing"

	"github.com/luxeave/entropy-course/internal/course/core"
)

func TestRootCommandMountsPhaseDAdapters(t *testing.T) {
	command := newRootCommand(context.Background())

	paths := [][]string{
		{"lesson", "block", "add"},
		{"lesson", "block", "list"},
		{"lesson", "block", "get"},
		{"lesson", "block", "update"},
		{"lesson", "block", "remove"},
		{"lesson", "block", "reorder"},
		{"quiz", "create"},
		{"quiz", "list"},
		{"quiz", "get"},
		{"quiz", "update"},
		{"quiz", "delete"},
		{"quiz", "question", "add"},
		{"quiz", "question", "list"},
		{"quiz", "question", "get"},
		{"quiz", "question", "update"},
		{"quiz", "question", "remove"},
		{"quiz", "question", "reorder"},
		{"practice", "create"},
		{"practice", "list"},
		{"practice", "get"},
		{"practice", "update"},
		{"practice", "delete"},
		{"practice", "testcase", "add"},
		{"practice", "testcase", "list"},
		{"practice", "testcase", "get"},
		{"practice", "testcase", "update"},
		{"practice", "testcase", "remove"},
		{"practice", "testcase", "reorder"},
		{"test", "create"},
		{"test", "list"},
		{"test", "get"},
		{"test", "update"},
		{"test", "delete"},
		{"test", "item", "add"},
		{"test", "item", "list"},
		{"test", "item", "get"},
		{"test", "item", "update"},
		{"test", "item", "remove"},
		{"test", "item", "reorder"},
		{"import", "plan"},
		{"import", "apply"},
		{"migrate", "up"},
		{"migrate", "down"},
		{"migrate", "status"},
		{"rest"},
		{"playground"},
	}

	for _, path := range paths {
		if _, _, err := command.Find(path); err != nil {
			t.Fatalf("expected command path %v to be mounted, got %v", path, err)
		}
	}
}

func TestRootCommandBindsPhaseDRuntimeFlags(t *testing.T) {
	command := newRootCommand(context.Background())

	if command.PersistentFlags().Lookup("api-token") == nil {
		t.Fatalf("expected api-token persistent flag")
	}
	if command.PersistentFlags().Lookup("db-url") == nil {
		t.Fatalf("expected db-url persistent flag")
	}

	restCommand, _, err := command.Find([]string{"rest"})
	if err != nil {
		t.Fatalf("expected rest command, got %v", err)
	}
	if restCommand.Flags().Lookup("addr") == nil {
		t.Fatalf("expected rest addr flag")
	}

	playgroundCommand, _, err := command.Find([]string{"playground"})
	if err != nil {
		t.Fatalf("expected playground command, got %v", err)
	}
	if playgroundCommand.Flags().Lookup("addr") == nil {
		t.Fatalf("expected playground addr flag")
	}
}

func TestDeferredPracticeServiceImplementsPort(t *testing.T) {
	var _ core.PracticeService = deferredPracticeService{}
}

func TestDeferredTestServiceImplementsPort(t *testing.T) {
	var _ core.TestService = deferredTestService{}
}

func TestDeferredImportServiceImplementsPort(t *testing.T) {
	var _ core.ImportService = deferredImportService{}
}
