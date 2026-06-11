package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	outputTable = "table"
	outputJSON  = "json"
	outputQuiet = "quiet"
)

var ErrUnsupportedOutputFormat = errors.New("unsupported output format")

type courseOutputRenderer struct {
	writer io.Writer
}

func newCourseOutputRenderer(writer io.Writer) CourseRenderer {
	return courseOutputRenderer{writer: defaultOutputWriter(writer)}
}

func (renderer courseOutputRenderer) RenderCreatedCourse(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer courseOutputRenderer) RenderCourseList(format string, courses []core.CourseView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		rows := make([][]string, 0, len(courses))
		for _, course := range courses {
			rows = append(rows, []string{
				course.ID,
				course.Title,
				course.Slug,
				course.Status,
				course.InstructorID,
				formatTime(course.UpdatedAt),
			})
		}

		return writeTable(renderer.writer, []string{"ID", "TITLE", "SLUG", "STATUS", "INSTRUCTOR_ID", "UPDATED_AT"}, rows)
	case outputJSON:
		return writeJSON(renderer.writer, courses)
	case outputQuiet:
		return writeIDs(renderer.writer, courseIDs(courses))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer courseOutputRenderer) RenderCourse(format string, course core.CourseView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		return writeTable(renderer.writer, []string{"FIELD", "VALUE"}, [][]string{
			{"ID", course.ID},
			{"TITLE", course.Title},
			{"SLUG", course.Slug},
			{"DESCRIPTION", course.Description},
			{"INSTRUCTOR_ID", course.InstructorID},
			{"STATUS", course.Status},
			{"CREATED_AT", formatTime(course.CreatedAt)},
			{"UPDATED_AT", formatTime(course.UpdatedAt)},
		})
	case outputJSON:
		return writeJSON(renderer.writer, course)
	case outputQuiet:
		return writeLine(renderer.writer, course.ID)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer courseOutputRenderer) RenderUpdatedCourse(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer courseOutputRenderer) RenderConfirmation(message string) error {
	return writeLine(renderer.writer, message)
}

type lessonOutputRenderer struct {
	writer io.Writer
}

func newLessonOutputRenderer(writer io.Writer) LessonRenderer {
	return lessonOutputRenderer{writer: defaultOutputWriter(writer)}
}

func (renderer lessonOutputRenderer) RenderCreatedLesson(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer lessonOutputRenderer) RenderLessonList(format string, lessons []core.LessonView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		rows := make([][]string, 0, len(lessons))
		for _, lesson := range lessons {
			rows = append(rows, []string{
				lesson.ID,
				lesson.CourseID,
				strconv.Itoa(lesson.Order),
				lesson.Title,
				formatTime(lesson.UpdatedAt),
			})
		}

		return writeTable(renderer.writer, []string{"ID", "COURSE_ID", "ORDER", "TITLE", "UPDATED_AT"}, rows)
	case outputJSON:
		return writeJSON(renderer.writer, lessons)
	case outputQuiet:
		return writeIDs(renderer.writer, lessonIDs(lessons))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer lessonOutputRenderer) RenderLesson(format string, lesson core.LessonView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		return writeTable(renderer.writer, []string{"FIELD", "VALUE"}, [][]string{
			{"ID", lesson.ID},
			{"COURSE_ID", lesson.CourseID},
			{"TITLE", lesson.Title},
			{"ORDER", strconv.Itoa(lesson.Order)},
			{"CREATED_AT", formatTime(lesson.CreatedAt)},
			{"UPDATED_AT", formatTime(lesson.UpdatedAt)},
		})
	case outputJSON:
		return writeJSON(renderer.writer, lesson)
	case outputQuiet:
		return writeLine(renderer.writer, lesson.ID)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer lessonOutputRenderer) RenderUpdatedLesson(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer lessonOutputRenderer) RenderCreatedBlock(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer lessonOutputRenderer) RenderBlockList(format string, blocks []core.BlockView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		rows := make([][]string, 0, len(blocks))
		for _, block := range blocks {
			rows = append(rows, blockRow(block))
		}

		return writeTable(renderer.writer, blockListHeaders(), rows)
	case outputJSON:
		return writeJSON(renderer.writer, blocks)
	case outputQuiet:
		return writeIDs(renderer.writer, blockIDs(blocks))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer lessonOutputRenderer) RenderBlock(format string, block core.BlockView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		return writeTable(renderer.writer, []string{"FIELD", "VALUE"}, [][]string{
			{"ID", block.ID},
			{"LESSON_ID", block.LessonID},
			{"KIND", block.Kind},
			{"POSITION", strconv.Itoa(block.Position)},
			{"MARKDOWN", block.Markdown},
			{"VIDEO_PROVIDER", block.VideoProvider},
			{"VIDEO_LOCATOR", block.VideoLocator},
			{"VIDEO_CAPTION", block.VideoCaption},
			{"QUIZ_REF", block.QuizRef},
			{"PRACTICE_REF", block.PracticeRef},
		})
	case outputJSON:
		return writeJSON(renderer.writer, block)
	case outputQuiet:
		return writeLine(renderer.writer, block.ID)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer lessonOutputRenderer) RenderUpdatedBlock(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer lessonOutputRenderer) RenderConfirmation(message string) error {
	return writeLine(renderer.writer, message)
}

