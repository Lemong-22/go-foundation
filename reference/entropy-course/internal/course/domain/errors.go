package domain

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrValidation                = errors.New("validation failed")
	ErrNotFound                  = errors.New("not found")
	ErrSlugTaken                 = errors.New("slug taken")
	ErrAlreadyPublished          = errors.New("course already published")
	ErrNotPublished              = errors.New("course not published")
	ErrQuizInUse                 = errors.New("quiz in use")
	ErrPracticeInUse             = errors.New("practice in use")
	ErrUnsupportedImportFormat   = errors.New("unsupported import format")
	ErrInvalidConflictStrategy   = errors.New("invalid conflict strategy")
	ErrUnresolvedImportConflicts = errors.New("unresolved import conflicts")
	ErrImportSourceParse         = errors.New("import source parse failed")
	ErrImportSourceLayout        = errors.New("import source layout invalid")
	ErrImportPlanHashMismatch    = errors.New("resolved plan zip hash mismatch")
)

type ValidationError struct {
	Field  string
	Reason string
}

func NewValidationError(field string, reason string) ValidationError {
	return ValidationError{Field: field, Reason: reason}
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Reason)
}

func (e ValidationError) Unwrap() error {
	return ErrValidation
}

type QuizInUseError struct {
	LessonIDs []LessonID
}

func NewQuizInUseError(lessonIDs []LessonID) QuizInUseError {
	copied := make([]LessonID, len(lessonIDs))
	copy(copied, lessonIDs)

	return QuizInUseError{LessonIDs: copied}
}

func (e QuizInUseError) Error() string {
	return "quiz is embedded in one or more lessons"
}

func (e QuizInUseError) Unwrap() error {
	return ErrQuizInUse
}

type PracticeInUseError struct {
	LessonIDs []LessonID
}

func NewPracticeInUseError(lessonIDs []LessonID) PracticeInUseError {
	copied := make([]LessonID, len(lessonIDs))
	copy(copied, lessonIDs)

	return PracticeInUseError{LessonIDs: copied}
}

func (e PracticeInUseError) Error() string {
	return "practice is embedded in one or more lessons"
}

func (e PracticeInUseError) Unwrap() error {
	return ErrPracticeInUse
}

type UnsupportedImportFormatError struct {
	Version string
	Allowed []string
}

func NewUnsupportedImportFormatError(version string, allowed []string) UnsupportedImportFormatError {
	copied := make([]string, 0, len(allowed))
	for _, value := range allowed {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			copied = append(copied, trimmed)
		}
	}

	return UnsupportedImportFormatError{
		Version: strings.TrimSpace(version),
		Allowed: copied,
	}
}

func (e UnsupportedImportFormatError) Error() string {
	if len(e.Allowed) == 0 {
		return fmt.Sprintf("unsupported import format version %q", e.Version)
	}

	return fmt.Sprintf("unsupported import format version %q; supported versions: %s", e.Version, strings.Join(e.Allowed, ", "))
}

func (e UnsupportedImportFormatError) Unwrap() error {
	return ErrUnsupportedImportFormat
}

type InvalidConflictStrategyError struct {
	Strategy string
}

func NewInvalidConflictStrategyError(strategy string) InvalidConflictStrategyError {
	return InvalidConflictStrategyError{Strategy: strings.TrimSpace(strategy)}
}

func (e InvalidConflictStrategyError) Error() string {
	return fmt.Sprintf("invalid conflict strategy %q", e.Strategy)
}

func (e InvalidConflictStrategyError) Unwrap() error {
	return ErrInvalidConflictStrategy
}

type UnresolvedImportConflictsError struct {
	conflicts []ImportConflict
}

func NewUnresolvedImportConflictsError(conflicts []ImportConflict) UnresolvedImportConflictsError {
	copied, _ := copyImportConflicts(conflicts)
	return UnresolvedImportConflictsError{conflicts: copied}
}

func (e UnresolvedImportConflictsError) Error() string {
	return fmt.Sprintf("%d unresolved import conflict(s)", len(e.conflicts))
}

func (e UnresolvedImportConflictsError) Unwrap() error {
	return ErrUnresolvedImportConflicts
}

func (e UnresolvedImportConflictsError) Conflicts() []ImportConflict {
	copied, _ := copyImportConflicts(e.conflicts)
	return copied
}

type ImportSourceParseError struct {
	Path   string
	Reason string
	Cause  error
}

func NewImportSourceParseError(path string, reason string, cause error) ImportSourceParseError {
	return ImportSourceParseError{
		Path:   strings.TrimSpace(path),
		Reason: strings.TrimSpace(reason),
		Cause:  cause,
	}
}

func (e ImportSourceParseError) Error() string {
	return importSourceErrorMessage("parse import source", e.Path, e.Reason)
}

func (e ImportSourceParseError) Unwrap() []error {
	return unwrapImportError(ErrImportSourceParse, e.Cause)
}

type ImportSourceLayoutError struct {
	Path   string
	Reason string
	Cause  error
}

func NewImportSourceLayoutError(path string, reason string, cause error) ImportSourceLayoutError {
	return ImportSourceLayoutError{
		Path:   strings.TrimSpace(path),
		Reason: strings.TrimSpace(reason),
		Cause:  cause,
	}
}

func (e ImportSourceLayoutError) Error() string {
	return importSourceErrorMessage("invalid import source layout", e.Path, e.Reason)
}

func (e ImportSourceLayoutError) Unwrap() []error {
	return unwrapImportError(ErrImportSourceLayout, e.Cause)
}

type ImportPlanHashMismatchError struct {
	ExpectedZipHash string
	ActualZipHash   string
}

func NewImportPlanHashMismatchError(expectedZipHash string, actualZipHash string) ImportPlanHashMismatchError {
	return ImportPlanHashMismatchError{
		ExpectedZipHash: strings.TrimSpace(expectedZipHash),
		ActualZipHash:   strings.TrimSpace(actualZipHash),
	}
}

func (e ImportPlanHashMismatchError) Error() string {
	return fmt.Sprintf("resolved plan zip hash mismatch: expected %q, got %q", e.ExpectedZipHash, e.ActualZipHash)
}

func (e ImportPlanHashMismatchError) Unwrap() error {
	return ErrImportPlanHashMismatch
}

func importSourceErrorMessage(prefix string, path string, reason string) string {
	switch {
	case path != "" && reason != "":
		return fmt.Sprintf("%s %q: %s", prefix, path, reason)
	case path != "":
		return fmt.Sprintf("%s %q", prefix, path)
	case reason != "":
		return fmt.Sprintf("%s: %s", prefix, reason)
	default:
		return prefix
	}
}

func unwrapImportError(sentinel error, cause error) []error {
	if cause == nil {
		return []error{sentinel}
	}

	return []error{sentinel, cause}
}
