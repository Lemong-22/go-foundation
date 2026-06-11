package core

import "time"

type CreatePracticeInput struct {
	CourseID    string
	Title       string
	Language    string
	Prompt      string
	StarterCode string
	Solution    string
}

type CreatePracticeOutput struct {
	ID string
}

type ListPracticesInput struct {
	CourseID string
}

type ListPracticesOutput struct {
	Practices []PracticeView
}

type GetPracticeInput struct {
	ID string
}

type GetPracticeOutput struct {
	Practice PracticeDetailView
}

type UpdatePracticeInput struct {
	ID          string
	Title       *string
	Prompt      *string
	StarterCode *string
	Solution    *string
}

type UpdatePracticeOutput struct {
	ID string
}

type DeletePracticeInput struct {
	ID string
}

type AddTestCaseInput struct {
	PracticeID     string
	Stdin          string
	ExpectedStdout string
	Name           string
	Position       *int
}

type AddTestCaseOutput struct {
	ID string
}

type ListTestCasesInput struct {
	PracticeID string
}

type ListTestCasesOutput struct {
	TestCases []TestCaseView
}

type GetTestCaseInput struct {
	ID string
}

type GetTestCaseOutput struct {
	TestCase TestCaseView
}

type UpdateTestCaseInput struct {
	ID             string
	Stdin          *string
	ExpectedStdout *string
	Name           *string
}

type UpdateTestCaseOutput struct {
	ID string
}

type RemoveTestCaseInput struct {
	ID string
}

type ReorderTestCasesInput struct {
	PracticeID string
	Order      []TestCasePlacementDTO
}

type TestCasePlacementDTO struct {
	TestCaseID string
	Position   int
}

type PracticeView struct {
	ID            string
	CourseID      string
	Title         string
	Language      string
	TestCaseCount int
	HasSolution   bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PracticeDetailView struct {
	PracticeView
	Prompt      string
	StarterCode string
	Solution    string
	TestCases   []TestCaseView
}

type TestCaseView struct {
	ID             string
	PracticeID     string
	Stdin          string
	ExpectedStdout string
	Name           string
	Position       int
}
