package core

import "time"

type CreateQuizInput struct {
	CourseID      string
	Title         string
	PassThreshold *float64
}

type CreateQuizOutput struct {
	ID string
}

type ListQuizzesInput struct {
	CourseID string
}

type ListQuizzesOutput struct {
	Quizzes []QuizView
}

type GetQuizInput struct {
	ID string
}

type GetQuizOutput struct {
	Quiz QuizDetailView
}

type UpdateQuizInput struct {
	ID            string
	Title         *string
	PassThreshold *float64
}

type UpdateQuizOutput struct {
	ID string
}

type DeleteQuizInput struct {
	ID string
}

type AddQuestionInput struct {
	QuizID         string
	Type           string
	Prompt         string
	Options        []string
	CorrectIndices []int
	Explanation    string
	Position       *int
}

type AddQuestionOutput struct {
	ID string
}

type ListQuestionsInput struct {
	QuizID string
}

type ListQuestionsOutput struct {
	Questions []QuestionView
}

type GetQuestionInput struct {
	ID string
}

type GetQuestionOutput struct {
	Question QuestionView
}

type UpdateQuestionInput struct {
	ID             string
	Prompt         *string
	Options        *[]string
	CorrectIndices *[]int
	Explanation    *string
}

type UpdateQuestionOutput struct {
	ID string
}

type RemoveQuestionInput struct {
	ID string
}

type ReorderQuestionsInput struct {
	QuizID string
	Order  []QuestionPlacementDTO
}

type QuestionPlacementDTO struct {
	QuestionID string
	Position   int
}

type QuizView struct {
	ID            string
	CourseID      string
	Title         string
	PassThreshold float64
	QuestionCount int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type QuizDetailView struct {
	QuizView
	Questions []QuestionView
}

type QuestionView struct {
	ID             string
	QuizID         string
	Type           string
	Prompt         string
	Explanation    string
	Options        []string
	CorrectIndices []int
	Position       int
}
