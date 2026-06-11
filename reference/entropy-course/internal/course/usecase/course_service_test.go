package usecase

import (
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	courseIDValue     = "550e8400-e29b-41d4-a716-446655440000"
	otherCourseID     = "550e8400-e29b-41d4-a716-446655440001"
	instructorIDValue = "550e8400-e29b-41d4-a716-446655440010"
)

func TestCreateCourseSavesDraftCourse(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 8, 0, 0, 0, time.UTC)}
	courses := newCourseRepositoryFake()
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, clock)

	out, err := service.CreateCourse(core.CreateCourseInput{
		Title:        "Intro to Go",
		Slug:         "intro-to-go",
		Description:  "Learn Go",
		InstructorID: instructorIDValue,
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	if out.ID != courseIDValue {
		t.Fatalf("expected id %q, got %q", courseIDValue, out.ID)
	}

	saved := courses.savedCourses[0]
	if saved.Title() != "Intro to Go" || saved.Status() != domain.Draft() {
		t.Fatalf("expected saved draft course, got title=%q status=%q", saved.Title(), saved.Status())
	}
	if saved.Slug().String() != "intro-to-go" || saved.InstructorID().String() != instructorIDValue {
		t.Fatalf("expected slug and instructor id to be saved")
	}
	if !saved.CreatedAt().Equal(clock.now) || !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected deterministic timestamps")
	}
}

func TestCreateCourseValidatesInputs(t *testing.T) {
	tests := []struct {
		name  string
		input core.CreateCourseInput
	}{
		{
			name: "invalid slug",
			input: core.CreateCourseInput{
				Title:        "Intro to Go",
				Slug:         "Intro To Go",
				InstructorID: instructorIDValue,
			},
		},
		{
			name: "invalid instructor id",
			input: core.CreateCourseInput{
				Title:        "Intro to Go",
				Slug:         "intro-to-go",
				InstructorID: "bad-id",
			},
		},
		{
			name: "missing title",
			input: core.CreateCourseInput{
				Title:        "   ",
				Slug:         "intro-to-go",
				InstructorID: instructorIDValue,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			courses := newCourseRepositoryFake()
			service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, fixedClock{})

			if _, err := service.CreateCourse(test.input); !errors.Is(err, domain.ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
			if len(courses.savedCourses) != 0 {
				t.Fatalf("expected invalid course not to be saved")
			}
		})
	}
}

func TestCreateCourseRejectsTakenSlug(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, otherCourseID, "Existing", "intro-to-go", domain.Draft()))
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, fixedClock{})

	_, err := service.CreateCourse(core.CreateCourseInput{
		Title:        "Intro to Go",
		Slug:         "intro-to-go",
		InstructorID: instructorIDValue,
	})
	if !errors.Is(err, domain.ErrSlugTaken) {
		t.Fatalf("expected slug taken error, got %v", err)
	}
}

func TestListCoursesSupportsNoFilter(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Draft Course", "draft-course", domain.Draft()))
	courses.store(mustCourse(t, otherCourseID, "Published Course", "published-course", domain.Published()))
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, fixedClock{})

	out, err := service.ListCourses(core.ListCoursesInput{})
	if err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}

	if len(out.Courses) != 2 {
		t.Fatalf("expected two courses, got %d", len(out.Courses))
	}
}

func TestListCoursesFiltersByStatus(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Draft Course", "draft-course", domain.Draft()))
	courses.store(mustCourse(t, otherCourseID, "Published Course", "published-course", domain.Published()))
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, fixedClock{})

	out, err := service.ListCourses(core.ListCoursesInput{Status: "published"})
	if err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}

	if len(out.Courses) != 1 {
		t.Fatalf("expected one published course, got %d", len(out.Courses))
	}
	if out.Courses[0].ID != otherCourseID || out.Courses[0].Status != "published" {
		t.Fatalf("expected published course view, got %+v", out.Courses[0])
	}
}

