package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	lessonIDValue      = "550e8400-e29b-41d4-a716-446655440020"
	otherLessonIDValue = "550e8400-e29b-41d4-a716-446655440021"
	blockIDValue       = "550e8400-e29b-41d4-a716-446655440022"
)

func TestScanLessonRestoresDomainLesson(t *testing.T) {
	createdAt := time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	lesson, err := scanLesson(lessonScannerFake{
		id:        lessonIDValue,
		courseID:  courseIDValue,
		title:     "First Lesson",
		content:   "Content",
		order:     2,
		createdAt: createdAt,
		updatedAt: updatedAt,
	})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if lesson.ID().String() != lessonIDValue || lesson.CourseID().String() != courseIDValue {
		t.Fatalf("expected persisted identity and course id to be restored")
	}
	if lesson.Title() != "First Lesson" || lesson.Content() != "Content" || lesson.Order().Int() != 2 {
		t.Fatalf("expected persisted lesson fields to be restored")
	}
	if !lesson.CreatedAt().Equal(createdAt) || !lesson.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected persisted timestamps to be restored")
	}
}

func TestScanLessonBlockRestoresTextBlock(t *testing.T) {
	lessonID, block, err := scanLessonBlock(valueScannerFake{values: []any{
		blockIDValue,
		lessonIDValue,
		"text",
		0,
		"Markdown",
		nil,
		nil,
		nil,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if lessonID.String() != lessonIDValue || block.ID().String() != blockIDValue || block.Kind() != domain.TextKind() {
		t.Fatalf("expected text block identity to be restored")
	}
	if body := block.Body().(domain.TextBody); body.Markdown != "Markdown" {
		t.Fatalf("expected text markdown to be restored, got %q", body.Markdown)
	}
}

func TestScanLessonBlockRestoresVideoBlock(t *testing.T) {
	lessonID, block, err := scanLessonBlock(valueScannerFake{values: []any{
		blockIDValue,
		lessonIDValue,
		"video",
		1,
		nil,
		"youtube",
		"https://youtu.be/dQw4w9WgXcQ",
		"Intro video",
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if lessonID.String() != lessonIDValue || block.Kind() != domain.VideoKind() || block.Position().Int() != 1 {
		t.Fatalf("expected video block fields to be restored")
	}
	body := block.Body().(domain.VideoBody)
	if body.Media.Provider() != domain.YouTubeProvider() || body.Media.Locator() != "https://youtu.be/dQw4w9WgXcQ" || body.Caption != "Intro video" {
		t.Fatalf("expected video payload to be restored, got %+v", body)
	}
}

func TestScanLessonBlockRestoresQuizBlock(t *testing.T) {
	lessonID, block, err := scanLessonBlock(valueScannerFake{values: []any{
		blockIDValue,
		lessonIDValue,
		"quiz",
		2,
		nil,
		nil,
		nil,
		nil,
		postgresQuizIDValue,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if lessonID.String() != lessonIDValue || block.Kind() != domain.QuizKind() || block.Position().Int() != 2 {
		t.Fatalf("expected quiz block fields to be restored")
	}
	body := block.Body().(domain.QuizBody)
	if body.QuizRef.String() != postgresQuizIDValue {
		t.Fatalf("expected quiz ref %q, got %q", postgresQuizIDValue, body.QuizRef.String())
	}
}

func TestScanLessonBlockRestoresPracticeBlock(t *testing.T) {
	lessonID, block, err := scanLessonBlock(valueScannerFake{values: []any{
		blockIDValue,
		lessonIDValue,
		"practice",
		3,
		nil,
		nil,
		nil,
		nil,
		nil,
		postgresPracticeIDValue,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if lessonID.String() != lessonIDValue || block.Kind() != domain.PracticeKind() || block.Position().Int() != 3 {
		t.Fatalf("expected practice block fields to be restored")
	}
	body := block.Body().(domain.PracticeBody)
	if body.PracticeRef.String() != postgresPracticeIDValue {
		t.Fatalf("expected practice ref %q, got %q", postgresPracticeIDValue, body.PracticeRef.String())
	}
}

func TestSaveCommitsLessonAndBlockReplacementInOneTransaction(t *testing.T) {
	tx := &lessonTransactionFake{}
	repo := newPostgresLessonRepositoryWithTransaction(tx)

	err := repo.Save(mustPostgresLessonWithBlocks(t,
		lessonIDValue,
		1,
		mustPostgresTextBlock(t, lessonIDValue, 0, "Intro"),
		mustPostgresVideoBlock(t, blockIDValue, 1, "https://youtu.be/dQw4w9WgXcQ", "Watch"),
		mustPostgresQuizBlock(t, otherLessonIDValue, 2, postgresQuizIDValue),
		mustPostgresPracticeBlock(t, postgresOtherPracticeID, 3, postgresPracticeIDValue),
	))
	if err != nil {
		t.Fatalf("expected save to succeed, got %v", err)
	}

	if tx.execCalls != 6 {
		t.Fatalf("expected upsert, delete, and four block inserts, got %d calls", tx.execCalls)
	}
	if !strings.Contains(strings.ToLower(tx.calls[0].sql), "insert into lessons") {
		t.Fatalf("expected first call to upsert lesson, got %q", tx.calls[0].sql)
	}
	if !strings.Contains(strings.ToLower(tx.calls[1].sql), "delete from content_blocks") {
		t.Fatalf("expected second call to delete old blocks, got %q", tx.calls[1].sql)
	}
	if !strings.Contains(strings.ToLower(tx.calls[2].sql), "insert into content_blocks") ||
		!strings.Contains(strings.ToLower(tx.calls[3].sql), "insert into content_blocks") ||
		!strings.Contains(strings.ToLower(tx.calls[4].sql), "insert into content_blocks") ||
		!strings.Contains(strings.ToLower(tx.calls[5].sql), "insert into content_blocks") {
		t.Fatalf("expected remaining calls to insert blocks")
	}
	if tx.calls[2].args[4] != "Intro" || tx.calls[2].args[5] != nil {
		t.Fatalf("expected text block payload columns, got %+v", tx.calls[2].args)
	}
	if tx.calls[3].args[4] != nil || tx.calls[3].args[5] != "youtube" || tx.calls[3].args[6] != "https://youtu.be/dQw4w9WgXcQ" || tx.calls[3].args[7] != "Watch" {
		t.Fatalf("expected video block payload columns, got %+v", tx.calls[3].args)
	}
	if tx.calls[4].args[4] != nil || tx.calls[4].args[5] != nil || tx.calls[4].args[8] != postgresQuizIDValue {
		t.Fatalf("expected quiz block payload columns, got %+v", tx.calls[4].args)
	}
	if tx.calls[5].args[4] != nil || tx.calls[5].args[8] != nil || tx.calls[5].args[9] != postgresPracticeIDValue {
		t.Fatalf("expected practice block payload columns, got %+v", tx.calls[5].args)
	}
	if !tx.committed || tx.rolledBack {
		t.Fatalf("expected successful save to commit without rollback")
	}
}

func TestSaveRollsBackWhenReplacingBlocksFails(t *testing.T) {
	errBoom := errors.New("insert block failed")
	tx := &lessonTransactionFake{execErrs: []error{nil, nil, errBoom}}
	repo := newPostgresLessonRepositoryWithTransaction(tx)

	err := repo.Save(mustPostgresLesson(t, lessonIDValue, 1))
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected block insert error, got %v", err)
	}
	if tx.committed {
		t.Fatalf("did not expect failed save to commit")
	}
	if !tx.rolledBack {
		t.Fatalf("expected failed save to roll back")
	}
}

func TestSaveAllCommitsLessonUpdatesInOneTransaction(t *testing.T) {
	tx := &lessonTransactionFake{}
	repo := newPostgresLessonRepositoryWithTransaction(tx)

	err := repo.SaveAll([]domain.Lesson{
		mustPostgresLesson(t, lessonIDValue, 1),
		mustPostgresLesson(t, otherLessonIDValue, 2),
	})
	if err != nil {
		t.Fatalf("expected save all to succeed, got %v", err)
	}

	if tx.execCalls != 6 {
		t.Fatalf("expected two lesson updates with block replacement calls, got %d", tx.execCalls)
	}
	if !tx.committed {
		t.Fatalf("expected transaction to commit")
	}
	if tx.rolledBack {
		t.Fatalf("did not expect successful transaction to roll back")
	}
}

func TestFindByBlockSQLUsesContentBlocksJoin(t *testing.T) {
	sql := strings.ToLower(selectLessonByBlockSQL)
	if !strings.Contains(sql, "join content_blocks") || !strings.Contains(sql, "where b.id = $1") {
		t.Fatalf("expected block lookup SQL to join content_blocks by block id, got %q", selectLessonByBlockSQL)
	}
}

func TestFindLessonsEmbeddingQuizSQLUsesRestrictLookup(t *testing.T) {
	sql := strings.ToLower(selectLessonsEmbeddingQuizSQL)
	if !strings.Contains(sql, "join content_blocks") ||
		!strings.Contains(sql, "b.kind = 'quiz'") ||
		!strings.Contains(sql, "b.quiz_ref = $1") {
		t.Fatalf("expected quiz embedding lookup SQL to filter quiz blocks by quiz_ref, got %q", selectLessonsEmbeddingQuizSQL)
	}
}

func TestFindLessonsEmbeddingPracticeSQLUsesRestrictLookup(t *testing.T) {
	sql := strings.ToLower(selectLessonsEmbeddingPracticeSQL)
	if !strings.Contains(sql, "join content_blocks") ||
		!strings.Contains(sql, "b.kind = 'practice'") ||
		!strings.Contains(sql, "b.practice_ref = $1") {
		t.Fatalf("expected practice embedding lookup SQL to filter practice blocks by practice_ref, got %q", selectLessonsEmbeddingPracticeSQL)
	}
}

func TestSaveAllRollsBackOnExecError(t *testing.T) {
	errBoom := errors.New("update failed")
	tx := &lessonTransactionFake{execErrs: []error{errBoom}}
	repo := newPostgresLessonRepositoryWithTransaction(tx)

	err := repo.SaveAll([]domain.Lesson{mustPostgresLesson(t, lessonIDValue, 1)})
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected update error, got %v", err)
	}
	if tx.committed {
		t.Fatalf("did not expect failed transaction to commit")
	}
	if !tx.rolledBack {
		t.Fatalf("expected failed transaction to roll back")
	}
}

func TestSaveAllRollsBackWhenLessonIsMissing(t *testing.T) {
	tx := &lessonTransactionFake{execTags: []pgconn.CommandTag{pgconn.NewCommandTag("UPDATE 0")}}
	repo := newPostgresLessonRepositoryWithTransaction(tx)

	err := repo.SaveAll([]domain.Lesson{mustPostgresLesson(t, lessonIDValue, 1)})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if tx.committed {
		t.Fatalf("did not expect failed transaction to commit")
	}
	if !tx.rolledBack {
		t.Fatalf("expected failed transaction to roll back")
	}
}

type lessonScannerFake struct {
	id        string
	courseID  string
	title     string
	content   string
	order     int
	createdAt time.Time
	updatedAt time.Time
	err       error
}

func (scanner lessonScannerFake) Scan(dest ...any) error {
	if scanner.err != nil {
		return scanner.err
	}

	values := []any{
		scanner.id,
		scanner.courseID,
		scanner.title,
		scanner.content,
		scanner.order,
		scanner.createdAt,
		scanner.updatedAt,
	}

	for i, value := range values {
		switch target := dest[i].(type) {
		case *string:
			*target = value.(string)
		case *int:
			*target = value.(int)
		case *time.Time:
			*target = value.(time.Time)
		}
	}

	return nil
}

type lessonTransactionFake struct {
	execTags []pgconn.CommandTag
	execErrs []error

	execCalls  int
	committed  bool
	rolledBack bool
	calls      []lessonExecCall
}

type lessonExecCall struct {
	sql  string
	args []any
}

func (tx *lessonTransactionFake) Exec(
	_ context.Context,
	sql string,
	args ...any,
) (pgconn.CommandTag, error) {
	callIndex := tx.execCalls
	tx.execCalls++
	tx.calls = append(tx.calls, lessonExecCall{sql: sql, args: args})

	if callIndex < len(tx.execErrs) && tx.execErrs[callIndex] != nil {
		return pgconn.CommandTag{}, tx.execErrs[callIndex]
	}

	if callIndex < len(tx.execTags) {
		return tx.execTags[callIndex], nil
	}

	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (tx *lessonTransactionFake) Commit(context.Context) error {
	tx.committed = true
	return nil
}

func (tx *lessonTransactionFake) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

func newPostgresLessonRepositoryWithTransaction(tx lessonTransaction) *PostgresLessonRepository {
	return &PostgresLessonRepository{
		beginTx: func(context.Context) (lessonTransaction, error) {
			return tx, nil
		},
	}
}

type valueScannerFake struct {
	values []any
	err    error
}

func (scanner valueScannerFake) Scan(dest ...any) error {
	if scanner.err != nil {
		return scanner.err
	}

	for i, value := range scanner.values {
		switch target := dest[i].(type) {
		case *string:
			*target = value.(string)
		case *int:
			*target = value.(int)
		case *float64:
			*target = value.(float64)
		case *time.Time:
			*target = value.(time.Time)
		case *pgtype.Int4:
			switch typed := value.(type) {
			case nil:
				*target = pgtype.Int4{}
			case int:
				*target = pgtype.Int4{Int32: int32(typed), Valid: true}
			case int32:
				*target = pgtype.Int4{Int32: typed, Valid: true}
			case pgtype.Int4:
				*target = typed
			}
		case *pgtype.Text:
			switch typed := value.(type) {
			case nil:
				*target = pgtype.Text{}
			case string:
				*target = pgtype.Text{String: typed, Valid: true}
			case pgtype.Text:
				*target = typed
			}
		}
	}

	return nil
}

func mustPostgresLesson(t *testing.T, idValue string, orderValue int) domain.Lesson {
	t.Helper()

	return mustPostgresLessonWithBlocks(
		t,
		idValue,
		orderValue,
		mustPostgresTextBlock(t, idValue, 0, "Content"),
	)
}

func mustPostgresLessonWithBlocks(
	t *testing.T,
	idValue string,
	orderValue int,
	blocks ...domain.ContentBlock,
) domain.Lesson {
	t.Helper()

	lesson, err := domain.NewLesson(
		mustPostgresLessonID(idValue),
		mustPostgresCourseID(courseIDValue),
		"First Lesson",
		blocks,
		mustPostgresLessonOrder(orderValue),
		time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected lesson fixture, got %v", err)
	}

	return lesson
}

func mustPostgresTextBlock(t *testing.T, idValue string, positionValue int, markdown string) domain.ContentBlock {
	t.Helper()

	position, err := domain.NewBlockPosition(positionValue)
	if err != nil {
		t.Fatalf("expected block position fixture, got %v", err)
	}

	block, err := domain.NewTextBlock(mustPostgresBlockID(idValue), position, markdown)
	if err != nil {
		t.Fatalf("expected text block fixture, got %v", err)
	}

	return block
}

func mustPostgresVideoBlock(t *testing.T, idValue string, positionValue int, locator string, caption string) domain.ContentBlock {
	t.Helper()

	position, err := domain.NewBlockPosition(positionValue)
	if err != nil {
		t.Fatalf("expected block position fixture, got %v", err)
	}

	media, err := domain.NewMediaRef(domain.YouTubeProvider(), locator)
	if err != nil {
		t.Fatalf("expected media ref fixture, got %v", err)
	}

	block, err := domain.NewVideoBlock(mustPostgresBlockID(idValue), position, media, caption)
	if err != nil {
		t.Fatalf("expected video block fixture, got %v", err)
	}

	return block
}

func mustPostgresQuizBlock(t *testing.T, idValue string, positionValue int, quizIDValue string) domain.ContentBlock {
	t.Helper()

	position, err := domain.NewBlockPosition(positionValue)
	if err != nil {
		t.Fatalf("expected block position fixture, got %v", err)
	}

	block, err := domain.NewQuizBlock(mustPostgresBlockID(idValue), position, mustPostgresQuizID(quizIDValue))
	if err != nil {
		t.Fatalf("expected quiz block fixture, got %v", err)
	}

	return block
}

func mustPostgresPracticeBlock(t *testing.T, idValue string, positionValue int, practiceIDValue string) domain.ContentBlock {
	t.Helper()

	position, err := domain.NewBlockPosition(positionValue)
	if err != nil {
		t.Fatalf("expected block position fixture, got %v", err)
	}

	block, err := domain.NewPracticeBlock(mustPostgresBlockID(idValue), position, mustPostgresPracticeID(practiceIDValue))
	if err != nil {
		t.Fatalf("expected practice block fixture, got %v", err)
	}

	return block
}

func mustPostgresLessonID(value string) domain.LessonID {
	id, err := domain.NewLessonID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustPostgresBlockID(value string) domain.BlockID {
	id, err := domain.NewBlockID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustPostgresCourseID(value string) domain.CourseID {
	id, err := domain.NewCourseID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustPostgresQuizID(value string) domain.QuizID {
	id, err := domain.NewQuizID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustPostgresLessonOrder(value int) domain.LessonOrder {
	order, err := domain.NewLessonOrder(value)
	if err != nil {
		panic(err)
	}

	return order
}