type quizOutputRenderer struct {
	writer io.Writer
}

func newQuizOutputRenderer(writer io.Writer) QuizRenderer {
	return quizOutputRenderer{writer: defaultOutputWriter(writer)}
}

func (renderer quizOutputRenderer) RenderCreatedQuiz(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer quizOutputRenderer) RenderQuizList(format string, quizzes []core.QuizView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		rows := make([][]string, 0, len(quizzes))
		for _, quiz := range quizzes {
			rows = append(rows, quizRow(quiz))
		}

		return writeTable(renderer.writer, quizListHeaders(), rows)
	case outputJSON:
		return writeJSON(renderer.writer, quizzes)
	case outputQuiet:
		return writeIDs(renderer.writer, quizIDs(quizzes))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer quizOutputRenderer) RenderQuiz(format string, quiz core.QuizDetailView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		return writeTable(renderer.writer, []string{"FIELD", "VALUE"}, [][]string{
			{"ID", quiz.ID},
			{"COURSE_ID", quiz.CourseID},
			{"TITLE", quiz.Title},
			{"PASS_THRESHOLD", formatFloat(quiz.PassThreshold)},
			{"QUESTION_COUNT", strconv.Itoa(quiz.QuestionCount)},
			{"QUESTIONS", questionSummaries(quiz.Questions)},
			{"CREATED_AT", formatTime(quiz.CreatedAt)},
			{"UPDATED_AT", formatTime(quiz.UpdatedAt)},
		})
	case outputJSON:
		return writeJSON(renderer.writer, quiz)
	case outputQuiet:
		return writeLine(renderer.writer, quiz.ID)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer quizOutputRenderer) RenderUpdatedQuiz(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer quizOutputRenderer) RenderCreatedQuestion(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer quizOutputRenderer) RenderQuestionList(format string, questions []core.QuestionView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		rows := make([][]string, 0, len(questions))
		for _, question := range questions {
			rows = append(rows, questionRow(question))
		}

		return writeTable(renderer.writer, questionListHeaders(), rows)
	case outputJSON:
		return writeJSON(renderer.writer, questions)
	case outputQuiet:
		return writeIDs(renderer.writer, questionIDs(questions))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer quizOutputRenderer) RenderQuestion(format string, question core.QuestionView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		return writeTable(renderer.writer, []string{"FIELD", "VALUE"}, [][]string{
			{"ID", question.ID},
			{"QUIZ_ID", question.QuizID},
			{"POSITION", strconv.Itoa(question.Position)},
			{"TYPE", question.Type},
			{"PROMPT", question.Prompt},
			{"OPTIONS", strings.Join(question.Options, " | ")},
			{"CORRECT_INDICES", joinInts(question.CorrectIndices)},
			{"EXPLANATION", question.Explanation},
		})
	case outputJSON:
		return writeJSON(renderer.writer, question)
	case outputQuiet:
		return writeLine(renderer.writer, question.ID)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer quizOutputRenderer) RenderUpdatedQuestion(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer quizOutputRenderer) RenderConfirmation(message string) error {
	return writeLine(renderer.writer, message)
}

type practiceOutputRenderer struct {
	writer io.Writer
}

func newPracticeOutputRenderer(writer io.Writer) PracticeRenderer {
	return practiceOutputRenderer{writer: defaultOutputWriter(writer)}
}

