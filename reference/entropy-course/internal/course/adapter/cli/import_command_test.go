package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const importZipHash = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"

func TestImportCommandExposesSubcommands(t *testing.T) {
	command := NewImportCommand(ImportCommandOptions{Service: &importServiceFake{}, Config: viper.New()})

	for _, name := range []string{"plan", "apply"} {
		if _, _, err := command.Find([]string{name}); err != nil {
			t.Fatalf("expected import %s command to exist, got %v", name, err)
		}
	}
}

func TestImportPlanMapsZipInstructorAndFormatToDTO(t *testing.T) {
	plan := importPlanFixture(t, nil, nil)
	service := &importServiceFake{planOut: core.PlanImportOutput{Plan: plan}}
	renderer := &importRendererFake{}
	config := viper.New()
	config.Set("instructor-id", instructorIDValue)

	err := executeImportCommand(
		NewImportCommand(ImportCommandOptions{Service: service, Renderer: renderer, Config: config}),
		"plan",
		"course.zip",
		"-o", "table",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "plan" {
		t.Fatalf("expected plan to be called, got %q", service.called)
	}
	if service.planIn.ZipPath != "course.zip" || service.planIn.InstructorID != instructorIDValue {
		t.Fatalf("expected primitive plan input, got %+v", service.planIn)
	}
	if renderer.planFormat != "table" || renderer.plan.ZipHash() != importZipHash {
		t.Fatalf("expected plan renderer to receive table plan")
	}
}

func TestImportPlanWritesJSONToOutputFile(t *testing.T) {
	plan := importPlanFixture(t, nil, nil)
	service := &importServiceFake{planOut: core.PlanImportOutput{Plan: plan}}
	outputPath := t.TempDir() + "/plan.json"

	err := executeImportCommand(
		NewImportCommand(ImportCommandOptions{Service: service, Config: viper.New()}),
		"plan",
		"course.zip",
		"--instructor-id", instructorIDValue,
		"--output", outputPath,
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("expected output file to be written, got %v", err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("expected plan json, got %v", err)
	}
	if payload["zip_hash"] != importZipHash {
		t.Fatalf("expected plan zip hash, got %q", data)
	}
}

func TestImportApplyMapsResolvedPlanStrategyForceAndInstructor(t *testing.T) {
	service := &importServiceFake{applyOut: core.ApplyPlanOutput{Result: emptyApplyResult(t)}}
	resolvedPlan := []byte(`{"format_version":"1"}`)
	resolvedPlanPath := t.TempDir() + "/resolved-plan.json"
	if err := os.WriteFile(resolvedPlanPath, resolvedPlan, 0o600); err != nil {
		t.Fatalf("expected resolved plan fixture, got %v", err)
	}

	err := executeImportCommand(
		NewImportCommand(ImportCommandOptions{Service: service, Config: viper.New()}),
		"apply",
		"course.zip",
		"--instructor-id", instructorIDValue,
		"--resolved-plan", resolvedPlanPath,
		"--conflict-strategy", "update",
		"--force",
	)
	if err != nil {
		t.Fatalf("expected command to succeed, got %v", err)
	}

	if service.called != "apply" {
		t.Fatalf("expected apply to be called, got %q", service.called)
	}
	if service.applyIn.ZipPath != "course.zip" || service.applyIn.InstructorID != instructorIDValue || service.applyIn.ConflictStrategy != "update" {
		t.Fatalf("expected primitive apply input, got %+v", service.applyIn)
	}
	if !bytes.Equal(service.applyIn.ResolvedPlanJSON, resolvedPlan) {
		t.Fatalf("expected resolved plan bytes to be passed through")
	}
}

func TestImportApplyForceBypassesConfirmationAndRendersDefaultTable(t *testing.T) {
	service := &importServiceFake{applyOut: core.ApplyPlanOutput{Result: emptyApplyResult(t)}}
	renderer := &importRendererFake{}
	prompter := &coursePrompterFake{confirmed: false}

	err := executeImportCommand(
		NewImportCommand(ImportCommandOptions{Service: service, Renderer: renderer, Prompter: prompter, Config: viper.New()}),
		"apply",
		"course.zip",
		"--instructor-id", instructorIDValue,
		"--force",
	)
	if err != nil {
		t.Fatalf("expected force apply to succeed, got %v", err)
	}

	if prompter.message != "" {
		t.Fatalf("expected --force to bypass confirmation, got prompt %q", prompter.message)
	}
	if service.called != "apply" {
		t.Fatalf("expected apply service call, got %q", service.called)
	}
	if renderer.resultFormat != "table" {
		t.Fatalf("expected default apply table output, got %q", renderer.resultFormat)
	}
}

func TestImportApplyRequiresConfirmationWithoutForce(t *testing.T) {
	service := &importServiceFake{}
	prompter := &coursePrompterFake{confirmed: false}

	err := executeImportCommand(
		NewImportCommand(ImportCommandOptions{Service: service, Prompter: prompter, Config: viper.New()}),
		"apply",
		"course.zip",
		"--instructor-id", instructorIDValue,
	)
	if !errors.Is(err, ErrConfirmationDeclined) {
		t.Fatalf("expected confirmation declined, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called, got %q", service.called)
	}
}

func TestImportApplyReportsUnresolvedConflictRefs(t *testing.T) {
	conflict := importConflictFixture(t, "course:intro")
	service := &importServiceFake{err: domain.NewUnresolvedImportConflictsError([]domain.ImportConflict{conflict})}

	err := executeImportCommand(
		NewImportCommand(ImportCommandOptions{Service: service, Config: viper.New()}),
		"apply",
		"course.zip",
		"--instructor-id", instructorIDValue,
		"--force",
	)
	if !errors.Is(err, domain.ErrUnresolvedImportConflicts) {
		t.Fatalf("expected unresolved conflicts error, got %v", err)
	}
	if !strings.Contains(err.Error(), "course:intro") {
		t.Fatalf("expected conflict ref in error, got %v", err)
	}
}

func TestImportPlanRequiresInstructorID(t *testing.T) {
	service := &importServiceFake{}

	err := executeImportCommand(
		NewImportCommand(ImportCommandOptions{Service: service, Config: viper.New()}),
		"plan",
		"course.zip",
	)
	if !errors.Is(err, ErrInstructorIDRequired) {
		t.Fatalf("expected instructor id required, got %v", err)
	}
	if service.called != "" {
		t.Fatalf("expected service not to be called")
	}
}

func executeImportCommand(command *cobra.Command, args ...string) error {
	command.SetArgs(args)
	command.SetOut(&bytes.Buffer{})
	command.SetErr(&bytes.Buffer{})
	return command.Execute()
}

type importServiceFake struct {
	called string
	err    error

	planIn  core.PlanImportInput
	planOut core.PlanImportOutput

	applyIn  core.ApplyPlanInput
	applyOut core.ApplyPlanOutput
}

func (service *importServiceFake) PlanImport(in core.PlanImportInput) (core.PlanImportOutput, error) {
	service.called = "plan"
	service.planIn = in
	if service.err != nil {
		return core.PlanImportOutput{}, service.err
	}

	return service.planOut, nil
}

func (service *importServiceFake) ApplyPlan(in core.ApplyPlanInput) (core.ApplyPlanOutput, error) {
	service.called = "apply"
	service.applyIn = in
	if service.err != nil {
		return core.ApplyPlanOutput{}, service.err
	}

	return service.applyOut, nil
}

type importRendererFake struct {
	planFormat string
	plan       domain.ImportPlan

	resultFormat string
	result       domain.ApplyResult
}

func (renderer *importRendererFake) RenderImportPlan(format string, plan domain.ImportPlan) error {
	renderer.planFormat = format
	renderer.plan = plan
	return nil
}

func (renderer *importRendererFake) RenderApplyResult(format string, result domain.ApplyResult) error {
	renderer.resultFormat = format
	renderer.result = result
	return nil
}

func importPlanFixture(t *testing.T, operations []domain.ImportOperation, conflicts []domain.ImportConflict) domain.ImportPlan {
	t.Helper()

	plan, err := domain.NewImportPlan("1", importZipHash, time.Date(2026, 5, 28, 18, 0, 0, 0, time.UTC), operations, conflicts)
	if err != nil {
		t.Fatalf("expected import plan fixture, got %v", err)
	}

	return plan
}

func importOperationFixture(t *testing.T, kind domain.OperationKind, entity domain.EntityType, ref string, targetID *string) domain.ImportOperation {
	t.Helper()

	operation, err := domain.NewImportOperation(kind, entity, ref, targetID, json.RawMessage(`{"title":"Intro"}`))
	if err != nil {
		t.Fatalf("expected import operation fixture, got %v", err)
	}

	return operation
}

func importConflictFixture(t *testing.T, ref string) domain.ImportConflict {
	t.Helper()

	candidate, err := domain.NewConflictCandidate(courseIDValue, "course Intro")
	if err != nil {
		t.Fatalf("expected candidate fixture, got %v", err)
	}
	conflict, err := domain.NewImportConflict(
		domain.CourseEntity(),
		ref,
		domain.SlugCollision(),
		[]domain.ConflictCandidate{candidate},
		domain.UpdateOperation(),
		json.RawMessage(`{"title":"Intro"}`),
	)
	if err != nil {
		t.Fatalf("expected import conflict fixture, got %v", err)
	}

	return conflict
}

func emptyApplyResult(t *testing.T) domain.ApplyResult {
	t.Helper()

	result, err := domain.NewApplyResult(nil, nil, nil, 0, 0)
	if err != nil {
		t.Fatalf("expected apply result fixture, got %v", err)
	}

	return result
}
