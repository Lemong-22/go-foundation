package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

var _ core.TestRepository = (*PostgresTestRepository)(nil)

type PostgresTestRepository struct {
	pool    *pgxpool.Pool
	beginTx func(context.Context) (testTransaction, error)
}

type testTransaction interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

func NewPostgresTestRepository(pool *pgxpool.Pool) *PostgresTestRepository {
	repo := &PostgresTestRepository{pool: pool}
	repo.beginTx = repo.beginPoolTransaction

	return repo
}

func (repo *PostgresTestRepository) Save(test domain.Test) error {
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

	if err := upsertTest(ctx, tx, test); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, deleteTestItemsSQL, test.ID().String()); err != nil {
		return err
	}
	for _, item := range test.Items() {
		if err := insertTestItem(ctx, tx, test.ID(), item); err != nil {
			return err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	committed = true

	return nil
}

func (repo *PostgresTestRepository) FindByID(id domain.TestID) (domain.Test, error) {
	record, err := scanTestRecord(repo.pool.QueryRow(
		context.Background(),
		selectTestByIDSQL,
		id.String(),
	))
	if err != nil {
		return domain.Test{}, mapTestRowError(err)
	}

	items, err := repo.findItemsByTestIDs([]domain.TestID{id})
	if err != nil {
		return domain.Test{}, err
	}

	return restoreTest(record, items[id])
}

func (repo *PostgresTestRepository) FindByCourse(courseID domain.CourseID) ([]domain.Test, error) {
	rows, err := repo.pool.Query(
		context.Background(),
		selectTestsByCourseSQL,
		courseID.String(),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	records := []testRecord{}
	testIDs := []domain.TestID{}
	for rows.Next() {
		record, err := scanTestRecord(rows)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
		testIDs = append(testIDs, record.id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items, err := repo.findItemsByTestIDs(testIDs)
	if err != nil {
		return nil, err
	}

	tests := make([]domain.Test, 0, len(records))
	for _, record := range records {
		test, err := restoreTest(record, items[record.id])
		if err != nil {
			return nil, err
		}
		tests = append(tests, test)
	}

	return tests, nil
}

func (repo *PostgresTestRepository) FindByItemID(id domain.TestItemID) (domain.Test, error) {
	record, err := scanTestRecord(repo.pool.QueryRow(
		context.Background(),
		selectTestByItemSQL,
		id.String(),
	))
	if err != nil {
		return domain.Test{}, mapTestRowError(err)
	}

	items, err := repo.findItemsByTestIDs([]domain.TestID{record.id})
	if err != nil {
		return domain.Test{}, err
	}

	return restoreTest(record, items[record.id])
}

func (repo *PostgresTestRepository) Delete(id domain.TestID) error {
	tag, err := repo.pool.Exec(context.Background(), deleteTestSQL, id.String())
	return mapTestDeleteResult(tag, err)
}

func (repo *PostgresTestRepository) DeleteByCourse(courseID domain.CourseID) error {
	_, err := repo.pool.Exec(context.Background(), deleteTestsByCourseSQL, courseID.String())
	return err
}

func (repo *PostgresTestRepository) beginPoolTransaction(ctx context.Context) (testTransaction, error) {
	return repo.pool.Begin(ctx)
}

func (repo *PostgresTestRepository) startTransaction(ctx context.Context) (testTransaction, error) {
	if repo.beginTx != nil {
		return repo.beginTx(ctx)
	}

	return repo.beginPoolTransaction(ctx)
}

const selectTestSQL = `SELECT
	id,
	course_id,
	title,
	time_limit_minutes,
	pass_threshold,
	solution_zip_provider,
	solution_zip_locator,
	solution_video_provider,
	solution_video_locator,
	solution_video_caption,
	created_at,
	updated_at
FROM tests`

const selectTestByIDSQL = selectTestSQL + ` WHERE id = $1`

const selectTestsByCourseSQL = selectTestSQL + ` WHERE course_id = $1 ORDER BY created_at DESC`

const selectTestByItemSQL = `SELECT
	t.id,
	t.course_id,
	t.title,
	t.time_limit_minutes,
	t.pass_threshold,
	t.solution_zip_provider,
	t.solution_zip_locator,
	t.solution_video_provider,
	t.solution_video_locator,
	t.solution_video_caption,
	t.created_at,
	t.updated_at
FROM tests AS t
JOIN test_items AS ti ON ti.test_id = t.id
WHERE ti.id = $1`

const upsertTestSQL = `INSERT INTO tests (
	id,
	course_id,
	title,
	time_limit_minutes,
	pass_threshold,
	solution_zip_provider,
	solution_zip_locator,
	solution_video_provider,
	solution_video_locator,
	solution_video_caption,
	created_at,
	updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (id) DO UPDATE SET
	course_id = EXCLUDED.course_id,
	title = EXCLUDED.title,
	time_limit_minutes = EXCLUDED.time_limit_minutes,
	pass_threshold = EXCLUDED.pass_threshold,
	solution_zip_provider = EXCLUDED.solution_zip_provider,
	solution_zip_locator = EXCLUDED.solution_zip_locator,
	solution_video_provider = EXCLUDED.solution_video_provider,
	solution_video_locator = EXCLUDED.solution_video_locator,
	solution_video_caption = EXCLUDED.solution_video_caption,
	updated_at = EXCLUDED.updated_at`

const deleteTestItemsSQL = `DELETE FROM test_items WHERE test_id = $1`

const insertTestItemSQL = `INSERT INTO test_items (
	id,
	test_id,
	kind,
	position,
	choice_type,
	choice_prompt,
	choice_options,
	choice_correct_indices,
	choice_explanation,
	coding_language,
	coding_prompt,
	starter_code,
	coding_solution,
	coding_test_cases
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb, $9, $10, $11, $12, $13, $14::jsonb)`

const selectTestItemsSQL = `SELECT
	id,
	test_id,
	kind,
	position,
	choice_type,
	choice_prompt,
	choice_options::text,
	choice_correct_indices::text,
	choice_explanation,
	coding_language,
	coding_prompt,
	starter_code,
	coding_solution,
	coding_test_cases::text
FROM test_items
WHERE test_id = ANY($1::uuid[])
ORDER BY test_id ASC, position ASC`

const deleteTestSQL = `DELETE FROM tests WHERE id = $1`

const deleteTestsByCourseSQL = `DELETE FROM tests WHERE course_id = $1`

type testScanner interface {
	Scan(dest ...any) error
}

type testRecord struct {
	id            domain.TestID
	courseID      domain.CourseID
	title         string
	timeLimit     *domain.TimeLimit
	passThreshold domain.PassThreshold
	solution      *domain.TestSolution
	createdAt     time.Time
	updatedAt     time.Time
}

func scanTestRecord(scanner testScanner) (testRecord, error) {
	var (
		idValue               string
		courseIDValue         string
		title                 string
		timeLimitMinutes      pgtype.Int4
		passThresholdValue    float64
		solutionZipProvider   pgtype.Text
		solutionZipLocator    pgtype.Text
		solutionVideoProvider pgtype.Text
		solutionVideoLocator  pgtype.Text
		solutionVideoCaption  pgtype.Text
		createdAt             time.Time
		updatedAt             time.Time
	)

	if err := scanner.Scan(
		&idValue,
		&courseIDValue,
		&title,
		&timeLimitMinutes,
		&passThresholdValue,
		&solutionZipProvider,
		&solutionZipLocator,
		&solutionVideoProvider,
		&solutionVideoLocator,
		&solutionVideoCaption,
		&createdAt,
		&updatedAt,
	); err != nil {
		return testRecord{}, err
	}

	id, err := domain.NewTestID(idValue)
	if err != nil {
		return testRecord{}, err
	}

	courseID, err := domain.NewCourseID(courseIDValue)
	if err != nil {
		return testRecord{}, err
	}

	timeLimit, err := testTimeLimitFromColumn(timeLimitMinutes)
	if err != nil {
		return testRecord{}, err
	}

	passThreshold, err := domain.NewPassThreshold(passThresholdValue)
	if err != nil {
		return testRecord{}, err
	}

	solution, err := testSolutionFromColumns(
		solutionZipProvider,
		solutionZipLocator,
		solutionVideoProvider,
		solutionVideoLocator,
		solutionVideoCaption,
	)
	if err != nil {
		return testRecord{}, err
	}

	return testRecord{
		id:            id,
		courseID:      courseID,
		title:         title,
		timeLimit:     timeLimit,
		passThreshold: passThreshold,
		solution:      solution,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

func restoreTest(record testRecord, items []domain.TestItem) (domain.Test, error) {
	return domain.RestoreTest(
		record.id,
		record.courseID,
		record.title,
		record.timeLimit,
		record.passThreshold,
		record.solution,
		items,
		record.createdAt,
		record.updatedAt,
	)
}

func upsertTest(ctx context.Context, tx testTransaction, test domain.Test) error {
	timeLimitMinutes, solutionZipProvider, solutionZipLocator, solutionVideoProvider, solutionVideoLocator, solutionVideoCaption := testPayloadColumns(test)

	_, err := tx.Exec(
		ctx,
		upsertTestSQL,
		test.ID().String(),
		test.CourseID().String(),
		test.Title(),
		timeLimitMinutes,
		test.PassThreshold().Float64(),
		solutionZipProvider,
		solutionZipLocator,
		solutionVideoProvider,
		solutionVideoLocator,
		solutionVideoCaption,
		test.CreatedAt(),
		test.UpdatedAt(),
	)

	return err
}

func testPayloadColumns(test domain.Test) (any, any, any, any, any, any) {
	var timeLimitMinutes any
	if timeLimit := test.TimeLimit(); timeLimit != nil {
		timeLimitMinutes = timeLimit.Minutes()
	}

	solution := test.Solution()
	if solution == nil {
		return timeLimitMinutes, nil, nil, nil, nil, nil
	}

	return timeLimitMinutes,
		solution.SolutionZip().Provider().String(),
		solution.SolutionZip().Locator(),
		solution.ExplanationVideo().Provider().String(),
		solution.ExplanationVideo().Locator(),
		solution.ExplanationCaption()
}

func insertTestItem(ctx context.Context, tx testTransaction, testID domain.TestID, item domain.TestItem) error {
	choiceType, choicePrompt, choiceOptions, choiceCorrectIndices, choiceExplanation, codingLanguage, codingPrompt, starterCode, codingSolution, codingTestCases, err := testItemPayloadColumns(item)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		ctx,
		insertTestItemSQL,
		item.ID().String(),
		testID.String(),
		item.Kind().String(),
		item.Position().Int(),
		choiceType,
		choicePrompt,
		choiceOptions,
		choiceCorrectIndices,
		choiceExplanation,
		codingLanguage,
		codingPrompt,
		starterCode,
		codingSolution,
		codingTestCases,
	)

	return err
}

func testItemPayloadColumns(item domain.TestItem) (any, any, any, any, any, any, any, any, any, any, error) {
	switch body := item.Body().(type) {
	case domain.ChoiceItemBody:
		optionsJSON, err := encodeJSON(body.Options())
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}
		correctIndicesJSON, err := encodeJSON(body.CorrectIndices())
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}

		return body.Type().String(),
			body.Prompt(),
			optionsJSON,
			correctIndicesJSON,
			body.Explanation(),
			nil,
			nil,
			nil,
			nil,
			nil,
			nil
	case domain.CodingItemBody:
		testCasesJSON, err := encodeCodingTestCases(body.TestCases())
		if err != nil {
			return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, err
		}

		return nil,
			nil,
			nil,
			nil,
			nil,
			body.Language().String(),
			body.Prompt(),
			body.StarterCode(),
			body.Solution(),
			testCasesJSON,
			nil
	default:
		return nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, domain.NewValidationError("body", "must be choice or coding")
	}
}

func (repo *PostgresTestRepository) findItemsByTestIDs(testIDs []domain.TestID) (map[domain.TestID][]domain.TestItem, error) {
	items := make(map[domain.TestID][]domain.TestItem, len(testIDs))
	if len(testIDs) == 0 {
		return items, nil
	}

	idValues := make([]string, 0, len(testIDs))
	for _, id := range testIDs {
		idValues = append(idValues, id.String())
		items[id] = nil
	}

	rows, err := repo.pool.Query(context.Background(), selectTestItemsSQL, idValues)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		testID, item, err := scanTestItem(rows)
		if err != nil {
			return nil, err
		}

		items[testID] = append(items[testID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func scanTestItem(scanner testScanner) (domain.TestID, domain.TestItem, error) {
	var (
		idValue              string
		testIDValue          string
		kindValue            string
		positionValue        int
		choiceType           pgtype.Text
		choicePrompt         pgtype.Text
		choiceOptions        pgtype.Text
		choiceCorrectIndices pgtype.Text
		choiceExplanation    pgtype.Text
		codingLanguage       pgtype.Text
		codingPrompt         pgtype.Text
		starterCode          pgtype.Text
		codingSolution       pgtype.Text
		codingTestCases      pgtype.Text
	)

	if err := scanner.Scan(
		&idValue,
		&testIDValue,
		&kindValue,
		&positionValue,
		&choiceType,
		&choicePrompt,
		&choiceOptions,
		&choiceCorrectIndices,
		&choiceExplanation,
		&codingLanguage,
		&codingPrompt,
		&starterCode,
		&codingSolution,
		&codingTestCases,
	); err != nil {
		return domain.TestID{}, domain.TestItem{}, err
	}

	id, err := domain.NewTestItemID(idValue)
	if err != nil {
		return domain.TestID{}, domain.TestItem{}, err
	}

	testID, err := domain.NewTestID(testIDValue)
	if err != nil {
		return domain.TestID{}, domain.TestItem{}, err
	}

	kind, err := domain.NewTestItemKind(kindValue)
	if err != nil {
		return domain.TestID{}, domain.TestItem{}, err
	}

	position, err := domain.NewTestItemPosition(positionValue)
	if err != nil {
		return domain.TestID{}, domain.TestItem{}, err
	}

	body, err := testItemBodyFromColumns(
		kind,
		choiceType,
		choicePrompt,
		choiceOptions,
		choiceCorrectIndices,
		choiceExplanation,
		codingLanguage,
		codingPrompt,
		starterCode,
		codingSolution,
		codingTestCases,
	)
	if err != nil {
		return domain.TestID{}, domain.TestItem{}, err
	}

	item, err := domain.NewTestItem(id, kind, body, position)
	if err != nil {
		return domain.TestID{}, domain.TestItem{}, err
	}

	return testID, item, nil
}

func testItemBodyFromColumns(
	kind domain.TestItemKind,
	choiceType pgtype.Text,
	choicePrompt pgtype.Text,
	choiceOptions pgtype.Text,
	choiceCorrectIndices pgtype.Text,
	choiceExplanation pgtype.Text,
	codingLanguage pgtype.Text,
	codingPrompt pgtype.Text,
	starterCode pgtype.Text,
	codingSolution pgtype.Text,
	codingTestCases pgtype.Text,
) (domain.TestItemBody, error) {
	if kind.IsChoice() {
		questionType, err := domain.NewChoiceQuestionType(nullableText(choiceType))
		if err != nil {
			return nil, err
		}
		options, err := decodeStringSlice(nullableText(choiceOptions))
		if err != nil {
			return nil, err
		}
		correctIndices, err := decodeIntSlice(nullableText(choiceCorrectIndices))
		if err != nil {
			return nil, err
		}

		return domain.NewChoiceItemBody(
			questionType,
			nullableText(choicePrompt),
			options,
			correctIndices,
			nullableText(choiceExplanation),
		)
	}

	if kind.IsCoding() {
		language, err := domain.NewLanguage(nullableText(codingLanguage))
		if err != nil {
			return nil, err
		}
		testCases, err := decodeCodingTestCases(nullableText(codingTestCases))
		if err != nil {
			return nil, err
		}

		return domain.NewCodingItemBody(
			language,
			nullableText(codingPrompt),
			nullableText(starterCode),
			nullableText(codingSolution),
			testCases,
		)
	}

	return nil, domain.NewValidationError("kind", "must be choice or coding")
}

func testTimeLimitFromColumn(value pgtype.Int4) (*domain.TimeLimit, error) {
	if !value.Valid {
		return nil, nil
	}

	limit, err := domain.NewTimeLimit(int(value.Int32))
	if err != nil {
		return nil, err
	}

	return &limit, nil
}

func testSolutionFromColumns(
	zipProvider pgtype.Text,
	zipLocator pgtype.Text,
	videoProvider pgtype.Text,
	videoLocator pgtype.Text,
	videoCaption pgtype.Text,
) (*domain.TestSolution, error) {
	setCount := 0
	for _, value := range []pgtype.Text{zipProvider, zipLocator, videoProvider, videoLocator} {
		if value.Valid {
			setCount++
		}
	}
	if setCount == 0 {
		return nil, nil
	}
	if setCount != 4 {
		return nil, domain.NewValidationError("solution", "must include zip and video media refs")
	}

	zipMediaProvider, err := domain.NewMediaProvider(zipProvider.String)
	if err != nil {
		return nil, err
	}
	zipMedia, err := domain.NewMediaRef(zipMediaProvider, zipLocator.String)
	if err != nil {
		return nil, err
	}

	videoMediaProvider, err := domain.NewMediaProvider(videoProvider.String)
	if err != nil {
		return nil, err
	}
	videoMedia, err := domain.NewMediaRef(videoMediaProvider, videoLocator.String)
	if err != nil {
		return nil, err
	}

	solution, err := domain.NewTestSolution(zipMedia, videoMedia, nullableText(videoCaption))
	if err != nil {
		return nil, err
	}

	return &solution, nil
}

type codingTestCaseJSON struct {
	Stdin          string `json:"stdin"`
	ExpectedStdout string `json:"expected_stdout"`
	Name           string `json:"name"`
}

func encodeCodingTestCases(testCases []domain.CodingTestCase) (string, error) {
	encoded := make([]codingTestCaseJSON, 0, len(testCases))
	for _, testCase := range testCases {
		encoded = append(encoded, codingTestCaseJSON{
			Stdin:          testCase.Stdin(),
			ExpectedStdout: testCase.ExpectedStdout(),
			Name:           testCase.Name(),
		})
	}

	return encodeJSON(encoded)
}

func decodeCodingTestCases(value string) ([]domain.CodingTestCase, error) {
	var decoded []codingTestCaseJSON
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return nil, err
	}

	testCases := make([]domain.CodingTestCase, 0, len(decoded))
	for _, testCase := range decoded {
		testCases = append(testCases, domain.NewCodingTestCase(testCase.Stdin, testCase.ExpectedStdout, testCase.Name))
	}

	return testCases, nil
}

func mapTestRowError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	return err
}

func mapTestDeleteResult(tag pgconn.CommandTag, err error) error {
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
