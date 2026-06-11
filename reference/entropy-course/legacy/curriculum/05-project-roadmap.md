# TaskForge Project Roadmap and Capstone Ideation

TaskForge is a realistic project management platform used to teach JavaScript, TypeScript, Go, and Rust through one connected product. You can build it as one monorepo or four separate repositories.

## Product vision

TaskForge helps small teams track projects, tasks, comments, assignments, due dates, imports, exports, and sync workflows. The learning version focuses on engineering depth rather than visual polish.

## Core domain

### Entities

```text
User
  id, name, email, role, createdAt

Project
  id, ownerId, name, description, status, createdAt, updatedAt

Task
  id, projectId, title, description, status, priority, assigneeId, dueDate,
  tags, createdAt, updatedAt, completedAt

Comment
  id, taskId, authorId, body, createdAt

TaskEvent
  id, taskId, actorId, type, payload, createdAt

SyncRecord
  id, source, localVersion, remoteVersion, status, conflictReason, createdAt
```

### Task statuses

```text
draft -> active -> blocked -> active -> completed -> archived
```

Allow status transitions intentionally. For example, an archived task cannot be completed without being restored first.

## API sketch

```text
GET    /health
GET    /ready
GET    /v1/projects
POST   /v1/projects
GET    /v1/projects/{projectId}
PATCH  /v1/projects/{projectId}
GET    /v1/projects/{projectId}/tasks
POST   /v1/projects/{projectId}/tasks
GET    /v1/tasks/{taskId}
PATCH  /v1/tasks/{taskId}
POST   /v1/tasks/{taskId}/complete
POST   /v1/tasks/{taskId}/comments
GET    /v1/events?projectId=...
POST   /v1/imports/validate
POST   /v1/sync/preview
POST   /v1/sync/apply
```

## Language ownership in the final system

| Language   | Product responsibility                                                | Why it fits                                                                                 |
| ---------- | --------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| JavaScript | Browser app and early Node.js API                                     | Best place to learn browser behavior, DOM, async workflows, and runtime fundamentals.       |
| TypeScript | Shared contracts, SDK, admin dashboard, optional BFF                  | Adds type safety and domain modeling on top of JavaScript.                                  |
| Go         | Core API, database-backed service, event worker                       | Strong fit for simple services, explicit errors, concurrency, and deployment.               |
| Rust       | CLI, sync/search/import validation, optional high-performance service | Strong fit for safe file handling, performance-sensitive tools, and robust local workflows. |

## End-to-end architecture

```text
Browser JS App ----> Go API ----> Database
      |                 |
      |                 +----> Go Event Worker
      |
TypeScript Admin Dashboard ----> TypeScript SDK ----> Go API
      |
      +----> Shared Contracts

Rust CLI ----> Local Cache
    |             |
    +---- sync ----> Go API
    |
    +---- optional ----> Rust Sync/Search Service
```

## Progressive milestone plan

### Milestone 1: Local task model

- Build in JavaScript first.
- Store users, projects, and tasks as plain data.
- Implement create, update, complete, filter, and summarize.
- Add sample fixtures and a README.

### Milestone 2: Browser task board

- Add DOM rendering, forms, events, and local storage.
- Add import/export JSON.
- Add validation and error messages.
- Add accessibility checks.

### Milestone 3: Remote API prototype

- Add a simple Node.js API or mock server.
- Replace local-only data access with repository adapters.
- Add async states, retries, and error handling.

### Milestone 4: Type-safe contracts and SDK

- Introduce TypeScript shared domain types.
- Create a typed SDK for the API.
- Validate external JSON at runtime.
- Replace fragile strings and booleans with unions and state machines.

### Milestone 5: Go core backend

- Build a Go API with persistent storage.
- Implement task, project, comment, and event endpoints.
- Add auth checks, pagination, validation, and tests.
- Add a worker that processes events.

### Milestone 6: Rust CLI and sync tooling