func TestListCoursesRejectsInvalidStatus(t *testing.T) {
	service := newCourseServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, fixedClock{})

	_, err := service.ListCourses(core.ListCoursesInput{Status: "archived"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestGetCourseReturnsMappedView(t *testing.T) {
	courses := newCourseRepositoryFake()
	course := mustCourse(t, courseIDValue, "Intro to Go", "intro-to-go", domain.Draft())
	course.ChangeDescription("Learn Go", course.CreatedAt().Add(time.Hour))
	courses.store(course)
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, fixedClock{})

	out, err := service.GetCourse(core.GetCourseInput{ID: courseIDValue})
	if err != nil {
		t.Fatalf("expected get to succeed, got %v", err)
	}

	want := core.CourseView{
		ID:           courseIDValue,
		Title:        "Intro to Go",
		Slug:         "intro-to-go",
		Description:  "Learn Go",
		InstructorID: instructorIDValue,
		Status:       "draft",
		CreatedAt:    course.CreatedAt(),
		UpdatedAt:    course.UpdatedAt(),
	}
	if out.Course != want {
		t.Fatalf("expected course view %+v, got %+v", want, out.Course)
	}
}

func TestGetCourseValidatesID(t *testing.T) {
	service := newCourseServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, fixedClock{})

	_, err := service.GetCourse(core.GetCourseInput{ID: "bad-id"})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestGetCourseReturnsNotFound(t *testing.T) {
	service := newCourseServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, fixedClock{})

	_, err := service.GetCourse(core.GetCourseInput{ID: courseIDValue})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateCourseChangesProvidedFields(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 9, 0, 0, 0, time.UTC)}
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro to Go", "intro-to-go", domain.Draft()))
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, clock)
	title := "Advanced Go"
	description := "Deep Go"
	slug := "advanced-go"

	out, err := service.UpdateCourse(core.UpdateCourseInput{
		ID:          courseIDValue,
		Title:       &title,
		Description: &description,
		Slug:        &slug,
	})
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if out.ID != courseIDValue {
		t.Fatalf("expected output id %q, got %q", courseIDValue, out.ID)
	}

	saved := courses.savedCourses[0]
	if saved.Title() != title || saved.Description() != description || saved.Slug().String() != slug {
		t.Fatalf("expected course fields to be updated")
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestUpdateCourseRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.UpdateCourseInput
		seed      []domain.Course
		wantError error
	}{
		{
			name: "not found",
			input: core.UpdateCourseInput{
				ID:    courseIDValue,
				Title: stringPointer("Advanced Go"),
			},
			wantError: domain.ErrNotFound,
		},
		{
			name: "nothing to update",
			input: core.UpdateCourseInput{
				ID: courseIDValue,
			},
			seed:      []domain.Course{mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft())},
			wantError: domain.ErrValidation,
		},
		{
			name: "invalid slug",
			input: core.UpdateCourseInput{
				ID:   courseIDValue,
				Slug: stringPointer("Bad Slug"),
			},
			seed:      []domain.Course{mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft())},
			wantError: domain.ErrValidation,
		},
		{
			name: "slug taken by another course",
			input: core.UpdateCourseInput{
				ID:   courseIDValue,
				Slug: stringPointer("taken"),
			},
			seed: []domain.Course{
				mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()),
				mustCourse(t, otherCourseID, "Other", "taken", domain.Draft()),
			},
			wantError: domain.ErrSlugTaken,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			courses := newCourseRepositoryFake()
			for _, course := range test.seed {
				courses.store(course)
			}

			service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, fixedClock{})
			_, err := service.UpdateCourse(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
		})
	}
}

func TestDeleteCourseDeletesDependentsBeforeCourse(t *testing.T) {
	operations := []string{}
	courses := newCourseRepositoryFake()
	courses.operations = &operations
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := &lessonRepositoryFake{operations: &operations}
	quizzes := newQuizRepositoryFake()
	quizzes.operations = &operations
	practices := newPracticeRepositoryFake()
	practices.operations = &operations
	tests := &courseTestRepositoryFake{operations: &operations}
	service := NewCourseService(courses, lessons, quizzes, fixedIDGenerator{courseID: mustCourseID(courseIDValue)}, fixedClock{}, practices, tests)

	if err := service.DeleteCourse(core.DeleteCourseInput{ID: courseIDValue}); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	wantCalls := []string{
		"lessons:" + courseIDValue,
		"quizzes:" + courseIDValue,
		"practices:" + courseIDValue,
		"tests:" + courseIDValue,
		"course:" + courseIDValue,
	}
	if !reflect.DeepEqual(operations, wantCalls) {
		t.Fatalf("expected delete order %v, got %v", wantCalls, operations)
	}
}