func (renderer practiceOutputRenderer) RenderCreatedPractice(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer practiceOutputRenderer) RenderPracticeList(format string, practices []core.PracticeView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		rows := make([][]string, 0, len(practices))
		for _, practice := range practices {
			rows = append(rows, practiceRow(practice))
		}

		return writeTable(renderer.writer, practiceListHeaders(), rows)
	case outputJSON:
		return writeJSON(renderer.writer, practices)
	case outputQuiet:
		return writeIDs(renderer.writer, practiceIDs(practices))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer practiceOutputRenderer) RenderPractice(format string, practice core.PracticeDetailView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		return writeTable(renderer.writer, []string{"FIELD", "VALUE"}, [][]string{
			{"ID", practice.ID},
			{"COURSE_ID", practice.CourseID},
			{"TITLE", practice.Title},
			{"LANGUAGE", practice.Language},
			{"PROMPT", practice.Prompt},
			{"STARTER_CODE", practice.StarterCode},
			{"SOLUTION", practice.Solution},
			{"TEST_CASE_COUNT", strconv.Itoa(practice.TestCaseCount)},
			{"TEST_CASES", testCaseSummaries(practice.TestCases)},
			{"CREATED_AT", formatTime(practice.CreatedAt)},
			{"UPDATED_AT", formatTime(practice.UpdatedAt)},
		})
	case outputJSON:
		return writeJSON(renderer.writer, practice)
	case outputQuiet:
		return writeLine(renderer.writer, practice.ID)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer practiceOutputRenderer) RenderUpdatedPractice(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer practiceOutputRenderer) RenderCreatedTestCase(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer practiceOutputRenderer) RenderTestCaseList(format string, testCases []core.TestCaseView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		rows := make([][]string, 0, len(testCases))
		for _, testCase := range testCases {
			rows = append(rows, testCaseRow(testCase))
		}

		return writeTable(renderer.writer, testCaseListHeaders(), rows)
	case outputJSON:
		return writeJSON(renderer.writer, testCases)
	case outputQuiet:
		return writeIDs(renderer.writer, testCaseIDs(testCases))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer practiceOutputRenderer) RenderTestCase(format string, testCase core.TestCaseView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		return writeTable(renderer.writer, []string{"FIELD", "VALUE"}, [][]string{
			{"ID", testCase.ID},
			{"PRACTICE_ID", testCase.PracticeID},
			{"POSITION", strconv.Itoa(testCase.Position)},
			{"NAME", testCase.Name},
			{"STDIN", testCase.Stdin},
			{"EXPECTED_STDOUT", testCase.ExpectedStdout},
		})
	case outputJSON:
		return writeJSON(renderer.writer, testCase)
	case outputQuiet:
		return writeLine(renderer.writer, testCase.ID)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer practiceOutputRenderer) RenderUpdatedTestCase(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer practiceOutputRenderer) RenderConfirmation(message string) error {
	return writeLine(renderer.writer, message)
}

type testOutputRenderer struct {
	writer io.Writer
}

func newTestOutputRenderer(writer io.Writer) TestRenderer {
	return testOutputRenderer{writer: defaultOutputWriter(writer)}
}

