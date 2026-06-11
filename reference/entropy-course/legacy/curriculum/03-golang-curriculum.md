# Go Curriculum: Beginner to Advanced

## Track goal

Build the **TaskForge core backend** in Go: a clean JSON API, database persistence, event processing worker, observability, admin tooling, and deployment workflow. The track emphasizes simple design, explicit errors, tests, concurrency, and production operations.

## What you will be able to do

- Build Go modules, packages, command-line programs, HTTP APIs, and workers.
- Model domains with structs, methods, interfaces, and explicit errors.
- Use goroutines, channels, context cancellation, synchronization, and race detection.
- Persist data, write tests, profile performance, add observability, and deploy services.
- Explain Go design tradeoffs clearly.

## Primary references used

- [Go Documentation](https://go.dev/doc/)
- [A Tour of Go](https://go.dev/tour/)
- [Go Generics Tutorial](https://go.dev/doc/tutorial/generics)
- [Go Concurrency Learning Resources](https://go.dev/wiki/LearnConcurrency)

## Project milestone map

| Stage | Product feature           | Concepts practiced                         | Evidence of completion                 |
| ----- | ------------------------- | ------------------------------------------ | -------------------------------------- |
| 1     | Health-check API skeleton | modules, packages, commands                | Service starts and is formatted        |
| 2     | Domain service            | structs, methods, interfaces, errors       | Task rules are tested                  |
| 3     | REST API                  | net/http, JSON, middleware                 | Clients can create and list tasks      |
| 4     | Persistent backend        | database, migrations, transactions         | Data survives restarts                 |
| 5     | Event worker              | goroutines, channels, context, race safety | Background jobs process task events    |
| 6     | Production service        | auth, observability, profiling, deployment | Containerized API with runbooks and CI |

## Curriculum modules

Each module is designed as a small product sprint. Do not move on until the deliverable works, has tests or manual verification notes, and is committed with a clear README update.

### Module 1: Setup, modules, packages, and Go workflow

**Level:** Beginner

**Concepts to learn**

- Go toolchain, modules, packages, and `go` commands
- `gofmt`, `go vet`, and simple project layout
- The `main` package versus library packages
- Reading compiler errors productively

**Project build**

- Create `services/go-api` with a `cmd/api` entry point and internal packages.
- Print configuration and start a minimal health-check server.
- Document how to run, test, and format the service.

**Practice tasks**

- Create and import one internal package.
- Run formatting and tests from a clean clone.
- Add a simple `make` or script file only if it improves repeatability.

**Completion checklist**

- [ ] The service starts with one command.
- [ ] Package names are short and meaningful.
- [ ] Formatting is automatic and consistent.

### Module 2: Basic syntax, values, functions, and control flow

**Level:** Beginner

**Concepts to learn**

- Variables, constants, zero values, and short declarations
- Functions, multiple return values, and named result caution
- Conditionals, loops, and `switch`
- Pointers at a practical beginner level

**Project build**

- Create functions for task priority scoring, due-date status, and slug generation.
- Use pointers only where mutation or optionality is intentional.
- Add command-line flags for local development configuration.

**Practice tasks**

- Convert JavaScript task fixture logic into Go functions.
- Write small examples that reveal zero values.
- Explain where pointers appear in your code and why.

**Completion checklist**

- [ ] Functions are simple and tested.
- [ ] Zero values are handled intentionally.
- [ ] You can explain basic pointer use without overusing pointers.

### Module 3: Structs, methods, interfaces, and domain modeling

**Level:** Beginner

**Concepts to learn**

- Struct fields, tags, methods, and composition
- Interfaces as behavior contracts
- Package-level visibility
- Small interfaces defined near consumers

**Project build**

- Model users, projects, tasks, comments, and events as Go structs.
- Create repository interfaces for task storage.
- Add JSON struct tags for API responses.

**Practice tasks**

- Use composition for shared metadata such as IDs and timestamps.
- Define a tiny interface for clock/time dependency in tests.
- Avoid premature interface creation; introduce interfaces when a consumer needs one.

**Completion checklist**

- [ ] Domain structs match API contracts clearly.
- [ ] Interfaces are small and purposeful.
- [ ] Public fields and methods are documented where useful.

### Module 4: Errors, validation, and explicit control flow

**Level:** Beginner

**Concepts to learn**

- Error values, wrapping, and sentinel errors
- Validation functions and domain invariants
- Early returns
- Logging errors with context

**Project build**

- Validate task creation, update, assignment, and completion.
- Return structured API errors from domain and handler layers.
- Wrap lower-level errors without losing their cause.

**Practice tasks**

- Create one validation error type and one not-found error.
- Write tests for invalid input and missing task IDs.
- Map domain errors to HTTP status codes.

**Completion checklist**

- [ ] Errors are explicit and not swallowed.
- [ ] Validation rules live close to the domain.
- [ ] Handlers return consistent error responses.

### Module 5: Collections, generics, and reusable helpers

**Level:** Beginner

**Concepts to learn**

- Slices, maps, and iteration
- Sorting and searching
- Generics for simple reusable functions
- When not to use generics

**Project build**

- Create pagination, filtering, and sorting helpers for tasks.
- Use maps for fast lookups by ID.
- Add one generic helper only where it removes real duplication.

**Practice tasks**

- Write non-generic and generic versions of a helper and compare readability.
- Handle nil and empty slices deliberately.
- Benchmark a simple lookup versus loop only after correctness.

**Completion checklist**

- [ ] Collections are initialized safely.
- [ ] Generic code is clear and constrained.
- [ ] Pagination behavior is tested.

### Module 6: Testing, table tests, fakes, and coverage that matters

**Level:** Intermediate

**Concepts to learn**

- The `testing` package
- Table-driven tests
- Fakes and dependency injection
- Coverage as a guide, not a goal

**Project build**

- Add tests for validation, repositories, handlers, and services.
- Create in-memory fake repositories for service tests.
- Add a repeatable test command to the README.

**Practice tasks**

- Write table tests for task validation rules.
- Create tests for both expected errors and happy paths.
- Refactor code that is painful to test.

**Completion checklist**

- [ ] Core domain behavior is covered by fast tests.
- [ ] Tests read like examples of the business rules.
- [ ] Fakes do not duplicate too much production logic.

### Module 7: HTTP server, routing, middleware, and JSON APIs

**Level:** Intermediate

**Concepts to learn**

- `net/http` fundamentals
- Routing and middleware
- JSON decoding and encoding
- HTTP status codes and idempotency

**Project build**

- Implement REST endpoints for projects and tasks.
- Add middleware for request IDs, logging, panic recovery, and content type.
- Return consistent JSON responses.

**Practice tasks**

- Implement handlers first with the standard library, then optionally compare a router package.
- Add integration tests using `httptest`.
- Document API examples with curl commands.

**Completion checklist**

- [ ] Endpoints follow consistent naming and status conventions.
- [ ] Bad JSON and validation errors are handled clearly.
- [ ] The API is usable from the JavaScript and TypeScript clients.

### Module 8: Persistence, migrations, and repository implementation

**Level:** Intermediate

**Concepts to learn**

- `database/sql` style APIs or a chosen database library
- Schema design and migrations
- Transactions
- Repository implementation and data mapping

**Project build**

- Persist users, projects, tasks, comments, and events in a database.
- Add migrations and seed data.
- Use transactions for multi-step task changes.

**Practice tasks**

- Model indexes for common queries such as project tasks by status.
- Write tests with a temporary database or clear repository boundary.
- Handle duplicate IDs and foreign key failures.

**Completion checklist**

- [ ] Schema supports product queries efficiently enough for the current scale.
- [ ] Migrations are repeatable.
- [ ] Repository code maps database rows to domain structs safely.

### Module 9: Contexts, cancellation, deadlines, and graceful shutdown

**Level:** Intermediate

**Concepts to learn**

- `context.Context` in request lifecycles
- Timeouts and cancellation
- Graceful server shutdown
- Avoiding leaked goroutines

**Project build**

- Pass request context through services and repositories.
- Add timeouts for external calls and database operations.
- Implement graceful shutdown for the API.

**Practice tasks**

- Cancel a client request and verify work stops where possible.
- Add shutdown handling for active requests.
- Write a test for timeout behavior.

**Completion checklist**

- [ ] Long-running operations receive context.
- [ ] The service shuts down cleanly.
- [ ] You can explain context values versus cancellation.

### Module 10: Goroutines, channels, select, and worker pipelines

**Level:** Intermediate

**Concepts to learn**

- Goroutines and channels
- `select`, buffered versus unbuffered channels
- Worker pools and fan-out/fan-in
- Backpressure and cancellation

**Project build**

- Create a background worker that processes TaskForge events.
- Send notifications, recompute project metrics, or index tasks asynchronously.
- Use contexts to stop workers cleanly.

**Practice tasks**

- Build a worker pool with bounded concurrency.
- Simulate slow jobs and failures.
- Add metrics for queue depth and processing duration.

**Completion checklist**

- [ ] Workers do not leak after shutdown.
- [ ] Channel ownership is clear.
- [ ] Backpressure behavior is documented.

### Module 11: Synchronization, race detection, and shared state

**Level:** Intermediate

**Concepts to learn**

- Mutexes, atomic operations at a high level, and safe shared state
- Race detector
- Immutable messages versus shared mutation
- Cache invalidation basics

**Project build**

- Add an in-memory cache for project summaries.
- Protect shared state correctly or redesign around message passing.
- Run tests with race detection.

**Practice tasks**

- Create a deliberate race in a toy example and observe the detector.
- Compare mutex-protected cache with channel-owned cache.
- Document cache invalidation rules.

**Completion checklist**

- [ ] Race detector passes.
- [ ] Cache invalidation is tied to domain events.
- [ ] You can explain when a mutex is simpler than a channel.

### Module 12: CLI tools, configuration, logging, and operational ergonomics

**Level:** Advanced

**Concepts to learn**

- Command-line flags and subcommands
- Configuration from environment and files
- Structured logging
- Admin and maintenance commands

**Project build**

- Create a Go admin CLI for migrations, seed data, user lookup, and event replay.
- Add structured logs with request IDs and job IDs.
- Validate configuration at startup.

**Practice tasks**

- Add a dry-run mode for dangerous commands.
- Create a sample `.env.example`.
- Document operational commands.

**Completion checklist**

- [ ] Operators can diagnose common problems.
- [ ] Dangerous commands have confirmations or dry-run mode.
- [ ] Logs are useful and not noisy.

### Module 13: Service architecture, boundaries, and maintainable Go packages

**Level:** Advanced

**Concepts to learn**

- Layering: handler, service, repository, worker
- Dependency direction
- Internal packages
- Avoiding framework-shaped architecture

**Project build**

- Refactor the Go API into clear packages.
- Keep domain rules independent from HTTP and database code.
- Add architecture decision records for package boundaries.

**Practice tasks**

- Draw package dependencies.
- Move one rule from a handler into the service or domain layer.
- Delete unused abstractions.

**Completion checklist**

- [ ] Handlers are thin.
- [ ] Domain rules are testable without a server.
- [ ] Package boundaries help future features.

### Module 14: API design, auth, authorization, and rate limiting

**Level:** Advanced

**Concepts to learn**

- Authentication versus authorization
- Ownership checks
- Pagination and filtering API design
- Rate limiting and abuse prevention

**Project build**

- Add user identity to requests.
- Enforce project membership before task modifications.
- Add rate limiting for write endpoints.

**Practice tasks**

- Write tests for cross-user access attempts.
- Document auth assumptions clearly.
- Add pagination metadata to list endpoints.

**Completion checklist**

- [ ] Authorization lives server-side.
- [ ] List endpoints are safe for large data sets.
- [ ] The API is hard to misuse.

### Module 15: Observability: metrics, tracing, health checks, and runbooks

**Level:** Advanced

**Concepts to learn**

- Structured logs, metrics, traces, and health checks
- Golden signals: latency, traffic, errors, saturation
- Runbooks and incident notes
- Alert design basics

**Project build**

- Add health and readiness endpoints.
- Record request duration, error count, worker job count, and queue depth.
- Create a runbook for high error rate and worker backlog.

**Practice tasks**

- Simulate a database outage and inspect logs.
- Create a simple dashboard or metrics text output.
- Write an incident note after a staged failure.

**Completion checklist**

- [ ] The service is observable during failure.
- [ ] Runbooks are actionable.
- [ ] Metrics answer product and operations questions.

### Module 16: Performance, benchmarks, profiling, and load testing

**Level:** Advanced

**Concepts to learn**

- Benchmarks with the testing package
- CPU and memory profiling
- Allocation awareness
- Load test basics and bottleneck analysis

**Project build**

- Benchmark task filtering and project summary generation.
- Profile one slow endpoint.
- Run a small load test and record results.

**Practice tasks**

- Optimize only after creating a baseline.
- Compare two implementations with benchmarks.
- Write a performance report with tradeoffs.

**Completion checklist**

- [ ] Optimizations are evidence-based.
- [ ] Benchmarks are stable enough to compare.
- [ ] You can explain the bottleneck before and after.

### Module 17: Deployment, containers, CI, and release discipline

**Level:** Advanced

**Concepts to learn**

- Static binaries and container images
- CI checks: format, test, vet, race, build
- Database migrations in deployment
- Rollback and compatibility

**Project build**

- Create a containerized Go API and worker.
- Add CI steps for formatting, tests, vetting, race checks, and builds.
- Document deployment and rollback steps.

**Practice tasks**

- Build from a clean environment.
- Run migrations against a fresh local database.
- Tag a release and write release notes.

**Completion checklist**

- [ ] A clean clone can build and test everything.
- [ ] Deployment docs include migrations and rollback.
- [ ] CI catches common mistakes before merge.

### Module 18: Go capstone: TaskForge core API and event worker

**Level:** Capstone

**Concepts to learn**

- End-to-end backend delivery
- HTTP, persistence, concurrency, testing, observability, deployment
- Operational ownership
- Portfolio documentation

**Project build**

- Complete a Go TaskForge API with database persistence, auth checks, project/task endpoints, event log, worker, metrics, and containerized deployment.
- Connect the TypeScript SDK and JavaScript app to the Go API.
- Write runbooks and architecture decision records.

**Practice tasks**

- Run a realistic demo from empty database to populated dashboard.
- Simulate failures: bad input, canceled request, worker backlog, database outage.
- Review the system using the assessment rubric.

**Completion checklist**

- [ ] The API supports the full frontend workflow.
- [ ] Workers process events reliably and shut down cleanly.
- [ ] The service is testable, observable, and deployable.

## Go capstone specification

Build `services/go-api` and `services/go-worker`.

Minimum features:

- Health, readiness, project, task, comment, and event endpoints.
- Domain validation and consistent error responses.
- Database migrations and repository implementation.
- Authentication stub and project membership authorization.
- Background event worker with graceful shutdown.
- Tests for domain rules, handlers, repositories, and worker behavior.
- Structured logs, health checks, metrics, and runbooks.
- Containerized build and CI checklist.

Stretch features:

- Outbox pattern for reliable event processing.
- OpenAPI contract generation.
- Rate limiting and audit logs.
- Load test report and performance tuning notes.
- Multi-tenant project isolation.

## Final review questions

- Which package owns each business rule?
- What happens when a request is canceled midway through database work?
- How do workers stop without losing or duplicating jobs?
- Which endpoints are safe for large projects and which need pagination or indexing?
- What metrics would you check first during a production incident?
