package domain

import (
	"errors"
	"testing"
)

const validUUID = "550e8400-e29b-41d4-a716-446655440000"

func TestIDValueObjectsValidateUUIDs(t *testing.T) {
	tests := []struct {
		name string
		make func(string) (string, error)
	}{
		{
			name: "course id",
			make: func(value string) (string, error) {
				id, err := NewCourseID(value)
				return id.String(), err
			},
		},
		{
			name: "lesson id",
			make: func(value string) (string, error) {
				id, err := NewLessonID(value)
				return id.String(), err
			},
		},
		{
			name: "block id",
			make: func(value string) (string, error) {
				id, err := NewBlockID(value)
				return id.String(), err
			},
		},
		{
			name: "quiz id",
			make: func(value string) (string, error) {
				id, err := NewQuizID(value)
				return id.String(), err
			},
		},
		{
			name: "practice id",
			make: func(value string) (string, error) {
				id, err := NewPracticeID(value)
				return id.String(), err
			},
		},
		{
			name: "test id",
			make: func(value string) (string, error) {
				id, err := NewTestID(value)
				return id.String(), err
			},
		},
		{
			name: "question id",
			make: func(value string) (string, error) {
				id, err := NewQuestionID(value)
				return id.String(), err
			},
		},
		{
			name: "test case id",
			make: func(value string) (string, error) {
				id, err := NewTestCaseID(value)
				return id.String(), err
			},
		},
		{
			name: "test item id",
			make: func(value string) (string, error) {
				id, err := NewTestItemID(value)
				return id.String(), err
			},
		},
		{
			name: "instructor id",
			make: func(value string) (string, error) {
				id, err := NewInstructorID(value)
				return id.String(), err
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, err := test.make(validUUID)
			if err != nil {
				t.Fatalf("expected valid UUID, got error %v", err)
			}

			if id != validUUID {
				t.Fatalf("expected canonical UUID %q, got %q", validUUID, id)
			}

			if _, err := test.make("not-a-uuid"); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error for invalid UUID, got %v", err)
			}

			if _, err := test.make("   "); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error for empty UUID, got %v", err)
			}
		})
	}
}

func TestNewSlug(t *testing.T) {
	validSlugs := []string{
		"go",
		"go-101",
		"intro-to-go",
		"course2",
	}

	for _, value := range validSlugs {
		t.Run(value, func(t *testing.T) {
			slug, err := NewSlug(value)
			if err != nil {
				t.Fatalf("expected valid slug, got error %v", err)
			}

			if slug.String() != value {
				t.Fatalf("expected slug %q, got %q", value, slug.String())
			}
		})
	}
}

func TestNewSlugRejectsInvalidFormat(t *testing.T) {
	invalidSlugs := []string{
		"",
		"   ",
		"Intro",
		"intro_to_go",
		"intro--go",
		"-intro",
		"intro-",
		"intro go",
	}

	for _, value := range invalidSlugs {
		t.Run(value, func(t *testing.T) {
			if _, err := NewSlug(value); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestCourseStatus(t *testing.T) {
	draft, err := NewCourseStatus("draft")
	if err != nil {
		t.Fatalf("expected draft status, got error %v", err)
	}

	if draft.String() != "draft" || draft.IsPublished() {
		t.Fatalf("expected draft status, got %q", draft.String())
	}

	published := Published()
	if published.String() != "published" || !published.IsPublished() {
		t.Fatalf("expected published status, got %q", published.String())
	}

}

func TestDeferredArchivedCourseStatusRemainsAbsent(t *testing.T) {
	if _, err := NewCourseStatus("archived"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected deferred archived status to remain absent, got %v", err)
	}
}

func TestLessonOrder(t *testing.T) {
	order, err := NewLessonOrder(0)
	if err != nil {
		t.Fatalf("expected zero order to be valid, got error %v", err)
	}

	if order.Int() != 0 {
		t.Fatalf("expected order 0, got %d", order.Int())
	}

	if _, err := NewLessonOrder(-1); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error for negative order, got %v", err)
	}
}

func TestBlockPosition(t *testing.T) {
	position, err := NewBlockPosition(0)
	if err != nil {
		t.Fatalf("expected zero position to be valid, got error %v", err)
	}

	if position.Int() != 0 {
		t.Fatalf("expected position 0, got %d", position.Int())
	}

	if _, err := NewBlockPosition(-1); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error for negative position, got %v", err)
	}
}

func TestQuestionPosition(t *testing.T) {
	position, err := NewQuestionPosition(0)
	if err != nil {
		t.Fatalf("expected zero position to be valid, got error %v", err)
	}

	if position.Int() != 0 {
		t.Fatalf("expected position 0, got %d", position.Int())
	}

	if _, err := NewQuestionPosition(-1); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error for negative position, got %v", err)
	}
}

func TestTestCasePosition(t *testing.T) {
	position, err := NewTestCasePosition(0)
	if err != nil {
		t.Fatalf("expected zero position to be valid, got error %v", err)
	}

	if position.Int() != 0 {
		t.Fatalf("expected position 0, got %d", position.Int())
	}

	if _, err := NewTestCasePosition(-1); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error for negative position, got %v", err)
	}
}

