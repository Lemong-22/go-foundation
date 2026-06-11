# Course Import Format

This document is the v1 import contract for humans and external agents. It
documents the zip package accepted by `ZipImportSource`, the JSON plan emitted by
`PlanImport`, and the apply result returned by CLI, REST, and console adapters.

The AI agent runtime is outside this repository. The agent can produce zip files,
run plan/apply entrypoints, and choose conflict resolutions. It must not depend
on an in-process Go agent, hidden persistence, or direct payload edits in a plan.

## ZIP Layout

One zip contains exactly one course:

```text
course.zip
|-- format_version.txt
|-- course.yaml
|-- lessons/
|   |-- 01-foundations.md
|   `-- 02-syntax.md
|-- quizzes/
|   `-- foundations-quiz.yaml
|-- practices/
|   `-- fizzbuzz.yaml
`-- tests/
    `-- midterm.yaml
```

Required files:

- `format_version.txt`
- `course.yaml`

Optional directories:

- `lessons/` with `*.md` files
- `quizzes/` with `*.yaml` files
- `practices/` with `*.yaml` files
- `tests/` with `*.yaml` files

Path rules:

- Paths are normalized to slash separators before validation.
- Directories are ignored; files must match the allowlist above.
- Absolute paths, `..` traversal, duplicate normalized paths, and unexpected
  files are layout errors.
- YAML files use `.yaml`, not `.yml`.
- YAML is decoded with known fields enabled, so unknown keys are parse errors.
- Files inside each supported directory are parsed in lexicographic path order.

The zip hash in a plan is a canonical SHA-256 digest over sorted normalized file
paths and file bytes. Zip entry order and compression metadata do not affect the
hash. Changing any accepted file path or byte requires a new plan.

## `format_version.txt`

`format_version.txt` contains the supported format version as plain text:

```text
1
```

The parser trims surrounding whitespace. Empty content is a layout error.
Anything other than `1` is an unsupported import format.

## `course.yaml`

```yaml
title: Intro to Go
slug: intro-to-go
description: A practical Go foundations course.
status: draft
```

Fields:

| Field | Required | Notes |
| --- | --- | --- |
| `title` | yes | Non-empty after trimming. |
| `slug` | yes | Lowercase letters, numbers, and single hyphens only. This is the persisted course slug. |
| `description` | no | Free text. Empty is allowed. |
| `status` | yes | `draft` or `published`. |

`instructor_id` is not part of the zip. CLI reads it from config or
`--instructor-id`; REST and console read it from runtime configuration.

## `lessons/*.md`

Lesson files are Markdown with YAML frontmatter. The frontmatter defines lesson
metadata and ordered blocks. The body is optional and can provide the Markdown
for a single text block that omits `markdown`.

```markdown
---
title: Foundations
order: 0
blocks:
  - kind: text
    position: 0
    markdown: |
      Welcome to the course.
  - kind: video
    position: 1
    video_provider: youtube
    video_locator: dQw4w9WgXcQ
    video_caption: Overview
  - kind: quiz
    position: 2
    quiz_ref: foundations-quiz
  - kind: practice
    position: 3
    practice_ref: fizzbuzz
---
```

Fields:

| Field | Required | Notes |
| --- | --- | --- |
| `title` | yes | Non-empty after trimming. Lessons match existing lessons by title within the course. |
| `order` | no | Zero-based lesson order. If omitted, file order index is used. |
| `blocks` | yes | Must contain at least one block. |

Block fields:

| Field | Required | Notes |
| --- | --- | --- |
| `kind` | yes | `text`, `video`, `quiz`, or `practice`. |
| `position` | no | Zero-based block position. If omitted, block list index is used. |
| `markdown` | text | Required for `text` unless the Markdown body after frontmatter is non-empty. |
| `video_provider` | video | Required for `video`; `url`, `youtube`, or `mux`. |
| `video_locator` | video | Required for `video`; validated by the selected provider. |
| `video_caption` | no | Optional video caption. |
| `quiz_ref` | quiz | Required for `quiz`; must reference a quiz slug in `quizzes/*.yaml`. |
| `practice_ref` | practice | Required for `practice`; must reference a practice slug in `practices/*.yaml`. |

Tests are course-level only in v1. There is no lesson block kind for tests.

## `quizzes/*.yaml`

```yaml
slug: foundations-quiz
title: Foundations Quiz
pass_threshold: 0.8
questions:
  - type: single
    position: 0
    prompt: Which command runs Go tests?
    options:
      - go test ./...
      - go fmt ./...
    correct_indices: [0]
    explanation: `go test ./...` runs tests for all packages.
```

Fields:

| Field | Required | Notes |
| --- | --- | --- |
| `slug` | yes | Import-local quiz identifier; must be unique among quizzes in the zip. |
| `title` | yes | Non-empty after trimming. Quizzes match existing quizzes by title within the course. |
| `pass_threshold` | no | Number from `0` through `1`. Defaults to domain behavior when omitted. |
| `questions` | no | Ordered list of choice questions. |

Question fields:

| Field | Required | Notes |
| --- | --- | --- |
| `type` | yes | `single` or `multiple`. |
| `position` | no | Zero-based question position. If omitted, list index is used. |
| `prompt` | yes | Non-empty after trimming. |
| `options` | yes | Must contain at least one entry at parse time; domain validation may require more. |
| `correct_indices` | yes for valid apply | Zero-based option indexes. Domain validation enforces choice rules. |
| `explanation` | no | Free text. |

## `practices/*.yaml`

```yaml
slug: fizzbuzz
title: FizzBuzz
language: golang
prompt: Print numbers with FizzBuzz substitutions.
starter_code: |
  package main
solution: |
  package main
test_cases:
  - position: 0
    name: sample
    stdin: "15\n"
    expected_stdout: "1\n2\nFizz\n"
```

Fields:

| Field | Required | Notes |
| --- | --- | --- |
| `slug` | yes | Import-local practice identifier; must be unique among practices in the zip. |
| `title` | yes | Non-empty after trimming. Practices match existing practices by title within the course. |
| `language` | yes | `javascript`, `typescript`, `golang`, or `rust`. |
| `prompt` | yes | Non-empty after trimming. |
| `starter_code` | no | Free text. |
| `solution` | no | Free text. |
| `test_cases` | no | Ordered list of practice test cases. |

Test case fields:

| Field | Required | Notes |
| --- | --- | --- |
| `position` | no | Zero-based test case position. If omitted, list index is used. |
| `stdin` | no | Free text. |
| `expected_stdout` | yes | Must not be empty. |
| `name` | no | Optional label. |

## `tests/*.yaml`

```yaml
slug: midterm
title: Midterm
time_limit_minutes: 45
pass_threshold: 0.75
solution:
  zip_provider: url
  zip_locator: https://example.com/solutions/midterm.zip
  video_provider: youtube
  video_locator: dQw4w9WgXcQ
  video_caption: Walkthrough
items:
  - kind: choice
    position: 0
    prompt: Pick one.
    choice_type: single
    options: [A, B]
    correct_indices: [0]
    explanation: A is correct.
  - kind: coding
    position: 1
    coding_prompt: Implement FizzBuzz.
    language: golang
    starter_code: |
      package main
    solution: |
      package main
    test_cases:
      - stdin: "3\n"
        expected_stdout: "1\n2\nFizz\n"
        name: sample
```

Fields:

| Field | Required | Notes |
| --- | --- | --- |
| `slug` | yes | Import-local test identifier; must be unique among tests in the zip. |
| `title` | yes | Non-empty after trimming. Tests match existing tests by title within the course. |
| `time_limit_minutes` | no | Positive integer when set; domain validation runs during planning/apply. |
| `pass_threshold` | no | Number from `0` through `1`. Defaults to domain behavior when omitted. |
| `solution` | no | Optional solution package metadata. |
| `items` | no | Ordered list of test items. |

Solution fields:

| Field | Required | Notes |
| --- | --- | --- |
| `zip_provider` | yes when `solution` is present | `url`, `youtube`, or `mux`; normally `url` for zip packages. |
| `zip_locator` | yes when `solution` is present | Valid locator for `zip_provider`. |
| `video_provider` | yes when `solution` is present | `url`, `youtube`, or `mux`. |
| `video_locator` | yes when `solution` is present | Valid locator for `video_provider`. |
| `video_caption` | no | Optional explanation caption. |

Test item fields:

| Field | Required | Notes |
| --- | --- | --- |
| `kind` | yes | `choice` or `coding`. |
| `position` | no | Zero-based item position. If omitted, list index is used. |
| `prompt` | choice | Non-empty choice prompt. |
| `choice_type` | choice | `single` or `multiple`. |
| `options` | choice | Choice options. |
| `correct_indices` | choice | Zero-based option indexes. |
| `explanation` | no | Free text for choice items. |
| `coding_prompt` | coding | Non-empty coding prompt. |
| `language` | coding | `javascript`, `typescript`, `golang`, or `rust`. |
| `starter_code` | no | Free text for coding items. |
| `solution` | no | Free text for coding items. |
| `test_cases` | no | Embedded coding test cases with `stdin`, `expected_stdout`, and `name`. |

## Slugs And References

- The course `slug` is persisted and is the course identity used by planning.
- Quiz, practice, and test `slug` fields are import-local only. They do not add
  domain slugs to those aggregates.
- Quiz slugs must be unique within `quizzes/*.yaml`.
- Practice slugs must be unique within `practices/*.yaml`.
- Test slugs must be unique within `tests/*.yaml`.
- Lesson `quiz_ref` values must match a quiz slug in the same zip.
- Lesson `practice_ref` values must match a practice slug in the same zip.
- Lesson titles are used as import-local refs for lesson operations.
- Question, block, test case, and test item refs are based on their parent ref
  plus resolved position.

Planning compares imported content to existing state this way:

| Entity | Existing match key | Conflict reason when content differs |
| --- | --- | --- |
| Course | `slug` | `slug_collision` |
| Lesson | title inside course | `title_in_parent_collision` |
| Quiz | title inside course | `title_in_parent_collision` |
| Practice | title inside course | `title_in_parent_collision` |
| Test | title inside course | `title_in_parent_collision` |
| Block | position inside lesson | `position_collision` |
| Question | position inside quiz | `position_collision` |
| Test case | position inside practice | `position_collision` |
| Test item | position inside test | `position_collision` |

If an agent wants a distinct new entity instead of updating or skipping a
collision, it must change the identity field in the zip, such as a course slug,
quiz title, lesson title, or child position, and run planning again.

## Validation Errors

Import validation can fail before planning, during planning, or during apply.

Layout errors include:

- missing `format_version.txt`
- empty format version
- missing `course.yaml`
- unexpected file path
- unsafe file path
- duplicate normalized file path
- missing required fields
- duplicate quiz, practice, or test slug
- unknown `quiz_ref` or `practice_ref`
- unknown lesson block kind
- unknown test item kind

Parse errors include:

- invalid zip file
- unreadable zip entry
- malformed YAML
- unknown YAML key
- missing or unterminated lesson frontmatter

Planning and apply validation include:

- unsupported `format_version`
- invalid course slug or status
- invalid media provider or locator
- invalid pass threshold
- invalid language
- invalid choice question content
- invalid resolved plan JSON
- resolved plan hash mismatch
- unresolved conflicts when strategy is `fail`

REST maps validation, parse, layout, unsupported format, invalid strategy, and
hash mismatch errors to `400`; unresolved conflicts to `409`; partial apply
failures to `500`; and successful plan/apply responses to `200`.

## Plan JSON Schema

`PlanImport` emits an `ImportPlan` JSON document. `ApplyPlan` can consume the
same shape as `resolved_plan` once every conflict has been converted into an
operation and the `conflicts` array is empty.

```json
{
  "format_version": "1",
  "zip_hash": "64 lowercase hex characters",
  "generated_at": "2026-05-28T18:00:00Z",
  "operations": [],
  "conflicts": []
}
```

### `ImportPlan`

| Field | Type | Notes |
| --- | --- | --- |
| `format_version` | string | Copied from `format_version.txt`; v1 is `1`. |
| `zip_hash` | string | Lowercase SHA-256 hex digest of canonical zip content. |
| `generated_at` | RFC3339 timestamp | Created by the service clock. |
| `operations` | `ImportOperation[]` | Planned `create`, `update`, `noop`, or `skip` operations. |
| `conflicts` | `ImportConflict[]` | Decisions that must be resolved before apply with `fail`. |

### `ImportOperation`

```json
{
  "kind": "create",
  "entity_type": "quiz",
  "entity_ref": "quiz:foundations-quiz",
  "target_id": "existing-id-for-update-noop-or-skip",
  "payload": {}
}
```

| Field | Type | Notes |
| --- | --- | --- |
| `kind` | enum | `create`, `update`, `noop`, or `skip`. |
| `entity_type` | enum | `course`, `lesson`, `block`, `quiz`, `question`, `practice`, `test_case`, `test`, or `test_item`. |
| `entity_ref` | string | Import-local human-readable reference. |
| `target_id` | string | Omitted for `create`; required for `update`, `noop`, and `skip`. |
| `payload` | object | Entity-specific data generated from the zip. Treat as read-only. |

Payload edits require changing the zip and replanning. A resolved plan may copy
the conflict payload into a new operation, but it must not mutate payload
content.

### `ImportConflict`

```json
{
  "entity_type": "course",
  "entity_ref": "course:intro-to-go",
  "reason": "slug_collision",
  "candidates": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "description": "course 550e8400-e29b-41d4-a716-446655440000 \"Intro to Go\""
    }
  ],
  "recommended": "update",
  "payload": {}
}
```

| Field | Type | Notes |
| --- | --- | --- |
| `entity_type` | enum | Same values as `ImportOperation.entity_type`. |
| `entity_ref` | string | Same ref that the promoted operation should use. |
| `reason` | enum | `slug_collision`, `title_in_parent_collision`, or `position_collision`. |
| `candidates` | `ConflictCandidate[]` | Existing entities that matched the imported identity. At least one candidate is required. |
| `recommended` | enum | Planner recommendation. In v1 this is `update` for unresolved conflicts. |
| `payload` | object | Same payload shape as the corresponding operation. Treat as read-only. |

### `ConflictCandidate`

| Field | Type | Notes |
| --- | --- | --- |
| `id` | string | Existing entity ID. |
| `description` | string | Human-readable summary for review. |

### Resolved Plan Requirements

A resolved plan submitted to `import apply` or `POST /v1/import/apply` must:

- keep `format_version` and `zip_hash` from the plan generated for the same zip
- leave `generated_at` as a valid timestamp
- contain only valid `ImportOperation` entries
- have an empty `conflicts` array
- use only the closed `kind`, `entity_type`, and `reason` enums
- use `target_id` only where valid for the operation kind
- keep operation payloads copied from the original plan

To resolve a conflict, choose `update` or `skip` against the intended candidate.
To create a separate new entity, edit the zip identity field and re-run
`PlanImport`; do not invent a new payload inside the resolved plan.

## Payload JSON By Entity Type

Payloads are generated by the planner and consumed by `ApplyPlan`. Agents should
read them for context, but should not edit them in a resolved plan.

| Entity type | Payload fields |
| --- | --- |
| `course` | `title`, `slug`, `description`, `instructor_id`, `status` |
| `lesson` | `course_id`, `title`, `order` |
| `block` | `lesson_id`, `kind`, `markdown`, `video_provider`, `video_locator`, `video_caption`, `quiz_ref`, `practice_ref`, `position` |
| `quiz` | `course_id`, `slug`, `title`, `pass_threshold`, `import_local_id` |
| `question` | `quiz_id`, `type`, `prompt`, `options`, `correct_indices`, `explanation`, `position` |
| `practice` | `course_id`, `slug`, `title`, `language`, `prompt`, `starter_code`, `solution`, `import_local_id` |
| `test_case` | `practice_id`, `stdin`, `expected_stdout`, `name`, `position` |
| `test` | `course_id`, `slug`, `title`, `time_limit_minutes`, `pass_threshold`, `solution`, `import_local_id` |
| `test_item` | `test_id`, `kind`, `position`, `prompt`, `choice_type`, `options`, `correct_indices`, `explanation`, `coding_prompt`, `language`, `starter_code`, `solution`, `test_cases` |

Placeholder IDs begin with `$`, such as `$course:intro-to-go` or
`$quiz:foundations-quiz`. Apply resolves placeholders to real UUIDs as parent
aggregates are created.

## Apply Result JSON

CLI JSON output, REST, and console use this apply result shape:

```json
{
  "applied": [
    {
      "kind": "create",
      "entity_type": "course",
      "entity_ref": "course:intro-to-go",
      "target_id": "",
      "message": "created course 550e8400-e29b-41d4-a716-446655440000"
    }
  ],
  "failed": [
    {
      "kind": "update",
      "entity_type": "quiz",
      "entity_ref": "quiz:foundations-quiz",
      "target_id": "550e8400-e29b-41d4-a716-446655440001",
      "error": "validation failed"
    }
  ],
  "skipped": [
    {
      "kind": "noop",
      "entity_type": "lesson",
      "entity_ref": "lesson:Foundations",
      "target_id": "550e8400-e29b-41d4-a716-446655440002"
    }
  ],
  "counts": {
    "create": 1,
    "update": 1,
    "noop": 1,
    "skip": 0
  },
  "aggregates_succeeded": 2,
  "aggregates_failed": 1
}
```

`applied` entries include a message from the service call. `failed` entries
include an error string. `skipped` includes both `noop` and `skip` operations.
`counts` includes operations from all three arrays.

## Entry Points

CLI:

```sh
course-cli import plan course.zip -o json --output plan.json
course-cli import apply course.zip --resolved-plan resolved-plan.json --force
course-cli import apply course.zip --conflict-strategy fail --force
```

REST:

```text
POST /v1/import/plan
  multipart field: zip

POST /v1/import/apply?conflict_strategy=fail
  multipart fields: zip, resolved_plan (optional JSON file or form value)
```

Console:

- Upload a zip.
- Review operations and conflicts.
- Resolve conflicts client-side into the same resolved plan JSON shape.
- Apply through the same import service.

All entrypoints call the same `ImportService` port. Adapters do not own import
business rules.

## Future Work

These items are intentionally deferred and are not hidden v1 requirements:

- AI agent runtime, prompts, model selection, or hosting
- auth and permissions beyond existing single-instructor configuration
- source code compile/run validation for imported exercises
- import history, undo, rollback, or audit tables
- multi-course zip packages
- persistent domain slugs for Quiz, Practice, or Test
- new Phase E database migrations or import-specific persistence tables
