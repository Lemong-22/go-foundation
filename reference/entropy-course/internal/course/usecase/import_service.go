package usecase

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const importFormatVersion = "1"

type ImportService struct {
	importSource core.ImportSource
	clock        core.Clock

	courses   core.CourseRepository
	lessons   core.LessonRepository
	quizzes   core.QuizRepository
	practices core.PracticeRepository
	tests     core.TestRepository

	courseService   core.CourseService
	lessonService   core.LessonService
	quizService     core.QuizService
	practiceService core.PracticeService
	testService     core.TestService
}

func NewImportService(
	importSource core.ImportSource,
	clock core.Clock,
	courses core.CourseRepository,
	lessons core.LessonRepository,
	quizzes core.QuizRepository,
	practices core.PracticeRepository,
	tests core.TestRepository,
	courseService core.CourseService,
	lessonService core.LessonService,
	quizService core.QuizService,
	practiceService core.PracticeService,
	testService core.TestService,
) *ImportService {
	return &ImportService{
		importSource:    importSource,
		clock:           clock,
		courses:         courses,
		lessons:         lessons,
		quizzes:         quizzes,
		practices:       practices,
		tests:           tests,
		courseService:   courseService,
		lessonService:   lessonService,
		quizService:     quizService,
		practiceService: practiceService,
		testService:     testService,
	}
}

func (service *ImportService) PlanImport(in core.PlanImportInput) (core.PlanImportOutput, error) {
	instructorID, err := domain.NewInstructorID(in.InstructorID)
	if err != nil {
		return core.PlanImportOutput{}, err
	}

	parsed, metadata, err := service.importSource.Open(in.ZipPath)
	if err != nil {
		return core.PlanImportOutput{}, err
	}
	if metadata.FormatVersion != importFormatVersion {
		return core.PlanImportOutput{}, domain.NewUnsupportedImportFormatError(metadata.FormatVersion, []string{importFormatVersion})
	}

	planner := importPlanner{
		service:      service,
		instructorID: instructorID,
		parsed:       parsed,
		metadata:     metadata,
		quizRefs:     map[string]string{},
		practiceRefs: map[string]string{},
	}
	if err := planner.plan(); err != nil {
		return core.PlanImportOutput{}, err
	}

	plan, err := domain.NewImportPlan(
		metadata.FormatVersion,
		metadata.ZipHash,
		service.clock.Now(),
		planner.operations,
		planner.conflicts,
	)
	if err != nil {
		return core.PlanImportOutput{}, err
	}

	return core.PlanImportOutput{Plan: plan}, nil
}

func (service *ImportService) ApplyPlan(in core.ApplyPlanInput) (core.ApplyPlanOutput, error) {
	if err := service.ensureApplyServices(); err != nil {
		return core.ApplyPlanOutput{}, err
	}
	if _, err := domain.NewInstructorID(in.InstructorID); err != nil {
		return core.ApplyPlanOutput{}, err
	}

	_, metadata, err := service.importSource.Open(in.ZipPath)
	if err != nil {
		return core.ApplyPlanOutput{}, err
	}
	if metadata.FormatVersion != importFormatVersion {
		return core.ApplyPlanOutput{}, domain.NewUnsupportedImportFormatError(metadata.FormatVersion, []string{importFormatVersion})
	}

	operations, err := service.applyOperations(in, metadata)
	if err != nil {
		return core.ApplyPlanOutput{}, err
	}

	applier := importApplier{
		service:     service,
		instructor:  in.InstructorID,
		resolvedIDs: map[string]string{},
	}
	result, err := applier.apply(operations)
	if err != nil {
		return core.ApplyPlanOutput{}, err
	}

	return core.ApplyPlanOutput{Result: result}, nil
}

func (service *ImportService) ensureApplyServices() error {
	if service.courseService == nil ||
		service.lessonService == nil ||
		service.quizService == nil ||
		service.practiceService == nil ||
		service.testService == nil {
		return domain.NewValidationError("import", "apply requires course, lesson, quiz, practice, and test services")
	}

	return nil
}

func (service *ImportService) applyOperations(in core.ApplyPlanInput, metadata core.ImportSourceMetadata) ([]domain.ImportOperation, error) {
	if len(in.ResolvedPlanJSON) > 0 {
		var plan domain.ImportPlan
		if err := json.Unmarshal(in.ResolvedPlanJSON, &plan); err != nil {
			return nil, err
		}
		if plan.ZipHash() != metadata.ZipHash {
			return nil, domain.NewImportPlanHashMismatchError(plan.ZipHash(), metadata.ZipHash)
		}
		if conflicts := plan.Conflicts(); len(conflicts) > 0 {
			return nil, domain.NewUnresolvedImportConflictsError(conflicts)
		}

		return plan.Operations(), nil
	}

	strategy, err := conflictStrategyOrDefault(in.ConflictStrategy)
	if err != nil {
		return nil, err
	}

	out, err := service.PlanImport(core.PlanImportInput{
		ZipPath:      in.ZipPath,
		InstructorID: in.InstructorID,
	})
	if err != nil {
		return nil, err
	}

	plan := out.Plan
	conflicts := plan.Conflicts()
	if len(conflicts) == 0 {
		return plan.Operations(), nil
	}
	if strategy.IsFail() {
		return nil, domain.NewUnresolvedImportConflictsError(conflicts)
	}

	operations := plan.Operations()
	for _, conflict := range conflicts {
		operation, err := operationFromConflict(conflict, strategy)
		if err != nil {
			return nil, err
		}
		operations = append(operations, operation)
	}

	return operations, nil
}

func conflictStrategyOrDefault(value string) (domain.ConflictStrategy, error) {
	if strings.TrimSpace(value) == "" {
		return domain.FailConflicts(), nil
	}

	strategy, err := domain.NewConflictStrategy(value)
	if err != nil {
		return domain.ConflictStrategy{}, domain.NewInvalidConflictStrategyError(value)
	}

	return strategy, nil
}

func operationFromConflict(conflict domain.ImportConflict, strategy domain.ConflictStrategy) (domain.ImportOperation, error) {
	candidates := conflict.Candidates()
	if len(candidates) == 0 {
		return domain.ImportOperation{}, domain.NewValidationError("conflict", "must include at least one candidate")
	}

	targetID := candidates[0].ID()
	kind := domain.SkipOperation()
	if strategy.IsUpdate() {
		kind = domain.UpdateOperation()
	}

	return domain.NewImportOperation(
		kind,
		conflict.EntityType(),
		conflict.EntityRef(),
		&targetID,
		conflict.Payload(),
	)
}

type importApplier struct {
	service    *ImportService
	instructor string

	resolvedIDs map[string]string

	applied []domain.AppliedOperation
	failed  []domain.FailedOperation
	skipped []domain.ImportOperation

	aggregatesSucceeded int
	aggregatesFailed    int
}

func (applier *importApplier) apply(operations []domain.ImportOperation) (domain.ApplyResult, error) {
	groups := groupImportOperations(operations)
	for _, group := range groups {
		failedBefore := len(applier.failed)
		applier.applyGroup(group)
		if len(applier.failed) > failedBefore {
			applier.aggregatesFailed++
			continue
		}
		applier.aggregatesSucceeded++
	}

	return domain.NewApplyResult(
		applier.applied,
		applier.failed,
		applier.skipped,
		applier.aggregatesSucceeded,
		applier.aggregatesFailed,
	)
}

func (applier *importApplier) applyGroup(group importOperationGroup) {
	for _, entry := range group.operations {
		operation := entry.operation
		if operation.Kind().IsNoop() || operation.Kind().IsSkip() {
			applier.rememberOperationID(operation, "")
			applier.skipped = append(applier.skipped, operation)
			continue
		}

		id, message, err := applier.applyOperation(operation)
		if err != nil {
			failed, failedErr := domain.NewFailedOperation(operation, err)
			if failedErr != nil {
				err = failedErr
			} else {
				applier.failed = append(applier.failed, failed)
			}
			return
		}

		applier.rememberOperationID(operation, id)
		applied, err := domain.NewAppliedOperation(operation, message)
		if err != nil {
			failed, failedErr := domain.NewFailedOperation(operation, err)
			if failedErr == nil {
				applier.failed = append(applier.failed, failed)
			}
			return
		}
		applier.applied = append(applier.applied, applied)
	}
}

