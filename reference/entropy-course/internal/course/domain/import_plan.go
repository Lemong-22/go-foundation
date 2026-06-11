package domain

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

const (
	operationKindCreate = "create"
	operationKindUpdate = "update"
	operationKindNoop   = "noop"
	operationKindSkip   = "skip"

	entityTypeCourse   = "course"
	entityTypeLesson   = "lesson"
	entityTypeBlock    = "block"
	entityTypeQuiz     = "quiz"
	entityTypeQuestion = "question"
	entityTypePractice = "practice"
	entityTypeTestCase = "test_case"
	entityTypeTest     = "test"
	entityTypeTestItem = "test_item"

	conflictReasonSlugCollision          = "slug_collision"
	conflictReasonTitleInParentCollision = "title_in_parent_collision"
	conflictReasonPositionCollision      = "position_collision"

	conflictStrategyFail   = "fail"
	conflictStrategySkip   = "skip"
	conflictStrategyUpdate = "update"
)

var importZipHashPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

type OperationKind struct {
	value string
}

func CreateOperation() OperationKind {
	return OperationKind{value: operationKindCreate}
}

func UpdateOperation() OperationKind {
	return OperationKind{value: operationKindUpdate}
}

func NoopOperation() OperationKind {
	return OperationKind{value: operationKindNoop}
}

func SkipOperation() OperationKind {
	return OperationKind{value: operationKindSkip}
}

func NewOperationKind(value string) (OperationKind, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case operationKindCreate:
		return CreateOperation(), nil
	case operationKindUpdate:
		return UpdateOperation(), nil
	case operationKindNoop:
		return NoopOperation(), nil
	case operationKindSkip:
		return SkipOperation(), nil
	default:
		return OperationKind{}, NewValidationError("operation_kind", "must be create, update, noop, or skip")
	}
}

func (kind OperationKind) String() string {
	return kind.value
}

func (kind OperationKind) IsCreate() bool {
	return kind.value == operationKindCreate
}

func (kind OperationKind) IsUpdate() bool {
	return kind.value == operationKindUpdate
}

func (kind OperationKind) IsNoop() bool {
	return kind.value == operationKindNoop
}

func (kind OperationKind) IsSkip() bool {
	return kind.value == operationKindSkip
}

func (kind OperationKind) MarshalJSON() ([]byte, error) {
	return json.Marshal(kind.String())
}

func (kind *OperationKind) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := NewOperationKind(value)
	if err != nil {
		return err
	}

	*kind = parsed
	return nil
}

type EntityType struct {
	value string
}

func CourseEntity() EntityType {
	return EntityType{value: entityTypeCourse}
}

func LessonEntity() EntityType {
	return EntityType{value: entityTypeLesson}
}

func BlockEntity() EntityType {
	return EntityType{value: entityTypeBlock}
}

func QuizEntity() EntityType {
	return EntityType{value: entityTypeQuiz}
}

func QuestionEntity() EntityType {
	return EntityType{value: entityTypeQuestion}
}

func PracticeEntity() EntityType {
	return EntityType{value: entityTypePractice}
}

func TestCaseEntity() EntityType {
	return EntityType{value: entityTypeTestCase}
}

func TestEntity() EntityType {
	return EntityType{value: entityTypeTest}
}

func TestItemEntity() EntityType {
	return EntityType{value: entityTypeTestItem}
}

func NewEntityType(value string) (EntityType, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case entityTypeCourse:
		return CourseEntity(), nil
	case entityTypeLesson:
		return LessonEntity(), nil
	case entityTypeBlock:
		return BlockEntity(), nil
	case entityTypeQuiz:
		return QuizEntity(), nil
	case entityTypeQuestion:
		return QuestionEntity(), nil
	case entityTypePractice:
		return PracticeEntity(), nil
	case entityTypeTestCase:
		return TestCaseEntity(), nil
	case entityTypeTest:
		return TestEntity(), nil
	case entityTypeTestItem:
		return TestItemEntity(), nil
	default:
		return EntityType{}, NewValidationError("entity_type", "must be course, lesson, block, quiz, question, practice, test_case, test, or test_item")
	}
}

func (entityType EntityType) String() string {
	return entityType.value
}

func (entityType EntityType) IsCourse() bool {
	return entityType.value == entityTypeCourse
}

func (entityType EntityType) IsLesson() bool {
	return entityType.value == entityTypeLesson
}

