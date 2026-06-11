package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

const (
	anotherValidUUID = "550e8400-e29b-41d4-a716-446655440001"
	blockIDValue     = "550e8400-e29b-41d4-a716-446655440002"
	otherBlockID     = "550e8400-e29b-41d4-a716-446655440003"
	thirdBlockID     = "550e8400-e29b-41d4-a716-446655440004"
)

func TestNewCourseCreatesDraftCourse(t *testing.T) {
	courseID := mustCourseID(t, validUUID)
	instructorID := mustInstructorID(t, anotherValidUUID)
	slug := mustSlug(t, "intro-to-go")
	now := time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC)

	course, err := NewCourse(courseID, "  Intro to Go  ", slug, "Learn Go", instructorID, now)
	if err != nil {
		t.Fatalf("expected course, got error %v", err)
	}

	if course.ID() != courseID {
		t.Fatalf("expected course id %q, got %q", courseID, course.ID())
	}
	if course.Title() != "Intro to Go" {
		t.Fatalf("expected trimmed title, got %q", course.Title())
	}
	if course.Slug() != slug {
		t.Fatalf("expected slug %q, got %q", slug, course.Slug())
	}
	if course.Description() != "Learn Go" {
		t.Fatalf("expected description, got %q", course.Description())
	}
	if course.InstructorID() != instructorID {
		t.Fatalf("expected instructor id %q, got %q", instructorID, course.InstructorID())
	}
	if course.Status() != Draft() {
		t.Fatalf("expected draft status, got %q", course.Status())
	}
	if !course.CreatedAt().Equal(now) || !course.UpdatedAt().Equal(now) {
		t.Fatalf("expected created and updated timestamps to equal %v", now)
	}
}

func TestNewCourseRejectsEmptyTitle(t *testing.T) {
	_, err := NewCourse(
		mustCourseID(t, validUUID),
		"   ",
		mustSlug(t, "intro-to-go"),
		"",
		mustInstructorID(t, anotherValidUUID),
		time.Now(),
	)

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRestoreCourseRehydratesPersistedState(t *testing.T) {
	createdAt := time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	course, err := RestoreCourse(
		mustCourseID(t, validUUID),
		"  Intro to Go  ",
		mustSlug(t, "intro-to-go"),
		"Learn Go",
		mustInstructorID(t, anotherValidUUID),
		Published(),
		createdAt,
		updatedAt,
	)
	if err != nil {
		t.Fatalf("expected restored course, got %v", err)
	}

	if course.Title() != "Intro to Go" || course.Status() != Published() {
		t.Fatalf("expected persisted title and status to be restored")
	}
	if !course.CreatedAt().Equal(createdAt) || !course.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected persisted timestamps to be restored")
	}
}

func TestCourseMutationsUpdateStateAndTimestamp(t *testing.T) {
	course := mustCourse(t)
	changedAt := course.CreatedAt().Add(time.Hour)
	newSlug := mustSlug(t, "advanced-go")

	if err := course.Rename("  Advanced Go  ", changedAt); err != nil {
		t.Fatalf("expected rename to succeed, got %v", err)
	}
	if course.Title() != "Advanced Go" || !course.UpdatedAt().Equal(changedAt) {
		t.Fatalf("rename did not update title and timestamp")
	}

	descriptionChangedAt := changedAt.Add(time.Hour)
	course.ChangeDescription("Deeper Go", descriptionChangedAt)
	if course.Description() != "Deeper Go" || !course.UpdatedAt().Equal(descriptionChangedAt) {
		t.Fatalf("description change did not update state and timestamp")
	}

	slugChangedAt := descriptionChangedAt.Add(time.Hour)
	course.ChangeSlug(newSlug, slugChangedAt)
	if course.Slug() != newSlug || !course.UpdatedAt().Equal(slugChangedAt) {
		t.Fatalf("slug change did not update state and timestamp")
	}
}

func TestCourseRenameRejectsEmptyTitleWithoutChangingTimestamp(t *testing.T) {
	course := mustCourse(t)
	updatedAt := course.UpdatedAt()

	err := course.Rename("   ", updatedAt.Add(time.Hour))
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if course.Title() != "Intro to Go" {
		t.Fatalf("expected title to remain unchanged, got %q", course.Title())
	}
	if !course.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updated timestamp to remain unchanged")
	}
}