func (applier *importApplier) applyOperation(operation domain.ImportOperation) (string, string, error) {
	switch {
	case operation.EntityType().IsCourse():
		return applier.applyCourse(operation)
	case operation.EntityType().IsQuiz():
		return applier.applyQuiz(operation)
	case operation.EntityType().IsQuestion():
		return applier.applyQuestion(operation)
	case operation.EntityType().IsPractice():
		return applier.applyPractice(operation)
	case operation.EntityType().IsTestCase():
		return applier.applyTestCase(operation)
	case operation.EntityType().IsTest():
		return applier.applyTest(operation)
	case operation.EntityType().IsTestItem():
		return applier.applyTestItem(operation)
	case operation.EntityType().IsLesson():
		return applier.applyLesson(operation)
	case operation.EntityType().IsBlock():
		return applier.applyBlock(operation)
	default:
		return "", "", domain.NewValidationError("entity_type", "unsupported import entity type")
	}
}

func (applier *importApplier) rememberOperationID(operation domain.ImportOperation, id string) {
	if id == "" {
		if targetID := operation.TargetID(); targetID != nil {
			id = *targetID
		}
	}
	if id == "" {
		return
	}

	applier.resolvedIDs[operation.EntityRef()] = id
	if placeholder := operationPlaceholder(operation); placeholder != "" {
		applier.resolvedIDs[placeholder] = id
	}
}

func (applier *importApplier) resolveID(value string) (string, error) {
	if !strings.HasPrefix(value, "$") {
		return value, nil
	}

	resolved, exists := applier.resolvedIDs[value]
	if !exists {
		return "", domain.NewValidationError("placeholder", fmt.Sprintf("unresolved import placeholder %q", value))
	}

	return resolved, nil
}

func (applier *importApplier) applyCourse(operation domain.ImportOperation) (string, string, error) {
	payload, err := decodePayload[courseImportPayload](operation)
	if err != nil {
		return "", "", err
	}

	if operation.Kind().IsCreate() {
		out, err := applier.service.courseService.CreateCourse(core.CreateCourseInput{
			Title:        payload.Title,
			Slug:         payload.Slug,
			Description:  payload.Description,
			InstructorID: applier.instructor,
		})
		if err != nil {
			return "", "", err
		}
		if err := applier.syncCourseStatus(out.ID, payload.Status); err != nil {
			return "", "", err
		}

		return out.ID, "created course " + out.ID, nil
	}

	targetID := requiredTargetID(operation)
	if _, err := applier.service.courseService.UpdateCourse(core.UpdateCourseInput{
		ID:          targetID,
		Title:       stringRef(payload.Title),
		Description: stringRef(payload.Description),
		Slug:        stringRef(payload.Slug),
	}); err != nil {
		return "", "", err
	}
	if err := applier.syncCourseStatus(targetID, payload.Status); err != nil {
		return "", "", err
	}

	return targetID, "updated course " + targetID, nil
}

func (applier *importApplier) syncCourseStatus(courseID string, statusValue string) error {
	status, err := domain.NewCourseStatus(statusValue)
	if err != nil {
		return err
	}

	id, err := domain.NewCourseID(courseID)
	if err != nil {
		return err
	}
	course, err := applier.service.courses.FindByID(id)
	if err != nil {
		return err
	}

	if status.IsPublished() {
		if course.Status().IsPublished() {
			return nil
		}
		return applier.service.courseService.PublishCourse(core.PublishCourseInput{ID: courseID})
	}

	if !course.Status().IsPublished() {
		return nil
	}
	return applier.service.courseService.UnpublishCourse(core.UnpublishCourseInput{ID: courseID})
}

func (applier *importApplier) applyQuiz(operation domain.ImportOperation) (string, string, error) {
	payload, err := decodePayload[quizImportPayload](operation)
	if err != nil {
		return "", "", err
	}

	courseID, err := applier.resolveID(payload.CourseID)
	if err != nil {
		return "", "", err
	}

	if operation.Kind().IsCreate() {
		out, err := applier.service.quizService.CreateQuiz(core.CreateQuizInput{
			CourseID:      courseID,
			Title:         payload.Title,
			PassThreshold: float64Ref(payload.PassThreshold),
		})
		if err != nil {
			return "", "", err
		}

		return out.ID, "created quiz " + out.ID, nil
	}

	targetID := requiredTargetID(operation)
	if _, err := applier.service.quizService.UpdateQuiz(core.UpdateQuizInput{
		ID:            targetID,
		Title:         stringRef(payload.Title),
		PassThreshold: float64Ref(payload.PassThreshold),
	}); err != nil {
		return "", "", err
	}

	return targetID, "updated quiz " + targetID, nil
}

func (applier *importApplier) applyQuestion(operation domain.ImportOperation) (string, string, error) {
	payload, err := decodePayload[questionImportPayload](operation)
	if err != nil {
		return "", "", err
	}

	quizID, err := applier.resolveID(payload.QuizID)
	if err != nil {
		return "", "", err
	}

	if operation.Kind().IsCreate() {
		out, err := applier.service.quizService.AddQuestion(core.AddQuestionInput{
			QuizID:         quizID,
			Type:           payload.Type,
			Prompt:         payload.Prompt,
			Options:        payload.Options,
			CorrectIndices: payload.CorrectIndices,
			Explanation:    payload.Explanation,
			Position:       intRef(payload.Position),
		})
		if err != nil {
			return "", "", err
		}

		return out.ID, "created question " + out.ID, nil
	}

	targetID := requiredTargetID(operation)
	if _, err := applier.service.quizService.UpdateQuestion(core.UpdateQuestionInput{
		ID:             targetID,
		Prompt:         stringRef(payload.Prompt),
		Options:        stringSliceRef(payload.Options),
		CorrectIndices: intSliceRef(payload.CorrectIndices),
		Explanation:    stringRef(payload.Explanation),
	}); err != nil {
		return "", "", err
	}

	return targetID, "updated question " + targetID, nil
}

func (applier *importApplier) applyPractice(operation domain.ImportOperation) (string, string, error) {
	payload, err := decodePayload[practiceImportPayload](operation)
	if err != nil {
		return "", "", err
	}

	courseID, err := applier.resolveID(payload.CourseID)
	if err != nil {
		return "", "", err
	}

	if operation.Kind().IsCreate() {
		out, err := applier.service.practiceService.CreatePractice(core.CreatePracticeInput{
			CourseID:    courseID,
			Title:       payload.Title,
			Language:    payload.Language,
			Prompt:      payload.Prompt,
			StarterCode: payload.StarterCode,
			Solution:    payload.Solution,
		})
		if err != nil {
			return "", "", err
		}

		return out.ID, "created practice " + out.ID, nil
	}

	targetID := requiredTargetID(operation)
	if _, err := applier.service.practiceService.UpdatePractice(core.UpdatePracticeInput{
		ID:          targetID,
		Title:       stringRef(payload.Title),
		Prompt:      stringRef(payload.Prompt),
		StarterCode: stringRef(payload.StarterCode),
		Solution:    stringRef(payload.Solution),
	}); err != nil {
		return "", "", err
	}

	return targetID, "updated practice " + targetID, nil
}

func (applier *importApplier) applyTestCase(operation domain.ImportOperation) (string, string, error) {
	payload, err := decodePayload[testCaseImportPayload](operation)
	if err != nil {
		return "", "", err
	}

	practiceID, err := applier.resolveID(payload.PracticeID)
	if err != nil {
		return "", "", err
	}

	if operation.Kind().IsCreate() {
		out, err := applier.service.practiceService.AddTestCase(core.AddTestCaseInput{
			PracticeID:     practiceID,
			Stdin:          payload.Stdin,
			ExpectedStdout: payload.ExpectedStdout,
			Name:           payload.Name,
			Position:       intRef(payload.Position),
		})
		if err != nil {
			return "", "", err
		}

		return out.ID, "created test case " + out.ID, nil
	}

	targetID := requiredTargetID(operation)
	if _, err := applier.service.practiceService.UpdateTestCase(core.UpdateTestCaseInput{
		ID:             targetID,
		Stdin:          stringRef(payload.Stdin),
		ExpectedStdout: stringRef(payload.ExpectedStdout),
		Name:           stringRef(payload.Name),
	}); err != nil {
		return "", "", err
	}

	return targetID, "updated test case " + targetID, nil
}

