package usecase

import (
	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

type TestService struct {
	courses core.CourseRepository
	tests   core.TestRepository
	ids     core.IDGenerator
	clock   core.Clock
}

func NewTestService(
	courses core.CourseRepository,
	tests core.TestRepository,
	ids core.IDGenerator,
	clock core.Clock,
) *TestService {
	return &TestService{
		courses: courses,
		tests:   tests,
		ids:     ids,
		clock:   clock,
	}
}

func (service *TestService) CreateTest(in core.CreateTestInput) (core.CreateTestOutput, error) {
	courseID, err := domain.NewCourseID(in.CourseID)
	if err != nil {
		return core.CreateTestOutput{}, err
	}

	if _, err := service.courses.FindByID(courseID); err != nil {
		return core.CreateTestOutput{}, err
	}

	timeLimit, err := timeLimitFromInput(in.TimeLimitMinutes)
	if err != nil {
		return core.CreateTestOutput{}, err
	}

	threshold := domain.DefaultPassThreshold()
	if in.PassThreshold != nil {
		threshold, err = domain.NewPassThreshold(*in.PassThreshold)
		if err != nil {
			return core.CreateTestOutput{}, err
		}
	}

	id := service.ids.NewTestID()
	test, err := domain.NewTest(id, courseID, in.Title, timeLimit, threshold, nil, nil, service.clock.Now())
	if err != nil {
		return core.CreateTestOutput{}, err
	}

	if err := service.tests.Save(test); err != nil {
		return core.CreateTestOutput{}, err
	}

	return core.CreateTestOutput{ID: id.String()}, nil
}

func (service *TestService) ListTests(in core.ListTestsInput) (core.ListTestsOutput, error) {
	courseID, err := domain.NewCourseID(in.CourseID)
	if err != nil {
		return core.ListTestsOutput{}, err
	}

	if _, err := service.courses.FindByID(courseID); err != nil {
		return core.ListTestsOutput{}, err
	}

	tests, err := service.tests.FindByCourse(courseID)
	if err != nil {
		return core.ListTestsOutput{}, err
	}

	views := make([]core.TestView, 0, len(tests))
	for _, test := range tests {
		views = append(views, testView(test))
	}

	return core.ListTestsOutput{Tests: views}, nil
}

func (service *TestService) GetTest(in core.GetTestInput) (core.GetTestOutput, error) {
	id, err := domain.NewTestID(in.ID)
	if err != nil {
		return core.GetTestOutput{}, err
	}

	test, err := service.tests.FindByID(id)
	if err != nil {
		return core.GetTestOutput{}, err
	}

	return core.GetTestOutput{Test: testDetailView(test)}, nil
}

func (service *TestService) UpdateTest(in core.UpdateTestInput) (core.UpdateTestOutput, error) {
	id, err := domain.NewTestID(in.ID)
	if err != nil {
		return core.UpdateTestOutput{}, err
	}

	if isEmptyTestUpdate(in) {
		return core.UpdateTestOutput{}, domain.NewValidationError("update", "must include at least one field")
	}

	solution, setSolution, err := solutionFromUpdate(in)
	if err != nil {
		return core.UpdateTestOutput{}, err
	}

	test, err := service.tests.FindByID(id)
	if err != nil {
		return core.UpdateTestOutput{}, err
	}

	now := service.clock.Now()
	if in.Title != nil {
		if err := test.Rename(*in.Title, now); err != nil {
			return core.UpdateTestOutput{}, err
		}
	}
	if in.TimeLimitMinutes != nil {
		if *in.TimeLimitMinutes == 0 {
			test.ChangeTimeLimit(nil, now)
		} else {
			timeLimit, err := domain.NewTimeLimit(*in.TimeLimitMinutes)
			if err != nil {
				return core.UpdateTestOutput{}, err
			}
			test.ChangeTimeLimit(&timeLimit, now)
		}
	}
	if in.PassThreshold != nil {
		threshold, err := domain.NewPassThreshold(*in.PassThreshold)
		if err != nil {
			return core.UpdateTestOutput{}, err
		}
		test.ChangePassThreshold(threshold, now)
	}
	if setSolution {
		test.SetSolution(solution, now)
	}

	if err := service.tests.Save(test); err != nil {
		return core.UpdateTestOutput{}, err
	}

	return core.UpdateTestOutput{ID: id.String()}, nil
}

func (service *TestService) DeleteTest(in core.DeleteTestInput) error {
	id, err := domain.NewTestID(in.ID)
	if err != nil {
		return err
	}

	if _, err := service.tests.FindByID(id); err != nil {
		return err
	}

	return service.tests.Delete(id)
}

func (service *TestService) AddTestItem(in core.AddTestItemInput) (core.AddTestItemOutput, error) {
	testID, err := domain.NewTestID(in.TestID)
	if err != nil {
		return core.AddTestItemOutput{}, err
	}

	test, err := service.tests.FindByID(testID)
	if err != nil {
		return core.AddTestItemOutput{}, err
	}

	kind, err := domain.NewTestItemKind(in.Kind)
	if err != nil {
		return core.AddTestItemOutput{}, err
	}

	body, err := testItemBodyFromAdd(kind, in)
	if err != nil {
		return core.AddTestItemOutput{}, err
	}

	position, err := service.testItemPosition(test, in.Position)
	if err != nil {
		return core.AddTestItemOutput{}, err
	}

	id := service.ids.NewTestItemID()
	item, err := domain.NewTestItem(id, kind, body, position)
	if err != nil {
		return core.AddTestItemOutput{}, err
	}

	if err := test.AddItem(item, service.clock.Now()); err != nil {
		return core.AddTestItemOutput{}, err
	}

	if err := service.tests.Save(test); err != nil {
		return core.AddTestItemOutput{}, err
	}

	return core.AddTestItemOutput{ID: id.String()}, nil
}

func (service *TestService) ListTestItems(in core.ListTestItemsInput) (core.ListTestItemsOutput, error) {
	testID, err := domain.NewTestID(in.TestID)
	if err != nil {
		return core.ListTestItemsOutput{}, err
	}

	test, err := service.tests.FindByID(testID)
	if err != nil {
		return core.ListTestItemsOutput{}, err
	}

	return core.ListTestItemsOutput{Items: testItemViews(test)}, nil
}

func (service *TestService) GetTestItem(in core.GetTestItemInput) (core.GetTestItemOutput, error) {
	id, err := domain.NewTestItemID(in.ID)
	if err != nil {
		return core.GetTestItemOutput{}, err
	}

	test, err := service.tests.FindByItemID(id)
	if err != nil {
		return core.GetTestItemOutput{}, err
	}

	item, err := test.Item(id)
	if err != nil {
		return core.GetTestItemOutput{}, err
	}

	return core.GetTestItemOutput{Item: testItemView(test.ID(), item)}, nil
}

func (service *TestService) UpdateTestItem(in core.UpdateTestItemInput) (core.UpdateTestItemOutput, error) {
	id, err := domain.NewTestItemID(in.ID)
	if err != nil {
		return core.UpdateTestItemOutput{}, err
	}

	if isEmptyTestItemUpdate(in) {
		return core.UpdateTestItemOutput{}, domain.NewValidationError("update", "must include at least one field")
	}

	test, err := service.tests.FindByItemID(id)
	if err != nil {
		return core.UpdateTestItemOutput{}, err
	}

	item, err := test.Item(id)
	if err != nil {
		return core.UpdateTestItemOutput{}, err
	}

	body, err := testItemBodyFromUpdate(item, in)
	if err != nil {
		return core.UpdateTestItemOutput{}, err
	}

	if err := test.ReplaceItemBody(id, body, service.clock.Now()); err != nil {
		return core.UpdateTestItemOutput{}, err
	}

	if err := service.tests.Save(test); err != nil {
		return core.UpdateTestItemOutput{}, err
	}

	return core.UpdateTestItemOutput{ID: id.String()}, nil
}

func (service *TestService) RemoveTestItem(in core.RemoveTestItemInput) error {
	id, err := domain.NewTestItemID(in.ID)
	if err != nil {
		return err
	}

	test, err := service.tests.FindByItemID(id)
	if err != nil {
		return err
	}

	if err := test.RemoveItem(id, service.clock.Now()); err != nil {
		return err
	}

	return service.tests.Save(test)
}

func (service *TestService) ReorderTestItems(in core.ReorderTestItemsInput) error {
	testID, err := domain.NewTestID(in.TestID)
	if err != nil {
		return err
	}

	test, err := service.tests.FindByID(testID)
	if err != nil {
		return err
	}

	placements, err := testItemPlacements(in.Order, test)
	if err != nil {
		return err
	}

	if err := test.ReorderItems(placements, service.clock.Now()); err != nil {
		return err
	}

	return service.tests.Save(test)
}

func (service *TestService) testItemPosition(test domain.Test, explicitPosition *int) (domain.TestItemPosition, error) {
	if explicitPosition != nil {
		return domain.NewTestItemPosition(*explicitPosition)
	}

	return domain.NewTestItemPosition(nextTestItemPosition(test.Items()))
}

func timeLimitFromInput(minutes *int) (*domain.TimeLimit, error) {
	if minutes == nil {
		return nil, nil
	}

	timeLimit, err := domain.NewTimeLimit(*minutes)
	if err != nil {
		return nil, err
	}

	return &timeLimit, nil
}

func isEmptyTestUpdate(in core.UpdateTestInput) bool {
	return in.Title == nil &&
		in.TimeLimitMinutes == nil &&
		in.PassThreshold == nil &&
		!hasAnySolutionField(in)
}

func hasAnySolutionField(in core.UpdateTestInput) bool {
	return in.SolutionZipProvider != nil ||
		in.SolutionZipLocator != nil ||
		in.SolutionVideoProvider != nil ||
		in.SolutionVideoLocator != nil ||
		in.SolutionVideoCaption != nil
}

func hasRequiredSolutionFields(in core.UpdateTestInput) bool {
	return in.SolutionZipProvider != nil &&
		in.SolutionZipLocator != nil &&
		in.SolutionVideoProvider != nil &&
		in.SolutionVideoLocator != nil
}

func solutionFromUpdate(in core.UpdateTestInput) (domain.TestSolution, bool, error) {
	if !hasAnySolutionField(in) {
		return domain.TestSolution{}, false, nil
	}
	if !hasRequiredSolutionFields(in) {
		return domain.TestSolution{}, false, domain.NewValidationError("solution", "must include zip provider, zip locator, video provider, and video locator together")
	}

	zipProvider, err := domain.NewMediaProvider(*in.SolutionZipProvider)
	if err != nil {
		return domain.TestSolution{}, false, err
	}
	zip, err := domain.NewMediaRef(zipProvider, *in.SolutionZipLocator)
	if err != nil {
		return domain.TestSolution{}, false, err
	}

	videoProvider, err := domain.NewMediaProvider(*in.SolutionVideoProvider)
	if err != nil {
		return domain.TestSolution{}, false, err
	}
	video, err := domain.NewMediaRef(videoProvider, *in.SolutionVideoLocator)
	if err != nil {
		return domain.TestSolution{}, false, err
	}

	caption := ""
	if in.SolutionVideoCaption != nil {
		caption = *in.SolutionVideoCaption
	}

	solution, err := domain.NewTestSolution(zip, video, caption)
	if err != nil {
		return domain.TestSolution{}, false, err
	}

	return solution, true, nil
}

func testView(test domain.Test) core.TestView {
	return core.TestView{
		ID:               test.ID().String(),
		CourseID:         test.CourseID().String(),
		Title:            test.Title(),
		TimeLimitMinutes: timeLimitMinutes(test.TimeLimit()),
		PassThreshold:    test.PassThreshold().Float64(),
		HasSolution:      test.Solution() != nil,
		ItemCount:        len(test.Items()),
		CreatedAt:        test.CreatedAt(),
		UpdatedAt:        test.UpdatedAt(),
	}
}

func testDetailView(test domain.Test) core.TestDetailView {
	return core.TestDetailView{
		TestView: testView(test),
		Solution: solutionView(test.Solution()),
		Items:    testItemViews(test),
	}
}

func solutionView(solution *domain.TestSolution) *core.TestSolutionView {
	if solution == nil {
		return nil
	}

	zip := solution.SolutionZip()
	video := solution.ExplanationVideo()

	return &core.TestSolutionView{
		ZipProvider:   zip.Provider().String(),
		ZipLocator:    zip.Locator(),
		VideoProvider: video.Provider().String(),
		VideoLocator:  video.Locator(),
		VideoCaption:  solution.ExplanationCaption(),
	}
}

func testItemViews(test domain.Test) []core.TestItemView {
	items := test.Items()
	views := make([]core.TestItemView, 0, len(items))
	for _, item := range items {
		views = append(views, testItemView(test.ID(), item))
	}

	return views
}

func testItemView(testID domain.TestID, item domain.TestItem) core.TestItemView {
	view := core.TestItemView{
		ID:       item.ID().String(),
		TestID:   testID.String(),
		Kind:     item.Kind().String(),
		Position: item.Position().Int(),
	}

	switch body := item.Body().(type) {
	case domain.ChoiceItemBody:
		view.ChoicePrompt = body.Prompt()
		view.ChoiceType = body.Type().String()
		view.ChoiceOptions = body.Options()
		view.ChoiceCorrectIndices = body.CorrectIndices()
		view.ChoiceExplanation = body.Explanation()
	case domain.CodingItemBody:
		view.CodingPrompt = body.Prompt()
		view.Language = body.Language().String()
		view.StarterCode = body.StarterCode()
		view.CodingSolution = body.Solution()
		view.TestCases = codingTestCaseDTOs(body.TestCases())
	}

	return view
}

func codingTestCaseDTOs(testCases []domain.CodingTestCase) []core.CodingTestCaseDTO {
	dtos := make([]core.CodingTestCaseDTO, 0, len(testCases))
	for _, testCase := range testCases {
		dtos = append(dtos, core.CodingTestCaseDTO{
			Stdin:          testCase.Stdin(),
			ExpectedStdout: testCase.ExpectedStdout(),
			Name:           testCase.Name(),
		})
	}

	return dtos
}

func timeLimitMinutes(timeLimit *domain.TimeLimit) *int {
	if timeLimit == nil {
		return nil
	}

	minutes := timeLimit.Minutes()
	return &minutes
}

func testItemBodyFromAdd(kind domain.TestItemKind, in core.AddTestItemInput) (domain.TestItemBody, error) {
	if kind.IsChoice() {
		questionType, err := domain.NewChoiceQuestionType(in.ChoiceType)
		if err != nil {
			return nil, err
		}

		return domain.NewChoiceItemBody(questionType, in.Prompt, in.Options, in.CorrectIndices, in.Explanation)
	}

	language, err := domain.NewLanguage(in.Language)
	if err != nil {
		return nil, err
	}

	testCases := codingTestCasesFromDTOs(in.TestCases)
	return domain.NewCodingItemBody(language, in.CodingPrompt, in.StarterCode, in.Solution, testCases)
}

func testItemBodyFromUpdate(item domain.TestItem, in core.UpdateTestItemInput) (domain.TestItemBody, error) {
	if item.Kind().IsChoice() {
		if hasCodingItemUpdateFields(in) {
			return nil, domain.NewValidationError("kind", "update fields must match item kind")
		}

		body, ok := item.Body().(domain.ChoiceItemBody)
		if !ok {
			return nil, domain.NewValidationError("body", "must match item kind")
		}

		return updatedChoiceItemBody(body, in)
	}

	if hasChoiceSpecificItemUpdateFields(in) {
		return nil, domain.NewValidationError("kind", "update fields must match item kind")
	}

	body, ok := item.Body().(domain.CodingItemBody)
	if !ok {
		return nil, domain.NewValidationError("body", "must match item kind")
	}

	return updatedCodingItemBody(body, in)
}

func updatedChoiceItemBody(body domain.ChoiceItemBody, in core.UpdateTestItemInput) (domain.ChoiceItemBody, error) {
	if in.Options != nil && in.CorrectIndices == nil {
		return domain.ChoiceItemBody{}, domain.NewValidationError("correct_indices", "must be provided when options are updated")
	}
	if in.Options == nil && in.CorrectIndices != nil {
		return domain.ChoiceItemBody{}, domain.NewValidationError("options", "must be provided when correct indices are updated")
	}

	questionType := body.Type()
	if in.ChoiceType != nil {
		updatedType, err := domain.NewChoiceQuestionType(*in.ChoiceType)
		if err != nil {
			return domain.ChoiceItemBody{}, err
		}
		questionType = updatedType
	}

	prompt := body.Prompt()
	if in.Prompt != nil {
		prompt = *in.Prompt
	}

	options := body.Options()
	correctIndices := body.CorrectIndices()
	if in.Options != nil {
		options = *in.Options
		correctIndices = *in.CorrectIndices
	}

	explanation := body.Explanation()
	if in.Explanation != nil {
		explanation = *in.Explanation
	}

	return domain.NewChoiceItemBody(questionType, prompt, options, correctIndices, explanation)
}

func updatedCodingItemBody(body domain.CodingItemBody, in core.UpdateTestItemInput) (domain.CodingItemBody, error) {
	language := body.Language()
	if in.Language != nil {
		updatedLanguage, err := domain.NewLanguage(*in.Language)
		if err != nil {
			return domain.CodingItemBody{}, err
		}
		language = updatedLanguage
	}

	prompt := body.Prompt()
	if in.CodingPrompt != nil {
		prompt = *in.CodingPrompt
	} else if in.Prompt != nil {
		prompt = *in.Prompt
	}

	starterCode := body.StarterCode()
	if in.StarterCode != nil {
		starterCode = *in.StarterCode
	}

	solution := body.Solution()
	if in.Solution != nil {
		solution = *in.Solution
	}

	testCases := body.TestCases()
	if in.TestCases != nil {
		testCases = codingTestCasesFromDTOs(*in.TestCases)
	}

	return domain.NewCodingItemBody(language, prompt, starterCode, solution, testCases)
}

func isEmptyTestItemUpdate(in core.UpdateTestItemInput) bool {
	return !hasChoiceItemUpdateFields(in) && !hasCodingItemUpdateFields(in)
}

func hasChoiceItemUpdateFields(in core.UpdateTestItemInput) bool {
	return in.Prompt != nil ||
		hasChoiceSpecificItemUpdateFields(in)
}

func hasChoiceSpecificItemUpdateFields(in core.UpdateTestItemInput) bool {
	return in.ChoiceType != nil ||
		in.Options != nil ||
		in.CorrectIndices != nil ||
		in.Explanation != nil
}

func hasCodingItemUpdateFields(in core.UpdateTestItemInput) bool {
	return in.CodingPrompt != nil ||
		in.Language != nil ||
		in.StarterCode != nil ||
		in.Solution != nil ||
		in.TestCases != nil
}

func codingTestCasesFromDTOs(dtos []core.CodingTestCaseDTO) []domain.CodingTestCase {
	testCases := make([]domain.CodingTestCase, 0, len(dtos))
	for _, dto := range dtos {
		testCases = append(testCases, domain.NewCodingTestCase(dto.Stdin, dto.ExpectedStdout, dto.Name))
	}

	return testCases
}

func nextTestItemPosition(items []domain.TestItem) int {
	maxPosition := -1
	for _, item := range items {
		if item.Position().Int() > maxPosition {
			maxPosition = item.Position().Int()
		}
	}

	return maxPosition + 1
}

func testItemPlacements(dtos []core.TestItemPlacementDTO, test domain.Test) ([]domain.TestItemPlacement, error) {
	current := make(map[string]struct{}, len(test.Items()))
	for _, item := range test.Items() {
		current[item.ID().String()] = struct{}{}
	}

	placements := make([]domain.TestItemPlacement, 0, len(dtos))
	for _, dto := range dtos {
		testItemID, err := domain.NewTestItemID(dto.TestItemID)
		if err != nil {
			return nil, err
		}
		if _, exists := current[testItemID.String()]; !exists {
			return nil, domain.NewValidationError("test_item_id", "must belong to the test")
		}

		position, err := domain.NewTestItemPosition(dto.Position)
		if err != nil {
			return nil, err
		}

		placements = append(placements, domain.TestItemPlacement{
			TestItemID: testItemID,
			Position:   position,
		})
	}

	return placements, nil
}
