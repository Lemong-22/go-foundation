package domain

import (
	"sort"
	"strings"
	"time"
)

const (
	testItemKindChoice = "choice"
	testItemKindCoding = "coding"
)

type TestItemKind struct {
	value string
}

func ChoiceKind() TestItemKind {
	return TestItemKind{value: testItemKindChoice}
}

func CodingKind() TestItemKind {
	return TestItemKind{value: testItemKindCoding}
}

func NewTestItemKind(value string) (TestItemKind, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case testItemKindChoice:
		return ChoiceKind(), nil
	case testItemKindCoding:
		return CodingKind(), nil
	default:
		return TestItemKind{}, NewValidationError("kind", "must be choice or coding")
	}
}

func (kind TestItemKind) String() string {
	return kind.value
}

func (kind TestItemKind) IsChoice() bool {
	return kind.value == testItemKindChoice
}

func (kind TestItemKind) IsCoding() bool {
	return kind.value == testItemKindCoding
}

type TestItemPosition struct {
	value int
}

func NewTestItemPosition(value int) (TestItemPosition, error) {
	if value < 0 {
		return TestItemPosition{}, NewValidationError("position", "must be greater than or equal to zero")
	}

	return TestItemPosition{value: value}, nil
}

func (position TestItemPosition) Int() int {
	return position.value
}

type TimeLimit struct {
	minutes int
}

func NewTimeLimit(minutes int) (TimeLimit, error) {
	if minutes <= 0 {
		return TimeLimit{}, NewValidationError("time_limit_minutes", "must be greater than zero")
	}

	return TimeLimit{minutes: minutes}, nil
}

func (limit TimeLimit) Minutes() int {
	return limit.minutes
}

type TestSolution struct {
	solutionZip        MediaRef
	explanationVideo   MediaRef
	explanationCaption string
}

func NewTestSolution(solutionZip MediaRef, explanationVideo MediaRef, explanationCaption string) (TestSolution, error) {
	if err := validateMediaRef("solution_zip", solutionZip); err != nil {
		return TestSolution{}, err
	}
	if err := validateMediaRef("explanation_video", explanationVideo); err != nil {
		return TestSolution{}, err
	}

	return TestSolution{
		solutionZip:        solutionZip,
		explanationVideo:   explanationVideo,
		explanationCaption: explanationCaption,
	}, nil
}

func (solution TestSolution) SolutionZip() MediaRef {
	return solution.solutionZip
}

func (solution TestSolution) ExplanationVideo() MediaRef {
	return solution.explanationVideo
}

func (solution TestSolution) ExplanationCaption() string {
	return solution.explanationCaption
}

type TestItemBody interface {
	Kind() TestItemKind
	isTestItemBody()
}

type ChoiceItemBody struct {
	questionType   ChoiceQuestionType
	prompt         string
	options        []string
	correctIndices []int
	explanation    string
}

func NewChoiceItemBody(
	questionType ChoiceQuestionType,
	prompt string,
	options []string,
	correctIndices []int,
	explanation string,
) (ChoiceItemBody, error) {
	normalizedPrompt, err := normalizeQuestionPrompt(prompt)
	if err != nil {
		return ChoiceItemBody{}, err
	}
	if err := validateChoiceQuestionType(questionType); err != nil {
		return ChoiceItemBody{}, err
	}
	if err := validateQuestionContent(questionType, options, correctIndices); err != nil {
		return ChoiceItemBody{}, err
	}

	return ChoiceItemBody{
		questionType:   questionType,
		prompt:         normalizedPrompt,
		options:        copyStrings(options),
		correctIndices: copyInts(correctIndices),
		explanation:    explanation,
	}, nil
}

func (body ChoiceItemBody) Kind() TestItemKind {
	return ChoiceKind()
}

func (ChoiceItemBody) isTestItemBody() {}

func (body ChoiceItemBody) Type() ChoiceQuestionType {
	return body.questionType
}

func (body ChoiceItemBody) Prompt() string {
	return body.prompt
}

func (body ChoiceItemBody) Options() []string {
	return copyStrings(body.options)
}

func (body ChoiceItemBody) CorrectIndices() []int {
	return copyInts(body.correctIndices)
}

func (body ChoiceItemBody) Explanation() string {
	return body.explanation
}

func (body ChoiceItemBody) validate() error {
	if err := validateChoiceQuestionType(body.questionType); err != nil {
		return err
	}
	if _, err := normalizeQuestionPrompt(body.prompt); err != nil {
		return err
	}

	return validateQuestionContent(body.questionType, body.options, body.correctIndices)
}

