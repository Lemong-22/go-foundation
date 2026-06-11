package domain

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"
)

const (
	importZipHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	targetIDValue = "550e8400-e29b-41d4-a716-446655440010"
)

func TestOperationKind(t *testing.T) {
	create, err := NewOperationKind("create")
	if err != nil {
		t.Fatalf("expected create operation, got error %v", err)
	}
	if !create.IsCreate() || create.IsUpdate() || create.IsNoop() || create.IsSkip() || create.String() != "create" {
		t.Fatalf("expected create operation kind, got %q", create.String())
	}

	update := UpdateOperation()
	if !update.IsUpdate() || update.IsCreate() || update.String() != "update" {
		t.Fatalf("expected update operation kind, got %q", update.String())
	}

	if _, err := NewOperationKind("delete"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected unsupported operation validation error, got %v", err)
	}
}

func TestEntityType(t *testing.T) {
	tests := []struct {
		value string
		want  func(EntityType) bool
	}{
		{value: "course", want: EntityType.IsCourse},
		{value: "lesson", want: EntityType.IsLesson},
		{value: "block", want: EntityType.IsBlock},
		{value: "quiz", want: EntityType.IsQuiz},
		{value: "question", want: EntityType.IsQuestion},
		{value: "practice", want: EntityType.IsPractice},
		{value: "test_case", want: EntityType.IsTestCase},
		{value: "test", want: EntityType.IsTest},
		{value: "test_item", want: EntityType.IsTestItem},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			entityType, err := NewEntityType(test.value)
			if err != nil {
				t.Fatalf("expected entity type, got error %v", err)
			}
			if !test.want(entityType) || entityType.String() != test.value {
				t.Fatalf("expected entity type %q, got %q", test.value, entityType.String())
			}
		})
	}

	if _, err := NewEntityType("module"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected unsupported entity type validation error, got %v", err)
	}
}

func TestConflictReasonAndStrategy(t *testing.T) {
	reason, err := NewConflictReason("slug_collision")
	if err != nil {
		t.Fatalf("expected conflict reason, got error %v", err)
	}
	if !reason.IsSlugCollision() || reason.IsPositionCollision() || reason.String() != "slug_collision" {
		t.Fatalf("expected slug collision, got %q", reason.String())
	}

	titleReason := TitleInParentCollision()
	if !titleReason.IsTitleInParentCollision() || titleReason.String() != "title_in_parent_collision" {
		t.Fatalf("expected title-in-parent collision, got %q", titleReason.String())
	}

	strategy, err := NewConflictStrategy("fail")
	if err != nil {
		t.Fatalf("expected conflict strategy, got error %v", err)
	}
	if !strategy.IsFail() || strategy.IsSkip() || strategy.IsUpdate() || strategy.String() != "fail" {
		t.Fatalf("expected fail strategy, got %q", strategy.String())
	}

	if _, err := NewConflictReason("ambiguous"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected unsupported reason validation error, got %v", err)
	}
	if _, err := NewConflictStrategy("merge"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected unsupported strategy validation error, got %v", err)
	}
}

func TestImportOperationValidatesTargetIDInvariants(t *testing.T) {
	if _, err := NewImportOperation(CreateOperation(), CourseEntity(), "course:intro", stringPtr(targetIDValue), json.RawMessage(`{"slug":"intro"}`)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected create operation target id validation error, got %v", err)
	}

	if _, err := NewImportOperation(UpdateOperation(), CourseEntity(), "course:intro", nil, json.RawMessage(`{"slug":"intro"}`)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected update operation target id validation error, got %v", err)
	}

	if _, err := NewImportOperation(UpdateOperation(), CourseEntity(), "course:intro", stringPtr("   "), json.RawMessage(`{"slug":"intro"}`)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected empty target id validation error, got %v", err)
	}
}

