# Plan: Instructor write endpoints for the REST adapter

Status: not started. Scope: add course + lesson **write** endpoints to
`internal/course/adapter/rest`, reusing the existing `CourseService` /
`LessonService`. Block, quiz, test, and practice writes already exist over REST
and are out of scope.

Pattern to follow: every handler is thin — decode JSON → call the injected
service → `writeJSON` / `WriteHeader`. Mirror the existing handlers in
`server.go` (e.g. `handleQuiz`, `handleTests`) and the read handlers in
`course_handlers.go`. No new logic belongs in this layer.

## 1. Endpoints to add

Courses:

| Method + path | Service call | Success |
|---|---|---|
| `POST /v1/courses` | `CreateCourse` | 201, `{ "ID": ... }` |
| `PATCH /v1/courses/{id}` | `UpdateCourse` | 200, `{ "ID": ... }` |
| `DELETE /v1/courses/{id}` | `DeleteCourse` | 204 |
| `POST /v1/courses/{id}/publish` | `PublishCourse` | 204 |
| `POST /v1/courses/{id}/unpublish` | `UnpublishCourse` | 204 |

Lessons:

| Method + path | Service call | Success |
|---|---|---|
| `POST /v1/courses/{id}/lessons` | `CreateLesson` (CourseID from path) | 201, `{ "ID": ... }` |
| `PATCH /v1/lessons/{id}` | `UpdateLesson` | 200, `{ "ID": ... }` |
| `DELETE /v1/lessons/{id}` | `DeleteLesson` | 204 |
| `POST /v1/courses/{id}/lessons/reorder` | `ReorderLessons` (CourseID from path) | 204 |

These extend handlers that currently only do GET, plus three new path shapes
(`/publish`, `/unpublish`, `/lessons/reorder`).

## 2. Request bodies (core DTO fields, decoded case-insensitively)

- `CreateCourseInput`: `Title`, `Slug`, `Description`. **Do NOT read
  `InstructorID` from the body** — set it from `server.instructorID`, exactly
  like `import_handlers.go`. Guard: if `server.instructorID == ""`, return a
  validation error (see import handler precedent at lines ~132).
- `UpdateCourseInput`: `Title *string`, `Description *string`, `Slug *string`;
  set `ID` from the path.
- `CreateLessonInput`: `Title`, `Order *int`; set `CourseID` from the path.
- `UpdateLessonInput`: `Title *string`; set `ID` from the path.
- `ReorderLessonsInput`: `Order []LessonPosition` (`LessonID`, `Position`); set
  `CourseID` from the path.

Pointer fields (`*string`, `*int`) are how the services express "field omitted
vs set" — keep them as pointers so PATCH stays partial.

## 3. Handlers — extend, don't duplicate

In `course_handlers.go`, widen the existing methods from GET-only to a method
switch (copy the shape of `handleQuiz` / `handleTest`):

- `handleCourses`: add `case http.MethodPost` → `CreateCourse` (inject
  instructor id) → 201. Keep GET. Update `methodNotAllowed(... GET, POST)`.
- `handleCourse`: add `PATCH` → `UpdateCourse`, `DELETE` → `DeleteCourse`. Keep
  GET. Allow `GET, PATCH, DELETE`.
- `handleCourseLessons`: add `POST` → `CreateLesson`. Keep GET. Allow
  `GET, POST`.
- `handleLesson`: add `PATCH` → `UpdateLesson`, `DELETE` → `DeleteLesson`. Keep
  GET. Allow `GET, PATCH, DELETE`.

New handlers (POST-only, follow `handleReorderQuestions` for the reorder one):

- `handleCoursePublish(w, r, courseID)` → `PublishCourse` → 204.
- `handleCourseUnpublish(w, r, courseID)` → `UnpublishCourse` → 204.
- `handleCourseLessonsReorder(w, r, courseID)` → `ReorderLessons` → 204.

## 4. Routing — `Server.route()` in `server.go`

The switch matches on `len(segments)` + literal segments, so cases are mutually
exclusive. Existing read cases stay; the method switch inside each handler picks
up the new verbs for paths that already route (`POST /v1/courses`,
`PATCH/DELETE /v1/courses/{id}`, `POST /v1/courses/{id}/lessons`,
`PATCH/DELETE /v1/lessons/{id}`).

Add three new cases for the new path shapes:

```go
case len(segments) == 4 && segments[1] == "courses" && segments[3] == "publish":
    server.handleCoursePublish(response, request, segments[2])
case len(segments) == 4 && segments[1] == "courses" && segments[3] == "unpublish":
    server.handleCourseUnpublish(response, request, segments[2])
case len(segments) == 5 && segments[1] == "courses" && segments[3] == "lessons" && segments[4] == "reorder":
    server.handleCourseLessonsReorder(response, request, segments[2])
```

Collision check: `len==4 courses/{id}/{publish|unpublish}` is distinct from the
existing `len==4 courses/{id}/{quizzes|tests|practices|lessons}`. The `len==5
courses/{id}/lessons/reorder` is distinct from everything else. No clashes.

## 5. Error mapping — already handled

`writeError` + `statusForError` already map `domain.ErrValidation` → 400,
`domain.ErrNotFound` → 404, and the in-use conflicts → 409. New handlers just
call `writeError(response, err)`; no new mapping needed unless a new domain
sentinel is introduced.

## 6. Tests (`course_handlers_test.go`, extend existing fakes)

The read tests already define `courseReadFake` (embeds `courseServiceFake`) and
`lessonReadFake` (embeds `lessonServiceFake`). Extend those fakes to record
write calls (createIn/updateIn/deleteIn/publishIn/unpublishIn/reorderIn and
canned outputs), or add sibling write-fakes — keep field names distinct from the
embedded fakes to avoid shadowing (see existing `lessonReadFake` for the
naming approach).

Cases to cover, mirroring `server_test.go` style:

- Create course: body parsed; **`InstructorID` comes from server config, not the
  body**; 201 + id echoed.
- Create course with empty server instructor id → validation error / 400.
- Update course: PATCH pointers set only for provided fields; 200.
- Delete course → 204.
- Publish / unpublish → 204, correct id forwarded.
- Create lesson: `CourseID` taken from path, not body; 201.
- Update lesson → 200; Delete lesson → 204.
- Reorder lessons: `CourseID` from path, `Order` placements forwarded; 204.
- Method rejection: e.g. `PUT /v1/courses/{id}` → 405 with `Allow` header.
- Auth: each write path returns 401 without the bearer token (covered globally
  by the `authenticate` middleware, so one representative case is enough).

## 7. Verify

```
gofmt -l internal/course/adapter/rest
go build ./...
go test ./internal/course/adapter/rest/...
```

(Use tabs for indentation; the repo is gofmt-clean.)

## 8. Follow-ups (not in this slice)

- **Wire casing**: responses are PascalCase (`ID`, `Title`) because core DTOs
  have no `json` tags. If the frontend wants camelCase, add `json:"..."` tags to
  the core view structs — but that's a breaking change to every endpoint, so
  decide before the frontend hardens against the current shape.
- **AuthZ granularity**: all writes currently sit behind a single shared bearer
  token (instructor). If students ever hit the same server, split read vs write
  auth before exposing these.
- **Lesson block writes** already exist (`POST /v1/lessons/{id}/blocks`, etc.) —
  no work needed there.