func (entityType EntityType) IsBlock() bool {
	return entityType.value == entityTypeBlock
}

func (entityType EntityType) IsQuiz() bool {
	return entityType.value == entityTypeQuiz
}

func (entityType EntityType) IsQuestion() bool {
	return entityType.value == entityTypeQuestion
}

func (entityType EntityType) IsPractice() bool {
	return entityType.value == entityTypePractice
}

func (entityType EntityType) IsTestCase() bool {
	return entityType.value == entityTypeTestCase
}

func (entityType EntityType) IsTest() bool {
	return entityType.value == entityTypeTest
}

func (entityType EntityType) IsTestItem() bool {
	return entityType.value == entityTypeTestItem
}

func (entityType EntityType) MarshalJSON() ([]byte, error) {
	return json.Marshal(entityType.String())
}

func (entityType *EntityType) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := NewEntityType(value)
	if err != nil {
		return err
	}

	*entityType = parsed
	return nil
}

type ConflictReason struct {
	value string
}

func SlugCollision() ConflictReason {
	return ConflictReason{value: conflictReasonSlugCollision}
}

func TitleInParentCollision() ConflictReason {
	return ConflictReason{value: conflictReasonTitleInParentCollision}
}

func PositionCollision() ConflictReason {
	return ConflictReason{value: conflictReasonPositionCollision}
}

func NewConflictReason(value string) (ConflictReason, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case conflictReasonSlugCollision:
		return SlugCollision(), nil
	case conflictReasonTitleInParentCollision:
		return TitleInParentCollision(), nil
	case conflictReasonPositionCollision:
		return PositionCollision(), nil
	default:
		return ConflictReason{}, NewValidationError("conflict_reason", "must be slug_collision, title_in_parent_collision, or position_collision")
	}
}

func (reason ConflictReason) String() string {
	return reason.value
}

func (reason ConflictReason) IsSlugCollision() bool {
	return reason.value == conflictReasonSlugCollision
}

func (reason ConflictReason) IsTitleInParentCollision() bool {
	return reason.value == conflictReasonTitleInParentCollision
}

func (reason ConflictReason) IsPositionCollision() bool {
	return reason.value == conflictReasonPositionCollision
}

func (reason ConflictReason) MarshalJSON() ([]byte, error) {
	return json.Marshal(reason.String())
}

func (reason *ConflictReason) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := NewConflictReason(value)
	if err != nil {
		return err
	}

	*reason = parsed
	return nil
}

type ConflictStrategy struct {
	value string
}

func FailConflicts() ConflictStrategy {
	return ConflictStrategy{value: conflictStrategyFail}
}

func SkipConflicts() ConflictStrategy {
	return ConflictStrategy{value: conflictStrategySkip}
}

func UpdateConflicts() ConflictStrategy {
	return ConflictStrategy{value: conflictStrategyUpdate}
}

func NewConflictStrategy(value string) (ConflictStrategy, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case conflictStrategyFail:
		return FailConflicts(), nil
	case conflictStrategySkip:
		return SkipConflicts(), nil
	case conflictStrategyUpdate:
		return UpdateConflicts(), nil
	default:
		return ConflictStrategy{}, NewValidationError("conflict_strategy", "must be fail, skip, or update")
	}
}

func (strategy ConflictStrategy) String() string {
	return strategy.value
}

func (strategy ConflictStrategy) IsFail() bool {
	return strategy.value == conflictStrategyFail
}

func (strategy ConflictStrategy) IsSkip() bool {
	return strategy.value == conflictStrategySkip
}

func (strategy ConflictStrategy) IsUpdate() bool {
	return strategy.value == conflictStrategyUpdate
}

func (strategy ConflictStrategy) MarshalJSON() ([]byte, error) {
	return json.Marshal(strategy.String())
}

func (strategy *ConflictStrategy) UnmarshalJSON(data []byte) error {
	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	parsed, err := NewConflictStrategy(value)
	if err != nil {
		return err
	}

	*strategy = parsed
	return nil
}

type ImportPlan struct {
	formatVersion string
	zipHash       string
	generatedAt   time.Time
	operations    []ImportOperation
	conflicts     []ImportConflict
}