func (renderer testOutputRenderer) RenderCreatedTest(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer testOutputRenderer) RenderTestList(format string, tests []core.TestView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		rows := make([][]string, 0, len(tests))
		for _, test := range tests {
			rows = append(rows, testRow(test))
		}

		return writeTable(renderer.writer, testListHeaders(), rows)
	case outputJSON:
		return writeJSON(renderer.writer, tests)
	case outputQuiet:
		return writeIDs(renderer.writer, testIDs(tests))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer testOutputRenderer) RenderTest(format string, test core.TestDetailView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		return writeTable(renderer.writer, []string{"FIELD", "VALUE"}, [][]string{
			{"ID", test.ID},
			{"COURSE_ID", test.CourseID},
			{"TITLE", test.Title},
			{"TIME_LIMIT_MINUTES", formatOptionalInt(test.TimeLimitMinutes)},
			{"PASS_THRESHOLD", formatFloat(test.PassThreshold)},
			{"HAS_SOLUTION", strconv.FormatBool(test.HasSolution)},
			{"SOLUTION", testSolutionSummary(test.Solution)},
			{"ITEM_COUNT", strconv.Itoa(test.ItemCount)},
			{"ITEMS", testItemSummaries(test.Items)},
			{"CREATED_AT", formatTime(test.CreatedAt)},
			{"UPDATED_AT", formatTime(test.UpdatedAt)},
		})
	case outputJSON:
		return writeJSON(renderer.writer, test)
	case outputQuiet:
		return writeLine(renderer.writer, test.ID)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer testOutputRenderer) RenderUpdatedTest(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer testOutputRenderer) RenderCreatedTestItem(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer testOutputRenderer) RenderTestItemList(format string, items []core.TestItemView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		rows := make([][]string, 0, len(items))
		for _, item := range items {
			rows = append(rows, testItemRow(item))
		}

		return writeTable(renderer.writer, testItemListHeaders(), rows)
	case outputJSON:
		return writeJSON(renderer.writer, items)
	case outputQuiet:
		return writeIDs(renderer.writer, testItemIDs(items))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer testOutputRenderer) RenderTestItem(format string, item core.TestItemView) error {
	switch normalizeOutputFormat(format) {
	case outputTable:
		return writeTable(renderer.writer, []string{"FIELD", "VALUE"}, [][]string{
			{"ID", item.ID},
			{"TEST_ID", item.TestID},
			{"KIND", item.Kind},
			{"POSITION", strconv.Itoa(item.Position)},
			{"CHOICE_PROMPT", item.ChoicePrompt},
			{"CHOICE_TYPE", item.ChoiceType},
			{"CHOICE_OPTIONS", strings.Join(item.ChoiceOptions, " | ")},
			{"CHOICE_CORRECT_INDICES", joinInts(item.ChoiceCorrectIndices)},
			{"CHOICE_EXPLANATION", item.ChoiceExplanation},
			{"CODING_PROMPT", item.CodingPrompt},
			{"LANGUAGE", item.Language},
			{"STARTER_CODE", item.StarterCode},
			{"CODING_SOLUTION", item.CodingSolution},
			{"TEST_CASES", codingTestCaseSummaries(item.TestCases)},
		})
	case outputJSON:
		return writeJSON(renderer.writer, item)
	case outputQuiet:
		return writeLine(renderer.writer, item.ID)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer testOutputRenderer) RenderUpdatedTestItem(id string) error {
	return writeLine(renderer.writer, id)
}

func (renderer testOutputRenderer) RenderConfirmation(message string) error {
	return writeLine(renderer.writer, message)
}

type importOutputRenderer struct {
	writer io.Writer
}

func newImportOutputRenderer(writer io.Writer) ImportRenderer {
	return importOutputRenderer{writer: defaultOutputWriter(writer)}
}

func (renderer importOutputRenderer) RenderImportPlan(format string, plan domain.ImportPlan) error {
	switch normalizeOutputFormat(format) {
	case outputJSON:
		return writeJSON(renderer.writer, plan)
	case outputTable:
		if err := writeLine(renderer.writer, "OPERATIONS"); err != nil {
			return err
		}
		if err := writeTable(renderer.writer, importOperationHeaders(), importOperationRows(plan.Operations())); err != nil {
			return err
		}
		if err := writeLine(renderer.writer, "CONFLICTS"); err != nil {
			return err
		}
		return writeTable(renderer.writer, importConflictHeaders(), importConflictRows(plan.Conflicts()))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func (renderer importOutputRenderer) RenderApplyResult(format string, result domain.ApplyResult) error {
	switch normalizeOutputFormat(format) {
	case outputJSON:
		return writeJSON(renderer.writer, applyResultOutput(result))
	case outputTable:
		if err := writeTable(renderer.writer, []string{"FIELD", "VALUE"}, applyResultRows(result)); err != nil {
			return err
		}
		failed := result.Failed()
		if len(failed) == 0 {
			return nil
		}
		if err := writeLine(renderer.writer, "FAILED"); err != nil {
			return err
		}
		return writeTable(renderer.writer, []string{"ENTITY", "REF", "KIND", "ERROR"}, failedOperationRows(failed))
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedOutputFormat, format)
	}
}

func normalizeOutputFormat(format string) string {
	format = strings.TrimSpace(strings.ToLower(format))
	if format == "" {
		return outputTable
	}

	return format
}

func defaultOutputWriter(writer io.Writer) io.Writer {
	if writer == nil {
		return io.Discard
	}

	return writer
}

func writeJSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeTable(writer io.Writer, headers []string, rows [][]string) error {
	table := tabwriter.NewWriter(writer, 0, 0, 2, ' ', 0)

	if err := writeTableRow(table, headers); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writeTableRow(table, row); err != nil {
			return err
		}
	}

	return table.Flush()
}

func writeTableRow(writer io.Writer, columns []string) error {
	_, err := fmt.Fprintln(writer, strings.Join(columns, "\t"))
	return err
}

func writeIDs(writer io.Writer, ids []string) error {
	for _, id := range ids {
		if err := writeLine(writer, id); err != nil {
			return err
		}
	}

	return nil
}

func writeLine(writer io.Writer, value string) error {
	_, err := fmt.Fprintln(writer, value)
	return err
}

func courseIDs(courses []core.CourseView) []string {
	ids := make([]string, 0, len(courses))
	for _, course := range courses {
		ids = append(ids, course.ID)
	}

	return ids
}

func lessonIDs(lessons []core.LessonView) []string {
	ids := make([]string, 0, len(lessons))
	for _, lesson := range lessons {
		ids = append(ids, lesson.ID)
	}

	return ids
}

func blockListHeaders() []string {
	return []string{
		"ID",
		"LESSON_ID",
		"POSITION",
		"KIND",
		"MARKDOWN",
		"VIDEO_PROVIDER",
		"VIDEO_LOCATOR",
		"VIDEO_CAPTION",
		"QUIZ_REF",
		"PRACTICE_REF",
	}
}

func blockRow(block core.BlockView) []string {
	return []string{
		block.ID,
		block.LessonID,
		strconv.Itoa(block.Position),
		block.Kind,
		block.Markdown,
		block.VideoProvider,
		block.VideoLocator,
		block.VideoCaption,
		block.QuizRef,
		block.PracticeRef,
	}
}

func blockIDs(blocks []core.BlockView) []string {
	ids := make([]string, 0, len(blocks))
	for _, block := range blocks {
		ids = append(ids, block.ID)
	}

	return ids
}

func quizListHeaders() []string {
	return []string{
		"ID",
		"COURSE_ID",
		"TITLE",
		"PASS_THRESHOLD",
		"QUESTION_COUNT",
		"UPDATED_AT",
	}
}

func quizRow(quiz core.QuizView) []string {
	return []string{
		quiz.ID,
		quiz.CourseID,
		quiz.Title,
		formatFloat(quiz.PassThreshold),
		strconv.Itoa(quiz.QuestionCount),
		formatTime(quiz.UpdatedAt),
	}
}

func quizIDs(quizzes []core.QuizView) []string {
	ids := make([]string, 0, len(quizzes))
	for _, quiz := range quizzes {
		ids = append(ids, quiz.ID)
	}

	return ids
}

func questionListHeaders() []string {
	return []string{
		"ID",
		"QUIZ_ID",
		"POSITION",
		"TYPE",
		"PROMPT",
		"OPTIONS",
		"CORRECT_INDICES",
		"EXPLANATION",
	}
}

func questionRow(question core.QuestionView) []string {
	return []string{
		question.ID,
		question.QuizID,
		strconv.Itoa(question.Position),
		question.Type,
		question.Prompt,
		strings.Join(question.Options, " | "),
		joinInts(question.CorrectIndices),
		question.Explanation,
	}
}

func questionIDs(questions []core.QuestionView) []string {
	ids := make([]string, 0, len(questions))
	for _, question := range questions {
		ids = append(ids, question.ID)
	}

	return ids
}

func questionSummaries(questions []core.QuestionView) string {
	summaries := make([]string, 0, len(questions))
	for _, question := range questions {
		summaries = append(summaries, fmt.Sprintf("%d:%s", question.Position, question.ID))
	}

	return strings.Join(summaries, ", ")
}

func practiceListHeaders() []string {
	return []string{
		"ID",
		"COURSE_ID",
		"TITLE",
		"LANGUAGE",
		"TEST_CASE_COUNT",
		"HAS_SOLUTION",
		"UPDATED_AT",
	}
}

func practiceRow(practice core.PracticeView) []string {
	return []string{
		practice.ID,
		practice.CourseID,
		practice.Title,
		practice.Language,
		strconv.Itoa(practice.TestCaseCount),
		strconv.FormatBool(practice.HasSolution),
		formatTime(practice.UpdatedAt),
	}
}

func practiceIDs(practices []core.PracticeView) []string {
	ids := make([]string, 0, len(practices))
	for _, practice := range practices {
		ids = append(ids, practice.ID)
	}

	return ids
}

func testCaseListHeaders() []string {
	return []string{
		"ID",
		"PRACTICE_ID",
		"POSITION",
		"NAME",
		"STDIN",
		"EXPECTED_STDOUT",
	}
}

func testCaseRow(testCase core.TestCaseView) []string {
	return []string{
		testCase.ID,
		testCase.PracticeID,
		strconv.Itoa(testCase.Position),
		testCase.Name,
		testCase.Stdin,
		testCase.ExpectedStdout,
	}
}

func testCaseIDs(testCases []core.TestCaseView) []string {
	ids := make([]string, 0, len(testCases))
	for _, testCase := range testCases {
		ids = append(ids, testCase.ID)
	}

	return ids
}

func testCaseSummaries(testCases []core.TestCaseView) string {
	summaries := make([]string, 0, len(testCases))
	for _, testCase := range testCases {
		summaries = append(summaries, fmt.Sprintf("%d:%s", testCase.Position, testCase.ID))
	}

	return strings.Join(summaries, ", ")
}

func testListHeaders() []string {
	return []string{
		"ID",
		"COURSE_ID",
		"TITLE",
		"TIME_LIMIT_MINUTES",
		"PASS_THRESHOLD",
		"HAS_SOLUTION",
		"ITEM_COUNT",
		"UPDATED_AT",
	}
}

func testRow(test core.TestView) []string {
	return []string{
		test.ID,
		test.CourseID,
		test.Title,
		formatOptionalInt(test.TimeLimitMinutes),
		formatFloat(test.PassThreshold),
		strconv.FormatBool(test.HasSolution),
		strconv.Itoa(test.ItemCount),
		formatTime(test.UpdatedAt),
	}
}

func testIDs(tests []core.TestView) []string {
	ids := make([]string, 0, len(tests))
	for _, test := range tests {
		ids = append(ids, test.ID)
	}

	return ids
}

func testSolutionSummary(solution *core.TestSolutionView) string {
	if solution == nil {
		return ""
	}

	return fmt.Sprintf("%s:%s -> %s:%s", solution.ZipProvider, solution.ZipLocator, solution.VideoProvider, solution.VideoLocator)
}

func testItemListHeaders() []string {
	return []string{
		"ID",
		"TEST_ID",
		"POSITION",
		"KIND",
		"PROMPT",
		"TYPE",
		"LANGUAGE",
		"TEST_CASE_COUNT",
	}
}

func testItemRow(item core.TestItemView) []string {
	return []string{
		item.ID,
		item.TestID,
		strconv.Itoa(item.Position),
		item.Kind,
		testItemPrompt(item),
		item.ChoiceType,
		item.Language,
		strconv.Itoa(len(item.TestCases)),
	}
}

func testItemPrompt(item core.TestItemView) string {
	if item.CodingPrompt != "" {
		return item.CodingPrompt
	}

	return item.ChoicePrompt
}

func testItemIDs(items []core.TestItemView) []string {
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}

	return ids
}