func TestCoursePublishAndUnpublish(t *testing.T) {
	course := mustCourse(t)
	publishedAt := course.CreatedAt().Add(time.Hour)

	if err := course.Publish(publishedAt); err != nil {
		t.Fatalf("expected publish to succeed, got %v", err)
	}
	if course.Status() != Published() || !course.UpdatedAt().Equal(publishedAt) {
		t.Fatalf("publish did not update status and timestamp")
	}

	if err := course.Publish(publishedAt.Add(time.Hour)); !errors.Is(err, ErrAlreadyPublished) {
		t.Fatalf("expected already published error, got %v", err)
	}

	unpublishedAt := publishedAt.Add(2 * time.Hour)
	if err := course.Unpublish(unpublishedAt); err != nil {
		t.Fatalf("expected unpublish to succeed, got %v", err)
	}
	if course.Status() != Draft() || !course.UpdatedAt().Equal(unpublishedAt) {
		t.Fatalf("unpublish did not update status and timestamp")
	}

	if err := course.Unpublish(unpublishedAt.Add(time.Hour)); !errors.Is(err, ErrNotPublished) {
		t.Fatalf("expected not published error, got %v", err)
	}
}

func TestNewLessonCreatesLesson(t *testing.T) {
	lessonID := mustLessonID(t, validUUID)
	courseID := mustCourseID(t, anotherValidUUID)
	order := mustLessonOrder(t, 2)
	now := time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC)
	blocks := []ContentBlock{mustTextBlock(t, blockIDValue, 0, "Content")}

	lesson, err := NewLesson(lessonID, courseID, "  First Lesson  ", blocks, order, now)
	if err != nil {
		t.Fatalf("expected lesson, got error %v", err)
	}

	if lesson.ID() != lessonID {
		t.Fatalf("expected lesson id %q, got %q", lessonID, lesson.ID())
	}
	if lesson.CourseID() != courseID {
		t.Fatalf("expected course id %q, got %q", courseID, lesson.CourseID())
	}
	if lesson.Title() != "First Lesson" {
		t.Fatalf("expected trimmed title, got %q", lesson.Title())
	}
	if got := lesson.Blocks(); !reflect.DeepEqual(got, blocks) {
		t.Fatalf("expected blocks %+v, got %+v", blocks, got)
	}
	if lesson.Content() != "Content" {
		t.Fatalf("expected compatibility content, got %q", lesson.Content())
	}
	if lesson.Order() != order {
		t.Fatalf("expected order %d, got %d", order.Int(), lesson.Order().Int())
	}
	if !lesson.CreatedAt().Equal(now) || !lesson.UpdatedAt().Equal(now) {
		t.Fatalf("expected created and updated timestamps to equal %v", now)
	}
}

func TestNewLessonRejectsEmptyTitle(t *testing.T) {
	_, err := NewLesson(
		mustLessonID(t, validUUID),
		mustCourseID(t, anotherValidUUID),
		"   ",
		nil,
		mustLessonOrder(t, 0),
		time.Now(),
	)

	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRestoreLessonRehydratesPersistedState(t *testing.T) {
	createdAt := time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	lesson, err := RestoreLesson(
		mustLessonID(t, validUUID),
		mustCourseID(t, anotherValidUUID),
		"  First Lesson  ",
		[]ContentBlock{mustTextBlock(t, blockIDValue, 0, "Content")},
		mustLessonOrder(t, 4),
		createdAt,
		updatedAt,
	)
	if err != nil {
		t.Fatalf("expected restored lesson, got %v", err)
	}

	if lesson.Title() != "First Lesson" || lesson.Order().Int() != 4 {
		t.Fatalf("expected persisted title and order to be restored")
	}
	if lesson.Content() != "Content" {
		t.Fatalf("expected restored blocks to expose compatibility content")
	}
	if !lesson.CreatedAt().Equal(createdAt) || !lesson.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected persisted timestamps to be restored")
	}
}

