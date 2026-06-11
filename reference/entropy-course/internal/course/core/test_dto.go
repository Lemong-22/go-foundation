package core

import "time"

type CreateTestInput struct {
	CourseID         string
	Title            string
	TimeLimitMinutes *int
	PassThreshold    *float64
}

type CreateTestOutput struct {
	ID string
}

type ListTestsInput struct {
	CourseID string
}

type ListTestsOutput struct {
	Tests []TestView
}

type GetTestInput struct {
	ID string
}

type GetTestOutput struct {
	Test TestDetailView
}

type UpdateTestInput struct {
	ID               string
	Title            *string
	TimeLimitMinutes *int
	PassThreshold    *float64

	SolutionZipProvider   *string
	SolutionZipLocator    *string
	SolutionVideoProvider *string
	SolutionVideoLocator  *string
	SolutionVideoCaption  *string
}

type UpdateTestOutput struct {
	ID string
}

type DeleteTestInput struct {
	ID string
}

type AddTestItemInput struct {
	TestID   string
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
	TestCases    []CodingTestCaseDTO
}

type AddTestItemOutput struct {
	ID string
}

type ListTestItemsInput struct {
	TestID string
}

type ListTestItemsOutput struct {
	Items []TestItemView
}

type GetTestItemInput struct {
	ID string
}

type GetTestItemOutput struct {
	Item TestItemView
}

type UpdateTestItemInput struct {
	ID string

	Prompt         *string
	ChoiceType     *string
	Options        *[]string
	CorrectIndices *[]int
	Explanation    *string

	CodingPrompt *string
	Language     *string
	StarterCode  *string
	Solution     *string
	TestCases    *[]CodingTestCaseDTO
}

type UpdateTestItemOutput struct {
	ID string
}

type RemoveTestItemInput struct {
	ID string
}

type ReorderTestItemsInput struct {
	TestID string
	Order  []TestItemPlacementDTO
}

type TestItemPlacementDTO struct {
	TestItemID string
	Position   int
}

type TestView struct {
	ID               string
	CourseID         string
	Title            string
	TimeLimitMinutes *int
	PassThreshold    float64
	HasSolution      bool
	ItemCount        int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type TestDetailView struct {
	TestView
	Solution *TestSolutionView
	Items    []TestItemView
}

type TestSolutionView struct {
	ZipProvider   string
	ZipLocator    string
	VideoProvider string
	VideoLocator  string
	VideoCaption  string
}

type TestItemView struct {
	ID       string
	TestID   string
	Kind     string
	Position int

	ChoicePrompt         string
	ChoiceType           string
	ChoiceOptions        []string
	ChoiceCorrectIndices []int
	ChoiceExplanation    string

	CodingPrompt   string
	Language       string
	StarterCode    string
	CodingSolution string
	TestCases      []CodingTestCaseDTO
}

type CodingTestCaseDTO struct {
	Stdin          string
	ExpectedStdout string
	Name           string
}
