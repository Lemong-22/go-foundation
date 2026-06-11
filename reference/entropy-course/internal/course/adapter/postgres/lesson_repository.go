package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

var _ core.LessonRepository = (*PostgresLessonRepository)(nil)

type PostgresLessonRepository struct {
	pool    *pgxpool.Pool
	beginTx func(context.Context) (lessonTransaction, error)
}

type lessonTransaction interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

func NewPostgresLessonRepository(pool *pgxpool.Pool) *PostgresLessonRepository {
	repo := &PostgresLessonRepository{pool: pool}
	repo.beginTx = repo.beginPoolTransaction

	return repo
}

func (repo *PostgresLessonRepository) Save(lesson domain.Lesson) error {
	ctx := context.Background()
	tx, err := repo.startTransaction(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if err := upsertLesson(ctx, tx, lesson); err != nil {
		return err
	}
	if err := replaceLessonBlocks(ctx, tx, lesson); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true

	return nil
}

func (repo *PostgresLessonRepository) SaveAll(lessons []domain.Lesson) error {
	ctx := context.Background()
	tx, err := repo.startTransaction(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	for _, lesson := range lessons {
		if err := updateLesson(ctx, tx, lesson); err != nil {
			return err
		}
		if err := replaceLessonBlocks(ctx, tx, lesson); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true

	return nil
}

func (repo *PostgresLessonRepository) FindByID(id domain.LessonID) (domain.Lesson, error) {
	record, err := scanLessonRecord(repo.pool.QueryRow(
		context.Background(),
		selectLessonSQL+` WHERE id = $1`,
		id.String(),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Lesson{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Lesson{}, err
	}

	blocks, err := repo.findBlocksByLessonIDs([]domain.LessonID{id})
	if err != nil {
		return domain.Lesson{}, err
	}

	return restoreLesson(record, blocks[id])
}

func (repo *PostgresLessonRepository) FindByCourse(courseID domain.CourseID) ([]domain.Lesson, error) {
	rows, err := repo.pool.Query(
		context.Background(),
		selectLessonSQL+` WHERE course_id = $1 ORDER BY "order" ASC`,
		courseID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []lessonRecord{}
	lessonIDs := []domain.LessonID{}
	for rows.Next() {
		record, err := scanLessonRecord(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
		lessonIDs = append(lessonIDs, record.id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	blocks, err := repo.findBlocksByLessonIDs(lessonIDs)
	if err != nil {
		return nil, err
	}

	lessons := make([]domain.Lesson, 0, len(records))
	for _, record := range records {
		lesson, err := restoreLesson(record, blocks[record.id])
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

func (repo *PostgresLessonRepository) FindByBlockID(id domain.BlockID) (domain.Lesson, error) {
	record, err := scanLessonRecord(repo.pool.QueryRow(
		context.Background(),
		selectLessonByBlockSQL,
		id.String(),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Lesson{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Lesson{}, err
	}

	blocks, err := repo.findBlocksByLessonIDs([]domain.LessonID{record.id})
	if err != nil {
		return domain.Lesson{}, err
	}

	return restoreLesson(record, blocks[record.id])
}

func (repo *PostgresLessonRepository) FindLessonsEmbeddingQuiz(quizID domain.QuizID) ([]domain.LessonID, error) {
	rows, err := repo.pool.Query(
		context.Background(),
		selectLessonsEmbeddingQuizSQL,
		quizID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessonIDs := []domain.LessonID{}
	for rows.Next() {
		var idValue string
		if err := rows.Scan(&idValue); err != nil {
			return nil, err
		}

		id, err := domain.NewLessonID(idValue)
		if err != nil {
			return nil, err
		}
		lessonIDs = append(lessonIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lessonIDs, nil
}

func (repo *PostgresLessonRepository) FindLessonsEmbeddingPractice(practiceID domain.PracticeID) ([]domain.LessonID, error) {
	rows, err := repo.pool.Query(
		context.Background(),
		selectLessonsEmbeddingPracticeSQL,
		practiceID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessonIDs := []domain.LessonID{}
	for rows.Next() {
		var idValue string
		if err := rows.Scan(&idValue); err != nil {
			return nil, err
		}

		id, err := domain.NewLessonID(idValue)
		if err != nil {
			return nil, err
		}
		lessonIDs = append(lessonIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return lessonIDs, nil
}

func (repo *PostgresLessonRepository) Delete(id domain.LessonID) error {
	tag, err := repo.pool.Exec(context.Background(), `DELETE FROM lessons WHERE id = $1`, id.String())
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *PostgresLessonRepository) DeleteByCourse(courseID domain.CourseID) error {
	_, err := repo.pool.Exec(context.Background(), `DELETE FROM lessons WHERE course_id = $1`, courseID.String())
	return err
}

func (repo *PostgresLessonRepository) beginPoolTransaction(ctx context.Context) (lessonTransaction, error) {
	return repo.pool.Begin(ctx)
}

func (repo *PostgresLessonRepository) startTransaction(ctx context.Context) (lessonTransaction, error) {
	if repo.beginTx != nil {
		return repo.beginTx(ctx)
	}

	return repo.beginPoolTransaction(ctx)
}

const selectLessonSQL = `SELECT
	id,
	course_id,
	title,
	content,
	"order",
	created_at,
	updated_at
FROM lessons`

const updateLessonSQL = `UPDATE lessons SET
	course_id = $2,
	title = $3,
	content = $4,
	"order" = $5,
	updated_at = $6
WHERE id = $1`

const upsertLessonSQL = `INSERT INTO lessons (
	id,
	course_id,
	title,
	content,
	"order",
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (id) DO UPDATE SET
	course_id = EXCLUDED.course_id,
	title = EXCLUDED.title,
	content = EXCLUDED.content,
	"order" = EXCLUDED."order",
	updated_at = EXCLUDED.updated_at`

const deleteLessonBlocksSQL = `DELETE FROM content_blocks WHERE lesson_id = $1`

const insertLessonBlockSQL = `INSERT INTO content_blocks (
	id,
	lesson_id,
	kind,
	position,
	text_markdown,
	video_provider,
	video_locator,
	video_caption,
	quiz_ref,
	practice_ref
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

const selectLessonBlocksSQL = `SELECT
	id,
	lesson_id,
	kind,
	position,
	text_markdown,
	video_provider,
	video_locator,
	video_caption,
	quiz_ref::text,
	practice_ref::text
FROM content_blocks
WHERE lesson_id = ANY($1::uuid[])
ORDER BY lesson_id ASC, position ASC`

const selectLessonByBlockSQL = `SELECT
	l.id,
	l.course_id,
	l.title,
	l.content,
	l."order",
	l.created_at,
	l.updated_at
FROM lessons AS l
JOIN content_blocks b ON b.lesson_id = l.id
WHERE b.id = $1`

const selectLessonsEmbeddingQuizSQL = `SELECT DISTINCT
	l.id
FROM lessons AS l
JOIN content_blocks AS b ON b.lesson_id = l.id
WHERE b.kind = 'quiz'
	AND b.quiz_ref = $1
ORDER BY l.id ASC`

const selectLessonsEmbeddingPracticeSQL = `SELECT DISTINCT
	l.id
FROM lessons AS l
JOIN content_blocks AS b ON b.lesson_id = l.id
WHERE b.kind = 'practice'
	AND b.practice_ref = $1
ORDER BY l.id ASC`

type lessonScanner interface {
	Scan(dest ...any) error
}

type lessonRecord struct {
	id        domain.LessonID
	courseID  domain.CourseID
	title     string
	content   string
	order     domain.LessonOrder
	createdAt time.Time
	updatedAt time.Time
}

func scanLesson(scanner lessonScanner) (domain.Lesson, error) {
	record, err := scanLessonRecord(scanner)
	if err != nil {
		return domain.Lesson{}, err
	}

	blocks, err := legacyLessonBlocks(record.id, record.content)
	if err != nil {
		return domain.Lesson{}, err
	}

	return restoreLesson(record, blocks)
}

func scanLessonRecord(scanner lessonScanner) (lessonRecord, error) {
	var (
		idValue       string
		courseIDValue string
		title         string
		content       string
		orderValue    int
		createdAt     time.Time
		updatedAt     time.Time
	)

	if err := scanner.Scan(
		&idValue,
		&courseIDValue,
		&title,
		&content,
		&orderValue,
		&createdAt,
		&updatedAt,
	); err != nil {
		return lessonRecord{}, err
	}

	id, err := domain.NewLessonID(idValue)
	if err != nil {
		return lessonRecord{}, err
	}

	courseID, err := domain.NewCourseID(courseIDValue)
	if err != nil {
		return lessonRecord{}, err
	}

	order, err := domain.NewLessonOrder(orderValue)
	if err != nil {
		return lessonRecord{}, err
	}

	return lessonRecord{
		id:        id,
		courseID:  courseID,
		title:     title,
		content:   content,
		order:     order,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

func restoreLesson(record lessonRecord, blocks []domain.ContentBlock) (domain.Lesson, error) {
	return domain.RestoreLesson(
		record.id,
		record.courseID,
		record.title,
		blocks,
		record.order,
		record.createdAt,
		record.updatedAt,
	)
}

func updateLesson(ctx context.Context, tx lessonTransaction, lesson domain.Lesson) error {
	tag, err := tx.Exec(
		ctx,
		updateLessonSQL,
		lesson.ID().String(),
		lesson.CourseID().String(),
		lesson.Title(),
		lesson.Content(),
		lesson.Order().Int(),
		lesson.UpdatedAt(),
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func upsertLesson(ctx context.Context, tx lessonTransaction, lesson domain.Lesson) error {
	_, err := tx.Exec(
		ctx,
		upsertLessonSQL,
		lesson.ID().String(),
		lesson.CourseID().String(),
		lesson.Title(),
		lesson.Content(),
		lesson.Order().Int(),
		lesson.CreatedAt(),
		lesson.UpdatedAt(),
	)

	return err
}

func replaceLessonBlocks(ctx context.Context, tx lessonTransaction, lesson domain.Lesson) error {
	if _, err := tx.Exec(ctx, deleteLessonBlocksSQL, lesson.ID().String()); err != nil {
		return err
	}

	for _, block := range lesson.Blocks() {
		textMarkdown, videoProvider, videoLocator, videoCaption, quizRef, practiceRef, err := blockPayloadColumns(block)
		if err != nil {
			return err
		}

		if _, err := tx.Exec(
			ctx,
			insertLessonBlockSQL,
			block.ID().String(),
			lesson.ID().String(),
			block.Kind().String(),
			block.Position().Int(),
			textMarkdown,
			videoProvider,
			videoLocator,
			videoCaption,
			quizRef,
			practiceRef,
		); err != nil {
			return err
		}
	}

	return nil
}

func blockPayloadColumns(block domain.ContentBlock) (any, any, any, any, any, any, error) {
	switch body := block.Body().(type) {
	case domain.TextBody:
		return body.Markdown, nil, nil, nil, nil, nil, nil
	case domain.VideoBody:
		return nil, body.Media.Provider().String(), body.Media.Locator(), body.Caption, nil, nil, nil
	case domain.QuizBody:
		return nil, nil, nil, nil, body.QuizRef.String(), nil, nil
	case domain.PracticeBody:
		return nil, nil, nil, nil, nil, body.PracticeRef.String(), nil
	default:
		return nil, nil, nil, nil, nil, nil, domain.NewValidationError("body", "must be text, video, quiz, or practice")
	}
}

func (repo *PostgresLessonRepository) findBlocksByLessonIDs(lessonIDs []domain.LessonID) (map[domain.LessonID][]domain.ContentBlock, error) {
	blocks := make(map[domain.LessonID][]domain.ContentBlock, len(lessonIDs))
	if len(lessonIDs) == 0 {
		return blocks, nil
	}

	idValues := make([]string, 0, len(lessonIDs))
	for _, id := range lessonIDs {
		idValues = append(idValues, id.String())
		blocks[id] = nil
	}

	rows, err := repo.pool.Query(context.Background(), selectLessonBlocksSQL, idValues)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		lessonID, block, err := scanLessonBlock(rows)
		if err != nil {
			return nil, err
		}

		blocks[lessonID] = append(blocks[lessonID], block)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func scanLessonBlock(scanner lessonScanner) (domain.LessonID, domain.ContentBlock, error) {
	var (
		idValue       string
		lessonIDValue string
		kindValue     string
		positionValue int
		textMarkdown  pgtype.Text
		videoProvider pgtype.Text
		videoLocator  pgtype.Text
		videoCaption  pgtype.Text
		quizRef       pgtype.Text
		practiceRef   pgtype.Text
	)

	if err := scanner.Scan(
		&idValue,
		&lessonIDValue,
		&kindValue,
		&positionValue,
		&textMarkdown,
		&videoProvider,
		&videoLocator,
		&videoCaption,
		&quizRef,
		&practiceRef,
	); err != nil {
		return domain.LessonID{}, domain.ContentBlock{}, err
	}

	id, err := domain.NewBlockID(idValue)
	if err != nil {
		return domain.LessonID{}, domain.ContentBlock{}, err
	}

	lessonID, err := domain.NewLessonID(lessonIDValue)
	if err != nil {
		return domain.LessonID{}, domain.ContentBlock{}, err
	}

	kind, err := domain.NewContentBlockKind(kindValue)
	if err != nil {
		return domain.LessonID{}, domain.ContentBlock{}, err
	}

	position, err := domain.NewBlockPosition(positionValue)
	if err != nil {
		return domain.LessonID{}, domain.ContentBlock{}, err
	}

	body, err := contentBodyFromColumns(kind, textMarkdown, videoProvider, videoLocator, videoCaption, quizRef, practiceRef)
	if err != nil {
		return domain.LessonID{}, domain.ContentBlock{}, err
	}

	block, err := domain.NewContentBlock(id, kind, position, body)
	if err != nil {
		return domain.LessonID{}, domain.ContentBlock{}, err
	}

	return lessonID, block, nil
}

func contentBodyFromColumns(
	kind domain.ContentBlockKind,
	textMarkdown pgtype.Text,
	videoProvider pgtype.Text,
	videoLocator pgtype.Text,
	videoCaption pgtype.Text,
	quizRef pgtype.Text,
	practiceRef pgtype.Text,
) (domain.ContentBody, error) {
	switch {
	case kind.IsText():
		return domain.TextBody{Markdown: nullableText(textMarkdown)}, nil
	case kind.IsVideo():
		provider, err := domain.NewMediaProvider(nullableText(videoProvider))
		if err != nil {
			return nil, err
		}

		media, err := domain.NewMediaRef(provider, nullableText(videoLocator))
		if err != nil {
			return nil, err
		}

		return domain.VideoBody{Media: media, Caption: nullableText(videoCaption)}, nil
	case kind.IsQuiz():
		id, err := domain.NewQuizID(nullableText(quizRef))
		if err != nil {
			return nil, err
		}

		return domain.QuizBody{QuizRef: id}, nil
	case kind.IsPractice():
		id, err := domain.NewPracticeID(nullableText(practiceRef))
		if err != nil {
			return nil, err
		}

		return domain.PracticeBody{PracticeRef: id}, nil
	default:
		return nil, domain.NewValidationError("kind", "must be text, video, quiz, or practice")
	}
}

func nullableText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func legacyLessonBlocks(id domain.LessonID, content string) ([]domain.ContentBlock, error) {
	if content == "" {
		return nil, nil
	}

	blockID, err := domain.NewBlockID(id.String())
	if err != nil {
		return nil, err
	}

	position, err := domain.NewBlockPosition(0)
	if err != nil {
		return nil, err
	}

	block, err := domain.NewTextBlock(blockID, position, content)
	if err != nil {
		return nil, err
	}

	return []domain.ContentBlock{block}, nil
}