func TestLessonMutationsUpdateStateAndTimestamp(t *testing.T) {
	lesson := mustLesson(t)
	renamedAt := lesson.CreatedAt().Add(time.Hour)

	if err := lesson.Rename("  Intro  ", renamedAt); err != nil {
		t.Fatalf("expected rename to succeed, got %v", err)
	}
	if lesson.Title() != "Intro" || !lesson.UpdatedAt().Equal(renamedAt) {
		t.Fatalf("rename did not update title and timestamp")
	}

	movedAt := renamedAt.Add(time.Hour)
	newOrder := mustLessonOrder(t, 4)
	lesson.MoveTo(newOrder, movedAt)
	if lesson.Order() != newOrder || !lesson.UpdatedAt().Equal(movedAt) {
		t.Fatalf("move did not update order and timestamp")
	}
}

func TestNewLessonRejectsInvalidBlockSets(t *testing.T) {
	tests := []struct {
		name   string
		blocks []ContentBlock
	}{
		{
			name: "duplicate ids",
			blocks: []ContentBlock{
				mustTextBlock(t, blockIDValue, 0, "One"),
				mustTextBlock(t, blockIDValue, 1, "Two"),
			},
		},
		{
			name: "duplicate positions",
			blocks: []ContentBlock{
				mustTextBlock(t, blockIDValue, 0, "One"),
				mustTextBlock(t, otherBlockID, 0, "Two"),
			},
		},
		{
			name: "position gap",
			blocks: []ContentBlock{
				mustTextBlock(t, blockIDValue, 0, "One"),
				mustTextBlock(t, otherBlockID, 2, "Two"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewLesson(
				mustLessonID(t, validUUID),
				mustCourseID(t, anotherValidUUID),
				"First Lesson",
				test.blocks,
				mustLessonOrder(t, 0),
				time.Now(),
			)
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestLessonAddBlockInsertsAndShiftsPositions(t *testing.T) {
	lesson := mustLessonWithBlocks(t,
		mustTextBlock(t, blockIDValue, 0, "One"),
		mustTextBlock(t, otherBlockID, 1, "Two"),
	)
	inserted := mustVideoBlock(t, thirdBlockID, 1, "https://youtu.be/dQw4w9WgXcQ")
	changedAt := lesson.CreatedAt().Add(time.Hour)

	if err := lesson.AddBlock(inserted, changedAt); err != nil {
		t.Fatalf("expected add block to succeed, got %v", err)
	}

	blocks := lesson.Blocks()
	if len(blocks) != 3 {
		t.Fatalf("expected three blocks, got %d", len(blocks))
	}
	if blocks[0].ID().String() != blockIDValue || blocks[0].Position().Int() != 0 {
		t.Fatalf("expected original first block to stay at position 0")
	}
	if blocks[1].ID().String() != thirdBlockID || blocks[1].Position().Int() != 1 {
		t.Fatalf("expected inserted block at position 1, got %+v", blocks[1])
	}
	if blocks[2].ID().String() != otherBlockID || blocks[2].Position().Int() != 2 {
		t.Fatalf("expected original second block to shift to position 2")
	}
	if !lesson.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp to change")
	}
}

func TestLessonAddBlockRejectsDuplicateIDAndOutOfRangePosition(t *testing.T) {
	lesson := mustLessonWithBlocks(t, mustTextBlock(t, blockIDValue, 0, "One"))

	if err := lesson.AddBlock(mustTextBlock(t, blockIDValue, 1, "Duplicate"), time.Now()); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected duplicate id validation error, got %v", err)
	}

	if err := lesson.AddBlock(mustTextBlock(t, otherBlockID, 2, "Gap"), time.Now()); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected out-of-range position validation error, got %v", err)
	}
}

func TestLessonUpdateBlockPreservesKindAndTimestamp(t *testing.T) {
	lesson := mustLessonWithBlocks(t, mustTextBlock(t, blockIDValue, 0, "One"))
	changedAt := lesson.CreatedAt().Add(time.Hour)

	err := lesson.UpdateBlock(mustBlockID(t, blockIDValue), TextBody{Markdown: "Updated"}, changedAt)
	if err != nil {
		t.Fatalf("expected update to succeed, got %v", err)
	}

	block, err := lesson.Block(mustBlockID(t, blockIDValue))
	if err != nil {
		t.Fatalf("expected block to exist, got %v", err)
	}
	body := block.Body().(TextBody)
	if body.Markdown != "Updated" {
		t.Fatalf("expected updated markdown, got %q", body.Markdown)
	}
	if !lesson.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}

	media := mustMediaRef(t, URLProvider(), "https://example.com/video.mp4")
	err = lesson.UpdateBlock(mustBlockID(t, blockIDValue), VideoBody{Media: media}, changedAt.Add(time.Hour))
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected kind/body mismatch validation error, got %v", err)
	}
}

