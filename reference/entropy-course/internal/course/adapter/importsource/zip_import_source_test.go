package importsource

import (
	"archive/zip"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/luxeave/entropy-course/internal/course/domain"
)

func TestZipImportSourceOpenParsesMinimalZip(t *testing.T) {
	zipPath := writeImportZip(t, []zipEntry{
		{name: "format_version.txt", body: "1\n"},
		{name: "course.yaml", body: `
title: Intro to Go
slug: intro-to-go
description: Learn Go
status: draft
`},
	})

	parsed, metadata, err := NewZipImportSource().Open(zipPath)
	if err != nil {
		t.Fatalf("expected import source to open, got error %v", err)
	}

	if metadata.FormatVersion != "1" || len(metadata.ZipHash) != 64 {
		t.Fatalf("expected import metadata, got %+v", metadata)
	}
	if parsed.FormatVersion != "1" || parsed.Course.Title != "Intro to Go" || parsed.Course.Slug != "intro-to-go" {
		t.Fatalf("expected parsed course, got %+v", parsed.Course)
	}
	if len(parsed.Lessons) != 0 || len(parsed.Quizzes) != 0 || len(parsed.Practices) != 0 || len(parsed.Tests) != 0 {
		t.Fatalf("expected minimal zip to contain only course metadata")
	}
}

func TestZipImportSourceOpenParsesFullZip(t *testing.T) {
	zipPath := writeImportZip(t, fullImportEntries())

	parsed, metadata, err := NewZipImportSource().Open(zipPath)
	if err != nil {
		t.Fatalf("expected import source to open, got error %v", err)
	}

	if metadata.FormatVersion != "1" || metadata.ZipHash == "" {
		t.Fatalf("expected metadata, got %+v", metadata)
	}
	if parsed.Course.Status != "published" {
		t.Fatalf("expected course status to be parsed, got %q", parsed.Course.Status)
	}

	if len(parsed.Lessons) != 1 {
		t.Fatalf("expected one lesson, got %d", len(parsed.Lessons))
	}
	lesson := parsed.Lessons[0]
	if lesson.Title != "Foundations" || lesson.Order == nil || *lesson.Order != 0 {
		t.Fatalf("expected lesson metadata, got %+v", lesson)
	}
	if len(lesson.Blocks) != 4 {
		t.Fatalf("expected four lesson blocks, got %d", len(lesson.Blocks))
	}
	if lesson.Blocks[0].Markdown != "Welcome to Go." {
		t.Fatalf("expected text block markdown, got %q", lesson.Blocks[0].Markdown)
	}
	if lesson.Blocks[1].VideoProvider != "youtube" || lesson.Blocks[1].VideoLocator != "dQw4w9WgXcQ" {
		t.Fatalf("expected video block fields, got %+v", lesson.Blocks[1])
	}
	if lesson.Blocks[2].QuizRef != "foundations-quiz" || lesson.Blocks[3].PracticeRef != "fizzbuzz" {
		t.Fatalf("expected quiz and practice refs, got %+v %+v", lesson.Blocks[2], lesson.Blocks[3])
	}

	if len(parsed.Quizzes) != 1 || parsed.Quizzes[0].Questions[0].CorrectIndices[0] != 0 {
		t.Fatalf("expected quiz content, got %+v", parsed.Quizzes)
	}
	if len(parsed.Practices) != 1 || parsed.Practices[0].TestCases[0].ExpectedStdout != "Fizz" {
		t.Fatalf("expected practice content, got %+v", parsed.Practices)
	}
	if len(parsed.Tests) != 1 || parsed.Tests[0].Items[1].TestCases[0].ExpectedStdout != "Fizz" {
		t.Fatalf("expected test content, got %+v", parsed.Tests)
	}
	if parsed.Tests[0].Solution == nil || parsed.Tests[0].Solution.VideoCaption != "Walkthrough" {
		t.Fatalf("expected test solution, got %+v", parsed.Tests[0].Solution)
	}
}

func TestZipImportSourceCanonicalHashIgnoresEntryOrder(t *testing.T) {
	entries := fullImportEntries()
	reversed := make([]zipEntry, len(entries))
	for i := range entries {
		reversed[len(entries)-1-i] = entries[i]
	}

	_, firstMetadata, err := NewZipImportSource().Open(writeImportZip(t, entries))
	if err != nil {
		t.Fatalf("expected first zip to parse, got %v", err)
	}
	_, secondMetadata, err := NewZipImportSource().Open(writeImportZip(t, reversed))
	if err != nil {
		t.Fatalf("expected second zip to parse, got %v", err)
	}

	if firstMetadata.ZipHash != secondMetadata.ZipHash {
		t.Fatalf("expected canonical hash to ignore entry order: %q != %q", firstMetadata.ZipHash, secondMetadata.ZipHash)
	}
}

