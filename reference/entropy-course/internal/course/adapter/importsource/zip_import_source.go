package importsource

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"path"
	"sort"
	"strings"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
	"gopkg.in/yaml.v3"
)

const supportedFormatVersion = "1"

var _ core.ImportSource = (*ZipImportSource)(nil)

type ZipImportSource struct{}

func NewZipImportSource() *ZipImportSource {
	return &ZipImportSource{}
}

func (source *ZipImportSource) Open(zipPath string) (core.ParsedImportSource, core.ImportSourceMetadata, error) {
	files, err := readZipFiles(zipPath)
	if err != nil {
		return core.ParsedImportSource{}, core.ImportSourceMetadata{}, err
	}

	formatVersion, err := parseFormatVersion(zipPath, files)
	if err != nil {
		return core.ParsedImportSource{}, core.ImportSourceMetadata{}, err
	}
	if formatVersion != supportedFormatVersion {
		return core.ParsedImportSource{}, core.ImportSourceMetadata{}, domain.NewUnsupportedImportFormatError(formatVersion, []string{supportedFormatVersion})
	}

	parsed, err := parseImportFiles(zipPath, formatVersion, files)
	if err != nil {
		return core.ParsedImportSource{}, core.ImportSourceMetadata{}, err
	}

	metadata := core.ImportSourceMetadata{
		ZipHash:       canonicalZipHash(files),
		FormatVersion: formatVersion,
	}

	return parsed, metadata, nil
}

type zipFileMap map[string][]byte

func readZipFiles(zipPath string) (zipFileMap, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, domain.NewImportSourceParseError(zipPath, "open zip", err)
	}
	defer reader.Close()

	files := zipFileMap{}
	for _, file := range reader.File {
		name, err := normalizeZipPath(file.Name)
		if err != nil {
			return nil, domain.NewImportSourceLayoutError(zipPath, err.Error(), err)
		}
		if name == "" {
			continue
		}
		if !isImportZipPath(name) {
			return nil, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("unexpected file %q", name), nil)
		}
		if _, exists := files[name]; exists {
			return nil, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("duplicate file %q", name), nil)
		}

		content, err := readZipFile(file)
		if err != nil {
			return nil, domain.NewImportSourceParseError(zipPath, fmt.Sprintf("read %s", name), err)
		}
		files[name] = content
	}

	return files, nil
}

