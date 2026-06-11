package domain

import (
	"sort"
	"strings"
	"time"
)

const (
	languageJavaScript = "javascript"
	languageTypeScript = "typescript"
	languageGolang     = "golang"
	languageRust       = "rust"
)

type Language struct {
	value string
}

func JavaScript() Language {
	return Language{value: languageJavaScript}
}

func TypeScript() Language {
	return Language{value: languageTypeScript}
}

func Golang() Language {
	return Language{value: languageGolang}
}

func Rust() Language {
	return Language{value: languageRust}
}

func NewLanguage(value string) (Language, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case languageJavaScript:
		return JavaScript(), nil
	case languageTypeScript:
		return TypeScript(), nil
	case languageGolang:
		return Golang(), nil
	case languageRust:
		return Rust(), nil
	default:
		return Language{}, NewValidationError("language", "must be javascript, typescript, golang, or rust")
	}
}

func (language Language) String() string {
	return language.value
}

func (language Language) IsJavaScript() bool {
	return language.value == languageJavaScript
}

func (language Language) IsTypeScript() bool {
	return language.value == languageTypeScript
}

func (language Language) IsGolang() bool {
	return language.value == languageGolang
}

func (language Language) IsRust() bool {
	return language.value == languageRust
}

type TestCasePosition struct {
	value int
}

func NewTestCasePosition(value int) (TestCasePosition, error) {
	if value < 0 {
		return TestCasePosition{}, NewValidationError("position", "must be greater than or equal to zero")
	}

	return TestCasePosition{value: value}, nil
}

func (position TestCasePosition) Int() int {
	return position.value
}

type TestCase struct {
	id             TestCaseID
	stdin          string
	expectedStdout string
	name           string
	position       TestCasePosition
}

func NewTestCase(
	id TestCaseID,
	stdin string,
	expectedStdout string,
	name string,
	position TestCasePosition,
) (TestCase, error) {
	if err := validateTestCaseIdentity(id); err != nil {
		return TestCase{}, err
	}
	if position.Int() < 0 {
		return TestCase{}, NewValidationError("position", "must be greater than or equal to zero")
	}

	return TestCase{
		id:             id,
		stdin:          stdin,
		expectedStdout: expectedStdout,
		name:           name,
		position:       position,
	}, nil
}

func (testCase TestCase) ID() TestCaseID {
	return testCase.id
}

func (testCase TestCase) Stdin() string {
	return testCase.stdin
}

func (testCase TestCase) ExpectedStdout() string {
	return testCase.expectedStdout
}

func (testCase TestCase) Name() string {
	return testCase.name
}

func (testCase TestCase) Position() TestCasePosition {
	return testCase.position
}

func (testCase *TestCase) ChangeStdin(stdin string) {
	testCase.stdin = stdin
}

func (testCase *TestCase) ChangeExpectedStdout(expected string) {
	testCase.expectedStdout = expected
}

func (testCase *TestCase) ChangeName(name string) {
	testCase.name = name
}

func (testCase *TestCase) MoveTo(position TestCasePosition) {
	testCase.position = position
}

func (testCase TestCase) validate() error {
	if err := validateTestCaseIdentity(testCase.id); err != nil {
		return err
	}
	if testCase.position.Int() < 0 {
		return NewValidationError("position", "must be greater than or equal to zero")
	}

	return nil
}

type Practice struct {
	id          PracticeID
	courseID    CourseID
	title       string
	language    Language
	prompt      string
	starterCode string
	solution    string
	testCases   []TestCase
	createdAt   time.Time
	updatedAt   time.Time
}

func NewPractice(
	id PracticeID,
	courseID CourseID,
	title string,
	language Language,
	prompt string,
	starterCode string,
	solution string,
	testCases []TestCase,
	now time.Time,
) (Practice, error) {
	return RestorePractice(id, courseID, title, language, prompt, starterCode, solution, testCases, now, now)
}

