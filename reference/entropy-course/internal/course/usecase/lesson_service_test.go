package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	lessonIDValue      = "550e8400-e29b-41d4-a716-446655440020"
	otherLessonIDValue = "550e8400-e29b-41d4-a716-446655440021"
	thirdLessonIDValue = "550e8400-e29b-41d4-a716-446655440022"
)

func TestCreateLessonSavesLessonWithExplicitOrder(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC)}
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := newLessonRepositoryFake()
	service := newLessonServiceFixture(courses, lessons, clock)
	order := 3

	out, err := service.CreateLesson(core.CreateLessonInput{
		CourseID: courseIDValue,
		Title:    "First Lesson",
		Order:    &order,
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	if out.ID != lessonIDValue {
		t.Fatalf("expected id %q, got %q", lessonIDValue, out.ID)
	}

	saved := lessons.savedLessons[0]
	if saved.CourseID().String() != courseIDValue || saved.Order().Int() != order {
		t.Fatalf("expected saved lesson with course id and explicit order")
	}
	if saved.Title() != "First Lesson" || len(saved.Blocks()) != 0 {
		t.Fatalf("expected title and empty blocks to be saved")
	}
	if !saved.CreatedAt().Equal(clock.now) || !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected deterministic timestamps")
	}
}

func TestCreateLessonAppendsWhenOrderIsUnset(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, otherLessonIDValue, courseIDValue, "Existing", "Content", 4))
	service := newLessonServiceFixture(courses, lessons, fixedClock{})

	_, err := service.CreateLesson(core.CreateLessonInput{
		CourseID: courseIDValue,
		Title:    "Next Lesson",
	})
	if err != nil {
		t.Fatalf("expected create to succeed, got %v", err)
	}

	if lessons.savedLessons[0].Order().Int() != 5 {
		t.Fatalf("expected appended order 5, got %d", lessons.savedLessons[0].Order().Int())
	}
}

func TestCreateLessonRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name       string
		input      core.CreateLessonInput
		seedCourse bool
		wantError  error
	}{
		{
			name: "invalid course id",
			input: core.CreateLessonInput{
				CourseID: "bad-id",
				Title:    "Lesson",
			},
			wantError: domain.ErrValidation,
		},
		{
			name: "course not found",
			input: core.CreateLessonInput{
				CourseID: courseIDValue,
				Title:    "Lesson",
			},
			wantError: domain.ErrNotFound,
		},
		{
			name: "missing title",
			input: core.CreateLessonInput{
				CourseID: courseIDValue,
				Title:    "   ",
			},
			seedCourse: true,
			wantError:  domain.ErrValidation,
		},
		{
			name: "negative order",
			input: core.CreateLessonInput{
				CourseID: courseIDValue,
				Title:    "Lesson",
				Order:    intPointer(-1),
			},
			seedCourse: true,
			wantError:  domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			courses := newCourseRepositoryFake()
			if test.seedCourse {
				courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
			}
			lessons := newLessonRepositoryFake()
			service := newLessonServiceFixture(courses, lessons, fixedClock{})

			_, err := service.CreateLesson(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(lessons.savedLessons) != 0 {
				t.Fatalf("expected invalid lesson not to be saved")
			}
		})
	}
}

func TestListLessonsValidatesCourseAndReturnsViews(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	lessons.store(mustLesson(t, otherLessonIDValue, courseIDValue, "Second", "More", 1))
	service := newLessonServiceFixture(courses, lessons, fixedClock{})

	out, err := service.ListLessons(core.ListLessonsInput{CourseID: courseIDValue})
	if err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}

	if len(out.Lessons) != 2 {
		t.Fatalf("expected two lessons, got %d", len(out.Lessons))
	}
	if out.Lessons[0].ID != lessonIDValue || out.Lessons[0].Order != 0 {
		t.Fatalf("expected first lesson view, got %+v", out.Lessons[0])
	}
}

func TestListLessonsReturnsCourseNotFound(t *testing.T) {
	service := newLessonServiceFixture(newCourseRepositoryFake(), newLessonRepositoryFake(), fixedClock{})

	_, err := service.ListLessons(core.ListLessonsInput{CourseID: courseIDValue})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestGetLessonReturnsMappedView(t *testing.T) {
	lessons := newLessonRepositoryFake()
	lesson := mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 2)
	lessons.store(lesson)
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, fixedClock{})

	out, err := service.GetLesson(core.GetLessonInput{ID: lessonIDValue})
	if err != nil {
		t.Fatalf("expected get to succeed, got %v", err)
	}

	want := core.LessonView{
		ID:        lessonIDValue,
		CourseID:  courseIDValue,
		Title:     "First",
		Order:     2,
		CreatedAt: lesson.CreatedAt(),
		UpdatedAt: lesson.UpdatedAt(),
	}
	if out.Lesson != want {
		t.Fatalf("expected lesson view %+v, got %+v", want, out.Lesson)
	}
}

