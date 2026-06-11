package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

var _ core.QuizRepository = (*PostgresQuizRepository)(nil)

type PostgresQuizRepository struct {
	pool    *pgxpool.Pool
	beginTx func(context.Context) (quizTransaction, error)
}

type quizTransaction interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

func NewPostgresQuizRepository(pool *pgxpool.Pool) *PostgresQuizRepository {
	repo := &PostgresQuizRepository{pool: pool}
	repo.beginTx = repo.beginPoolTransaction

	return repo
}

func (repo *PostgresQuizRepository) Save(quiz domain.Quiz) error {
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

	if err := upsertQuiz(ctx, tx, quiz); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, deleteQuizQuestionsSQL, quiz.ID().String()); err != nil {
		return err
	}
	for _, question := range quiz.Questions() {
		if err := insertQuizQuestion(ctx, tx, quiz.ID(), question); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true

	return nil
}

func (repo *PostgresQuizRepository) FindByID(id domain.QuizID) (domain.Quiz, error) {
	record, err := scanQuizRecord(repo.pool.QueryRow(
		context.Background(),
		selectQuizByIDSQL,
		id.String(),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Quiz{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Quiz{}, err
	}

	questions, err := repo.findQuestionsByQuizIDs([]domain.QuizID{id})
	if err != nil {
		return domain.Quiz{}, err
	}

	return restoreQuiz(record, questions[id])
}

func (repo *PostgresQuizRepository) FindByCourse(courseID domain.CourseID) ([]domain.Quiz, error) {
	rows, err := repo.pool.Query(
		context.Background(),
		selectQuizzesByCourseSQL,
		courseID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []quizRecord{}
	quizIDs := []domain.QuizID{}
	for rows.Next() {
		record, err := scanQuizRecord(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
		quizIDs = append(quizIDs, record.id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	questions, err := repo.findQuestionsByQuizIDs(quizIDs)
	if err != nil {
		return nil, err
	}

	quizzes := make([]domain.Quiz, 0, len(records))
	for _, record := range records {
		quiz, err := restoreQuiz(record, questions[record.id])
		if err != nil {
			return nil, err
		}
		quizzes = append(quizzes, quiz)
	}

	return quizzes, nil
}

func (repo *PostgresQuizRepository) FindByQuestionID(id domain.QuestionID) (domain.Quiz, error) {
	record, err := scanQuizRecord(repo.pool.QueryRow(
		context.Background(),
		selectQuizByQuestionSQL,
		id.String(),
	))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Quiz{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Quiz{}, err
	}

	questions, err := repo.findQuestionsByQuizIDs([]domain.QuizID{record.id})
	if err != nil {
		return domain.Quiz{}, err
	}

	return restoreQuiz(record, questions[record.id])
}

func (repo *PostgresQuizRepository) Delete(id domain.QuizID) error {
	tag, err := repo.pool.Exec(context.Background(), `DELETE FROM quizzes WHERE id = $1`, id.String())
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *PostgresQuizRepository) DeleteByCourse(courseID domain.CourseID) error {
	_, err := repo.pool.Exec(context.Background(), `DELETE FROM quizzes WHERE course_id = $1`, courseID.String())
	return err
}

func (repo *PostgresQuizRepository) beginPoolTransaction(ctx context.Context) (quizTransaction, error) {
	return repo.pool.Begin(ctx)
}

func (repo *PostgresQuizRepository) startTransaction(ctx context.Context) (quizTransaction, error) {
	if repo.beginTx != nil {
		return repo.beginTx(ctx)
	}

	return repo.beginPoolTransaction(ctx)
}

const selectQuizSQL = `SELECT
	id,
	course_id,
	title,
	pass_threshold,
	created_at,
	updated_at
FROM quizzes`

const selectQuizByIDSQL = selectQuizSQL + ` WHERE id = $1`

const selectQuizzesByCourseSQL = selectQuizSQL + ` WHERE course_id = $1 ORDER BY created_at DESC`

const selectQuizByQuestionSQL = `SELECT
	q.id,
	q.course_id,
	q.title,
	q.pass_threshold,
	q.created_at,
	q.updated_at
FROM quizzes AS q
JOIN quiz_questions AS qq ON qq.quiz_id = q.id
WHERE qq.id = $1`

const upsertQuizSQL = `INSERT INTO quizzes (
	id,
	course_id,
	title,
	pass_threshold,
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO UPDATE SET
	course_id = EXCLUDED.course_id,
	title = EXCLUDED.title,
	pass_threshold = EXCLUDED.pass_threshold,
	updated_at = EXCLUDED.updated_at`

const deleteQuizQuestionsSQL = `DELETE FROM quiz_questions WHERE quiz_id = $1`

const insertQuizQuestionSQL = `INSERT INTO quiz_questions (
	id,
	quiz_id,
	type,
	prompt,
	options,
	correct_indices,
	explanation,
	position
) VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8)`

const selectQuizQuestionsSQL = `SELECT
	id,
	quiz_id,
	type,
	prompt,
	options::text,
	correct_indices::text,
	explanation,
	position
FROM quiz_questions
WHERE quiz_id = ANY($1::uuid[])
ORDER BY quiz_id ASC, position ASC`

type quizScanner interface {
	Scan(dest ...any) error
}

type quizRecord struct {
	id            domain.QuizID
	courseID      domain.CourseID
	title         string
	passThreshold domain.PassThreshold
	createdAt     time.Time
	updatedAt     time.Time
}

func scanQuizRecord(scanner quizScanner) (quizRecord, error) {
	var (
		idValue            string
		courseIDValue      string
		title              string
		passThresholdValue float64
		createdAt          time.Time
		updatedAt          time.Time
	)

	if err := scanner.Scan(
		&idValue,
		&courseIDValue,
		&title,
		&passThresholdValue,
		&createdAt,
		&updatedAt,
	); err != nil {
		return quizRecord{}, err
	}

	id, err := domain.NewQuizID(idValue)
	if err != nil {
		return quizRecord{}, err
	}

	courseID, err := domain.NewCourseID(courseIDValue)
	if err != nil {
		return quizRecord{}, err
	}

	passThreshold, err := domain.NewPassThreshold(passThresholdValue)
	if err != nil {
		return quizRecord{}, err
	}

	return quizRecord{
		id:            id,
		courseID:      courseID,
		title:         title,
		passThreshold: passThreshold,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

func restoreQuiz(record quizRecord, questions []domain.ChoiceQuestion) (domain.Quiz, error) {
	return domain.RestoreQuiz(
		record.id,
		record.courseID,
		record.title,
		record.passThreshold,
		questions,
		record.createdAt,
		record.updatedAt,
	)
}

func upsertQuiz(ctx context.Context, tx quizTransaction, quiz domain.Quiz) error {
	_, err := tx.Exec(
		ctx,
		upsertQuizSQL,
		quiz.ID().String(),
		quiz.CourseID().String(),
		quiz.Title(),
		quiz.PassThreshold().Float64(),
		quiz.CreatedAt(),
		quiz.UpdatedAt(),
	)

	return err
}

func insertQuizQuestion(ctx context.Context, tx quizTransaction, quizID domain.QuizID, question domain.ChoiceQuestion) error {
	optionsJSON, err := encodeJSON(question.Options())
	if err != nil {
		return err
	}
	correctIndicesJSON, err := encodeJSON(question.CorrectIndices())
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		insertQuizQuestionSQL,
		question.ID().String(),
		quizID.String(),
		question.Type().String(),
		question.Prompt(),
		optionsJSON,
		correctIndicesJSON,
		question.Explanation(),
		question.Position().Int(),
	)

	return err
}

func (repo *PostgresQuizRepository) findQuestionsByQuizIDs(quizIDs []domain.QuizID) (map[domain.QuizID][]domain.ChoiceQuestion, error) {
	questions := make(map[domain.QuizID][]domain.ChoiceQuestion, len(quizIDs))
	if len(quizIDs) == 0 {
		return questions, nil
	}

	idValues := make([]string, 0, len(quizIDs))
	for _, id := range quizIDs {
		idValues = append(idValues, id.String())
		questions[id] = nil
	}

	rows, err := repo.pool.Query(context.Background(), selectQuizQuestionsSQL, idValues)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		quizID, question, err := scanQuizQuestion(rows)
		if err != nil {
			return nil, err
		}

		questions[quizID] = append(questions[quizID], question)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}

func scanQuizQuestion(scanner quizScanner) (domain.QuizID, domain.ChoiceQuestion, error) {
	var (
		idValue            string
		quizIDValue        string
		typeValue          string
		prompt             string
		optionsJSON        string
		correctIndicesJSON string
		explanation        string
		positionValue      int
	)

	if err := scanner.Scan(
		&idValue,
		&quizIDValue,
		&typeValue,
		&prompt,
		&optionsJSON,
		&correctIndicesJSON,
		&explanation,
		&positionValue,
	); err != nil {
		return domain.QuizID{}, domain.ChoiceQuestion{}, err
	}

	id, err := domain.NewQuestionID(idValue)
	if err != nil {
		return domain.QuizID{}, domain.ChoiceQuestion{}, err
	}

	quizID, err := domain.NewQuizID(quizIDValue)
	if err != nil {
		return domain.QuizID{}, domain.ChoiceQuestion{}, err
	}

	questionType, err := domain.NewChoiceQuestionType(typeValue)
	if err != nil {
		return domain.QuizID{}, domain.ChoiceQuestion{}, err
	}

	options, err := decodeStringSlice(optionsJSON)
	if err != nil {
		return domain.QuizID{}, domain.ChoiceQuestion{}, err
	}

	correctIndices, err := decodeIntSlice(correctIndicesJSON)
	if err != nil {
		return domain.QuizID{}, domain.ChoiceQuestion{}, err
	}

	position, err := domain.NewQuestionPosition(positionValue)
	if err != nil {
		return domain.QuizID{}, domain.ChoiceQuestion{}, err
	}

	question, err := domain.NewChoiceQuestion(
		id,
		questionType,
		prompt,
		options,
		correctIndices,
		explanation,
		position,
	)
	if err != nil {
		return domain.QuizID{}, domain.ChoiceQuestion{}, err
	}

	return quizID, question, nil
}

func encodeJSON(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}

func decodeStringSlice(value string) ([]string, error) {
	var decoded []string
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}

func decodeIntSlice(value string) ([]int, error) {
	var decoded []int
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return nil, err
	}

	return decoded, nil
}
