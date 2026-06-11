package rest

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

var (
	ErrMissingCourseService   = errors.New("course service is required")
	ErrMissingLessonService   = errors.New("lesson service is required")
	ErrMissingQuizService     = errors.New("quiz service is required")
	ErrMissingPracticeService = errors.New("practice service is required")
	ErrMissingTestService     = errors.New("test service is required")
	ErrMissingImportService   = errors.New("import service is required")
	ErrMissingAPIToken        = errors.New("api token is required")
	ErrUnauthorized           = errors.New("unauthorized")
)

const DefaultAddress = "127.0.0.1:8788"

type Server struct {
	course       core.CourseService
	lesson       core.LessonService
	quiz         core.QuizService
	practice     core.PracticeService
	test         core.TestService
	imports      core.ImportService
	token        string
	instructorID string
}

type Options struct {
	Course       core.CourseService
	Lesson       core.LessonService
	Quiz         core.QuizService
	Practice     core.PracticeService
	Test         core.TestService
	Import       core.ImportService
	Token        string
	InstructorID string
}

type errorResponse struct {
	Error string `json:"error"`
}

type quizInUseResponse struct {
	Error     string   `json:"error"`
	LessonIDs []string `json:"lessonIds"`
}

type practiceInUseResponse struct {
	Error     string   `json:"error"`
	LessonIDs []string `json:"lessonIds"`
}

func NewServer(options Options) (*Server, error) {
	token := strings.TrimSpace(options.Token)
	if options.Course == nil {
		return nil, ErrMissingCourseService
	}
	if options.Lesson == nil {
		return nil, ErrMissingLessonService
	}
	if options.Quiz == nil {
		return nil, ErrMissingQuizService
	}
	if options.Practice == nil {
		return nil, ErrMissingPracticeService
	}
	if options.Test == nil {
		return nil, ErrMissingTestService
	}
	if token == "" {
		return nil, ErrMissingAPIToken
	}

	return &Server{
		course:       options.Course,
		lesson:       options.Lesson,
		quiz:         options.Quiz,
		practice:     options.Practice,
		test:         options.Test,
		imports:      options.Import,
		token:        token,
		instructorID: strings.TrimSpace(options.InstructorID),
	}, nil
}

func (server *Server) Handler() http.Handler {
	return server.authenticate(http.HandlerFunc(server.route))
}

func (server *Server) ListenAndServe(ctx context.Context, address string) error {
	if strings.TrimSpace(address) == "" {
		address = DefaultAddress
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}

	httpServer := &http.Server{
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(listener)
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown rest server: %w", err)
		}
	}

	err = <-errCh
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (server *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if !server.authorized(request.Header.Get("Authorization")) {
			writeError(response, ErrUnauthorized)
			return
		}

		next.ServeHTTP(response, request)
	})
}

func (server *Server) authorized(header string) bool {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || scheme != "Bearer" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(token), []byte(server.token)) == 1
}

