package rest

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/luxeave/entropy-course/internal/course/core"
	"github.com/luxeave/entropy-course/internal/course/domain"
)

const importMultipartMemory = 32 << 20

type applyResultResponse struct {
	Applied             []appliedOperationResponse `json:"applied"`
	Failed              []failedOperationResponse  `json:"failed"`
	Skipped             []importOperationResponse  `json:"skipped"`
	Counts              map[string]int             `json:"counts"`
	AggregatesSucceeded int                        `json:"aggregates_succeeded"`
	AggregatesFailed    int                        `json:"aggregates_failed"`
}

type appliedOperationResponse struct {
	importOperationResponse
	Message string `json:"message"`
}

type failedOperationResponse struct {
	importOperationResponse
	Error string `json:"error"`
}

type importOperationResponse struct {
	Kind       string `json:"kind"`
	EntityType string `json:"entity_type"`
	EntityRef  string `json:"entity_ref"`
	TargetID   string `json:"target_id"`
}

func (server *Server) handleImportPlan(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	if err := parseImportMultipart(request); err != nil {
		writeError(response, err)
		return
	}
	defer cleanupMultipartForm(request)

	zipPath, cleanup, err := importZipTempfile(request)
	if err != nil {
		writeError(response, err)
		return
	}
	defer cleanup()

	service, err := server.importService()
	if err != nil {
		writeError(response, err)
		return
	}

	out, err := service.PlanImport(core.PlanImportInput{
		ZipPath:      zipPath,
		InstructorID: server.instructorID,
	})
	if err != nil {
		writeError(response, err)
		return
	}

	writeJSON(response, http.StatusOK, out.Plan)
}

func (server *Server) handleImportApply(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		methodNotAllowed(response, http.MethodPost)
		return
	}

	if err := parseImportMultipart(request); err != nil {
		writeError(response, err)
		return
	}
	defer cleanupMultipartForm(request)

	zipPath, cleanup, err := importZipTempfile(request)
	if err != nil {
		writeError(response, err)
		return
	}
	defer cleanup()

	resolvedPlan, err := optionalMultipartJSON(request, "resolved_plan")
	if err != nil {
		writeError(response, err)
		return
	}

	service, err := server.importService()
	if err != nil {
		writeError(response, err)
		return
	}

	out, err := service.ApplyPlan(core.ApplyPlanInput{
		ZipPath:          zipPath,
		InstructorID:     server.instructorID,
		ResolvedPlanJSON: resolvedPlan,
		ConflictStrategy: request.URL.Query().Get("conflict_strategy"),
	})
	if err != nil {
		writeError(response, err)
		return
	}

	status := http.StatusOK
	if out.Result.AggregatesFailed() > 0 || len(out.Result.Failed()) > 0 {
		status = http.StatusInternalServerError
	}
	writeJSON(response, status, applyResultOutput(out.Result))
}

func (server *Server) importService() (core.ImportService, error) {
	if server.imports == nil {
		return nil, ErrMissingImportService
	}
	if server.instructorID == "" {
		return nil, domain.NewValidationError("instructor_id", "is required")
	}

	return server.imports, nil
}

func parseImportMultipart(request *http.Request) error {
	if request.MultipartForm != nil {
		return nil
	}
	if err := request.ParseMultipartForm(importMultipartMemory); err != nil {
		return domain.NewValidationError("request", "must be multipart form data")
	}

	return nil
}

func cleanupMultipartForm(request *http.Request) {
	if request.MultipartForm != nil {
		_ = request.MultipartForm.RemoveAll()
	}
}

func importZipTempfile(request *http.Request) (string, func(), error) {
	file, header, err := request.FormFile("zip")
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", func() {}, domain.NewValidationError("zip", "is required")
		}
		return "", func() {}, err
	}
	defer file.Close()

	temp, err := os.CreateTemp("", "course-import-*.zip")
	if err != nil {
		return "", func() {}, err
	}

	path := temp.Name()
	cleanup := func() { _ = os.Remove(path) }
	if err := copyUpload(temp, file, header); err != nil {
		_ = temp.Close()
		cleanup()
		return "", func() {}, err
	}
	if err := temp.Close(); err != nil {
		cleanup()
		return "", func() {}, err
	}

	return path, cleanup, nil
}

func copyUpload(dst *os.File, src multipart.File, header *multipart.FileHeader) error {
	if header == nil || header.Filename == "" {
		return domain.NewValidationError("zip", "filename is required")
	}
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

func optionalMultipartJSON(request *http.Request, field string) ([]byte, error) {
	file, _, err := request.FormFile(field)
	if err == nil {
		defer file.Close()
		return io.ReadAll(file)
	}
	if !errors.Is(err, http.ErrMissingFile) {
		return nil, err
	}
	if request.MultipartForm == nil {
		return nil, nil
	}
	values := request.MultipartForm.Value[field]
	if len(values) == 0 {
		return nil, nil
	}

	return []byte(values[0]), nil
}

func applyResultOutput(result domain.ApplyResult) applyResultResponse {
	return applyResultResponse{
		Applied:             appliedOperationOutputs(result.Applied()),
		Failed:              failedOperationOutputs(result.Failed()),
		Skipped:             importOperationOutputs(result.Skipped()),
		Counts:              applyResultCounts(result),
		AggregatesSucceeded: result.AggregatesSucceeded(),
		AggregatesFailed:    result.AggregatesFailed(),
	}
}

func appliedOperationOutputs(operations []domain.AppliedOperation) []appliedOperationResponse {
	outputs := make([]appliedOperationResponse, 0, len(operations))
	for _, operation := range operations {
		outputs = append(outputs, appliedOperationResponse{
			importOperationResponse: importOperationOutput(operation.Operation()),
			Message:                 operation.Message(),
		})
	}

	return outputs
}

func failedOperationOutputs(operations []domain.FailedOperation) []failedOperationResponse {
	outputs := make([]failedOperationResponse, 0, len(operations))
	for _, operation := range operations {
		outputs = append(outputs, failedOperationResponse{
			importOperationResponse: importOperationOutput(operation.Operation()),
			Error:                   operation.Err().Error(),
		})
	}

	return outputs
}

func importOperationOutputs(operations []domain.ImportOperation) []importOperationResponse {
	outputs := make([]importOperationResponse, 0, len(operations))
	for _, operation := range operations {
		outputs = append(outputs, importOperationOutput(operation))
	}

	return outputs
}

func importOperationOutput(operation domain.ImportOperation) importOperationResponse {
	return importOperationResponse{
		Kind:       operation.Kind().String(),
		EntityType: operation.EntityType().String(),
		EntityRef:  operation.EntityRef(),
		TargetID:   importTargetID(operation),
	}
}

func importTargetID(operation domain.ImportOperation) string {
	targetID := operation.TargetID()
	if targetID == nil {
		return ""
	}

	return *targetID
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
