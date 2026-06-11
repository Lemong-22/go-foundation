package rest

import (
	"net/http"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

// handleCourses serves the student/instructor course catalog and course
// creation endpoint. An optional ?status= query filter is forwarded on reads.
func (server *Server) handleCourses(response http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		out, err := server.course.ListCourses(core.ListCoursesInput{
			Status: request.URL.Query().Get("status"),
		})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPost:
		if server.instructorID == "" {
			writeError(response, domain.NewValidationError("instructor_id", "is required"))
			return
		}

		var input core.CreateCourseInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.InstructorID = server.instructorID

		out, err := server.course.CreateCourse(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusCreated, out)
	default:
		methodNotAllowed(response, http.MethodGet, http.MethodPost)
	}
}

// handleCourse serves GET/PATCH/DELETE /v1/courses/{id}.
func (server *Server) handleCourse(response http.ResponseWriter, request *http.Request, courseID string) {
	switch request.Method {
	case http.MethodGet:
		out, err := server.course.GetCourse(core.GetCourseInput{ID: courseID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPatch:
		var input core.UpdateCourseInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.ID = courseID

		out, err := server.course.UpdateCourse(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodDelete:
		if err := server.course.DeleteCourse(core.DeleteCourseInput{ID: courseID}); err != nil {
			writeError(response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodDelete, http.MethodGet, http.MethodPatch)
	}
}

// handleCourseLessons serves GET/POST /v1/courses/{id}/lessons.
func (server *Server) handleCourseLessons(response http.ResponseWriter, request *http.Request, courseID string) {
	switch request.Method {
	case http.MethodGet:
		out, err := server.lesson.ListLessons(core.ListLessonsInput{CourseID: courseID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPost:
		var input core.CreateLessonInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.CourseID = courseID

		out, err := server.lesson.CreateLesson(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusCreated, out)
	default:
		methodNotAllowed(response, http.MethodGet, http.MethodPost)
	}
}

// handleLesson serves GET/PATCH/DELETE /v1/lessons/{id}. The lesson's content
// blocks are served separately by /v1/lessons/{id}/blocks.
func (server *Server) handleLesson(response http.ResponseWriter, request *http.Request, lessonID string) {
	switch request.Method {
	case http.MethodGet:
		out, err := server.lesson.GetLesson(core.GetLessonInput{ID: lessonID})
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodPatch:
		var input core.UpdateLessonInput
		if err := decodeJSON(request, &input); err != nil {
			writeError(response, err)
			return
		}
		input.ID = lessonID

		out, err := server.lesson.UpdateLesson(input)
		if err != nil {
			writeError(response, err)
			return
		}
		writeJSON(response, http.StatusOK, out)
	case http.MethodDelete:
		if err := server.lesson.DeleteLesson(core.DeleteLessonInput{ID: lessonID}); err != nil {
			writeError(response, err)
			return
		}
		response.WriteHeader(http.StatusNoContent)
	default:
		methodNotAllowed(response, http.MethodDelete, http.MethodGet, http.MethodPatch)
	}
}

func (server *Server) handleCoursePublish(response http.ResponseWriter, request *http.Request, courseID string) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	if err := server.course.PublishCourse(core.PublishCourseInput{ID: courseID}); err != nil {
		writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (server *Server) handleCourseUnpublish(response http.ResponseWriter, request *http.Request, courseID string) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	if err := server.course.UnpublishCourse(core.UnpublishCourseInput{ID: courseID}); err != nil {
		writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}

func (server *Server) handleCourseLessonsReorder(response http.ResponseWriter, request *http.Request, courseID string) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	var input core.ReorderLessonsInput
	if err := decodeJSON(request, &input); err != nil {
		writeError(response, err)
		return
	}
	input.CourseID = courseID

	if err := server.lesson.ReorderLessons(input); err != nil {
		writeError(response, err)
		return
	}
	response.WriteHeader(http.StatusNoContent)
}
