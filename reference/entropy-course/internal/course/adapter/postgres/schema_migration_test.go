package postgres

import (
	"os"
	"strings"
	"testing"
)

func TestCourseSchemaMigrationMatchesRepositories(t *testing.T) {
	schema := readCourseSchemaMigration(t)

	requiredFragments := []string{
		"create table courses",
		"id uuid primary key",
		"title text not null",
		"slug text not null",
		"description text not null default ''",
		"instructor_id uuid not null",
		"status text not null default 'draft'",
		"created_at timestamptz not null default now()",
		"updated_at timestamptz not null default now()",
		"constraint courses_slug_key unique (slug)",
		"constraint courses_status_check check (status in ('draft', 'published'))",
		"create table lessons",
		"course_id uuid not null",
		"content text not null default ''",
		`"order" integer not null default 0`,
		"foreign key (course_id)",
		"references courses(id)",
		"on delete cascade",
		`constraint lessons_order_check check ("order" >= 0)`,
		`create index lessons_course_order_idx on lessons (course_id, "order")`,
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("expected schema migration to contain %q", fragment)
		}
	}
}

func TestCourseSchemaMigrationDoesNotOwnIdentitySchema(t *testing.T) {
	schema := readCourseSchemaMigration(t)

	if strings.Contains(schema, "references users") {
		t.Fatalf("course schema must not own identity-context user relations")
	}
}

func TestContentBlocksMigrationMatchesRepository(t *testing.T) {
	schema := readContentBlocksMigration(t)

	requiredFragments := []string{
		"create table content_blocks",
		"id uuid primary key",
		"lesson_id uuid not null",
		"kind text not null",
		"position integer not null",
		"text_markdown text",
		"video_provider text",
		"video_locator text",
		"video_caption text",
		"foreign key (lesson_id)",
		"references lessons(id)",
		"on delete cascade",
		"constraint content_blocks_kind_check check (kind in ('text', 'video'))",
		"constraint content_blocks_position_check check (position >= 0)",
		"constraint content_blocks_lesson_position_key unique (lesson_id, position)",
		"create index content_blocks_lesson_position_idx on content_blocks (lesson_id, position)",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("expected content blocks migration to contain %q", fragment)
		}
	}
}

func TestContentBlocksMigrationBackfillsLegacyLessonContent(t *testing.T) {
	schema := readContentBlocksMigration(t)

	requiredFragments := []string{
		"insert into content_blocks",
		"select",
		"id,",
		"'text'",
		"0",
		"content",
		"from lessons",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("expected content blocks backfill to contain %q", fragment)
		}
	}

	if strings.Contains(schema, "drop column content") {
		t.Fatalf("content blocks migration must leave lessons.content as rollback safety")
	}
}

func TestContentBlocksDownMigrationDropsTable(t *testing.T) {
	contents, err := os.ReadFile("../../../../migrations/000002_add_content_blocks.down.sql")
	if err != nil {
		t.Fatalf("expected content blocks down migration to be readable, got %v", err)
	}

	if !strings.Contains(strings.ToLower(string(contents)), "drop table if exists content_blocks") {
		t.Fatalf("expected down migration to drop content_blocks")
	}
}

func TestQuizMigrationMatchesRepositories(t *testing.T) {
	schema := readQuizMigration(t)

	requiredFragments := []string{
		"create table quizzes",
		"id uuid primary key",
		"course_id uuid not null",
		"title text not null",
		"pass_threshold double precision not null default 0.7",
		"references courses(id)",
		"on delete cascade",
		"constraint quizzes_pass_threshold_check",
		"create table quiz_questions",
		"quiz_id uuid not null",
		"type text not null",
		"prompt text not null",
		"options jsonb not null",
		"correct_indices jsonb not null",
		"explanation text not null default ''",
		"position integer not null",
		"references quizzes(id)",
		"constraint quiz_questions_type_check",
		"check (type in ('single', 'multiple'))",
		"constraint quiz_questions_position_check",
		"check (position >= 0)",
		"constraint quiz_questions_position_unique",
		"unique (quiz_id, position)",
		"create index quiz_questions_quiz_position_idx",
		"drop constraint content_blocks_kind_check",
		"check (kind in ('text', 'video', 'quiz'))",
		"add column quiz_ref uuid",
		"references quizzes(id)",
		"on delete restrict",
		"create index content_blocks_quiz_ref_idx",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("expected quiz migration to contain %q", fragment)
		}
	}
}

func TestQuizDownMigrationReversesContentBlockChangesBeforeDroppingQuizzes(t *testing.T) {
	contents, err := os.ReadFile("../../../../migrations/000003_add_quizzes.down.sql")
	if err != nil {
		t.Fatalf("expected quiz down migration to be readable, got %v", err)
	}

	schema := strings.ToLower(string(contents))
	requiredFragments := []string{
		"drop index if exists content_blocks_quiz_ref_idx",
		"drop constraint if exists content_blocks_quiz_ref_fkey",
		"drop column if exists quiz_ref",
		"check (kind in ('text', 'video'))",
		"drop table if exists quiz_questions",
		"drop table if exists quizzes",
		"rollback",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("expected quiz down migration to contain %q", fragment)
		}
	}
	if strings.Index(schema, "drop table if exists quiz_questions") > strings.Index(schema, "drop table if exists quizzes") {
		t.Fatalf("expected quiz_questions to drop before quizzes")
	}
}