func TestZipImportSourceOpenReturnsTypedErrors(t *testing.T) {
	tests := []struct {
		name    string
		entries []zipEntry
		want    error
	}{
		{
			name: "missing format version",
			entries: []zipEntry{
				{name: "course.yaml", body: validCourseYAML()},
			},
			want: domain.ErrImportSourceLayout,
		},
		{
			name: "unsupported format version",
			entries: []zipEntry{
				{name: "format_version.txt", body: "2"},
				{name: "course.yaml", body: validCourseYAML()},
			},
			want: domain.ErrUnsupportedImportFormat,
		},
		{
			name: "malformed course yaml",
			entries: []zipEntry{
				{name: "format_version.txt", body: "1"},
				{name: "course.yaml", body: "title: ["},
			},
			want: domain.ErrImportSourceParse,
		},
		{
			name: "missing lesson frontmatter",
			entries: append([]zipEntry{
				{name: "format_version.txt", body: "1"},
				{name: "course.yaml", body: validCourseYAML()},
			}, quizEntry(), zipEntry{name: "lessons/01-foundations.md", body: "# No frontmatter"}),
			want: domain.ErrImportSourceParse,
		},
		{
			name: "duplicate quiz slug",
			entries: []zipEntry{
				{name: "format_version.txt", body: "1"},
				{name: "course.yaml", body: validCourseYAML()},
				quizEntry(),
				{name: "quizzes/duplicate.yaml", body: quizYAMLContent("foundations-quiz")},
			},
			want: domain.ErrImportSourceLayout,
		},
		{
			name: "unknown quiz reference",
			entries: []zipEntry{
				{name: "format_version.txt", body: "1"},
				{name: "course.yaml", body: validCourseYAML()},
				{name: "lessons/01-foundations.md", body: `---
title: Foundations
blocks:
  - kind: quiz
    quiz_ref: missing-quiz
---
`},
			},
			want: domain.ErrImportSourceLayout,
		},
		{
			name: "unexpected layout path",
			entries: []zipEntry{
				{name: "format_version.txt", body: "1"},
				{name: "course.yaml", body: validCourseYAML()},
				{name: "notes/readme.txt", body: "not part of the format"},
			},
			want: domain.ErrImportSourceLayout,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := NewZipImportSource().Open(writeImportZip(t, test.entries))
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}

type zipEntry struct {
	name string
	body string
}

func writeImportZip(t *testing.T, entries []zipEntry) string {
	t.Helper()

	zipPath := filepath.Join(t.TempDir(), "course.zip")
	file, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}

	writer := zip.NewWriter(file)
	for _, entry := range entries {
		entryWriter, err := writer.Create(entry.name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := entryWriter.Write([]byte(entry.body)); err != nil {
			t.Fatalf("write zip entry: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close zip file: %v", err)
	}

	return zipPath
}

func fullImportEntries() []zipEntry {
	return []zipEntry{
		{name: "format_version.txt", body: "1\n"},
		{name: "course.yaml", body: `
title: Intro to Go
slug: intro-to-go
description: Learn Go
status: published
`},
		{name: "quizzes/foundations-quiz.yaml", body: quizYAMLContent("foundations-quiz")},
		{name: "practices/fizzbuzz.yaml", body: `
slug: fizzbuzz
title: FizzBuzz
language: golang
prompt: Print Fizz for multiples of three.
starter_code: package main
solution: package main
test_cases:
  - stdin: "3"
    expected_stdout: Fizz
    name: multiple of three
    position: 0
`},
		{name: "tests/midterm.yaml", body: `
slug: midterm
title: Midterm
time_limit_minutes: 30
pass_threshold: 0.8
solution:
  zip_provider: url
  zip_locator: https://example.com/solution.zip
  video_provider: youtube
  video_locator: dQw4w9WgXcQ
  video_caption: Walkthrough
items:
  - kind: choice
    prompt: Pick one
    choice_type: single
    options:
      - A
      - B
    correct_indices:
      - 0
    explanation: Because A.
    position: 0
  - kind: coding
    coding_prompt: Write FizzBuzz
    language: golang
    starter_code: package main
    solution: package main
    test_cases:
      - stdin: "3"
        expected_stdout: Fizz
        name: multiple of three
    position: 1
`},
		{name: "lessons/01-foundations.md", body: `---
title: Foundations
order: 0
blocks:
  - kind: text
    markdown: Welcome to Go.
    position: 0
  - kind: video
    video_provider: youtube
    video_locator: dQw4w9WgXcQ
    video_caption: Setup
    position: 1
  - kind: quiz
    quiz_ref: foundations-quiz
    position: 2
  - kind: practice
    practice_ref: fizzbuzz
    position: 3
---
Lesson body.
`},
	}
}

func validCourseYAML() string {
	return `
title: Intro to Go
slug: intro-to-go
description: Learn Go
status: draft
`
}

func quizEntry() zipEntry {
	return zipEntry{name: "quizzes/foundations-quiz.yaml", body: quizYAMLContent("foundations-quiz")}
}

func quizYAMLContent(slug string) string {
	return `
slug: ` + slug + `
title: Foundations Quiz
pass_threshold: 0.7
questions:
  - type: single
    prompt: Pick one
    options:
      - A
      - B
    correct_indices:
      - 0
    explanation: Because A.
    position: 0
`
}