func (applier *importApplier) applyTest(operation domain.ImportOperation) (string, string, error) {
	payload, err := decodePayload[testImportPayload](operation)
	if err != nil {
		return "", "", err
	}

	courseID, err := applier.resolveID(payload.CourseID)
	if err != nil {
		return "", "", err
	}

	if operation.Kind().IsCreate() {
		out, err := applier.service.testService.CreateTest(core.CreateTestInput{
			CourseID:         courseID,
			Title:            payload.Title,
			TimeLimitMinutes: payload.TimeLimitMinutes,
			PassThreshold:    float64Ref(payload.PassThreshold),
		})
		if err != nil {
			return "", "", err
		}
		if payload.Solution != nil {
			if _, err := applier.service.testService.UpdateTest(testSolutionUpdateInput(out.ID, payload)); err != nil {
				return "", "", err
			}
		}

		return out.ID, "created test " + out.ID, nil
	}

	targetID := requiredTargetID(operation)
	if _, err := applier.service.testService.UpdateTest(testUpdateInput(targetID, payload)); err != nil {
		return "", "", err
	}

	return targetID, "updated test " + targetID, nil
}

func (applier *importApplier) applyTestItem(operation domain.ImportOperation) (string, string, error) {
	payload, err := decodePayload[testItemImportPayload](operation)
	if err != nil {
		return "", "", err
	}

	testID, err := applier.resolveID(payload.TestID)
	if err != nil {
		return "", "", err
	}

	if operation.Kind().IsCreate() {
		out, err := applier.service.testService.AddTestItem(core.AddTestItemInput{
			TestID:         testID,
			Kind:           payload.Kind,
			Position:       intRef(payload.Position),
			Prompt:         payload.Prompt,
			ChoiceType:     payload.ChoiceType,
			Options:        payload.Options,
			CorrectIndices: payload.CorrectIndices,
			Explanation:    payload.Explanation,
			CodingPrompt:   payload.CodingPrompt,
			Language:       payload.Language,
			StarterCode:    payload.StarterCode,
			Solution:       payload.Solution,
			TestCases:      payload.TestCases,
		})
		if err != nil {
			return "", "", err
		}

		return out.ID, "created test item " + out.ID, nil
	}

	targetID := requiredTargetID(operation)
	if _, err := applier.service.testService.UpdateTestItem(core.UpdateTestItemInput{
		ID:             targetID,
		Prompt:         stringRef(payload.Prompt),
		ChoiceType:     stringRef(payload.ChoiceType),
		Options:        stringSliceRef(payload.Options),
		CorrectIndices: intSliceRef(payload.CorrectIndices),
		Explanation:    stringRef(payload.Explanation),
		CodingPrompt:   stringRef(payload.CodingPrompt),
		Language:       stringRef(payload.Language),
		StarterCode:    stringRef(payload.StarterCode),
		Solution:       stringRef(payload.Solution),
		TestCases:      codingTestCasesRef(payload.TestCases),
	}); err != nil {
		return "", "", err
	}

	return targetID, "updated test item " + targetID, nil
}

func (applier *importApplier) applyLesson(operation domain.ImportOperation) (string, string, error) {
	payload, err := decodePayload[lessonImportPayload](operation)
	if err != nil {
		return "", "", err
	}

	courseID, err := applier.resolveID(payload.CourseID)
	if err != nil {
		return "", "", err
	}

	if operation.Kind().IsCreate() {
		out, err := applier.service.lessonService.CreateLesson(core.CreateLessonInput{
			CourseID: courseID,
			Title:    payload.Title,
			Order:    intRef(payload.Order),
		})
		if err != nil {
			return "", "", err
		}

		return out.ID, "created lesson " + out.ID, nil
	}

	targetID := requiredTargetID(operation)
	if _, err := applier.service.lessonService.UpdateLesson(core.UpdateLessonInput{
		ID:    targetID,
		Title: stringRef(payload.Title),
	}); err != nil {
		return "", "", err
	}

	return targetID, "updated lesson " + targetID, nil
}

func (applier *importApplier) applyBlock(operation domain.ImportOperation) (string, string, error) {
	payload, err := decodePayload[blockImportPayload](operation)
	if err != nil {
		return "", "", err
	}

	lessonID, err := applier.resolveID(payload.LessonID)
	if err != nil {
		return "", "", err
	}
	quizRef, err := applier.resolveOptionalID(payload.QuizRef)
	if err != nil {
		return "", "", err
	}
	practiceRef, err := applier.resolveOptionalID(payload.PracticeRef)
	if err != nil {
		return "", "", err
	}

	if operation.Kind().IsCreate() {
		out, err := applier.service.lessonService.AddLessonBlock(core.AddLessonBlockInput{
			LessonID:      lessonID,
			Kind:          payload.Kind,
			Markdown:      payload.Markdown,
			VideoProvider: payload.VideoProvider,
			VideoLocator:  payload.VideoLocator,
			VideoCaption:  payload.VideoCaption,
			QuizRef:       quizRef,
			PracticeRef:   practiceRef,
			Position:      intRef(payload.Position),
		})
		if err != nil {
			return "", "", err
		}

		return out.ID, "created block " + out.ID, nil
	}

	targetID := requiredTargetID(operation)
	update, err := blockUpdateInput(targetID, payload)
	if err != nil {
		return "", "", err
	}
	if _, err := applier.service.lessonService.UpdateLessonBlock(update); err != nil {
		return "", "", err
	}

	return targetID, "updated block " + targetID, nil
}

func (applier *importApplier) resolveOptionalID(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", nil
	}

	return applier.resolveID(value)
}

type courseImportPayload struct {
	Title        string `json:"title"`
	Slug         string `json:"slug"`
	Description  string `json:"description"`
	InstructorID string `json:"instructor_id"`
	Status       string `json:"status"`
}

type quizImportPayload struct {
	CourseID      string  `json:"course_id"`
	Slug          string  `json:"slug"`
	Title         string  `json:"title"`
	PassThreshold float64 `json:"pass_threshold"`
	ImportLocalID string  `json:"import_local_id"`
}

type questionImportPayload struct {
	QuizID         string   `json:"quiz_id"`
	Type           string   `json:"type"`
	Prompt         string   `json:"prompt"`
	Options        []string `json:"options"`
	CorrectIndices []int    `json:"correct_indices"`
	Explanation    string   `json:"explanation"`
	Position       int      `json:"position"`
}

type practiceImportPayload struct {
	CourseID      string `json:"course_id"`
	Slug          string `json:"slug"`
	Title         string `json:"title"`
	Language      string `json:"language"`
	Prompt        string `json:"prompt"`
	StarterCode   string `json:"starter_code"`
	Solution      string `json:"solution"`
	ImportLocalID string `json:"import_local_id"`
}

type testCaseImportPayload struct {
	PracticeID     string `json:"practice_id"`
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
	Name           string `json:"name"`
	Position       int    `json:"position"`
}

type testImportPayload struct {
	CourseID         string                     `json:"course_id"`
	Slug             string                     `json:"slug"`
	Title            string                     `json:"title"`
	TimeLimitMinutes *int                       `json:"time_limit_minutes"`
	PassThreshold    float64                    `json:"pass_threshold"`
	Solution         *testSolutionImportPayload `json:"solution"`
	ImportLocalID    string                     `json:"import_local_id"`
}

type testSolutionImportPayload struct {
	ZipProvider   string `json:"zip_provider"`
	ZipLocator    string `json:"zip_locator"`
	VideoProvider string `json:"video_provider"`
	VideoLocator  string `json:"video_locator"`
	VideoCaption  string `json:"video_caption"`
}

type testItemImportPayload struct {
	TestID         string                   `json:"test_id"`
	Kind           string                   `json:"kind"`
	Position       int                      `json:"position"`
	Prompt         string                   `json:"prompt"`
	ChoiceType     string                   `json:"choice_type"`
	Options        []string                 `json:"options"`
	CorrectIndices []int                    `json:"correct_indices"`
	Explanation    string                   `json:"explanation"`
	CodingPrompt   string                   `json:"coding_prompt"`
	Language       string                   `json:"language"`
	StarterCode    string                   `json:"starter_code"`
	Solution       string                   `json:"solution"`
	TestCases      []core.CodingTestCaseDTO `json:"test_cases"`
}

