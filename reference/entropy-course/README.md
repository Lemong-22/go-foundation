# Entropy Course

Go CLI, loopback browser playground, REST API, and Bun/Next.js learner reader
for the `course` bounded context. The backend manages courses, lessons, lesson
blocks, quizzes, practices, tests, and course imports through the same hexagonal
use cases across CLI, playground, and REST adapters.

For command-level detail, see [docs/cli-guide.md](docs/cli-guide.md). For the
course zip format and import plan schema, see
[docs/import-format.md](docs/import-format.md). The web reader has its own
notes in [web/README.md](web/README.md).

## Prerequisites

- Go 1.22+
- PostgreSQL 15+
- Bun 1.3+ and Node.js 20.9+ for the `web/` reader
- Docker, if you use the local Postgres command below

## Configure

Start a local Postgres instance:

```sh
docker run --name entropy-course-postgres \
  -e POSTGRES_DB=entropy_course \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  -d postgres:17-alpine
```

Export the runtime configuration:

```sh
export COURSE_CLI_DB_URL="postgres://postgres:password@127.0.0.1:5432/entropy_course?sslmode=disable"
export COURSE_CLI_INSTRUCTOR_ID="550e8400-e29b-41d4-a716-446655440010"
export COURSE_CLI_API_TOKEN="dev-token"
```

Or store the same values at `~/.config/course-cli/config.yaml`:

```yaml
db_url: postgres://postgres:password@127.0.0.1:5432/entropy_course?sslmode=disable
instructor_id: 550e8400-e29b-41d4-a716-446655440010
api_token: dev-token
```

Apply all embedded migrations and check status:

```sh
go run ./cmd/course-cli migrate up
go run ./cmd/course-cli migrate status
```

## Run The CLI

Run from source:

```sh
go run ./cmd/course-cli --help
go run ./cmd/course-cli course list --output table
go run ./cmd/course-cli import plan courses/javascript-foundations.zip -o json --output plan.json
```

Build a local binary:

```sh
mkdir -p bin
go build -o bin/course-cli ./cmd/course-cli
./bin/course-cli lesson list --course-id <course-id> --output json
```

Useful runtime commands:

```sh
./bin/course-cli migrate up
./bin/course-cli migrate down
./bin/course-cli migrate status
```

## Run The Playground

The playground is loopback-only by default and exposes the bounded-context
course, lesson, block, quiz, practice, test, and import commands through a
browser UI. It does not launch runtime/admin commands such as `rest`,
`playground`, or migrations.

```sh
go run ./cmd/course-cli playground
```

Open:

```text
http://127.0.0.1:8787
```

Use a different local port if needed:

```sh
go run ./cmd/course-cli playground --addr 127.0.0.1:8790
```

## Run The REST API

The REST server uses `COURSE_CLI_API_TOKEN` or `api_token` from the config file
as a bearer token. It defaults to `127.0.0.1:8788`.

```sh
go run ./cmd/course-cli rest
```

Use a different bind address if needed:

```sh
go run ./cmd/course-cli rest --addr 127.0.0.1:8080
```

Example authenticated read:

```sh
curl -H "Authorization: Bearer $COURSE_CLI_API_TOKEN" \
  "http://127.0.0.1:8788/v1/courses?status=published"
```

REST import endpoints use multipart uploads:

```text
POST /v1/import/plan
  multipart field: zip

POST /v1/import/apply?conflict_strategy=fail
  multipart fields: zip, resolved_plan (optional)
```

## Run The Learner Reader

The `web/` app is a Bun-managed Next.js App Router reader. It calls the Go REST
API from server-only code so `COURSE_CLI_API_TOKEN` never reaches the browser.
Published catalog, course, and lesson reads are data-backed; quiz, practice, and
course-test activity surfaces are still clearly labeled sample data until the
learner-safe projection handoff is wired into the components.

Start the Go REST API first, then in another terminal:

```sh
cd web
export COURSE_API_BASE_URL=http://127.0.0.1:8788
export COURSE_CLI_API_TOKEN=dev-token
export COURSE_REVALIDATION_SECRET=dev-secret
bun install
bun run dev
```

Open:

```text
http://localhost:3000
```

## Import Demo Content

The JavaScript Foundations package can make a local database reader-ready:

```sh
./scripts/load-javascript-foundations.sh
./scripts/load-javascript-foundations.sh --with-postgres
```

The script validates `courses/javascript-foundations.zip`, runs migrations,
builds `bin/course-cli`, applies the import, and lists published courses. See
[docs/runbooks/2026-06-06-test-ready-load-js-foundations.md](docs/runbooks/2026-06-06-test-ready-load-js-foundations.md)
for the step-by-step runbook.

The import workflow is:

```text
normalize -> plan -> resolve -> apply
```

Do not edit `payload` values inside a resolved plan. Change the zip and replan
when imported content or identity must change.

## Test

Focused adapter coverage:

```sh
go test ./internal/course/adapter/playground ./internal/course/adapter/cli ./internal/course/adapter/rest ./cmd/course-cli
```

Full Go suite:

```sh
go test ./...
```

Web checks:

```sh
cd web
bun run test
bun run lint
bun run typecheck
```