type CodingTestCase struct {
	stdin          string
	expectedStdout string
	name           string
}

func NewCodingTestCase(stdin string, expectedStdout string, name string) CodingTestCase {
	return CodingTestCase{
		stdin:          stdin,
		expectedStdout: expectedStdout,
		name:           name,
	}
}

func (testCase CodingTestCase) Stdin() string {
	return testCase.stdin
}

func (testCase CodingTestCase) ExpectedStdout() string {
	return testCase.expectedStdout
}

func (testCase CodingTestCase) Name() string {
	return testCase.name
}

type CodingItemBody struct {
	language    Language
	prompt      string
	starterCode string
	solution    string
	testCases   []CodingTestCase
}

func NewCodingItemBody(
	language Language,
	prompt string,
	starterCode string,
	solution string,
	testCases []CodingTestCase,
) (CodingItemBody, error) {
	if err := validateLanguage(language); err != nil {
		return CodingItemBody{}, err
	}
	normalizedPrompt, err := normalizePracticePrompt(prompt)
	if err != nil {
		return CodingItemBody{}, err
	}
	if len(testCases) == 0 {
		return CodingItemBody{}, NewValidationError("test_cases", "must include at least one test case")
	}

	return CodingItemBody{
		language:    language,
		prompt:      normalizedPrompt,
		starterCode: starterCode,
		solution:    solution,
		testCases:   copyCodingTestCases(testCases),
	}, nil
}

func (body CodingItemBody) Kind() TestItemKind {
	return CodingKind()
}

func (CodingItemBody) isTestItemBody() {}

func (body CodingItemBody) Language() Language {
	return body.language
}

func (body CodingItemBody) Prompt() string {
	return body.prompt
}

func (body CodingItemBody) StarterCode() string {
	return body.starterCode
}

func (body CodingItemBody) Solution() string {
	return body.solution
}

func (body CodingItemBody) TestCases() []CodingTestCase {
	return copyCodingTestCases(body.testCases)
}

func (body CodingItemBody) validate() error {
	if err := validateLanguage(body.language); err != nil {
		return err
	}
	if _, err := normalizePracticePrompt(body.prompt); err != nil {
		return err
	}
	if len(body.testCases) == 0 {
		return NewValidationError("test_cases", "must include at least one test case")
	}

	return nil
}

type TestItem struct {
	id       TestItemID
	kind     TestItemKind
	body     TestItemBody
	position TestItemPosition
}

func NewTestItem(
	id TestItemID,
	kind TestItemKind,
	body TestItemBody,
	position TestItemPosition,
) (TestItem, error) {
	item := TestItem{
		id:       id,
		kind:     kind,
		body:     body,
		position: position,
	}
	if err := item.validate(); err != nil {
		return TestItem{}, err
	}

	return item, nil
}

func (item TestItem) ID() TestItemID {
	return item.id
}

func (item TestItem) Kind() TestItemKind {
	return item.kind
}

func (item TestItem) Body() TestItemBody {
	return item.body
}

func (item TestItem) Position() TestItemPosition {
	return item.position
}

func (item *TestItem) ReplaceBody(body TestItemBody) error {
	if body == nil {
		return NewValidationError("body", "must not be nil")
	}
	if body.Kind() != item.kind {
		return NewValidationError("body", "must match item kind")
	}
	if err := validateTestItemBody(body); err != nil {
		return err
	}

	item.body = body
	return nil
}

func (item *TestItem) MoveTo(position TestItemPosition) {
	item.position = position
}

func (item TestItem) validate() error {
	if item.id.String() == "" {
		return NewValidationError("test_item_id", "must not be empty")
	}
	if err := validateTestItemKind(item.kind); err != nil {
		return err
	}
	if item.body == nil {
		return NewValidationError("body", "must not be nil")
	}
	if item.body.Kind() != item.kind {
		return NewValidationError("body", "must match item kind")
	}

	return validateTestItemBody(item.body)
}

type Test struct {
	id            TestID
	courseID      CourseID
	title         string
	timeLimit     *TimeLimit
	passThreshold PassThreshold
	solution      *TestSolution
	items         []TestItem
	createdAt     time.Time
	updatedAt     time.Time
}

