package core

type ImportSource interface {
	Open(zipPath string) (ParsedImportSource, ImportSourceMetadata, error)
}

type ImportSourceMetadata struct {
	ZipHash       string
	FormatVersion string
}

type ParsedImportSource struct {
	FormatVersion string
	Course        ParsedCourse
	Lessons       []ParsedLesson
	Quizzes       []ParsedQuiz
	Practices     []ParsedPractice
	Tests         []ParsedTest
}

type ParsedCourse struct {
	Title       string
	Slug        string
	Description string
	Status      string
}

type ParsedLesson struct {
	Title  string
	Order  *int
	Blocks []ParsedLessonBlock
}

type ParsedLessonBlock struct {
	Kind          string
	Markdown      string
	VideoProvider string
	VideoLocator  string
	VideoCaption  string
	QuizRef       string
	PracticeRef   string
	Position      *int
}

type ParsedQuiz struct {
	Slug          string
	Title         string
	PassThreshold *float64
	Questions     []ParsedQuestion
}

type ParsedQuestion struct {
	Type           string
	Prompt         string
	Options        []string
	CorrectIndices []int
	Explanation    string
	Position       *int
}

type ParsedPractice struct {
	Slug        string
	Title       string
	Language    string
	Prompt      string
	StarterCode string
	Solution    string
	TestCases   []ParsedPracticeTestCase
}

type ParsedPracticeTestCase struct {
	Stdin          string
	ExpectedStdout string
	Name           string
	Position       *int
}

type ParsedTest struct {
	Slug             string
	Title            string
	TimeLimitMinutes *int
	PassThreshold    *float64
	Solution         *ParsedTestSolution
	Items            []ParsedTestItem
}

type ParsedTestSolution struct {
	ZipProvider   string
	ZipLocator    string
	VideoProvider string
	VideoLocator  string
	VideoCaption  string
}

type ParsedTestItem struct {
	Kind     string
	Position *int

	Prompt         string
	ChoiceType     string
	Options        []string
	CorrectIndices []int
	Explanation    string

	CodingPrompt string
	Language     string
	StarterCode  string
	Solution     string
	TestCases    []ParsedCodingTestCase
}

type ParsedCodingTestCase struct {
	Stdin          string
	ExpectedStdout string
	Name           string
}