type lessonImportPayload struct {
	CourseID string `json:"course_id"`
	Title    string `json:"title"`
	Order    int    `json:"order"`
}

type blockImportPayload struct {
	LessonID      string `json:"lesson_id"`
	Kind          string `json:"kind"`
	Markdown      string `json:"markdown"`
	VideoProvider string `json:"video_provider"`
	VideoLocator  string `json:"video_locator"`
	VideoCaption  string `json:"video_caption"`
	QuizRef       string `json:"quiz_ref"`
	PracticeRef   string `json:"practice_ref"`
	Position      int    `json:"position"`
}

func decodePayload[T any](operation domain.ImportOperation) (T, error) {
	var payload T
	if err := json.Unmarshal(operation.Payload(), &payload); err != nil {
		return payload, err
	}

	return payload, nil
}

func requiredTargetID(operation domain.ImportOperation) string {
	targetID := operation.TargetID()
	if targetID == nil {
		return ""
	}

	return *targetID
}

func operationPlaceholder(operation domain.ImportOperation) string {
	entityRef := operation.EntityRef()
	switch {
	case operation.EntityType().IsCourse() && strings.HasPrefix(entityRef, "course:"):
		return placeholderID("course", strings.TrimPrefix(entityRef, "course:"))
	case operation.EntityType().IsQuiz() && strings.HasPrefix(entityRef, "quiz:"):
		return placeholderID("quiz", strings.TrimPrefix(entityRef, "quiz:"))
	case operation.EntityType().IsPractice() && strings.HasPrefix(entityRef, "practice:"):
		return placeholderID("practice", strings.TrimPrefix(entityRef, "practice:"))
	case operation.EntityType().IsTest() && strings.HasPrefix(entityRef, "test:"):
		return placeholderID("test", strings.TrimPrefix(entityRef, "test:"))
	case operation.EntityType().IsLesson() && strings.HasPrefix(entityRef, "lesson:"):
		return placeholderID("lesson", strings.TrimPrefix(entityRef, "lesson:"))
	default:
		return ""
	}
}

func testUpdateInput(id string, payload testImportPayload) core.UpdateTestInput {
	in := core.UpdateTestInput{
		ID:            id,
		Title:         stringRef(payload.Title),
		PassThreshold: float64Ref(payload.PassThreshold),
	}
	if payload.TimeLimitMinutes == nil {
		in.TimeLimitMinutes = intRef(0)
	} else {
		in.TimeLimitMinutes = payload.TimeLimitMinutes
	}
	if payload.Solution != nil {
		in.SolutionZipProvider = stringRef(payload.Solution.ZipProvider)
		in.SolutionZipLocator = stringRef(payload.Solution.ZipLocator)
		in.SolutionVideoProvider = stringRef(payload.Solution.VideoProvider)
		in.SolutionVideoLocator = stringRef(payload.Solution.VideoLocator)
		in.SolutionVideoCaption = stringRef(payload.Solution.VideoCaption)
	}

	return in
}

func testSolutionUpdateInput(id string, payload testImportPayload) core.UpdateTestInput {
	return core.UpdateTestInput{
		ID:                    id,
		SolutionZipProvider:   stringRef(payload.Solution.ZipProvider),
		SolutionZipLocator:    stringRef(payload.Solution.ZipLocator),
		SolutionVideoProvider: stringRef(payload.Solution.VideoProvider),
		SolutionVideoLocator:  stringRef(payload.Solution.VideoLocator),
		SolutionVideoCaption:  stringRef(payload.Solution.VideoCaption),
	}
}

func blockUpdateInput(id string, payload blockImportPayload) (core.UpdateLessonBlockInput, error) {
	switch payload.Kind {
	case domain.TextKind().String():
		return core.UpdateLessonBlockInput{
			ID:       id,
			Markdown: stringRef(payload.Markdown),
		}, nil
	case domain.VideoKind().String():
		return core.UpdateLessonBlockInput{
			ID:            id,
			VideoProvider: stringRef(payload.VideoProvider),
			VideoLocator:  stringRef(payload.VideoLocator),
			VideoCaption:  stringRef(payload.VideoCaption),
		}, nil
	default:
		return core.UpdateLessonBlockInput{}, domain.NewValidationError("block", "updating quiz or practice block references is not supported")
	}
}

func stringRef(value string) *string {
	return &value
}

func float64Ref(value float64) *float64 {
	return &value
}

func intRef(value int) *int {
	return &value
}

func stringSliceRef(value []string) *[]string {
	return &value
}

func intSliceRef(value []int) *[]int {
	return &value
}

func codingTestCasesRef(value []core.CodingTestCaseDTO) *[]core.CodingTestCaseDTO {
	return &value
}

type importOperationEntry struct {
	operation domain.ImportOperation
	index     int
}

type importOperationGroup struct {
	key        string
	rank       int
	firstIndex int
	operations []importOperationEntry
}

func groupImportOperations(operations []domain.ImportOperation) []importOperationGroup {
	indexed := map[string]int{}
	groups := []importOperationGroup{}
	for index, operation := range operations {
		key, rank := aggregateGroupKey(operation)
		groupIndex, exists := indexed[key]
		if !exists {
			groupIndex = len(groups)
			indexed[key] = groupIndex
			groups = append(groups, importOperationGroup{
				key:        key,
				rank:       rank,
				firstIndex: index,
			})
		}

		groups[groupIndex].operations = append(groups[groupIndex].operations, importOperationEntry{
			operation: operation,
			index:     index,
		})
	}

	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].rank == groups[j].rank {
			return groups[i].firstIndex < groups[j].firstIndex
		}

		return groups[i].rank < groups[j].rank
	})
	for index := range groups {
		sort.SliceStable(groups[index].operations, func(i, j int) bool {
			left := operationDepth(groups[index].operations[i].operation)
			right := operationDepth(groups[index].operations[j].operation)
			if left == right {
				return groups[index].operations[i].index < groups[index].operations[j].index
			}

			return left < right
		})
	}

	return groups
}

func aggregateGroupKey(operation domain.ImportOperation) (string, int) {
	switch {
	case operation.EntityType().IsCourse():
		return "course:" + operation.EntityRef(), 0
	case operation.EntityType().IsQuiz():
		return operation.EntityRef(), 1
	case operation.EntityType().IsQuestion():
		return "quiz:" + positionedParentRef(operation.EntityRef()), 1
	case operation.EntityType().IsPractice():
		return operation.EntityRef(), 2
	case operation.EntityType().IsTestCase():
		return "practice:" + positionedParentRef(operation.EntityRef()), 2
	case operation.EntityType().IsTest():
		return operation.EntityRef(), 3
	case operation.EntityType().IsTestItem():
		return "test:" + positionedParentRef(operation.EntityRef()), 3
	case operation.EntityType().IsLesson():
		return operation.EntityRef(), 4
	case operation.EntityType().IsBlock():
		return "lesson:" + positionedParentRef(operation.EntityRef()), 4
	default:
		return operation.EntityRef(), 5
	}
}

func positionedParentRef(entityRef string) string {
	withoutEntityType := entityRef
	if colon := strings.Index(withoutEntityType, ":"); colon >= 0 {
		withoutEntityType = withoutEntityType[colon+1:]
	}
	if colon := strings.LastIndex(withoutEntityType, ":"); colon >= 0 {
		return withoutEntityType[:colon]
	}

	return withoutEntityType
}

func operationDepth(operation domain.ImportOperation) int {
	switch {
	case operation.EntityType().IsCourse(),
		operation.EntityType().IsQuiz(),
		operation.EntityType().IsPractice(),
		operation.EntityType().IsTest(),
		operation.EntityType().IsLesson():
		return 0
	default:
		return 1
	}
}

type importPlanner struct {
	service      *ImportService
	instructorID domain.InstructorID
	parsed       core.ParsedImportSource
	metadata     core.ImportSourceMetadata

	targetCourseID *domain.CourseID
	courseRef      string
	quizRefs       map[string]string
	practiceRefs   map[string]string

	existingLessons   []domain.Lesson
	existingQuizzes   []domain.Quiz
	existingPractices []domain.Practice
	existingTests     []domain.Test

	operations []domain.ImportOperation
	conflicts  []domain.ImportConflict
}

