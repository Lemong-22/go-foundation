package domain

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

const (
	quizIDValue      = "550e8400-e29b-41d4-a716-446655440005"
	questionIDValue  = "550e8400-e29b-41d4-a716-446655440006"
	otherQuestionID  = "550e8400-e29b-41d4-a716-446655440007"
	thirdQuestionID  = "550e8400-e29b-41d4-a716-446655440008"
	fourthQuestionID = "550e8400-e29b-41d4-a716-446655440009"
)

func TestNewQuizCreatesAggregate(t *testing.T) {
	now := time.Date(2026, 5, 26, 7, 0, 0, 0, time.UTC)
	questions := []ChoiceQuestion{mustChoiceQuestion(t, questionIDValue, 0)}

	quiz, err := NewQuiz(
		mustQuizID(t, quizIDValue),
		mustCourseID(t, validUUID),
		"  Go Basics Quiz  ",
		DefaultPassThreshold(),
		questions,
		now,
	)
	if err != nil {
		t.Fatalf("expected quiz, got error %v", err)
	}

	if quiz.ID().String() != quizIDValue {
		t.Fatalf("expected quiz id %q, got %q", quizIDValue, quiz.ID().String())
	}
	if quiz.CourseID().String() != validUUID {
		t.Fatalf("expected course id %q, got %q", validUUID, quiz.CourseID().String())
	}
	if quiz.Title() != "Go Basics Quiz" {
		t.Fatalf("expected trimmed title, got %q", quiz.Title())
	}
	if quiz.PassThreshold() != DefaultPassThreshold() {
		t.Fatalf("expected default pass threshold")
	}
	if got := quiz.Questions(); !reflect.DeepEqual(got, questions) {
		t.Fatalf("expected questions %+v, got %+v", questions, got)
	}
	if !quiz.CreatedAt().Equal(now) || !quiz.UpdatedAt().Equal(now) {
		t.Fatalf("expected created and updated timestamps to equal %v", now)
	}
}

