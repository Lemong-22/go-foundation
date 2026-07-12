package course

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// --- Stub in-memory repo untuk test (tidak butuh Postgres) ---

type stubRepo struct {
	courses map[string]*Course
	lessons []Lesson
}

func newStubRepo() *stubRepo {
	return &stubRepo{
		courses: make(map[string]*Course),
		lessons: make([]Lesson, 0),
	}
}

func (s *stubRepo) Save(_ context.Context, c *Course) error {
	s.courses[c.ID] = c
	return nil
}

func (s *stubRepo) FindByID(_ context.Context, id string) (*Course, error) {
	c, ok := s.courses[id]
	if !ok {
		return nil, ErrNotFound
	}
	return c, nil
}

func (s *stubRepo) FindAll(_ context.Context) ([]Course, error) {
	result := make([]Course, 0, len(s.courses))
	for _, c := range s.courses {
		result = append(result, *c)
	}
	return result, nil
}

func (s *stubRepo) Delete(_ context.Context, id string) error {
	if _, ok := s.courses[id]; !ok {
		return ErrNotFound
	}
	delete(s.courses, id)
	return nil
}

func (s *stubRepo) SaveLesson(_ context.Context, l *Lesson) error {
	s.lessons = append(s.lessons, *l)
	return nil
}

func (s *stubRepo) FindLessonsByCourseID(_ context.Context, courseID string) ([]Lesson, error) {
	result := make([]Lesson, 0)
	for _, l := range s.lessons {
		if l.CourseID == courseID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (s *stubRepo) DeleteLesson(_ context.Context, id string) error {
	for i, l := range s.lessons {
		if l.ID == id {
			s.lessons = append(s.lessons[:i], s.lessons[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

// --- Helper ---

func newRequest(method, target string, body []byte) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, target, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, target, nil)
	}
	return req
}

// --- Tests: Course handlers ---

func TestCreateCourseHandler_Success(t *testing.T) {
	repo := newStubRepo()
	h := NewCourseHandler(repo)

	body, _ := json.Marshal(map[string]string{
		"title": "Intro to Go",
		"slug":  "intro-go",
	})
	req := newRequest(http.MethodPost, "/courses", body)
	rr := httptest.NewRecorder()

	h.CreateCourseHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var got Course
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Title != "Intro to Go" {
		t.Errorf("expected title 'Intro to Go', got %q", got.Title)
	}
	if got.Slug != "intro-go" {
		t.Errorf("expected slug 'intro-go', got %q", got.Slug)
	}
	if got.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", got.Status)
	}
}

func TestCreateCourseHandler_MissingFields(t *testing.T) {
	repo := newStubRepo()
	h := NewCourseHandler(repo)

	body, _ := json.Marshal(map[string]string{"title": "No Slug"})
	req := newRequest(http.MethodPost, "/courses", body)
	rr := httptest.NewRecorder()

	h.CreateCourseHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestListCoursesHandler_Empty(t *testing.T) {
	repo := newStubRepo()
	h := NewCourseHandler(repo)

	req := newRequest(http.MethodGet, "/courses", nil)
	rr := httptest.NewRecorder()

	h.ListCoursesHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var got []Course
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d courses", len(got))
	}
}

// --- Tests: Lesson handlers ---

// seedCourse menyiapkan satu course di stub repo dan mengembalikan ID-nya.
func seedCourse(t *testing.T, repo *stubRepo) string {
	t.Helper()
	c := &Course{
		ID:    "CRSM-TEST-001",
		Title: "Test Course",
		Slug:  "test-course",
	}
	if err := repo.Save(context.Background(), c); err != nil {
		t.Fatalf("seedCourse: %v", err)
	}
	return c.ID
}

func TestCreateLessonHandler_Success(t *testing.T) {
	repo := newStubRepo()
	h := NewCourseHandler(repo)
	courseID := seedCourse(t, repo)

	body, _ := json.Marshal(map[string]any{
		"title":       "Bab 1",
		"slug":        "bab-1",
		"content":     "Materi pertama",
		"order_index": 0,
	})

	// Pakai httptest.NewRequest dengan path value manual
	req := httptest.NewRequest(http.MethodPost, "/courses/"+courseID+"/lessons", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Set path value supaya r.PathValue("id") bekerja
	req.SetPathValue("id", courseID)
	rr := httptest.NewRecorder()

	h.CreateLessonHandler(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var got Lesson
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if got.Title != "Bab 1" {
		t.Errorf("expected title 'Bab 1', got %q", got.Title)
	}
	if got.CourseID != courseID {
		t.Errorf("expected course_id %q, got %q", courseID, got.CourseID)
	}
}

func TestCreateLessonHandler_CourseNotFound(t *testing.T) {
	repo := newStubRepo()
	h := NewCourseHandler(repo)

	body, _ := json.Marshal(map[string]any{
		"title": "Bab 1",
		"slug":  "bab-1",
	})
	req := httptest.NewRequest(http.MethodPost, "/courses/CRSM-BOGUS/lessons", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", "CRSM-BOGUS")
	rr := httptest.NewRecorder()

	h.CreateLessonHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestCreateLessonHandler_MissingFields(t *testing.T) {
	repo := newStubRepo()
	h := NewCourseHandler(repo)
	courseID := seedCourse(t, repo)

	body, _ := json.Marshal(map[string]any{"title": "No Slug"})
	req := httptest.NewRequest(http.MethodPost, "/courses/"+courseID+"/lessons", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.SetPathValue("id", courseID)
	rr := httptest.NewRecorder()

	h.CreateLessonHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestListLessonsHandler_Success(t *testing.T) {
	repo := newStubRepo()
	h := NewCourseHandler(repo)
	courseID := seedCourse(t, repo)

	// Seed 2 lessons langsung ke repo
	_ = repo.SaveLesson(context.Background(), &Lesson{
		ID: "LES-001", CourseID: courseID, Title: "Bab 1", Slug: "bab-1", OrderIndex: 0,
	})
	_ = repo.SaveLesson(context.Background(), &Lesson{
		ID: "LES-002", CourseID: courseID, Title: "Bab 2", Slug: "bab-2", OrderIndex: 1,
	})

	req := httptest.NewRequest(http.MethodGet, "/courses/"+courseID+"/lessons", nil)
	req.SetPathValue("id", courseID)
	rr := httptest.NewRecorder()

	h.ListLessonsHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", rr.Code, rr.Body.String())
	}

	var got []Lesson
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 lessons, got %d", len(got))
	}
}

func TestListLessonsHandler_CourseNotFound(t *testing.T) {
	repo := newStubRepo()
	h := NewCourseHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/courses/CRSM-BOGUS/lessons", nil)
	req.SetPathValue("id", "CRSM-BOGUS")
	rr := httptest.NewRecorder()

	h.ListLessonsHandler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
