package domain

import (
	"math"
	"sort"
	"strings"
	"time"
)

const (
	choiceQuestionTypeSingle   = "single"
	choiceQuestionTypeMultiple = "multiple"
)

type QuestionPosition struct {
	value int
}

func NewQuestionPosition(value int) (QuestionPosition, error) {
	if value < 0 {
		return QuestionPosition{}, NewValidationError("position", "must be greater than or equal to zero")
	}

	return QuestionPosition{value: value}, nil
}

func (position QuestionPosition) Int() int {
	return position.value
}

type ChoiceQuestionType struct {
	value string
}

func SingleChoice() ChoiceQuestionType {
	return ChoiceQuestionType{value: choiceQuestionTypeSingle}
}

func MultipleChoice() ChoiceQuestionType {
	return ChoiceQuestionType{value: choiceQuestionTypeMultiple}
}

func NewChoiceQuestionType(value string) (ChoiceQuestionType, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case choiceQuestionTypeSingle:
		return SingleChoice(), nil
	case choiceQuestionTypeMultiple:
		return MultipleChoice(), nil
	default:
		return ChoiceQuestionType{}, NewValidationError("question_type", "must be single or multiple")
	}
}

func (questionType ChoiceQuestionType) String() string {
	return questionType.value
}

func (questionType ChoiceQuestionType) IsSingle() bool {
	return questionType.value == choiceQuestionTypeSingle
}

func (questionType ChoiceQuestionType) IsMultiple() bool {
	return questionType.value == choiceQuestionTypeMultiple
}

type PassThreshold struct {
	value float64
}

func NewPassThreshold(value float64) (PassThreshold, error) {
	if math.IsNaN(value) || value < 0 || value > 1 {
		return PassThreshold{}, NewValidationError("pass_threshold", "must be between zero and one")
	}

	return PassThreshold{value: value}, nil
}

func DefaultPassThreshold() PassThreshold {
	return PassThreshold{value: 0.7}
}

func (threshold PassThreshold) Float64() float64 {
	return threshold.value
}

type ChoiceQuestion struct {
	id             QuestionID
	questionType   ChoiceQuestionType
	prompt         string
	options        []string
	correctIndices []int
	explanation    string
	position       QuestionPosition
}

func NewChoiceQuestion(
	id QuestionID,
	questionType ChoiceQuestionType,
	prompt string,
	options []string,
	correctIndices []int,
	explanation string,
	position QuestionPosition,
) (ChoiceQuestion, error) {
	normalizedPrompt, err := normalizeQuestionPrompt(prompt)
	if err != nil {
		return ChoiceQuestion{}, err
	}
	if err := validateQuestionIdentity(id); err != nil {
		return ChoiceQuestion{}, err
	}
	if err := validateChoiceQuestionType(questionType); err != nil {
		return ChoiceQuestion{}, err
	}
	if err := validateQuestionContent(questionType, options, correctIndices); err != nil {
		return ChoiceQuestion{}, err
	}

	return ChoiceQuestion{
		id:             id,
		questionType:   questionType,
		prompt:         normalizedPrompt,
		options:        copyStrings(options),
		correctIndices: copyInts(correctIndices),
		explanation:    explanation,
		position:       position,
	}, nil
}

func (question ChoiceQuestion) ID() QuestionID {
	return question.id
}

func (question ChoiceQuestion) Type() ChoiceQuestionType {
	return question.questionType
}

func (question ChoiceQuestion) Prompt() string {
	return question.prompt
}

func (question ChoiceQuestion) Options() []string {
	return copyStrings(question.options)
}

func (question ChoiceQuestion) CorrectIndices() []int {
	return copyInts(question.correctIndices)
}

func (question ChoiceQuestion) Explanation() string {
	return question.explanation
}

func (question ChoiceQuestion) Position() QuestionPosition {
	return question.position
}

func (question *ChoiceQuestion) ChangePrompt(prompt string) error {
	normalizedPrompt, err := normalizeQuestionPrompt(prompt)
	if err != nil {
		return err
	}

	question.prompt = normalizedPrompt
	return nil
}

func (question *ChoiceQuestion) ChangeContent(options []string, correctIndices []int) error {
	if err := validateQuestionContent(question.questionType, options, correctIndices); err != nil {
		return err
	}

	question.options = copyStrings(options)
	question.correctIndices = copyInts(correctIndices)
	return nil
}

func (question *ChoiceQuestion) ChangeExplanation(explanation string) {
	question.explanation = explanation
}

func (question *ChoiceQuestion) MoveTo(position QuestionPosition) {
	question.position = position
}

func (question ChoiceQuestion) validate() error {
	if err := validateQuestionIdentity(question.id); err != nil {
		return err
	}
	if err := validateChoiceQuestionType(question.questionType); err != nil {
		return err
	}
	if _, err := normalizeQuestionPrompt(question.prompt); err != nil {
		return err
	}

	return validateQuestionContent(question.questionType, question.options, question.correctIndices)
}