func testItemSummaries(items []core.TestItemView) string {
	summaries := make([]string, 0, len(items))
	for _, item := range items {
		summaries = append(summaries, fmt.Sprintf("%d:%s", item.Position, item.ID))
	}

	return strings.Join(summaries, ", ")
}

func codingTestCaseSummaries(testCases []core.CodingTestCaseDTO) string {
	summaries := make([]string, 0, len(testCases))
	for index, testCase := range testCases {
		name := testCase.Name
		if name == "" {
			name = strconv.Itoa(index)
		}
		summaries = append(summaries, fmt.Sprintf("%d:%s", index, name))
	}

	return strings.Join(summaries, ", ")
}

func importOperationHeaders() []string {
	return []string{"KIND", "ENTITY", "REF", "TARGET_ID"}
}

func importOperationRows(operations []domain.ImportOperation) [][]string {
	rows := make([][]string, 0, len(operations))
	for _, operation := range operations {
		rows = append(rows, []string{
			operation.Kind().String(),
			operation.EntityType().String(),
			operation.EntityRef(),
			importTargetID(operation),
		})
	}

	return rows
}

func importConflictHeaders() []string {
	return []string{"ENTITY", "REF", "REASON", "RECOMMENDED", "CANDIDATES"}
}

func importConflictRows(conflicts []domain.ImportConflict) [][]string {
	rows := make([][]string, 0, len(conflicts))
	for _, conflict := range conflicts {
		rows = append(rows, []string{
			conflict.EntityType().String(),
			conflict.EntityRef(),
			conflict.Reason().String(),
			conflict.Recommended().String(),
			conflictCandidateSummary(conflict.Candidates()),
		})
	}

	return rows
}

