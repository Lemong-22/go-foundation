# Course CLI Guide

This guide explains how to install, configure, and use `course-cli` to manage
courses, lessons, lesson blocks, quizzes, practices, tests, imports, migrations,
the loopback playground, and the REST server.

The CLI is the scriptable management interface for the course bounded context.
It writes directly to PostgreSQL through the same use cases used by REST and the
browser playground.

## Prerequisites

- Go 1.22 or newer
- PostgreSQL
- A migrated course database
- An instructor UUID for commands that create or import instructor-owned content

## Build Or Run

Run from source:

```sh
go run ./cmd/course-cli --help
go run ./cmd/course-cli course list --output table
```

Build a local binary:

```sh
mkdir -p bin
go build -o bin/course-cli ./cmd/course-cli
./bin/course-cli --help
```

The examples below use `course-cli`. If you are running from source, replace
`course-cli` with `go run ./cmd/course-cli`.

## Configuration

Most commands need a database URL:

```sh
export COURSE_CLI_DB_URL="postgres://postgres:password@127.0.0.1:5432/entropy_course?sslmode=disable"
```

Create and import commands also need an instructor id unless it is supplied by a
command flag:

```sh
export COURSE_CLI_INSTRUCTOR_ID="550e8400-e29b-41d4-a716-446655440010"
```

The REST server uses a bearer token:

```sh
export COURSE_CLI_API_TOKEN="local-dev-token"
```

You can also store config in `~/.config/course-cli/config.yaml`:

```yaml
db_url: postgres://postgres:password@127.0.0.1:5432/entropy_course?sslmode=disable
instructor_id: 550e8400-e29b-41d4-a716-446655440010
api_token: local-dev-token
```

Config precedence is: CLI flags, environment variables, config file, defaults.
The root flags are:

```sh
course-cli --db-url "$COURSE_CLI_DB_URL" --api-token "$COURSE_CLI_API_TOKEN" ...
```

## Database Migrations

Apply all pending migrations:

```sh
course-cli migrate up
```

Rollback the latest applied migration:

```sh
course-cli migrate down
```

Show applied and pending migrations:

```sh
course-cli migrate status
```

## Output Formats

Read commands generally support:

- `--output table` or `-o table` for human-readable tables
- `--output json` or `-o json` for automation
- `--output quiet` or `-o quiet` for id-only output

Create and update commands print the affected id. Delete, reorder, publish, and
unpublish commands print a short confirmation message.

## Destructive Commands

Delete and remove commands prompt for confirmation by default. Use `--force` in
scripts:

```sh
course-cli lesson delete "$LESSON_ID" --force
course-cli quiz question remove "$QUESTION_ID" --force
```

Quiz and practice deletion can fail if the item is embedded in one or more
lessons. Remove the lesson block references first.

## Course Commands

Create a course:

```sh
course-cli course create \
  --title "Intro to Go" \
  --slug intro-to-go \
  --description "Learn Go from first principles"
```

Use `--instructor-id` to override configured instructor id:

```sh
course-cli course create \
  --title "Intro to Go" \
  --slug intro-to-go \
  --instructor-id 550e8400-e29b-41d4-a716-446655440010
```

Common operations:

```sh
course-cli course list --status draft --output table
course-cli course get "$COURSE_ID" --output json
course-cli course update "$COURSE_ID" --title "Advanced Go"
course-cli course publish "$COURSE_ID"
course-cli course unpublish "$COURSE_ID"
course-cli course delete "$COURSE_ID" --force
```

Course commands:

| Command | Purpose |
| --- | --- |
| `course create --title <title> --slug <slug> [--description <text>] [--instructor-id <uuid>]` | Create a draft course. |
| `course list [--status draft\|published] [-o table\|json\|quiet]` | List courses. |
| `course get <course-id> [-o table\|json\|quiet]` | Show one course. |
| `course update <course-id> [--title <title>] [--description <text>] [--slug <slug>]` | Update supplied fields. |
| `course publish <course-id>` | Publish a draft course. |
| `course unpublish <course-id>` | Move a course back to draft. |
| `course delete <course-id> [--force]` | Delete a course. |

