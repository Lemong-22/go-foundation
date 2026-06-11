package usecase

import (
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

type QuizService struct {
	courses core.CourseRepository
	lessons core.LessonRepository
	quizzes core.QuizRepository
	ids     core.IDGenerator
	clock   core.Clock
}

func NewQuizService(
	courses core.CourseRepository,
	lessons core.LessonRepository,
	quizzes core.QuizRepository,
	ids core.IDGenerator,
	clock core.Clock,
) *QuizService {
	return &QuizService{
		courses: courses,
		lessons: lessons,
		quizzes: quizzes,
		ids:     ids,
		clock:   clock,
	}
}

func (service *QuizService) CreateQuiz(in core.CreateQuizInput) (core.CreateQuizOutput, error) {
	courseID, err := domain.NewCourseID(in.CourseID)
	if err != nil {
		return core.CreateQuizOutput{}, err
	}

	if _, err := service.courses.FindByID(courseID); err != nil {
		return core.CreateQuizOutput{}, err
	}

	threshold := domain.DefaultPassThreshold()
	if in.PassThreshold != nil {
		threshold, err = domain.NewPassThreshold(*in.PassThreshold)
		if err != nil {
			return core.CreateQuizOutput{}, err
		}
	}

	id := service.ids.NewQuizID()
	quiz, err := domain.NewQuiz(id, courseID, in.Title, threshold, nil, service.clock.Now())
	if err != nil {
		return core.CreateQuizOutput{}, err
	}

	if err := service.quizzes.Save(quiz); err != nil {
		return core.CreateQuizOutput{}, err
	}

	return core.CreateQuizOutput{ID: id.String()}, nil
}

func (service *QuizService) ListQuizzes(in core.ListQuizzesInput) (core.ListQuizzesOutput, error) {
	courseID, err := domain.NewCourseID(in.CourseID)
	if err != nil {
		return core.ListQuizzesOutput{}, err
	}

	if _, err := service.courses.FindByID(courseID); err != nil {
		return core.ListQuizzesOutput{}, err
	}

	quizzes, err := service.quizzes.FindByCourse(courseID)
	if err != nil {
		return core.ListQuizzesOutput{}, err
	}

	views := make([]core.QuizView, 0, len(quizzes))
	for _, quiz := range quizzes {
		views = append(views, quizView(quiz))
	}

	return core.ListQuizzesOutput{Quizzes: views}, nil
}

func (service *QuizService) GetQuiz(in core.GetQuizInput) (core.GetQuizOutput, error) {
	id, err := domain.NewQuizID(in.ID)
	if err != nil {
		return core.GetQuizOutput{}, err
	}

	quiz, err := service.quizzes.FindByID(id)
	if err != nil {
		return core.GetQuizOutput{}, err
	}

	return core.GetQuizOutput{Quiz: quizDetailView(quiz)}, nil
}

func (service *QuizService) UpdateQuiz(in core.UpdateQuizInput) (core.UpdateQuizOutput, error) {
	id, err := domain.NewQuizID(in.ID)
	if err != nil {
		return core.UpdateQuizOutput{}, err
	}

	if in.Title == nil && in.PassThreshold == nil {
		return core.UpdateQuizOutput{}, domain.NewValidationError("update", "must include at least one field")
	}

	quiz, err := service.quizzes.FindByID(id)
	if err != nil {
		return core.UpdateQuizOutput{}, err
	}

	now := service.clock.Now()
	if in.Title != nil {
		if err := quiz.Rename(*in.Title, now); err != nil {
			return core.UpdateQuizOutput{}, err
		}
	}

	if in.PassThreshold != nil {
		threshold, err := domain.NewPassThreshold(*in.PassThreshold)
		if err != nil {
			return core.UpdateQuizOutput{}, err
		}
		quiz.ChangePassThreshold(threshold, now)
	}

	if err := service.quizzes.Save(quiz); err != nil {
		return core.UpdateQuizOutput{}, err
	}

	return core.UpdateQuizOutput{ID: id.String()}, nil
}

func (service *QuizService) DeleteQuiz(in core.DeleteQuizInput) error {
	id, err := domain.NewQuizID(in.ID)
	if err != nil {
		return err
	}

	if _, err := service.quizzes.FindByID(id); err != nil {
		return err
	}

	lessonIDs, err := service.lessons.FindLessonsEmbeddingQuiz(id)
	if err != nil {
		return err
	}
	if len(lessonIDs) > 0 {
		return domain.NewQuizInUseError(lessonIDs)
	}

	return service.quizzes.Delete(id)
}

func (service *QuizService) AddQuestion(in core.AddQuestionInput) (core.AddQuestionOutput, error) {
	quizID, err := domain.NewQuizID(in.QuizID)
	if err != nil {
		return core.AddQuestionOutput{}, err
	}

	quiz, err := service.quizzes.FindByID(quizID)
	if err != nil {
		return core.AddQuestionOutput{}, err
	}

	questionType, err := domain.NewChoiceQuestionType(in.Type)
	if err != nil {
		return core.AddQuestionOutput{}, err
	}

	position, err := service.questionPosition(quiz, in.Position)
	if err != nil {
		return core.AddQuestionOutput{}, err
	}

	id := service.ids.NewQuestionID()
	question, err := domain.NewChoiceQuestion(
		id,
		questionType,
		in.Prompt,
		in.Options,
		in.CorrectIndices,
		in.Explanation,
		position,
	)
	if err != nil {
		return core.AddQuestionOutput{}, err
	}

	if err := quiz.AddQuestion(question, service.clock.Now()); err != nil {
		return core.AddQuestionOutput{}, err
	}

	if err := service.quizzes.Save(quiz); err != nil {
		return core.AddQuestionOutput{}, err
	}

	return core.AddQuestionOutput{ID: id.String()}, nil
}

func (service *QuizService) ListQuestions(in core.ListQuestionsInput) (core.ListQuestionsOutput, error) {
	quizID, err := domain.NewQuizID(in.QuizID)
	if err != nil {
		return core.ListQuestionsOutput{}, err
	}

	quiz, err := service.quizzes.FindByID(quizID)
	if err != nil {
		return core.ListQuestionsOutput{}, err
	}

	return core.ListQuestionsOutput{Questions: questionViews(quiz)}, nil
}

func (service *QuizService) GetQuestion(in core.GetQuestionInput) (core.GetQuestionOutput, error) {
	id, err := domain.NewQuestionID(in.ID)
	if err != nil {
		return core.GetQuestionOutput{}, err
	}

	quiz, err := service.quizzes.FindByQuestionID(id)
	if err != nil {
		return core.GetQuestionOutput{}, err
	}

	question, err := quiz.Question(id)
	if err != nil {
		return core.GetQuestionOutput{}, err
	}

	return core.GetQuestionOutput{Question: questionView(quiz.ID(), question)}, nil
}

func (service *QuizService) UpdateQuestion(in core.UpdateQuestionInput) (core.UpdateQuestionOutput, error) {
	id, err := domain.NewQuestionID(in.ID)
	if err != nil {
		return core.UpdateQuestionOutput{}, err
	}

	if in.Prompt == nil && in.Options == nil && in.CorrectIndices == nil && in.Explanation == nil {
		return core.UpdateQuestionOutput{}, domain.NewValidationError("update", "must include at least one field")
	}
	if in.Options != nil && in.CorrectIndices == nil {
		return core.UpdateQuestionOutput{}, domain.NewValidationError("correct_indices", "must be provided when options are updated")
	}
	if in.Options == nil && in.CorrectIndices != nil {
		return core.UpdateQuestionOutput{}, domain.NewValidationError("options", "must be provided when correct indices are updated")
	}

	quiz, err := service.quizzes.FindByQuestionID(id)
	if err != nil {
		return core.UpdateQuestionOutput{}, err
	}

	now := service.clock.Now()
	if in.Prompt != nil {
		if err := quiz.ChangeQuestionPrompt(id, *in.Prompt, now); err != nil {
			return core.UpdateQuestionOutput{}, err
		}
	}
	if in.Options != nil {
		if err := quiz.ChangeQuestionContent(id, *in.Options, *in.CorrectIndices, now); err != nil {
			return core.UpdateQuestionOutput{}, err
		}
	}
	if in.Explanation != nil {
		if err := quiz.ChangeQuestionExplanation(id, *in.Explanation, now); err != nil {
			return core.UpdateQuestionOutput{}, err
		}
	}

	if err := service.quizzes.Save(quiz); err != nil {
		return core.UpdateQuestionOutput{}, err
	}

	return core.UpdateQuestionOutput{ID: id.String()}, nil
}

func (service *QuizService) RemoveQuestion(in core.RemoveQuestionInput) error {
	id, err := domain.NewQuestionID(in.ID)
	if err != nil {
		return err
	}

	quiz, err := service.quizzes.FindByQuestionID(id)
	if err != nil {
		return err
	}

	if err := quiz.RemoveQuestion(id, service.clock.Now()); err != nil {
		return err
	}

	return service.quizzes.Save(quiz)
}

func (service *QuizService) ReorderQuestions(in core.ReorderQuestionsInput) error {
	quizID, err := domain.NewQuizID(in.QuizID)
	if err != nil {
		return err
	}

	quiz, err := service.quizzes.FindByID(quizID)
	if err != nil {
		return err
	}

	placements, err := questionPlacements(in.Order, quiz)
	if err != nil {
		return err
	}

	if err := quiz.ReorderQuestions(placements, service.clock.Now()); err != nil {
		return err
	}

	return service.quizzes.Save(quiz)
}

func (service *QuizService) questionPosition(quiz domain.Quiz, explicitPosition *int) (domain.QuestionPosition, error) {
	if explicitPosition != nil {
		return domain.NewQuestionPosition(*explicitPosition)
	}

	return domain.NewQuestionPosition(nextQuestionPosition(quiz.Questions()))
}

func quizView(quiz domain.Quiz) core.QuizView {
	return core.QuizView{
		ID:            quiz.ID().String(),
		CourseID:      quiz.CourseID().String(),
		Title:         quiz.Title(),
		PassThreshold: quiz.PassThreshold().Float64(),
		QuestionCount: len(quiz.Questions()),
		CreatedAt:     quiz.CreatedAt(),
		UpdatedAt:     quiz.UpdatedAt(),
	}
}

func quizDetailView(quiz domain.Quiz) core.QuizDetailView {
	return core.QuizDetailView{
		QuizView:  quizView(quiz),
		Questions: questionViews(quiz),
	}
}

func questionViews(quiz domain.Quiz) []core.QuestionView {
	questions := quiz.Questions()
	views := make([]core.QuestionView, 0, len(questions))
	for _, question := range questions {
		views = append(views, questionView(quiz.ID(), question))
	}

	return views
}

func questionView(quizID domain.QuizID, question domain.ChoiceQuestion) core.QuestionView {
	return core.QuestionView{
		ID:             question.ID().String(),
		QuizID:         quizID.String(),
		Type:           question.Type().String(),
		Prompt:         question.Prompt(),
		Explanation:    question.Explanation(),
		Options:        question.Options(),
		CorrectIndices: question.CorrectIndices(),
		Position:       question.Position().Int(),
	}
}

func nextQuestionPosition(questions []domain.ChoiceQuestion) int {
	maxPosition := -1
	for _, question := range questions {
		if question.Position().Int() > maxPosition {
			maxPosition = question.Position().Int()
		}
	}

	return maxPosition + 1
}

func questionPlacements(dtos []core.QuestionPlacementDTO, quiz domain.Quiz) ([]domain.QuestionPlacement, error) {
	current := make(map[string]struct{}, len(quiz.Questions()))
	for _, question := range quiz.Questions() {
		current[question.ID().String()] = struct{}{}
	}

	placements := make([]domain.QuestionPlacement, 0, len(dtos))
	for _, dto := range dtos {
		questionID, err := domain.NewQuestionID(dto.QuestionID)
		if err != nil {
			return nil, err
		}
		if _, exists := current[questionID.String()]; !exists {
			return nil, domain.NewValidationError("question_id", "must belong to the quiz")
		}

		position, err := domain.NewQuestionPosition(dto.Position)
		if err != nil {
			return nil, err
		}

		placements = append(placements, domain.QuestionPlacement{
			QuestionID: questionID,
			Position:   position,
		})
	}

	return placements, nil
}