func TestGetLessonFailureModes(t *testing.T) {
	service := newLessonServiceFixture(newCourseRepositoryFake(), newLessonRepositoryFake(), fixedClock{})

	if _, err := service.GetLesson(core.GetLessonInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if _, err := service.GetLesson(core.GetLessonInput{ID: lessonIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateLessonChangesProvidedFields(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 13, 0, 0, 0, time.UTC)}
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, clock)
	title := "Updated"

	out, err := service.UpdateLesson(core.UpdateLessonInput{
		ID:    lessonIDValue,
		Title: &title,
	})
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	if out.ID != lessonIDValue {
		t.Fatalf("expected output id %q, got %q", lessonIDValue, out.ID)
	}

	saved := lessons.savedLessons[0]
	if saved.Title() != title || saved.Content() != "Content" {
		t.Fatalf("expected title to update and blocks to remain unchanged")
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestUpdateLessonRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.UpdateLessonInput
		seed      bool
		wantError error
	}{
		{
			name:      "invalid id",
			input:     core.UpdateLessonInput{ID: "bad-id", Title: stringPointer("Updated")},
			wantError: domain.ErrValidation,
		},
		{
			name:      "nothing to update",
			input:     core.UpdateLessonInput{ID: lessonIDValue},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name:      "not found",
			input:     core.UpdateLessonInput{ID: lessonIDValue, Title: stringPointer("Updated")},
			wantError: domain.ErrNotFound,
		},
		{
			name:      "empty title",
			input:     core.UpdateLessonInput{ID: lessonIDValue, Title: stringPointer("   ")},
			seed:      true,
			wantError: domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lessons := newLessonRepositoryFake()
			if test.seed {
				lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
			}
			service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, fixedClock{})

			_, err := service.UpdateLesson(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if test.name == "nothing to update" && err.Error() != "update: nothing to update" {
				t.Fatalf("expected nothing to update error, got %v", err)
			}
		})
	}
}

func TestDeleteLessonDeletesByID(t *testing.T) {
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, fixedClock{})

	if err := service.DeleteLesson(core.DeleteLessonInput{ID: lessonIDValue}); err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}

	if _, exists := lessons.lessons[lessonIDValue]; exists {
		t.Fatalf("expected lesson to be deleted")
	}
}

