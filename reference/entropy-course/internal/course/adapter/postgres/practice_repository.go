package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

var _ core.PracticeRepository = (*PostgresPracticeRepository)(nil)

type PostgresPracticeRepository struct {
	pool    *pgxpool.Pool
	beginTx func(context.Context) (practiceTransaction, error)
}

type practiceTransaction interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

func NewPostgresPracticeRepository(pool *pgxpool.Pool) *PostgresPracticeRepository {
	repo := &PostgresPracticeRepository{pool: pool}
	repo.beginTx = repo.beginPoolTransaction

	return repo
}

func (repo *PostgresPracticeRepository) Save(practice domain.Practice) error {
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

	if err := upsertPractice(ctx, tx, practice); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, deletePracticeTestCasesSQL, practice.ID().String()); err != nil {
		return err
	}
	for _, testCase := range practice.TestCases() {
		if err := insertPracticeTestCase(ctx, tx, practice.ID(), testCase); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true

	return nil
}

func (repo *PostgresPracticeRepository) FindByID(id domain.PracticeID) (domain.Practice, error) {
	record, err := scanPracticeRecord(repo.pool.QueryRow(
		context.Background(),
		selectPracticeByIDSQL,
		id.String(),
	))
	if err != nil {
		return domain.Practice{}, mapPracticeRowError(err)
	}

	testCases, err := repo.findTestCasesByPracticeIDs([]domain.PracticeID{id})
	if err != nil {
		return domain.Practice{}, err
	}

	return restorePractice(record, testCases[id])
}

