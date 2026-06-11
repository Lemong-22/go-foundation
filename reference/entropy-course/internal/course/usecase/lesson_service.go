package usecase

import (
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

type LessonService struct {
	courses   core.CourseRepository
	lessons   core.LessonRepository
	quizzes   core.QuizRepository
	practices core.PracticeRepository
	ids       core.IDGenerator
	clock     core.Clock
}

func NewLessonService(
	courses core.CourseRepository,
	lessons core.LessonRepository,
	quizzes core.QuizRepository,
	ids core.IDGenerator,
	clock core.Clock,
	practices ...core.PracticeRepository,
) *LessonService {
	var practiceRepo core.PracticeRepository
	if len(practices) > 0 {
		practiceRepo = practices[0]
	}

	return &LessonService{
		courses:   courses,
		lessons:   lessons,
		quizzes:   quizzes,
		practices: practiceRepo,
		ids:       ids,
		clock:     clock,
	}
}

func (service *LessonService) CreateLesson(in core.CreateLessonInput) (core.CreateLessonOutput, error) {
	courseID, err := domain.NewCourseID(in.CourseID)
	if err != nil {
		return core.CreateLessonOutput{}, err
	}

	if _, err := service.courses.FindByID(courseID); err != nil {
		return core.CreateLessonOutput{}, err
	}

	order, err := service.lessonOrder(courseID, in.Order)
	if err != nil {
		return core.CreateLessonOutput{}, err
	}

	id := service.ids.NewLessonID()
	lesson, err := domain.NewLesson(id, courseID, in.Title, nil, order, service.clock.Now())
	if err != nil {
		return core.CreateLessonOutput{}, err
	}

	if err := service.lessons.Save(lesson); err != nil {
		return core.CreateLessonOutput{}, err
	}

	return core.CreateLessonOutput{ID: id.String()}, nil
}

func (service *LessonService) ListLessons(in core.ListLessonsInput) (core.ListLessonsOutput, error) {
	courseID, err := domain.NewCourseID(in.CourseID)
	if err != nil {
		return core.ListLessonsOutput{}, err
	}

	if _, err := service.courses.FindByID(courseID); err != nil {
		return core.ListLessonsOutput{}, err
	}

	lessons, err := service.lessons.FindByCourse(courseID)
	if err != nil {
		return core.ListLessonsOutput{}, err
	}

	views := make([]core.LessonView, 0, len(lessons))
	for _, lesson := range lessons {
		views = append(views, lessonView(lesson))
	}

	return core.ListLessonsOutput{Lessons: views}, nil
}

func (service *LessonService) GetLesson(in core.GetLessonInput) (core.GetLessonOutput, error) {
	id, err := domain.NewLessonID(in.ID)
	if err != nil {
		return core.GetLessonOutput{}, err
	}

	lesson, err := service.lessons.FindByID(id)
	if err != nil {
		return core.GetLessonOutput{}, err
	}

	return core.GetLessonOutput{Lesson: lessonView(lesson)}, nil
}

func (service *LessonService) UpdateLesson(in core.UpdateLessonInput) (core.UpdateLessonOutput, error) {
	id, err := domain.NewLessonID(in.ID)
	if err != nil {
		return core.UpdateLessonOutput{}, err
	}

	if in.Title == nil {
		return core.UpdateLessonOutput{}, domain.NewValidationError("update", "nothing to update")
	}

	lesson, err := service.lessons.FindByID(id)
	if err != nil {
		return core.UpdateLessonOutput{}, err
	}

	now := service.clock.Now()
	if in.Title != nil {
		if err := lesson.Rename(*in.Title, now); err != nil {
			return core.UpdateLessonOutput{}, err
		}
	}

	if err := service.lessons.Save(lesson); err != nil {
		return core.UpdateLessonOutput{}, err
	}

	return core.UpdateLessonOutput{ID: id.String()}, nil
}

func (service *LessonService) DeleteLesson(in core.DeleteLessonInput) error {
	id, err := domain.NewLessonID(in.ID)
	if err != nil {
		return err
	}

	return service.lessons.Delete(id)
}

func (service *LessonService) ReorderLessons(in core.ReorderLessonsInput) error {
	courseID, err := domain.NewCourseID(in.CourseID)
	if err != nil {
		return err
	}

	if _, err := service.courses.FindByID(courseID); err != nil {
		return err
	}

	lessons, err := service.lessons.FindByCourse(courseID)
	if err != nil {
		return err
	}

	indexedLessons := indexLessonsByID(lessons)
	updates, err := lessonOrderUpdates(in.Order, courseID, indexedLessons)
	if err != nil {
		return err
	}

	now := service.clock.Now()
	updatedLessons := make([]domain.Lesson, 0, len(updates))
	for _, update := range updates {
		lesson := update.lesson
		lesson.MoveTo(update.order, now)
		updatedLessons = append(updatedLessons, lesson)
	}

	return service.lessons.SaveAll(updatedLessons)
}

func (service *LessonService) AddLessonBlock(in core.AddLessonBlockInput) (core.AddLessonBlockOutput, error) {
	lessonID, err := domain.NewLessonID(in.LessonID)
	if err != nil {
		return core.AddLessonBlockOutput{}, err
	}

	lesson, err := service.lessons.FindByID(lessonID)
	if err != nil {
		return core.AddLessonBlockOutput{}, err
	}

	kind, err := domain.NewContentBlockKind(in.Kind)
	if err != nil {
		return core.AddLessonBlockOutput{}, err
	}

	body, err := service.blockBody(kind, lesson, in)
	if err != nil {
		return core.AddLessonBlockOutput{}, err
	}

	position, err := service.blockPosition(lesson, in.Position)
	if err != nil {
		return core.AddLessonBlockOutput{}, err
	}

	id := service.ids.NewBlockID()
	block, err := domain.NewContentBlock(id, kind, position, body)
	if err != nil {
		return core.AddLessonBlockOutput{}, err
	}

	if err := lesson.AddBlock(block, service.clock.Now()); err != nil {
		return core.AddLessonBlockOutput{}, err
	}

	if err := service.lessons.Save(lesson); err != nil {
		return core.AddLessonBlockOutput{}, err
	}

	return core.AddLessonBlockOutput{ID: id.String()}, nil
}

func (service *LessonService) ListLessonBlocks(in core.ListLessonBlocksInput) (core.ListLessonBlocksOutput, error) {
	lessonID, err := domain.NewLessonID(in.LessonID)
	if err != nil {
		return core.ListLessonBlocksOutput{}, err
	}

	lesson, err := service.lessons.FindByID(lessonID)
	if err != nil {
		return core.ListLessonBlocksOutput{}, err
	}

	return core.ListLessonBlocksOutput{Blocks: blockViews(lesson)}, nil
}

func (service *LessonService) GetLessonBlock(in core.GetLessonBlockInput) (core.GetLessonBlockOutput, error) {
	blockID, err := domain.NewBlockID(in.ID)
	if err != nil {
		return core.GetLessonBlockOutput{}, err
	}

	lesson, err := service.lessons.FindByBlockID(blockID)
	if err != nil {
		return core.GetLessonBlockOutput{}, err
	}

	block, err := lesson.Block(blockID)
	if err != nil {
		return core.GetLessonBlockOutput{}, err
	}

	return core.GetLessonBlockOutput{Block: blockView(lesson.ID(), block)}, nil
}

func (service *LessonService) UpdateLessonBlock(in core.UpdateLessonBlockInput) (core.UpdateLessonBlockOutput, error) {
	blockID, err := domain.NewBlockID(in.ID)
	if err != nil {
		return core.UpdateLessonBlockOutput{}, err
	}
	if !hasBlockUpdate(in) {
		return core.UpdateLessonBlockOutput{}, domain.NewValidationError("update", "must include at least one field")
	}

	lesson, err := service.lessons.FindByBlockID(blockID)
	if err != nil {
		return core.UpdateLessonBlockOutput{}, err
	}

	block, err := lesson.Block(blockID)
	if err != nil {
		return core.UpdateLessonBlockOutput{}, err
	}

	body, err := updatedBlockBody(block, in)
	if err != nil {
		return core.UpdateLessonBlockOutput{}, err
	}

	if err := lesson.UpdateBlock(blockID, body, service.clock.Now()); err != nil {
		return core.UpdateLessonBlockOutput{}, err
	}

	if err := service.lessons.Save(lesson); err != nil {
		return core.UpdateLessonBlockOutput{}, err
	}

	return core.UpdateLessonBlockOutput{ID: blockID.String()}, nil
}

func (service *LessonService) RemoveLessonBlock(in core.RemoveLessonBlockInput) error {
	blockID, err := domain.NewBlockID(in.ID)
	if err != nil {
		return err
	}

	lesson, err := service.lessons.FindByBlockID(blockID)
	if err != nil {
		return err
	}

	if err := lesson.RemoveBlock(blockID, service.clock.Now()); err != nil {
		return err
	}

	return service.lessons.Save(lesson)
}

func (service *LessonService) ReorderLessonBlocks(in core.ReorderLessonBlocksInput) error {
	lessonID, err := domain.NewLessonID(in.LessonID)
	if err != nil {
		return err
	}

	lesson, err := service.lessons.FindByID(lessonID)
	if err != nil {
		return err
	}

	placements, err := blockPlacements(in.Order)
	if err != nil {
		return err
	}

	if err := lesson.ReorderBlocks(placements, service.clock.Now()); err != nil {
		return err
	}

	return service.lessons.Save(lesson)
}

func (service *LessonService) blockPosition(lesson domain.Lesson, explicitPosition *int) (domain.BlockPosition, error) {
	if explicitPosition != nil {
		return domain.NewBlockPosition(*explicitPosition)
	}

	return domain.NewBlockPosition(nextBlockPosition(lesson.Blocks()))
}

func (service *LessonService) lessonOrder(courseID domain.CourseID, explicitOrder *int) (domain.LessonOrder, error) {
	if explicitOrder != nil {
		return domain.NewLessonOrder(*explicitOrder)
	}

	lessons, err := service.lessons.FindByCourse(courseID)
	if err != nil {
		return domain.LessonOrder{}, err
	}

	return domain.NewLessonOrder(nextLessonOrder(lessons))
}

type lessonOrderUpdate struct {
	lesson domain.Lesson
	order  domain.LessonOrder
}

func lessonOrderUpdates(
	positions []core.LessonPosition,
	courseID domain.CourseID,
	lessons map[string]domain.Lesson,
) ([]lessonOrderUpdate, error) {
	updates := make([]lessonOrderUpdate, 0, len(positions))
	usedPositions := make(map[int]struct{}, len(positions))

	for _, position := range positions {
		lessonID, err := domain.NewLessonID(position.LessonID)
		if err != nil {
			return nil, err
		}

		lesson, exists := lessons[lessonID.String()]
		if !exists || lesson.CourseID() != courseID {
			return nil, domain.NewValidationError("lesson_id", "must belong to the course")
		}

		order, err := domain.NewLessonOrder(position.Position)
		if err != nil {
			return nil, err
		}

		if _, exists := usedPositions[position.Position]; exists {
			return nil, domain.NewValidationError("position", "must be unique")
		}
		usedPositions[position.Position] = struct{}{}

		updates = append(updates, lessonOrderUpdate{lesson: lesson, order: order})
	}

	return updates, nil
}

func indexLessonsByID(lessons []domain.Lesson) map[string]domain.Lesson {
	indexed := make(map[string]domain.Lesson, len(lessons))
	for _, lesson := range lessons {
		indexed[lesson.ID().String()] = lesson
	}

	return indexed
}

func nextLessonOrder(lessons []domain.Lesson) int {
	maxOrder := -1
	for _, lesson := range lessons {
		if lesson.Order().Int() > maxOrder {
			maxOrder = lesson.Order().Int()
		}
	}

	return maxOrder + 1
}

func lessonView(lesson domain.Lesson) core.LessonView {
	return core.LessonView{
		ID:        lesson.ID().String(),
		CourseID:  lesson.CourseID().String(),
		Title:     lesson.Title(),
		Order:     lesson.Order().Int(),
		CreatedAt: lesson.CreatedAt(),
		UpdatedAt: lesson.UpdatedAt(),
	}
}

func (service *LessonService) blockBody(
	kind domain.ContentBlockKind,
	lesson domain.Lesson,
	in core.AddLessonBlockInput,
) (domain.ContentBody, error) {
	switch {
	case kind.IsText():
		return domain.TextBody{Markdown: in.Markdown}, nil
	case kind.IsVideo():
		provider, err := domain.NewMediaProvider(in.VideoProvider)
		if err != nil {
			return nil, err
		}

		media, err := domain.NewMediaRef(provider, in.VideoLocator)
		if err != nil {
			return nil, err
		}

		return domain.VideoBody{Media: media, Caption: in.VideoCaption}, nil
	case kind.IsQuiz():
		quizID, err := domain.NewQuizID(in.QuizRef)
		if err != nil {
			return nil, err
		}

		quiz, err := service.quizzes.FindByID(quizID)
		if err != nil {
			return nil, err
		}
		if quiz.CourseID() != lesson.CourseID() {
			return nil, domain.NewValidationError("quiz_ref", "must belong to the lesson course")
		}

		return domain.QuizBody{QuizRef: quizID}, nil
	case kind.IsPractice():
		practiceID, err := domain.NewPracticeID(in.PracticeRef)
		if err != nil {
			return nil, err
		}

		if service.practices == nil {
			return nil, domain.NewValidationError("practice_ref", "practice repository is not configured")
		}

		practice, err := service.practices.FindByID(practiceID)
		if err != nil {
			return nil, err
		}
		if practice.CourseID() != lesson.CourseID() {
			return nil, domain.NewValidationError("practice_ref", "must belong to the lesson course")
		}

		return domain.PracticeBody{PracticeRef: practiceID}, nil
	default:
		return nil, domain.NewValidationError("kind", "must be text, video, quiz, or practice")
	}
}

func nextBlockPosition(blocks []domain.ContentBlock) int {
	maxPosition := -1
	for _, block := range blocks {
		if block.Position().Int() > maxPosition {
			maxPosition = block.Position().Int()
		}
	}

	return maxPosition + 1
}

func blockViews(lesson domain.Lesson) []core.BlockView {
	blocks := lesson.Blocks()
	views := make([]core.BlockView, 0, len(blocks))
	for _, block := range blocks {
		views = append(views, blockView(lesson.ID(), block))
	}

	return views
}

func blockView(lessonID domain.LessonID, block domain.ContentBlock) core.BlockView {
	view := core.BlockView{
		ID:       block.ID().String(),
		LessonID: lessonID.String(),
		Kind:     block.Kind().String(),
		Position: block.Position().Int(),
	}

	switch body := block.Body().(type) {
	case domain.TextBody:
		view.Markdown = body.Markdown
	case domain.VideoBody:
		view.VideoProvider = body.Media.Provider().String()
		view.VideoLocator = body.Media.Locator()
		view.VideoCaption = body.Caption
	case domain.QuizBody:
		view.QuizRef = body.QuizRef.String()
	case domain.PracticeBody:
		view.PracticeRef = body.PracticeRef.String()
	}

	return view
}

func hasBlockUpdate(in core.UpdateLessonBlockInput) bool {
	return in.Markdown != nil ||
		in.VideoProvider != nil ||
		in.VideoLocator != nil ||
		in.VideoCaption != nil
}

func updatedBlockBody(block domain.ContentBlock, in core.UpdateLessonBlockInput) (domain.ContentBody, error) {
	switch body := block.Body().(type) {
	case domain.TextBody:
		if in.Markdown != nil {
			body.Markdown = *in.Markdown
		}

		return body, nil
	case domain.VideoBody:
		provider := body.Media.Provider()
		if in.VideoProvider != nil {
			updatedProvider, err := domain.NewMediaProvider(*in.VideoProvider)
			if err != nil {
				return nil, err
			}
			provider = updatedProvider
		}

		locator := body.Media.Locator()
		if in.VideoLocator != nil {
			locator = *in.VideoLocator
		}

		media, err := domain.NewMediaRef(provider, locator)
		if err != nil {
			return nil, err
		}

		caption := body.Caption
		if in.VideoCaption != nil {
			caption = *in.VideoCaption
		}

		return domain.VideoBody{Media: media, Caption: caption}, nil
	default:
		return nil, domain.NewValidationError("body", "must be text or video")
	}
}

func blockPlacements(dtos []core.BlockPlacementDTO) ([]domain.BlockPlacement, error) {
	placements := make([]domain.BlockPlacement, 0, len(dtos))
	for _, dto := range dtos {
		blockID, err := domain.NewBlockID(dto.BlockID)
		if err != nil {
			return nil, err
		}

		position, err := domain.NewBlockPosition(dto.Position)
		if err != nil {
			return nil, err
		}

		placements = append(placements, domain.BlockPlacement{
			BlockID:  blockID,
			Position: position,
		})
	}

	return placements, nil
}
