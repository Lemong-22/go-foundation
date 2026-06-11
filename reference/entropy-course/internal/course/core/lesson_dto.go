package core

import "time"

type CreateLessonInput struct {
	CourseID string
	Title    string
	Order    *int
}

type CreateLessonOutput struct {
	ID string
}

type ListLessonsInput struct {
	CourseID string
}

type ListLessonsOutput struct {
	Lessons []LessonView
}

type GetLessonInput struct {
	ID string
}

type GetLessonOutput struct {
	Lesson LessonView
}

type UpdateLessonInput struct {
	ID    string
	Title *string
}

type UpdateLessonOutput struct {
	ID string
}

type DeleteLessonInput struct {
	ID string
}

type ReorderLessonsInput struct {
	CourseID string
	Order    []LessonPosition
}

type LessonPosition struct {
	LessonID string
	Position int
}

type LessonView struct {
	ID        string
	CourseID  string
	Title     string
	Order     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type AddLessonBlockInput struct {
	LessonID      string
	Kind          string
	Markdown      string
	VideoProvider string
	VideoLocator  string
	VideoCaption  string
	QuizRef       string
	PracticeRef   string
	Position      *int
}

type AddLessonBlockOutput struct {
	ID string
}

type ListLessonBlocksInput struct {
	LessonID string
}

type ListLessonBlocksOutput struct {
	Blocks []BlockView
}

type GetLessonBlockInput struct {
	ID string
}

type GetLessonBlockOutput struct {
	Block BlockView
}

type UpdateLessonBlockInput struct {
	ID            string
	Markdown      *string
	VideoProvider *string
	VideoLocator  *string
	VideoCaption  *string
}

type UpdateLessonBlockOutput struct {
	ID string
}

type RemoveLessonBlockInput struct {
	ID string
}

type ReorderLessonBlocksInput struct {
	LessonID string
	Order    []BlockPlacementDTO
}

type BlockPlacementDTO struct {
	BlockID  string
	Position int
}

type BlockView struct {
	ID            string
	LessonID      string
	Kind          string
	Position      int
	Markdown      string
	VideoProvider string
	VideoLocator  string
	VideoCaption  string
	QuizRef       string
	PracticeRef   string
}
