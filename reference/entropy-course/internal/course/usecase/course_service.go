package usecase

import (
	"errors"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

type CourseService struct {
	courses   core.CourseRepository
	lessons   core.LessonRepository
	quizzes   core.QuizRepository
	practices core.PracticeRepository
	tests     core.TestRepository
	ids       core.IDGenerator
	clock     core.Clock
}

func NewCourseService(
	courses core.CourseRepository,
	lessons core.LessonRepository,
	quizzes core.QuizRepository,
	ids core.IDGenerator,
	clock core.Clock,
	practices core.PracticeRepository,
	tests ...core.TestRepository,
) *CourseService {
	var testRepo core.TestRepository
	if len(tests) > 0 {
		testRepo = tests[0]
	}

	return &CourseService{
		courses:   courses,
		lessons:   lessons,
		quizzes:   quizzes,
		practices: practices,
		tests:     testRepo,
		ids:       ids,
		clock:     clock,
	}
}

func (service *CourseService) CreateCourse(in core.CreateCourseInput) (core.CreateCourseOutput, error) {
	slug, err := domain.NewSlug(in.Slug)
	if err != nil {
		return core.CreateCourseOutput{}, err
	}

	instructorID, err := domain.NewInstructorID(in.InstructorID)
	if err != nil {
		return core.CreateCourseOutput{}, err
	}

	if err := service.ensureSlugAvailable(slug, nil); err != nil {
		return core.CreateCourseOutput{}, err
	}

	id := service.ids.NewCourseID()
	course, err := domain.NewCourse(id, in.Title, slug, in.Description, instructorID, service.clock.Now())
	if err != nil {
		return core.CreateCourseOutput{}, err
	}

	if err := service.courses.Save(course); err != nil {
		return core.CreateCourseOutput{}, err
	}

	return core.CreateCourseOutput{ID: id.String()}, nil
}

func (service *CourseService) ListCourses(in core.ListCoursesInput) (core.ListCoursesOutput, error) {
	filter, err := buildCourseFilter(in.Status)
	if err != nil {
		return core.ListCoursesOutput{}, err
	}

	courses, err := service.courses.FindAll(filter)
	if err != nil {
		return core.ListCoursesOutput{}, err
	}

	views := make([]core.CourseView, 0, len(courses))
	for _, course := range courses {
		views = append(views, courseView(course))
	}

	return core.ListCoursesOutput{Courses: views}, nil
}

func (service *CourseService) GetCourse(in core.GetCourseInput) (core.GetCourseOutput, error) {
	id, err := domain.NewCourseID(in.ID)
	if err != nil {
		return core.GetCourseOutput{}, err
	}

	course, err := service.courses.FindByID(id)
	if err != nil {
		return core.GetCourseOutput{}, err
	}

	return core.GetCourseOutput{Course: courseView(course)}, nil
}

func (service *CourseService) UpdateCourse(in core.UpdateCourseInput) (core.UpdateCourseOutput, error) {
	id, err := domain.NewCourseID(in.ID)
	if err != nil {
		return core.UpdateCourseOutput{}, err
	}

	if in.Title == nil && in.Description == nil && in.Slug == nil {
		return core.UpdateCourseOutput{}, domain.NewValidationError("update", "must include at least one field")
	}

	course, err := service.courses.FindByID(id)
	if err != nil {
		return core.UpdateCourseOutput{}, err
	}

	now := service.clock.Now()
	if in.Slug != nil {
		slug, err := domain.NewSlug(*in.Slug)
		if err != nil {
			return core.UpdateCourseOutput{}, err
		}

		if err := service.ensureSlugAvailable(slug, &id); err != nil {
			return core.UpdateCourseOutput{}, err
		}

		course.ChangeSlug(slug, now)
	}

	if in.Title != nil {
		if err := course.Rename(*in.Title, now); err != nil {
			return core.UpdateCourseOutput{}, err
		}
	}

	if in.Description != nil {
		course.ChangeDescription(*in.Description, now)
	}

	if err := service.courses.Save(course); err != nil {
		return core.UpdateCourseOutput{}, err
	}

	return core.UpdateCourseOutput{ID: id.String()}, nil
}

func (service *CourseService) DeleteCourse(in core.DeleteCourseInput) error {
	id, err := domain.NewCourseID(in.ID)
	if err != nil {
		return err
	}

	if _, err := service.courses.FindByID(id); err != nil {
		return err
	}

	if err := service.lessons.DeleteByCourse(id); err != nil {
		return err
	}

	if err := service.quizzes.DeleteByCourse(id); err != nil {
		return err
	}

	if service.practices != nil {
		if err := service.practices.DeleteByCourse(id); err != nil {
			return err
		}
	}

	if service.tests != nil {
		if err := service.tests.DeleteByCourse(id); err != nil {
			return err
		}
	}

	return service.courses.Delete(id)
}

func (service *CourseService) PublishCourse(in core.PublishCourseInput) error {
	id, err := domain.NewCourseID(in.ID)
	if err != nil {
		return err
	}

	course, err := service.courses.FindByID(id)
	if err != nil {
		return err
	}

	if err := course.Publish(service.clock.Now()); err != nil {
		return err
	}

	return service.courses.Save(course)
}

func (service *CourseService) UnpublishCourse(in core.UnpublishCourseInput) error {
	id, err := domain.NewCourseID(in.ID)
	if err != nil {
		return err
	}

	course, err := service.courses.FindByID(id)
	if err != nil {
		return err
	}

	if err := course.Unpublish(service.clock.Now()); err != nil {
		return err
	}

	return service.courses.Save(course)
}

func (service *CourseService) ensureSlugAvailable(slug domain.Slug, currentCourseID *domain.CourseID) error {
	course, err := service.courses.FindBySlug(slug)
	if errors.Is(err, domain.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	if currentCourseID != nil && course.ID() == *currentCourseID {
		return nil
	}

	return domain.ErrSlugTaken
}

func buildCourseFilter(statusValue string) (core.CourseFilter, error) {
	if statusValue == "" {
		return core.CourseFilter{}, nil
	}

	status, err := domain.NewCourseStatus(statusValue)
	if err != nil {
		return core.CourseFilter{}, err
	}

	return core.CourseFilter{Status: &status}, nil
}

func courseView(course domain.Course) core.CourseView {
	return core.CourseView{
		ID:           course.ID().String(),
		Title:        course.Title(),
		Slug:         course.Slug().String(),
		Description:  course.Description(),
		InstructorID: course.InstructorID().String(),
		Status:       course.Status().String(),
		CreatedAt:    course.CreatedAt(),
		UpdatedAt:    course.UpdatedAt(),
	}
}