func importTargetID(operation domain.ImportOperation) string {
	targetID := operation.TargetID()
	if targetID == nil {
		return ""
	}

	return *targetID
}

func conflictCandidateSummary(candidates []domain.ConflictCandidate) string {
	parts := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		parts = append(parts, candidate.ID()+":"+candidate.Description())
	}

	return strings.Join(parts, ", ")
}

func applyResultRows(result domain.ApplyResult) [][]string {
	counts := applyResultCounts(result)
	return [][]string{
		{"APPLIED", strconv.Itoa(len(result.Applied()))},
		{"FAILED", strconv.Itoa(len(result.Failed()))},
		{"SKIPPED", strconv.Itoa(len(result.Skipped()))},
		{"CREATES", strconv.Itoa(counts["create"])},
		{"UPDATES", strconv.Itoa(counts["update"])},
		{"NOOPS", strconv.Itoa(counts["noop"])},
		{"SKIPS", strconv.Itoa(counts["skip"])},
		{"AGGREGATES_SUCCEEDED", strconv.Itoa(result.AggregatesSucceeded())},
		{"AGGREGATES_FAILED", strconv.Itoa(result.AggregatesFailed())},
	}
}

func failedOperationRows(failed []domain.FailedOperation) [][]string {
	rows := make([][]string, 0, len(failed))
	for _, failure := range failed {
		operation := failure.Operation()
		rows = append(rows, []string{
			operation.EntityType().String(),
			operation.EntityRef(),
			operation.Kind().String(),
			failure.Err().Error(),
		})
	}

	return rows
}

