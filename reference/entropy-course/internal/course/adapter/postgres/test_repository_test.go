package postgres

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	postgresTestIDValue      = "550e8400-e29b-41d4-a716-446655440070"
	postgresOtherTestID      = "550e8400-e29b-41d4-a716-446655440071"
	postgresTestItemIDValue  = "550e8400-e29b-41d4-a716-446655440080"
	postgresOtherTestItemID  = "550e8400-e29b-41d4-a716-446655440081"
	postgresSolutionZipURL   = "https://example.com/solutions/test.zip"
	postgresSolutionVideoURL = "https://example.com/explanations/test.mp4"
)

func TestScanTestRecordRestoresNilOptionalFields(t *testing.T) {
	createdAt := time.Date(2026, 5, 27, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	record, err := scanTestRecord(valueScannerFake{values: []any{
		postgresTestIDValue,
		courseIDValue,
		"Final Test",
		nil,
		0.7,
		nil,
		nil,
		nil,
		nil,
		nil,
		createdAt,
		updatedAt,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if record.id.String() != postgresTestIDValue || record.courseID.String() != courseIDValue {
		t.Fatalf("expected persisted identity fields")
	}
	if record.title != "Final Test" || record.passThreshold.Float64() != 0.7 {
		t.Fatalf("expected metadata fields to be restored")
	}
	if record.timeLimit != nil || record.solution != nil {
		t.Fatalf("expected nil optional fields, got time limit %v solution %v", record.timeLimit, record.solution)
	}
	if !record.createdAt.Equal(createdAt) || !record.updatedAt.Equal(updatedAt) {
		t.Fatalf("expected timestamps to be restored")
	}
}

func TestScanTestRecordRestoresTimeLimitAndSolution(t *testing.T) {
	record, err := scanTestRecord(valueScannerFake{values: []any{
		postgresTestIDValue,
		courseIDValue,
		"Final Test",
		45,
		0.85,
		"url",
		postgresSolutionZipURL,
		"url",
		postgresSolutionVideoURL,
		"Walkthrough",
		time.Date(2026, 5, 27, 7, 0, 0, 0, time.UTC),
		time.Date(2026, 5, 27, 8, 0, 0, 0, time.UTC),
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if record.timeLimit == nil || record.timeLimit.Minutes() != 45 {
		t.Fatalf("expected time limit to be restored, got %v", record.timeLimit)
	}
	if record.solution == nil {
		t.Fatalf("expected solution package to be restored")
	}
	if record.solution.SolutionZip().Locator() != postgresSolutionZipURL ||
		record.solution.ExplanationVideo().Locator() != postgresSolutionVideoURL ||
		record.solution.ExplanationCaption() != "Walkthrough" {
		t.Fatalf("expected solution refs and caption to be restored")
	}
}

func TestScanTestItemRestoresChoiceItem(t *testing.T) {
	testID, item, err := scanTestItem(valueScannerFake{values: []any{
		postgresTestItemIDValue,
		postgresTestIDValue,
		"choice",
		2,
		"multiple",
		"Pick many",
		`["A","B","C"]`,
		`[0,2]`,
		"Because A and C",
		nil,
		nil,
		nil,
		nil,
		nil,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if testID.String() != postgresTestIDValue || item.ID().String() != postgresTestItemIDValue {
		t.Fatalf("expected persisted ids")
	}
	if !item.Kind().IsChoice() || item.Position().Int() != 2 {
		t.Fatalf("expected choice item at position 2")
	}
	body, ok := item.Body().(domain.ChoiceItemBody)
	if !ok {
		t.Fatalf("expected choice body, got %T", item.Body())
	}
	if !body.Type().IsMultiple() || body.Prompt() != "Pick many" || body.Explanation() != "Because A and C" {
		t.Fatalf("expected choice body metadata to be restored")
	}
	if !reflect.DeepEqual(body.Options(), []string{"A", "B", "C"}) ||
		!reflect.DeepEqual(body.CorrectIndices(), []int{0, 2}) {
		t.Fatalf("expected structured choice JSON to be restored")
	}
}

func TestScanTestItemRestoresCodingItem(t *testing.T) {
	testID, item, err := scanTestItem(valueScannerFake{values: []any{
		postgresOtherTestItemID,
		postgresTestIDValue,
		"coding",
		1,
		nil,
		nil,
		nil,
		nil,
		nil,
		"golang",
		"Write fizz buzz",
		"package main",
		"func main() {}",
		`[{"stdin":"1\n","expected_stdout":"1\n","name":"sample"}]`,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if testID.String() != postgresTestIDValue || item.ID().String() != postgresOtherTestItemID {
		t.Fatalf("expected persisted ids")
	}
	if !item.Kind().IsCoding() || item.Position().Int() != 1 {
		t.Fatalf("expected coding item at position 1")
	}
	body, ok := item.Body().(domain.CodingItemBody)
	if !ok {
		t.Fatalf("expected coding body, got %T", item.Body())
	}
	if body.Language() != domain.Golang() || body.Prompt() != "Write fizz buzz" {
		t.Fatalf("expected coding metadata to be restored")
	}
	if body.StarterCode() != "package main" || body.Solution() != "func main() {}" {
		t.Fatalf("expected source fields to be restored")
	}
	cases := body.TestCases()
	if len(cases) != 1 || cases[0].Stdin() != "1\n" || cases[0].ExpectedStdout() != "1\n" || cases[0].Name() != "sample" {
		t.Fatalf("expected structured coding test cases, got %+v", cases)
	}
}

func TestSaveCommitsEmptyTestWithNilSolutionInOneTransaction(t *testing.T) {
	tx := &lessonTransactionFake{}
	repo := newPostgresTestRepositoryWithTransaction(tx)

	err := repo.Save(mustPostgresTest(t, postgresTestIDValue, "Final Test", nil, nil))
	if err != nil {
		t.Fatalf("expected save to succeed, got %v", err)
	}

	if tx.execCalls != 2 {
		t.Fatalf("expected upsert and item delete for empty test, got %d calls", tx.execCalls)
	}
	if !strings.Contains(strings.ToLower(tx.calls[0].sql), "insert into tests") {
		t.Fatalf("expected first call to upsert test, got %q", tx.calls[0].sql)
	}
	if !strings.Contains(strings.ToLower(tx.calls[1].sql), "delete from test_items") {
		t.Fatalf("expected second call to replace items, got %q", tx.calls[1].sql)
	}
	if tx.calls[0].args[3] != nil {
		t.Fatalf("expected nil time limit, got %+v", tx.calls[0].args)
	}
	for _, index := range []int{5, 6, 7, 8, 9} {
		if tx.calls[0].args[index] != nil {
			t.Fatalf("expected absent solution column %d to be nil, got %+v", index, tx.calls[0].args[index])
		}
	}
	if !tx.committed || tx.rolledBack {
		t.Fatalf("expected successful save to commit without rollback")
	}
}

func TestSaveCommitsTestAndItemReplacementInOneTransaction(t *testing.T) {
	tx := &lessonTransactionFake{}
	repo := newPostgresTestRepositoryWithTransaction(tx)
	timeLimit := mustPostgresTimeLimit(45)
	solution := mustPostgresTestSolution(t)
	test := mustPostgresTest(
		t,
		postgresTestIDValue,
		"Final Test",
		&timeLimit,
		&solution,
		mustPostgresChoiceTestItem(t, postgresTestItemIDValue, 0),
		mustPostgresCodingTestItem(t, postgresOtherTestItemID, 1),
	)
	err := test.ReorderItems([]domain.TestItemPlacement{
		{TestItemID: mustPostgresTestItemID(postgresOtherTestItemID), Position: mustPostgresTestItemPosition(0)},
		{TestItemID: mustPostgresTestItemID(postgresTestItemIDValue), Position: mustPostgresTestItemPosition(1)},
	}, time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("expected reorder fixture to succeed, got %v", err)
	}

	if err := repo.Save(test); err != nil {
		t.Fatalf("expected save to succeed, got %v", err)
	}

	if tx.execCalls != 4 {
		t.Fatalf("expected upsert, delete, and two item inserts, got %d calls", tx.execCalls)
	}
	if tx.calls[0].args[3] != 45 || tx.calls[0].args[5] != "url" || tx.calls[0].args[6] != postgresSolutionZipURL ||
		tx.calls[0].args[7] != "url" || tx.calls[0].args[8] != postgresSolutionVideoURL || tx.calls[0].args[9] != "Walkthrough" {
		t.Fatalf("expected time limit and solution args, got %+v", tx.calls[0].args)
	}
	if tx.calls[2].args[0] != postgresOtherTestItemID || tx.calls[2].args[2] != "coding" || tx.calls[2].args[3] != 0 {
		t.Fatalf("expected reordered coding item to persist first at position 0, got %+v", tx.calls[2].args)
	}
	if tx.calls[2].args[9] != "golang" || tx.calls[2].args[13] != `[{"stdin":"stdin","expected_stdout":"stdout","name":"sample"}]` {
		t.Fatalf("expected coding item payload, got %+v", tx.calls[2].args)
	}
	if tx.calls[3].args[0] != postgresTestItemIDValue || tx.calls[3].args[2] != "choice" || tx.calls[3].args[3] != 1 {
		t.Fatalf("expected reordered choice item to persist second at position 1, got %+v", tx.calls[3].args)
	}
	if tx.calls[3].args[6] != `["A","B"]` || tx.calls[3].args[7] != `[0]` {
		t.Fatalf("expected structured choice JSON payloads, got %+v", tx.calls[3].args)
	}
	if !tx.committed || tx.rolledBack {
		t.Fatalf("expected successful save to commit without rollback")
	}
}

func TestSaveRollsBackWhenTestItemInsertFails(t *testing.T) {
	errBoom := errors.New("insert item failed")
	tx := &lessonTransactionFake{execErrs: []error{nil, nil, errBoom}}
	repo := newPostgresTestRepositoryWithTransaction(tx)

	err := repo.Save(mustPostgresTest(
		t,
		postgresTestIDValue,
		"Final Test",
		nil,
		nil,
		mustPostgresChoiceTestItem(t, postgresTestItemIDValue, 0),
	))
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected insert item error, got %v", err)
	}
	if tx.committed {
		t.Fatalf("did not expect failed save to commit")
	}
	if !tx.rolledBack {
		t.Fatalf("expected failed save to roll back")
	}
}

func TestTestLookupSQLMatchesRepositoryContract(t *testing.T) {
	listSQL := strings.ToLower(selectTestsByCourseSQL)
	if !strings.Contains(listSQL, "where course_id = $1") || !strings.Contains(listSQL, "order by created_at desc") {
		t.Fatalf("expected course list SQL to filter by course and order by newest, got %q", selectTestsByCourseSQL)
	}

	itemSQL := strings.ToLower(selectTestByItemSQL)
	if !strings.Contains(itemSQL, "join test_items") || !strings.Contains(itemSQL, "where ti.id = $1") {
		t.Fatalf("expected item lookup SQL to join test_items by item id, got %q", selectTestByItemSQL)
	}

	itemsSQL := strings.ToLower(selectTestItemsSQL)
	if !strings.Contains(itemsSQL, "order by test_id asc, position asc") {
		t.Fatalf("expected item hydration SQL to order by position, got %q", selectTestItemsSQL)
	}

	if !strings.Contains(strings.ToLower(deleteTestSQL), "delete from tests where id = $1") ||
		!strings.Contains(strings.ToLower(deleteTestsByCourseSQL), "delete from tests where course_id = $1") {
		t.Fatalf("expected test delete SQL to target tests directly")
	}
}

func TestMapTestRowErrorMapsNoRowsToNotFound(t *testing.T) {
	if err := mapTestRowError(pgx.ErrNoRows); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestMapTestDeleteResultMapsMissingRowsToNotFound(t *testing.T) {
	if err := mapTestDeleteResult(pgconn.NewCommandTag("DELETE 0"), nil); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func newPostgresTestRepositoryWithTransaction(tx testTransaction) *PostgresTestRepository {
	return &PostgresTestRepository{
		beginTx: func(context.Context) (testTransaction, error) {
			return tx, nil
		},
	}
}

func mustPostgresTest(
	t *testing.T,
	idValue string,
	title string,
	timeLimit *domain.TimeLimit,
	solution *domain.TestSolution,
	items ...domain.TestItem,
) domain.Test {
	t.Helper()

	test, err := domain.NewTest(
		mustPostgresTestID(idValue),
		mustPostgresCourseID(courseIDValue),
		title,
		timeLimit,
		domain.DefaultPassThreshold(),
		solution,
		items,
		time.Date(2026, 5, 27, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected test fixture, got %v", err)
	}

	return test
}

func mustPostgresChoiceTestItem(t *testing.T, idValue string, positionValue int) domain.TestItem {
	t.Helper()

	body, err := domain.NewChoiceItemBody(
		domain.SingleChoice(),
		"Pick one",
		[]string{"A", "B"},
		[]int{0},
		"Because A",
	)
	if err != nil {
		t.Fatalf("expected choice body fixture, got %v", err)
	}

	item, err := domain.NewTestItem(
		mustPostgresTestItemID(idValue),
		domain.ChoiceKind(),
		body,
		mustPostgresTestItemPosition(positionValue),
	)
	if err != nil {
		t.Fatalf("expected choice item fixture, got %v", err)
	}

	return item
}

func mustPostgresCodingTestItem(t *testing.T, idValue string, positionValue int) domain.TestItem {
	t.Helper()

	body, err := domain.NewCodingItemBody(
		domain.Golang(),
		"Write fizz buzz",
		"package main",
		"func main() {}",
		[]domain.CodingTestCase{domain.NewCodingTestCase("stdin", "stdout", "sample")},
	)
	if err != nil {
		t.Fatalf("expected coding body fixture, got %v", err)
	}

	item, err := domain.NewTestItem(
		mustPostgresTestItemID(idValue),
		domain.CodingKind(),
		body,
		mustPostgresTestItemPosition(positionValue),
	)
	if err != nil {
		t.Fatalf("expected coding item fixture, got %v", err)
	}

	return item
}

func mustPostgresTestSolution(t *testing.T) domain.TestSolution {
	t.Helper()

	zipRef, err := domain.NewMediaRef(domain.URLProvider(), postgresSolutionZipURL)
	if err != nil {
		t.Fatalf("expected solution zip fixture, got %v", err)
	}
	videoRef, err := domain.NewMediaRef(domain.URLProvider(), postgresSolutionVideoURL)
	if err != nil {
		t.Fatalf("expected solution video fixture, got %v", err)
	}
	solution, err := domain.NewTestSolution(zipRef, videoRef, "Walkthrough")
	if err != nil {
		t.Fatalf("expected solution fixture, got %v", err)
	}

	return solution
}

func mustPostgresTimeLimit(value int) domain.TimeLimit {
	limit, err := domain.NewTimeLimit(value)
	if err != nil {
		panic(err)
	}

	return limit
}

func mustPostgresTestID(value string) domain.TestID {
	id, err := domain.NewTestID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustPostgresTestItemID(value string) domain.TestItemID {
	id, err := domain.NewTestItemID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustPostgresTestItemPosition(value int) domain.TestItemPosition {
	position, err := domain.NewTestItemPosition(value)
	if err != nil {
		panic(err)
	}

	return position
}