func (server *Server) route(response http.ResponseWriter, request *http.Request) {
	segments := pathSegments(request.URL.Path)
	if len(segments) == 0 || segments[0] != "v1" {
		http.NotFound(response, request)
		return
	}

	switch {
	case len(segments) == 3 && segments[1] == "import" && segments[2] == "plan":
		server.handleImportPlan(response, request)
	case len(segments) == 3 && segments[1] == "import" && segments[2] == "apply":
		server.handleImportApply(response, request)
	case len(segments) == 2 && segments[1] == "courses":
		server.handleCourses(response, request)
	case len(segments) == 4 && segments[1] == "courses" && segments[3] == "publish":
		server.handleCoursePublish(response, request, segments[2])
	case len(segments) == 4 && segments[1] == "courses" && segments[3] == "unpublish":
		server.handleCourseUnpublish(response, request, segments[2])
	case len(segments) == 4 && segments[1] == "courses" && segments[3] == "lessons":
		server.handleCourseLessons(response, request, segments[2])
	case len(segments) == 5 && segments[1] == "courses" && segments[3] == "lessons" && segments[4] == "reorder":
		server.handleCourseLessonsReorder(response, request, segments[2])
	case len(segments) == 3 && segments[1] == "courses":
		server.handleCourse(response, request, segments[2])
	case len(segments) == 3 && segments[1] == "lessons":
		server.handleLesson(response, request, segments[2])
	case len(segments) == 3 && segments[1] == "test-items":
		server.handleTestItem(response, request, segments[2])
	case len(segments) == 2 && segments[1] == "tests":
		server.handleTests(response, request)
	case len(segments) == 4 && segments[1] == "courses" && segments[3] == "tests":
		server.handleCourseTests(response, request, segments[2])
	case len(segments) == 3 && segments[1] == "tests":
		server.handleTest(response, request, segments[2])
	case len(segments) == 4 && segments[1] == "tests" && segments[3] == "items":
		server.handleTestItems(response, request, segments[2])
	case len(segments) == 5 && segments[1] == "tests" && segments[3] == "items" && segments[4] == "reorder":
		server.handleReorderTestItems(response, request, segments[2])
	case len(segments) == 3 && segments[1] == "practices":
		server.handlePractice(response, request, segments[2])
	case len(segments) == 3 && segments[1] == "testcases":
		server.handleTestCase(response, request, segments[2])
	case len(segments) == 2 && segments[1] == "practices":
		server.handlePractices(response, request)
	case len(segments) == 4 && segments[1] == "courses" && segments[3] == "practices":
		server.handleCoursePractices(response, request, segments[2])
	case len(segments) == 4 && segments[1] == "practices" && segments[3] == "testcases":
		server.handlePracticeTestCases(response, request, segments[2])
	case len(segments) == 5 && segments[1] == "practices" && segments[3] == "testcases" && segments[4] == "reorder":
		server.handleReorderTestCases(response, request, segments[2])
	case len(segments) == 3 && segments[1] == "quizzes":
		server.handleQuiz(response, request, segments[2])
	case len(segments) == 3 && segments[1] == "questions":
		server.handleQuestion(response, request, segments[2])
	case len(segments) == 2 && segments[1] == "quizzes":
		server.handleQuizzes(response, request)
	case len(segments) == 4 && segments[1] == "courses" && segments[3] == "quizzes":
		server.handleCourseQuizzes(response, request, segments[2])
	case len(segments) == 4 && segments[1] == "quizzes" && segments[3] == "questions":
		server.handleQuizQuestions(response, request, segments[2])
	case len(segments) == 5 && segments[1] == "quizzes" && segments[3] == "questions" && segments[4] == "reorder":
		server.handleReorderQuestions(response, request, segments[2])
	case len(segments) == 4 && segments[1] == "lessons" && segments[3] == "blocks":
		server.handleLessonBlocks(response, request, segments[2])
	case len(segments) == 5 && segments[1] == "lessons" && segments[3] == "blocks" && segments[4] == "reorder":
		server.handleReorderLessonBlocks(response, request, segments[2])
	case len(segments) == 3 && segments[1] == "blocks":
		server.handleBlock(response, request, segments[2])
	default:
		http.NotFound(response, request)
	}
}

func (server *Server) handleQuizzes(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	var input core.CreateQuizInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(response, err)
		return
	}

	out, err := server.quiz.CreateQuiz(input)
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusCreated, out)
}

func (server *Server) handleCourseQuizzes(response http.ResponseWriter, request *http.Request, courseID string) {
	if request.Method != http.MethodGet {
		methodNotAllowed(response, http.MethodGet)
		return
	}

	out, err := server.quiz.ListQuizzes(core.ListQuizzesInput{CourseID: courseID})
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, out)
}

