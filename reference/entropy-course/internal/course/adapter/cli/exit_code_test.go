package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/luxeave/entropy-course/internal/course/domain"
)

func TestExitCodeMapsKnownErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "nil", err: nil, want: ExitOK},
		{name: "domain validation", err: domain.NewValidationError("id", "invalid"), want: ExitValidation},
		{name: "slug taken", err: domain.ErrSlugTaken, want: ExitValidation},
		{name: "required flag", err: fmt.Errorf("%w: --title", ErrRequiredFlagMissing), want: ExitValidation},
		{name: "invalid lesson order", err: fmt.Errorf("%w: bad", ErrInvalidLessonOrder), want: ExitValidation},
		{name: "invalid block order", err: fmt.Errorf("%w: bad", ErrInvalidBlockOrder), want: ExitValidation},
		{name: "invalid question order", err: fmt.Errorf("%w: bad", ErrInvalidQuestionOrder), want: ExitValidation},
		{name: "invalid test case order", err: fmt.Errorf("%w: bad", ErrInvalidTestCaseOrder), want: ExitValidation},
		{name: "quiz in use", err: domain.NewQuizInUseError(nil), want: ExitValidation},
		{name: "practice in use", err: domain.NewPracticeInUseError(nil), want: ExitValidation},
		{name: "unresolved import conflicts", err: domain.NewUnresolvedImportConflictsError(nil), want: ExitValidation},
		{name: "import hash mismatch", err: domain.NewImportPlanHashMismatchError("a", "b"), want: ExitValidation},
		{name: "not found", err: domain.ErrNotFound, want: ExitNotFound},
		{name: "permission reserved", err: ErrPermissionDenied, want: ExitPermission},
		{name: "internal", err: errors.New("database unavailable"), want: ExitInternal},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ExitCode(test.err); got != test.want {
				t.Fatalf("expected exit code %d, got %d", test.want, got)
			}
		})
	}
}