func TestNewQuizRejectsInvalidState(t *testing.T) {
	now := time.Date(2026, 5, 26, 7, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		quizID    QuizID
		courseID  CourseID
		title     string
		questions []ChoiceQuestion
		createdAt time.Time
		updatedAt time.Time
	}{
		{
			name:      "empty quiz id",
			quizID:    QuizID{},
			courseID:  mustCourseID(t, validUUID),
			title:     "Quiz",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "empty course id",
			quizID:    mustQuizID(t, quizIDValue),
			courseID:  CourseID{},
			title:     "Quiz",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "empty title",
			quizID:    mustQuizID(t, quizIDValue),
			courseID:  mustCourseID(t, validUUID),
			title:     "   ",
			createdAt: now,
			updatedAt: now,
		},
		{
			name:      "updated before created",
			quizID:    mustQuizID(t, quizIDValue),
			courseID:  mustCourseID(t, validUUID),
			title:     "Quiz",
			createdAt: now,
			updatedAt: now.Add(-time.Minute),
		},
		{
			name:     "duplicate question ids",
			quizID:   mustQuizID(t, quizIDValue),
			courseID: mustCourseID(t, validUUID),
			title:    "Quiz",
			questions: []ChoiceQuestion{
				mustChoiceQuestion(t, questionIDValue, 0),
				mustChoiceQuestion(t, questionIDValue, 1),
			},
			createdAt: now,
			updatedAt: now,
		},
		{
			name:     "position gap",
			quizID:   mustQuizID(t, quizIDValue),
			courseID: mustCourseID(t, validUUID),
			title:    "Quiz",
			questions: []ChoiceQuestion{
				mustChoiceQuestion(t, questionIDValue, 0),
				mustChoiceQuestion(t, otherQuestionID, 2),
			},
			createdAt: now,
			updatedAt: now,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := RestoreQuiz(
				test.quizID,
				test.courseID,
				test.title,
				DefaultPassThreshold(),
				test.questions,
				test.createdAt,
				test.updatedAt,
			)
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestNewChoiceQuestionValidatesContent(t *testing.T) {
	tests := []struct {
		name           string
		questionType   ChoiceQuestionType
		prompt         string
		options        []string
		correctIndices []int
	}{
		{
			name:           "empty prompt",
			questionType:   SingleChoice(),
			prompt:         "   ",
			options:        []string{"A", "B"},
			correctIndices: []int{0},
		},
		{
			name:           "too few options",
			questionType:   SingleChoice(),
			prompt:         "Pick one",
			options:        []string{"A"},
			correctIndices: []int{0},
		},
		{
			name:           "no correct indices",
			questionType:   SingleChoice(),
			prompt:         "Pick one",
			options:        []string{"A", "B"},
			correctIndices: nil,
		},
		{
			name:           "out of range correct index",
			questionType:   SingleChoice(),
			prompt:         "Pick one",
			options:        []string{"A", "B"},
			correctIndices: []int{2},
		},
		{
			name:           "duplicate correct index",
			questionType:   MultipleChoice(),
			prompt:         "Pick many",
			options:        []string{"A", "B"},
			correctIndices: []int{0, 0},
		},
		{
			name:           "single choice with multiple correct indices",
			questionType:   SingleChoice(),
			prompt:         "Pick one",
			options:        []string{"A", "B"},
			correctIndices: []int{0, 1},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewChoiceQuestion(
				mustQuestionID(t, questionIDValue),
				test.questionType,
				test.prompt,
				test.options,
				test.correctIndices,
				"",
				mustQuestionPosition(t, 0),
			)
			if !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestQuizAddQuestionInsertsAndShiftsPositions(t *testing.T) {
	quiz := mustQuizWithQuestions(t,
		mustChoiceQuestion(t, questionIDValue, 0),
		mustChoiceQuestion(t, otherQuestionID, 1),
	)
	inserted := mustChoiceQuestion(t, thirdQuestionID, 1)
	changedAt := quiz.CreatedAt().Add(time.Hour)

	if err := quiz.AddQuestion(inserted, changedAt); err != nil {
		t.Fatalf("expected add question to succeed, got %v", err)
	}

	questions := quiz.Questions()
	if len(questions) != 3 {
		t.Fatalf("expected three questions, got %d", len(questions))
	}
	if questions[0].ID().String() != questionIDValue || questions[0].Position().Int() != 0 {
		t.Fatalf("expected original first question to stay at position 0")
	}
	if questions[1].ID().String() != thirdQuestionID || questions[1].Position().Int() != 1 {
		t.Fatalf("expected inserted question at position 1, got %+v", questions[1])
	}
	if questions[2].ID().String() != otherQuestionID || questions[2].Position().Int() != 2 {
		t.Fatalf("expected original second question to shift to position 2")
	}
	if !quiz.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}
}

func TestQuizAddQuestionRejectsDuplicateIDAndOutOfRangePosition(t *testing.T) {
	quiz := mustQuizWithQuestions(t, mustChoiceQuestion(t, questionIDValue, 0))

	if err := quiz.AddQuestion(mustChoiceQuestion(t, questionIDValue, 1), time.Now()); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected duplicate id validation error, got %v", err)
	}

	if err := quiz.AddQuestion(mustChoiceQuestion(t, otherQuestionID, 2), time.Now()); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected out-of-range position validation error, got %v", err)
	}
}

func TestQuizRemoveQuestionCompactsPositions(t *testing.T) {
	quiz := mustQuizWithQuestions(t,
		mustChoiceQuestion(t, questionIDValue, 0),
		mustChoiceQuestion(t, otherQuestionID, 1),
		mustChoiceQuestion(t, thirdQuestionID, 2),
	)
	changedAt := quiz.CreatedAt().Add(time.Hour)

	if err := quiz.RemoveQuestion(mustQuestionID(t, otherQuestionID), changedAt); err != nil {
		t.Fatalf("expected remove to succeed, got %v", err)
	}

	questions := quiz.Questions()
	if len(questions) != 2 {
		t.Fatalf("expected two questions, got %d", len(questions))
	}
	if questions[0].ID().String() != questionIDValue || questions[0].Position().Int() != 0 {
		t.Fatalf("expected first question at position 0")
	}
	if questions[1].ID().String() != thirdQuestionID || questions[1].Position().Int() != 1 {
		t.Fatalf("expected third question compacted to position 1")
	}
	if !quiz.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}
}

func TestQuizReorderQuestionsRequiresPermutation(t *testing.T) {
	quiz := mustQuizWithQuestions(t,
		mustChoiceQuestion(t, questionIDValue, 0),
		mustChoiceQuestion(t, otherQuestionID, 1),
	)
	changedAt := quiz.CreatedAt().Add(time.Hour)

	err := quiz.ReorderQuestions([]QuestionPlacement{
		{QuestionID: mustQuestionID(t, otherQuestionID), Position: mustQuestionPosition(t, 0)},
		{QuestionID: mustQuestionID(t, questionIDValue), Position: mustQuestionPosition(t, 1)},
	}, changedAt)
	if err != nil {
		t.Fatalf("expected reorder to succeed, got %v", err)
	}

	questions := quiz.Questions()
	if questions[0].ID().String() != otherQuestionID || questions[1].ID().String() != questionIDValue {
		t.Fatalf("expected questions to reorder by position")
	}
	if !quiz.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected updated timestamp")
	}

	tests := []struct {
		name  string
		order []QuestionPlacement
	}{
		{
			name:  "missing question",
			order: []QuestionPlacement{{QuestionID: mustQuestionID(t, questionIDValue), Position: mustQuestionPosition(t, 0)}},
		},
		{
			name: "unknown question",
			order: []QuestionPlacement{
				{QuestionID: mustQuestionID(t, questionIDValue), Position: mustQuestionPosition(t, 0)},
				{QuestionID: mustQuestionID(t, thirdQuestionID), Position: mustQuestionPosition(t, 1)},
			},
		},
		{
			name: "duplicate position",
			order: []QuestionPlacement{
				{QuestionID: mustQuestionID(t, questionIDValue), Position: mustQuestionPosition(t, 0)},
				{QuestionID: mustQuestionID(t, otherQuestionID), Position: mustQuestionPosition(t, 0)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			quiz := mustQuizWithQuestions(t,
				mustChoiceQuestion(t, questionIDValue, 0),
				mustChoiceQuestion(t, otherQuestionID, 1),
			)
			if err := quiz.ReorderQuestions(test.order, changedAt); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestQuizQuestionMutationsUpdateThroughAggregate(t *testing.T) {
	quiz := mustQuizWithQuestions(t, mustChoiceQuestion(t, questionIDValue, 0))
	changedAt := quiz.CreatedAt().Add(time.Hour)
	questionID := mustQuestionID(t, questionIDValue)

	if err := quiz.ChangeQuestionPrompt(questionID, "  Updated prompt?  ", changedAt); err != nil {
		t.Fatalf("expected prompt change to succeed, got %v", err)
	}
	question, err := quiz.Question(questionID)
	if err != nil {
		t.Fatalf("expected question to exist, got %v", err)
	}
	if question.Prompt() != "Updated prompt?" || !quiz.UpdatedAt().Equal(changedAt) {
		t.Fatalf("expected prompt and timestamp to update")
	}

	contentChangedAt := changedAt.Add(time.Hour)
	if err := quiz.ChangeQuestionContent(questionID, []string{"A", "B", "C"}, []int{2}, contentChangedAt); err != nil {
		t.Fatalf("expected content change to succeed, got %v", err)
	}
	question, err = quiz.Question(questionID)
	if err != nil {
		t.Fatalf("expected question to exist, got %v", err)
	}
	if !reflect.DeepEqual(question.Options(), []string{"A", "B", "C"}) || !reflect.DeepEqual(question.CorrectIndices(), []int{2}) {
		t.Fatalf("expected content to update atomically")
	}
	if !quiz.UpdatedAt().Equal(contentChangedAt) {
		t.Fatalf("expected content change timestamp")
	}

	updatedAt := quiz.UpdatedAt()
	if err := quiz.ChangeQuestionContent(questionID, []string{"A", "B"}, []int{0, 1}, updatedAt.Add(time.Hour)); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected invalid single-choice content validation error, got %v", err)
	}
	question, err = quiz.Question(questionID)
	if err != nil {
		t.Fatalf("expected question to exist, got %v", err)
	}
	if !reflect.DeepEqual(question.CorrectIndices(), []int{2}) || !quiz.UpdatedAt().Equal(updatedAt) {
		t.Fatalf("expected invalid content update to leave state unchanged")
	}

	explanationChangedAt := updatedAt.Add(2 * time.Hour)
	if err := quiz.ChangeQuestionExplanation(questionID, "Because it is correct.", explanationChangedAt); err != nil {
		t.Fatalf("expected explanation change to succeed, got %v", err)
	}
	question, err = quiz.Question(questionID)
	if err != nil {
		t.Fatalf("expected question to exist, got %v", err)
	}
	if question.Explanation() != "Because it is correct." || !quiz.UpdatedAt().Equal(explanationChangedAt) {
		t.Fatalf("expected explanation and timestamp to update")
	}
}

func TestQuizBodyKind(t *testing.T) {
	quizID := mustQuizID(t, quizIDValue)
	body := QuizBody{QuizRef: quizID}

	if body.Kind() != QuizKind() || body.QuizRef != quizID {
		t.Fatalf("expected quiz body to keep quiz ref and report quiz kind")
	}

	block, err := NewQuizBlock(mustBlockID(t, blockIDValue), mustBlockPosition(t, 0), quizID)
	if err != nil {
		t.Fatalf("expected quiz block, got error %v", err)
	}
	if block.Kind() != QuizKind() {
		t.Fatalf("expected quiz block kind, got %q", block.Kind().String())
	}
}

func mustQuizWithQuestions(t *testing.T, questions ...ChoiceQuestion) Quiz {
	t.Helper()

	quiz, err := NewQuiz(
		mustQuizID(t, quizIDValue),
		mustCourseID(t, validUUID),
		"Go Basics Quiz",
		DefaultPassThreshold(),
		questions,
		time.Date(2026, 5, 26, 7, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected quiz fixture, got error %v", err)
	}

	return quiz
}

func mustChoiceQuestion(t *testing.T, id string, position int) ChoiceQuestion {
	t.Helper()

	question, err := NewChoiceQuestion(
		mustQuestionID(t, id),
		SingleChoice(),
		"Pick one",
		[]string{"A", "B"},
		[]int{0},
		"",
		mustQuestionPosition(t, position),
	)
	if err != nil {
		t.Fatalf("expected question fixture, got error %v", err)
	}

	return question
}

func mustQuizID(t *testing.T, value string) QuizID {
	t.Helper()

	id, err := NewQuizID(value)
	if err != nil {
		t.Fatalf("expected quiz id fixture, got error %v", err)
	}

	return id
}

func mustQuestionID(t *testing.T, value string) QuestionID {
	t.Helper()

	id, err := NewQuestionID(value)
	if err != nil {
		t.Fatalf("expected question id fixture, got error %v", err)
	}

	return id
}

func mustQuestionPosition(t *testing.T, value int) QuestionPosition {
	t.Helper()

	position, err := NewQuestionPosition(value)
	if err != nil {
		t.Fatalf("expected question position fixture, got error %v", err)
	}

	return position
}