func NewImportPlan(
	formatVersion string,
	zipHash string,
	generatedAt time.Time,
	operations []ImportOperation,
	conflicts []ImportConflict,
) (ImportPlan, error) {
	normalizedVersion := strings.TrimSpace(formatVersion)
	if normalizedVersion == "" {
		return ImportPlan{}, NewValidationError("format_version", "must not be empty")
	}

	normalizedHash := strings.TrimSpace(zipHash)
	if !importZipHashPattern.MatchString(normalizedHash) {
		return ImportPlan{}, NewValidationError("zip_hash", "must be a lowercase SHA-256 hex digest")
	}

	if generatedAt.IsZero() {
		return ImportPlan{}, NewValidationError("generated_at", "must not be zero")
	}

	copiedOperations, err := copyImportOperations(operations)
	if err != nil {
		return ImportPlan{}, err
	}

	copiedConflicts, err := copyImportConflicts(conflicts)
	if err != nil {
		return ImportPlan{}, err
	}

	return ImportPlan{
		formatVersion: normalizedVersion,
		zipHash:       normalizedHash,
		generatedAt:   generatedAt,
		operations:    copiedOperations,
		conflicts:     copiedConflicts,
	}, nil
}

func (plan ImportPlan) FormatVersion() string {
	return plan.formatVersion
}

func (plan ImportPlan) ZipHash() string {
	return plan.zipHash
}

func (plan ImportPlan) GeneratedAt() time.Time {
	return plan.generatedAt
}

func (plan ImportPlan) Operations() []ImportOperation {
	copied, _ := copyImportOperations(plan.operations)
	return copied
}

func (plan ImportPlan) Conflicts() []ImportConflict {
	copied, _ := copyImportConflicts(plan.conflicts)
	return copied
}

func (plan ImportPlan) IsResolved() bool {
	return len(plan.conflicts) == 0
}

func (plan ImportPlan) MarshalJSON() ([]byte, error) {
	return json.Marshal(importPlanJSON{
		FormatVersion: plan.formatVersion,
		ZipHash:       plan.zipHash,
		GeneratedAt:   plan.generatedAt,
		Operations:    plan.operations,
		Conflicts:     plan.conflicts,
	})
}

func (plan *ImportPlan) UnmarshalJSON(data []byte) error {
	var dto importPlanJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}

	parsed, err := NewImportPlan(dto.FormatVersion, dto.ZipHash, dto.GeneratedAt, dto.Operations, dto.Conflicts)
	if err != nil {
		return err
	}

	*plan = parsed
	return nil
}

type ImportOperation struct {
	kind       OperationKind
	entityType EntityType
	entityRef  string
	targetID   *string
	payload    json.RawMessage
}

func NewImportOperation(
	kind OperationKind,
	entityType EntityType,
	entityRef string,
	targetID *string,
	payload json.RawMessage,
) (ImportOperation, error) {
	normalizedKind, err := NewOperationKind(kind.String())
	if err != nil {
		return ImportOperation{}, err
	}

	normalizedEntityType, err := NewEntityType(entityType.String())
	if err != nil {
		return ImportOperation{}, err
	}

	normalizedRef, err := normalizeImportRef(entityRef)
	if err != nil {
		return ImportOperation{}, err
	}

	normalizedTargetID, err := normalizeImportOperationTargetID(normalizedKind, targetID)
	if err != nil {
		return ImportOperation{}, err
	}

	normalizedPayload, err := normalizeImportPayload("payload", payload)
	if err != nil {
		return ImportOperation{}, err
	}

	return ImportOperation{
		kind:       normalizedKind,
		entityType: normalizedEntityType,
		entityRef:  normalizedRef,
		targetID:   normalizedTargetID,
		payload:    normalizedPayload,
	}, nil
}

func (operation ImportOperation) Kind() OperationKind {
	return operation.kind
}

func (operation ImportOperation) EntityType() EntityType {
	return operation.entityType
}

func (operation ImportOperation) EntityRef() string {
	return operation.entityRef
}

func (operation ImportOperation) TargetID() *string {
	if operation.targetID == nil {
		return nil
	}

	copied := *operation.targetID
	return &copied
}

func (operation ImportOperation) Payload() json.RawMessage {
	return copyRawMessage(operation.payload)
}

func (operation ImportOperation) MarshalJSON() ([]byte, error) {
	return json.Marshal(importOperationJSON{
		Kind:       operation.kind,
		EntityType: operation.entityType,
		EntityRef:  operation.entityRef,
		TargetID:   operation.TargetID(),
		Payload:    operation.Payload(),
	})
}

