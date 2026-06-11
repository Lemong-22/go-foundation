package rest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const importInstructorIDValue = "550e8400-e29b-41d4-a716-446655440010"

func TestImportPlanRouteMapsMultipartZipToService(t *testing.T) {
	service := &importServiceFake{
		planOut: core.PlanImportOutput{
			Plan: restImportPlanFixture(t),
		},
	}

	response := authedImportMultipartRequest(t, service, http.MethodPost, "/v1/import/plan", importMultipartRequest{
		zip: []byte("zip-bytes"),
	})

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "plan" {
		t.Fatalf("expected plan call, got %q", service.called)
	}
	if service.planIn.InstructorID != importInstructorIDValue {
		t.Fatalf("expected instructor id, got %+v", service.planIn)
	}
	if service.planZipContent != "zip-bytes" {
		t.Fatalf("expected uploaded zip bytes during service call, got %q", service.planZipContent)
	}
	if _, err := os.Stat(service.planIn.ZipPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected temp zip to be deleted after request, got %v", err)
	}

	var out map[string]any
	decodeResponse(t, response, &out)
	if out["format_version"] != "1" || out["zip_hash"] == "" {
		t.Fatalf("expected import plan json response, got %+v", out)
	}
}

func TestImportApplyRouteMapsMultipartZipPlanAndStrategyToService(t *testing.T) {
	service := &importServiceFake{
		applyOut: core.ApplyPlanOutput{
			Result: restApplyResultFixture(t, nil),
		},
	}

	response := authedImportMultipartRequest(t, service, http.MethodPost, "/v1/import/apply?conflict_strategy=update", importMultipartRequest{
		zip:                []byte("zip-bytes"),
		resolvedPlanJSON:   []byte(`{"format_version":"1"}`),
		resolvedPlanAsFile: true,
	})

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "apply" {
		t.Fatalf("expected apply call, got %q", service.called)
	}
	if service.applyIn.InstructorID != importInstructorIDValue || service.applyIn.ConflictStrategy != "update" {
		t.Fatalf("expected instructor and strategy in input, got %+v", service.applyIn)
	}
	if string(service.applyIn.ResolvedPlanJSON) != `{"format_version":"1"}` {
		t.Fatalf("expected resolved plan bytes, got %s", string(service.applyIn.ResolvedPlanJSON))
	}
	if service.applyZipContent != "zip-bytes" {
		t.Fatalf("expected uploaded zip bytes during service call, got %q", service.applyZipContent)
	}
	if _, err := os.Stat(service.applyIn.ZipPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected temp zip to be deleted after request, got %v", err)
	}

	var out applyResultResponse
	decodeResponse(t, response, &out)
	if out.AggregatesSucceeded != 1 || out.AggregatesFailed != 0 {
		t.Fatalf("expected apply result response, got %+v", out)
	}
}

func TestImportApplyAcceptsResolvedPlanFormField(t *testing.T) {
	service := &importServiceFake{
		applyOut: core.ApplyPlanOutput{
			Result: restApplyResultFixture(t, nil),
		},
	}

	response := authedImportMultipartRequest(t, service, http.MethodPost, "/v1/import/apply", importMultipartRequest{
		zip:              []byte("zip-bytes"),
		resolvedPlanJSON: []byte(`{"zip_hash":"abc"}`),
	})

	if response.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d body=%s", response.Code, response.Body.String())
	}
	if string(service.applyIn.ResolvedPlanJSON) != `{"zip_hash":"abc"}` {
		t.Fatalf("expected resolved plan form field bytes, got %s", string(service.applyIn.ResolvedPlanJSON))
	}
}

func TestImportApplyReturnsPartialFailureDetails(t *testing.T) {
	failed := restFailedOperationFixture(t)
	service := &importServiceFake{
		applyOut: core.ApplyPlanOutput{
			Result: restApplyResultFixture(t, []domain.FailedOperation{failed}),
		},
	}

	response := authedImportMultipartRequest(t, service, http.MethodPost, "/v1/import/apply", importMultipartRequest{
		zip: []byte("zip-bytes"),
	})

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected internal server error response, got %d body=%s", response.Code, response.Body.String())
	}

	var out applyResultResponse
	decodeResponse(t, response, &out)
	if out.AggregatesFailed != 1 || len(out.Failed) != 1 || out.Failed[0].Error != "course create failed" {
		t.Fatalf("expected failed operation details, got %+v", out)
	}
}