## Lesson Commands

Create and manage lessons inside a course:

```sh
course-cli lesson create --course-id "$COURSE_ID" --title "Setup" --order 0
course-cli lesson list --course-id "$COURSE_ID" --output table
course-cli lesson get "$LESSON_ID" --output json
course-cli lesson update "$LESSON_ID" --title "Workspace Setup"
course-cli lesson reorder --course-id "$COURSE_ID" --order "$LESSON_A:0,$LESSON_B:1"
course-cli lesson delete "$LESSON_ID" --force
```

Lesson commands:

| Command | Purpose |
| --- | --- |
| `lesson create --course-id <id> --title <title> [--order <int>]` | Create a lesson. |
| `lesson list --course-id <id> [-o table\|json\|quiet]` | List lessons in order. |
| `lesson get <lesson-id> [-o table\|json\|quiet]` | Show one lesson. |
| `lesson update <lesson-id> [--title <title>]` | Rename a lesson. |
| `lesson reorder --course-id <id> --order <lesson-id:position,...>` | Replace lesson order. |
| `lesson delete <lesson-id> [--force]` | Delete a lesson. |

Positions are zero-based.

## Lesson Block Commands

Lessons are built from ordered blocks. Supported block kinds are `text`,
`video`, `quiz`, and `practice`.

Add a text block:

```sh
course-cli lesson block add \
  --lesson-id "$LESSON_ID" \
  --kind text \
  --text "## Welcome" \
  --position 0
```

Add a video block:

```sh
course-cli lesson block add \
  --lesson-id "$LESSON_ID" \
  --kind video \
  --video-provider youtube \
  --video-locator dQw4w9WgXcQ \
  --video-caption "Intro walkthrough"
```

Embed a quiz or practice:

```sh
course-cli lesson block add --lesson-id "$LESSON_ID" --kind quiz --quiz-id "$QUIZ_ID"
course-cli lesson block add --lesson-id "$LESSON_ID" --kind practice --practice-id "$PRACTICE_ID"
```

Manage blocks:

```sh
course-cli lesson block list --lesson-id "$LESSON_ID" --output table
course-cli lesson block get "$BLOCK_ID" --output json
course-cli lesson block update "$BLOCK_ID" --text "Updated markdown"
course-cli lesson block reorder --lesson-id "$LESSON_ID" --order "$BLOCK_A:0,$BLOCK_B:1"
course-cli lesson block remove "$BLOCK_ID" --force
```

## Quiz Commands

Create a quiz:

```sh
course-cli quiz create \
  --course-id "$COURSE_ID" \
  --title "Go Basics Quiz" \
  --pass-threshold 0.8
```

Manage quiz metadata:

```sh
course-cli quiz list --course-id "$COURSE_ID" --output table
course-cli quiz get "$QUIZ_ID" --output json
course-cli quiz update "$QUIZ_ID" --title "Updated Quiz" --pass-threshold 0.75
course-cli quiz delete "$QUIZ_ID" --force
```

Quiz commands:

| Command | Purpose |
| --- | --- |
| `quiz create --course-id <id> --title <title> [--pass-threshold <0-1>]` | Create a quiz. |
| `quiz list --course-id <id> [-o table\|json\|quiet]` | List quizzes. |
| `quiz get <quiz-id> [-o table\|json\|quiet]` | Show quiz detail and questions. |
| `quiz update <quiz-id> [--title <title>] [--pass-threshold <0-1>]` | Update supplied fields. |
| `quiz delete <quiz-id> [--force]` | Delete a quiz not embedded in lessons. |

## Quiz Question Commands

Question types are `single` and `multiple`. Options are supplied by repeating
`--option`. Correct answers are zero-based option indexes.

Add a single-choice question:

```sh
course-cli quiz question add \
  --quiz-id "$QUIZ_ID" \
  --type single \
  --prompt "Which keyword starts a goroutine?" \
  --option go \
  --option defer \
  --correct 0 \
  --explanation "go starts a new goroutine"
```

Add a multiple-choice question:

```sh
course-cli quiz question add \
  --quiz-id "$QUIZ_ID" \
  --type multiple \
  --prompt "Pick the built-in Go types" \
  --option string \
  --option banana \
  --option int \
  --correct 0,2
```

Manage questions:

```sh
course-cli quiz question list --quiz-id "$QUIZ_ID" --output table
course-cli quiz question get "$QUESTION_ID" --output json
course-cli quiz question update "$QUESTION_ID" --prompt "Updated prompt" --correct 1
course-cli quiz question reorder --quiz-id "$QUIZ_ID" --order "$QUESTION_A:0,$QUESTION_B:1"
course-cli quiz question remove "$QUESTION_ID" --force
```

## Practice Commands

Practice languages are `javascript`, `typescript`, `golang`, and `rust`.

Create a coding practice:

```sh
course-cli practice create \
  --course-id "$COURSE_ID" \
  --title "FizzBuzz" \
  --language golang \
  --prompt "Print FizzBuzz" \
  --starter-code "package main" \
  --solution "package main"
```

Manage practices:

```sh
course-cli practice list --course-id "$COURSE_ID" --output table
course-cli practice get "$PRACTICE_ID" --output json
course-cli practice update "$PRACTICE_ID" --title "Updated FizzBuzz" --prompt "Updated prompt"
course-cli practice delete "$PRACTICE_ID" --force
```

Practice commands:

| Command | Purpose |
| --- | --- |
| `practice create --course-id <id> --title <title> --language <language> --prompt <text> [--starter-code <code>] [--solution <code>]` | Create a coding practice. |
| `practice list --course-id <id> [-o table\|json\|quiet]` | List practices. |
| `practice get <practice-id> [-o table\|json\|quiet]` | Show practice detail and test cases. |
| `practice update <practice-id> [--title <title>] [--prompt <text>] [--starter-code <code>] [--solution <code>]` | Update supplied fields. |
| `practice delete <practice-id> [--force]` | Delete a practice not embedded in lessons. |

## Practice Test Case Commands

Add test cases to a practice:

```sh
course-cli practice testcase add \
  --practice-id "$PRACTICE_ID" \
  --stdin "3" \
  --expected-stdout "Fizz" \
  --name "multiple of three" \
  --position 0
```

Manage practice test cases:

```sh
course-cli practice testcase list --practice-id "$PRACTICE_ID" --output table
course-cli practice testcase get "$TESTCASE_ID" --output json
course-cli practice testcase update "$TESTCASE_ID" --stdin "5" --expected-stdout "Buzz"
course-cli practice testcase reorder --practice-id "$PRACTICE_ID" --order "$CASE_A:0,$CASE_B:1"
course-cli practice testcase remove "$TESTCASE_ID" --force
```

`--stdin`, `--expected-stdout`, and `--name` may be empty when intentionally
supplied.

## Test Commands

Create a test:

```sh
course-cli test create \
  --course-id "$COURSE_ID" \
  --title "Final Test" \
  --time-limit-minutes 45 \
  --pass-threshold 0.8
```

Manage tests:

```sh
course-cli test list --course-id "$COURSE_ID" --output table
course-cli test get "$TEST_ID" --output json
course-cli test update "$TEST_ID" --title "Updated Final Test" --time-limit-minutes 60
course-cli test delete "$TEST_ID" --force
```

Attach solution references:

```sh
course-cli test update "$TEST_ID" \
  --solution-zip-provider url \
  --solution-zip-locator https://example.com/solution.zip \
  --solution-video-provider url \
  --solution-video-locator https://example.com/video.mp4 \
  --solution-video-caption "Walkthrough"
```

Test commands:

| Command | Purpose |
| --- | --- |
| `test create --course-id <id> --title <title> [--time-limit-minutes <int>] [--pass-threshold <0-1>]` | Create a test. |
| `test list --course-id <id> [-o table\|json\|quiet]` | List tests. |
| `test get <test-id> [-o table\|json\|quiet]` | Show test detail and items. |
| `test update <test-id> [flags]` | Update metadata or solution references. |
| `test delete <test-id> [--force]` | Delete a test. |

## Test Item Commands

Test items can be `choice` or `coding`.

