# PRD: Course CLI Module

**Status:** Draft  
**Author:** Stephen Antoni  
**Date:** 2026-05-24  
**Version:** 0.1.0

---

## 1. Overview

The **Course CLI** is a greenfield, single-binary command-line tool written in Go that serves as the backend management interface for a course and lesson platform. It allows **Instructors** to manage their own content and **Platform Admins** to manage all content and users platform-wide. The CLI connects directly to a PostgreSQL database and is the authoritative backend for course and lesson data operations.

---

## 2. Goals

- Provide a fast, scriptable CLI for managing courses and lessons
- Enforce a clear ownership-based permission model between Instructors and Admins
- Be automation-friendly (JSON output, predictable exit codes, env var support)
- Ship as a single self-contained Go binary
- Lay the foundation for a future web frontend or API layer backed by the same database

## 3. Non-Goals

- No web UI or REST API in this phase
- No student-facing features (enrollment, progress tracking) in this phase
- Authentication and authorization are deferred to a future iteration (see §10)
- No multi-tenancy / organization support in this phase

---

## 4. User Roles

The CLI recognizes two actor roles, enforced by auth context (to be implemented in a future iteration):

| Capability | Instructor | Platform Admin |
|---|---|---|
| Create / edit / delete **own** courses | ✅ | ✅ |
| Create / edit / delete **any** course | ❌ | ✅ |
| Add / edit / delete lessons in **own** course | ✅ | ✅ |
| Publish / unpublish **own** course | ✅ | ✅ |
| Publish / unpublish **any** course | ❌ | ✅ |
| View enrollment / student data for **own** courses | ✅ | ✅ |
| View enrollment / student data for **all** courses | ❌ | ✅ |
| Manage instructors (create, assign, remove) | ❌ | ✅ |
| Run platform-wide admin commands | ❌ | ✅ |

---

## 5. Data Model

### 5.1 Course