func TestDeleteCourseReturnsTestCascadeError(t *testing.T) {
	operations := []string{}
	courses := newCourseRepositoryFake()
	courses.operations = &operations
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := &lessonRepositoryFake{operations: &operations}
	quizzes := newQuizRepositoryFake()
	quizzes.operations = &operations
	practices := newPracticeRepositoryFake()
	practices.operations = &operations
	wantErr := errors.New("delete tests")
	tests := &courseTestRepositoryFake{operations: &operations, deleteByCourseErr: wantErr}
	service := NewCourseService(courses, lessons, quizzes, fixedIDGenerator{courseID: mustCourseID(courseIDValue)}, fixedClock{}, practices, tests)

	err := service.DeleteCourse(core.DeleteCourseInput{ID: courseIDValue})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected test cascade error, got %v", err)
	}

	wantCalls := []string{
		"lessons:" + courseIDValue,
		"quizzes:" + courseIDValue,
		"practices:" + courseIDValue,
		"tests:" + courseIDValue,
	}
	if !reflect.DeepEqual(operations, wantCalls) {
		t.Fatalf("expected delete order %v, got %v", wantCalls, operations)
	}
	if _, err := courses.FindByID(mustCourseID(courseIDValue)); err != nil {
		t.Fatalf("expected course deletion to stop after test cascade error, got %v", err)
	}
}

func TestDeleteCourseReturnsNotFound(t *testing.T) {
	service := newCourseServiceFixture(newCourseRepositoryFake(), &lessonRepositoryFake{}, fixedClock{})

	err := service.DeleteCourse(core.DeleteCourseInput{ID: courseIDValue})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestPublishCourse(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)}
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, clock)

	if err := service.PublishCourse(core.PublishCourseInput{ID: courseIDValue}); err != nil {
		t.Fatalf("expected publish to succeed, got %v", err)
	}

	saved := courses.savedCourses[0]
	if saved.Status() != domain.Published() || !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected published course with updated timestamp")
	}
}

func TestPublishCourseRejectsAlreadyPublished(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Published()))
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, fixedClock{})

	err := service.PublishCourse(core.PublishCourseInput{ID: courseIDValue})
	if !errors.Is(err, domain.ErrAlreadyPublished) {
		t.Fatalf("expected already published error, got %v", err)
	}
}

func TestUnpublishCourse(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 11, 0, 0, 0, time.UTC)}
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Published()))
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, clock)

	if err := service.UnpublishCourse(core.UnpublishCourseInput{ID: courseIDValue}); err != nil {
		t.Fatalf("expected unpublish to succeed, got %v", err)
	}

	saved := courses.savedCourses[0]
	if saved.Status() != domain.Draft() || !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected draft course with updated timestamp")
	}
}

func TestUnpublishCourseRejectsDraftCourse(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	service := newCourseServiceFixture(courses, &lessonRepositoryFake{}, fixedClock{})

	err := service.UnpublishCourse(core.UnpublishCourseInput{ID: courseIDValue})
	if !errors.Is(err, domain.ErrNotPublished) {
		t.Fatalf("expected not published error, got %v", err)
	}
}

func newCourseServiceFixture(
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	clock fixedClock,
) *CourseService {
	return NewCourseService(courses, lessons, newQuizRepositoryFake(), fixedIDGenerator{courseID: mustCourseID(courseIDValue)}, clock, newPracticeRepositoryFake())
}

type courseRepositoryFake struct {
	courses      map[string]domain.Course
	savedCourses []domain.Course
	operations   *[]string
}

func newCourseRepositoryFake() *courseRepositoryFake {
	return &courseRepositoryFake{courses: make(map[string]domain.Course)}
}

func (repo *courseRepositoryFake) Save(course domain.Course) error {
	repo.savedCourses = append(repo.savedCourses, course)
	repo.store(course)
	return nil
}