func (planner *importPlanner) plan() error {
	if err := planner.planCourse(); err != nil {
		return err
	}
	if planner.targetCourseID != nil {
		if err := planner.loadExistingAggregates(*planner.targetCourseID); err != nil {
			return err
		}
	}
	if err := planner.planQuizzes(); err != nil {
		return err
	}
	if err := planner.planPractices(); err != nil {
		return err
	}
	if err := planner.planTests(); err != nil {
		return err
	}
	if err := planner.planLessons(); err != nil {
		return err
	}

	return nil
}

func (planner *importPlanner) planCourse() error {
	parsed := planner.parsed.Course
	status, err := domain.NewCourseStatus(parsed.Status)
	if err != nil {
		return err
	}

	slug, err := domain.NewSlug(parsed.Slug)
	if err != nil {
		return err
	}

	entityRef := courseEntityRef(parsed)
	parsedPayload, err := payload(map[string]any{
		"title":         parsed.Title,
		"slug":          parsed.Slug,
		"description":   parsed.Description,
		"instructor_id": planner.instructorID.String(),
		"status":        status.String(),
	})
	if err != nil {
		return err
	}

	existing, err := planner.service.courses.FindBySlug(slug)
	if errors.Is(err, domain.ErrNotFound) {
		planner.courseRef = placeholderID("course", parsed.Slug)
		return planner.appendOperation(domain.CreateOperation(), domain.CourseEntity(), entityRef, nil, parsedPayload)
	}
	if err != nil {
		return err
	}

	courseID := existing.ID()
	planner.targetCourseID = &courseID
	planner.courseRef = courseID.String()

	existingPayload, err := payload(map[string]any{
		"title":         existing.Title(),
		"slug":          existing.Slug().String(),
		"description":   existing.Description(),
		"instructor_id": existing.InstructorID().String(),
		"status":        existing.Status().String(),
	})
	if err != nil {
		return err
	}

	targetID := existing.ID().String()
	if samePayload(parsedPayload, existingPayload) {
		return planner.appendOperation(domain.NoopOperation(), domain.CourseEntity(), entityRef, &targetID, parsedPayload)
	}

	return planner.appendConflict(
		domain.CourseEntity(),
		entityRef,
		domain.SlugCollision(),
		[]domain.ConflictCandidate{courseCandidate(existing)},
		domain.UpdateOperation(),
		parsedPayload,
	)
}

func (planner *importPlanner) loadExistingAggregates(courseID domain.CourseID) error {
	lessons, err := planner.service.lessons.FindByCourse(courseID)
	if err != nil {
		return err
	}
	quizzes, err := planner.service.quizzes.FindByCourse(courseID)
	if err != nil {
		return err
	}
	practices, err := planner.service.practices.FindByCourse(courseID)
	if err != nil {
		return err
	}
	tests, err := planner.service.tests.FindByCourse(courseID)
	if err != nil {
		return err
	}

	sortLessons(lessons)
	sortQuizzes(quizzes)
	sortPractices(practices)
	sortTests(tests)

	planner.existingLessons = lessons
	planner.existingQuizzes = quizzes
	planner.existingPractices = practices
	planner.existingTests = tests
	return nil
}

func (planner *importPlanner) planQuizzes() error {
	for _, parsed := range planner.parsed.Quizzes {
		if err := planner.planQuiz(parsed); err != nil {
			return err
		}
	}

	return nil
}

func (planner *importPlanner) planQuiz(parsed core.ParsedQuiz) error {
	entityRef := quizEntityRef(parsed)
	parentID := planner.courseRef
	quizRef := placeholderID("quiz", parsed.Slug)
	existing, exists := findQuizByTitle(planner.existingQuizzes, parsed.Title)
	if exists {
		quizRef = existing.ID().String()
	}
	planner.quizRefs[parsed.Slug] = quizRef

	parsedPayload, err := quizPayload(parentID, parsed)
	if err != nil {
		return err
	}

	if !exists {
		if err := planner.appendOperation(domain.CreateOperation(), domain.QuizEntity(), entityRef, nil, parsedPayload); err != nil {
			return err
		}

		return planner.planParsedQuestions(quizRef, parsed.Slug, parsed.Questions, nil)
	}

	targetID := existing.ID().String()
	existingPayload, err := existingQuizPayload(parentID, parsed.Slug, existing)
	if err != nil {
		return err
	}
	if samePayload(parsedPayload, existingPayload) {
		if err := planner.appendOperation(domain.NoopOperation(), domain.QuizEntity(), entityRef, &targetID, parsedPayload); err != nil {
			return err
		}
	} else if err := planner.appendConflict(
		domain.QuizEntity(),
		entityRef,
		domain.TitleInParentCollision(),
		[]domain.ConflictCandidate{quizCandidate(existing)},
		domain.UpdateOperation(),
		parsedPayload,
	); err != nil {
		return err
	}

	return planner.planParsedQuestions(quizRef, parsed.Slug, parsed.Questions, existing.Questions())
}

func (planner *importPlanner) planParsedQuestions(
	quizID string,
	quizSlug string,
	questions []core.ParsedQuestion,
	existing []domain.ChoiceQuestion,
) error {
	existingByPosition := questionsByPosition(existing)
	for index, parsed := range questions {
		position := positionOrIndex(parsed.Position, index)
		entityRef := positionedEntityRef("question", quizSlug, position)
		parsedPayload, err := questionPayload(quizID, parsed, position)
		if err != nil {
			return err
		}

		existingQuestion, exists := existingByPosition[position]
		if !exists {
			if err := planner.appendOperation(domain.CreateOperation(), domain.QuestionEntity(), entityRef, nil, parsedPayload); err != nil {
				return err
			}
			continue
		}

		targetID := existingQuestion.ID().String()
		existingPayload, err := existingQuestionPayload(quizID, existingQuestion)
		if err != nil {
			return err
		}
		if samePayload(parsedPayload, existingPayload) {
			if err := planner.appendOperation(domain.NoopOperation(), domain.QuestionEntity(), entityRef, &targetID, parsedPayload); err != nil {
				return err
			}
			continue
		}
		if err := planner.appendConflict(
			domain.QuestionEntity(),
			entityRef,
			domain.PositionCollision(),
			[]domain.ConflictCandidate{questionCandidate(existingQuestion)},
			domain.UpdateOperation(),
			parsedPayload,
		); err != nil {
			return err
		}
	}

	return nil
}

func (planner *importPlanner) planPractices() error {
	for _, parsed := range planner.parsed.Practices {
		if err := planner.planPractice(parsed); err != nil {
			return err
		}
	}

	return nil
}

func (planner *importPlanner) planPractice(parsed core.ParsedPractice) error {
	entityRef := practiceEntityRef(parsed)
	parentID := planner.courseRef
	practiceRef := placeholderID("practice", parsed.Slug)
	existing, exists := findPracticeByTitle(planner.existingPractices, parsed.Title)
	if exists {
		practiceRef = existing.ID().String()
	}
	planner.practiceRefs[parsed.Slug] = practiceRef

	parsedPayload, err := practicePayload(parentID, parsed)
	if err != nil {
		return err
	}

	if !exists {
		if err := planner.appendOperation(domain.CreateOperation(), domain.PracticeEntity(), entityRef, nil, parsedPayload); err != nil {
			return err
		}

		return planner.planParsedPracticeTestCases(practiceRef, parsed.Slug, parsed.TestCases, nil)
	}

	targetID := existing.ID().String()
	existingPayload, err := existingPracticePayload(parentID, parsed.Slug, existing)
	if err != nil {
		return err
	}
	if samePayload(parsedPayload, existingPayload) {
		if err := planner.appendOperation(domain.NoopOperation(), domain.PracticeEntity(), entityRef, &targetID, parsedPayload); err != nil {
			return err
		}
	} else if err := planner.appendConflict(
		domain.PracticeEntity(),
		entityRef,
		domain.TitleInParentCollision(),
		[]domain.ConflictCandidate{practiceCandidate(existing)},
		domain.UpdateOperation(),
		parsedPayload,
	); err != nil {
		return err
	}

	return planner.planParsedPracticeTestCases(practiceRef, parsed.Slug, parsed.TestCases, existing.TestCases())
}