func NewTest(
	id TestID,
	courseID CourseID,
	title string,
	timeLimit *TimeLimit,
	passThreshold PassThreshold,
	solution *TestSolution,
	items []TestItem,
	now time.Time,
) (Test, error) {
	return RestoreTest(id, courseID, title, timeLimit, passThreshold, solution, items, now, now)
}

func RestoreTest(
	id TestID,
	courseID CourseID,
	title string,
	timeLimit *TimeLimit,
	passThreshold PassThreshold,
	solution *TestSolution,
	items []TestItem,
	createdAt time.Time,
	updatedAt time.Time,
) (Test, error) {
	if id.String() == "" {
		return Test{}, NewValidationError("test_id", "must not be empty")
	}
	if courseID.String() == "" {
		return Test{}, NewValidationError("course_id", "must not be empty")
	}
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return Test{}, err
	}
	if timeLimit != nil {
		if _, err := NewTimeLimit(timeLimit.Minutes()); err != nil {
			return Test{}, err
		}
	}
	if _, err := NewPassThreshold(passThreshold.Float64()); err != nil {
		return Test{}, err
	}
	if solution != nil {
		if _, err := NewTestSolution(solution.SolutionZip(), solution.ExplanationVideo(), solution.ExplanationCaption()); err != nil {
			return Test{}, err
		}
	}
	if updatedAt.Before(createdAt) {
		return Test{}, NewValidationError("updated_at", "must be greater than or equal to created_at")
	}

	normalizedItems, err := normalizeTestItems(items)
	if err != nil {
		return Test{}, err
	}

	return Test{
		id:            id,
		courseID:      courseID,
		title:         normalizedTitle,
		timeLimit:     copyTimeLimit(timeLimit),
		passThreshold: passThreshold,
		solution:      copyTestSolution(solution),
		items:         normalizedItems,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

func (test Test) ID() TestID {
	return test.id
}

func (test Test) CourseID() CourseID {
	return test.courseID
}

func (test Test) Title() string {
	return test.title
}

func (test Test) TimeLimit() *TimeLimit {
	return copyTimeLimit(test.timeLimit)
}

func (test Test) PassThreshold() PassThreshold {
	return test.passThreshold
}

func (test Test) Solution() *TestSolution {
	return copyTestSolution(test.solution)
}

func (test Test) Items() []TestItem {
	items := make([]TestItem, len(test.items))
	copy(items, test.items)

	return items
}

func (test Test) Item(id TestItemID) (TestItem, error) {
	for _, item := range test.items {
		if item.ID() == id {
			return item, nil
		}
	}

	return TestItem{}, ErrNotFound
}

func (test Test) CreatedAt() time.Time {
	return test.createdAt
}

func (test Test) UpdatedAt() time.Time {
	return test.updatedAt
}

func (test *Test) Rename(title string, now time.Time) error {
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return err
	}

	test.title = normalizedTitle
	test.touch(now)

	return nil
}

func (test *Test) ChangeTimeLimit(timeLimit *TimeLimit, now time.Time) {
	test.timeLimit = copyTimeLimit(timeLimit)
	test.touch(now)
}

func (test *Test) ChangePassThreshold(threshold PassThreshold, now time.Time) {
	test.passThreshold = threshold
	test.touch(now)
}

func (test *Test) SetSolution(solution TestSolution, now time.Time) {
	test.solution = copyTestSolution(&solution)
	test.touch(now)
}

func (test *Test) AddItem(item TestItem, now time.Time) error {
	if err := item.validate(); err != nil {
		return err
	}
	if test.hasItemID(item.ID()) {
		return NewValidationError("test_item_id", "must be unique within the test")
	}
	if item.Position().Int() > len(test.items) {
		return NewValidationError("position", "must be less than or equal to item count")
	}

	items := test.Items()
	for i := range items {
		if items[i].Position().Int() >= item.Position().Int() {
			items[i].MoveTo(TestItemPosition{value: items[i].Position().Int() + 1})
		}
	}
	items = append(items, item)
	test.items = sortTestItemsByPosition(items)
	test.touch(now)

	return nil
}

func (test *Test) RemoveItem(id TestItemID, now time.Time) error {
	items := make([]TestItem, 0, len(test.items))
	removed := false

	for _, item := range test.items {
		if item.ID() == id {
			removed = true
			continue
		}

		item.MoveTo(TestItemPosition{value: len(items)})
		items = append(items, item)
	}

	if !removed {
		return ErrNotFound
	}

	test.items = items
	test.touch(now)

	return nil
}