```sql
CREATE TABLE courses (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title          TEXT NOT NULL,
  slug           TEXT NOT NULL UNIQUE,
  description    TEXT,
  instructor_id  UUID NOT NULL REFERENCES users(id),
  status         TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Fields:**
- `id` — UUID primary key, auto-generated
- `title` — human-readable course name (required)
- `slug` — URL-friendly unique identifier (required, unique)
- `description` — full course description (optional, markdown)
- `instructor_id` — owning instructor (FK to users)
- `status` — lifecycle state: `draft` | `published` | `archived`
- `created_at`, `updated_at` — audit timestamps

### 5.2 Lesson

```sql
CREATE TABLE lessons (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  course_id   UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  title       TEXT NOT NULL,
  content     TEXT,
  order       INTEGER NOT NULL DEFAULT 0,
  status      TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published')),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Fields:**
- `id` — UUID primary key, auto-generated
- `course_id` — parent course (FK, cascades on delete)
- `title` — lesson name (required)
- `content` — lesson body (optional, markdown/plain text)
- `order` — integer position within the course (for sequencing)
- `status` — `draft` | `published`
- `created_at`, `updated_at` — audit timestamps

### 5.3 User (stub — for auth iteration)

```sql
CREATE TABLE users (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email      TEXT NOT NULL UNIQUE,
  role       TEXT NOT NULL CHECK (role IN ('instructor', 'admin')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

---

## 6. Command Reference

Binary name: `course-cli`  
Command pattern: `course-cli <noun> <verb> [args] [flags]`

### 6.1 Course Commands

```
course-cli course create   --title <title> --slug <slug> [--description <text>]
course-cli course list     [--status draft|published|archived] [--output json]
course-cli course get      <course-id>    [--output json]
course-cli course update   <course-id>    [--title <title>] [--description <text>] [--slug <slug>]
course-cli course delete   <course-id>    [--force]
course-cli course publish  <course-id>
course-cli course unpublish <course-id>
```

| Command | Description | Instructor | Admin |
|---|---|---|---|
| `course create` | Create a new course (status: draft) | ✅ own | ✅ any |
| `course list` | List courses | Own only | All |
| `course get` | Get course detail | Own only | Any |
| `course update` | Update course metadata | Own only | Any |
| `course delete` | Delete a course (and all lessons) | Own only | Any |
| `course publish` | Set status to `published` | Own only | Any |
| `course unpublish` | Set status back to `draft` | Own only | Any |

### 6.2 Lesson Commands

```
course-cli lesson create  --course-id <id> --title <title> [--content <text>] [--order <int>]
course-cli lesson list    --course-id <id>  [--output json]
course-cli lesson get     <lesson-id>       [--output json]
course-cli lesson update  <lesson-id>       [--title <title>] [--content <text>]
course-cli lesson delete  <lesson-id>       [--force]
course-cli lesson reorder --course-id <id>  --order <lesson-id:pos,lesson-id:pos,...>
```

| Command | Description | Instructor | Admin |
|---|---|---|---|
| `lesson create` | Add a lesson to a course | Own course | Any course |
| `lesson list` | List all lessons in a course | Own course | Any course |
| `lesson get` | Get lesson detail | Own course | Any course |
| `lesson update` | Edit lesson content/title | Own course | Any course |
| `lesson delete` | Remove a lesson | Own course | Any course |
| `lesson reorder` | Resequence lessons within a course | Own course | Any course |

### 6.3 Admin Commands

_Admin-only. Returns permission denied for instructor role._

```
course-cli admin instructor list
course-cli admin instructor create  --email <email>
course-cli admin instructor delete  <instructor-id>  [--force]
course-cli admin course list        [--instructor-id <id>] [--status <status>] [--output json]
```

| Command | Description |
|---|---|
| `admin instructor list` | List all instructors |
| `admin instructor create` | Create a new instructor account |
| `admin instructor delete` | Remove an instructor (and optionally reassign their courses) |
| `admin course list` | List all courses across all instructors with optional filters |

### 6.4 Migration Commands

```
course-cli migrate up      ← apply all pending migrations
course-cli migrate down    ← rollback the last applied migration
course-cli migrate status  ← show list of applied and pending migrations
```

Uses `golang-migrate` under the hood. Migration files live in `./migrations/` and are embedded in the binary at build time via Go's `embed` package.

---

## 7. Global Flags

These flags are available on every command:

| Flag | Short | Description |
|---|---|---|
| `--output` | `-o` | Output format: `table` (default) or `json` |
| `--quiet` | `-q` | Print only the resource ID (useful for piping) |
| `--force` | `-f` | Skip confirmation prompts on destructive commands |
| `--verbose` | `-v` | Print debug info, stack traces, and SQL statements |
| `--db-url` | | PostgreSQL connection string (overrides config file) |

---

## 8. Output Formats

### Default (table)
```
ID        TITLE              STATUS     INSTRUCTOR
────────────────────────────────────────────────────
a1b2c3    Intro to Go        published  jane@...
d4e5f6    Advanced Postgres  draft      john@...
```

### JSON (`--output json`)
```json
[
  {"id": "a1b2c3", "title": "Intro to Go", "status": "published", "instructor_id": "..."},
  {"id": "d4e5f6", "title": "Advanced Postgres", "status": "draft", "instructor_id": "..."}
]
```

### Quiet (`--quiet` / `-q`)
```
a1b2c3
```
Only the resource ID — intended for shell scripting and piping.

**Rule:** All data output goes to **stdout**. All errors and warnings go to **stderr**.

---

## 9. Error Handling & Exit Codes

| Scenario | Exit Code | Example stderr |
|---|---|---|
| Success | `0` | — |
| Validation error (bad/missing input) | `1` | `Error: --title is required` |
| Not found (resource doesn't exist) | `2` | `Error: course a1b2c3 not found` |
| Permission denied | `3` | `Error: permission denied` |
| Internal / database error | `5` | `Error: internal error (run with --verbose for details)` |

**Destructive command confirmation:**
```
$ course-cli course delete a1b2c3
Delete course "Intro to Go" and all its lessons? [y/N]:
```
Bypassed with `--force` / `-f`.

---

## 10. Authentication (Deferred)

> ⚠️ **TODO: Auth is deferred to a future iteration.**

Planned approach:
- `course-cli auth login` / `course-cli auth logout` flow
- Token stored in `~/.config/course-cli/config.yaml`
- Environment variable override: `COURSE_CLI_TOKEN=xxx course-cli ...`
- Token encodes user role server-side; no role logic on the client

For this iteration, role enforcement is scaffolded in code but not enforced.

---

## 11. Configuration

Config file location: `~/.config/course-cli/config.yaml`

```yaml
db_url: postgres://user:password@localhost:5432/coursedb
# auth_token: <deferred>
```

Priority order (highest to lowest):
1. CLI flags (e.g., `--db-url`)
2. Environment variables (e.g., `COURSE_CLI_DB_URL`)
3. Config file
4. Defaults

---

## 12. Tech Stack

| Concern | Choice |
|---|---|
| Language | Go |
| CLI framework | `cobra` + `viper` |
| Database | PostgreSQL |
| DB driver | `pgx` |
| Migrations | `golang-migrate` (embedded via `embed`) |
| Table output | `tablewriter` |
| UUID generation | `google/uuid` |

---

## 13. Project Structure (Proposed)

```
course-cli/
├── cmd/
│   ├── root.go           ← global flags, config loading
│   ├── course.go         ← course subcommands
│   ├── lesson.go         ← lesson subcommands
│   ├── admin.go          ← admin subcommands
│   └── migrate.go        ← migration subcommands
├── internal/
│   ├── db/               ← postgres connection, query helpers
│   ├── domain/           ← Course, Lesson, User structs
│   ├── repository/       ← DB queries per entity
│   ├── service/          ← business logic, permission checks
│   └── output/           ← table/json/quiet formatters
├── migrations/           ← SQL migration files (embedded)
├── main.go
└── go.mod
```

---

## 14. Success Metrics

- All CRUD operations for courses and lessons work end-to-end against a live Postgres instance
- `--output json` is parseable by `jq` for all list and get commands
- Exit codes are correct for scripted error handling
- `migrate up` runs cleanly on a fresh database
- Admin commands correctly scope access (scaffolded, enforced post-auth)

---

## 15. Out of Scope / Future Iterations

- Authentication & authorization (role enforcement)
- Student/enrollment management
- Video/media asset management (`video_url`, `duration` fields)
- Course tags, categories, prerequisites
- REST or gRPC API layer
- Web frontend
- Multi-tenancy / organization support