func (planner *importPlanner) planParsedPracticeTestCases(
	practiceID string,
	practiceSlug string,
	testCases []core.ParsedPracticeTestCase,
	existing []domain.TestCase,
) error {
	existingByPosition := testCasesByPosition(existing)
	for index, parsed := range testCases {
		position := positionOrIndex(parsed.Position, index)
		entityRef := positionedEntityRef("test_case", practiceSlug, position)
		parsedPayload, err := practiceTestCasePayload(practiceID, parsed, position)
		if err != nil {
			return err
		}

		existingCase, exists := existingByPosition[position]
		if !exists {
			if err := planner.appendOperation(domain.CreateOperation(), domain.TestCaseEntity(), entityRef, nil, parsedPayload); err != nil {
				return err
			}
			continue
		}

		targetID := existingCase.ID().String()
		existingPayload, err := existingPracticeTestCasePayload(practiceID, existingCase)
		if err != nil {
			return err
		}
		if samePayload(parsedPayload, existingPayload) {
			if err := planner.appendOperation(domain.NoopOperation(), domain.TestCaseEntity(), entityRef, &targetID, parsedPayload); err != nil {
				return err
			}
			continue
		}
		if err := planner.appendConflict(
			domain.TestCaseEntity(),
			entityRef,
			domain.PositionCollision(),
			[]domain.ConflictCandidate{testCaseCandidate(existingCase)},
			domain.UpdateOperation(),
			parsedPayload,
		); err != nil {
			return err
		}
	}

	return nil
}

func (planner *importPlanner) planTests() error {
	for _, parsed := range planner.parsed.Tests {
		if err := planner.planTest(parsed); err != nil {
			return err
		}
	}

	return nil
}

func (planner *importPlanner) planTest(parsed core.ParsedTest) error {
	entityRef := testEntityRef(parsed)
	parentID := planner.courseRef
	testRef := placeholderID("test", parsed.Slug)
	existing, exists := findTestByTitle(planner.existingTests, parsed.Title)
	if exists {
		testRef = existing.ID().String()
	}

	parsedPayload, err := testPayload(parentID, parsed)
	if err != nil {
		return err
	}

	if !exists {
		if err := planner.appendOperation(domain.CreateOperation(), domain.TestEntity(), entityRef, nil, parsedPayload); err != nil {
			return err
		}

		return planner.planParsedTestItems(testRef, parsed.Slug, parsed.Items, nil)
	}

	targetID := existing.ID().String()
	existingPayload, err := existingTestPayload(parentID, parsed.Slug, existing)
	if err != nil {
		return err
	}
	if samePayload(parsedPayload, existingPayload) {
		if err := planner.appendOperation(domain.NoopOperation(), domain.TestEntity(), entityRef, &targetID, parsedPayload); err != nil {
			return err
		}
	} else if err := planner.appendConflict(
		domain.TestEntity(),
		entityRef,
		domain.TitleInParentCollision(),
		[]domain.ConflictCandidate{testCandidate(existing)},
		domain.UpdateOperation(),
		parsedPayload,
	); err != nil {
		return err
	}

	return planner.planParsedTestItems(testRef, parsed.Slug, parsed.Items, existing.Items())
}

func (planner *importPlanner) planParsedTestItems(
	testID string,
	testSlug string,
	items []core.ParsedTestItem,
	existing []domain.TestItem,
) error {
	existingByPosition := testItemsByPosition(existing)
	for index, parsed := range items {
		position := positionOrIndex(parsed.Position, index)
		entityRef := positionedEntityRef("test_item", testSlug, position)
		parsedPayload, err := testItemPayload(testID, parsed, position)
		if err != nil {
			return err
		}

		existingItem, exists := existingByPosition[position]
		if !exists {
			if err := planner.appendOperation(domain.CreateOperation(), domain.TestItemEntity(), entityRef, nil, parsedPayload); err != nil {
				return err
			}
			continue
		}

		targetID := existingItem.ID().String()
		existingPayload, err := existingTestItemPayload(testID, existingItem)
		if err != nil {
			return err
		}
		if samePayload(parsedPayload, existingPayload) {
			if err := planner.appendOperation(domain.NoopOperation(), domain.TestItemEntity(), entityRef, &targetID, parsedPayload); err != nil {
				return err
			}
			continue
		}
		if err := planner.appendConflict(
			domain.TestItemEntity(),
			entityRef,
			domain.PositionCollision(),
			[]domain.ConflictCandidate{testItemCandidate(existingItem)},
			domain.UpdateOperation(),
			parsedPayload,
		); err != nil {
			return err
		}
	}

	return nil
}

func (planner *importPlanner) planLessons() error {
	for index, parsed := range planner.parsed.Lessons {
		if err := planner.planLesson(parsed, index); err != nil {
			return err
		}
	}

	return nil
}

func (planner *importPlanner) planLesson(parsed core.ParsedLesson, index int) error {
	entityRef := lessonEntityRef(parsed)
	parentID := planner.courseRef
	lessonRef := placeholderID("lesson", parsed.Title)
	existing, exists := findLessonByTitle(planner.existingLessons, parsed.Title)
	if exists {
		lessonRef = existing.ID().String()
	}

	parsedPayload, err := lessonPayload(parentID, parsed, positionOrIndex(parsed.Order, index))
	if err != nil {
		return err
	}

	if !exists {
		if err := planner.appendOperation(domain.CreateOperation(), domain.LessonEntity(), entityRef, nil, parsedPayload); err != nil {
			return err
		}

		return planner.planParsedLessonBlocks(lessonRef, parsed.Title, parsed.Blocks, nil)
	}

	targetID := existing.ID().String()
	existingPayload, err := existingLessonPayload(parentID, existing)
	if err != nil {
		return err
	}
	if samePayload(parsedPayload, existingPayload) {
		if err := planner.appendOperation(domain.NoopOperation(), domain.LessonEntity(), entityRef, &targetID, parsedPayload); err != nil {
			return err
		}
	} else if err := planner.appendConflict(
		domain.LessonEntity(),
		entityRef,
		domain.TitleInParentCollision(),
		[]domain.ConflictCandidate{lessonCandidate(existing)},
		domain.UpdateOperation(),
		parsedPayload,
	); err != nil {
		return err
	}

	return planner.planParsedLessonBlocks(lessonRef, parsed.Title, parsed.Blocks, existing.Blocks())
}

func (planner *importPlanner) planParsedLessonBlocks(
	lessonID string,
	lessonTitle string,
	blocks []core.ParsedLessonBlock,
	existing []domain.ContentBlock,
) error {
	existingByPosition := blocksByPosition(existing)
	for index, parsed := range blocks {
		position := positionOrIndex(parsed.Position, index)
		entityRef := positionedEntityRef("block", lessonTitle, position)
		parsedPayload, err := planner.lessonBlockPayload(lessonID, parsed, position)
		if err != nil {
			return err
		}

		existingBlock, exists := existingByPosition[position]
		if !exists {
			if err := planner.appendOperation(domain.CreateOperation(), domain.BlockEntity(), entityRef, nil, parsedPayload); err != nil {
				return err
			}
			continue
		}

		targetID := existingBlock.ID().String()
		existingPayload, err := existingLessonBlockPayload(lessonID, existingBlock)
		if err != nil {
			return err
		}
		if samePayload(parsedPayload, existingPayload) {
			if err := planner.appendOperation(domain.NoopOperation(), domain.BlockEntity(), entityRef, &targetID, parsedPayload); err != nil {
				return err
			}
			continue
		}
		if err := planner.appendConflict(
			domain.BlockEntity(),
			entityRef,
			domain.PositionCollision(),
			[]domain.ConflictCandidate{blockCandidate(existingBlock)},
			domain.UpdateOperation(),
			parsedPayload,
		); err != nil {
			return err
		}
	}

	return nil
}