type Quiz struct {
	id            QuizID
	courseID      CourseID
	title         string
	passThreshold PassThreshold
	questions     []ChoiceQuestion
	createdAt     time.Time
	updatedAt     time.Time
}

func NewQuiz(
	id QuizID,
	courseID CourseID,
	title string,
	passThreshold PassThreshold,
	questions []ChoiceQuestion,
	now time.Time,
) (Quiz, error) {
	return RestoreQuiz(id, courseID, title, passThreshold, questions, now, now)
}

func RestoreQuiz(
	id QuizID,
	courseID CourseID,
	title string,
	passThreshold PassThreshold,
	questions []ChoiceQuestion,
	createdAt time.Time,
	updatedAt time.Time,
) (Quiz, error) {
	if id.String() == "" {
		return Quiz{}, NewValidationError("quiz_id", "must not be empty")
	}
	if courseID.String() == "" {
		return Quiz{}, NewValidationError("course_id", "must not be empty")
	}
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return Quiz{}, err
	}
	if _, err := NewPassThreshold(passThreshold.Float64()); err != nil {
		return Quiz{}, err
	}
	if updatedAt.Before(createdAt) {
		return Quiz{}, NewValidationError("updated_at", "must be greater than or equal to created_at")
	}

	normalizedQuestions, err := normalizeQuizQuestions(questions)
	if err != nil {
		return Quiz{}, err
	}

	return Quiz{
		id:            id,
		courseID:      courseID,
		title:         normalizedTitle,
		passThreshold: passThreshold,
		questions:     normalizedQuestions,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

func (quiz Quiz) ID() QuizID {
	return quiz.id
}

func (quiz Quiz) CourseID() CourseID {
	return quiz.courseID
}

func (quiz Quiz) Title() string {
	return quiz.title
}

func (quiz Quiz) PassThreshold() PassThreshold {
	return quiz.passThreshold
}

func (quiz Quiz) Questions() []ChoiceQuestion {
	questions := make([]ChoiceQuestion, len(quiz.questions))
	copy(questions, quiz.questions)

	return questions
}

func (quiz Quiz) Question(id QuestionID) (ChoiceQuestion, error) {
	for _, question := range quiz.questions {
		if question.ID() == id {
			return question, nil
		}
	}

	return ChoiceQuestion{}, ErrNotFound
}

func (quiz Quiz) CreatedAt() time.Time {
	return quiz.createdAt
}

func (quiz Quiz) UpdatedAt() time.Time {
	return quiz.updatedAt
}

func (quiz *Quiz) Rename(title string, now time.Time) error {
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return err
	}

	quiz.title = normalizedTitle
	quiz.touch(now)

	return nil
}

func (quiz *Quiz) ChangePassThreshold(threshold PassThreshold, now time.Time) {
	quiz.passThreshold = threshold
	quiz.touch(now)
}

func (quiz *Quiz) AddQuestion(question ChoiceQuestion, now time.Time) error {
	if err := question.validate(); err != nil {
		return err
	}
	if quiz.hasQuestionID(question.ID()) {
		return NewValidationError("question_id", "must be unique within the quiz")
	}
	if question.Position().Int() > len(quiz.questions) {
		return NewValidationError("position", "must be less than or equal to question count")
	}

	questions := quiz.Questions()
	for i := range questions {
		if questions[i].Position().Int() >= question.Position().Int() {
			questions[i].MoveTo(QuestionPosition{value: questions[i].Position().Int() + 1})
		}
	}
	questions = append(questions, question)
	quiz.questions = sortQuestionsByPosition(questions)
	quiz.touch(now)

	return nil
}

func (quiz *Quiz) RemoveQuestion(id QuestionID, now time.Time) error {
	questions := make([]ChoiceQuestion, 0, len(quiz.questions))
	removed := false

	for _, question := range quiz.questions {
		if question.ID() == id {
			removed = true
			continue
		}

		question.MoveTo(QuestionPosition{value: len(questions)})
		questions = append(questions, question)
	}

	if !removed {
		return ErrNotFound
	}

	quiz.questions = questions
	quiz.touch(now)

	return nil
}

func (quiz *Quiz) ReorderQuestions(order []QuestionPlacement, now time.Time) error {
	if len(order) != len(quiz.questions) {
		return NewValidationError("order", "must include every question exactly once")
	}

	current := make(map[string]ChoiceQuestion, len(quiz.questions))
	for _, question := range quiz.questions {
		current[question.ID().String()] = question
	}

	usedQuestions := make(map[string]struct{}, len(order))
	usedPositions := make(map[int]struct{}, len(order))
	positions := make(map[string]QuestionPosition, len(order))

	for _, placement := range order {
		id := placement.QuestionID.String()
		if _, exists := current[id]; !exists {
			return NewValidationError("question_id", "must belong to the quiz")
		}
		if _, exists := usedQuestions[id]; exists {
			return NewValidationError("question_id", "must be unique")
		}

		position := placement.Position.Int()
		if position >= len(order) {
			return NewValidationError("position", "must be contiguous from zero")
		}
		if _, exists := usedPositions[position]; exists {
			return NewValidationError("position", "must be unique")
		}

		usedQuestions[id] = struct{}{}
		usedPositions[position] = struct{}{}
		positions[id] = placement.Position
	}

	questions := make([]ChoiceQuestion, 0, len(quiz.questions))
	for _, question := range quiz.questions {
		question.MoveTo(positions[question.ID().String()])
		questions = append(questions, question)
	}

	quiz.questions = sortQuestionsByPosition(questions)
	quiz.touch(now)

	return nil
}

