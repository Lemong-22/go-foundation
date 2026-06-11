package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/domain"
)

func TestDTOsCarryPrimitiveBoundaryData(t *testing.T) {
	title := "Intro to Go"
	description := "Learn Go"
	slug := "intro-to-go"
	order := 1
	passThreshold := 0.8

	courseInput := UpdateCourseInput{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Title:       &title,
		Description: &description,
		Slug:        &slug,
	}
	if *courseInput.Title != title || *courseInput.Description != description || *courseInput.Slug != slug {
		t.Fatalf("expected course DTO to carry primitive update values")
	}

	lessonInput := CreateLessonInput{
		CourseID: "550e8400-e29b-41d4-a716-446655440000",
		Title:    "First Lesson",
		Order:    &order,
	}
	if *lessonInput.Order != order {
		t.Fatalf("expected lesson DTO to carry primitive order")
	}

	blockInput := AddLessonBlockInput{
		LessonID:      "550e8400-e29b-41d4-a716-446655440020",
		Kind:          "video",
		VideoProvider: "youtube",
		VideoLocator:  "dQw4w9WgXcQ",
		QuizRef:       "550e8400-e29b-41d4-a716-446655440030",
		PracticeRef:   "550e8400-e29b-41d4-a716-446655440040",
	}
	if blockInput.Kind != "video" || blockInput.VideoProvider != "youtube" || blockInput.QuizRef == "" || blockInput.PracticeRef == "" {
		t.Fatalf("expected block DTO to carry primitive payload values")
	}

	quizInput := UpdateQuizInput{
		ID:            "550e8400-e29b-41d4-a716-446655440030",
		Title:         &title,
		PassThreshold: &passThreshold,
	}
	if *quizInput.Title != title || *quizInput.PassThreshold != passThreshold {
		t.Fatalf("expected quiz DTO to carry primitive update values")
	}

	options := []string{"A", "B"}
	correct := []int{0}
	questionInput := UpdateQuestionInput{
		ID:             "550e8400-e29b-41d4-a716-446655440031",
		Prompt:         &title,
		Options:        &options,
		CorrectIndices: &correct,
	}
	if (*questionInput.Options)[0] != "A" || (*questionInput.CorrectIndices)[0] != 0 {
		t.Fatalf("expected question DTO to carry primitive content slices")
	}

	practiceInput := UpdatePracticeInput{
		ID:          "550e8400-e29b-41d4-a716-446655440040",
		Title:       &title,
		Prompt:      &description,
		StarterCode: &slug,
		Solution:    &slug,
	}
	if *practiceInput.Title != title || *practiceInput.Prompt != description || *practiceInput.StarterCode != slug || *practiceInput.Solution != slug {
		t.Fatalf("expected practice DTO to carry primitive update values")
	}

	stdin := "input"
	expected := "output"
	name := "sample"
	testCaseInput := UpdateTestCaseInput{
		ID:             "550e8400-e29b-41d4-a716-446655440041",
		Stdin:          &stdin,
		ExpectedStdout: &expected,
		Name:           &name,
	}
	if *testCaseInput.Stdin != stdin || *testCaseInput.ExpectedStdout != expected || *testCaseInput.Name != name {
		t.Fatalf("expected test case DTO to carry primitive update values")
	}

	timeLimit := 45
	testInput := UpdateTestInput{
		ID:                    "550e8400-e29b-41d4-a716-446655440050",
		Title:                 &title,
		TimeLimitMinutes:      &timeLimit,
		PassThreshold:         &passThreshold,
		SolutionZipProvider:   &slug,
		SolutionZipLocator:    &description,
		SolutionVideoProvider: &slug,
		SolutionVideoLocator:  &description,
		SolutionVideoCaption:  &title,
	}
	if *testInput.TimeLimitMinutes != timeLimit || *testInput.PassThreshold != passThreshold || *testInput.SolutionVideoCaption != title {
		t.Fatalf("expected test DTO to carry primitive metadata and solution values")
	}

	testCases := []CodingTestCaseDTO{{Stdin: "input", ExpectedStdout: "output", Name: "sample"}}
	testItemInput := UpdateTestItemInput{
		ID:             "550e8400-e29b-41d4-a716-446655440051",
		Prompt:         &title,
		ChoiceType:     &slug,
		Options:        &options,
		CorrectIndices: &correct,
		Explanation:    &description,
		CodingPrompt:   &description,
		StarterCode:    &slug,
		Solution:       &title,
		TestCases:      &testCases,
	}
	if *testItemInput.Prompt != title || *testItemInput.CodingPrompt != description || (*testItemInput.TestCases)[0].ExpectedStdout != "output" {
		t.Fatalf("expected test item DTO to preserve choice and coding prompt fields")
	}
}

