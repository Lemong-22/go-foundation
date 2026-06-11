package domain

import "time"

type Course struct {
	id           CourseID
	title        string
	slug         Slug
	description  string
	instructorID InstructorID
	status       CourseStatus
	createdAt    time.Time
	updatedAt    time.Time
}

func NewCourse(
	id CourseID,
	title string,
	slug Slug,
	description string,
	instructorID InstructorID,
	now time.Time,
) (Course, error) {
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return Course{}, err
	}

	return Course{
		id:           id,
		title:        normalizedTitle,
		slug:         slug,
		description:  description,
		instructorID: instructorID,
		status:       Draft(),
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func RestoreCourse(
	id CourseID,
	title string,
	slug Slug,
	description string,
	instructorID InstructorID,
	status CourseStatus,
	createdAt time.Time,
	updatedAt time.Time,
) (Course, error) {
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return Course{}, err
	}

	return Course{
		id:           id,
		title:        normalizedTitle,
		slug:         slug,
		description:  description,
		instructorID: instructorID,
		status:       status,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}

func (course Course) ID() CourseID {
	return course.id
}

func (course Course) Title() string {
	return course.title
}

func (course Course) Slug() Slug {
	return course.slug
}

func (course Course) Description() string {
	return course.description
}

func (course Course) InstructorID() InstructorID {
	return course.instructorID
}

func (course Course) Status() CourseStatus {
	return course.status
}

func (course Course) CreatedAt() time.Time {
	return course.createdAt
}

func (course Course) UpdatedAt() time.Time {
	return course.updatedAt
}

func (course *Course) Rename(title string, now time.Time) error {
	normalizedTitle, err := normalizeTitle(title)
	if err != nil {
		return err
	}

	course.title = normalizedTitle
	course.touch(now)

	return nil
}

func (course *Course) ChangeDescription(description string, now time.Time) {
	course.description = description
	course.touch(now)
}

func (course *Course) ChangeSlug(slug Slug, now time.Time) {
	course.slug = slug
	course.touch(now)
}

func (course *Course) Publish(now time.Time) error {
	if course.status.IsPublished() {
		return ErrAlreadyPublished
	}

	course.status = Published()
	course.touch(now)

	return nil
}

func (course *Course) Unpublish(now time.Time) error {
	if !course.status.IsPublished() {
		return ErrNotPublished
	}

	course.status = Draft()
	course.touch(now)

	return nil
}

func (course *Course) touch(now time.Time) {
	course.updatedAt = mutationTime(course.createdAt, now)
}