func (test *Test) ReorderItems(order []TestItemPlacement, now time.Time) error {
	if len(order) != len(test.items) {
		return NewValidationError("order", "must include every item exactly once")
	}

	current := make(map[string]TestItem, len(test.items))
	for _, item := range test.items {
		current[item.ID().String()] = item
	}

	usedItems := make(map[string]struct{}, len(order))
	usedPositions := make(map[int]struct{}, len(order))
	positions := make(map[string]TestItemPosition, len(order))

	for _, placement := range order {
		id := placement.TestItemID.String()
		if _, exists := current[id]; !exists {
			return NewValidationError("test_item_id", "must belong to the test")
		}
		if _, exists := usedItems[id]; exists {
			return NewValidationError("test_item_id", "must be unique")
		}

		position := placement.Position.Int()
		if position >= len(order) {
			return NewValidationError("position", "must be contiguous from zero")
		}
		if _, exists := usedPositions[position]; exists {
			return NewValidationError("position", "must be unique")
		}

		usedItems[id] = struct{}{}
		usedPositions[position] = struct{}{}
		positions[id] = placement.Position
	}

	items := make([]TestItem, 0, len(test.items))
	for _, item := range test.items {
		item.MoveTo(positions[item.ID().String()])
		items = append(items, item)
	}

	test.items = sortTestItemsByPosition(items)
	test.touch(now)

	return nil
}

func (test *Test) ReplaceItemBody(id TestItemID, body TestItemBody, now time.Time) error {
	items := test.Items()
	for i := range items {
		if items[i].ID() != id {
			continue
		}
		if err := items[i].ReplaceBody(body); err != nil {
			return err
		}

		test.items = items
		test.touch(now)
		return nil
	}

	return ErrNotFound
}

func (test *Test) touch(now time.Time) {
	test.updatedAt = mutationTime(test.createdAt, now)
}

func (test Test) hasItemID(id TestItemID) bool {
	for _, item := range test.items {
		if item.ID() == id {
			return true
		}
	}

	return false
}

type TestItemPlacement struct {
	TestItemID TestItemID
	Position   TestItemPosition
}

func validateTestItemKind(kind TestItemKind) error {
	switch kind {
	case ChoiceKind(), CodingKind():
		return nil
	default:
		return NewValidationError("kind", "must be choice or coding")
	}
}

func validateTestItemBody(body TestItemBody) error {
	switch typed := body.(type) {
	case ChoiceItemBody:
		return typed.validate()
	case CodingItemBody:
		return typed.validate()
	default:
		return NewValidationError("body", "must be a supported test item body")
	}
}

func validateMediaRef(field string, ref MediaRef) error {
	if ref.Locator() == "" {
		return NewValidationError(field, "must not be empty")
	}
	if !isValidMediaLocator(ref.Provider(), ref.Locator()) {
		return NewValidationError(field, "must be a valid media ref")
	}

	return nil
}

func normalizeTestItems(items []TestItem) ([]TestItem, error) {
	normalized := make([]TestItem, len(items))
	copy(normalized, items)

	ids := make(map[string]struct{}, len(normalized))
	positions := make(map[int]struct{}, len(normalized))
	for _, item := range normalized {
		if err := item.validate(); err != nil {
			return nil, err
		}

		id := item.ID().String()
		if _, exists := ids[id]; exists {
			return nil, NewValidationError("test_item_id", "must be unique within the test")
		}
		ids[id] = struct{}{}

		position := item.Position().Int()
		if _, exists := positions[position]; exists {
			return nil, NewValidationError("position", "must be unique")
		}
		positions[position] = struct{}{}
	}

	normalized = sortTestItemsByPosition(normalized)
	for i, item := range normalized {
		if item.Position().Int() != i {
			return nil, NewValidationError("position", "must be contiguous from zero")
		}
	}

	return normalized, nil
}

func sortTestItemsByPosition(items []TestItem) []TestItem {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Position().Int() < items[j].Position().Int()
	})

	return items
}

func copyTimeLimit(limit *TimeLimit) *TimeLimit {
	if limit == nil {
		return nil
	}

	copied := *limit
	return &copied
}

func copyTestSolution(solution *TestSolution) *TestSolution {
	if solution == nil {
		return nil
	}

	copied := *solution
	return &copied
}

func copyCodingTestCases(testCases []CodingTestCase) []CodingTestCase {
	copied := make([]CodingTestCase, len(testCases))
	copy(copied, testCases)

	return copied
}
