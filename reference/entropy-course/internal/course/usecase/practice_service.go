package usecase

import (
	"strings"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

type PracticeService struct {
	courses   core.CourseRepository
	lessons   core.LessonRepository
	practices core.PracticeRepository
	ids       core.IDGenerator
	clock     core.Clock
}

func NewPracticeService(
	courses core.CourseRepository,
	lessons core.LessonRepository,
	practices core.PracticeRepository,
	ids core.IDGenerator,
	clock core.Clock,
) *PracticeService {
	return &PracticeService{
		courses:   courses,
		lessons:   lessons,
		practices: practices,
		ids:       ids,
		clock:     clock,
	}
}

func (service *PracticeService) CreatePractice(in core.CreatePracticeInput) (core.CreatePracticeOutput, error) {
	courseID, err := domain.NewCourseID(in.CourseID)
	if err != nil {
		return core.CreatePracticeOutput{}, err
	}

	if _, err := service.courses.FindByID(courseID); err != nil {
		return core.CreatePracticeOutput{}, err
	}

	language, err := domain.NewLanguage(in.Language)
	if err != nil {
		return core.CreatePracticeOutput{}, err
	}

	id := service.ids.NewPracticeID()
	practice, err := domain.NewPractice(
		id,
		courseID,
		in.Title,
		language,
		in.Prompt,
		in.StarterCode,
		in.Solution,
		nil,
		service.clock.Now(),
	)
	if err != nil {
		return core.CreatePracticeOutput{}, err
	}

	if err := service.practices.Save(practice); err != nil {
		return core.CreatePracticeOutput{}, err
	}

	return core.CreatePracticeOutput{ID: id.String()}, nil
}

func (service *PracticeService) ListPractices(in core.ListPracticesInput) (core.ListPracticesOutput, error) {
	courseID, err := domain.NewCourseID(in.CourseID)
	if err != nil {
		return core.ListPracticesOutput{}, err
	}

	if _, err := service.courses.FindByID(courseID); err != nil {
		return core.ListPracticesOutput{}, err
	}

	practices, err := service.practices.FindByCourse(courseID)
	if err != nil {
		return core.ListPracticesOutput{}, err
	}

	views := make([]core.PracticeView, 0, len(practices))
	for _, practice := range practices {
		views = append(views, practiceView(practice))
	}

	return core.ListPracticesOutput{Practices: views}, nil
}

func (service *PracticeService) GetPractice(in core.GetPracticeInput) (core.GetPracticeOutput, error) {
	id, err := domain.NewPracticeID(in.ID)
	if err != nil {
		return core.GetPracticeOutput{}, err
	}

	practice, err := service.practices.FindByID(id)
	if err != nil {
		return core.GetPracticeOutput{}, err
	}

	return core.GetPracticeOutput{Practice: practiceDetailView(practice)}, nil
}

func (service *PracticeService) UpdatePractice(in core.UpdatePracticeInput) (core.UpdatePracticeOutput, error) {
	id, err := domain.NewPracticeID(in.ID)
	if err != nil {
		return core.UpdatePracticeOutput{}, err
	}

	if in.Title == nil && in.Prompt == nil && in.StarterCode == nil && in.Solution == nil {
		return core.UpdatePracticeOutput{}, domain.NewValidationError("update", "must include at least one field")
	}

	practice, err := service.practices.FindByID(id)
	if err != nil {
		return core.UpdatePracticeOutput{}, err
	}

	now := service.clock.Now()
	if in.Title != nil {
		if err := practice.Rename(*in.Title, now); err != nil {
			return core.UpdatePracticeOutput{}, err
		}
	}
	if in.Prompt != nil {
		if err := practice.ChangePrompt(*in.Prompt, now); err != nil {
			return core.UpdatePracticeOutput{}, err
		}
	}
	if in.StarterCode != nil {
		practice.ChangeStarterCode(*in.StarterCode, now)
	}
	if in.Solution != nil {
		practice.ChangeSolution(*in.Solution, now)
	}

	if err := service.practices.Save(practice); err != nil {
		return core.UpdatePracticeOutput{}, err
	}

	return core.UpdatePracticeOutput{ID: id.String()}, nil
}

func (service *PracticeService) DeletePractice(in core.DeletePracticeInput) error {
	id, err := domain.NewPracticeID(in.ID)
	if err != nil {
		return err
	}

	if _, err := service.practices.FindByID(id); err != nil {
		return err
	}

	lessonIDs, err := service.lessons.FindLessonsEmbeddingPractice(id)
	if err != nil {
		return err
	}
	if len(lessonIDs) > 0 {
		return domain.NewPracticeInUseError(lessonIDs)
	}

	return service.practices.Delete(id)
}

func (service *PracticeService) AddTestCase(in core.AddTestCaseInput) (core.AddTestCaseOutput, error) {
	practiceID, err := domain.NewPracticeID(in.PracticeID)
	if err != nil {
		return core.AddTestCaseOutput{}, err
	}

	practice, err := service.practices.FindByID(practiceID)
	if err != nil {
		return core.AddTestCaseOutput{}, err
	}

	position, err := service.testCasePosition(practice, in.Position)
	if err != nil {
		return core.AddTestCaseOutput{}, err
	}

	id := service.ids.NewTestCaseID()
	testCase, err := domain.NewTestCase(id, in.Stdin, in.ExpectedStdout, in.Name, position)
	if err != nil {
		return core.AddTestCaseOutput{}, err
	}

	if err := practice.AddTestCase(testCase, service.clock.Now()); err != nil {
		return core.AddTestCaseOutput{}, err
	}

	if err := service.practices.Save(practice); err != nil {
		return core.AddTestCaseOutput{}, err
	}

	return core.AddTestCaseOutput{ID: id.String()}, nil
}

func (service *PracticeService) ListTestCases(in core.ListTestCasesInput) (core.ListTestCasesOutput, error) {
	practiceID, err := domain.NewPracticeID(in.PracticeID)
	if err != nil {
		return core.ListTestCasesOutput{}, err
	}

	practice, err := service.practices.FindByID(practiceID)
	if err != nil {
		return core.ListTestCasesOutput{}, err
	}

	return core.ListTestCasesOutput{TestCases: testCaseViews(practice)}, nil
}

func (service *PracticeService) GetTestCase(in core.GetTestCaseInput) (core.GetTestCaseOutput, error) {
	id, err := domain.NewTestCaseID(in.ID)
	if err != nil {
		return core.GetTestCaseOutput{}, err
	}

	practice, err := service.practices.FindByTestCaseID(id)
	if err != nil {
		return core.GetTestCaseOutput{}, err
	}

	testCase, err := practice.TestCase(id)
	if err != nil {
		return core.GetTestCaseOutput{}, err
	}

	return core.GetTestCaseOutput{TestCase: testCaseView(practice.ID(), testCase)}, nil
}

func (service *PracticeService) UpdateTestCase(in core.UpdateTestCaseInput) (core.UpdateTestCaseOutput, error) {
	id, err := domain.NewTestCaseID(in.ID)
	if err != nil {
		return core.UpdateTestCaseOutput{}, err
	}

	if in.Stdin == nil && in.ExpectedStdout == nil && in.Name == nil {
		return core.UpdateTestCaseOutput{}, domain.NewValidationError("update", "must include at least one field")
	}

	practice, err := service.practices.FindByTestCaseID(id)
	if err != nil {
		return core.UpdateTestCaseOutput{}, err
	}

	now := service.clock.Now()
	if in.Stdin != nil {
		if err := practice.ChangeTestCaseStdin(id, *in.Stdin, now); err != nil {
			return core.UpdateTestCaseOutput{}, err
		}
	}
	if in.ExpectedStdout != nil {
		if err := practice.ChangeTestCaseExpectedStdout(id, *in.ExpectedStdout, now); err != nil {
			return core.UpdateTestCaseOutput{}, err
		}
	}
	if in.Name != nil {
		if err := practice.ChangeTestCaseName(id, *in.Name, now); err != nil {
			return core.UpdateTestCaseOutput{}, err
		}
	}

	if err := service.practices.Save(practice); err != nil {
		return core.UpdateTestCaseOutput{}, err
	}

	return core.UpdateTestCaseOutput{ID: id.String()}, nil
}

func (service *PracticeService) RemoveTestCase(in core.RemoveTestCaseInput) error {
	id, err := domain.NewTestCaseID(in.ID)
	if err != nil {
		return err
	}

	practice, err := service.practices.FindByTestCaseID(id)
	if err != nil {
		return err
	}

	if err := practice.RemoveTestCase(id, service.clock.Now()); err != nil {
		return err
	}

	return service.practices.Save(practice)
}

func (service *PracticeService) ReorderTestCases(in core.ReorderTestCasesInput) error {
	practiceID, err := domain.NewPracticeID(in.PracticeID)
	if err != nil {
		return err
	}

	practice, err := service.practices.FindByID(practiceID)
	if err != nil {
		return err
	}

	placements, err := testCasePlacements(in.Order, practice)
	if err != nil {
		return err
	}

	if err := practice.ReorderTestCases(placements, service.clock.Now()); err != nil {
		return err
	}

	return service.practices.Save(practice)
}

func (service *PracticeService) testCasePosition(practice domain.Practice, explicitPosition *int) (domain.TestCasePosition, error) {
	if explicitPosition != nil {
		return domain.NewTestCasePosition(*explicitPosition)
	}

	return domain.NewTestCasePosition(nextTestCasePosition(practice.TestCases()))
}

func practiceView(practice domain.Practice) core.PracticeView {
	return core.PracticeView{
		ID:            practice.ID().String(),
		CourseID:      practice.CourseID().String(),
		Title:         practice.Title(),
		Language:      practice.Language().String(),
		TestCaseCount: len(practice.TestCases()),
		HasSolution:   strings.TrimSpace(practice.Solution()) != "",
		CreatedAt:     practice.CreatedAt(),
		UpdatedAt:     practice.UpdatedAt(),
	}
}

func practiceDetailView(practice domain.Practice) core.PracticeDetailView {
	return core.PracticeDetailView{
		PracticeView: practiceView(practice),
		Prompt:       practice.Prompt(),
		StarterCode:  practice.StarterCode(),
		Solution:     practice.Solution(),
		TestCases:    testCaseViews(practice),
	}
}

func testCaseViews(practice domain.Practice) []core.TestCaseView {
	testCases := practice.TestCases()
	views := make([]core.TestCaseView, 0, len(testCases))
	for _, testCase := range testCases {
		views = append(views, testCaseView(practice.ID(), testCase))
	}

	return views
}

func testCaseView(practiceID domain.PracticeID, testCase domain.TestCase) core.TestCaseView {
	return core.TestCaseView{
		ID:             testCase.ID().String(),
		PracticeID:     practiceID.String(),
		Stdin:          testCase.Stdin(),
		ExpectedStdout: testCase.ExpectedStdout(),
		Name:           testCase.Name(),
		Position:       testCase.Position().Int(),
	}
}

func nextTestCasePosition(testCases []domain.TestCase) int {
	maxPosition := -1
	for _, testCase := range testCases {
		if testCase.Position().Int() > maxPosition {
			maxPosition = testCase.Position().Int()
		}
	}

	return maxPosition + 1
}

func testCasePlacements(dtos []core.TestCasePlacementDTO, practice domain.Practice) ([]domain.TestCasePlacement, error) {
	current := make(map[string]struct{}, len(practice.TestCases()))
	for _, testCase := range practice.TestCases() {
		current[testCase.ID().String()] = struct{}{}
	}

	placements := make([]domain.TestCasePlacement, 0, len(dtos))
	for _, dto := range dtos {
		testCaseID, err := domain.NewTestCaseID(dto.TestCaseID)
		if err != nil {
			return nil, err
		}
		if _, exists := current[testCaseID.String()]; !exists {
			return nil, domain.NewValidationError("test_case_id", "must belong to the practice")
		}

		position, err := domain.NewTestCasePosition(dto.Position)
		if err != nil {
			return nil, err
		}

		placements = append(placements, domain.TestCasePlacement{
			TestCaseID: testCaseID,
			Position:   position,
		})
	}

	return placements, nil
}