func TestDeleteLessonFailureModes(t *testing.T) {
	service := newLessonServiceFixture(newCourseRepositoryFake(), newLessonRepositoryFake(), fixedClock{})

	if err := service.DeleteLesson(core.DeleteLessonInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if err := service.DeleteLesson(core.DeleteLessonInput{ID: lessonIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestReorderLessonsUpdatesSelectedLessonsWithSharedTimestamp(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 14, 0, 0, 0, time.UTC)}
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	lessons.store(mustLesson(t, otherLessonIDValue, courseIDValue, "Second", "More", 1))
	lessons.store(mustLesson(t, thirdLessonIDValue, courseIDValue, "Third", "Rest", 2))
	service := newLessonServiceFixture(courses, lessons, clock)

	err := service.ReorderLessons(core.ReorderLessonsInput{
		CourseID: courseIDValue,
		Order: []core.LessonPosition{
			{LessonID: otherLessonIDValue, Position: 0},
			{LessonID: lessonIDValue, Position: 1},
		},
	})
	if err != nil {
		t.Fatalf("expected reorder to succeed, got %v", err)
	}

	if len(lessons.savedAllLessons) != 2 {
		t.Fatalf("expected two reordered lessons, got %d", len(lessons.savedAllLessons))
	}

	for _, saved := range lessons.savedAllLessons {
		if !saved.UpdatedAt().Equal(clock.now) {
			t.Fatalf("expected shared timestamp %v, got %v", clock.now, saved.UpdatedAt())
		}
	}

	if lessons.lessons[otherLessonIDValue].Order().Int() != 0 {
		t.Fatalf("expected second lesson to move to position 0")
	}
	if lessons.lessons[lessonIDValue].Order().Int() != 1 {
		t.Fatalf("expected first lesson to move to position 1")
	}
	if lessons.lessons[thirdLessonIDValue].Order().Int() != 2 {
		t.Fatalf("expected unmentioned lesson to remain unchanged")
	}
}

func TestReorderLessonsRejectsDuplicatePositions(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	lessons.store(mustLesson(t, otherLessonIDValue, courseIDValue, "Second", "More", 1))
	service := newLessonServiceFixture(courses, lessons, fixedClock{})

	err := service.ReorderLessons(core.ReorderLessonsInput{
		CourseID: courseIDValue,
		Order: []core.LessonPosition{
			{LessonID: lessonIDValue, Position: 1},
			{LessonID: otherLessonIDValue, Position: 1},
		},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(lessons.savedAllLessons) != 0 {
		t.Fatalf("expected invalid reorder not to be saved")
	}
}

func TestReorderLessonsRejectsUnknownLesson(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	service := newLessonServiceFixture(courses, lessons, fixedClock{})

	err := service.ReorderLessons(core.ReorderLessonsInput{
		CourseID: courseIDValue,
		Order:    []core.LessonPosition{{LessonID: otherLessonIDValue, Position: 0}},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(lessons.savedAllLessons) != 0 {
		t.Fatalf("expected unknown lesson not to be saved")
	}
}

func TestReorderLessonsRejectsLessonOutsideCourse(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := newLessonRepositoryFake()
	lessons.findByCourseExtras = []domain.Lesson{
		mustLesson(t, lessonIDValue, otherCourseID, "Other", "Content", 0),
	}
	service := newLessonServiceFixture(courses, lessons, fixedClock{})

	err := service.ReorderLessons(core.ReorderLessonsInput{
		CourseID: courseIDValue,
		Order:    []core.LessonPosition{{LessonID: lessonIDValue, Position: 0}},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(lessons.savedAllLessons) != 0 {
		t.Fatalf("expected outside-course lesson not to be saved")
	}
}

func TestReorderLessonsRejectsInvalidPosition(t *testing.T) {
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	service := newLessonServiceFixture(courses, lessons, fixedClock{})

	err := service.ReorderLessons(core.ReorderLessonsInput{
		CourseID: courseIDValue,
		Order:    []core.LessonPosition{{LessonID: lessonIDValue, Position: -1}},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(lessons.savedAllLessons) != 0 {
		t.Fatalf("expected invalid position not to be saved")
	}
}

func TestReorderLessonsReturnsCourseNotFound(t *testing.T) {
	service := newLessonServiceFixture(newCourseRepositoryFake(), newLessonRepositoryFake(), fixedClock{})

	err := service.ReorderLessons(core.ReorderLessonsInput{
		CourseID: courseIDValue,
		Order:    []core.LessonPosition{{LessonID: lessonIDValue, Position: 0}},
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestReorderLessonsPropagatesSaveAllError(t *testing.T) {
	errBoom := errors.New("save all failed")
	courses := newCourseRepositoryFake()
	courses.store(mustCourse(t, courseIDValue, "Intro", "intro", domain.Draft()))
	lessons := newLessonRepositoryFake()
	lessons.saveAllErr = errBoom
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	service := newLessonServiceFixture(courses, lessons, fixedClock{})

	err := service.ReorderLessons(core.ReorderLessonsInput{
		CourseID: courseIDValue,
		Order:    []core.LessonPosition{{LessonID: lessonIDValue, Position: 2}},
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected save all error, got %v", err)
	}
}

func TestAddLessonBlockAppendsTextBlockByDefault(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 15, 0, 0, 0, time.UTC)}
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, clock)

	out, err := service.AddLessonBlock(core.AddLessonBlockInput{
		LessonID: lessonIDValue,
		Kind:     "text",
		Markdown: "Second block",
	})
	if err != nil {
		t.Fatalf("expected add block to succeed, got %v", err)
	}

	if out.ID != thirdLessonIDValue {
		t.Fatalf("expected generated block id %q, got %q", thirdLessonIDValue, out.ID)
	}

	saved := lessons.savedLessons[0]
	blocks := saved.Blocks()
	if len(blocks) != 2 {
		t.Fatalf("expected two blocks, got %d", len(blocks))
	}
	if blocks[1].ID().String() != thirdLessonIDValue || blocks[1].Position().Int() != 1 {
		t.Fatalf("expected new block appended at position 1, got %+v", blocks[1])
	}
	if body := blocks[1].Body().(domain.TextBody); body.Markdown != "Second block" {
		t.Fatalf("expected markdown to map, got %q", body.Markdown)
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestAddLessonBlockInsertsVideoBlockAtExplicitPosition(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 15, 0, 0, 0, time.UTC)}
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, clock)
	position := 0

	_, err := service.AddLessonBlock(core.AddLessonBlockInput{
		LessonID:      lessonIDValue,
		Kind:          "video",
		VideoProvider: "youtube",
		VideoLocator:  "https://youtu.be/dQw4w9WgXcQ",
		VideoCaption:  "Intro video",
		Position:      &position,
	})
	if err != nil {
		t.Fatalf("expected add block to succeed, got %v", err)
	}

	blocks := lessons.savedLessons[0].Blocks()
	if blocks[0].ID().String() != thirdLessonIDValue || blocks[0].Position().Int() != 0 {
		t.Fatalf("expected video block inserted at position 0, got %+v", blocks[0])
	}
	body := blocks[0].Body().(domain.VideoBody)
	if body.Media.Provider() != domain.YouTubeProvider() || body.Media.Locator() != "https://youtu.be/dQw4w9WgXcQ" || body.Caption != "Intro video" {
		t.Fatalf("expected video payload to map, got %+v", body)
	}
	if blocks[1].ID().String() != lessonIDValue || blocks[1].Position().Int() != 1 {
		t.Fatalf("expected existing block to shift to position 1, got %+v", blocks[1])
	}
}

func TestAddLessonBlockAddsQuizBlockForSameCourse(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 15, 0, 0, 0, time.UTC)}
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, courseIDValue, "Basics Quiz", 0.7))
	service := newLessonServiceFixtureWithQuizzes(newCourseRepositoryFake(), lessons, quizzes, clock)

	out, err := service.AddLessonBlock(core.AddLessonBlockInput{
		LessonID: lessonIDValue,
		Kind:     "quiz",
		QuizRef:  quizIDValue,
	})
	if err != nil {
		t.Fatalf("expected add quiz block to succeed, got %v", err)
	}

	if out.ID != thirdLessonIDValue {
		t.Fatalf("expected generated block id %q, got %q", thirdLessonIDValue, out.ID)
	}

	saved := lessons.savedLessons[0]
	blocks := saved.Blocks()
	if len(blocks) != 2 {
		t.Fatalf("expected two blocks, got %d", len(blocks))
	}
	if blocks[1].Kind() != domain.QuizKind() || blocks[1].Position().Int() != 1 {
		t.Fatalf("expected appended quiz block at position 1, got %+v", blocks[1])
	}
	body := blocks[1].Body().(domain.QuizBody)
	if body.QuizRef.String() != quizIDValue {
		t.Fatalf("expected quiz ref %q, got %q", quizIDValue, body.QuizRef.String())
	}

	listOut, err := service.ListLessonBlocks(core.ListLessonBlocksInput{LessonID: lessonIDValue})
	if err != nil {
		t.Fatalf("expected list blocks to succeed, got %v", err)
	}
	if listOut.Blocks[1].QuizRef != quizIDValue {
		t.Fatalf("expected quiz ref in block view, got %+v", listOut.Blocks[1])
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestAddLessonBlockRejectsMissingQuiz(t *testing.T) {
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	service := newLessonServiceFixtureWithQuizzes(newCourseRepositoryFake(), lessons, newQuizRepositoryFake(), fixedClock{})

	_, err := service.AddLessonBlock(core.AddLessonBlockInput{
		LessonID: lessonIDValue,
		Kind:     "quiz",
		QuizRef:  quizIDValue,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if len(lessons.savedLessons) != 0 {
		t.Fatalf("expected missing quiz block not to be saved")
	}
}

func TestAddLessonBlockRejectsCrossCourseQuiz(t *testing.T) {
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	quizzes := newQuizRepositoryFake()
	quizzes.store(mustQuiz(t, quizIDValue, otherCourseID, "Other Course Quiz", 0.7))
	service := newLessonServiceFixtureWithQuizzes(newCourseRepositoryFake(), lessons, quizzes, fixedClock{})

	_, err := service.AddLessonBlock(core.AddLessonBlockInput{
		LessonID: lessonIDValue,
		Kind:     "quiz",
		QuizRef:  quizIDValue,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(lessons.savedLessons) != 0 {
		t.Fatalf("expected cross-course quiz block not to be saved")
	}
}

func TestAddLessonBlockAddsPracticeBlockForSameCourse(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 26, 17, 0, 0, 0, time.UTC)}
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(t, practiceIDValue, courseIDValue, "FizzBuzz", "golang", "Prompt", "", "solution"))
	service := newLessonServiceFixtureWithPractices(newCourseRepositoryFake(), lessons, practices, clock)

	out, err := service.AddLessonBlock(core.AddLessonBlockInput{
		LessonID:    lessonIDValue,
		Kind:        "practice",
		PracticeRef: practiceIDValue,
	})
	if err != nil {
		t.Fatalf("expected add practice block to succeed, got %v", err)
	}

	if out.ID != thirdLessonIDValue {
		t.Fatalf("expected generated block id %q, got %q", thirdLessonIDValue, out.ID)
	}

	saved := lessons.savedLessons[0]
	blocks := saved.Blocks()
	if len(blocks) != 2 {
		t.Fatalf("expected two blocks, got %d", len(blocks))
	}
	if blocks[1].Kind() != domain.PracticeKind() || blocks[1].Position().Int() != 1 {
		t.Fatalf("expected appended practice block at position 1, got %+v", blocks[1])
	}
	body := blocks[1].Body().(domain.PracticeBody)
	if body.PracticeRef.String() != practiceIDValue {
		t.Fatalf("expected practice ref %q, got %q", practiceIDValue, body.PracticeRef.String())
	}

	listOut, err := service.ListLessonBlocks(core.ListLessonBlocksInput{LessonID: lessonIDValue})
	if err != nil {
		t.Fatalf("expected list blocks to succeed, got %v", err)
	}
	if listOut.Blocks[1].PracticeRef != practiceIDValue {
		t.Fatalf("expected practice ref in block view, got %+v", listOut.Blocks[1])
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestAddLessonBlockRejectsMissingPractice(t *testing.T) {
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	service := newLessonServiceFixtureWithPractices(newCourseRepositoryFake(), lessons, newPracticeRepositoryFake(), fixedClock{})

	_, err := service.AddLessonBlock(core.AddLessonBlockInput{
		LessonID:    lessonIDValue,
		Kind:        "practice",
		PracticeRef: practiceIDValue,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if len(lessons.savedLessons) != 0 {
		t.Fatalf("expected missing practice block not to be saved")
	}
}

func TestAddLessonBlockRejectsCrossCoursePractice(t *testing.T) {
	lessons := newLessonRepositoryFake()
	lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
	practices := newPracticeRepositoryFake()
	practices.store(mustPractice(t, practiceIDValue, otherCourseID, "Other Practice", "golang", "Prompt", "", "solution"))
	service := newLessonServiceFixtureWithPractices(newCourseRepositoryFake(), lessons, practices, fixedClock{})

	_, err := service.AddLessonBlock(core.AddLessonBlockInput{
		LessonID:    lessonIDValue,
		Kind:        "practice",
		PracticeRef: practiceIDValue,
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if len(lessons.savedLessons) != 0 {
		t.Fatalf("expected cross-course practice block not to be saved")
	}
}

func TestAddLessonBlockRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.AddLessonBlockInput
		seed      bool
		wantError error
	}{
		{
			name:      "invalid lesson id",
			input:     core.AddLessonBlockInput{LessonID: "bad-id", Kind: "text"},
			wantError: domain.ErrValidation,
		},
		{
			name:      "missing lesson",
			input:     core.AddLessonBlockInput{LessonID: lessonIDValue, Kind: "text"},
			wantError: domain.ErrNotFound,
		},
		{
			name:      "invalid practice ref",
			input:     core.AddLessonBlockInput{LessonID: lessonIDValue, Kind: "practice"},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name:      "invalid quiz ref",
			input:     core.AddLessonBlockInput{LessonID: lessonIDValue, Kind: "quiz", QuizRef: "bad-id"},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "invalid media ref",
			input: core.AddLessonBlockInput{
				LessonID:      lessonIDValue,
				Kind:          "video",
				VideoProvider: "youtube",
				VideoLocator:  "https://example.com/watch?v=dQw4w9WgXcQ",
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name: "invalid position",
			input: core.AddLessonBlockInput{
				LessonID: lessonIDValue,
				Kind:     "text",
				Position: intPointer(-1),
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lessons := newLessonRepositoryFake()
			if test.seed {
				lessons.store(mustLesson(t, lessonIDValue, courseIDValue, "First", "Content", 0))
			}
			service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, fixedClock{})

			_, err := service.AddLessonBlock(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(lessons.savedLessons) != 0 {
				t.Fatalf("expected invalid block not to be saved")
			}
		})
	}
}

func TestListLessonBlocksReturnsOrderedBlockViews(t *testing.T) {
	lessons := newLessonRepositoryFake()
	lessons.store(mustLessonWithBlocks(t,
		mustTextBlock(t, lessonIDValue, 0, "Content"),
		mustVideoBlock(t, thirdLessonIDValue, 1, "https://youtu.be/dQw4w9WgXcQ", "Intro video"),
		mustQuizBlock(t, otherLessonIDValue, 2, quizIDValue),
	))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, fixedClock{})

	out, err := service.ListLessonBlocks(core.ListLessonBlocksInput{LessonID: lessonIDValue})
	if err != nil {
		t.Fatalf("expected list blocks to succeed, got %v", err)
	}

	want := []core.BlockView{
		{
			ID:       lessonIDValue,
			LessonID: lessonIDValue,
			Kind:     "text",
			Position: 0,
			Markdown: "Content",
		},
		{
			ID:            thirdLessonIDValue,
			LessonID:      lessonIDValue,
			Kind:          "video",
			Position:      1,
			VideoProvider: "youtube",
			VideoLocator:  "https://youtu.be/dQw4w9WgXcQ",
			VideoCaption:  "Intro video",
		},
		{
			ID:       otherLessonIDValue,
			LessonID: lessonIDValue,
			Kind:     "quiz",
			Position: 2,
			QuizRef:  quizIDValue,
		},
	}
	if len(out.Blocks) != len(want) {
		t.Fatalf("expected %d blocks, got %d", len(want), len(out.Blocks))
	}
	for index := range want {
		if out.Blocks[index] != want[index] {
			t.Fatalf("expected block view %+v at index %d, got %+v", want[index], index, out.Blocks[index])
		}
	}
}

func TestListLessonBlocksRejectsFailureModes(t *testing.T) {
	service := newLessonServiceFixture(newCourseRepositoryFake(), newLessonRepositoryFake(), fixedClock{})

	if _, err := service.ListLessonBlocks(core.ListLessonBlocksInput{LessonID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.ListLessonBlocks(core.ListLessonBlocksInput{LessonID: lessonIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestGetLessonBlockLoadsOwningLessonAndReturnsView(t *testing.T) {
	lessons := newLessonRepositoryFake()
	lessons.store(mustLessonWithBlocks(t,
		mustTextBlock(t, lessonIDValue, 0, "Content"),
		mustVideoBlock(t, thirdLessonIDValue, 1, "https://youtu.be/dQw4w9WgXcQ", "Intro video"),
	))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, fixedClock{})

	out, err := service.GetLessonBlock(core.GetLessonBlockInput{ID: thirdLessonIDValue})
	if err != nil {
		t.Fatalf("expected get block to succeed, got %v", err)
	}

	want := core.BlockView{
		ID:            thirdLessonIDValue,
		LessonID:      lessonIDValue,
		Kind:          "video",
		Position:      1,
		VideoProvider: "youtube",
		VideoLocator:  "https://youtu.be/dQw4w9WgXcQ",
		VideoCaption:  "Intro video",
	}
	if out.Block != want {
		t.Fatalf("expected block view %+v, got %+v", want, out.Block)
	}
}

func TestGetLessonBlockRejectsFailureModes(t *testing.T) {
	service := newLessonServiceFixture(newCourseRepositoryFake(), newLessonRepositoryFake(), fixedClock{})

	if _, err := service.GetLessonBlock(core.GetLessonBlockInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if _, err := service.GetLessonBlock(core.GetLessonBlockInput{ID: thirdLessonIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestUpdateLessonBlockUpdatesTextBlock(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC)}
	lessons := newLessonRepositoryFake()
	lessons.store(mustLessonWithBlocks(t, mustTextBlock(t, lessonIDValue, 0, "Content")))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, clock)
	markdown := "Updated content"
	caption := "ignored"

	out, err := service.UpdateLessonBlock(core.UpdateLessonBlockInput{
		ID:           lessonIDValue,
		Markdown:     &markdown,
		VideoCaption: &caption,
	})
	if err != nil {
		t.Fatalf("expected update block to succeed, got %v", err)
	}

	if out.ID != lessonIDValue {
		t.Fatalf("expected output id %q, got %q", lessonIDValue, out.ID)
	}

	saved := lessons.savedLessons[0]
	block, err := saved.Block(mustBlockID(lessonIDValue))
	if err != nil {
		t.Fatalf("expected block to exist, got %v", err)
	}
	if body := block.Body().(domain.TextBody); body.Markdown != markdown {
		t.Fatalf("expected updated markdown, got %q", body.Markdown)
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestUpdateLessonBlockSupportsPartialVideoCaptionUpdate(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC)}
	lessons := newLessonRepositoryFake()
	lessons.store(mustLessonWithBlocks(t, mustVideoBlock(t, thirdLessonIDValue, 0, "https://youtu.be/dQw4w9WgXcQ", "Old caption")))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, clock)
	caption := "New caption"
	markdown := "ignored"

	_, err := service.UpdateLessonBlock(core.UpdateLessonBlockInput{
		ID:           thirdLessonIDValue,
		Markdown:     &markdown,
		VideoCaption: &caption,
	})
	if err != nil {
		t.Fatalf("expected update block to succeed, got %v", err)
	}

	block, err := lessons.savedLessons[0].Block(mustBlockID(thirdLessonIDValue))
	if err != nil {
		t.Fatalf("expected block to exist, got %v", err)
	}
	body := block.Body().(domain.VideoBody)
	if body.Caption != caption {
		t.Fatalf("expected updated caption, got %q", body.Caption)
	}
	if body.Media.Provider() != domain.YouTubeProvider() || body.Media.Locator() != "https://youtu.be/dQw4w9WgXcQ" {
		t.Fatalf("expected existing media ref to remain, got %+v", body.Media)
	}
}

func TestUpdateLessonBlockRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.UpdateLessonBlockInput
		seed      bool
		wantError error
	}{
		{
			name:      "invalid id",
			input:     core.UpdateLessonBlockInput{ID: "bad-id", Markdown: stringPointer("content")},
			wantError: domain.ErrValidation,
		},
		{
			name:      "empty update",
			input:     core.UpdateLessonBlockInput{ID: lessonIDValue},
			seed:      true,
			wantError: domain.ErrValidation,
		},
		{
			name:      "not found",
			input:     core.UpdateLessonBlockInput{ID: lessonIDValue, Markdown: stringPointer("content")},
			wantError: domain.ErrNotFound,
		},
		{
			name: "invalid media ref",
			input: core.UpdateLessonBlockInput{
				ID:           thirdLessonIDValue,
				VideoLocator: stringPointer("https://example.com/watch?v=dQw4w9WgXcQ"),
			},
			seed:      true,
			wantError: domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lessons := newLessonRepositoryFake()
			if test.seed {
				lessons.store(mustLessonWithBlocks(t, mustVideoBlock(t, thirdLessonIDValue, 0, "https://youtu.be/dQw4w9WgXcQ", "Caption")))
			}
			service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, fixedClock{})

			_, err := service.UpdateLessonBlock(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(lessons.savedLessons) != 0 {
				t.Fatalf("expected invalid update not to be saved")
			}
		})
	}
}

func TestRemoveLessonBlockCompactsPositionsAndSaves(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC)}
	lessons := newLessonRepositoryFake()
	lessons.store(mustLessonWithBlocks(t,
		mustTextBlock(t, lessonIDValue, 0, "First"),
		mustTextBlock(t, otherLessonIDValue, 1, "Second"),
		mustTextBlock(t, thirdLessonIDValue, 2, "Third"),
	))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, clock)

	if err := service.RemoveLessonBlock(core.RemoveLessonBlockInput{ID: otherLessonIDValue}); err != nil {
		t.Fatalf("expected remove block to succeed, got %v", err)
	}

	saved := lessons.savedLessons[0]
	blocks := saved.Blocks()
	if len(blocks) != 2 {
		t.Fatalf("expected two remaining blocks, got %d", len(blocks))
	}
	if blocks[0].ID().String() != lessonIDValue || blocks[0].Position().Int() != 0 {
		t.Fatalf("expected first block at position 0, got %+v", blocks[0])
	}
	if blocks[1].ID().String() != thirdLessonIDValue || blocks[1].Position().Int() != 1 {
		t.Fatalf("expected third block compacted to position 1, got %+v", blocks[1])
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected updated timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestRemoveLessonBlockRejectsFailureModes(t *testing.T) {
	service := newLessonServiceFixture(newCourseRepositoryFake(), newLessonRepositoryFake(), fixedClock{})

	if err := service.RemoveLessonBlock(core.RemoveLessonBlockInput{ID: "bad-id"}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if err := service.RemoveLessonBlock(core.RemoveLessonBlockInput{ID: lessonIDValue}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestReorderLessonBlocksSavesPermutationWithSharedTimestamp(t *testing.T) {
	clock := fixedClock{now: time.Date(2026, 5, 24, 16, 0, 0, 0, time.UTC)}
	lessons := newLessonRepositoryFake()
	lessons.store(mustLessonWithBlocks(t,
		mustTextBlock(t, lessonIDValue, 0, "First"),
		mustTextBlock(t, otherLessonIDValue, 1, "Second"),
		mustTextBlock(t, thirdLessonIDValue, 2, "Third"),
	))
	service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, clock)

	err := service.ReorderLessonBlocks(core.ReorderLessonBlocksInput{
		LessonID: lessonIDValue,
		Order: []core.BlockPlacementDTO{
			{BlockID: thirdLessonIDValue, Position: 0},
			{BlockID: lessonIDValue, Position: 1},
			{BlockID: otherLessonIDValue, Position: 2},
		},
	})
	if err != nil {
		t.Fatalf("expected reorder blocks to succeed, got %v", err)
	}

	saved := lessons.savedLessons[0]
	blocks := saved.Blocks()
	if blocks[0].ID().String() != thirdLessonIDValue || blocks[1].ID().String() != lessonIDValue || blocks[2].ID().String() != otherLessonIDValue {
		t.Fatalf("expected blocks to reorder, got %+v", blocks)
	}
	for index, block := range blocks {
		if block.Position().Int() != index {
			t.Fatalf("expected block %s at position %d, got %d", block.ID().String(), index, block.Position().Int())
		}
	}
	if !saved.UpdatedAt().Equal(clock.now) {
		t.Fatalf("expected shared timestamp %v, got %v", clock.now, saved.UpdatedAt())
	}
}

func TestReorderLessonBlocksRejectsFailureModes(t *testing.T) {
	tests := []struct {
		name      string
		input     core.ReorderLessonBlocksInput
		seed      []domain.Lesson
		wantError error
	}{
		{
			name:      "invalid lesson id",
			input:     core.ReorderLessonBlocksInput{LessonID: "bad-id"},
			wantError: domain.ErrValidation,
		},
		{
			name:      "missing lesson",
			input:     core.ReorderLessonBlocksInput{LessonID: lessonIDValue},
			wantError: domain.ErrNotFound,
		},
		{
			name: "invalid block id",
			input: core.ReorderLessonBlocksInput{
				LessonID: lessonIDValue,
				Order:    []core.BlockPlacementDTO{{BlockID: "bad-id", Position: 0}},
			},
			seed:      []domain.Lesson{mustLessonWithBlocks(t, mustTextBlock(t, lessonIDValue, 0, "First"))},
			wantError: domain.ErrValidation,
		},
		{
			name: "duplicate positions",
			input: core.ReorderLessonBlocksInput{
				LessonID: lessonIDValue,
				Order: []core.BlockPlacementDTO{
					{BlockID: lessonIDValue, Position: 0},
					{BlockID: otherLessonIDValue, Position: 0},
				},
			},
			seed: []domain.Lesson{mustLessonWithBlocks(t,
				mustTextBlock(t, lessonIDValue, 0, "First"),
				mustTextBlock(t, otherLessonIDValue, 1, "Second"),
			)},
			wantError: domain.ErrValidation,
		},
		{
			name: "missing placement",
			input: core.ReorderLessonBlocksInput{
				LessonID: lessonIDValue,
				Order:    []core.BlockPlacementDTO{{BlockID: lessonIDValue, Position: 0}},
			},
			seed: []domain.Lesson{mustLessonWithBlocks(t,
				mustTextBlock(t, lessonIDValue, 0, "First"),
				mustTextBlock(t, otherLessonIDValue, 1, "Second"),
			)},
			wantError: domain.ErrValidation,
		},
		{
			name: "extra placement",
			input: core.ReorderLessonBlocksInput{
				LessonID: lessonIDValue,
				Order: []core.BlockPlacementDTO{
					{BlockID: lessonIDValue, Position: 0},
					{BlockID: otherLessonIDValue, Position: 1},
				},
			},
			seed:      []domain.Lesson{mustLessonWithBlocks(t, mustTextBlock(t, lessonIDValue, 0, "First"))},
			wantError: domain.ErrValidation,
		},
		{
			name: "block from another lesson",
			input: core.ReorderLessonBlocksInput{
				LessonID: lessonIDValue,
				Order: []core.BlockPlacementDTO{
					{BlockID: lessonIDValue, Position: 0},
					{BlockID: thirdLessonIDValue, Position: 1},
				},
			},
			seed: []domain.Lesson{
				mustLessonWithBlocks(t, mustTextBlock(t, lessonIDValue, 0, "First"), mustTextBlock(t, otherLessonIDValue, 1, "Second")),
				mustLessonFixture(t, otherLessonIDValue, courseIDValue, "Other", []domain.ContentBlock{mustTextBlock(t, thirdLessonIDValue, 0, "Other block")}, 1),
			},
			wantError: domain.ErrValidation,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lessons := newLessonRepositoryFake()
			for _, lesson := range test.seed {
				lessons.store(lesson)
			}
			service := newLessonServiceFixture(newCourseRepositoryFake(), lessons, fixedClock{})

			err := service.ReorderLessonBlocks(test.input)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("expected %v, got %v", test.wantError, err)
			}
			if len(lessons.savedLessons) != 0 {
				t.Fatalf("expected invalid reorder not to be saved")
			}
		})
	}
}

func newLessonServiceFixture(
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	clock fixedClock,
) *LessonService {
	return newLessonServiceFixtureWithQuizzes(courses, lessons, newQuizRepositoryFake(), clock)
}

func newLessonServiceFixtureWithQuizzes(
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	quizzes *quizRepositoryFake,
	clock fixedClock,
) *LessonService {
	return newLessonServiceFixtureWithDependencies(courses, lessons, quizzes, newPracticeRepositoryFake(), clock)
}

func newLessonServiceFixtureWithPractices(
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	practices *practiceRepositoryFake,
	clock fixedClock,
) *LessonService {
	return newLessonServiceFixtureWithDependencies(courses, lessons, newQuizRepositoryFake(), practices, clock)
}

func newLessonServiceFixtureWithDependencies(
	courses *courseRepositoryFake,
	lessons *lessonRepositoryFake,
	quizzes *quizRepositoryFake,
	practices *practiceRepositoryFake,
	clock fixedClock,
) *LessonService {
	ids := fixedIDGenerator{
		courseID: mustCourseID(courseIDValue),
		lessonID: mustLessonID(lessonIDValue),
		blockID:  mustBlockID(thirdLessonIDValue),
	}
	return NewLessonService(courses, lessons, quizzes, ids, clock, practices)
}

func newLessonRepositoryFake() *lessonRepositoryFake {
	return &lessonRepositoryFake{lessons: make(map[string]domain.Lesson)}
}

func mustLesson(t *testing.T, idValue string, courseIDValue string, title string, content string, orderValue int) domain.Lesson {
	t.Helper()

	lesson, err := domain.NewLesson(
		mustLessonID(idValue),
		mustCourseID(courseIDValue),
		title,
		mustLegacyLessonBlocks(t, idValue, content),
		mustLessonOrder(orderValue),
		time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected lesson fixture, got %v", err)
	}

	return lesson
}

func mustLegacyLessonBlocks(t *testing.T, lessonIDValue string, content string) []domain.ContentBlock {
	t.Helper()

	if content == "" {
		return nil
	}

	blockID, err := domain.NewBlockID(lessonIDValue)
	if err != nil {
		t.Fatalf("expected block id fixture, got %v", err)
	}

	position, err := domain.NewBlockPosition(0)
	if err != nil {
		t.Fatalf("expected block position fixture, got %v", err)
	}

	block, err := domain.NewTextBlock(blockID, position, content)
	if err != nil {
		t.Fatalf("expected text block fixture, got %v", err)
	}

	return []domain.ContentBlock{block}
}

func mustLessonWithBlocks(t *testing.T, blocks ...domain.ContentBlock) domain.Lesson {
	t.Helper()

	return mustLessonFixture(t, lessonIDValue, courseIDValue, "First Lesson", blocks, 0)
}

func mustLessonFixture(
	t *testing.T,
	idValue string,
	courseIDValue string,
	title string,
	blocks []domain.ContentBlock,
	orderValue int,
) domain.Lesson {
	t.Helper()

	lesson, err := domain.NewLesson(
		mustLessonID(idValue),
		mustCourseID(courseIDValue),
		title,
		blocks,
		mustLessonOrder(orderValue),
		time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected lesson fixture, got %v", err)
	}

	return lesson
}

func mustTextBlock(t *testing.T, idValue string, positionValue int, markdown string) domain.ContentBlock {
	t.Helper()

	position, err := domain.NewBlockPosition(positionValue)
	if err != nil {
		t.Fatalf("expected block position fixture, got %v", err)
	}

	block, err := domain.NewTextBlock(mustBlockID(idValue), position, markdown)
	if err != nil {
		t.Fatalf("expected text block fixture, got %v", err)
	}

	return block
}

func mustVideoBlock(t *testing.T, idValue string, positionValue int, locator string, caption string) domain.ContentBlock {
	t.Helper()

	position, err := domain.NewBlockPosition(positionValue)
	if err != nil {
		t.Fatalf("expected block position fixture, got %v", err)
	}

	media, err := domain.NewMediaRef(domain.YouTubeProvider(), locator)
	if err != nil {
		t.Fatalf("expected media ref fixture, got %v", err)
	}

	block, err := domain.NewVideoBlock(mustBlockID(idValue), position, media, caption)
	if err != nil {
		t.Fatalf("expected video block fixture, got %v", err)
	}

	return block
}

func mustLessonID(value string) domain.LessonID {
	id, err := domain.NewLessonID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustBlockID(value string) domain.BlockID {
	id, err := domain.NewBlockID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustLessonOrder(value int) domain.LessonOrder {
	order, err := domain.NewLessonOrder(value)
	if err != nil {
		panic(err)
	}

	return order
}

func intPointer(value int) *int {
	return &value
}