func (repo *PostgresPracticeRepository) FindByCourse(courseID domain.CourseID) ([]domain.Practice, error) {
	rows, err := repo.pool.Query(
		context.Background(),
		selectPracticesByCourseSQL,
		courseID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []practiceRecord{}
	practiceIDs := []domain.PracticeID{}
	for rows.Next() {
		record, err := scanPracticeRecord(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
		practiceIDs = append(practiceIDs, record.id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	testCases, err := repo.findTestCasesByPracticeIDs(practiceIDs)
	if err != nil {
		return nil, err
	}

	practices := make([]domain.Practice, 0, len(records))
	for _, record := range records {
		practice, err := restorePractice(record, testCases[record.id])
		if err != nil {
			return nil, err
		}
		practices = append(practices, practice)
	}

	return practices, nil
}

func (repo *PostgresPracticeRepository) FindByTestCaseID(id domain.TestCaseID) (domain.Practice, error) {
	record, err := scanPracticeRecord(repo.pool.QueryRow(
		context.Background(),
		selectPracticeByTestCaseSQL,
		id.String(),
	))
	if err != nil {
		return domain.Practice{}, mapPracticeRowError(err)
	}

	testCases, err := repo.findTestCasesByPracticeIDs([]domain.PracticeID{record.id})
	if err != nil {
		return domain.Practice{}, err
	}

	return restorePractice(record, testCases[record.id])
}

func (repo *PostgresPracticeRepository) Delete(id domain.PracticeID) error {
	tag, err := repo.pool.Exec(context.Background(), deletePracticeSQL, id.String())
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *PostgresPracticeRepository) DeleteByCourse(courseID domain.CourseID) error {
	_, err := repo.pool.Exec(context.Background(), deletePracticesByCourseSQL, courseID.String())
	return err
}

func (repo *PostgresPracticeRepository) beginPoolTransaction(ctx context.Context) (practiceTransaction, error) {
	return repo.pool.Begin(ctx)
}

func (repo *PostgresPracticeRepository) startTransaction(ctx context.Context) (practiceTransaction, error) {
	if repo.beginTx != nil {
		return repo.beginTx(ctx)
	}

	return repo.beginPoolTransaction(ctx)
}

const selectPracticeSQL = `SELECT
	id,
	course_id,
	title,
	language,
	prompt,
	starter_code,
	solution,
	created_at,
	updated_at
FROM practices`

const selectPracticeByIDSQL = selectPracticeSQL + ` WHERE id = $1`

const selectPracticesByCourseSQL = selectPracticeSQL + ` WHERE course_id = $1 ORDER BY created_at DESC`

const selectPracticeByTestCaseSQL = `SELECT
	p.id,
	p.course_id,
	p.title,
	p.language,
	p.prompt,
	p.starter_code,
	p.solution,
	p.created_at,
	p.updated_at
FROM practices AS p
JOIN practice_test_cases AS tc ON tc.practice_id = p.id
WHERE tc.id = $1`

const upsertPracticeSQL = `INSERT INTO practices (
	id,
	course_id,
	title,
	language,
	prompt,
	starter_code,
	solution,
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
ON CONFLICT (id) DO UPDATE SET
	course_id = EXCLUDED.course_id,
	title = EXCLUDED.title,
	language = EXCLUDED.language,
	prompt = EXCLUDED.prompt,
	starter_code = EXCLUDED.starter_code,
	solution = EXCLUDED.solution,
	updated_at = EXCLUDED.updated_at`

const deletePracticeTestCasesSQL = `DELETE FROM practice_test_cases WHERE practice_id = $1`

const insertPracticeTestCaseSQL = `INSERT INTO practice_test_cases (
	id,
	practice_id,
	stdin,
	expected_stdout,
	name,
	position
) VALUES ($1, $2, $3, $4, $5, $6)`

const selectPracticeTestCasesSQL = `SELECT
	id,
	practice_id,
	stdin,
	expected_stdout,
	name,
	position
FROM practice_test_cases
WHERE practice_id = ANY($1::uuid[])
ORDER BY practice_id ASC, position ASC`

const deletePracticeSQL = `DELETE FROM practices WHERE id = $1`

const deletePracticesByCourseSQL = `DELETE FROM practices WHERE course_id = $1`

type practiceScanner interface {
	Scan(dest ...any) error
}

type practiceRecord struct {
	id          domain.PracticeID
	courseID    domain.CourseID
	title       string
	language    domain.Language
	prompt      string
	starterCode string
	solution    string
	createdAt   time.Time
	updatedAt   time.Time
}

func scanPracticeRecord(scanner practiceScanner) (practiceRecord, error) {
	var (
		idValue       string
		courseIDValue string
		title         string
		languageValue string
		prompt        string
		starterCode   string
		solution      string
		createdAt     time.Time
		updatedAt     time.Time
	)

	if err := scanner.Scan(
		&idValue,
		&courseIDValue,
		&title,
		&languageValue,
		&prompt,
		&starterCode,
		&solution,
		&createdAt,
		&updatedAt,
	); err != nil {
		return practiceRecord{}, err
	}

	id, err := domain.NewPracticeID(idValue)
	if err != nil {
		return practiceRecord{}, err
	}

	courseID, err := domain.NewCourseID(courseIDValue)
	if err != nil {
		return practiceRecord{}, err
	}

	language, err := domain.NewLanguage(languageValue)
	if err != nil {
		return practiceRecord{}, err
	}

	return practiceRecord{
		id:          id,
		courseID:    courseID,
		title:       title,
		language:    language,
		prompt:      prompt,
		starterCode: starterCode,
		solution:    solution,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
	}, nil
}

func restorePractice(record practiceRecord, testCases []domain.TestCase) (domain.Practice, error) {
	return domain.RestorePractice(
		record.id,
		record.courseID,
		record.title,
		record.language,
		record.prompt,
		record.starterCode,
		record.solution,
		testCases,
		record.createdAt,
		record.updatedAt,
	)
}

func upsertPractice(ctx context.Context, tx practiceTransaction, practice domain.Practice) error {
	_, err := tx.Exec(
		ctx,
		upsertPracticeSQL,
		practice.ID().String(),
		practice.CourseID().String(),
		practice.Title(),
		practice.Language().String(),
		practice.Prompt(),
		practice.StarterCode(),
		practice.Solution(),
		practice.CreatedAt(),
		practice.UpdatedAt(),
	)

	return err
}

func insertPracticeTestCase(
	ctx context.Context,
	tx practiceTransaction,
	practiceID domain.PracticeID,
	testCase domain.TestCase,
) error {
	_, err := tx.Exec(
		ctx,
		insertPracticeTestCaseSQL,
		testCase.ID().String(),
		practiceID.String(),
		testCase.Stdin(),
		testCase.ExpectedStdout(),
		testCase.Name(),
		testCase.Position().Int(),
	)

	return err
}

func (repo *PostgresPracticeRepository) findTestCasesByPracticeIDs(
	practiceIDs []domain.PracticeID,
) (map[domain.PracticeID][]domain.TestCase, error) {
	testCases := make(map[domain.PracticeID][]domain.TestCase, len(practiceIDs))
	if len(practiceIDs) == 0 {
		return testCases, nil
	}

	idValues := make([]string, 0, len(practiceIDs))
	for _, id := range practiceIDs {
		idValues = append(idValues, id.String())
		testCases[id] = nil
	}

	rows, err := repo.pool.Query(context.Background(), selectPracticeTestCasesSQL, idValues)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		practiceID, testCase, err := scanPracticeTestCase(rows)
		if err != nil {
			return nil, err
		}

		testCases[practiceID] = append(testCases[practiceID], testCase)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return testCases, nil
}

func scanPracticeTestCase(scanner practiceScanner) (domain.PracticeID, domain.TestCase, error) {
	var (
		idValue         string
		practiceIDValue string
		stdin           string
		expectedStdout  string
		name            string
		positionValue   int
	)

	if err := scanner.Scan(
		&idValue,
		&practiceIDValue,
		&stdin,
		&expectedStdout,
		&name,
		&positionValue,
	); err != nil {
		return domain.PracticeID{}, domain.TestCase{}, err
	}

	id, err := domain.NewTestCaseID(idValue)
	if err != nil {
		return domain.PracticeID{}, domain.TestCase{}, err
	}

	practiceID, err := domain.NewPracticeID(practiceIDValue)
	if err != nil {
		return domain.PracticeID{}, domain.TestCase{}, err
	}

	position, err := domain.NewTestCasePosition(positionValue)
	if err != nil {
		return domain.PracticeID{}, domain.TestCase{}, err
	}

	testCase, err := domain.NewTestCase(id, stdin, expectedStdout, name, position)
	if err != nil {
		return domain.PracticeID{}, domain.TestCase{}, err
	}

	return practiceID, testCase, nil
}

func mapPracticeRowError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	return err
}
