package domain

import (
	"net/url"
	"regexp"
	"strings"
)

const (
	contentBlockKindText     = "text"
	contentBlockKindVideo    = "video"
	contentBlockKindQuiz     = "quiz"
	contentBlockKindPractice = "practice"

	mediaProviderURL     = "url"
	mediaProviderYouTube = "youtube"
	mediaProviderMux     = "mux"
)

var (
	youtubeIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{6,}$`)
	muxIDPattern     = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

type ContentBlockKind struct {
	value string
}

func TextKind() ContentBlockKind {
	return ContentBlockKind{value: contentBlockKindText}
}

func VideoKind() ContentBlockKind {
	return ContentBlockKind{value: contentBlockKindVideo}
}

func QuizKind() ContentBlockKind {
	return ContentBlockKind{value: contentBlockKindQuiz}
}

func PracticeKind() ContentBlockKind {
	return ContentBlockKind{value: contentBlockKindPractice}
}

func NewContentBlockKind(value string) (ContentBlockKind, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case contentBlockKindText:
		return TextKind(), nil
	case contentBlockKindVideo:
		return VideoKind(), nil
	case contentBlockKindQuiz:
		return QuizKind(), nil
	case contentBlockKindPractice:
		return PracticeKind(), nil
	default:
		return ContentBlockKind{}, NewValidationError("kind", "must be text, video, quiz, or practice")
	}
}

func (kind ContentBlockKind) String() string {
	return kind.value
}

func (kind ContentBlockKind) IsText() bool {
	return kind.value == contentBlockKindText
}

func (kind ContentBlockKind) IsVideo() bool {
	return kind.value == contentBlockKindVideo
}

func (kind ContentBlockKind) IsQuiz() bool {
	return kind.value == contentBlockKindQuiz
}

func (kind ContentBlockKind) IsPractice() bool {
	return kind.value == contentBlockKindPractice
}

type BlockPosition struct {
	value int
}

func NewBlockPosition(value int) (BlockPosition, error) {
	if value < 0 {
		return BlockPosition{}, NewValidationError("position", "must be greater than or equal to zero")
	}

	return BlockPosition{value: value}, nil
}

func (position BlockPosition) Int() int {
	return position.value
}

type MediaProvider struct {
	value string
}

func URLProvider() MediaProvider {
	return MediaProvider{value: mediaProviderURL}
}

func YouTubeProvider() MediaProvider {
	return MediaProvider{value: mediaProviderYouTube}
}

func MuxProvider() MediaProvider {
	return MediaProvider{value: mediaProviderMux}
}

func NewMediaProvider(value string) (MediaProvider, error) {
	trimmed := strings.TrimSpace(value)
	switch trimmed {
	case mediaProviderURL:
		return URLProvider(), nil
	case mediaProviderYouTube:
		return YouTubeProvider(), nil
	case mediaProviderMux:
		return MuxProvider(), nil
	default:
		return MediaProvider{}, NewValidationError("media_provider", "must be url, youtube, or mux")
	}
}

func (provider MediaProvider) String() string {
	return provider.value
}

type MediaRef struct {
	provider MediaProvider
	locator  string
}

func NewMediaRef(provider MediaProvider, locator string) (MediaRef, error) {
	trimmed := strings.TrimSpace(locator)
	if trimmed == "" {
		return MediaRef{}, NewValidationError("media_locator", "must not be empty")
	}

	if !isValidMediaLocator(provider, trimmed) {
		return MediaRef{}, NewValidationError("media_locator", "is invalid for provider")
	}

	return MediaRef{provider: provider, locator: trimmed}, nil
}

func (ref MediaRef) Provider() MediaProvider {
	return ref.provider
}

func (ref MediaRef) Locator() string {
	return ref.locator
}

type ContentBody interface {
	Kind() ContentBlockKind
	isContentBody()
}

type TextBody struct {
	Markdown string
}

func (body TextBody) Kind() ContentBlockKind {
	return TextKind()
}

func (TextBody) isContentBody() {}

type VideoBody struct {
	Media   MediaRef
	Caption string
}

func (body VideoBody) Kind() ContentBlockKind {
	return VideoKind()
}

func (VideoBody) isContentBody() {}

type QuizBody struct {
	QuizRef QuizID
}

func (body QuizBody) Kind() ContentBlockKind {
	return QuizKind()
}

func (QuizBody) isContentBody() {}

type PracticeBody struct {
	PracticeRef PracticeID
}

func (body PracticeBody) Kind() ContentBlockKind {
	return PracticeKind()
}

func (PracticeBody) isContentBody() {}

type ContentBlock struct {
	id       BlockID
	kind     ContentBlockKind
	position BlockPosition
	body     ContentBody
}

func NewContentBlock(
	id BlockID,
	kind ContentBlockKind,
	position BlockPosition,
	body ContentBody,
) (ContentBlock, error) {
	if body == nil {
		return ContentBlock{}, NewValidationError("body", "must not be nil")
	}
	if body.Kind() != kind {
		return ContentBlock{}, NewValidationError("body", "must match block kind")
	}

	return ContentBlock{
		id:       id,
		kind:     kind,
		position: position,
		body:     body,
	}, nil
}

func NewTextBlock(id BlockID, position BlockPosition, markdown string) (ContentBlock, error) {
	return NewContentBlock(id, TextKind(), position, TextBody{Markdown: markdown})
}

func NewVideoBlock(
	id BlockID,
	position BlockPosition,
	media MediaRef,
	caption string,
) (ContentBlock, error) {
	return NewContentBlock(id, VideoKind(), position, VideoBody{Media: media, Caption: caption})
}

func NewQuizBlock(id BlockID, position BlockPosition, quizRef QuizID) (ContentBlock, error) {
	return NewContentBlock(id, QuizKind(), position, QuizBody{QuizRef: quizRef})
}

func NewPracticeBlock(id BlockID, position BlockPosition, practiceRef PracticeID) (ContentBlock, error) {
	return NewContentBlock(id, PracticeKind(), position, PracticeBody{PracticeRef: practiceRef})
}

func (block ContentBlock) ID() BlockID {
	return block.id
}

func (block ContentBlock) Kind() ContentBlockKind {
	return block.kind
}

func (block ContentBlock) Position() BlockPosition {
	return block.position
}

func (block ContentBlock) Body() ContentBody {
	return block.body
}

func (block *ContentBlock) ChangeBody(body ContentBody) error {
	if body == nil {
		return NewValidationError("body", "must not be nil")
	}
	if body.Kind() != block.kind {
		return NewValidationError("body", "must match block kind")
	}

	block.body = body
	return nil
}

func (block *ContentBlock) MoveTo(position BlockPosition) {
	block.position = position
}

type BlockPlacement struct {
	BlockID  BlockID
	Position BlockPosition
}

func isValidMediaLocator(provider MediaProvider, locator string) bool {
	switch provider {
	case URLProvider():
		return isValidAbsoluteHTTPURL(locator)
	case YouTubeProvider():
		return isValidYouTubeLocator(locator)
	case MuxProvider():
		return muxIDPattern.MatchString(locator)
	default:
		return false
	}
}

func isValidAbsoluteHTTPURL(locator string) bool {
	parsed, err := url.Parse(locator)
	if err != nil {
		return false
	}

	return parsed.IsAbs() &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.Host != ""
}

func isValidYouTubeLocator(locator string) bool {
	if youtubeIDPattern.MatchString(locator) {
		return true
	}

	parsed, err := url.Parse(locator)
	if err != nil || !parsed.IsAbs() {
		return false
	}

	host := strings.ToLower(parsed.Hostname())
	switch {
	case host == "youtu.be":
		return youtubeIDPattern.MatchString(strings.Trim(parsed.Path, "/"))
	case host == "youtube.com" || strings.HasSuffix(host, ".youtube.com"):
		if id := parsed.Query().Get("v"); id != "" {
			return youtubeIDPattern.MatchString(id)
		}

		parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		if len(parts) == 2 && (parts[0] == "embed" || parts[0] == "shorts") {
			return youtubeIDPattern.MatchString(parts[1])
		}
	}

	return false
}