- Import and validate exported TaskForge data.
- Summarize, search, and filter local tasks.
- Sync with the Go API and handle conflicts.
- Package the CLI and measure performance on large data.

### Milestone 7: Integration and production polish

- Connect JavaScript and TypeScript clients to the Go API.
- Use Rust CLI for import/export/sync workflows.
- Add CI, release notes, runbooks, screenshots, and architecture diagrams.
- Create a final portfolio demo script.

## Feature backlog by difficulty

### Beginner features

- Create task.
- Edit task title and description.
- Complete task.
- Delete or archive task.
- Filter by status.
- Save data locally.
- Import and export JSON.

### Intermediate features

- Search and sort tasks.
- Project dashboard metrics.
- Form validation with field-specific errors.
- API client and remote sync.
- Comments and activity history.
- Typed SDK and shared contracts.
- Database persistence.

### Advanced features

- Role-based authorization.
- Event worker and audit log.
- Offline queue and conflict resolution.
- Bulk import with progress and cancellation.
- Search indexing.
- Observability and runbooks.
- Load testing and performance reports.
- Versioned API contracts.

## Alternate project ideas

Use these if TaskForge does not excite you. Keep the same learning structure: start local, add UI, add types, add API, add concurrency, add Rust tooling.

### 1. Inventory and orders platform

- JavaScript: local inventory dashboard.
- TypeScript: typed admin panel and SDK.
- Go: order and stock API with background reconciliation worker.
- Rust: CSV importer and stock audit CLI.

### 2. Personal finance tracker

- JavaScript: transaction categorizer UI.
- TypeScript: typed budgeting dashboard.
- Go: accounts, transactions, and reports API.
- Rust: statement parser and reconciliation CLI.

### 3. Job application tracker

- JavaScript: local board for applications.
- TypeScript: typed dashboard and email template manager.
- Go: API with reminders and event worker.
- Rust: resume keyword scanner and import tool.

### 4. Learning management system

- JavaScript: course progress tracker.
- TypeScript: typed instructor/admin dashboard.
- Go: courses, lessons, enrollments, progress API.
- Rust: content import validator and search indexer.

### 5. Incident management tool

- JavaScript: incident timeline UI.
- TypeScript: typed event SDK and dashboard.
- Go: incident API, notifications, and escalation worker.
- Rust: log parser and report generator.

## Suggested final demo script

1. Start the Go API and worker.
2. Open the TypeScript admin dashboard.
3. Create a project and several tasks.
4. Use the JavaScript browser app to edit and complete tasks.
5. Export data.
6. Run the Rust CLI to validate, summarize, and search the export.
7. Use the Rust CLI to sync a local change back to the Go API.
8. Show logs, metrics, tests, and release notes.
9. Explain one tradeoff in each language.

## Architecture decision records to write

Create `docs/adr/` and write short notes for these decisions:

- Why use a shared TaskForge domain across all language tracks?
- Why keep domain logic separate from UI, HTTP, and database code?
- Why use TypeScript for SDK and contracts?
- Why use Go for the core API and worker?
- Why use Rust for CLI, sync, import validation, or search?
- How are errors represented across API, SDK, UI, and CLI?
- How are API contract changes versioned?
- How does sync conflict resolution work?

## Data set ideas for realistic testing

Create generated fixtures for:

- 5 users, 3 projects, 30 tasks for quick manual testing.
- 50 users, 25 projects, 5,000 tasks for performance and pagination.
- 10 malformed imports for validation tests.
- 5 conflict scenarios for sync testing.
- 3 security edge cases with HTML-like task titles and unexpected fields.

## Integration checklist

- JavaScript app can use local storage or API storage through an adapter.
- TypeScript SDK points to the Go API and validates responses.
- Go API follows the documented contract.
- Rust CLI can import JavaScript exports and sync with Go API.
- Errors have consistent codes across API, SDK, UI, and CLI.
- Release notes explain cross-language changes.