func (quiz *Quiz) ChangeQuestionPrompt(id QuestionID, prompt string, now time.Time) error {
	questions := quiz.Questions()
	for i := range questions {
		if questions[i].ID() != id {
			continue
		}
		if err := questions[i].ChangePrompt(prompt); err != nil {
			return err
		}

		quiz.questions = questions
		quiz.touch(now)
		return nil
	}

	return ErrNotFound
}

func (quiz *Quiz) ChangeQuestionContent(id QuestionID, options []string, correctIndices []int, now time.Time) error {
	questions := quiz.Questions()
	for i := range questions {
		if questions[i].ID() != id {
			continue
		}
		if err := questions[i].ChangeContent(options, correctIndices); err != nil {
			return err
		}

		quiz.questions = questions
		quiz.touch(now)
		return nil
	}

	return ErrNotFound
}

func (quiz *Quiz) ChangeQuestionExplanation(id QuestionID, explanation string, now time.Time) error {
	questions := quiz.Questions()
	for i := range questions {
		if questions[i].ID() != id {
			continue
		}

		questions[i].ChangeExplanation(explanation)
		quiz.questions = questions
		quiz.touch(now)
		return nil
	}

	return ErrNotFound
}

func (quiz *Quiz) touch(now time.Time) {
	quiz.updatedAt = mutationTime(quiz.createdAt, now)
}

func (quiz Quiz) hasQuestionID(id QuestionID) bool {
	for _, question := range quiz.questions {
		if question.ID() == id {
			return true
		}
	}

	return false
}

type QuestionPlacement struct {
	QuestionID QuestionID
	Position   QuestionPosition
}

func normalizeQuestionPrompt(prompt string) (string, error) {
	trimmed := strings.TrimSpace(prompt)
	if trimmed == "" {
		return "", NewValidationError("prompt", "must not be empty")
	}

	return trimmed, nil
}

func validateQuestionIdentity(id QuestionID) error {
	if id.String() == "" {
		return NewValidationError("question_id", "must not be empty")
	}

	return nil
}

func validateChoiceQuestionType(questionType ChoiceQuestionType) error {
	switch questionType {
	case SingleChoice(), MultipleChoice():
		return nil
	default:
		return NewValidationError("question_type", "must be single or multiple")
	}
}

func validateQuestionContent(questionType ChoiceQuestionType, options []string, correctIndices []int) error {
	if len(options) < 2 {
		return NewValidationError("options", "must include at least two choices")
	}
	if len(correctIndices) == 0 {
		return NewValidationError("correct_indices", "must include at least one correct answer")
	}
	if questionType.IsSingle() && len(correctIndices) != 1 {
		return NewValidationError("correct_indices", "must include exactly one correct answer for single-choice questions")
	}

	seen := make(map[int]struct{}, len(correctIndices))
	for _, index := range correctIndices {
		if index < 0 || index >= len(options) {
			return NewValidationError("correct_indices", "must reference existing options")
		}
		if _, exists := seen[index]; exists {
			return NewValidationError("correct_indices", "must not contain duplicates")
		}

		seen[index] = struct{}{}
	}

	return nil
}

func normalizeQuizQuestions(questions []ChoiceQuestion) ([]ChoiceQuestion, error) {
	normalized := make([]ChoiceQuestion, len(questions))
	copy(normalized, questions)

	ids := make(map[string]struct{}, len(normalized))
	positions := make(map[int]struct{}, len(normalized))
	for _, question := range normalized {
		if err := question.validate(); err != nil {
			return nil, err
		}

		id := question.ID().String()
		if _, exists := ids[id]; exists {
			return nil, NewValidationError("question_id", "must be unique within the quiz")
		}
		ids[id] = struct{}{}

		position := question.Position().Int()
		if _, exists := positions[position]; exists {
			return nil, NewValidationError("position", "must be unique")
		}
		positions[position] = struct{}{}
	}

	normalized = sortQuestionsByPosition(normalized)
	for i, question := range normalized {
		if question.Position().Int() != i {
			return nil, NewValidationError("position", "must be contiguous from zero")
		}
	}

	return normalized, nil
}

func sortQuestionsByPosition(questions []ChoiceQuestion) []ChoiceQuestion {
	sort.Slice(questions, func(i, j int) bool {
		return questions[i].Position().Int() < questions[j].Position().Int()
	})

	return questions
}

func copyStrings(values []string) []string {
	copied := make([]string, len(values))
	copy(copied, values)

	return copied
}

func copyInts(values []int) []int {
	copied := make([]int, len(values))
	copy(copied, values)

	return copied
}