func TestImportApplyMapsUnresolvedConflictsAndCleansTempfile(t *testing.T) {
	service := &importServiceFake{
		applyErr: domain.NewUnresolvedImportConflictsError([]domain.ImportConflict{restImportConflictFixture(t)}),
	}

	response := authedImportMultipartRequest(t, service, http.MethodPost, "/v1/import/apply", importMultipartRequest{
		zip: []byte("zip-bytes"),
	})

	if response.Code != http.StatusConflict {
		t.Fatalf("expected conflict response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "apply" {
		t.Fatalf("expected apply call, got %q", service.called)
	}
	if _, err := os.Stat(service.applyIn.ZipPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected temp zip to be deleted after apply error, got %v", err)
	}
}

func TestImportEndpointsMapImportErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "parse", err: domain.NewImportSourceParseError("course.zip", "invalid yaml", nil), want: http.StatusBadRequest},
		{name: "layout", err: domain.NewImportSourceLayoutError("course.zip", "missing course.yaml", nil), want: http.StatusBadRequest},
		{name: "unsupported format", err: domain.NewUnsupportedImportFormatError("2", []string{"1"}), want: http.StatusBadRequest},
		{name: "invalid strategy", err: domain.NewInvalidConflictStrategyError("merge"), want: http.StatusBadRequest},
		{name: "hash mismatch", err: domain.NewImportPlanHashMismatchError(strings.Repeat("a", 64), strings.Repeat("b", 64)), want: http.StatusBadRequest},
		{name: "unresolved conflicts", err: domain.NewUnresolvedImportConflictsError([]domain.ImportConflict{restImportConflictFixture(t)}), want: http.StatusConflict},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &importServiceFake{planErr: test.err}
			response := authedImportMultipartRequest(t, service, http.MethodPost, "/v1/import/plan", importMultipartRequest{
				zip: []byte("zip-bytes"),
			})

			if response.Code != test.want {
				t.Fatalf("expected status %d, got %d body=%s", test.want, response.Code, response.Body.String())
			}
		})
	}
}

func TestImportPlanRejectsMissingZipBeforeServiceCall(t *testing.T) {
	service := &importServiceFake{}
	response := authedImportMultipartRequest(t, service, http.MethodPost, "/v1/import/plan", importMultipartRequest{})

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "" {
		t.Fatalf("expected missing zip not to reach service, got %q", service.called)
	}
}

func TestImportPlanRequiresConfiguredInstructorID(t *testing.T) {
	service := &importServiceFake{}
	server, err := NewServer(Options{
		Course:   courseServiceFake{},
		Lesson:   &lessonServiceFake{},
		Quiz:     &quizServiceFake{},
		Practice: &practiceServiceFake{},
		Test:     &testServiceFake{},
		Import:   service,
		Token:    apiTokenValue,
	})
	if err != nil {
		t.Fatalf("expected test server, got %v", err)
	}

	request, contentType := newImportMultipartHTTPRequest(t, http.MethodPost, "/v1/import/plan", importMultipartRequest{
		zip: []byte("zip-bytes"),
	})
	request.Header.Set("Authorization", "Bearer "+apiTokenValue)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request response, got %d body=%s", response.Code, response.Body.String())
	}
	if service.called != "" {
		t.Fatalf("expected missing instructor id not to reach service, got %q", service.called)
	}
}

type importMultipartRequest struct {
	zip                []byte
	resolvedPlanJSON   []byte
	resolvedPlanAsFile bool
}

func authedImportMultipartRequest(t *testing.T, service *importServiceFake, method string, path string, multipartRequest importMultipartRequest) *httptest.ResponseRecorder {
	t.Helper()

	server := newImportTestServer(t, service)
	request, contentType := newImportMultipartHTTPRequest(t, method, path, multipartRequest)
	request.Header.Set("Authorization", "Bearer "+apiTokenValue)
	request.Header.Set("Content-Type", contentType)
	response := httptest.NewRecorder()

	server.Handler().ServeHTTP(response, request)
	return response
}

