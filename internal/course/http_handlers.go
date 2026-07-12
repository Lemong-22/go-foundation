package course

import (
	"encoding/json"
	"net/http"
	"time"
)

// CourseHandler memegang repository database agar bisa transaksi data via HTTP
type CourseHandler struct {
	repo CourseRepository
}

func NewCourseHandler(repo CourseRepository) *CourseHandler {
	return &CourseHandler{repo: repo}
}

// CreateCourseHandler menangani rute "POST /courses"
func (h *CourseHandler) CreateCourseHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Validasi method wajib POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Bikin DTO penampung JSON dari user
	var input struct {
		Title       string `json:"title"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}

	// 3. Bongkar isi JSON dari user (Unmarshaling)
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validasi fail-fast basic
	if input.Title == "" || input.Slug == "" {
		http.Error(w, "Title and slug are required", http.StatusBadRequest)
		return
	}

	// 4. Masukkan data ke Entity Course asli
	// ID generator taktis nanoseconds mirip buatan VIN di CLI kemarin
	newCourse := Course{
		ID:          "CRSM-" + time.Now().Format("20060102150405"), // format waktu unik
		Title:       input.Title,
		Slug:        input.Slug,
		Description: input.Description,
		Status:      "draft",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 5. Simpan ke database Postgres pakai repo yang sudah kita bikin di Day 4!
	if err := h.repo.Save(r.Context(), &newCourse); err != nil {
		http.Error(w, "Failed to save course: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Kirim balik response sukses dalam bentuk JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCourse)
}

// ListCoursesHandler menangani rute "GET /courses"
func (h *CourseHandler) ListCoursesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	courses, err := h.repo.FindAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch courses", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(courses)
}

// CreateLessonHandler menangani rute "POST /courses/{id}/lessons"
func (h *CourseHandler) CreateLessonHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Ambil course ID dari path value (Go 1.22 native routing)
	courseID := r.PathValue("id")
	if courseID == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

	// 2. Validasi parent course ada dulu (fail-fast sebelum hit FK)
	if _, err := h.repo.FindByID(r.Context(), courseID); err != nil {
		if err == ErrNotFound {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify course: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. DTO input dari JSON body
	var input struct {
		Title      string `json:"title"`
		Slug       string `json:"slug"`
		Content    string `json:"content"`
		OrderIndex int    `json:"order_index"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if input.Title == "" || input.Slug == "" {
		http.Error(w, "Title and slug are required", http.StatusBadRequest)
		return
	}

	// 4. Bangun entity Lesson
	now := time.Now()
	newLesson := Lesson{
		ID:         "LES-" + now.Format("20060102150405.000000000"),
		CourseID:   courseID,
		Title:      input.Title,
		Slug:       input.Slug,
		Content:    input.Content,
		OrderIndex: input.OrderIndex,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// 5. Simpan ke Postgres
	if err := h.repo.SaveLesson(r.Context(), &newLesson); err != nil {
		http.Error(w, "Failed to save lesson: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 6. Response 201 Created
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newLesson)
}

// ListLessonsHandler menangani rute "GET /courses/{id}/lessons"
func (h *CourseHandler) ListLessonsHandler(w http.ResponseWriter, r *http.Request) {
	courseID := r.PathValue("id")
	if courseID == "" {
		http.Error(w, "Course ID is required", http.StatusBadRequest)
		return
	}

	// Validasi parent course ada
	if _, err := h.repo.FindByID(r.Context(), courseID); err != nil {
		if err == ErrNotFound {
			http.Error(w, "Course not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to verify course: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lessons, err := h.repo.FindLessonsByCourseID(r.Context(), courseID)
	if err != nil {
		http.Error(w, "Failed to fetch lessons", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(lessons)
}