func (planner *importPlanner) lessonBlockPayload(lessonID string, block core.ParsedLessonBlock, position int) (json.RawMessage, error) {
	quizRef := ""
	if block.Kind == "quiz" {
		resolved, exists := planner.quizRefs[block.QuizRef]
		if !exists {
			return nil, domain.NewValidationError("quiz_ref", "must reference an imported quiz")
		}
		quizRef = resolved
	}

	practiceRef := ""
	if block.Kind == "practice" {
		resolved, exists := planner.practiceRefs[block.PracticeRef]
		if !exists {
			return nil, domain.NewValidationError("practice_ref", "must reference an imported practice")
		}
		practiceRef = resolved
	}

	return payload(map[string]any{
		"lesson_id":      lessonID,
		"kind":           block.Kind,
		"markdown":       block.Markdown,
		"video_provider": block.VideoProvider,
		"video_locator":  block.VideoLocator,
		"video_caption":  block.VideoCaption,
		"quiz_ref":       quizRef,
		"practice_ref":   practiceRef,
		"position":       position,
	})
}

func (planner *importPlanner) appendOperation(
	kind domain.OperationKind,
	entityType domain.EntityType,
	entityRef string,
	targetID *string,
	payload json.RawMessage,
) error {
	operation, err := domain.NewImportOperation(kind, entityType, entityRef, targetID, payload)
	if err != nil {
		return err
	}

	planner.operations = append(planner.operations, operation)
	return nil
}

func (planner *importPlanner) appendConflict(
	entityType domain.EntityType,
	entityRef string,
	reason domain.ConflictReason,
	candidates []domain.ConflictCandidate,
	recommended domain.OperationKind,
	payload json.RawMessage,
) error {
	conflict, err := domain.NewImportConflict(entityType, entityRef, reason, candidates, recommended, payload)
	if err != nil {
		return err
	}

	planner.conflicts = append(planner.conflicts, conflict)
	return nil
}

func quizPayload(courseID string, quiz core.ParsedQuiz) (json.RawMessage, error) {
	return payload(map[string]any{
		"course_id":       courseID,
		"slug":            quiz.Slug,
		"title":           quiz.Title,
		"pass_threshold":  passThresholdValue(quiz.PassThreshold),
		"import_local_id": quiz.Slug,
	})
}

func existingQuizPayload(courseID string, slug string, quiz domain.Quiz) (json.RawMessage, error) {
	threshold := quiz.PassThreshold().Float64()
	return payload(map[string]any{
		"course_id":       courseID,
		"slug":            slug,
		"title":           quiz.Title(),
		"pass_threshold":  threshold,
		"import_local_id": slug,
	})
}

func questionPayload(quizID string, question core.ParsedQuestion, position int) (json.RawMessage, error) {
	return payload(map[string]any{
		"quiz_id":         quizID,
		"type":            question.Type,
		"prompt":          question.Prompt,
		"options":         question.Options,
		"correct_indices": question.CorrectIndices,
		"explanation":     question.Explanation,
		"position":        position,
	})
}

func existingQuestionPayload(quizID string, question domain.ChoiceQuestion) (json.RawMessage, error) {
	return payload(map[string]any{
		"quiz_id":         quizID,
		"type":            question.Type().String(),
		"prompt":          question.Prompt(),
		"options":         question.Options(),
		"correct_indices": question.CorrectIndices(),
		"explanation":     question.Explanation(),
		"position":        question.Position().Int(),
	})
}

func practicePayload(courseID string, practice core.ParsedPractice) (json.RawMessage, error) {
	return payload(map[string]any{
		"course_id":       courseID,
		"slug":            practice.Slug,
		"title":           practice.Title,
		"language":        practice.Language,
		"prompt":          practice.Prompt,
		"starter_code":    practice.StarterCode,
		"solution":        practice.Solution,
		"import_local_id": practice.Slug,
	})
}

func existingPracticePayload(courseID string, slug string, practice domain.Practice) (json.RawMessage, error) {
	return payload(map[string]any{
		"course_id":       courseID,
		"slug":            slug,
		"title":           practice.Title(),
		"language":        practice.Language().String(),
		"prompt":          practice.Prompt(),
		"starter_code":    practice.StarterCode(),
		"solution":        practice.Solution(),
		"import_local_id": slug,
	})
}

func practiceTestCasePayload(practiceID string, testCase core.ParsedPracticeTestCase, position int) (json.RawMessage, error) {
	return payload(map[string]any{
		"practice_id":     practiceID,
		"stdin":           testCase.Stdin,
		"expected_stdout": testCase.ExpectedStdout,
		"name":            testCase.Name,
		"position":        position,
	})
}

func existingPracticeTestCasePayload(practiceID string, testCase domain.TestCase) (json.RawMessage, error) {
	return payload(map[string]any{
		"practice_id":     practiceID,
		"stdin":           testCase.Stdin(),
		"expected_stdout": testCase.ExpectedStdout(),
		"name":            testCase.Name(),
		"position":        testCase.Position().Int(),
	})
}

func testPayload(courseID string, test core.ParsedTest) (json.RawMessage, error) {
	return payload(map[string]any{
		"course_id":          courseID,
		"slug":               test.Slug,
		"title":              test.Title,
		"time_limit_minutes": pointerValue(test.TimeLimitMinutes),
		"pass_threshold":     passThresholdValue(test.PassThreshold),
		"solution":           parsedTestSolutionPayload(test.Solution),
		"import_local_id":    test.Slug,
	})
}

func existingTestPayload(courseID string, slug string, test domain.Test) (json.RawMessage, error) {
	return payload(map[string]any{
		"course_id":          courseID,
		"slug":               slug,
		"title":              test.Title(),
		"time_limit_minutes": timeLimitPayload(test.TimeLimit()),
		"pass_threshold":     test.PassThreshold().Float64(),
		"solution":           testSolutionPayload(test.Solution()),
		"import_local_id":    slug,
	})
}

func testItemPayload(testID string, item core.ParsedTestItem, position int) (json.RawMessage, error) {
	return payload(map[string]any{
		"test_id":         testID,
		"kind":            item.Kind,
		"position":        position,
		"prompt":          item.Prompt,
		"choice_type":     item.ChoiceType,
		"options":         item.Options,
		"correct_indices": item.CorrectIndices,
		"explanation":     item.Explanation,
		"coding_prompt":   item.CodingPrompt,
		"language":        item.Language,
		"starter_code":    item.StarterCode,
		"solution":        item.Solution,
		"test_cases":      item.TestCases,
	})
}

func existingTestItemPayload(testID string, item domain.TestItem) (json.RawMessage, error) {
	payloadFields := map[string]any{
		"test_id":  testID,
		"kind":     item.Kind().String(),
		"position": item.Position().Int(),
	}

	switch body := item.Body().(type) {
	case domain.ChoiceItemBody:
		payloadFields["prompt"] = body.Prompt()
		payloadFields["choice_type"] = body.Type().String()
		payloadFields["options"] = body.Options()
		payloadFields["correct_indices"] = body.CorrectIndices()
		payloadFields["explanation"] = body.Explanation()
		payloadFields["coding_prompt"] = ""
		payloadFields["language"] = ""
		payloadFields["starter_code"] = ""
		payloadFields["solution"] = ""
		payloadFields["test_cases"] = []core.CodingTestCaseDTO(nil)
	case domain.CodingItemBody:
		payloadFields["prompt"] = ""
		payloadFields["choice_type"] = ""
		payloadFields["options"] = []string(nil)
		payloadFields["correct_indices"] = []int(nil)
		payloadFields["explanation"] = ""
		payloadFields["coding_prompt"] = body.Prompt()
		payloadFields["language"] = body.Language().String()
		payloadFields["starter_code"] = body.StarterCode()
		payloadFields["solution"] = body.Solution()
		payloadFields["test_cases"] = codingTestCaseDTOs(body.TestCases())
	}

	return payload(payloadFields)
}

func lessonPayload(courseID string, lesson core.ParsedLesson, order int) (json.RawMessage, error) {
	return payload(map[string]any{
		"course_id": courseID,
		"title":     lesson.Title,
		"order":     order,
	})
}

func existingLessonPayload(courseID string, lesson domain.Lesson) (json.RawMessage, error) {
	return payload(map[string]any{
		"course_id": courseID,
		"title":     lesson.Title(),
		"order":     lesson.Order().Int(),
	})
}

