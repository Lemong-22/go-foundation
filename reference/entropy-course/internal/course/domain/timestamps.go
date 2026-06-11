package domain

import "time"

func mutationTime(createdAt time.Time, now time.Time) time.Time {
	if now.Before(createdAt) {
		return createdAt
	}

	return now
}