func readZipFile(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func normalizeZipPath(name string) (string, error) {
	normalized := path.Clean(strings.TrimSpace(strings.ReplaceAll(name, "\\", "/")))
	if normalized == "." || strings.HasSuffix(name, "/") {
		return "", nil
	}
	if strings.HasPrefix(normalized, "../") || normalized == ".." || strings.HasPrefix(normalized, "/") {
		return "", fmt.Errorf("unsafe file path %q", name)
	}

	return normalized, nil
}

func isImportZipPath(name string) bool {
	switch {
	case name == "format_version.txt" || name == "course.yaml":
		return true
	case strings.HasPrefix(name, "lessons/") && strings.HasSuffix(name, ".md"):
		return true
	case strings.HasPrefix(name, "quizzes/") && strings.HasSuffix(name, ".yaml"):
		return true
	case strings.HasPrefix(name, "practices/") && strings.HasSuffix(name, ".yaml"):
		return true
	case strings.HasPrefix(name, "tests/") && strings.HasSuffix(name, ".yaml"):
		return true
	default:
		return false
	}
}

func parseFormatVersion(zipPath string, files zipFileMap) (string, error) {
	content, exists := files["format_version.txt"]
	if !exists {
		return "", domain.NewImportSourceLayoutError(zipPath, "missing format_version.txt", nil)
	}

	formatVersion := strings.TrimSpace(string(content))
	if formatVersion == "" {
		return "", domain.NewImportSourceLayoutError(zipPath, "missing format version", nil)
	}

	return formatVersion, nil
}

func parseImportFiles(zipPath string, formatVersion string, files zipFileMap) (core.ParsedImportSource, error) {
	course, err := parseCourse(zipPath, files)
	if err != nil {
		return core.ParsedImportSource{}, err
	}

	quizzes, err := parseQuizzes(zipPath, files)
	if err != nil {
		return core.ParsedImportSource{}, err
	}

	practices, err := parsePractices(zipPath, files)
	if err != nil {
		return core.ParsedImportSource{}, err
	}

	tests, err := parseTests(zipPath, files)
	if err != nil {
		return core.ParsedImportSource{}, err
	}

	lessons, err := parseLessons(zipPath, files, quizSlugSet(quizzes), practiceSlugSet(practices))
	if err != nil {
		return core.ParsedImportSource{}, err
	}

	return core.ParsedImportSource{
		FormatVersion: formatVersion,
		Course:        course,
		Lessons:       lessons,
		Quizzes:       quizzes,
		Practices:     practices,
		Tests:         tests,
	}, nil
}

func parseCourse(zipPath string, files zipFileMap) (core.ParsedCourse, error) {
	content, exists := files["course.yaml"]
	if !exists {
		return core.ParsedCourse{}, domain.NewImportSourceLayoutError(zipPath, "missing course.yaml", nil)
	}

	var course courseYAML
	if err := unmarshalYAML("course.yaml", content, &course); err != nil {
		return core.ParsedCourse{}, domain.NewImportSourceParseError(zipPath, "parse course.yaml", err)
	}

	parsed := core.ParsedCourse{
		Title:       strings.TrimSpace(course.Title),
		Slug:        strings.TrimSpace(course.Slug),
		Description: course.Description,
		Status:      strings.TrimSpace(course.Status),
	}
	if parsed.Title == "" {
		return core.ParsedCourse{}, domain.NewImportSourceLayoutError(zipPath, "course.yaml missing title", nil)
	}
	if parsed.Slug == "" {
		return core.ParsedCourse{}, domain.NewImportSourceLayoutError(zipPath, "course.yaml missing slug", nil)
	}
	if parsed.Status == "" {
		return core.ParsedCourse{}, domain.NewImportSourceLayoutError(zipPath, "course.yaml missing status", nil)
	}

	return parsed, nil
}

func parseLessons(zipPath string, files zipFileMap, quizSlugs map[string]struct{}, practiceSlugs map[string]struct{}) ([]core.ParsedLesson, error) {
	names := sortedFiles(files, "lessons/", ".md")
	lessons := make([]core.ParsedLesson, 0, len(names))
	for _, name := range names {
		lesson, err := parseLesson(zipPath, name, files[name], quizSlugs, practiceSlugs)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

func parseLesson(zipPath string, name string, content []byte, quizSlugs map[string]struct{}, practiceSlugs map[string]struct{}) (core.ParsedLesson, error) {
	frontmatter, body, err := splitFrontmatter(content)
	if err != nil {
		return core.ParsedLesson{}, domain.NewImportSourceParseError(zipPath, fmt.Sprintf("parse %s frontmatter", name), err)
	}

	var lesson lessonYAML
	if err := unmarshalYAML(name, frontmatter, &lesson); err != nil {
		return core.ParsedLesson{}, domain.NewImportSourceParseError(zipPath, fmt.Sprintf("parse %s", name), err)
	}

	parsed := core.ParsedLesson{
		Title: strings.TrimSpace(lesson.Title),
		Order: lesson.Order,
	}
	if parsed.Title == "" {
		return core.ParsedLesson{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s missing title", name), nil)
	}
	if len(lesson.Blocks) == 0 {
		return core.ParsedLesson{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s missing blocks", name), nil)
	}

	for i, block := range lesson.Blocks {
		parsedBlock, err := parseLessonBlock(zipPath, name, i, block, body, quizSlugs, practiceSlugs)
		if err != nil {
			return core.ParsedLesson{}, err
		}
		parsed.Blocks = append(parsed.Blocks, parsedBlock)
	}

	return parsed, nil
}

func parseLessonBlock(
	zipPath string,
	fileName string,
	index int,
	block lessonBlockYAML,
	body string,
	quizSlugs map[string]struct{},
	practiceSlugs map[string]struct{},
) (core.ParsedLessonBlock, error) {
	kind := strings.TrimSpace(block.Kind)
	if kind == "" {
		return core.ParsedLessonBlock{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s block %d missing kind", fileName, index), nil)
	}

	parsed := core.ParsedLessonBlock{
		Kind:          kind,
		Markdown:      block.Markdown,
		VideoProvider: strings.TrimSpace(block.VideoProvider),
		VideoLocator:  strings.TrimSpace(block.VideoLocator),
		VideoCaption:  block.VideoCaption,
		QuizRef:       strings.TrimSpace(block.QuizRef),
		PracticeRef:   strings.TrimSpace(block.PracticeRef),
		Position:      block.Position,
	}

	switch kind {
	case "text":
		if strings.TrimSpace(parsed.Markdown) == "" && strings.TrimSpace(body) != "" {
			parsed.Markdown = body
		}
		if strings.TrimSpace(parsed.Markdown) == "" {
			return core.ParsedLessonBlock{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s text block %d missing markdown", fileName, index), nil)
		}
	case "video":
		if parsed.VideoProvider == "" || parsed.VideoLocator == "" {
			return core.ParsedLessonBlock{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s video block %d missing provider or locator", fileName, index), nil)
		}
	case "quiz":
		if parsed.QuizRef == "" {
			return core.ParsedLessonBlock{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s quiz block %d missing quiz_ref", fileName, index), nil)
		}
		if _, exists := quizSlugs[parsed.QuizRef]; !exists {
			return core.ParsedLessonBlock{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s quiz block %d references unknown quiz %q", fileName, index, parsed.QuizRef), nil)
		}
	case "practice":
		if parsed.PracticeRef == "" {
			return core.ParsedLessonBlock{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s practice block %d missing practice_ref", fileName, index), nil)
		}
		if _, exists := practiceSlugs[parsed.PracticeRef]; !exists {
			return core.ParsedLessonBlock{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s practice block %d references unknown practice %q", fileName, index, parsed.PracticeRef), nil)
		}
	default:
		return core.ParsedLessonBlock{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s block %d has unknown kind %q", fileName, index, kind), nil)
	}

	return parsed, nil
}

func parseQuizzes(zipPath string, files zipFileMap) ([]core.ParsedQuiz, error) {
	names := sortedFiles(files, "quizzes/", ".yaml")
	quizzes := make([]core.ParsedQuiz, 0, len(names))
	slugs := map[string]struct{}{}
	for _, name := range names {
		var quiz quizYAML
		if err := unmarshalYAML(name, files[name], &quiz); err != nil {
			return nil, domain.NewImportSourceParseError(zipPath, fmt.Sprintf("parse %s", name), err)
		}

		parsed, err := parseQuiz(zipPath, name, quiz)
		if err != nil {
			return nil, err
		}
		if err := recordSlug(zipPath, "quiz", parsed.Slug, slugs); err != nil {
			return nil, err
		}
		quizzes = append(quizzes, parsed)
	}

	return quizzes, nil
}

func parseQuiz(zipPath string, name string, quiz quizYAML) (core.ParsedQuiz, error) {
	parsed := core.ParsedQuiz{
		Slug:          strings.TrimSpace(quiz.Slug),
		Title:         strings.TrimSpace(quiz.Title),
		PassThreshold: quiz.PassThreshold,
	}
	if parsed.Slug == "" {
		return core.ParsedQuiz{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s missing slug", name), nil)
	}
	if parsed.Title == "" {
		return core.ParsedQuiz{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s missing title", name), nil)
	}

	for i, question := range quiz.Questions {
		parsedQuestion := core.ParsedQuestion{
			Type:           strings.TrimSpace(question.Type),
			Prompt:         strings.TrimSpace(question.Prompt),
			Options:        question.Options,
			CorrectIndices: question.CorrectIndices,
			Explanation:    question.Explanation,
			Position:       question.Position,
		}
		if parsedQuestion.Type == "" || parsedQuestion.Prompt == "" {
			return core.ParsedQuiz{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s question %d missing type or prompt", name, i), nil)
		}
		if len(parsedQuestion.Options) == 0 {
			return core.ParsedQuiz{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s question %d missing options", name, i), nil)
		}
		parsed.Questions = append(parsed.Questions, parsedQuestion)
	}

	return parsed, nil
}

func parsePractices(zipPath string, files zipFileMap) ([]core.ParsedPractice, error) {
	names := sortedFiles(files, "practices/", ".yaml")
	practices := make([]core.ParsedPractice, 0, len(names))
	slugs := map[string]struct{}{}
	for _, name := range names {
		var practice practiceYAML
		if err := unmarshalYAML(name, files[name], &practice); err != nil {
			return nil, domain.NewImportSourceParseError(zipPath, fmt.Sprintf("parse %s", name), err)
		}

		parsed, err := parsePractice(zipPath, name, practice)
		if err != nil {
			return nil, err
		}
		if err := recordSlug(zipPath, "practice", parsed.Slug, slugs); err != nil {
			return nil, err
		}
		practices = append(practices, parsed)
	}

	return practices, nil
}

func parsePractice(zipPath string, name string, practice practiceYAML) (core.ParsedPractice, error) {
	parsed := core.ParsedPractice{
		Slug:        strings.TrimSpace(practice.Slug),
		Title:       strings.TrimSpace(practice.Title),
		Language:    strings.TrimSpace(practice.Language),
		Prompt:      strings.TrimSpace(practice.Prompt),
		StarterCode: practice.StarterCode,
		Solution:    practice.Solution,
	}
	if parsed.Slug == "" {
		return core.ParsedPractice{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s missing slug", name), nil)
	}
	if parsed.Title == "" || parsed.Language == "" || parsed.Prompt == "" {
		return core.ParsedPractice{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s missing title, language, or prompt", name), nil)
	}

	for i, testCase := range practice.TestCases {
		parsedCase := core.ParsedPracticeTestCase{
			Stdin:          testCase.Stdin,
			ExpectedStdout: testCase.ExpectedStdout,
			Name:           testCase.Name,
			Position:       testCase.Position,
		}
		if parsedCase.ExpectedStdout == "" {
			return core.ParsedPractice{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s test case %d missing expected_stdout", name, i), nil)
		}
		parsed.TestCases = append(parsed.TestCases, parsedCase)
	}

	return parsed, nil
}

func parseTests(zipPath string, files zipFileMap) ([]core.ParsedTest, error) {
	names := sortedFiles(files, "tests/", ".yaml")
	tests := make([]core.ParsedTest, 0, len(names))
	slugs := map[string]struct{}{}
	for _, name := range names {
		var test testYAML
		if err := unmarshalYAML(name, files[name], &test); err != nil {
			return nil, domain.NewImportSourceParseError(zipPath, fmt.Sprintf("parse %s", name), err)
		}

		parsed, err := parseTest(zipPath, name, test)
		if err != nil {
			return nil, err
		}
		if err := recordSlug(zipPath, "test", parsed.Slug, slugs); err != nil {
			return nil, err
		}
		tests = append(tests, parsed)
	}

	return tests, nil
}

func parseTest(zipPath string, name string, test testYAML) (core.ParsedTest, error) {
	parsed := core.ParsedTest{
		Slug:             strings.TrimSpace(test.Slug),
		Title:            strings.TrimSpace(test.Title),
		TimeLimitMinutes: test.TimeLimitMinutes,
		PassThreshold:    test.PassThreshold,
	}
	if parsed.Slug == "" {
		return core.ParsedTest{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s missing slug", name), nil)
	}
	if parsed.Title == "" {
		return core.ParsedTest{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s missing title", name), nil)
	}
	if test.Solution != nil {
		parsed.Solution = &core.ParsedTestSolution{
			ZipProvider:   strings.TrimSpace(test.Solution.ZipProvider),
			ZipLocator:    strings.TrimSpace(test.Solution.ZipLocator),
			VideoProvider: strings.TrimSpace(test.Solution.VideoProvider),
			VideoLocator:  strings.TrimSpace(test.Solution.VideoLocator),
			VideoCaption:  test.Solution.VideoCaption,
		}
	}

	for i, item := range test.Items {
		parsedItem, err := parseTestItem(zipPath, name, i, item)
		if err != nil {
			return core.ParsedTest{}, err
		}
		parsed.Items = append(parsed.Items, parsedItem)
	}

	return parsed, nil
}

func parseTestItem(zipPath string, name string, index int, item testItemYAML) (core.ParsedTestItem, error) {
	kind := strings.TrimSpace(item.Kind)
	if kind == "" {
		return core.ParsedTestItem{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s item %d missing kind", name, index), nil)
	}

	parsed := core.ParsedTestItem{
		Kind:           kind,
		Position:       item.Position,
		Prompt:         strings.TrimSpace(item.Prompt),
		ChoiceType:     strings.TrimSpace(item.ChoiceType),
		Options:        item.Options,
		CorrectIndices: item.CorrectIndices,
		Explanation:    item.Explanation,
		CodingPrompt:   strings.TrimSpace(item.CodingPrompt),
		Language:       strings.TrimSpace(item.Language),
		StarterCode:    item.StarterCode,
		Solution:       item.Solution,
	}

	switch kind {
	case "choice":
		if parsed.Prompt == "" || parsed.ChoiceType == "" || len(parsed.Options) == 0 {
			return core.ParsedTestItem{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s choice item %d missing prompt, choice_type, or options", name, index), nil)
		}
	case "coding":
		if parsed.CodingPrompt == "" || parsed.Language == "" {
			return core.ParsedTestItem{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s coding item %d missing coding_prompt or language", name, index), nil)
		}
	default:
		return core.ParsedTestItem{}, domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s item %d has unknown kind %q", name, index, kind), nil)
	}

	for _, testCase := range item.TestCases {
		parsed.TestCases = append(parsed.TestCases, core.ParsedCodingTestCase{
			Stdin:          testCase.Stdin,
			ExpectedStdout: testCase.ExpectedStdout,
			Name:           testCase.Name,
		})
	}

	return parsed, nil
}

func splitFrontmatter(content []byte) ([]byte, string, error) {
	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return nil, "", fmt.Errorf("missing YAML frontmatter")
	}

	rest := normalized[len("---\n"):]
	end := strings.Index(rest, "\n---\n")
	if end < 0 {
		if strings.HasSuffix(rest, "\n---") {
			end = len(rest) - len("\n---")
		} else {
			return nil, "", fmt.Errorf("unterminated YAML frontmatter")
		}
	}

	frontmatter := rest[:end]
	bodyStart := end + len("\n---\n")
	body := ""
	if bodyStart <= len(rest) {
		body = rest[bodyStart:]
	}

	return []byte(frontmatter), strings.Trim(body, "\n"), nil
}

func unmarshalYAML(name string, content []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}

	return nil
}

func sortedFiles(files zipFileMap, prefix string, suffix string) []string {
	names := []string{}
	for name := range files {
		if strings.HasPrefix(name, prefix) && strings.HasSuffix(name, suffix) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func recordSlug(zipPath string, entityName string, slug string, slugs map[string]struct{}) error {
	if slug == "" {
		return domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("%s missing slug", entityName), nil)
	}
	if _, exists := slugs[slug]; exists {
		return domain.NewImportSourceLayoutError(zipPath, fmt.Sprintf("duplicate %s slug %q", entityName, slug), nil)
	}

	slugs[slug] = struct{}{}
	return nil
}

func quizSlugSet(quizzes []core.ParsedQuiz) map[string]struct{} {
	slugs := map[string]struct{}{}
	for _, quiz := range quizzes {
		slugs[quiz.Slug] = struct{}{}
	}
	return slugs
}

func practiceSlugSet(practices []core.ParsedPractice) map[string]struct{} {
	slugs := map[string]struct{}{}
	for _, practice := range practices {
		slugs[practice.Slug] = struct{}{}
	}
	return slugs
}

func canonicalZipHash(files zipFileMap) string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	hash := sha256.New()
	for _, name := range names {
		hash.Write([]byte(name))
		hash.Write([]byte{0})
		hash.Write(files[name])
		hash.Write([]byte{0})
	}

	return hex.EncodeToString(hash.Sum(nil))
}

type courseYAML struct {
	Title       string `yaml:"title"`
	Slug        string `yaml:"slug"`
	Description string `yaml:"description"`
	Status      string `yaml:"status"`
}

type lessonYAML struct {
	Title  string            `yaml:"title"`
	Order  *int              `yaml:"order"`
	Blocks []lessonBlockYAML `yaml:"blocks"`
}

type lessonBlockYAML struct {
	Kind          string `yaml:"kind"`
	Markdown      string `yaml:"markdown"`
	VideoProvider string `yaml:"video_provider"`
	VideoLocator  string `yaml:"video_locator"`
	VideoCaption  string `yaml:"video_caption"`
	QuizRef       string `yaml:"quiz_ref"`
	PracticeRef   string `yaml:"practice_ref"`
	Position      *int   `yaml:"position"`
}

type quizYAML struct {
	Slug          string         `yaml:"slug"`
	Title         string         `yaml:"title"`
	PassThreshold *float64       `yaml:"pass_threshold"`
	Questions     []questionYAML `yaml:"questions"`
}

type questionYAML struct {
	Type           string   `yaml:"type"`
	Prompt         string   `yaml:"prompt"`
	Options        []string `yaml:"options"`
	CorrectIndices []int    `yaml:"correct_indices"`
	Explanation    string   `yaml:"explanation"`
	Position       *int     `yaml:"position"`
}

type practiceYAML struct {
	Slug        string                 `yaml:"slug"`
	Title       string                 `yaml:"title"`
	Language    string                 `yaml:"language"`
	Prompt      string                 `yaml:"prompt"`
	StarterCode string                 `yaml:"starter_code"`
	Solution    string                 `yaml:"solution"`
	TestCases   []practiceTestCaseYAML `yaml:"test_cases"`
}

type practiceTestCaseYAML struct {
	Stdin          string `yaml:"stdin"`
	ExpectedStdout string `yaml:"expected_stdout"`
	Name           string `yaml:"name"`
	Position       *int   `yaml:"position"`
}

type testYAML struct {
	Slug             string            `yaml:"slug"`
	Title            string            `yaml:"title"`
	TimeLimitMinutes *int              `yaml:"time_limit_minutes"`
	PassThreshold    *float64          `yaml:"pass_threshold"`
	Solution         *testSolutionYAML `yaml:"solution"`
	Items            []testItemYAML    `yaml:"items"`
}

type testSolutionYAML struct {
	ZipProvider   string `yaml:"zip_provider"`
	ZipLocator    string `yaml:"zip_locator"`
	VideoProvider string `yaml:"video_provider"`
	VideoLocator  string `yaml:"video_locator"`
	VideoCaption  string `yaml:"video_caption"`
}

type testItemYAML struct {
	Kind     string `yaml:"kind"`
	Position *int   `yaml:"position"`

	Prompt         string   `yaml:"prompt"`
	ChoiceType     string   `yaml:"choice_type"`
	Options        []string `yaml:"options"`
	CorrectIndices []int    `yaml:"correct_indices"`
	Explanation    string   `yaml:"explanation"`

	CodingPrompt string               `yaml:"coding_prompt"`
	Language     string               `yaml:"language"`
	StarterCode  string               `yaml:"starter_code"`
	Solution     string               `yaml:"solution"`
	TestCases    []codingTestCaseYAML `yaml:"test_cases"`
}

type codingTestCaseYAML struct {
	Stdin          string `yaml:"stdin"`
	ExpectedStdout string `yaml:"expected_stdout"`
	Name           string `yaml:"name"`
}