func newImportMultipartHTTPRequest(t *testing.T, method string, path string, multipartRequest importMultipartRequest) (*http.Request, string) {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	if multipartRequest.zip != nil {
		writeMultipartFile(t, writer, "zip", "course.zip", multipartRequest.zip)
	}
	if multipartRequest.resolvedPlanJSON != nil {
		if multipartRequest.resolvedPlanAsFile {
			writeMultipartFile(t, writer, "resolved_plan", "plan.json", multipartRequest.resolvedPlanJSON)
		} else if err := writer.WriteField("resolved_plan", string(multipartRequest.resolvedPlanJSON)); err != nil {
			t.Fatalf("expected resolved plan field, got %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("expected multipart body to close, got %v", err)
	}

	request := httptest.NewRequest(method, path, body)
	return request, writer.FormDataContentType()
}

func writeMultipartFile(t *testing.T, writer *multipart.Writer, field string, filename string, content []byte) {
	t.Helper()

	part, err := writer.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("expected multipart file part, got %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("expected multipart file content, got %v", err)
	}
}

func newImportTestServer(t *testing.T, service core.ImportService) *Server {
	t.Helper()

	server, err := NewServer(Options{
		Course:       courseServiceFake{},
		Lesson:       &lessonServiceFake{},
		Quiz:         &quizServiceFake{},
		Practice:     &practiceServiceFake{},
		Test:         &testServiceFake{},
		Import:       service,
		Token:        apiTokenValue,
		InstructorID: importInstructorIDValue,
	})
	if err != nil {
		t.Fatalf("expected test server, got %v", err)
	}

	return server
}

type importServiceFake struct {
	called string

	planIn         core.PlanImportInput
	planOut        core.PlanImportOutput
	planErr        error
	planZipContent string

	applyIn         core.ApplyPlanInput
	applyOut        core.ApplyPlanOutput
	applyErr        error
	applyZipContent string
}

func (service *importServiceFake) PlanImport(in core.PlanImportInput) (core.PlanImportOutput, error) {
	service.called = "plan"
	service.planIn = in
	service.planZipContent = readImportZipForTest(in.ZipPath)
	if service.planErr != nil {
		return core.PlanImportOutput{}, service.planErr
	}

	return service.planOut, nil
}

func (service *importServiceFake) ApplyPlan(in core.ApplyPlanInput) (core.ApplyPlanOutput, error) {
	service.called = "apply"
	service.applyIn = in
	service.applyZipContent = readImportZipForTest(in.ZipPath)
	if service.applyErr != nil {
		return core.ApplyPlanOutput{}, service.applyErr
	}

	return service.applyOut, nil
}

func readImportZipForTest(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return ""
	}

	return string(data)
}

func restImportPlanFixture(t *testing.T) domain.ImportPlan {
	t.Helper()

	plan, err := domain.NewImportPlan("1", strings.Repeat("a", 64), time.Unix(1, 0).UTC(), nil, nil)
	if err != nil {
		t.Fatalf("expected import plan fixture, got %v", err)
	}

	return plan
}

func restApplyResultFixture(t *testing.T, failed []domain.FailedOperation) domain.ApplyResult {
	t.Helper()

	aggregatesSucceeded := 1
	aggregatesFailed := 0
	if len(failed) > 0 {
		aggregatesSucceeded = 0
		aggregatesFailed = 1
	}
	result, err := domain.NewApplyResult(nil, failed, nil, aggregatesSucceeded, aggregatesFailed)
	if err != nil {
		t.Fatalf("expected apply result fixture, got %v", err)
	}

	return result
}

func restFailedOperationFixture(t *testing.T) domain.FailedOperation {
	t.Helper()

	failed, err := domain.NewFailedOperation(restImportOperationFixture(t), errors.New("course create failed"))
	if err != nil {
		t.Fatalf("expected failed operation fixture, got %v", err)
	}

	return failed
}

func restImportOperationFixture(t *testing.T) domain.ImportOperation {
	t.Helper()

	operation, err := domain.NewImportOperation(
		domain.CreateOperation(),
		domain.CourseEntity(),
		"course:intro",
		nil,
		json.RawMessage(`{"title":"Intro"}`),
	)
	if err != nil {
		t.Fatalf("expected import operation fixture, got %v", err)
	}

	return operation
}

func restImportConflictFixture(t *testing.T) domain.ImportConflict {
	t.Helper()

	candidate, err := domain.NewConflictCandidate(courseIDValue, "Existing course")
	if err != nil {
		t.Fatalf("expected conflict candidate fixture, got %v", err)
	}
	conflict, err := domain.NewImportConflict(
		domain.CourseEntity(),
		"course:intro",
		domain.SlugCollision(),
		[]domain.ConflictCandidate{candidate},
		domain.SkipOperation(),
		json.RawMessage(`{"title":"Intro"}`),
	)
	if err != nil {
		t.Fatalf("expected import conflict fixture, got %v", err)
	}

	return conflict
}
