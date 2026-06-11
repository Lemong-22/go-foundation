# Course Management Test Coverage

The Course Management bounded context is tested independently from the rest of
the repository. Use this focused command for validation:

```bash
go test ./internal/course/...
```

Coverage is split by hexagonal boundary:

- `internal/course/domain`: value object invariants, aggregate construction,
  aggregate mutation behavior, and deferred domain concepts that must remain
  absent for this iteration.
- `internal/course/usecase`: all 13 command usecases with fake repositories,
  fake ID generation, and fake clocks.
- `internal/course/adapter/postgres`: PostgreSQL row mapping, schema alignment,
  and lesson reorder transaction behavior where practical without requiring a
  live database.
- `internal/course/adapter/cli`: command parsing for the 13 spec command flows,
  output format selection, destructive confirmation behavior, and error-to-exit
  code mapping.
- `internal/course/app`: Viper-compatible config resolution, provider
  construction, and composition-root wiring.

Deferred scope is intentionally asserted by tests rather than implemented:
archived course commands/status, lesson publish status/commands, identity/admin
commands, auth enforcement, and migration command wiring remain outside the
Course bounded context.