func RestorePractice(
	id PracticeID,
	courseID CourseID,
	title string,
	language Language,
	prompt string,
	starterCode string,
	solution string,
	testCases []TestCase,
	createdAt time.Time,
	updatedAt time.Time,
) (Practice, error) {
	if id.String() == "" {
		return Practice{}, NewValidationError("practice_id", "must not be empty")
	}
	if courseID.String() == "" {
		return Practice{}, NewValidationError("course_id", "must not be empty")
	}
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return Practice{}, err
	}
	if err := validateLanguage(language); err != nil {
		return Practice{}, err
	}
	normalizedPrompt, err := normalizePracticePrompt(prompt)
	if err != nil {
		return Practice{}, err
	}
	if updatedAt.Before(createdAt) {
		return Practice{}, NewValidationError("updated_at", "must be greater than or equal to created_at")
	}

	normalizedTestCases, err := normalizePracticeTestCases(testCases)
	if err != nil {
		return Practice{}, err
	}

	return Practice{
		id:          id,
		courseID:    courseID,
		title:       normalizedTitle,
		language:    language,
		prompt:      normalizedPrompt,
		starterCode: starterCode,
		solution:    solution,
		testCases:   normalizedTestCases,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

func (practice Practice) ID() PracticeID {
	return practice.id
}

func (practice Practice) CourseID() CourseID {
	return practice.courseID
}

func (practice Practice) Title() string {
	return practice.title
}

func (practice Practice) Language() Language {
	return practice.language
}

func (practice Practice) Prompt() string {
	return practice.prompt
}

func (practice Practice) StarterCode() string {
	return practice.starterCode
}

func (practice Practice) Solution() string {
	return practice.solution
}

func (practice Practice) TestCases() []TestCase {
	testCases := make([]TestCase, len(practice.testCases))
	copy(testCases, practice.testCases)

	return testCases
}

func (practice Practice) TestCase(id TestCaseID) (TestCase, error) {
	for _, testCase := range practice.testCases {
		if testCase.ID() == id {
			return testCase, nil
		}
	}

	return TestCase{}, ErrNotFound
}

func (practice Practice) CreatedAt() time.Time {
	return practice.createdAt
}

func (practice Practice) UpdatedAt() time.Time {
	return practice.updatedAt
}

func (practice *Practice) Rename(title string, now time.Time) error {
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return err
	}

	practice.title = normalizedTitle
	practice.touch(now)

	return nil
}

func (practice *Practice) ChangePrompt(prompt string, now time.Time) error {
	normalizedPrompt, err := normalizePracticePrompt(prompt)
	if err != nil {
		return err
	}

	practice.prompt = normalizedPrompt
	practice.touch(now)

	return nil
}

func (practice *Practice) ChangeStarterCode(code string, now time.Time) {
	practice.starterCode = code
	practice.touch(now)
}

func (practice *Practice) ChangeSolution(source string, now time.Time) {
	practice.solution = source
	practice.touch(now)
}

func (practice *Practice) AddTestCase(testCase TestCase, now time.Time) error {
	if err := testCase.validate(); err != nil {
		return err
	}
	if practice.hasTestCaseID(testCase.ID()) {
		return NewValidationError("test_case_id", "must be unique within the practice")
	}
	if testCase.Position().Int() > len(practice.testCases) {
		return NewValidationError("position", "must be less than or equal to test case count")
	}

	testCases := practice.TestCases()
	for i := range testCases {
		if testCases[i].Position().Int() >= testCase.Position().Int() {
			testCases[i].MoveTo(TestCasePosition{value: testCases[i].Position().Int() + 1})
		}
	}
	testCases = append(testCases, testCase)
	practice.testCases = sortTestCasesByPosition(testCases)
	practice.touch(now)

	return nil
}

func (practice *Practice) RemoveTestCase(id TestCaseID, now time.Time) error {
	testCases := make([]TestCase, 0, len(practice.testCases))
	removed := false

	for _, testCase := range practice.testCases {
		if testCase.ID() == id {
			removed = true
			continue
		}

		testCase.MoveTo(TestCasePosition{value: len(testCases)})
		testCases = append(testCases, testCase)
	}

	if !removed {
		return ErrNotFound
	}

	practice.testCases = testCases
	practice.touch(now)

	return nil
}