func existingLessonBlockPayload(lessonID string, block domain.ContentBlock) (json.RawMessage, error) {
	payloadFields := map[string]any{
		"lesson_id":      lessonID,
		"kind":           block.Kind().String(),
		"markdown":       "",
		"video_provider": "",
		"video_locator":  "",
		"video_caption":  "",
		"quiz_ref":       "",
		"practice_ref":   "",
		"position":       block.Position().Int(),
	}

	switch body := block.Body().(type) {
	case domain.TextBody:
		payloadFields["markdown"] = body.Markdown
	case domain.VideoBody:
		payloadFields["video_provider"] = body.Media.Provider().String()
		payloadFields["video_locator"] = body.Media.Locator()
		payloadFields["video_caption"] = body.Caption
	case domain.QuizBody:
		payloadFields["quiz_ref"] = body.QuizRef.String()
	case domain.PracticeBody:
		payloadFields["practice_ref"] = body.PracticeRef.String()
	}

	return payload(payloadFields)
}

func payload(value any) (json.RawMessage, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return json.RawMessage(data), nil
}

func samePayload(left json.RawMessage, right json.RawMessage) bool {
	return string(left) == string(right)
}

func passThresholdValue(value *float64) float64 {
	if value == nil {
		return domain.DefaultPassThreshold().Float64()
	}

	return *value
}

func pointerValue(value *int) any {
	if value == nil {
		return nil
	}

	return *value
}

func timeLimitPayload(value *domain.TimeLimit) any {
	if value == nil {
		return nil
	}

	return value.Minutes()
}

func parsedTestSolutionPayload(solution *core.ParsedTestSolution) any {
	if solution == nil {
		return nil
	}

	return map[string]any{
		"zip_provider":   solution.ZipProvider,
		"zip_locator":    solution.ZipLocator,
		"video_provider": solution.VideoProvider,
		"video_locator":  solution.VideoLocator,
		"video_caption":  solution.VideoCaption,
	}
}

func testSolutionPayload(solution *domain.TestSolution) any {
	if solution == nil {
		return nil
	}

	zip := solution.SolutionZip()
	video := solution.ExplanationVideo()
	return map[string]any{
		"zip_provider":   zip.Provider().String(),
		"zip_locator":    zip.Locator(),
		"video_provider": video.Provider().String(),
		"video_locator":  video.Locator(),
		"video_caption":  solution.ExplanationCaption(),
	}
}

func positionOrIndex(position *int, index int) int {
	if position == nil {
		return index
	}

	return *position
}

func placeholderID(entityType string, localID string) string {
	return fmt.Sprintf("$%s:%s", entityType, localID)
}

func courseEntityRef(course core.ParsedCourse) string {
	return "course:" + course.Slug
}

func quizEntityRef(quiz core.ParsedQuiz) string {
	return "quiz:" + quiz.Slug
}

func practiceEntityRef(practice core.ParsedPractice) string {
	return "practice:" + practice.Slug
}

func testEntityRef(test core.ParsedTest) string {
	return "test:" + test.Slug
}

func lessonEntityRef(lesson core.ParsedLesson) string {
	return "lesson:" + lesson.Title
}

func positionedEntityRef(entityType string, parentRef string, position int) string {
	return fmt.Sprintf("%s:%s:%d", entityType, parentRef, position)
}

func findQuizByTitle(quizzes []domain.Quiz, title string) (domain.Quiz, bool) {
	for _, quiz := range quizzes {
		if quiz.Title() == title {
			return quiz, true
		}
	}

	return domain.Quiz{}, false
}

func findPracticeByTitle(practices []domain.Practice, title string) (domain.Practice, bool) {
	for _, practice := range practices {
		if practice.Title() == title {
			return practice, true
		}
	}

	return domain.Practice{}, false
}

func findTestByTitle(tests []domain.Test, title string) (domain.Test, bool) {
	for _, test := range tests {
		if test.Title() == title {
			return test, true
		}
	}

	return domain.Test{}, false
}

func findLessonByTitle(lessons []domain.Lesson, title string) (domain.Lesson, bool) {
	for _, lesson := range lessons {
		if lesson.Title() == title {
			return lesson, true
		}
	}

	return domain.Lesson{}, false
}

func questionsByPosition(questions []domain.ChoiceQuestion) map[int]domain.ChoiceQuestion {
	indexed := map[int]domain.ChoiceQuestion{}
	for _, question := range questions {
		indexed[question.Position().Int()] = question
	}

	return indexed
}

func testCasesByPosition(testCases []domain.TestCase) map[int]domain.TestCase {
	indexed := map[int]domain.TestCase{}
	for _, testCase := range testCases {
		indexed[testCase.Position().Int()] = testCase
	}

	return indexed
}

func testItemsByPosition(items []domain.TestItem) map[int]domain.TestItem {
	indexed := map[int]domain.TestItem{}
	for _, item := range items {
		indexed[item.Position().Int()] = item
	}

	return indexed
}

func blocksByPosition(blocks []domain.ContentBlock) map[int]domain.ContentBlock {
	indexed := map[int]domain.ContentBlock{}
	for _, block := range blocks {
		indexed[block.Position().Int()] = block
	}

	return indexed
}

func sortLessons(lessons []domain.Lesson) {
	sort.Slice(lessons, func(i, j int) bool {
		if lessons[i].Title() == lessons[j].Title() {
			return lessons[i].ID().String() < lessons[j].ID().String()
		}

		return lessons[i].Title() < lessons[j].Title()
	})
}

func sortQuizzes(quizzes []domain.Quiz) {
	sort.Slice(quizzes, func(i, j int) bool {
		if quizzes[i].Title() == quizzes[j].Title() {
			return quizzes[i].ID().String() < quizzes[j].ID().String()
		}

		return quizzes[i].Title() < quizzes[j].Title()
	})
}

func sortPractices(practices []domain.Practice) {
	sort.Slice(practices, func(i, j int) bool {
		if practices[i].Title() == practices[j].Title() {
			return practices[i].ID().String() < practices[j].ID().String()
		}

		return practices[i].Title() < practices[j].Title()
	})
}

func sortTests(tests []domain.Test) {
	sort.Slice(tests, func(i, j int) bool {
		if tests[i].Title() == tests[j].Title() {
			return tests[i].ID().String() < tests[j].ID().String()
		}

		return tests[i].Title() < tests[j].Title()
	})
}

func courseCandidate(course domain.Course) domain.ConflictCandidate {
	candidate, _ := domain.NewConflictCandidate(course.ID().String(), fmt.Sprintf("course %s %q", course.ID().String(), course.Title()))
	return candidate
}

func lessonCandidate(lesson domain.Lesson) domain.ConflictCandidate {
	candidate, _ := domain.NewConflictCandidate(lesson.ID().String(), fmt.Sprintf("lesson %s %q", lesson.ID().String(), lesson.Title()))
	return candidate
}

func blockCandidate(block domain.ContentBlock) domain.ConflictCandidate {
	candidate, _ := domain.NewConflictCandidate(block.ID().String(), fmt.Sprintf("block %s position %d", block.ID().String(), block.Position().Int()))
	return candidate
}

func quizCandidate(quiz domain.Quiz) domain.ConflictCandidate {
	candidate, _ := domain.NewConflictCandidate(quiz.ID().String(), fmt.Sprintf("quiz %s %q", quiz.ID().String(), quiz.Title()))
	return candidate
}

func questionCandidate(question domain.ChoiceQuestion) domain.ConflictCandidate {
	candidate, _ := domain.NewConflictCandidate(question.ID().String(), fmt.Sprintf("question %s position %d", question.ID().String(), question.Position().Int()))
	return candidate
}

func practiceCandidate(practice domain.Practice) domain.ConflictCandidate {
	candidate, _ := domain.NewConflictCandidate(practice.ID().String(), fmt.Sprintf("practice %s %q", practice.ID().String(), practice.Title()))
	return candidate
}

func testCaseCandidate(testCase domain.TestCase) domain.ConflictCandidate {
	candidate, _ := domain.NewConflictCandidate(testCase.ID().String(), fmt.Sprintf("test case %s position %d", testCase.ID().String(), testCase.Position().Int()))
	return candidate
}

func testCandidate(test domain.Test) domain.ConflictCandidate {
	candidate, _ := domain.NewConflictCandidate(test.ID().String(), fmt.Sprintf("test %s %q", test.ID().String(), test.Title()))
	return candidate
}

func testItemCandidate(item domain.TestItem) domain.ConflictCandidate {
	candidate, _ := domain.NewConflictCandidate(item.ID().String(), fmt.Sprintf("test item %s position %d", item.ID().String(), item.Position().Int()))
	return candidate
}