func TestViewsCarryFlatOutputData(t *testing.T) {
	now := time.Date(2026, 5, 24, 7, 0, 0, 0, time.UTC)

	course := CourseView{
		ID:           "course-id",
		Title:        "Intro to Go",
		Slug:         "intro-to-go",
		Description:  "Learn Go",
		InstructorID: "instructor-id",
		Status:       "draft",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if course.Status != "draft" || !course.CreatedAt.Equal(now) {
		t.Fatalf("expected flat course view data")
	}

	lesson := LessonView{
		ID:        "lesson-id",
		CourseID:  "course-id",
		Title:     "First Lesson",
		Order:     0,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if lesson.Order != 0 || !lesson.UpdatedAt.Equal(now) {
		t.Fatalf("expected flat lesson view data")
	}

	block := BlockView{
		ID:            "block-id",
		LessonID:      "lesson-id",
		Kind:          "video",
		Position:      0,
		VideoProvider: "youtube",
		VideoLocator:  "dQw4w9WgXcQ",
		VideoCaption:  "Intro",
		QuizRef:       "quiz-id",
		PracticeRef:   "practice-id",
	}
	if block.Kind != "video" || block.Position != 0 || block.VideoCaption != "Intro" || block.QuizRef != "quiz-id" || block.PracticeRef != "practice-id" {
		t.Fatalf("expected flat block view data")
	}

	quiz := QuizDetailView{
		QuizView: QuizView{
			ID:            "quiz-id",
			CourseID:      "course-id",
			Title:         "Basics Quiz",
			PassThreshold: 0.7,
			QuestionCount: 1,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		Questions: []QuestionView{
			{
				ID:             "question-id",
				QuizID:         "quiz-id",
				Type:           "single",
				Prompt:         "Pick one",
				Options:        []string{"A", "B"},
				CorrectIndices: []int{0},
				Position:       0,
			},
		},
	}
	if quiz.ID != "quiz-id" || quiz.Questions[0].CorrectIndices[0] != 0 {
		t.Fatalf("expected flat quiz and question view data")
	}

	practice := PracticeDetailView{
		PracticeView: PracticeView{
			ID:            "practice-id",
			CourseID:      "course-id",
			Title:         "FizzBuzz",
			Language:      "golang",
			TestCaseCount: 1,
			HasSolution:   true,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
		Prompt:      "Print fizz buzz",
		StarterCode: "package main",
		Solution:    "package main",
		TestCases: []TestCaseView{
			{
				ID:             "test-case-id",
				PracticeID:     "practice-id",
				Stdin:          "3",
				ExpectedStdout: "Fizz",
				Name:           "multiple of three",
				Position:       0,
			},
		},
	}
	if practice.ID != "practice-id" || practice.TestCases[0].ExpectedStdout != "Fizz" {
		t.Fatalf("expected flat practice and test case view data")
	}

	timeLimit := 30
	test := TestDetailView{
		TestView: TestView{
			ID:               "test-id",
			CourseID:         "course-id",
			Title:            "Final Test",
			TimeLimitMinutes: &timeLimit,
			PassThreshold:    0.8,
			HasSolution:      true,
			ItemCount:        2,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
		Solution: &TestSolutionView{
			ZipProvider:   "url",
			ZipLocator:    "https://example.com/solution.zip",
			VideoProvider: "youtube",
			VideoLocator:  "dQw4w9WgXcQ",
			VideoCaption:  "Walkthrough",
		},
		Items: []TestItemView{
			{
				ID:                   "choice-item-id",
				TestID:               "test-id",
				Kind:                 "choice",
				ChoicePrompt:         "Pick one",
				ChoiceType:           "single",
				ChoiceOptions:        []string{"A", "B"},
				ChoiceCorrectIndices: []int{0},
				Position:             0,
			},
			{
				ID:             "coding-item-id",
				TestID:         "test-id",
				Kind:           "coding",
				CodingPrompt:   "Write code",
				Language:       "golang",
				StarterCode:    "package main",
				CodingSolution: "package main",
				TestCases: []CodingTestCaseDTO{
					{Stdin: "3", ExpectedStdout: "Fizz", Name: "multiple of three"},
				},
				Position: 1,
			},
		},
	}
	if test.ID != "test-id" || test.Solution.VideoCaption != "Walkthrough" || test.Items[0].ChoicePrompt != "Pick one" || test.Items[1].CodingPrompt != "Write code" {
		t.Fatalf("expected flat test, solution, item, and coding test case view data")
	}
}

func TestPortInterfacesUseDomainTypesForCoreDependencies(t *testing.T) {
	var _ TestService = testServiceStub{}
	var _ CourseRepository = courseRepositoryStub{}
	var _ LessonRepository = lessonRepositoryStub{}
	var _ QuizRepository = quizRepositoryStub{}
	var _ PracticeRepository = practiceRepositoryStub{}
	var _ TestRepository = testRepositoryStub{}
	var _ IDGenerator = idGeneratorStub{}
	var _ Clock = clockStub{}
}

func TestLessonRepositoryDoesNotGainTestEmbeddingLookups(t *testing.T) {
	repository := reflect.TypeOf((*LessonRepository)(nil)).Elem()
	if _, exists := repository.MethodByName("FindLessonsEmbeddingTest"); exists {
		t.Fatalf("expected tests to stay course-level without lesson embedding lookups")
	}
}

type testServiceStub struct{}

func (testServiceStub) CreateTest(in CreateTestInput) (CreateTestOutput, error) {
	return CreateTestOutput{}, nil
}

func (testServiceStub) ListTests(in ListTestsInput) (ListTestsOutput, error) {
	return ListTestsOutput{}, nil
}

func (testServiceStub) GetTest(in GetTestInput) (GetTestOutput, error) {
	return GetTestOutput{}, nil
}

func (testServiceStub) UpdateTest(in UpdateTestInput) (UpdateTestOutput, error) {
	return UpdateTestOutput{}, nil
}

func (testServiceStub) DeleteTest(in DeleteTestInput) error {
	return nil
}

func (testServiceStub) AddTestItem(in AddTestItemInput) (AddTestItemOutput, error) {
	return AddTestItemOutput{}, nil
}

func (testServiceStub) ListTestItems(in ListTestItemsInput) (ListTestItemsOutput, error) {
	return ListTestItemsOutput{}, nil
}

func (testServiceStub) GetTestItem(in GetTestItemInput) (GetTestItemOutput, error) {
	return GetTestItemOutput{}, nil
}

func (testServiceStub) UpdateTestItem(in UpdateTestItemInput) (UpdateTestItemOutput, error) {
	return UpdateTestItemOutput{}, nil
}

func (testServiceStub) RemoveTestItem(in RemoveTestItemInput) error {
	return nil
}

func (testServiceStub) ReorderTestItems(in ReorderTestItemsInput) error {
	return nil
}

type courseRepositoryStub struct{}

func (courseRepositoryStub) Save(course domain.Course) error {
	return nil
}

func (courseRepositoryStub) FindByID(id domain.CourseID) (domain.Course, error) {
	return domain.Course{}, nil
}

func (courseRepositoryStub) FindBySlug(slug domain.Slug) (domain.Course, error) {
	return domain.Course{}, nil
}

func (courseRepositoryStub) FindAll(filter CourseFilter) ([]domain.Course, error) {
	return nil, nil
}

func (courseRepositoryStub) Delete(id domain.CourseID) error {
	return nil
}

type lessonRepositoryStub struct{}

func (lessonRepositoryStub) Save(lesson domain.Lesson) error {
	return nil
}

func (lessonRepositoryStub) SaveAll(lessons []domain.Lesson) error {
	return nil
}

func (lessonRepositoryStub) FindByID(id domain.LessonID) (domain.Lesson, error) {
	return domain.Lesson{}, nil
}

func (lessonRepositoryStub) FindByCourse(courseID domain.CourseID) ([]domain.Lesson, error) {
	return nil, nil
}

func (lessonRepositoryStub) FindByBlockID(id domain.BlockID) (domain.Lesson, error) {
	return domain.Lesson{}, nil
}

func (lessonRepositoryStub) FindLessonsEmbeddingQuiz(quizID domain.QuizID) ([]domain.LessonID, error) {
	return nil, nil
}

func (lessonRepositoryStub) FindLessonsEmbeddingPractice(practiceID domain.PracticeID) ([]domain.LessonID, error) {
	return nil, nil
}

func (lessonRepositoryStub) Delete(id domain.LessonID) error {
	return nil
}

func (lessonRepositoryStub) DeleteByCourse(courseID domain.CourseID) error {
	return nil
}

type quizRepositoryStub struct{}

func (quizRepositoryStub) Save(quiz domain.Quiz) error {
	return nil
}

func (quizRepositoryStub) FindByID(id domain.QuizID) (domain.Quiz, error) {
	return domain.Quiz{}, nil
}

func (quizRepositoryStub) FindByCourse(courseID domain.CourseID) ([]domain.Quiz, error) {
	return nil, nil
}

func (quizRepositoryStub) FindByQuestionID(id domain.QuestionID) (domain.Quiz, error) {
	return domain.Quiz{}, nil
}

func (quizRepositoryStub) Delete(id domain.QuizID) error {
	return nil
}

func (quizRepositoryStub) DeleteByCourse(courseID domain.CourseID) error {
	return nil
}

type practiceRepositoryStub struct{}

func (practiceRepositoryStub) Save(practice domain.Practice) error {
	return nil
}

func (practiceRepositoryStub) FindByID(id domain.PracticeID) (domain.Practice, error) {
	return domain.Practice{}, nil
}

func (practiceRepositoryStub) FindByCourse(courseID domain.CourseID) ([]domain.Practice, error) {
	return nil, nil
}

func (practiceRepositoryStub) FindByTestCaseID(id domain.TestCaseID) (domain.Practice, error) {
	return domain.Practice{}, nil
}

func (practiceRepositoryStub) Delete(id domain.PracticeID) error {
	return nil
}

func (practiceRepositoryStub) DeleteByCourse(courseID domain.CourseID) error {
	return nil
}

type testRepositoryStub struct{}

func (testRepositoryStub) Save(test domain.Test) error {
	return nil
}

func (testRepositoryStub) FindByID(id domain.TestID) (domain.Test, error) {
	return domain.Test{}, nil
}

func (testRepositoryStub) FindByCourse(courseID domain.CourseID) ([]domain.Test, error) {
	return nil, nil
}

func (testRepositoryStub) FindByItemID(id domain.TestItemID) (domain.Test, error) {
	return domain.Test{}, nil
}

func (testRepositoryStub) Delete(id domain.TestID) error {
	return nil
}

func (testRepositoryStub) DeleteByCourse(courseID domain.CourseID) error {
	return nil
}

type idGeneratorStub struct{}

func (idGeneratorStub) NewCourseID() domain.CourseID {
	return domain.CourseID{}
}

func (idGeneratorStub) NewLessonID() domain.LessonID {
	return domain.LessonID{}
}

func (idGeneratorStub) NewBlockID() domain.BlockID {
	return domain.BlockID{}
}

func (idGeneratorStub) NewQuizID() domain.QuizID {
	return domain.QuizID{}
}

func (idGeneratorStub) NewQuestionID() domain.QuestionID {
	return domain.QuestionID{}
}

func (idGeneratorStub) NewPracticeID() domain.PracticeID {
	return domain.PracticeID{}
}

func (idGeneratorStub) NewTestCaseID() domain.TestCaseID {
	return domain.TestCaseID{}
}

func (idGeneratorStub) NewTestID() domain.TestID {
	return domain.TestID{}
}

func (idGeneratorStub) NewTestItemID() domain.TestItemID {
	return domain.TestItemID{}
}

type clockStub struct{}

func (clockStub) Now() time.Time {
	return time.Time{}
}