func TestPracticeMigrationMatchesRepositories(t *testing.T) {
	schema := readPracticeMigration(t)

	requiredFragments := []string{
		"create table practices",
		"id uuid primary key",
		"course_id uuid not null",
		"title text not null",
		"language text not null",
		"prompt text not null",
		"starter_code text not null default ''",
		"solution text not null default ''",
		"references courses(id)",
		"on delete cascade",
		"constraint practices_language_check",
		"check (language in ('javascript', 'typescript', 'golang', 'rust'))",
		"create table practice_test_cases",
		"practice_id uuid not null",
		"stdin text not null default ''",
		"expected_stdout text not null default ''",
		"name text not null default ''",
		"position integer not null",
		"references practices(id)",
		"constraint practice_test_cases_position_check",
		"check (position >= 0)",
		"constraint practice_test_cases_position_unique",
		"unique (practice_id, position)",
		"create index practice_test_cases_practice_position_idx",
		"drop constraint content_blocks_kind_check",
		"check (kind in ('text', 'video', 'quiz', 'practice'))",
		"add column practice_ref uuid",
		"references practices(id)",
		"on delete restrict",
		"create index content_blocks_practice_ref_idx",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("expected practice migration to contain %q", fragment)
		}
	}
}

func TestPracticeDownMigrationReversesContentBlockChangesBeforeDroppingPractices(t *testing.T) {
	contents, err := os.ReadFile("../../../../migrations/000004_add_practices.down.sql")
	if err != nil {
		t.Fatalf("expected practice down migration to be readable, got %v", err)
	}

	schema := strings.ToLower(string(contents))
	requiredFragments := []string{
		"drop index if exists content_blocks_practice_ref_idx",
		"drop constraint if exists content_blocks_practice_ref_fkey",
		"drop column if exists practice_ref",
		"check (kind in ('text', 'video', 'quiz'))",
		"drop table if exists practice_test_cases",
		"drop table if exists practices",
		"rollback",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("expected practice down migration to contain %q", fragment)
		}
	}
	if strings.Index(schema, "drop table if exists practice_test_cases") > strings.Index(schema, "drop table if exists practices") {
		t.Fatalf("expected practice_test_cases to drop before practices")
	}
}

func TestTestMigrationMatchesRepository(t *testing.T) {
	schema := readTestMigration(t)

	requiredFragments := []string{
		"create table tests",
		"id uuid primary key",
		"course_id uuid not null",
		"title text not null",
		"time_limit_minutes integer",
		"pass_threshold double precision not null default 0.7",
		"solution_zip_provider text",
		"solution_zip_locator text",
		"solution_video_provider text",
		"solution_video_locator text",
		"solution_video_caption text",
		"references courses(id)",
		"on delete cascade",
		"constraint tests_time_limit_minutes_check",
		"check (time_limit_minutes is null or time_limit_minutes > 0)",
		"constraint tests_pass_threshold_check",
		"check (pass_threshold >= 0 and pass_threshold <= 1)",
		"constraint tests_solution_group_check",
		"create table test_items",
		"test_id uuid not null",
		"kind text not null",
		"position integer not null",
		"choice_type text",
		"choice_prompt text",
		"choice_options jsonb",
		"choice_correct_indices jsonb",
		"choice_explanation text",
		"coding_language text",
		"coding_prompt text",
		"starter_code text",
		"coding_solution text",
		"coding_test_cases jsonb",
		"references tests(id)",
		"constraint test_items_kind_check",
		"check (kind in ('choice', 'coding'))",
		"constraint test_items_position_check",
		"check (position >= 0)",
		"constraint test_items_body_shape_check",
		"constraint test_items_position_unique",
		"unique (test_id, position)",
		"create index test_items_test_position_idx",
		"on test_items (test_id, position)",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("expected test migration to contain %q", fragment)
		}
	}
}

func TestTestDownMigrationDropsItemsBeforeTests(t *testing.T) {
	contents, err := os.ReadFile("../../../../migrations/000005_add_tests.down.sql")
	if err != nil {
		t.Fatalf("expected test down migration to be readable, got %v", err)
	}

	schema := strings.ToLower(string(contents))
	requiredFragments := []string{
		"drop table if exists test_items",
		"drop table if exists tests",
	}

	for _, fragment := range requiredFragments {
		if !strings.Contains(schema, fragment) {
			t.Fatalf("expected test down migration to contain %q", fragment)
		}
	}
	if strings.Index(schema, "drop table if exists test_items") > strings.Index(schema, "drop table if exists tests") {
		t.Fatalf("expected test_items to drop before tests")
	}
}

func readCourseSchemaMigration(t *testing.T) string {
	t.Helper()

	contents, err := os.ReadFile("../../../../migrations/000001_create_course_schema.up.sql")
	if err != nil {
		t.Fatalf("expected course schema migration to be readable, got %v", err)
	}

	return strings.ToLower(string(contents))
}

func readContentBlocksMigration(t *testing.T) string {
	t.Helper()

	contents, err := os.ReadFile("../../../../migrations/000002_add_content_blocks.up.sql")
	if err != nil {
		t.Fatalf("expected content blocks migration to be readable, got %v", err)
	}

	return strings.ToLower(string(contents))
}

func readQuizMigration(t *testing.T) string {
	t.Helper()

	contents, err := os.ReadFile("../../../../migrations/000003_add_quizzes.up.sql")
	if err != nil {
		t.Fatalf("expected quiz migration to be readable, got %v", err)
	}

	return strings.ToLower(string(contents))
}

func readPracticeMigration(t *testing.T) string {
	t.Helper()

	contents, err := os.ReadFile("../../../../migrations/000004_add_practices.up.sql")
	if err != nil {
		t.Fatalf("expected practice migration to be readable, got %v", err)
	}

	return strings.ToLower(string(contents))
}

func readTestMigration(t *testing.T) string {
	t.Helper()

	contents, err := os.ReadFile("../../../../migrations/000005_add_tests.up.sql")
	if err != nil {
		t.Fatalf("expected test migration to be readable, got %v", err)
	}

	return strings.ToLower(string(contents))
}