func (operation *ImportOperation) UnmarshalJSON(data []byte) error {
	var dto importOperationJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}

	parsed, err := NewImportOperation(dto.Kind, dto.EntityType, dto.EntityRef, dto.TargetID, dto.Payload)
	if err != nil {
		return err
	}

	*operation = parsed
	return nil
}

func (operation ImportOperation) validate() error {
	_, err := NewImportOperation(operation.kind, operation.entityType, operation.entityRef, operation.targetID, operation.payload)
	return err
}

func (operation ImportOperation) copy() ImportOperation {
	copied, _ := NewImportOperation(operation.kind, operation.entityType, operation.entityRef, operation.targetID, operation.payload)
	return copied
}

type ImportConflict struct {
	entityType  EntityType
	entityRef   string
	reason      ConflictReason
	candidates  []ConflictCandidate
	recommended OperationKind
	payload     json.RawMessage
}

func NewImportConflict(
	entityType EntityType,
	entityRef string,
	reason ConflictReason,
	candidates []ConflictCandidate,
	recommended OperationKind,
	payload json.RawMessage,
) (ImportConflict, error) {
	normalizedEntityType, err := NewEntityType(entityType.String())
	if err != nil {
		return ImportConflict{}, err
	}

	normalizedRef, err := normalizeImportRef(entityRef)
	if err != nil {
		return ImportConflict{}, err
	}

	normalizedReason, err := NewConflictReason(reason.String())
	if err != nil {
		return ImportConflict{}, err
	}

	normalizedRecommended, err := NewOperationKind(recommended.String())
	if err != nil {
		return ImportConflict{}, err
	}

	if normalizedRecommended.IsNoop() {
		return ImportConflict{}, NewValidationError("recommended", "must not be noop for an unresolved conflict")
	}

	copiedCandidates, err := copyConflictCandidates(candidates)
	if err != nil {
		return ImportConflict{}, err
	}
	if len(copiedCandidates) == 0 {
		return ImportConflict{}, NewValidationError("candidates", "must include at least one candidate")
	}

	normalizedPayload, err := normalizeImportPayload("payload", payload)
	if err != nil {
		return ImportConflict{}, err
	}

	return ImportConflict{
		entityType:  normalizedEntityType,
		entityRef:   normalizedRef,
		reason:      normalizedReason,
		candidates:  copiedCandidates,
		recommended: normalizedRecommended,
		payload:     normalizedPayload,
	}, nil
}

func (conflict ImportConflict) EntityType() EntityType {
	return conflict.entityType
}

func (conflict ImportConflict) EntityRef() string {
	return conflict.entityRef
}

func (conflict ImportConflict) Reason() ConflictReason {
	return conflict.reason
}

func (conflict ImportConflict) Candidates() []ConflictCandidate {
	copied, _ := copyConflictCandidates(conflict.candidates)
	return copied
}

func (conflict ImportConflict) Recommended() OperationKind {
	return conflict.recommended
}

func (conflict ImportConflict) Payload() json.RawMessage {
	return copyRawMessage(conflict.payload)
}

func (conflict ImportConflict) MarshalJSON() ([]byte, error) {
	return json.Marshal(importConflictJSON{
		EntityType:  conflict.entityType,
		EntityRef:   conflict.entityRef,
		Reason:      conflict.reason,
		Candidates:  conflict.Candidates(),
		Recommended: conflict.recommended,
		Payload:     conflict.Payload(),
	})
}

func (conflict *ImportConflict) UnmarshalJSON(data []byte) error {
	var dto importConflictJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}

	parsed, err := NewImportConflict(dto.EntityType, dto.EntityRef, dto.Reason, dto.Candidates, dto.Recommended, dto.Payload)
	if err != nil {
		return err
	}

	*conflict = parsed
	return nil
}

func (conflict ImportConflict) validate() error {
	_, err := NewImportConflict(
		conflict.entityType,
		conflict.entityRef,
		conflict.reason,
		conflict.candidates,
		conflict.recommended,
		conflict.payload,
	)
	return err
}

func (conflict ImportConflict) copy() ImportConflict {
	copied, _ := NewImportConflict(
		conflict.entityType,
		conflict.entityRef,
		conflict.reason,
		conflict.candidates,
		conflict.recommended,
		conflict.payload,
	)
	return copied
}

type ConflictCandidate struct {
	id          string
	description string
}