func (practice *Practice) ReorderTestCases(order []TestCasePlacement, now time.Time) error {
	if len(order) != len(practice.testCases) {
		return NewValidationError("order", "must include every test case exactly once")
	}

	current := make(map[string]TestCase, len(practice.testCases))
	for _, testCase := range practice.testCases {
		current[testCase.ID().String()] = testCase
	}

	usedTestCases := make(map[string]struct{}, len(order))
	usedPositions := make(map[int]struct{}, len(order))
	positions := make(map[string]TestCasePosition, len(order))

	for _, placement := range order {
		id := placement.TestCaseID.String()
		if _, exists := current[id]; !exists {
			return NewValidationError("test_case_id", "must belong to the practice")
		}
		if _, exists := usedTestCases[id]; exists {
			return NewValidationError("test_case_id", "must be unique")
		}

		position := placement.Position.Int()
		if position < 0 || position >= len(order) {
			return NewValidationError("position", "must be contiguous from zero")
		}
		if _, exists := usedPositions[position]; exists {
			return NewValidationError("position", "must be unique")
		}

		usedTestCases[id] = struct{}{}
		usedPositions[position] = struct{}{}
		positions[id] = placement.Position
	}

	testCases := make([]TestCase, 0, len(practice.testCases))
	for _, testCase := range practice.testCases {
		testCase.MoveTo(positions[testCase.ID().String()])
		testCases = append(testCases, testCase)
	}

	practice.testCases = sortTestCasesByPosition(testCases)
	practice.touch(now)

	return nil
}

func (practice *Practice) ChangeTestCaseStdin(id TestCaseID, stdin string, now time.Time) error {
	testCases := practice.TestCases()
	for i := range testCases {
		if testCases[i].ID() != id {
			continue
		}

		testCases[i].ChangeStdin(stdin)
		practice.testCases = testCases
		practice.touch(now)
		return nil
	}

	return ErrNotFound
}

func (practice *Practice) ChangeTestCaseExpectedStdout(id TestCaseID, expected string, now time.Time) error {
	testCases := practice.TestCases()
	for i := range testCases {
		if testCases[i].ID() != id {
			continue
		}

		testCases[i].ChangeExpectedStdout(expected)
		practice.testCases = testCases
		practice.touch(now)
		return nil
	}

	return ErrNotFound
}

func (practice *Practice) ChangeTestCaseName(id TestCaseID, name string, now time.Time) error {
	testCases := practice.TestCases()
	for i := range testCases {
		if testCases[i].ID() != id {
			continue
		}

		testCases[i].ChangeName(name)
		practice.testCases = testCases
		practice.touch(now)
		return nil
	}

	return ErrNotFound
}

func (practice *Practice) touch(now time.Time) {
	practice.updatedAt = mutationTime(practice.createdAt, now)
}

func (practice Practice) hasTestCaseID(id TestCaseID) bool {
	for _, testCase := range practice.testCases {
		if testCase.ID() == id {
			return true
		}
	}

	return false
}

type TestCasePlacement struct {
	TestCaseID TestCaseID
	Position   TestCasePosition
}

func normalizePracticePrompt(prompt string) (string, error) {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "", NewValidationError("prompt", "must not be empty")
	}

	return trimmed, nil
}

func validateLanguage(language Language) error {
	switch language {
	case JavaScript(), TypeScript(), Golang(), Rust():
		return nil
	default:
		return NewValidationError("language", "must be javascript, typescript, golang, or rust")
	}
}

func validateTestCaseIdentity(id TestCaseID) error {
	if id.String() == "" {
		return NewValidationError("test_case_id", "must not be empty")
	}

	return nil
}

func normalizePracticeTestCases(testCases []TestCase) ([]TestCase, error) {
	normalized := make([]TestCase, len(testCases))
	copy(normalized, testCases)

	ids := make(map[string]struct{}, len(normalized))
	positions := make(map[int]struct{}, len(normalized))
	for _, testCase := range normalized {
		if err := testCase.validate(); err != nil {
			return nil, err
		}

		id := testCase.ID().String()
		if _, exists := ids[id]; exists {
			return nil, NewValidationError("test_case_id", "must be unique within the practice")
		}
		ids[id] = struct{}{}

		position := testCase.Position().Int()
		if _, exists := positions[position]; exists {
			return nil, NewValidationError("position", "must be unique")
		}
		positions[position] = struct{}{}
	}

	normalized = sortTestCasesByPosition(normalized)
	for i, testCase := range normalized {
		if testCase.Position().Int() != i {
			return nil, NewValidationError("position", "must be contiguous from zero")
		}
	}

	return normalized, nil
}

func sortTestCasesByPosition(testCases []TestCase) []TestCase {
	sort.Slice(testCases, func(i, j int) bool {
		return testCases[i].Position().Int() < testCases[j].Position().Int()
	})

	return testCases
}
