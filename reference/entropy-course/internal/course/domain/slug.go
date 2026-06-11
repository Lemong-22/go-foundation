package domain

import (
	"regexp"
	"strings"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Slug struct {
	value string
}

func NewSlug(value string) (Slug, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return Slug{}, NewValidationError("slug", "must not be empty")
	}

	if !slugPattern.MatchString(trimmed) {
		return Slug{}, NewValidationError("slug", "must contain lowercase letters, numbers, and single hyphens only")
	}

	return Slug{value: trimmed}, nil
}

func (slug Slug) String() string {
	return slug.value
}