func (repo *courseRepositoryFake) FindByID(id domain.CourseID) (domain.Course, error) {
	course, exists := repo.courses[id.String()]
	if !exists {
		return domain.Course{}, domain.ErrNotFound
	}

	return course, nil
}

func (repo *courseRepositoryFake) FindBySlug(slug domain.Slug) (domain.Course, error) {
	for _, course := range repo.courses {
		if course.Slug() == slug {
			return course, nil
		}
	}

	return domain.Course{}, domain.ErrNotFound
}

func (repo *courseRepositoryFake) FindAll(filter core.CourseFilter) ([]domain.Course, error) {
	courses := make([]domain.Course, 0, len(repo.courses))
	for _, course := range repo.courses {
		if filter.Status != nil && course.Status() != *filter.Status {
			continue
		}

		courses = append(courses, course)
	}

	return courses, nil
}

func (repo *courseRepositoryFake) Delete(id domain.CourseID) error {
	if _, exists := repo.courses[id.String()]; !exists {
		return domain.ErrNotFound
	}

	if repo.operations != nil {
		*repo.operations = append(*repo.operations, "course:"+id.String())
	}
	delete(repo.courses, id.String())

	return nil
}

func (repo *courseRepositoryFake) store(course domain.Course) {
	repo.courses[course.ID().String()] = course
}

type lessonRepositoryFake struct {
	operations *[]string
	lessons    map[string]domain.Lesson

	savedLessons       []domain.Lesson
	savedAllLessons    []domain.Lesson
	findByCourseExtras []domain.Lesson
	saveAllErr         error
}

func (repo *lessonRepositoryFake) Save(lesson domain.Lesson) error {
	repo.savedLessons = append(repo.savedLessons, lesson)
	repo.store(lesson)

	return nil
}

func (repo *lessonRepositoryFake) SaveAll(lessons []domain.Lesson) error {
	if repo.saveAllErr != nil {
		return repo.saveAllErr
	}

	repo.savedAllLessons = append(repo.savedAllLessons, lessons...)
	for _, lesson := range lessons {
		repo.store(lesson)
	}

	return nil
}

func (repo *lessonRepositoryFake) FindByID(id domain.LessonID) (domain.Lesson, error) {
	lesson, exists := repo.lessons[id.String()]
	if !exists {
		return domain.Lesson{}, domain.ErrNotFound
	}

	return lesson, nil
}

func (repo *lessonRepositoryFake) FindByCourse(courseID domain.CourseID) ([]domain.Lesson, error) {
	lessons := make([]domain.Lesson, 0, len(repo.lessons))
	for _, lesson := range repo.lessons {
		if lesson.CourseID() == courseID {
			lessons = append(lessons, lesson)
		}
	}
	lessons = append(lessons, repo.findByCourseExtras...)

	sort.Slice(lessons, func(i, j int) bool {
		return lessons[i].Order().Int() < lessons[j].Order().Int()
	})

	return lessons, nil
}

func (repo *lessonRepositoryFake) FindByBlockID(id domain.BlockID) (domain.Lesson, error) {
	for _, lesson := range repo.lessons {
		if _, err := lesson.Block(id); err == nil {
			return lesson, nil
		}
	}

	return domain.Lesson{}, domain.ErrNotFound
}

func (repo *lessonRepositoryFake) FindLessonsEmbeddingQuiz(quizID domain.QuizID) ([]domain.LessonID, error) {
	lessonIDs := []domain.LessonID{}
	for _, lesson := range repo.lessons {
		for _, block := range lesson.Blocks() {
			body, ok := block.Body().(domain.QuizBody)
			if ok && body.QuizRef == quizID {
				lessonIDs = append(lessonIDs, lesson.ID())
				break
			}
		}
	}

	return lessonIDs, nil
}

func (repo *lessonRepositoryFake) FindLessonsEmbeddingPractice(practiceID domain.PracticeID) ([]domain.LessonID, error) {
	lessonIDs := []domain.LessonID{}
	for _, lesson := range repo.lessons {
		for _, block := range lesson.Blocks() {
			body, ok := block.Body().(domain.PracticeBody)
			if ok && body.PracticeRef == practiceID {
				lessonIDs = append(lessonIDs, lesson.ID())
				break
			}
		}
	}

	return lessonIDs, nil
}