func NewConflictCandidate(id string, description string) (ConflictCandidate, error) {
	normalizedID := strings.TrimSpace(id)
	if normalizedID == "" {
		return ConflictCandidate{}, NewValidationError("candidate_id", "must not be empty")
	}

	normalizedDescription := strings.TrimSpace(description)
	if normalizedDescription == "" {
		return ConflictCandidate{}, NewValidationError("candidate_description", "must not be empty")
	}

	return ConflictCandidate{id: normalizedID, description: normalizedDescription}, nil
}

func (candidate ConflictCandidate) ID() string {
	return candidate.id
}

func (candidate ConflictCandidate) Description() string {
	return candidate.description
}

func (candidate ConflictCandidate) MarshalJSON() ([]byte, error) {
	return json.Marshal(conflictCandidateJSON{
		ID:          candidate.id,
		Description: candidate.description,
	})
}

func (candidate *ConflictCandidate) UnmarshalJSON(data []byte) error {
	var dto conflictCandidateJSON
	if err := json.Unmarshal(data, &dto); err != nil {
		return err
	}

	parsed, err := NewConflictCandidate(dto.ID, dto.Description)
	if err != nil {
		return err
	}

	*candidate = parsed
	return nil
}

func (candidate ConflictCandidate) validate() error {
	_, err := NewConflictCandidate(candidate.id, candidate.description)
	return err
}

type ApplyResult struct {
	applied             []AppliedOperation
	failed              []FailedOperation
	skipped             []ImportOperation
	aggregatesSucceeded int
	aggregatesFailed    int
}

func NewApplyResult(
	applied []AppliedOperation,
	failed []FailedOperation,
	skipped []ImportOperation,
	aggregatesSucceeded int,
	aggregatesFailed int,
) (ApplyResult, error) {
	if aggregatesSucceeded < 0 {
		return ApplyResult{}, NewValidationError("aggregates_succeeded", "must be greater than or equal to zero")
	}
	if aggregatesFailed < 0 {
		return ApplyResult{}, NewValidationError("aggregates_failed", "must be greater than or equal to zero")
	}

	copiedApplied, err := copyAppliedOperations(applied)
	if err != nil {
		return ApplyResult{}, err
	}

	copiedFailed, err := copyFailedOperations(failed)
	if err != nil {
		return ApplyResult{}, err
	}

	copiedSkipped, err := copyImportOperations(skipped)
	if err != nil {
		return ApplyResult{}, err
	}

	return ApplyResult{
		applied:             copiedApplied,
		failed:              copiedFailed,
		skipped:             copiedSkipped,
		aggregatesSucceeded: aggregatesSucceeded,
		aggregatesFailed:    aggregatesFailed,
	}, nil
}

func (result ApplyResult) Applied() []AppliedOperation {
	copied, _ := copyAppliedOperations(result.applied)
	return copied
}

func (result ApplyResult) Failed() []FailedOperation {
	copied, _ := copyFailedOperations(result.failed)
	return copied
}

func (result ApplyResult) Skipped() []ImportOperation {
	copied, _ := copyImportOperations(result.skipped)
	return copied
}

func (result ApplyResult) AggregatesSucceeded() int {
	return result.aggregatesSucceeded
}

func (result ApplyResult) AggregatesFailed() int {
	return result.aggregatesFailed
}

type AppliedOperation struct {
	operation ImportOperation
	message   string
}

func NewAppliedOperation(operation ImportOperation, message string) (AppliedOperation, error) {
	if err := operation.validate(); err != nil {
		return AppliedOperation{}, err
	}

	return AppliedOperation{
		operation: operation.copy(),
		message:   strings.TrimSpace(message),
	}, nil
}

func (operation AppliedOperation) Operation() ImportOperation {
	return operation.operation.copy()
}

func (operation AppliedOperation) Message() string {
	return operation.message
}

func (operation AppliedOperation) validate() error {
	_, err := NewAppliedOperation(operation.operation, operation.message)
	return err
}

func (operation AppliedOperation) copy() AppliedOperation {
	copied, _ := NewAppliedOperation(operation.operation, operation.message)
	return copied
}

type FailedOperation struct {
	operation ImportOperation
	err       error
}

func NewFailedOperation(operation ImportOperation, err error) (FailedOperation, error) {
	if err == nil {
		return FailedOperation{}, NewValidationError("error", "must not be nil")
	}
	if operationErr := operation.validate(); operationErr != nil {
		return FailedOperation{}, operationErr
	}

	return FailedOperation{
		operation: operation.copy(),
		err:       err,
	}, nil
}

