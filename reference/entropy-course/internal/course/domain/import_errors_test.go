package domain

import (
	"errors"
	"testing"
)

func TestImportErrorsWrapSentinels(t *testing.T) {
	cause := errors.New("bad yaml")

	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "unsupported format",
			err:  NewUnsupportedImportFormatError("2", []string{"1"}),
			want: ErrUnsupportedImportFormat,
		},
		{
			name: "invalid strategy",
			err:  NewInvalidConflictStrategyError("merge"),
			want: ErrInvalidConflictStrategy,
		},
		{
			name: "unresolved conflicts",
			err:  NewUnresolvedImportConflictsError([]ImportConflict{mustImportConflict(t)}),
			want: ErrUnresolvedImportConflicts,
		},
		{
			name: "parse failure",
			err:  NewImportSourceParseError("course.zip", "invalid course.yaml", cause),
			want: ErrImportSourceParse,
		},
		{
			name: "layout failure",
			err:  NewImportSourceLayoutError("course.zip", "missing course.yaml", cause),
			want: ErrImportSourceLayout,
		},
		{
			name: "hash mismatch",
			err:  NewImportPlanHashMismatchError(importZipHash, "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
			want: ErrImportPlanHashMismatch,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if !errors.Is(test.err, test.want) {
				t.Fatalf("expected %v to wrap %v", test.err, test.want)
			}
		})
	}

	parseErr := NewImportSourceParseError("course.zip", "invalid course.yaml", cause)
	if !errors.Is(parseErr, cause) {
		t.Fatalf("expected parse error to retain cause")
	}
}

func TestUnresolvedImportConflictsErrorCopiesConflicts(t *testing.T) {
	err := NewUnresolvedImportConflictsError([]ImportConflict{mustImportConflict(t)})

	conflicts := err.Conflicts()
	conflicts[0] = mustImportConflictForRef(t, "course:other")

	if err.Conflicts()[0].EntityRef() != "course:intro" {
		t.Fatalf("expected conflicts accessor to return a copy")
	}
}

func mustImportConflictForRef(t *testing.T, entityRef string) ImportConflict {
	t.Helper()

	conflict, err := NewImportConflict(
		CourseEntity(),
		entityRef,
		SlugCollision(),
		[]ConflictCandidate{mustConflictCandidate(t, targetIDValue, "course Intro")},
		UpdateOperation(),
		[]byte(`{"slug":"intro"}`),
	)
	if err != nil {
		t.Fatalf("expected import conflict, got error %v", err)
	}

	return conflict
}
