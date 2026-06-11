#!/usr/bin/env bash
#
# Make the repo test-ready and load the "JavaScript Foundations" course.
#
#   ./scripts/load-javascript-foundations.sh              # migrate + build + import
#   ./scripts/load-javascript-foundations.sh --with-postgres   # also boot a local docker Postgres first
#
# Prereqs: Go 1.22+, PostgreSQL 15+ reachable via $COURSE_CLI_DB_URL
# (use --with-postgres to start the docker instance from the README).
#
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO"
ZIP="courses/javascript-foundations.zip"

# --- configuration (override by exporting before running) -------------------
export COURSE_CLI_DB_URL="${COURSE_CLI_DB_URL:-postgres://postgres:password@127.0.0.1:5432/entropy_course?sslmode=disable}"
export COURSE_CLI_INSTRUCTOR_ID="${COURSE_CLI_INSTRUCTOR_ID:-550e8400-e29b-41d4-a716-446655440010}"
export COURSE_CLI_API_TOKEN="${COURSE_CLI_API_TOKEN:-dev-token}"

say(){ printf '\n\033[1;36m== %s\033[0m\n' "$*"; }

# --- 0. optional: boot a local Postgres (README settings) -------------------
if [[ "${1:-}" == "--with-postgres" ]]; then
  say "Starting local Postgres (docker)"
  docker start entropy-course-postgres 2>/dev/null || docker run --name entropy-course-postgres \
    -e POSTGRES_DB=entropy_course -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password \
    -p 5432:5432 -d postgres:17-alpine
  echo "waiting for Postgres..."; sleep 4
fi

# --- 1. pre-flight: validate the package before touching the DB -------------
say "Validating the course package (static, no DB)"
python3 scripts/validate_course_zip.py "$ZIP"

# --- 2. run database migrations ---------------------------------------------
say "Applying database migrations (course-cli migrate up)"
go run ./cmd/course-cli migrate up
go run ./cmd/course-cli migrate status || true

# --- 3. build the CLI binary ------------------------------------------------
say "Building the course-cli binary"
mkdir -p bin
go build -o bin/course-cli ./cmd/course-cli
echo "built ./bin/course-cli"

# --- 4. import + apply (fresh course, fail on any conflict) -----------------
say "Importing the course (import apply, conflict-strategy=fail)"
# Re-run safe: if it already exists this will report a conflict; use --conflict-strategy update to overwrite.
./bin/course-cli import apply "$ZIP" --conflict-strategy fail --force -o table

# --- 5. confirm it published ------------------------------------------------
say "Published courses now in the database"
./bin/course-cli course list --status published -o table

cat <<EOF

Done. The course is loaded and published (course.yaml status: published).

Next — see it in the learner reader and prove answer-safety:

  # terminal A — REST API (binds 127.0.0.1:8788)
  ./bin/course-cli rest

  # terminal B — the Next.js reader
  cd web
  export COURSE_API_BASE_URL=http://127.0.0.1:8788
  export COURSE_CLI_API_TOKEN=$COURSE_CLI_API_TOKEN
  export COURSE_REVALIDATION_SECRET=dev-secret
  bun install && bun run dev      # open http://localhost:3000

  # prove a learner read strips answers (no correct_indices / solutions):
  COURSE_ID=\$(./bin/course-cli course list --status published -o json | python3 -c "import sys,json;print(json.load(sys.stdin)[0]['id'])")
  # then fetch a quiz id from that course and compare:
  #   curl -H "Authorization: Bearer $COURSE_CLI_API_TOKEN" ".../v1/quizzes/<id>?view=learner"   # safe
  #   curl -H "Authorization: Bearer $COURSE_CLI_API_TOKEN" ".../v1/quizzes/<id>"                # full
EOF
