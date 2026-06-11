package domain

import "strings"

const (
	courseStatusDraft     = "draft"
	courseStatusPublished = "published"
)

type CourseStatus struct {
	value string
}

func Draft() CourseStatus {
	return CourseStatus{value: courseStatusDraft}
}

func Published() CourseStatus {
	return CourseStatus{value: courseStatusPublished}
}

func NewCourseStatus(value string) (CourseStatus, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case courseStatusDraft:
		return Draft(), nil
	case courseStatusPublished:
		return Published(), nil
	default:
		return CourseStatus{}, NewValidationError("status", "must be draft or published")
	}
}

func (status CourseStatus) String() string {
	return status.value
}

func (status CourseStatus) IsPublished() bool {
	return status.value == courseStatusPublished
}
