package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	postgresPracticeIDValue = "550e8400-e29b-41d4-a716-446655440050"
	postgresOtherPracticeID = "550e8400-e29b-41d4-a716-446655440051"
	postgresTestCaseIDValue = "550e8400-e29b-41d4-a716-446655440060"
	postgresOtherTestCaseID = "550e8400-e29b-41d4-a716-446655440061"
)

func TestScanPracticeRecordRestoresPersistedFields(t *testing.T) {
	createdAt := time.Date(2026, 5, 26, 11, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	record, err := scanPracticeRecord(valueScannerFake{values: []any{
		postgresPracticeIDValue,
		courseIDValue,
		"FizzBuzz",
		"golang",
		"Print fizz buzz",
		"starter",
		"solution",
		createdAt,
		updatedAt,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if record.id.String() != postgresPracticeIDValue || record.courseID.String() != courseIDValue {
		t.Fatalf("expected persisted identity fields")
	}
	if record.title != "FizzBuzz" || record.language != domain.Golang() || record.prompt != "Print fizz buzz" {
		t.Fatalf("expected practice metadata to be restored")
	}
	if record.starterCode != "starter" || record.solution != "solution" {
		t.Fatalf("expected source fields to be restored")
	}
	if !record.createdAt.Equal(createdAt) || !record.updatedAt.Equal(updatedAt) {
		t.Fatalf("expected timestamps to be restored")
	}
}

func TestScanPracticeTestCaseRestoresTestCase(t *testing.T) {
	practiceID, testCase, err := scanPracticeTestCase(valueScannerFake{values: []any{
		postgresTestCaseIDValue,
		postgresPracticeIDValue,
		"input\n",
		"output\n",
		"sample",
		2,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if practiceID.String() != postgresPracticeIDValue || testCase.ID().String() != postgresTestCaseIDValue {
		t.Fatalf("expected persisted ids")
	}
	if testCase.Stdin() != "input\n" || testCase.ExpectedStdout() != "output\n" || testCase.Name() != "sample" {
		t.Fatalf("expected test case payload to be restored")
	}
	if testCase.Position().Int() != 2 {
		t.Fatalf("expected position 2, got %d", testCase.Position().Int())
	}
}

func TestSaveCommitsPracticeAndTestCaseReplacementInOneTransaction(t *testing.T) {
	tx := &lessonTransactionFake{}
	repo := newPostgresPracticeRepositoryWithTransaction(tx)

	err := repo.Save(mustPostgresPractice(
		t,
		postgresPracticeIDValue,
		"FizzBuzz",
		"golang",
		"Print fizz buzz",
		"starter",
		"solution",
		mustPostgresTestCase(t, postgresTestCaseIDValue, 0),
		mustPostgresTestCase(t, postgresOtherTestCaseID, 1),
	))
	if err != nil {
		t.Fatalf("expected save to succeed, got %v", err)
	}

	if tx.execCalls != 4 {
		t.Fatalf("expected upsert, delete, and two test case inserts, got %d calls", tx.execCalls)
	}
	if !strings.Contains(strings.ToLower(tx.calls[0].sql), "insert into practices") {
		t.Fatalf("expected first call to upsert practice, got %q", tx.calls[0].sql)
	}
	if !strings.Contains(strings.ToLower(tx.calls[1].sql), "delete from practice_test_cases") {
		t.Fatalf("expected second call to delete old test cases, got %q", tx.calls[1].sql)
	}
	if !strings.Contains(strings.ToLower(tx.calls[2].sql), "insert into practice_test_cases") ||
		!strings.Contains(strings.ToLower(tx.calls[3].sql), "insert into practice_test_cases") {
		t.Fatalf("expected remaining calls to insert test cases")
	}
	if tx.calls[0].args[0] != postgresPracticeIDValue ||
		tx.calls[0].args[3] != "golang" ||
		tx.calls[0].args[5] != "starter" ||
		tx.calls[0].args[6] != "solution" {
		t.Fatalf("expected practice upsert args, got %+v", tx.calls[0].args)
	}
	if tx.calls[2].args[0] != postgresTestCaseIDValue || tx.calls[2].args[1] != postgresPracticeIDValue || tx.calls[2].args[5] != 0 {
		t.Fatalf("expected first test case insert args, got %+v", tx.calls[2].args)
	}
	if tx.calls[3].args[0] != postgresOtherTestCaseID || tx.calls[3].args[5] != 1 {
		t.Fatalf("expected second test case insert args, got %+v", tx.calls[3].args)
	}
	if !tx.committed || tx.rolledBack {
		t.Fatalf("expected successful save to commit without rollback")
	}
}

func TestSaveRollsBackWhenPracticeTestCaseInsertFails(t *testing.T) {
	errBoom := errors.New("insert test case failed")
	tx := &lessonTransactionFake{execErrs: []error{nil, nil, errBoom}}
	repo := newPostgresPracticeRepositoryWithTransaction(tx)

	err := repo.Save(mustPostgresPractice(t, postgresPracticeIDValue, "FizzBuzz", "golang", "Prompt", "", "", mustPostgresTestCase(t, postgresTestCaseIDValue, 0)))
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected insert test case error, got %v", err)
	}
	if tx.committed {
		t.Fatalf("did not expect failed save to commit")
	}
	if !tx.rolledBack {
		t.Fatalf("expected failed save to roll back")
	}
}

func TestPracticeLookupSQLMatchesRepositoryContract(t *testing.T) {
	listSQL := strings.ToLower(selectPracticesByCourseSQL)
	if !strings.Contains(listSQL, "where course_id = $1") || !strings.Contains(listSQL, "order by created_at desc") {
		t.Fatalf("expected course list SQL to filter by course and order by newest, got %q", selectPracticesByCourseSQL)
	}

	testCaseSQL := strings.ToLower(selectPracticeByTestCaseSQL)
	if !strings.Contains(testCaseSQL, "join practice_test_cases") || !strings.Contains(testCaseSQL, "where tc.id = $1") {
		t.Fatalf("expected test case lookup SQL to join practice_test_cases by test case id, got %q", selectPracticeByTestCaseSQL)
	}

	testCasesSQL := strings.ToLower(selectPracticeTestCasesSQL)
	if !strings.Contains(testCasesSQL, "order by practice_id asc, position asc") {
		t.Fatalf("expected test case hydration SQL to order by position, got %q", selectPracticeTestCasesSQL)
	}

	if !strings.Contains(strings.ToLower(deletePracticeSQL), "delete from practices where id = $1") ||
		!strings.Contains(strings.ToLower(deletePracticesByCourseSQL), "delete from practices where course_id = $1") {
		t.Fatalf("expected practice delete SQL to target practices directly")
	}
}

func TestMapPracticeRowErrorMapsNoRowsToNotFound(t *testing.T) {
	if err := mapPracticeRowError(pgx.ErrNoRows); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func newPostgresPracticeRepositoryWithTransaction(tx practiceTransaction) *PostgresPracticeRepository {
	return &PostgresPracticeRepository{
		beginTx: func(context.Context) (practiceTransaction, error) {
			return tx, nil
		},
	}
}

func mustPostgresPractice(
	t *testing.T,
	idValue string,
	title string,
	languageValue string,
	prompt string,
	starterCode string,
	solution string,
	testCases ...domain.TestCase,
) domain.Practice {
	t.Helper()

	language, err := domain.NewLanguage(languageValue)
	if err != nil {
		t.Fatalf("expected language fixture, got %v", err)
	}

	practice, err := domain.NewPractice(
		mustPostgresPracticeID(idValue),
		mustPostgresCourseID(courseIDValue),
		title,
		language,
		prompt,
		starterCode,
		solution,
		testCases,
		time.Date(2026, 5, 26, 11, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected practice fixture, got %v", err)
	}

	return practice
}

func mustPostgresTestCase(t *testing.T, idValue string, positionValue int) domain.TestCase {
	t.Helper()

	testCase, err := domain.NewTestCase(
		mustPostgresTestCaseID(idValue),
		"stdin",
		"stdout",
		"case",
		mustPostgresTestCasePosition(positionValue),
	)
	if err != nil {
		t.Fatalf("expected test case fixture, got %v", err)
	}

	return testCase
}

func mustPostgresPracticeID(value string) domain.PracticeID {
	id, err := domain.NewPracticeID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustPostgresTestCaseID(value string) domain.TestCaseID {
	id, err := domain.NewTestCaseID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustPostgresTestCasePosition(value int) domain.TestCasePosition {
	position, err := domain.NewTestCasePosition(value)
	if err != nil {
		panic(err)
	}

	return position
}