func (server *Server) handleQuiz(response http.ResponseWriter, request *http.Request, quizID string) {
	switch request.Method {
	case http.MethodGet:
		learnerRead, err := learnerReadRequested(request)
		if err != nil {
			writeError(response, err)
			return
		}

		out, err := server.quiz.GetQuiz(core.GetQuizInput{ID: quizID})
		if err != nil {
			writeError(response, err)
			return
		}
		if learnerRead {
			if err := server.ensureLearnerCoursePublished(out.Quiz.CourseID); err != nil {
				writeError(response, err)
				return
			}
			writeJSON(response, http.StatusOK, learnerQuizOutput(out))
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPatch:
		var input core.UpdateQuizInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.ID = quizID

		out, err := server.quiz.UpdateQuiz(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodDelete:
		if err := server.quiz.DeleteQuiz(core.DeleteQuizInput{ID: quizID}); err != nil {
			writeError(response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodDelete, http.MethodGet, http.MethodPatch)
	}
}

func (server *Server) handleQuizQuestions(response http.ResponseWriter, request *http.Request, quizID string) {
	switch request.Method {
	case http.MethodPost:
		var input core.AddQuestionInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.QuizID = quizID

		out, err := server.quiz.AddQuestion(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusCreated, out)
	case http.MethodGet:
		out, err := server.quiz.ListQuestions(core.ListQuestionsInput{QuizID: quizID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	default:
		methodNotAllowed(response, http.MethodGet, http.MethodPost)
	}
}

func (server *Server) handleQuestion(response http.ResponseWriter, request *http.Request, questionID string) {
	switch request.Method {
	case http.MethodGet:
		out, err := server.quiz.GetQuestion(core.GetQuestionInput{ID: questionID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPatch:
		var input core.UpdateQuestionInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.ID = questionID

		out, err := server.quiz.UpdateQuestion(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodDelete:
		if err := server.quiz.RemoveQuestion(core.RemoveQuestionInput{ID: questionID}); err != nil {
			writeError(response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodDelete, http.MethodGet, http.MethodPatch)
	}
}

func (server *Server) handleReorderQuestions(response http.ResponseWriter, request *http.Request, quizID string) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	var input core.ReorderQuestionsInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(response, err)
		return
	}
	input.QuizID = quizID

	if err := server.quiz.ReorderQuestions(input); err != nil {
		writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (server *Server) handleTests(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	var input core.CreateTestInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(response, err)
		return
	}

	out, err := server.test.CreateTest(input)
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusCreated, out)
}

func (server *Server) handleCourseTests(response http.ResponseWriter, request *http.Request, courseID string) {
	if request.Method != http.MethodGet {
		methodNotAllowed(response, http.MethodGet)
		return
	}

	out, err := server.test.ListTests(core.ListTestsInput{CourseID: courseID})
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, out)
}

func (server *Server) handleTest(response http.ResponseWriter, request *http.Request, testID string) {
	switch request.Method {
	case http.MethodGet:
		learnerRead, err := learnerReadRequested(request)
		if err != nil {
			writeError(response, err)
			return
		}

		out, err := server.test.GetTest(core.GetTestInput{ID: testID})
		if err != nil {
			writeError(response, err)
			return
		}
		if learnerRead {
			if err := server.ensureLearnerCoursePublished(out.Test.CourseID); err != nil {
				writeError(response, err)
				return
			}
			writeJSON(response, http.StatusOK, learnerTestOutput(out))
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPatch:
		var input core.UpdateTestInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.ID = testID

		out, err := server.test.UpdateTest(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodDelete:
		if err := server.test.DeleteTest(core.DeleteTestInput{ID: testID}); err != nil {
			writeError(response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodDelete, http.MethodGet, http.MethodPatch)
	}
}

func (server *Server) handleTestItems(response http.ResponseWriter, request *http.Request, testID string) {
	switch request.Method {
	case http.MethodPost:
		var input core.AddTestItemInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.TestID = testID

		out, err := server.test.AddTestItem(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusCreated, out)
	case http.MethodGet:
		out, err := server.test.ListTestItems(core.ListTestItemsInput{TestID: testID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	default:
		methodNotAllowed(response, http.MethodGet, http.MethodPost)
	}
}

func (server *Server) handleTestItem(response http.ResponseWriter, request *http.Request, itemID string) {
	switch request.Method {
	case http.MethodGet:
		out, err := server.test.GetTestItem(core.GetTestItemInput{ID: itemID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPatch:
		var input core.UpdateTestItemInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.ID = itemID

		out, err := server.test.UpdateTestItem(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodDelete:
		if err := server.test.RemoveTestItem(core.RemoveTestItemInput{ID: itemID}); err != nil {
			writeError(response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodDelete, http.MethodGet, http.MethodPatch)
	}
}

func (server *Server) handleReorderTestItems(response http.ResponseWriter, request *http.Request, testID string) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	var input core.ReorderTestItemsInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(response, err)
		return
	}
	input.TestID = testID

	if err := server.test.ReorderTestItems(input); err != nil {
		writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (server *Server) handlePractices(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	var input core.CreatePracticeInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(response, err)
		return
	}

	out, err := server.practice.CreatePractice(input)
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusCreated, out)
}

func (server *Server) handleCoursePractices(response http.ResponseWriter, request *http.Request, courseID string) {
	if request.Method != http.MethodGet {
		methodNotAllowed(response, http.MethodGet)
		return
	}

	out, err := server.practice.ListPractices(core.ListPracticesInput{CourseID: courseID})
	if err != nil {
		writeError(response, err)
		return
	}
	writeJSON(response, http.StatusOK, out)
}

func (server *Server) handlePractice(response http.ResponseWriter, request *http.Request, practiceID string) {
	switch request.Method {
	case http.MethodGet:
		learnerRead, err := learnerReadRequested(request)
		if err != nil {
			writeError(response, err)
			return
		}

		out, err := server.practice.GetPractice(core.GetPracticeInput{ID: practiceID})
		if err != nil {
			writeError(response, err)
			return
		}
		if learnerRead {
			if err := server.ensureLearnerCoursePublished(out.Practice.CourseID); err != nil {
				writeError(response, err)
				return
			}
			writeJSON(response, http.StatusOK, learnerPracticeOutput(out))
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPatch:
		var input core.UpdatePracticeInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.ID = practiceID

		out, err := server.practice.UpdatePractice(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodDelete:
		if err := server.practice.DeletePractice(core.DeletePracticeInput{ID: practiceID}); err != nil {
			writeError(response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodDelete, http.MethodGet, http.MethodPatch)
	}
}

func (server *Server) handlePracticeTestCases(response http.ResponseWriter, request *http.Request, practiceID string) {
	switch request.Method {
	case http.MethodPost:
		var input core.AddTestCaseInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.PracticeID = practiceID

		out, err := server.practice.AddTestCase(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusCreated, out)
	case http.MethodGet:
		out, err := server.practice.ListTestCases(core.ListTestCasesInput{PracticeID: practiceID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	default:
		methodNotAllowed(response, http.MethodGet, http.MethodPost)
	}
}

func (server *Server) handleTestCase(response http.ResponseWriter, request *http.Request, testCaseID string) {
	switch request.Method {
	case http.MethodGet:
		out, err := server.practice.GetTestCase(core.GetTestCaseInput{ID: testCaseID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPatch:
		var input core.UpdateTestCaseInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.ID = testCaseID

		out, err := server.practice.UpdateTestCase(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodDelete:
		if err := server.practice.RemoveTestCase(core.RemoveTestCaseInput{ID: testCaseID}); err != nil {
			writeError(response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodDelete, http.MethodGet, http.MethodPatch)
	}
}

func (server *Server) handleReorderTestCases(response http.ResponseWriter, request *http.Request, practiceID string) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	var input core.ReorderTestCasesInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(response, err)
		return
	}
	input.PracticeID = practiceID

	if err := server.practice.ReorderTestCases(input); err != nil {
		writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (server *Server) handleLessonBlocks(response http.ResponseWriter, request *http.Request, lessonID string) {
	switch request.Method {
	case http.MethodPost:
		var input core.AddLessonBlockInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.LessonID = lessonID

		out, err := server.lesson.AddLessonBlock(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusCreated, out)
	case http.MethodGet:
		out, err := server.lesson.ListLessonBlocks(core.ListLessonBlocksInput{LessonID: lessonID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	default:
		methodNotAllowed(response, http.MethodGet, http.MethodPost)
	}
}

func (server *Server) handleBlock(response http.ResponseWriter, request *http.Request, blockID string) {
	switch request.Method {
	case http.MethodGet:
		out, err := server.lesson.GetLessonBlock(core.GetLessonBlockInput{ID: blockID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPatch:
		var input core.UpdateLessonBlockInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.ID = blockID

		out, err := server.lesson.UpdateLessonBlock(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodDelete:
		if err := server.lesson.RemoveLessonBlock(core.RemoveLessonBlockInput{ID: blockID}); err != nil {
			writeError(response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodDelete, http.MethodGet, http.MethodPatch)
	}
}

func (server *Server) handleReorderLessonBlocks(response http.ResponseWriter, request *http.Request, lessonID string) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	var input core.ReorderLessonBlocksInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(response, err)
		return
	}
	input.LessonID = lessonID

	if err := server.lesson.ReorderLessonBlocks(input); err != nil {
		writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func pathSegments(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}

	return strings.Split(trimmed, "/")
}

func decodeJSON(request *http.Request, value any) error {
	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(value); err != nil {
		return domain.NewValidationError("request", "invalid json")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return domain.NewValidationError("request", "must contain one json value")
	}

	return nil
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func writeError(response http.ResponseWriter, err error) {
	var quizInUse domain.QuizInUseError
	if errors.As(err, &quizInUse) {
		writeJSON(response, http.StatusConflict, quizInUseResponse{
			Error:     err.Error(),
			LessonIDs: lessonIDStrings(quizInUse.LessonIDs),
		})
		return
	}

	var practiceInUse domain.PracticeInUseError
	if errors.As(err, &practiceInUse) {
		writeJSON(response, http.StatusConflict, practiceInUseResponse{
			Error:     err.Error(),
			LessonIDs: lessonIDStrings(practiceInUse.LessonIDs),
		})
		return
	}

	status := statusForError(err)
	message := err.Error()
	if status == http.StatusInternalServerError {
		message = "internal error"
	}

	writeJSON(response, status, errorResponse{Error: message})
}

func statusForError(err error) int {
	switch {
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrQuizInUse):
		return http.StatusConflict
	case errors.Is(err, domain.ErrPracticeInUse):
		return http.StatusConflict
	case errors.Is(err, domain.ErrUnsupportedImportFormat):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrInvalidConflictStrategy):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrImportSourceParse):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrImportSourceLayout):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrImportPlanHashMismatch):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrUnresolvedImportConflicts):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func lessonIDStrings(ids []domain.LessonID) []string {
	values := make([]string, 0, len(ids))
	for _, id := range ids {
		values = append(values, id.String())
	}

	return values
}

func methodNotAllowed(response http.ResponseWriter, methods ...string) {
	response.Header().Set("Allow", strings.Join(methods, ", "))
	http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
}
