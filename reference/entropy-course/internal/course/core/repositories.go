package core

import "github.com/luxeave/entropy-course/internal/course/domain"

type CourseRepository interface {
	Save(course domain.Course) error
	FindByID(id domain.CourseID) (domain.Course, error)
	FindBySlug(slug domain.Slug) (domain.Course, error)
	FindAll(filter CourseFilter) ([]domain.Course, error)
	Delete(id domain.CourseID) error
}

type CourseFilter struct {
	Status *domain.CourseStatus
}

type LessonRepository interface {
	Save(lesson domain.Lesson) error
	SaveAll(lessons []domain.Lesson) error
	FindByID(id domain.LessonID) (domain.Lesson, error)
	FindByCourse(courseID domain.CourseID) ([]domain.Lesson, error)
	FindByBlockID(id domain.BlockID) (domain.Lesson, error)
	FindLessonsEmbeddingQuiz(quizID domain.QuizID) ([]domain.LessonID, error)
	FindLessonsEmbeddingPractice(practiceID domain.PracticeID) ([]domain.LessonID, error)
	Delete(id domain.LessonID) error
	DeleteByCourse(courseID domain.CourseID) error
}

type QuizRepository interface {
	Save(quiz domain.Quiz) error
	FindByID(id domain.QuizID) (domain.Quiz, error)
	FindByCourse(courseID domain.CourseID) ([]domain.Quiz, error)
	FindByQuestionID(id domain.QuestionID) (domain.Quiz, error)
	Delete(id domain.QuizID) error
	DeleteByCourse(courseID domain.CourseID) error
}

type PracticeRepository interface {
	Save(practice domain.Practice) error
	FindByID(id domain.PracticeID) (domain.Practice, error)
	FindByCourse(courseID domain.CourseID) ([]domain.Practice, error)
	FindByTestCaseID(id domain.TestCaseID) (domain.Practice, error)
	Delete(id domain.PracticeID) error
	DeleteByCourse(courseID domain.CourseID) error
}

type TestRepository interface {
	Save(test domain.Test) error
	FindByID(id domain.TestID) (domain.Test, error)
	FindByCourse(courseID domain.CourseID) ([]domain.Test, error)
	FindByItemID(id domain.TestItemID) (domain.Test, error)
	Delete(id domain.TestID) error
	DeleteByCourse(courseID domain.CourseID) error
}