func TestLessonRemoveBlockCompactsPositions(t *testing.T) {
	lesson := mustLessonWithBlocks(t,
		mustTextBlock(t, blockIDValue, 0, "One"),
		mustTextBlock(t, otherBlockID, 1, "Two"),
		mustTextBlock(t, thirdBlockID, 2, "Three"),
	)
	changedAt := lesson.CreatedAt().Add(time.Hour)

	if err := lesson.RemoveBlock(mustBlockID(t, otherBlockID), changedAt); err != nil {
		t.Fatalf("expected remove to succeed, got %v", err)
	}

	blocks := lesson.Blocks()
	if len(blocks) != 2 {
		t.Fatalf("expected two blocks, got %d", len(blocks))
	}
	if blocks[0].ID().String() != blockIDValue || blocks[0].Position().Int() != 0 {
		t.Fatalf("expected first block at position 0")
	}
	if blocks[1].ID().String() != thirdBlockID || blocks[1].Position().Int() != 1 {
		t.Fatalf("expected third block compacted to position 1")
	}
	if !lesson.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}
}

func TestLessonReorderBlocksRequiresPermutation(t *testing.T) {
	lesson := mustLessonWithBlocks(t,
		mustTextBlock(t, blockIDValue, 0, "One"),
		mustTextBlock(t, otherBlockID, 1, "Two"),
	)
	changedAt := lesson.CreatedAt().Add(time.Hour)

	err := lesson.ReorderBlocks([]BlockPlacement{
		{BlockID: mustBlockID(t, otherBlockID), Position: mustBlockPosition(t, 0)},
		{BlockID: mustBlockID(t, blockIDValue), Position: mustBlockPosition(t, 1)},
	}, changedAt)
	if err != nil {
		t.Fatalf("expected reorder to succeed, got %v", err)
	}

	blocks := lesson.Blocks()
	if blocks[0].ID().String() != otherBlockID || blocks[1].ID().String() != blockIDValue {
		t.Fatalf("expected blocks to reorder by position")
	}
	if !lesson.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}

	tests := []struct {
		name  string
		order []BlockPlacement
	}{
		{
			name:  "missing block",
			order: []BlockPlacement{{BlockID: mustBlockID(t, blockIDValue), Position: mustBlockPosition(t, 0)}},
		},
		{
			name: "unknown block",
			order: []BlockPlacement{
				{BlockID: mustBlockID(t, blockIDValue), Position: mustBlockPosition(t, 0)},
				{BlockID: mustBlockID(t, thirdBlockID), Position: mustBlockPosition(t, 1)},
			},
		},
		{
			name: "duplicate position",
			order: []BlockPlacement{
				{BlockID: mustBlockID(t, blockIDValue), Position: mustBlockPosition(t, 0)},
				{BlockID: mustBlockID(t, otherBlockID), Position: mustBlockPosition(t, 0)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			lesson := mustLessonWithBlocks(t,
				mustTextBlock(t, blockIDValue, 0, "One"),
				mustTextBlock(t, otherBlockID, 1, "Two"),
			)
			if err := lesson.ReorderBlocks(test.order, changedAt); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestLessonRenameRejectsEmptyTitleWithoutChangingTimestamp(t *testing.T) {
	lesson := mustLesson(t)
	updatedAt := lesson.UpdatedAt()

	err := lesson.Rename("   ", updatedAt.Add(time.Hour))
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}

	if lesson.Title() != "First Lesson" {
		t.Fatalf("expected title to remain unchanged, got %q", lesson.Title())
	}
	if !lesson.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected updated timestamp to remain unchanged")
	}
}

func TestDeferredLessonStatusRemainsAbsent(t *testing.T) {
	lessonType := reflect.TypeOf(Lesson{})

	if _, exists := lessonType.MethodByName("Status"); exists {
		t.Fatalf("expected deferred lesson status behavior to remain absent")
	}
}

func mustCourse(t *testing.T) Course {
	t.Helper()

	course, err := NewCourse(
		mustCourseID(t, validUUID),
		"Intro to Go",
		mustSlug(t, "intro-to-go"),
		"Learn Go",
		mustInstructorID(t, anotherValidUUID),
		time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected course fixture, got error %v", err)
	}

	return course
}

func mustLesson(t *testing.T) Lesson {
	t.Helper()

	lesson, err := NewLesson(
		mustLessonID(t, validUUID),
		mustCourseID(t, anotherValidUUID),
		"First Lesson",
		[]ContentBlock{mustTextBlock(t, blockIDValue, 0, "Content")},
		mustLessonOrder(t, 1),
		time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected lesson fixture, got error %v", err)
	}

	return lesson
}

func mustCourseID(t *testing.T, value string) CourseID {
	t.Helper()

	id, err := NewCourseID(value)
	if err != nil {
		t.Fatalf("expected course id fixture, got error %v", err)
	}

	return id
}

func mustLessonID(t *testing.T, value string) LessonID {
	t.Helper()

	id, err := NewLessonID(value)
	if err != nil {
		t.Fatalf("expected lesson id fixture, got error %v", err)
	}

	return id
}

func mustInstructorID(t *testing.T, value string) InstructorID {
	t.Helper()

	id, err := NewInstructorID(value)
	if err != nil {
		t.Fatalf("expected instructor id fixture, got error %v", err)
	}

	return id
}

func mustSlug(t *testing.T, value string) Slug {
	t.Helper()

	slug, err := NewSlug(value)
	if err != nil {
		t.Fatalf("expected slug fixture, got error %v", err)
	}

	return slug
}

func mustLessonOrder(t *testing.T, value int) LessonOrder {
	t.Helper()

	order, err := NewLessonOrder(value)
	if err != nil {
		t.Fatalf("expected lesson order fixture, got error %v", err)
	}

	return order
}

func mustLessonWithBlocks(t *testing.T, blocks ...ContentBlock) Lesson {
	t.Helper()

	lesson, err := NewLesson(
		mustLessonID(t, validUUID),
		mustCourseID(t, anotherValidUUID),
		"First Lesson",
		blocks,
		mustLessonOrder(t, 1),
		time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected lesson fixture, got error %v", err)
	}

	return lesson
}

func mustBlockID(t *testing.T, value string) BlockID {
	t.Helper()

	id, err := NewBlockID(value)
	if err != nil {
		t.Fatalf("expected block id fixture, got error %v", err)
	}

	return id
}

func mustBlockPosition(t *testing.T, value int) BlockPosition {
	t.Helper()

	position, err := NewBlockPosition(value)
	if err != nil {
		t.Fatalf("expected block position fixture, got error %v", err)
	}

	return position
}

func mustTextBlock(t *testing.T, id string, position int, markdown string) ContentBlock {
	t.Helper()

	block, err := NewTextBlock(mustBlockID(t, id), mustBlockPosition(t, position), markdown)
	if err != nil {
		t.Fatalf("expected text block fixture, got error %v", err)
	}

	return block
}

func mustVideoBlock(t *testing.T, id string, position int, locator string) ContentBlock {
	t.Helper()

	block, err := NewVideoBlock(
		mustBlockID(t, id),
		mustBlockPosition(t, position),
		mustMediaRef(t, YouTubeProvider(), locator),
		"Watch this",
	)
	if err != nil {
		t.Fatalf("expected video block fixture, got error %v", err)
	}

	return block
}

func mustMediaRef(t *testing.T, provider MediaProvider, locator string) MediaRef {
	t.Helper()

	ref, err := NewMediaRef(provider, locator)
	if err != nil {
		t.Fatalf("expected media ref fixture, got error %v", err)
	}

	return ref
}