func (repo *lessonRepositoryFake) Delete(id domain.LessonID) error {
	if _, exists := repo.lessons[id.String()]; !exists {
		return domain.ErrNotFound
	}

	delete(repo.lessons, id.String())

	return nil
}

func (repo *lessonRepositoryFake) DeleteByCourse(courseID domain.CourseID) error {
	if repo.operations != nil {
		*repo.operations = append(*repo.operations, "lessons:"+courseID.String())
	}

	return nil
}

func (repo *lessonRepositoryFake) store(lesson domain.Lesson) {
	if repo.lessons == nil {
		repo.lessons = make(map[string]domain.Lesson)
	}

	repo.lessons[lesson.ID().String()] = lesson
}

type courseTestRepositoryFake struct {
	operations        *[]string
	deleteByCourseErr error
}

func (repo *courseTestRepositoryFake) Save(domain.Test) error {
	return nil
}

func (repo *courseTestRepositoryFake) FindByID(domain.TestID) (domain.Test, error) {
	return domain.Test{}, domain.ErrNotFound
}

func (repo *courseTestRepositoryFake) FindByCourse(domain.CourseID) ([]domain.Test, error) {
	return nil, nil
}

func (repo *courseTestRepositoryFake) FindByItemID(domain.TestItemID) (domain.Test, error) {
	return domain.Test{}, domain.ErrNotFound
}

func (repo *courseTestRepositoryFake) Delete(domain.TestID) error {
	return nil
}

func (repo *courseTestRepositoryFake) DeleteByCourse(courseID domain.CourseID) error {
	if repo.operations != nil {
		*repo.operations = append(*repo.operations, "tests:"+courseID.String())
	}
	if repo.deleteByCourseErr != nil {
		return repo.deleteByCourseErr
	}

	return nil
}

type fixedIDGenerator struct {
	courseID   domain.CourseID
	lessonID   domain.LessonID
	blockID    domain.BlockID
	quizID     domain.QuizID
	questionID domain.QuestionID
	practiceID domain.PracticeID
	testCaseID domain.TestCaseID
	testID     domain.TestID
	testItemID domain.TestItemID
}

func (generator fixedIDGenerator) NewCourseID() domain.CourseID {
	return generator.courseID
}

func (generator fixedIDGenerator) NewLessonID() domain.LessonID {
	return generator.lessonID
}

func (generator fixedIDGenerator) NewBlockID() domain.BlockID {
	return generator.blockID
}

func (generator fixedIDGenerator) NewQuizID() domain.QuizID {
	return generator.quizID
}

func (generator fixedIDGenerator) NewQuestionID() domain.QuestionID {
	return generator.questionID
}

func (generator fixedIDGenerator) NewPracticeID() domain.PracticeID {
	return generator.practiceID
}

func (generator fixedIDGenerator) NewTestCaseID() domain.TestCaseID {
	return generator.testCaseID
}

func (generator fixedIDGenerator) NewTestID() domain.TestID {
	return generator.testID
}

func (generator fixedIDGenerator) NewTestItemID() domain.TestItemID {
	return generator.testItemID
}

type fixedClock struct {
	now time.Time
}

func (clock fixedClock) Now() time.Time {
	return clock.now
}

func mustCourse(t *testing.T, idValue string, title string, slugValue string, status domain.CourseStatus) domain.Course {
	t.Helper()

	course, err := domain.NewCourse(
		mustCourseID(idValue),
		title,
		mustSlug(slugValue),
		"",
		mustInstructorID(instructorIDValue),
		time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected course fixture, got %v", err)
	}

	if status.IsPublished() {
		if err := course.Publish(course.CreatedAt()); err != nil {
			t.Fatalf("expected published fixture, got %v", err)
		}
	}

	return course
}

func mustCourseID(value string) domain.CourseID {
	id, err := domain.NewCourseID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustInstructorID(value string) domain.InstructorID {
	id, err := domain.NewInstructorID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustSlug(value string) domain.Slug {
	slug, err := domain.NewSlug(value)
	if err != nil {
		panic(err)
	}

	return slug
}

func stringPointer(value string) *string {
	return &value
}
