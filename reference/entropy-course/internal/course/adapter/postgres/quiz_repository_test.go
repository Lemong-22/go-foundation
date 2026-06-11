package postgres

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	postgresQuizIDValue     = "550e8400-e29b-41d4-a716-446655440030"
	postgresOtherQuizID     = "550e8400-e29b-41d4-a716-446655440031"
	postgresQuestionIDValue = "550e8400-e29b-41d4-a716-446655440040"
	postgresOtherQuestionID = "550e8400-e29b-41d4-a716-446655440041"
)

func TestScanQuizRecordRestoresPersistedFields(t *testing.T) {
	createdAt := time.Date(2026, 5, 26, 7, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	record, err := scanQuizRecord(valueScannerFake{values: []any{
		postgresQuizIDValue,
		courseIDValue,
		"Basics Quiz",
		0.8,
		createdAt,
		updatedAt,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if record.id.String() != postgresQuizIDValue || record.courseID.String() != courseIDValue {
		t.Fatalf("expected persisted identity fields")
	}
	if record.title != "Basics Quiz" || record.passThreshold.Float64() != 0.8 {
		t.Fatalf("expected title and threshold to be restored")
	}
	if !record.createdAt.Equal(createdAt) || !record.updatedAt.Equal(updatedAt) {
		t.Fatalf("expected timestamps to be restored")
	}
}

func TestScanQuizQuestionRestoresChoiceQuestion(t *testing.T) {
	quizID, question, err := scanQuizQuestion(valueScannerFake{values: []any{
		postgresQuestionIDValue,
		postgresQuizIDValue,
		"multiple",
		"Pick many",
		`["A","B","C"]`,
		`[0,2]`,
		"Because A and C",
		2,
	}})
	if err != nil {
		t.Fatalf("expected scan to succeed, got %v", err)
	}

	if quizID.String() != postgresQuizIDValue || question.ID().String() != postgresQuestionIDValue {
		t.Fatalf("expected persisted ids")
	}
	if !question.Type().IsMultiple() || question.Prompt() != "Pick many" || question.Position().Int() != 2 {
		t.Fatalf("expected question metadata to be restored")
	}
	if !reflect.DeepEqual(question.Options(), []string{"A", "B", "C"}) ||
		!reflect.DeepEqual(question.CorrectIndices(), []int{0, 2}) {
		t.Fatalf("expected JSON question content to be restored")
	}
	if question.Explanation() != "Because A and C" {
		t.Fatalf("expected explanation to be restored")
	}
}

func TestSaveCommitsQuizAndQuestionReplacementInOneTransaction(t *testing.T) {
	tx := &lessonTransactionFake{}
	repo := newPostgresQuizRepositoryWithTransaction(tx)

	err := repo.Save(mustPostgresQuiz(t,
		postgresQuizIDValue,
		"Basics Quiz",
		0.75,
		mustPostgresQuestion(t, postgresQuestionIDValue, 0),
		mustPostgresMultipleQuestion(t, postgresOtherQuestionID, 1),
	))
	if err != nil {
		t.Fatalf("expected save to succeed, got %v", err)
	}

	if tx.execCalls != 4 {
		t.Fatalf("expected upsert, delete, and two question inserts, got %d calls", tx.execCalls)
	}
	if !strings.Contains(strings.ToLower(tx.calls[0].sql), "insert into quizzes") {
		t.Fatalf("expected first call to upsert quiz, got %q", tx.calls[0].sql)
	}
	if !strings.Contains(strings.ToLower(tx.calls[1].sql), "delete from quiz_questions") {
		t.Fatalf("expected second call to delete old questions, got %q", tx.calls[1].sql)
	}
	if !strings.Contains(strings.ToLower(tx.calls[2].sql), "insert into quiz_questions") ||
		!strings.Contains(strings.ToLower(tx.calls[3].sql), "insert into quiz_questions") {
		t.Fatalf("expected remaining calls to insert questions")
	}
	if tx.calls[0].args[0] != postgresQuizIDValue || tx.calls[0].args[3] != 0.75 {
		t.Fatalf("expected quiz upsert args, got %+v", tx.calls[0].args)
	}
	if tx.calls[2].args[4] != `["A","B"]` || tx.calls[2].args[5] != `[0]` {
		t.Fatalf("expected single-choice JSON payloads, got %+v", tx.calls[2].args)
	}
	if tx.calls[3].args[4] != `["A","B","C"]` || tx.calls[3].args[5] != `[0,2]` {
		t.Fatalf("expected multiple-choice JSON payloads, got %+v", tx.calls[3].args)
	}
	if !tx.committed || tx.rolledBack {
		t.Fatalf("expected successful save to commit without rollback")
	}
}

func TestSaveRollsBackWhenQuestionInsertFails(t *testing.T) {
	errBoom := errors.New("insert question failed")
	tx := &lessonTransactionFake{execErrs: []error{nil, nil, errBoom}}
	repo := newPostgresQuizRepositoryWithTransaction(tx)

	err := repo.Save(mustPostgresQuiz(t, postgresQuizIDValue, "Basics Quiz", 0.75, mustPostgresQuestion(t, postgresQuestionIDValue, 0)))
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected insert question error, got %v", err)
	}
	if tx.committed {
		t.Fatalf("did not expect failed save to commit")
	}
	if !tx.rolledBack {
		t.Fatalf("expected failed save to roll back")
	}
}

func TestQuizLookupSQLMatchesRepositoryContract(t *testing.T) {
	listSQL := strings.ToLower(selectQuizzesByCourseSQL)
	if !strings.Contains(listSQL, "where course_id = $1") || !strings.Contains(listSQL, "order by created_at desc") {
		t.Fatalf("expected course list SQL to filter by course and order by newest, got %q", selectQuizzesByCourseSQL)
	}

	questionSQL := strings.ToLower(selectQuizByQuestionSQL)
	if !strings.Contains(questionSQL, "join quiz_questions") || !strings.Contains(questionSQL, "where qq.id = $1") {
		t.Fatalf("expected question lookup SQL to join quiz_questions by question id, got %q", selectQuizByQuestionSQL)
	}

	questionsSQL := strings.ToLower(selectQuizQuestionsSQL)
	if !strings.Contains(questionsSQL, "order by quiz_id asc, position asc") {
		t.Fatalf("expected question hydration SQL to order by position, got %q", selectQuizQuestionsSQL)
	}
}

func newPostgresQuizRepositoryWithTransaction(tx quizTransaction) *PostgresQuizRepository {
	return &PostgresQuizRepository{
		beginTx: func(context.Context) (quizTransaction, error) {
			return tx, nil
		},
	}
}

func mustPostgresQuiz(
	t *testing.T,
	idValue string,
	title string,
	thresholdValue float64,
	questions ...domain.ChoiceQuestion,
) domain.Quiz {
	t.Helper()

	threshold, err := domain.NewPassThreshold(thresholdValue)
	if err != nil {
		t.Fatalf("expected pass threshold fixture, got %v", err)
	}

	quiz, err := domain.NewQuiz(
		mustPostgresQuizID(idValue),
		mustPostgresCourseID(courseIDValue),
		title,
		threshold,
		questions,
		time.Date(2026, 5, 26, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected quiz fixture, got %v", err)
	}

	return quiz
}

func mustPostgresQuestion(t *testing.T, idValue string, positionValue int) domain.ChoiceQuestion {
	t.Helper()

	question, err := domain.NewChoiceQuestion(
		mustPostgresQuestionID(idValue),
		domain.SingleChoice(),
		"Pick one",
		[]string{"A", "B"},
		[]int{0},
		"Because A",
		mustPostgresQuestionPosition(positionValue),
	)
	if err != nil {
		t.Fatalf("expected question fixture, got %v", err)
	}

	return question
}

func mustPostgresMultipleQuestion(t *testing.T, idValue string, positionValue int) domain.ChoiceQuestion {
	t.Helper()

	question, err := domain.NewChoiceQuestion(
		mustPostgresQuestionID(idValue),
		domain.MultipleChoice(),
		"Pick many",
		[]string{"A", "B", "C"},
		[]int{0, 2},
		"Because A and C",
		mustPostgresQuestionPosition(positionValue),
	)
	if err != nil {
		t.Fatalf("expected multiple-choice question fixture, got %v", err)
	}

	return question
}

func mustPostgresQuestionID(value string) domain.QuestionID {
	id, err := domain.NewQuestionID(value)
	if err != nil {
		panic(err)
	}

	return id
}

func mustPostgresQuestionPosition(value int) domain.QuestionPosition {
	position, err := domain.NewQuestionPosition(value)
	if err != nil {
		panic(err)
	}

	return position
}