func (operation FailedOperation) Operation() ImportOperation {
	return operation.operation.copy()
}

func (operation FailedOperation) Err() error {
	return operation.err
}

func (operation FailedOperation) validate() error {
	if operation.err == nil {
		return NewValidationError("error", "must not be nil")
	}

	return operation.operation.validate()
}

func (operation FailedOperation) copy() FailedOperation {
	copied, _ := NewFailedOperation(operation.operation, operation.err)
	return copied
}

type importPlanJSON struct {
	FormatVersion string            `json:"format_version"`
	ZipHash       string            `json:"zip_hash"`
	GeneratedAt   time.Time         `json:"generated_at"`
	Operations    []ImportOperation `json:"operations"`
	Conflicts     []ImportConflict  `json:"conflicts"`
}

type importOperationJSON struct {
	Kind       OperationKind   `json:"kind"`
	EntityType EntityType      `json:"entity_type"`
	EntityRef  string          `json:"entity_ref"`
	TargetID   *string         `json:"target_id,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

type importConflictJSON struct {
	EntityType  EntityType          `json:"entity_type"`
	EntityRef   string              `json:"entity_ref"`
	Reason      ConflictReason      `json:"reason"`
	Candidates  []ConflictCandidate `json:"candidates"`
	Recommended OperationKind       `json:"recommended"`
	Payload     json.RawMessage     `json:"payload"`
}

type conflictCandidateJSON struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

func normalizeImportRef(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", NewValidationError("entity_ref", "must not be empty")
	}

	return trimmed, nil
}

func normalizeImportOperationTargetID(kind OperationKind, targetID *string) (*string, error) {
	if kind.IsCreate() {
		if targetID != nil {
			return nil, NewValidationError("target_id", "must be nil for create operations")
		}

		return nil, nil
	}

	if targetID == nil {
		return nil, NewValidationError("target_id", "must not be nil for update, noop, or skip operations")
	}

	trimmed := strings.TrimSpace(*targetID)
	if trimmed == "" {
		return nil, NewValidationError("target_id", "must not be empty")
	}

	return &trimmed, nil
}

func normalizeImportPayload(field string, payload json.RawMessage) (json.RawMessage, error) {
	if len(payload) == 0 {
		return nil, nil
	}
	if !json.Valid(payload) {
		return nil, NewValidationError(field, "must be valid JSON")
	}

	return copyRawMessage(payload), nil
}

func copyRawMessage(value json.RawMessage) json.RawMessage {
	if value == nil {
		return nil
	}

	copied := make(json.RawMessage, len(value))
	copy(copied, value)
	return copied
}

func copyImportOperations(operations []ImportOperation) ([]ImportOperation, error) {
	if operations == nil {
		return nil, nil
	}

	copied := make([]ImportOperation, len(operations))
	for i, operation := range operations {
		if err := operation.validate(); err != nil {
			return nil, err
		}
		copied[i] = operation.copy()
	}

	return copied, nil
}

func copyImportConflicts(conflicts []ImportConflict) ([]ImportConflict, error) {
	if conflicts == nil {
		return nil, nil
	}

	copied := make([]ImportConflict, len(conflicts))
	for i, conflict := range conflicts {
		if err := conflict.validate(); err != nil {
			return nil, err
		}
		copied[i] = conflict.copy()
	}

	return copied, nil
}

func copyConflictCandidates(candidates []ConflictCandidate) ([]ConflictCandidate, error) {
	if candidates == nil {
		return nil, nil
	}

	copied := make([]ConflictCandidate, len(candidates))
	for i, candidate := range candidates {
		if err := candidate.validate(); err != nil {
			return nil, err
		}
		copied[i] = candidate
	}

	return copied, nil
}

func copyAppliedOperations(operations []AppliedOperation) ([]AppliedOperation, error) {
	if operations == nil {
		return nil, nil
	}

	copied := make([]AppliedOperation, len(operations))
	for i, operation := range operations {
		if err := operation.validate(); err != nil {
			return nil, err
		}
		copied[i] = operation.copy()
	}

	return copied, nil
}

func copyFailedOperations(operations []FailedOperation) ([]FailedOperation, error) {
	if operations == nil {
		return nil, nil
	}

	copied := make([]FailedOperation, len(operations))
	for i, operation := range operations {
		if err := operation.validate(); err != nil {
			return nil, err
		}
		copied[i] = operation.copy()
	}

	return copied, nil
}
