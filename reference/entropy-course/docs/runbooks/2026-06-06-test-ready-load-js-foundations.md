# Runbook: make the repo test-ready & load JavaScript Foundations

**Date:** 2026-06-06 · **Why:** Phase C1 — load real course material so the learner reader can be honestly evaluated (and the `?view=learner` projections exercised with real quizzes, practices, and a test).

This is the payoff of the 2026-06-05 checkpoint: the chosen next feature was *"Load real course material,"* with the note *"help me create a short but meaningful JavaScript course."* That course now exists at `courses/javascript-foundations/` (source) and `courses/javascript-foundations.zip` (import package).

## One command

```sh
./scripts/load-javascript-foundations.sh                 # Postgres already running
./scripts/load-javascript-foundations.sh --with-postgres # also boot the docker Postgres from the README
```

It validates the package, runs `course-cli migrate up`, builds `bin/course-cli`, and applies the import (the course is `status: published`, so it appears in the reader immediately). Override `COURSE_CLI_DB_URL`, `COURSE_CLI_INSTRUCTOR_ID`, `COURSE_CLI_API_TOKEN` by exporting them first; defaults match the README's local docker Postgres.

## Or step by step

```sh
# 1. validate before touching the DB (no Go/DB needed)
python3 scripts/validate_course_zip.py courses/javascript-foundations.zip

# 2. migrate
export COURSE_CLI_DB_URL="postgres://postgres:password@127.0.0.1:5432/entropy_course?sslmode=disable"
go run ./cmd/course-cli migrate up

# 3. build
go build -o bin/course-cli ./cmd/course-cli

# 4. import + publish (fresh course; --conflict-strategy update to overwrite a re-run)
export COURSE_CLI_INSTRUCTOR_ID="550e8400-e29b-41d4-a716-446655440010"
./bin/course-cli import apply courses/javascript-foundations.zip --conflict-strategy fail --force -o table
./bin/course-cli course list --status published -o table
```

## See it in the reader & prove answer-safety

```sh
# terminal A
./bin/course-cli rest                         # 127.0.0.1:8788

# terminal B
cd web
export COURSE_API_BASE_URL=http://127.0.0.1:8788
export COURSE_CLI_API_TOKEN=dev-token         # must match the CLI/REST token
export COURSE_REVALIDATION_SECRET=dev-secret
bun install && bun run dev                    # http://localhost:3000
```

Then compare a learner read to a full read — the learner one must omit `correct_indices`, `explanation` (pre-submit), practice `solution`/`test_cases`/`expected_stdout`, and the test `solution`:

```sh
curl -H "Authorization: Bearer dev-token" ".../v1/quizzes/<id>?view=learner"   # safe (stripped)
curl -H "Authorization: Bearer dev-token" ".../v1/quizzes/<id>"                # full fidelity
```

## What's in the course

Three lessons (text + video + quiz, plus practices), three quizzes, two JavaScript practices, and one course test — deliberately covering every block kind and every answer-bearing field the learner projections strip.

| Entity | Items |
| --- | --- |
| Lessons | Getting Started · Variables & Data Types · Functions & Control Flow |
| Quizzes | getting-started-quiz · variables-quiz · functions-quiz |
| Practices | count-to-five · fizzbuzz-lite (`language: javascript`) |
| Test | foundations-checkpoint (choice ×2 + coding ×1) |

**Verified here:** the package passes `scripts/validate_course_zip.py` (layout, schema, refs, enums, choice bounds), and every JavaScript solution was executed in Node and produces exactly its declared `expected_stdout` (count-to-five, fizzbuzz-lite, and the test's evens-2..10 coding item). The migrate/build/import steps run against your Postgres — they could not be executed in the authoring sandbox (no Go/Postgres, no network).

## Editing the course

Edit the readable source under `courses/javascript-foundations/`, then repackage:

```sh
python3 - <<'PY'
import shutil; shutil.make_archive("courses/javascript-foundations","zip","courses/javascript-foundations")
PY
python3 scripts/validate_course_zip.py courses/javascript-foundations.zip
```

(Or re-run `scripts/build_javascript_foundations.py` if you prefer regenerating from the data-driven builder.) `slug` is the course identity — re-applying with the same slug updates the existing course when you pass `--conflict-strategy update`.