func TestContentBlockKind(t *testing.T) {
	text, err := NewContentBlockKind("text")
	if err != nil {
		t.Fatalf("expected text kind, got error %v", err)
	}
	if !text.IsText() || text.IsVideo() || text.IsQuiz() || text.IsPractice() || text.String() != "text" {
		t.Fatalf("expected text kind, got %q", text.String())
	}

	video := VideoKind()
	if !video.IsVideo() || video.IsText() || video.IsQuiz() || video.IsPractice() || video.String() != "video" {
		t.Fatalf("expected video kind, got %q", video.String())
	}

	quiz, err := NewContentBlockKind("quiz")
	if err != nil {
		t.Fatalf("expected quiz kind, got error %v", err)
	}
	if !quiz.IsQuiz() || quiz.IsText() || quiz.IsVideo() || quiz.IsPractice() || quiz.String() != "quiz" {
		t.Fatalf("expected quiz kind, got %q", quiz.String())
	}

	practice, err := NewContentBlockKind("practice")
	if err != nil {
		t.Fatalf("expected practice kind, got error %v", err)
	}
	if !practice.IsPractice() || practice.IsText() || practice.IsVideo() || practice.IsQuiz() || practice.String() != "practice" {
		t.Fatalf("expected practice kind, got %q", practice.String())
	}

	if _, err := NewContentBlockKind("assessment"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error for unsupported kind, got %v", err)
	}
}

func TestChoiceQuestionType(t *testing.T) {
	single, err := NewChoiceQuestionType("single")
	if err != nil {
		t.Fatalf("expected single-choice type, got error %v", err)
	}
	if !single.IsSingle() || single.IsMultiple() || single.String() != "single" {
		t.Fatalf("expected single-choice type, got %q", single.String())
	}

	multiple := MultipleChoice()
	if !multiple.IsMultiple() || multiple.IsSingle() || multiple.String() != "multiple" {
		t.Fatalf("expected multiple-choice type, got %q", multiple.String())
	}

	if _, err := NewChoiceQuestionType("short-answer"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error for unsupported question type, got %v", err)
	}
}

func TestPassThreshold(t *testing.T) {
	threshold, err := NewPassThreshold(0.8)
	if err != nil {
		t.Fatalf("expected pass threshold, got error %v", err)
	}
	if threshold.Float64() != 0.8 {
		t.Fatalf("expected threshold 0.8, got %f", threshold.Float64())
	}
	if DefaultPassThreshold().Float64() != 0.7 {
		t.Fatalf("expected default threshold 0.7, got %f", DefaultPassThreshold().Float64())
	}

	for _, value := range []float64{-0.1, 1.1} {
		if _, err := NewPassThreshold(value); !errors.Is(err, ErrValidation) {
			t.Fatalf("expected validation error for threshold %f, got %v", value, err)
		}
	}
}

func TestContentBlockRejectsBodyKindMismatch(t *testing.T) {
	_, err := NewContentBlock(
		mustBlockID(t, blockIDValue),
		TextKind(),
		mustBlockPosition(t, 0),
		VideoBody{Media: mustMediaRef(t, URLProvider(), "https://example.com/video.mp4")},
	)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected body kind mismatch validation error, got %v", err)
	}
}

func TestMediaProvider(t *testing.T) {
	tests := []struct {
		value string
		want  MediaProvider
	}{
		{value: "url", want: URLProvider()},
		{value: "youtube", want: YouTubeProvider()},
		{value: "mux", want: MuxProvider()},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			provider, err := NewMediaProvider(test.value)
			if err != nil {
				t.Fatalf("expected provider, got error %v", err)
			}

			if provider != test.want || provider.String() != test.value {
				t.Fatalf("expected provider %q, got %q", test.want.String(), provider.String())
			}
		})
	}

	if _, err := NewMediaProvider("vimeo"); !errors.Is(err, ErrValidation) {
		t.Fatalf("expected unsupported provider validation error, got %v", err)
	}
}

func TestMediaRefValidatesProviderLocators(t *testing.T) {
	valid := []struct {
		name     string
		provider MediaProvider
		locator  string
	}{
		{name: "absolute url", provider: URLProvider(), locator: "https://example.com/video.mp4"},
		{name: "youtube id", provider: YouTubeProvider(), locator: "dQw4w9WgXcQ"},
		{name: "youtube watch url", provider: YouTubeProvider(), locator: "https://www.youtube.com/watch?v=dQw4w9WgXcQ"},
		{name: "youtube short url", provider: YouTubeProvider(), locator: "https://youtu.be/dQw4w9WgXcQ"},
		{name: "mux playback id", provider: MuxProvider(), locator: "muxPlayback_123"},
	}

	for _, test := range valid {
		t.Run(test.name, func(t *testing.T) {
			ref, err := NewMediaRef(test.provider, test.locator)
			if err != nil {
				t.Fatalf("expected media ref, got error %v", err)
			}
			if ref.Provider() != test.provider || ref.Locator() != test.locator {
				t.Fatalf("expected provider and locator to be retained")
			}
		})
	}
}

func TestMediaRefRejectsInvalidLocators(t *testing.T) {
	invalid := []struct {
		name     string
		provider MediaProvider
		locator  string
	}{
		{name: "empty", provider: URLProvider(), locator: "   "},
		{name: "relative url", provider: URLProvider(), locator: "/video.mp4"},
		{name: "ftp url", provider: URLProvider(), locator: "ftp://example.com/video.mp4"},
		{name: "non-youtube url", provider: YouTubeProvider(), locator: "https://example.com/watch?v=dQw4w9WgXcQ"},
		{name: "invalid mux id", provider: MuxProvider(), locator: "mux playback"},
	}

	for _, test := range invalid {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewMediaRef(test.provider, test.locator); !errors.Is(err, ErrValidation) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestValidationErrorWrapsSentinel(t *testing.T) {
	err := NewValidationError("slug", "must not be empty")
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error to wrap ErrValidation")
	}
}
