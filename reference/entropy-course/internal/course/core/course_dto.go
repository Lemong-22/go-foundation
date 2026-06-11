package core

import "time"

type CreateCourseInput struct {
	Title        string
	Slug         string
	Description  string
	InstructorID string
}

type CreateCourseOutput struct {
	ID string
}

type ListCoursesInput struct {
	Status string
}

type ListCoursesOutput struct {
	Courses []CourseView
}

type GetCourseInput struct {
	ID string
}

type GetCourseOutput struct {
	Course CourseView
}

type UpdateCourseInput struct {
	ID          string
	Title       *string
	Description *string
	Slug        *string
}

type UpdateCourseOutput struct {
	ID string
}

type DeleteCourseInput struct {
	ID string
}

type PublishCourseInput struct {
	ID string
}

type UnpublishCourseInput struct {
	ID string
}

type CourseView struct {
	ID           string
	Title        string
	Slug         string
	Description  string
	InstructorID string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
