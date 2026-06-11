package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

func TestListCoursesReturnsCatalog(t *testing.T) {
	course := &courseReadFake{listOut: core.ListCoursesOutput{Courses: []core.CourseView{courseViewFixture()}}}
	response := readRequest(t, newReadServer(t, course, nil), http.MethodGet, "/v1/courses")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if course.called != "list" {
		t.Fatalf("expected list call, got %q", course.called)
	}

	var out core.ListCoursesOutput
	decodeResponse(t, response, &out)
	if len(out.Courses) != 1 || out.Courses[0].ID != courseIDValue {
		t.Fatalf("expected one course in catalog, got %+v", out)
	}
}

func TestListCoursesForwardsStatusFilter(t *testing.T) {
	course := &courseReadFake{}
	response := readRequest(t, newReadServer(t, course, nil), http.MethodGet, "/v1/courses?status=published")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if course.gotListCourses.Status != "published" {
		t.Fatalf("expected status filter forwarded, got %+v", course.gotListCourses)
	}
}

func TestCreateCourseParsesRequestAndUsesConfiguredInstructor(t *testing.T) {
	course := &courseReadFake{createOut: core.CreateCourseOutput{ID: courseIDValue}}
	response := courseRequest(t, newReadServer(t, course, nil), http.MethodPost, "/v1/courses", `{
		"title": "JavaScript",
		"slug": "javascript",
		"description": "Intro to JavaScript",
		"instructorID": "body-instructor"
	}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
	}
	if course.called != "create" {
		t.Fatalf("expected create call, got %q", course.called)
	}
	want := core.CreateCourseInput{
		Title:        "JavaScript",
		Slug:         "javascript",
		Description:  "Intro to JavaScript",
		InstructorID: "instructor-1",
	}
	if course.gotCreateCourse != want {
		t.Fatalf("expected create course input %+v, got %+v", want, course.gotCreateCourse)
	}

	var out core.CreateCourseOutput
	decodeResponse(t, response, &out)
	if out.ID != courseIDValue {
		t.Fatalf("expected course id response, got %+v", out)
	}
}

func TestCreateCourseRequiresConfiguredInstructor(t *testing.T) {
	course := &courseReadFake{}
	response := courseRequest(t, newReadServerWithInstructor(t, course, nil, ""), http.MethodPost, "/v1/courses", `{
		"title": "JavaScript",
		"slug": "javascript"
	}`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected validation response, got %d body=%s", response.Code, response.Body.String())
	}
	if course.called != "" {
		t.Fatalf("expected request not to reach service, got %q", course.called)
	}
}

func TestGetCourseRoutesByID(t *testing.T) {
	course := &courseReadFake{getOut: core.GetCourseOutput{Course: courseViewFixture()}}
	response := readRequest(t, newReadServer(t, course, nil), http.MethodGet, "/v1/courses/"+courseIDValue)

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if course.called != "get" || course.gotGetCourse.ID != courseIDValue {
		t.Fatalf("expected get course by id, got called=%q input=%+v", course.called, course.gotGetCourse)
	}

	var out core.GetCourseOutput
	decodeResponse(t, response, &out)
	if out.Course.ID != courseIDValue {
		t.Fatalf("expected course detail response, got %+v", out)
	}
}

func TestUpdateCourseParsesPointerFields(t *testing.T) {
	course := &courseReadFake{updateOut: core.UpdateCourseOutput{ID: courseIDValue}}
	response := courseRequest(t, newReadServer(t, course, nil), http.MethodPatch, "/v1/courses/"+courseIDValue, `{
		"title": "Advanced JavaScript",
		"description": ""
	}`)

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if course.called != "update" || course.gotUpdateCourse.ID != courseIDValue {
		t.Fatalf("expected update course input, got called=%q input=%+v", course.called, course.gotUpdateCourse)
	}
	if course.gotUpdateCourse.Title == nil || *course.gotUpdateCourse.Title != "Advanced JavaScript" {
		t.Fatalf("expected title pointer to be set, got %+v", course.gotUpdateCourse)
	}
	if course.gotUpdateCourse.Description == nil || *course.gotUpdateCourse.Description != "" {
		t.Fatalf("expected empty description pointer to be set, got %+v", course.gotUpdateCourse)
	}
	if course.gotUpdateCourse.Slug != nil {
		t.Fatalf("expected omitted slug to remain nil, got %+v", course.gotUpdateCourse)
	}

	var out core.UpdateCourseOutput
	decodeResponse(t, response, &out)
	if out.ID != courseIDValue {
		t.Fatalf("expected update id response, got %+v", out)
	}
}

func TestDeleteCourseRoutesByID(t *testing.T) {
	course := &courseReadFake{}
	response := courseRequest(t, newReadServer(t, course, nil), http.MethodDelete, "/v1/courses/"+courseIDValue, "")

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
	}
	if course.called != "delete" || course.gotDeleteCourse.ID != courseIDValue {
		t.Fatalf("expected delete course input, got called=%q input=%+v", course.called, course.gotDeleteCourse)
	}
}

func TestPublishAndUnpublishCourseRoutesByID(t *testing.T) {
	t.Run("publish", func(t *testing.T) {
		course := &courseReadFake{}
		response := courseRequest(t, newReadServer(t, course, nil), http.MethodPost, "/v1/courses/"+courseIDValue+"/publish", "")

		if response.Code != http.StatusNoContent {
			t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
		}
		if course.called != "publish" || course.gotPublishCourse.ID != courseIDValue {
			t.Fatalf("expected publish course input, got called=%q input=%+v", course.called, course.gotPublishCourse)
		}
	})

	t.Run("unpublish", func(t *testing.T) {
		course := &courseReadFake{}
		response := courseRequest(t, newReadServer(t, course, nil), http.MethodPost, "/v1/courses/"+courseIDValue+"/unpublish", "")

		if response.Code != http.StatusNoContent {
			t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
		}
		if course.called != "unpublish" || course.gotUnpublishCourse.ID != courseIDValue {
			t.Fatalf("expected unpublish course input, got called=%q input=%+v", course.called, course.gotUnpublishCourse)
		}
	})
}

func TestGetCourseNotFoundMapsTo404(t *testing.T) {
	course := &courseReadFake{err: domain.ErrNotFound}
	response := readRequest(t, newReadServer(t, course, nil), http.MethodGet, "/v1/courses/"+courseIDValue)

	if response.Code != http.StatusNotFound {
		t.Fatalf("expected not found response, got %d body=%s", response.Code, response.Body.String())
	}
}

func TestListCourseLessonsRoutesByCourseID(t *testing.T) {
	lesson := &lessonReadFake{lessonsOut: core.ListLessonsOutput{Lessons: []core.LessonView{lessonViewFixture()}}}
	response := readRequest(t, newReadServer(t, nil, lesson), http.MethodGet, "/v1/courses/"+courseIDValue+"/lessons")

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if lesson.readCalled != "list" || lesson.gotListLessons.CourseID != courseIDValue {
		t.Fatalf("expected list lessons by course id, got called=%q input=%+v", lesson.readCalled, lesson.gotListLessons)
	}

	var out core.ListLessonsOutput
	decodeResponse(t, response, &out)
	if len(out.Lessons) != 1 || out.Lessons[0].ID != lessonIDValue {
		t.Fatalf("expected one lesson, got %+v", out)
	}
}

func TestCreateLessonTakesCourseIDFromPath(t *testing.T) {
	order := 2
	lesson := &lessonReadFake{createOut: core.CreateLessonOutput{ID: lessonIDValue}}
	response := courseRequest(t, newReadServer(t, nil, lesson), http.MethodPost, "/v1/courses/"+courseIDValue+"/lessons", `{
		"courseID": "ignored-course",
		"title": "Variables and Data Types",
		"order": 2
	}`)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected created response, got %d body=%s", response.Code, response.Body.String())
	}
	if lesson.writeCalled != "create" || lesson.gotCreateLesson.CourseID != courseIDValue {
		t.Fatalf("expected create lesson input, got called=%q input=%+v", lesson.writeCalled, lesson.gotCreateLesson)
	}
	if lesson.gotCreateLesson.Title != "Variables and Data Types" || lesson.gotCreateLesson.Order == nil || *lesson.gotCreateLesson.Order != order {
		t.Fatalf("expected lesson body fields to map, got %+v", lesson.gotCreateLesson)
	}

	var out core.CreateLessonOutput
	decodeResponse(t, response, &out)
	if out.ID != lessonIDValue {
		t.Fatalf("expected lesson id response, got %+v", out)
	}
}

func TestGetLessonRoutesByID(t *testing.T) {
	lesson := &lessonReadFake{lessonOut: core.GetLessonOutput{Lesson: lessonViewFixture()}}
	response := readRequest(t, newReadServer(t, nil, lesson), http.MethodGet, "/v1/lessons/"+lessonIDValue)

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if lesson.readCalled != "get" || lesson.gotGetLesson.ID != lessonIDValue {
		t.Fatalf("expected get lesson by id, got called=%q input=%+v", lesson.readCalled, lesson.gotGetLesson)
	}

	var out core.GetLessonOutput
	decodeResponse(t, response, &out)
	if out.Lesson.ID != lessonIDValue {
		t.Fatalf("expected lesson detail response, got %+v", out)
	}
}

func TestUpdateLessonRoutesByID(t *testing.T) {
	lesson := &lessonReadFake{updateOut: core.UpdateLessonOutput{ID: lessonIDValue}}
	response := courseRequest(t, newReadServer(t, nil, lesson), http.MethodPatch, "/v1/lessons/"+lessonIDValue, `{
		"title": "Updated Lesson"
	}`)

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if lesson.writeCalled != "update" || lesson.gotUpdateLesson.ID != lessonIDValue {
		t.Fatalf("expected update lesson input, got called=%q input=%+v", lesson.writeCalled, lesson.gotUpdateLesson)
	}
	if lesson.gotUpdateLesson.Title == nil || *lesson.gotUpdateLesson.Title != "Updated Lesson" {
		t.Fatalf("expected title pointer to be set, got %+v", lesson.gotUpdateLesson)
	}

	var out core.UpdateLessonOutput
	decodeResponse(t, response, &out)
	if out.ID != lessonIDValue {
		t.Fatalf("expected update id response, got %+v", out)
	}
}

func TestDeleteLessonRoutesByID(t *testing.T) {
	lesson := &lessonReadFake{}
	response := courseRequest(t, newReadServer(t, nil, lesson), http.MethodDelete, "/v1/lessons/"+lessonIDValue, "")

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
	}
	if lesson.writeCalled != "delete" || lesson.gotDeleteLesson.ID != lessonIDValue {
		t.Fatalf("expected delete lesson input, got called=%q input=%+v", lesson.writeCalled, lesson.gotDeleteLesson)
	}
}

func TestReorderLessonsRoutesByCourseID(t *testing.T) {
	lesson := &lessonReadFake{}
	response := courseRequest(t, newReadServer(t, nil, lesson), http.MethodPost, "/v1/courses/"+courseIDValue+"/lessons/reorder", `{
		"order": [
			{"lessonID": "`+lessonIDValue+`", "position": 1},
			{"lessonID": "`+otherLessonIDValue+`", "position": 0}
		]
	}`)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected no content response, got %d body=%s", response.Code, response.Body.String())
	}
	if lesson.writeCalled != "reorder" || lesson.gotReorderLessons.CourseID != courseIDValue {
		t.Fatalf("expected reorder lessons input, got called=%q input=%+v", lesson.writeCalled, lesson.gotReorderLessons)
	}
	if len(lesson.gotReorderLessons.Order) != 2 || lesson.gotReorderLessons.Order[0].LessonID != lessonIDValue || lesson.gotReorderLessons.Order[1].Position != 0 {
		t.Fatalf("expected lesson reorder placements, got %+v", lesson.gotReorderLessons.Order)
	}
}

func TestCourseEndpointsRejectUnsupportedMethods(t *testing.T) {
	response := readRequest(t, newReadServer(t, &courseReadFake{}, nil), http.MethodPut, "/v1/courses/"+courseIDValue)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected method not allowed, got %d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Allow") != "DELETE, GET, PATCH" {
		t.Fatalf("expected allow header for course endpoint, got %q", response.Header().Get("Allow"))
	}
}

func TestReadEndpointsRequireAuth(t *testing.T) {
	server := newReadServer(t, &courseReadFake{}, nil)
	request := httptest.NewRequest(http.MethodGet, "/v1/courses", nil)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d body=%s", response.Code, response.Body.String())
	}
}

func TestWriteEndpointsRequireAuth(t *testing.T) {
	course := &courseReadFake{}
	server := newReadServer(t, course, nil)
	request := httptest.NewRequest(http.MethodPost, "/v1/courses", strings.NewReader(`{"title":"JavaScript"}`))
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d body=%s", response.Code, response.Body.String())
	}
	if course.called != "" {
		t.Fatalf("expected unauthorized request not to reach service, got %q", course.called)
	}
}

// helpers

func newReadServer(t *testing.T, course core.CourseService, lesson core.LessonService) *Server {
	t.Helper()

	return newReadServerWithInstructor(t, course, lesson, "instructor-1")
}

func newReadServerWithInstructor(t *testing.T, course core.CourseService, lesson core.LessonService, instructorID string) *Server {
	t.Helper()

	if course == nil {
		course = courseServiceFake{}
	}
	if lesson == nil {
		lesson = &lessonServiceFake{}
	}

	server, err := NewServer(Options{
		Course:       course,
		Lesson:       lesson,
		Quiz:         &quizServiceFake{},
		Practice:     &practiceServiceFake{},
		Test:         &testServiceFake{},
		Token:        apiTokenValue,
		InstructorID: instructorID,
	})
	if err != nil {
		t.Fatalf("expected read test server, got %v", err)
	}

	return server
}

func readRequest(t *testing.T, server *Server, method string, path string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("Authorization", "Bearer "+apiTokenValue)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	return response
}

func courseRequest(t *testing.T, server *Server, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	request.Header.Set("Authorization", "Bearer "+apiTokenValue)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	return response
}

func courseViewFixture() core.CourseView {
	return core.CourseView{
		ID:           courseIDValue,
		Title:        "JavaScript",
		Slug:         "javascript",
		Description:  "Intro to JavaScript",
		InstructorID: "instructor-1",
		Status:       "published",
	}
}

func lessonViewFixture() core.LessonView {
	return core.LessonView{
		ID:       lessonIDValue,
		CourseID: courseIDValue,
		Title:    "Variables and Data Types",
		Order:    1,
	}
}

// courseReadFake records read calls and serves canned course responses. It embeds
// courseServiceFake to satisfy the rest of the CourseService interface.
type courseReadFake struct {
	courseServiceFake

	called             string
	err                error
	gotCreateCourse    core.CreateCourseInput
	createOut          core.CreateCourseOutput
	gotListCourses     core.ListCoursesInput
	listOut            core.ListCoursesOutput
	gotGetCourse       core.GetCourseInput
	getOut             core.GetCourseOutput
	gotUpdateCourse    core.UpdateCourseInput
	updateOut          core.UpdateCourseOutput
	gotDeleteCourse    core.DeleteCourseInput
	gotPublishCourse   core.PublishCourseInput
	gotUnpublishCourse core.UnpublishCourseInput
}

func (fake *courseReadFake) CreateCourse(in core.CreateCourseInput) (core.CreateCourseOutput, error) {
	fake.called = "create"
	fake.gotCreateCourse = in
	if fake.err != nil {
		return core.CreateCourseOutput{}, fake.err
	}
	return fake.createOut, nil
}

func (fake *courseReadFake) ListCourses(in core.ListCoursesInput) (core.ListCoursesOutput, error) {
	fake.called = "list"
	fake.gotListCourses = in
	if fake.err != nil {
		return core.ListCoursesOutput{}, fake.err
	}
	return fake.listOut, nil
}

func (fake *courseReadFake) GetCourse(in core.GetCourseInput) (core.GetCourseOutput, error) {
	fake.called = "get"
	fake.gotGetCourse = in
	if fake.err != nil {
		return core.GetCourseOutput{}, fake.err
	}
	return fake.getOut, nil
}

func (fake *courseReadFake) UpdateCourse(in core.UpdateCourseInput) (core.UpdateCourseOutput, error) {
	fake.called = "update"
	fake.gotUpdateCourse = in
	if fake.err != nil {
		return core.UpdateCourseOutput{}, fake.err
	}
	return fake.updateOut, nil
}

func (fake *courseReadFake) DeleteCourse(in core.DeleteCourseInput) error {
	fake.called = "delete"
	fake.gotDeleteCourse = in
	return fake.err
}

func (fake *courseReadFake) PublishCourse(in core.PublishCourseInput) error {
	fake.called = "publish"
	fake.gotPublishCourse = in
	return fake.err
}

func (fake *courseReadFake) UnpublishCourse(in core.UnpublishCourseInput) error {
	fake.called = "unpublish"
	fake.gotUnpublishCourse = in
	return fake.err
}

// lessonReadFake records lesson read calls. It embeds lessonServiceFake to satisfy
// the block-related parts of the LessonService interface; field names are kept
// distinct from the embedded fake to avoid shadowing confusion.
type lessonReadFake struct {
	lessonServiceFake

	readCalled        string
	readErr           error
	writeCalled       string
	writeErr          error
	gotCreateLesson   core.CreateLessonInput
	createOut         core.CreateLessonOutput
	gotListLessons    core.ListLessonsInput
	lessonsOut        core.ListLessonsOutput
	gotGetLesson      core.GetLessonInput
	lessonOut         core.GetLessonOutput
	gotUpdateLesson   core.UpdateLessonInput
	updateOut         core.UpdateLessonOutput
	gotDeleteLesson   core.DeleteLessonInput
	gotReorderLessons core.ReorderLessonsInput
}

func (fake *lessonReadFake) CreateLesson(in core.CreateLessonInput) (core.CreateLessonOutput, error) {
	fake.writeCalled = "create"
	fake.gotCreateLesson = in
	if fake.writeErr != nil {
		return core.CreateLessonOutput{}, fake.writeErr
	}
	return fake.createOut, nil
}

func (fake *lessonReadFake) ListLessons(in core.ListLessonsInput) (core.ListLessonsOutput, error) {
	fake.readCalled = "list"
	fake.gotListLessons = in
	if fake.readErr != nil {
		return core.ListLessonsOutput{}, fake.readErr
	}
	return fake.lessonsOut, nil
}

func (fake *lessonReadFake) GetLesson(in core.GetLessonInput) (core.GetLessonOutput, error) {
	fake.readCalled = "get"
	fake.gotGetLesson = in
	if fake.readErr != nil {
		return core.GetLessonOutput{}, fake.readErr
	}
	return fake.lessonOut, nil
}

func (fake *lessonReadFake) UpdateLesson(in core.UpdateLessonInput) (core.UpdateLessonOutput, error) {
	fake.writeCalled = "update"
	fake.gotUpdateLesson = in
	if fake.writeErr != nil {
		return core.UpdateLessonOutput{}, fake.writeErr
	}
	return fake.updateOut, nil
}

func (fake *lessonReadFake) DeleteLesson(in core.DeleteLessonInput) error {
	fake.writeCalled = "delete"
	fake.gotDeleteLesson = in
	return fake.writeErr
}

func (fake *lessonReadFake) ReorderLessons(in core.ReorderLessonsInput) error {
	fake.writeCalled = "reorder"
	fake.gotReorderLessons = in
	return fake.writeErr
}
