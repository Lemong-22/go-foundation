# Rust Curriculum: Beginner to Advanced

## Track goal

Build the **TaskForge CLI and sync/search tooling** in Rust. This track teaches ownership, borrowing, error handling, traits, lifetimes, async, concurrency, performance, and packaging through a practical tool that imports, validates, summarizes, searches, and syncs TaskForge data.

## What you will be able to do

- Use Cargo, rustfmt, Clippy, tests, docs, modules, and crates confidently.
- Design safe APIs with ownership, borrowing, lifetimes, `Option`, `Result`, enums, traits, and generics.
- Build CLIs, file import/export tools, async HTTP clients, and optional Rust services.
- Apply concurrency, performance measurement, and safe systems design.
- Package and present a Rust project as part of a larger product architecture.

## Primary references used

- [The Rust Programming Language Book](https://doc.rust-lang.org/book/)
- [Rust Ownership Chapter](https://doc.rust-lang.org/book/ch04-00-understanding-ownership.html)
- [The Cargo Book](https://doc.rust-lang.org/cargo/)
- [Rust Clippy Documentation](https://doc.rust-lang.org/stable/clippy/)
- [Asynchronous Programming in Rust](https://rust-lang.github.io/async-book/)

## Project milestone map

| Stage | Product feature          | Concepts practiced                      | Evidence of completion                      |
| ----- | ------------------------ | --------------------------------------- | ------------------------------------------- |
| 1     | TaskForge CLI skeleton   | Cargo, modules, tests                   | CLI reads sample JSON and prints summary    |
| 2     | Safe data model          | structs, enums, Result, ownership       | Invalid input returns useful errors         |
| 3     | Reusable library         | modules, traits, generics, lifetimes    | CLI and tests use shared domain library     |
| 4     | Local storage and import | file I/O, serialization, migrations     | Safe local cache with import/export         |
| 5     | Async sync/search        | futures, HTTP, concurrency, performance | CLI syncs with Go API and handles conflicts |
| 6     | Packaged tool/service    | CI, release, profiling, docs            | Installable CLI and optional Rust service   |

## Curriculum modules

Each module is designed as a small product sprint. Do not move on until the deliverable works, has tests or manual verification notes, and is committed with a clear README update.

### Module 1: Setup, Cargo, rustfmt, Clippy, and first CLI

**Level:** Beginner

**Concepts to learn**

- Rust toolchain, Cargo packages, crates, modules, and binaries
- `cargo check`, `cargo build`, `cargo run`, `cargo test`, `cargo fmt`, and `cargo clippy`
- Compiler diagnostics as a learning tool
- Basic CLI input and output

**Project build**

- Create `services/rust-cli` with a `taskforge` command.
- Read a JSON file of tasks and print a summary.
- Add format, lint, test, and run commands to the README.

**Practice tasks**

- Create a library crate and a binary crate in the same package.
- Run Clippy and fix warnings you understand.
- Write a tiny test for task summary formatting.

**Completion checklist**

- [ ] The CLI runs from a clean clone.
- [ ] Cargo commands are documented.
- [ ] You are comfortable reading compiler messages slowly.

### Module 2: Values, control flow, functions, structs, and enums

**Level:** Beginner

**Concepts to learn**

- Variables, mutability, shadowing, scalar and compound types
- Functions, expressions, conditionals, loops, and pattern matching
- Structs and enums
- `Option` and `Result` as core modeling tools

**Project build**

- Model TaskForge tasks, projects, statuses, priorities, and sync states.
- Create summary commands for overdue, blocked, and completed tasks.
- Use enums instead of stringly typed statuses.

**Practice tasks**

- Convert invalid string statuses into typed enum parsing errors.
- Use pattern matching to render status labels.
- Write tests for summary calculations.

**Completion checklist**

- [ ] Task states are represented as enums.
- [ ] Invalid input returns `Result`, not panics.
- [ ] You can explain expression-oriented control flow.

### Module 3: Ownership, moves, borrowing, and references

**Level:** Beginner

**Concepts to learn**

- Ownership rules and moves
- Borrowing with immutable and mutable references
- Slices and borrowing parts of data
- Designing APIs that avoid unnecessary cloning

**Project build**

- Parse task data and pass borrowed references to summary functions.
- Refactor functions that clone too much.
- Add comments explaining ownership decisions around parsed data.

**Practice tasks**

- Create small examples that intentionally fail borrow checking, then fix them.
- Change function signatures from owned values to borrowed values where appropriate.
- Measure clone use by searching the codebase and justifying each clone.

**Completion checklist**

- [ ] Core functions borrow data when they do not need ownership.
- [ ] Clones are intentional.
- [ ] You can explain one borrow checker error you fixed.

### Module 4: Collections, strings, iterators, and data transformations

**Level:** Beginner

**Concepts to learn**

- `Vec`, `HashMap`, `String`, `&str`, and ownership-aware collection use
- Iterators, adapters, and consumers
- Sorting and grouping data
- Avoiding needless allocation

**Project build**

- Implement filters for project, owner, status, due date, and search text.
- Group tasks by project and owner using maps.
- Create iterator-based dashboard summaries.

**Practice tasks**

- Write loop-based and iterator-based versions, then compare readability.
- Handle Unicode search expectations explicitly.
- Add tests for empty collections and missing fields.

**Completion checklist**

- [ ] Transformations are clear and tested.
- [ ] String ownership is handled deliberately.
- [ ] You can explain when iterators improve code and when loops are clearer.

### Module 5: Modules, visibility, crates, and project layout

**Level:** Beginner

**Concepts to learn**

- Modules and visibility rules
- Library versus binary code
- Crate features at a high level
- Public API design

**Project build**

- Split the Rust CLI into domain, parser, commands, storage, and output modules.
- Expose a small library API for task parsing and summaries.
- Keep command-line concerns out of domain logic.

**Practice tasks**

- Move one public function to private and update callers.
- Document public functions with examples.
- Create a small architecture note for module boundaries.

**Completion checklist**

- [ ] Domain logic is reusable outside the CLI.
- [ ] Public APIs are intentionally small.
- [ ] Documentation examples compile or are manually verified.

### Module 6: Error handling with Result, custom errors, and context

**Level:** Intermediate

**Concepts to learn**

- `Result`, `?`, custom error enums, and error conversion
- Library errors versus application errors
- Recoverable versus unrecoverable failures
- Human-friendly CLI output

**Project build**

- Create error types for invalid files, parse failures, validation failures, and sync failures.
- Add helpful messages for CLI users.
- Preserve technical cause chains for debugging.

**Practice tasks**

- Convert one `unwrap` into proper error handling.
- Add tests for failure paths.
- Return different exit codes for common CLI failures.

**Completion checklist**

- [ ] The CLI does not panic on user mistakes.
- [ ] Errors are useful to users and developers.
- [ ] You can explain where `unwrap` is acceptable in tests but not production paths.

### Module 7: Traits, generics, and reusable abstractions

**Level:** Intermediate

**Concepts to learn**

- Traits as behavior contracts
- Generic functions and structs
- Trait bounds and `impl Trait`
- Static dispatch at a practical level

**Project build**

- Create storage traits for local file storage, API storage, and in-memory test storage.
- Create generic output renderers for JSON, text, and table formats.
- Use trait bounds to keep functions flexible but readable.

**Practice tasks**

- Implement a fake storage adapter for tests.
- Compare generics with trait objects for one use case.
- Document why each abstraction exists.

**Completion checklist**

- [ ] Traits are driven by real consumers.
- [ ] Generic code remains readable.
- [ ] Tests can swap storage implementations.

### Module 8: Lifetimes and API design

**Level:** Intermediate

**Concepts to learn**

- Lifetime annotations as relationships between references
- Lifetime elision rules at a practical level
- Owned versus borrowed return values
- Designing structs that borrow data only when it pays off

**Project build**

- Create borrowed task views for filtered summaries.
- Refactor lifetime-heavy code into owned types where it improves clarity.
- Document one lifetime annotation in the code.

**Practice tasks**

- Write examples of returning references incorrectly, then fix them.
- Compare a borrowed view with an owned DTO.
- Use compiler suggestions but verify you understand the relationship being expressed.

**Completion checklist**

- [ ] Lifetime annotations are minimal and meaningful.
- [ ] You avoid borrowing in public APIs when ownership is simpler.
- [ ] You can explain one lifetime in terms of who must outlive whom.

### Module 9: Testing, documentation, fixtures, and CLI verification

**Level:** Intermediate

**Concepts to learn**

- Unit tests, integration tests, and doc tests
- Fixtures and golden files
- Testing command-line behavior
- Documentation as part of API design

**Project build**

- Add tests for parsing, filtering, summaries, errors, and CLI commands.
- Create sample TaskForge export fixtures.
- Add examples to public functions.

**Practice tasks**

- Test a CLI command success and failure case.
- Add a golden output file for table rendering.
- Keep fixtures small and readable.

**Completion checklist**

- [ ] Tests are fast and repeatable.
- [ ] Documentation shows real usage.
- [ ] CLI behavior is verified, not only library logic.

### Module 10: File I/O, serialization, and local storage

**Level:** Intermediate

**Concepts to learn**

- Reading and writing files safely
- Serialization and deserialization
- Atomic writes and backups
- Config files and user data directories

**Project build**

- Store a local TaskForge task cache.
- Import and export JSON compatible with the JavaScript and Go services.
- Add backup before destructive writes.

**Practice tasks**

- Handle missing, unreadable, and malformed files.
- Add a migration for an old local cache format.
- Create tests with temporary directories.

**Completion checklist**

- [ ] File operations fail safely.
- [ ] Local data is compatible with the rest of TaskForge.
- [ ] Cache migration is tested.

### Module 11: Smart pointers, interior mutability, and shared ownership

**Level:** Advanced

**Concepts to learn**

- `Box`, `Rc`, `Arc`, `RefCell`, `Mutex`, and when each appears
- Shared ownership tradeoffs
- Interior mutability as a deliberate escape hatch
- Avoiding overcomplicated pointer graphs

**Project build**

- Add a plugin-style command registry or shared configuration object.
- Use `Arc` for shared configuration in concurrent or async code where needed.
- Refactor away unnecessary shared ownership.

**Practice tasks**

- Build a toy example with `Rc<RefCell<T>>`, then discuss why it may be a smell.
- Use `Arc<Mutex<T>>` only where shared mutation is required.
- Document ownership for shared state.

**Completion checklist**

- [ ] Smart pointers solve real ownership needs.
- [ ] Shared mutable state is limited and documented.
- [ ] You can explain why Rust makes shared mutation explicit.

### Module 12: Concurrency with threads, channels, and synchronization

**Level:** Advanced

**Concepts to learn**

- Threads and scoped work at a high level
- Channels and message passing
- Mutexes and atomics at a practical level
- Data race prevention

**Project build**

- Parallelize validation or indexing of a large task import.
- Send progress messages to the CLI while work continues.
- Cancel or stop work safely when a failure occurs.

**Practice tasks**

- Compare single-threaded and parallel import performance.
- Add tests around deterministic parts and manual checks around timing-heavy parts.
- Document concurrency limits.

**Completion checklist**

- [ ] Concurrent code has clear ownership and shutdown behavior.
- [ ] Progress reporting works during long operations.
- [ ] You can explain why data races are prevented.

### Module 13: Async Rust, futures, streams, and HTTP clients

**Level:** Advanced

**Concepts to learn**

- `async`, `.await`, futures, tasks, and runtimes
- Async HTTP requests
- Streams and backpressure basics
- Cancellation and timeouts

**Project build**

- Add `taskforge sync` to call the Go API asynchronously.
- Fetch projects and tasks, reconcile local cache, and report conflicts.
- Add timeouts and retry policy for safe operations.

**Practice tasks**

- Run multiple API calls concurrently with limits.
- Simulate a slow server and verify timeouts.
- Create conflict examples: local edit versus remote edit.

**Completion checklist**

- [ ] Async code does not block the runtime unnecessarily.
- [ ] Sync behavior handles errors and conflicts.
- [ ] You can explain futures at a practical level.

### Module 14: Rust web service for sync, search, or import processing

**Level:** Advanced

**Concepts to learn**

- HTTP service structure in Rust
- Request extractors and response types in a chosen framework
- Shared state in async services
- Service-level error handling

**Project build**

- Create `services/rust-sync` as a small API for high-throughput import validation, search indexing, or local sync.
- Expose endpoints for validate-import, search, and sync-preview.
- Reuse library code from the CLI.

**Practice tasks**

- Keep framework code at the edges.
- Add tests for pure domain logic and handler-level behavior.
- Document how the Rust service complements the Go API.

**Completion checklist**

- [ ] The service has a clear reason to exist.
- [ ] Domain code is reusable by CLI and API.
- [ ] Errors are mapped consistently to HTTP responses.

### Module 15: Database access, caching, and data consistency

**Level:** Advanced

**Concepts to learn**

- Async database access options
- Connection pools and transactions
- Cache design and invalidation
- Consistency tradeoffs between local and remote data

**Project build**

- Add optional persistence for Rust search index or sync metadata.
- Store sync cursors and conflict records.
- Use transactions where a partial write would be dangerous.

**Practice tasks**

- Write a migration or setup script for local development.
- Test conflict record storage.
- Document consistency guarantees.

**Completion checklist**

- [ ] Persistence has clear ownership and schema rules.
- [ ] Sync metadata survives restarts.
- [ ] Consistency tradeoffs are documented.

### Module 16: Unsafe Rust, FFI, and boundaries

**Level:** Advanced

**Concepts to learn**

- What `unsafe` allows and does not allow
- FFI concepts
- Safety invariants
- Prefer safe abstractions over unsafe shortcuts

**Project build**

- Audit the project and dependencies for any unsafe boundary you can identify.
- Write a small educational example of wrapping an unsafe operation safely, or skip unsafe in production code and document why.
- Create a safety checklist for future contributors.

**Practice tasks**

- Read one safe abstraction that hides unsafe internally.
- Document invariants that a safe wrapper must enforce.
- Avoid adding unsafe code unless there is a measured need.

**Completion checklist**

- [ ] Unsafe is not used casually.
- [ ] Safety invariants are explicit.
- [ ] You can explain why most application Rust needs little or no unsafe code.

### Module 17: Performance, memory, benchmarking, and profiling

**Level:** Advanced

**Concepts to learn**

- Benchmarking methodology
- Allocation awareness
- Profiling CPU and memory
- Tradeoffs between clarity and speed

**Project build**

- Benchmark import validation, search indexing, and sync reconciliation.
- Optimize one hot path using evidence.
- Write a performance report comparing baseline and optimized versions.

**Practice tasks**

- Replace one avoidable allocation and measure the effect.
- Compare sequential, threaded, and async approaches for the same workload.
- Keep the clearest version when performance gains are insignificant.

**Completion checklist**

- [ ] Performance claims have measurements.
- [ ] Optimizations do not obscure core logic unnecessarily.
- [ ] You can explain memory ownership effects on performance.

### Module 18: Packaging, releases, CI, and cross-platform distribution

**Level:** Advanced

**Concepts to learn**

- Release builds
- Cross-platform CLI considerations
- CI for format, lint, test, and build
- Versioning and changelogs

**Project build**

- Create release builds for the Rust CLI.
- Add CI checks for `cargo fmt`, `cargo clippy`, tests, and build.
- Write install and upgrade instructions.

**Practice tasks**

- Test the CLI from a clean directory.
- Create a sample release archive.
- Write a changelog entry and rollback note.

**Completion checklist**

- [ ] The CLI can be installed and used by another person.
- [ ] CI catches formatting, lint, and test failures.
- [ ] Release notes explain user-visible changes.

### Module 19: Rust capstone: TaskForge CLI and sync/search service

**Level:** Capstone

**Concepts to learn**

- End-to-end Rust delivery
- Ownership, errors, traits, async, concurrency, performance, packaging
- Safe systems design
- Portfolio documentation

**Project build**

- Complete a Rust CLI that imports, validates, summarizes, searches, and syncs TaskForge data.
- Optionally complete a Rust service for import validation, sync preview, or search indexing.
- Package the CLI and document architecture, safety, and performance decisions.

**Practice tasks**

- Run a large import and sync demo against the Go API.
- Create a conflict scenario and resolve it.
- Review with the assessment rubric.

**Completion checklist**

- [ ] The CLI is reliable, helpful, and well tested.
- [ ] Async sync and local file handling fail safely.
- [ ] The capstone demonstrates why Rust was chosen for this part of the system.

## Rust capstone specification

Build `services/rust-cli` and optionally `services/rust-sync`.

Minimum features:

- CLI commands for import, validate, summarize, search, export, and sync.
- Typed domain model using enums and `Result`-based errors.
- Safe file handling with backups or atomic writes.
- Reusable library code separated from CLI concerns.
- Async sync with the Go API and clear conflict handling.
- Tests for parsing, validation, summaries, storage, and core sync decisions.
- Format, Clippy, tests, release build, and install instructions.
- Performance report for large import or search workloads.

Stretch features:

- Rust search/indexing microservice.
- Streaming import of very large files.
- Conflict resolution strategies with preview and apply phases.
- Cross-platform release artifacts.
- Safe FFI wrapper demonstration, only as an educational extension.

## Final review questions

- Which functions own data, which borrow it, and why?
- Where did you choose clearer owned data instead of complicated lifetimes?
- How does the CLI avoid data loss during failed writes or sync conflicts?
- Which workload did you measure, and what optimization was actually justified?
- Why is Rust a good fit for this part of TaskForge?
