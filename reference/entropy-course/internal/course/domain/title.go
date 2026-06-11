package domain

import "strings"

func normalizeTitle(title string) (string, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "", NewValidationError("title", "must not be empty")
	}

	return trimmed, nil
}
