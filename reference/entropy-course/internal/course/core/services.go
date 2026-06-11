package core

type CourseService interface {
	CreateCourse(in CreateCourseInput) (CreateCourseOutput, error)
	ListCourses(in ListCoursesInput) (ListCoursesOutput, error)
	GetCourse(in GetCourseInput) (GetCourseOutput, error)
	UpdateCourse(in UpdateCourseInput) (UpdateCourseOutput, error)
	DeleteCourse(in DeleteCourseInput) error
	PublishCourse(in PublishCourseInput) error
	UnpublishCourse(in UnpublishCourseInput) error
}

type LessonService interface {
	CreateLesson(in CreateLessonInput) (CreateLessonOutput, error)
	ListLessons(in ListLessonsInput) (ListLessonsOutput, error)
	GetLesson(in GetLessonInput) (GetLessonOutput, error)
	UpdateLesson(in UpdateLessonInput) (UpdateLessonOutput, error)
	DeleteLesson(in DeleteLessonInput) error
	ReorderLessons(in ReorderLessonsInput) error
	AddLessonBlock(in AddLessonBlockInput) (AddLessonBlockOutput, error)
	ListLessonBlocks(in ListLessonBlocksInput) (ListLessonBlocksOutput, error)
	GetLessonBlock(in GetLessonBlockInput) (GetLessonBlockOutput, error)
	UpdateLessonBlock(in UpdateLessonBlockInput) (UpdateLessonBlockOutput, error)
	RemoveLessonBlock(in RemoveLessonBlockInput) error
	ReorderLessonBlocks(in ReorderLessonBlocksInput) error
}

type QuizService interface {
	CreateQuiz(in CreateQuizInput) (CreateQuizOutput, error)
	ListQuizzes(in ListQuizzesInput) (ListQuizzesOutput, error)
	GetQuiz(in GetQuizInput) (GetQuizOutput, error)
	UpdateQuiz(in UpdateQuizInput) (UpdateQuizOutput, error)
	DeleteQuiz(in DeleteQuizInput) error
	AddQuestion(in AddQuestionInput) (AddQuestionOutput, error)
	ListQuestions(in ListQuestionsInput) (ListQuestionsOutput, error)
	GetQuestion(in GetQuestionInput) (GetQuestionOutput, error)
	UpdateQuestion(in UpdateQuestionInput) (UpdateQuestionOutput, error)
	RemoveQuestion(in RemoveQuestionInput) error
	ReorderQuestions(in ReorderQuestionsInput) error
}

type PracticeService interface {
	CreatePractice(in CreatePracticeInput) (CreatePracticeOutput, error)
	ListPractices(in ListPracticesInput) (ListPracticesOutput, error)
	GetPractice(in GetPracticeInput) (GetPracticeOutput, error)
	UpdatePractice(in UpdatePracticeInput) (UpdatePracticeOutput, error)
	DeletePractice(in DeletePracticeInput) error
	AddTestCase(in AddTestCaseInput) (AddTestCaseOutput, error)
	ListTestCases(in ListTestCasesInput) (ListTestCasesOutput, error)
	GetTestCase(in GetTestCaseInput) (GetTestCaseOutput, error)
	UpdateTestCase(in UpdateTestCaseInput) (UpdateTestCaseOutput, error)
	RemoveTestCase(in RemoveTestCaseInput) error
	ReorderTestCases(in ReorderTestCasesInput) error
}

type TestService interface {
	CreateTest(in CreateTestInput) (CreateTestOutput, error)
	ListTests(in ListTestsInput) (ListTestsOutput, error)
	GetTest(in GetTestInput) (GetTestOutput, error)
	UpdateTest(in UpdateTestInput) (UpdateTestOutput, error)
	DeleteTest(in DeleteTestInput) error
	AddTestItem(in AddTestItemInput) (AddTestItemOutput, error)
	ListTestItems(in ListTestItemsInput) (ListTestItemsOutput, error)
	GetTestItem(in GetTestItemInput) (GetTestItemOutput, error)
	UpdateTestItem(in UpdateTestItemInput) (UpdateTestItemOutput, error)
	RemoveTestItem(in RemoveTestItemInput) error
	ReorderTestItems(in ReorderTestItemsInput) error
}

type ImportService interface {
	PlanImport(in PlanImportInput) (PlanImportOutput, error)
	ApplyPlan(in ApplyPlanInput) (ApplyPlanOutput, error)
}