func TestImportOperationCopiesPayloadAndTargetID(t *testing.T) {
	targetID := targetIDValue
	payload := json.RawMessage(`{"title":"Intro"}`)

	operation, err := NewImportOperation(UpdateOperation(), CourseEntity(), " course:intro ", &targetID, payload)
	if err != nil {
		t.Fatalf("expected import operation, got error %v", err)
	}

	targetID = "changed"
	payload[2] = 'x'

	if operation.EntityRef() != "course:intro" {
		t.Fatalf("expected entity ref to be trimmed, got %q", operation.EntityRef())
	}
	if got := operation.TargetID(); got == nil || *got != targetIDValue {
		t.Fatalf("expected copied target id, got %v", got)
	}
	if string(operation.Payload()) != `{"title":"Intro"}` {
		t.Fatalf("expected copied payload, got %s", operation.Payload())
	}

	gotTargetID := operation.TargetID()
	*gotTargetID = "mutated"
	gotPayload := operation.Payload()
	gotPayload[2] = 'z'

	if *operation.TargetID() != targetIDValue {
		t.Fatalf("expected target id accessor to return a copy")
	}
	if string(operation.Payload()) != `{"title":"Intro"}` {
		t.Fatalf("expected payload accessor to return a copy")
	}
}

func TestImportConflictValidatesCandidatesAndRecommendedKind(t *testing.T) {
	candidate := mustConflictCandidate(t, targetIDValue, "course Intro")

	if _, err := NewImportConflict(CourseEntity(), "course:intro", SlugCollision(), nil, UpdateOperation(), json.RawMessage(`{"slug":"intro"}`)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected missing candidate validation error, got %v", err)
	}
	if _, err := NewImportConflict(CourseEntity(), "course:intro", SlugCollision(), []ConflictCandidate{candidate}, NoopOperation(), json.RawMessage(`{"slug":"intro"}`)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected noop recommendation validation error, got %v", err)
	}
}

func TestImportPlanJSONRoundTrips(t *testing.T) {
	generatedAt := time.Date(2026, 5, 28, 16, 0, 0, 0, time.UTC)
	createOperation := mustImportOperation(t, CreateOperation(), CourseEntity(), "course:intro", nil, `{"slug":"intro"}`)
	updateOperation := mustImportOperation(t, UpdateOperation(), LessonEntity(), "lesson:Intro", stringPtr(targetIDValue), `{"title":"Intro"}`)
	conflict := mustImportConflict(t)

	plan, err := NewImportPlan("1", importZipHash, generatedAt, []ImportOperation{createOperation, updateOperation}, []ImportConflict{conflict})
	if err != nil {
		t.Fatalf("expected import plan, got error %v", err)
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("expected plan to marshal, got error %v", err)
	}

	var decoded ImportPlan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("expected plan to unmarshal, got error %v", err)
	}

	if decoded.FormatVersion() != "1" || decoded.ZipHash() != importZipHash || !decoded.GeneratedAt().Equal(generatedAt) {
		t.Fatalf("decoded plan metadata mismatch")
	}
	if decoded.IsResolved() {
		t.Fatalf("expected decoded plan to retain conflicts")
	}

	operations := decoded.Operations()
	if len(operations) != 2 {
		t.Fatalf("expected two operations, got %d", len(operations))
	}
	if !operations[0].Kind().IsCreate() || operations[0].TargetID() != nil {
		t.Fatalf("expected create operation without target id")
	}
	if got := operations[1].TargetID(); got == nil || *got != targetIDValue {
		t.Fatalf("expected update operation target id, got %v", got)
	}

	conflicts := decoded.Conflicts()
	if len(conflicts) != 1 || !conflicts[0].Reason().IsSlugCollision() || !conflicts[0].Recommended().IsUpdate() {
		t.Fatalf("decoded conflict mismatch")
	}
	if got := conflicts[0].Candidates(); len(got) != 1 || got[0].ID() != targetIDValue {
		t.Fatalf("decoded conflict candidates mismatch: %+v", got)
	}
}

