package cli

import (
	"errors"

	"github.com/luxeave/entropy-course/internal/course/domain"
)

const (
	ExitOK         = 0
	ExitValidation = 1
	ExitNotFound   = 2
	ExitPermission = 3
	ExitInternal   = 5
)

var ErrPermissionDenied = errors.New("permission denied")

func ExitCode(err error) int {
	if err == nil {
		return ExitOK
	}

	if isValidationError(err) {
		return ExitValidation
	}
	if errors.Is(err, domain.ErrNotFound) {
		return ExitNotFound
	}
	if errors.Is(err, ErrPermissionDenied) {
		return ExitPermission
	}

	return ExitInternal
}

func isValidationError(err error) bool {
	return errors.Is(err, domain.ErrValidation) ||
		errors.Is(err, domain.ErrSlugTaken) ||
		errors.Is(err, domain.ErrAlreadyPublished) ||
		errors.Is(err, domain.ErrNotPublished) ||
		errors.Is(err, domain.ErrQuizInUse) ||
		errors.Is(err, domain.ErrPracticeInUse) ||
		errors.Is(err, domain.ErrUnsupportedImportFormat) ||
		errors.Is(err, domain.ErrInvalidConflictStrategy) ||
		errors.Is(err, domain.ErrUnresolvedImportConflicts) ||
		errors.Is(err, domain.ErrImportSourceParse) ||
		errors.Is(err, domain.ErrImportSourceLayout) ||
		errors.Is(err, domain.ErrImportPlanHashMismatch) ||
		errors.Is(err, ErrRequiredFlagMissing) ||
		errors.Is(err, ErrInstructorIDRequired) ||
		errors.Is(err, ErrInvalidLessonOrder) ||
		errors.Is(err, ErrInvalidBlockOrder) ||
		errors.Is(err, ErrInvalidQuestionOrder) ||
		errors.Is(err, ErrInvalidTestCaseOrder) ||
		errors.Is(err, ErrUnsupportedOutputFormat) ||
		errors.Is(err, ErrConfirmationDeclined) ||
		errors.Is(err, ErrConfirmationRequired)
}