func applyResultOutput(result domain.ApplyResult) map[string]any {
	return map[string]any{
		"applied":              appliedOperationOutputs(result.Applied()),
		"failed":               failedOperationOutputs(result.Failed()),
		"skipped":              importOperationOutputs(result.Skipped()),
		"counts":               applyResultCounts(result),
		"aggregates_succeeded": result.AggregatesSucceeded(),
		"aggregates_failed":    result.AggregatesFailed(),
	}
}

func applyResultCounts(result domain.ApplyResult) map[string]int {
	counts := map[string]int{
		"create": 0,
		"update": 0,
		"noop":   0,
		"skip":   0,
	}
	for _, applied := range result.Applied() {
		counts[applied.Operation().Kind().String()]++
	}
	for _, failed := range result.Failed() {
		counts[failed.Operation().Kind().String()]++
	}
	for _, skipped := range result.Skipped() {
		counts[skipped.Kind().String()]++
	}

	return counts
}

func appliedOperationOutputs(operations []domain.AppliedOperation) []map[string]any {
	outputs := make([]map[string]any, 0, len(operations))
	for _, operation := range operations {
		output := importOperationOutput(operation.Operation())
		output["message"] = operation.Message()
		outputs = append(outputs, output)
	}

	return outputs
}

func failedOperationOutputs(operations []domain.FailedOperation) []map[string]any {
	outputs := make([]map[string]any, 0, len(operations))
	for _, operation := range operations {
		output := importOperationOutput(operation.Operation())
		output["error"] = operation.Err().Error()
		outputs = append(outputs, output)
	}

	return outputs
}

func importOperationOutputs(operations []domain.ImportOperation) []map[string]any {
	outputs := make([]map[string]any, 0, len(operations))
	for _, operation := range operations {
		outputs = append(outputs, importOperationOutput(operation))
	}

	return outputs
}

func importOperationOutput(operation domain.ImportOperation) map[string]any {
	return map[string]any{
		"kind":        operation.Kind().String(),
		"entity_type": operation.EntityType().String(),
		"entity_ref":  operation.EntityRef(),
		"target_id":   importTargetID(operation),
	}
}

func formatOptionalInt(value *int) string {
	if value == nil {
		return ""
	}

	return strconv.Itoa(*value)
}

func joinInts(values []int) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.Itoa(value))
	}

	return strings.Join(parts, ",")
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}

	return value.Format(time.RFC3339)
}