func TestImportPlanCopiesSlices(t *testing.T) {
	operation := mustImportOperation(t, CreateOperation(), CourseEntity(), "course:intro", nil, `{"slug":"intro"}`)
	plan, err := NewImportPlan(
		"1",
		importZipHash,
		time.Date(2026, 5, 28, 16, 0, 0, 0, time.UTC),
		[]ImportOperation{operation},
		nil,
	)
	if err != nil {
		t.Fatalf("expected import plan, got error %v", err)
	}

	gotOperations := plan.Operations()
	gotOperations[0] = mustImportOperation(t, CreateOperation(), CourseEntity(), "course:other", nil, `{"slug":"other"}`)

	if plan.Operations()[0].EntityRef() != "course:intro" {
		t.Fatalf("expected operations accessor to return a copy")
	}
}

func TestImportPlanRejectsInvalidJSONPayload(t *testing.T) {
	if _, err := NewImportOperation(CreateOperation(), CourseEntity(), "course:intro", nil, json.RawMessage(`{"slug"`)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid JSON payload validation error, got %v", err)
	}
}

func TestApplyResultCopiesOperationsAndRequiresFailedOperationError(t *testing.T) {
	operation := mustImportOperation(t, SkipOperation(), CourseEntity(), "course:intro", stringPtr(targetIDValue), `{"slug":"intro"}`)
	applied, err := NewAppliedOperation(operation, "created")
	if err != nil {
		t.Fatalf("expected applied operation, got error %v", err)
	}
	failed, err := NewFailedOperation(operation, ErrNotFound)
	if err != nil {
		t.Fatalf("expected failed operation, got error %v", err)
	}

	result, err := NewApplyResult([]AppliedOperation{applied}, []FailedOperation{failed}, []ImportOperation{operation}, 1, 1)
	if err != nil {
		t.Fatalf("expected apply result, got error %v", err)
	}

	if result.AggregatesSucceeded() != 1 || result.AggregatesFailed() != 1 {
		t.Fatalf("expected aggregate counts to be retained")
	}
	if !errors.Is(result.Failed()[0].Err(), ErrNotFound) {
		t.Fatalf("expected failed operation error to be retained")
	}

	mutated := result.Skipped()
	mutated[0] = mustImportOperation(t, SkipOperation(), CourseEntity(), "course:other", stringPtr(targetIDValue), `{"slug":"other"}`)
	if reflect.DeepEqual(mutated, result.Skipped()) {
		t.Fatalf("expected skipped operations accessor to return a copy")
	}

	if _, err := NewFailedOperation(operation, nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected nil failed operation error validation error, got %v", err)
	}
}

func stringPtr(value string) *string {
	return &value
}

func mustImportOperation(
	t *testing.T,
	kind OperationKind,
	entityType EntityType,
	entityRef string,
	targetID *string,
	payload string,
) ImportOperation {
	t.Helper()

	operation, err := NewImportOperation(kind, entityType, entityRef, targetID, json.RawMessage(payload))
	if err != nil {
		t.Fatalf("expected import operation, got error %v", err)
	}

	return operation
}

func mustImportConflict(t *testing.T) ImportConflict {
	t.Helper()

	conflict, err := NewImportConflict(
		CourseEntity(),
		"course:intro",
		SlugCollision(),
		[]ConflictCandidate{mustConflictCandidate(t, targetIDValue, "course Intro")},
		UpdateOperation(),
		json.RawMessage(`{"slug":"intro"}`),
	)
	if err != nil {
		t.Fatalf("expected import conflict, got error %v", err)
	}

	return conflict
}

func mustConflictCandidate(t *testing.T, id string, description string) ConflictCandidate {
	t.Helper()

	candidate, err := NewConflictCandidate(id, description)
	if err != nil {
		t.Fatalf("expected conflict candidate, got error %v", err)
	}

	return candidate
}