Add a choice item:

```sh
course-cli test item add \
  --test-id "$TEST_ID" \
  --kind choice \
  --prompt "Pick two" \
  --type multiple \
  --option A \
  --option B \
  --correct 0 \
  --correct 1 \
  --explanation "A and B"
```

Add a coding item. Coding test cases use
`stdin::expected_stdout[::name]` and can be repeated:

```sh
course-cli test item add \
  --test-id "$TEST_ID" \
  --kind coding \
  --prompt "Print hello" \
  --language javascript \
  --starter-code "console.log('')" \
  --solution "console.log('hello')" \
  --testcase "::hello::sample"
```

Manage test items:

```sh
course-cli test item list --test-id "$TEST_ID" --output table
course-cli test item get "$ITEM_ID" --output json
course-cli test item update "$ITEM_ID" --prompt "Updated prompt"
course-cli test item reorder --test-id "$TEST_ID" --order "$ITEM_A:0,$ITEM_B:1"
course-cli test item remove "$ITEM_ID" --force
```

## Course Import Workflow

The import workflow is intended for agents and automation. The external agent
normalizes source material into the documented v1 zip format, then the CLI
plans, resolves, and applies that zip.

Read the full zip format and plan schema in [import-format.md](import-format.md).

Plan an import:

```sh
course-cli import plan course.zip -o json --output plan.json
```

If there are no conflicts, apply the zip:

```sh
course-cli import apply course.zip --force
```

If there are conflicts, resolve them outside the CLI by converting each
`conflicts[]` entry into an `operations[]` entry with `kind` set to `update` or
`skip`, the selected existing `target_id`, the original `payload`, and
`conflicts: []`. Do not edit payload values inside the plan JSON. To create a
distinct entity after a collision, change the zip identity field and run
planning again.

Apply a resolved plan:

```sh
course-cli import apply course.zip --resolved-plan resolved-plan.json --force
```

Use a batch conflict strategy only when it is acceptable to treat every
unresolved conflict the same way:

```sh
course-cli import apply course.zip --conflict-strategy fail --force
course-cli import apply course.zip --conflict-strategy skip --force
course-cli import apply course.zip --conflict-strategy update --force
```

Import commands:

| Command | Purpose |
| --- | --- |
| `import plan <zip-path> [-o json\|table] [--output <file>] [--instructor-id <uuid>]` | Parse a course zip and produce an import plan. Default output is JSON. |
| `import apply <zip-path> [--resolved-plan <file>] [--conflict-strategy fail\|skip\|update] [-o table\|json] [--force] [--instructor-id <uuid>]` | Apply a fresh or resolved import plan. |

## Playground

The playground serves a loopback-only browser UI that runs the same Cobra
command handlers as the CLI:

```sh
course-cli playground
```

Open:

```text
http://127.0.0.1:8787
```

Use a different local port:

```sh
course-cli playground --addr 127.0.0.1:8790
```

The playground includes a command runner and an upload-based import console.

## REST Server

Start the REST API:

```sh
course-cli rest
```

Use a different bind address:

```sh
course-cli rest --addr 127.0.0.1:8080
```

The REST server uses the configured `api_token` or `COURSE_CLI_API_TOKEN` as its
bearer token. REST import endpoints accept multipart uploads; see
[import-format.md](import-format.md) for the import request shapes.

## Troubleshooting

- `missing database url`: set `COURSE_CLI_DB_URL`, add `db_url` to config, or
  pass `--db-url`.
- `instructor id is required`: set `COURSE_CLI_INSTRUCTOR_ID`, add
  `instructor_id` to config, or pass `--instructor-id` where supported.
- `unsupported output format`: use `table`, `json`, or `quiet` for read
  commands. Import plan supports `json` and `table`; import apply supports
  `table` and `json`.
- `confirmation declined`: rerun with `--force` for intentional destructive
  operations in scripts.
- `quiz is embedded in one or more lessons` or `practice is embedded in one or
  more lessons`: remove the embedding lesson block before deleting the quiz or
  practice.
- `unresolved import conflicts`: resolve the plan JSON or use an explicit
  conflict strategy.
