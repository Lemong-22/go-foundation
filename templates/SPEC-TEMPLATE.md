# Technical Spec — [Bounded Context Name]

**Status:** Draft | Approved | Superseded  
**Architecture:** [Hexagonal + Clean + DDD | Layered | Microservice]  
**Source:** `[PRD file]` vX.Y.Z · Spec author: [nama lo], YYYY-MM-DD

> **Template origin:** Di-extract dari `vendor/entropy-course/docs/spec.md` (Stephen Antoni, 2026-05-24).

---

## 0. Overview

- **Purpose:** [1-2 kalimat: apa yang context ini handle, sebagai authoritative backend untuk apa]
- **Language / stack:** [Bahasa, framework, driver]
- **In scope:** [List 3-5 hal yang JELAS dalam scope]
- **Out of scope:**
  - [Concern A] — separate [nama context]. This context references by id only.
  - [Concern B] — infrastructure tooling, ships in same binary but wired outside this context.
  - [Concern C] — deferred. See §6.

**[Catatan tentang adapters].** [Jelaskan boundary antara adapter dan core. Misal:] CLI flags, config parsing, output formatters, exit codes are all part of the **inbound CLI adapter** — they are not domain logic. The core returns typed domain errors; the CLI adapter is the only component that knows the exit-code table.

---

## 1. CLI Commands

[Opsional — kalau project CLI. Tabel sangat membantu.]

The activities a user can perform against this context. There are [N] commands; each maps to **exactly one inbound port method and one usecase**.

### [Entity A] commands

| Command | Description | Inputs | Success output | Failure modes |
|---------|-------------|--------|----------------|---------------|
| `entity_a create` | Create a new entity | `--field-1`, `--field-2` (opt) | prints new id | missing field, invalid format |
| `entity_a list` | List entities, newest first | `--filter` (opt), `-o/--output` | table / JSON | — |
| `entity_a get` | Show one entity's detail | `<id>`, `-o/--output` | entity detail | id not found |
| `entity_a update` | Edit entity | `<id>`, `--field-1` (opt) | prints updated id | id not found, nothing to update |
| `entity_a delete` | Delete entity | `<id>`, `--force` (opt) | confirmation | id not found |

---

## 2. Domain Model

[Aggregate = consistency boundary. DDD rule: aggregate references another aggregate by id only, never by holding the object.]

### Aggregates

**[Entity A]** — [1 kalimat: apa ini, role-nya apa]

- Identity: `ID EntityAID`
- Fields:
  - `Field1 string`
  - `Field2 Field2`  (VO)
  - `Status Status`  (VO)
  - `CreatedAt time.Time`, `UpdatedAt time.Time`
- Invariants:
  - `Field1` is non-empty after trimming.
  - `Field2` is always valid (guaranteed by VO).
  - `Status` is one of `draft` | `active` (guaranteed by VO).
  - `UpdatedAt >= CreatedAt`.
- Behaviour (each mutating method takes current time and bumps `UpdatedAt`; entity holds no `Clock` — timestamp passed in by usecase):
  - `Rename(name string, now time.Time) error` — re-validates non-empty.
  - `ChangeField2(f Field2, now time.Time)`.
  - `Activate(now time.Time) error` — sets `Status` to `active`; errors with `ErrAlreadyActive`.

**[Entity B]** — [sama format]

### Value Objects

All VOs are immutable; construction (`NewX`) fails with domain error when invariant violated. Existing VO is always valid.

- **EntityAID** — wraps `string`; invariant: non-empty, parseable as UUID.
- **Field2** — wraps `string`; invariant: matches regex `^...$`. *Uniqueness NOT a VO invariant* — cross-row rule enforced by usecase via repository.
- **Status** — wraps `string`; invariant: one of `draft` | `active`. Exposes `Draft()` / `Active()` constructors and `IsActive()` predicate.
- **Order** — wraps `int`; invariant: `>= 0`.

### Domain notes

- **Vertical slices.** Each command is its own slice through the stack: CLI → inbound port → usecase → outbound ports → adapter. No command shares a usecase with another.
- **[Plain string field] stays plain string.** A field with no invariant beyond non-empty doesn't need a VO. Wrapping is ceremony without payoff.
- **Cross-aggregate delete.** [When A delete must also remove A's children]: not a silent DB cascade hidden from domain — `DeleteA` usecase explicitly orchestrates across both repos.
- **Timestamps are passed in, not pulled.** Entities never call a clock. Constructors and mutating methods receive `now time.Time`; the usecase gets it from `Clock` outbound port. Keeps domain pure, makes time deterministic in tests.
- **Domain errors.** Core returns typed sentinel errors — `ErrNotFound`, `ErrAlreadyTaken`, `ErrAlreadyActive`, and validation errors from VO constructors. CLI adapter maps to exit codes; core itself knows nothing about exit codes.

---

## 3. Ports

### Inbound ports (driving — called by adapters)

One interface per noun, mirroring CLI grouping. Every method corresponds to exactly one command from §1.

```go
// EntityAService is the inbound port for all entity_a commands.
type EntityAService interface {
    CreateEntityA(in CreateEntityAInput) (CreateEntityAOutput, error)
    ListEntityAs(in ListEntityAsInput) (ListEntityAsOutput, error)
    GetEntityA(in GetEntityAInput) (GetEntityAOutput, error)
    UpdateEntityA(in UpdateEntityAInput) (UpdateEntityAOutput, error)
    DeleteEntityA(in DeleteEntityAInput) error
}

// EntityBService is the inbound port for entity_b commands.
type EntityBService interface {
    CreateEntityB(in CreateEntityBInput) (CreateEntityBOutput, error)
    // ...
}
```

Input/output DTOs carry only primitives and ids — boundary between (untyped) CLI world and (VO-typed) domain. Usecases build VOs from primitives and return validation errors when construction fails.

```go
// --- EntityA DTOs ---
type CreateEntityAInput struct {
    Field1 string
    Field2 string
    // ...
}
type CreateEntityAOutput struct{ ID string }

type UpdateEntityAInput struct {
    ID     string
    Field1 *string // nil = leave unchanged
    Field2 *string
}

// EntityAView is a flat read-model for output formatting.
type EntityAView struct {
    ID, Field1, Field2, Status string
    CreatedAt, UpdatedAt time.Time
}
```

### Outbound ports (driven — implemented by adapters)

Declared by core, depend only on domain types, implemented on outside.

```go
type EntityARepository interface {
    Save(e EntityA) error                       // INSERT or UPDATE
    FindByID(id EntityAID) (EntityA, error)      // ErrNotFound if absent
    FindAll(filter EntityAFilter) ([]EntityA, error)
    Delete(id EntityAID) error                   // ErrNotFound if absent
}

type IDGenerator interface {
    NewEntityAID() EntityAID
}

type Clock interface {
    Now() time.Time
}
```

**The dependency rule.** Every port above is declared inside core and speaks only domain types (`EntityA`, `EntityAID`, `Field2`, ...). Core says "I need somewhere to save an EntityA"; never "I need Postgres." Adapters on outside implement these interfaces. Dependencies point inward.

---

## 4. Adapters & Usecases

### Outbound adapters (side effects)

**PostgresEntityARepository** implements `EntityARepository`
- Construction dependencies: `*pgxpool.Pool`
- `Save` — `INSERT INTO entity_a ... ON CONFLICT (id) DO UPDATE`
- `FindByID` — `SELECT ... FROM entity_a WHERE id = $1`
- Maps "0 rows" to `ErrNotFound`; maps unique-constraint violation to `ErrAlreadyTaken` as backstop to usecase's pre-check.

**UUIDGenerator** implements `IDGenerator`
- `NewEntityAID` — pure; wraps `google/uuid` v4 in the id VO

**SystemClock** implements `Clock`
- `Now` — returns `time.Now().UTC()`

> The **inbound CLI adapter** (cobra commands) is also an adapter: it parses flags/config into input DTOs, calls inbound ports, formats output, runs `--force` prompts, maps domain errors to exit codes. It holds no business logic.

### Usecases (implement inbound ports, depend on outbound ports)

Each usecase implements one inbound port method, is constructor-injected with outbound ports (interfaces, never concrete adapters), and holds application logic.

**CreateEntityA** — implements `EntityAService.CreateEntityA`
- Depends on: `EntityARepository`, `IDGenerator`, `Clock`
- Steps:
  1. build `Field2` VO from `in.Field2` — return validation error if invalid
  2. `EntityARepository.FindByField2(...)` — if exists, return `ErrAlreadyTaken`
  3. `id := IDGenerator.NewEntityAID()`; `now := Clock.Now()`
  4. construct `EntityA` via `NewEntityA(id, in.Field1, field2, now)` — validates non-empty field1
  5. `EntityARepository.Save(entityA)`
  6. return `CreateEntityAOutput{ ID: id.String() }`

**[Other usecases]** — [sama format]

---

## 5. Container & Wiring

The composition root — the single place where concrete types are named and wired. Everything else in the spec speaks only in interfaces.

```go
type Config struct {
    DBURL string // resolved by config: flag > env > config file > default
}

type CLI struct {
    EntityA EntityAService
    EntityB EntityBService
}

func BuildContainer(ctx context.Context, cfg Config) (*CLI, error) {
    // 1. construct outbound adapters from config
    pool, err := pgxpool.New(ctx, cfg.DBURL)
    if err != nil {
        return nil, fmt.Errorf("connect db: %w", err)
    }
    entityARepo := NewPostgresEntityARepository(pool)
    ids := NewUUIDGenerator()
    clock := NewSystemClock()

    // 2. construct usecases — adapters injected through outbound ports only
    entityASvc := NewEntityAServiceImpl(entityARepo, ids, clock)

    // 3. bind usecases to adapters (CLI/REST)
    return &CLI{EntityA: entityASvc}, nil
}
```

Every usecase receives repositories / `Clock` / `IDGenerator` as **interfaces**, never concrete `*PostgresEntityARepository`. Payoff: swap Postgres for in-memory fake in tests — change 2 lines in `BuildContainer`, nowhere else.

The `migrate` commands and future `auth` commands are wired in `main.go` *outside* `BuildContainer` — they are not part of this bounded context.

---

## 6. Deferred

Judgment-call items considered during design and deliberately not built now. Parked here in writing, not forgotten.

| Item | Kind | Why deferred |
|------|------|--------------|
| [Feature X] | anticipatory | [Reason] |
| [Auth] | out of context | Belongs to separate [Identity] context. This context references by [X]ID only. |
| [Concern] | optimization | [Reason] |
| [Soft delete] | anticipatory | [Reason] |

---

## Catatan Pakai Template Ini

1. **§0 In/Out scope** — tulis DULUAN. Kalau ga jelas, stop. Diskusi scope dulu, baru nulis detail.
2. **§2 Aggregates** — aggregate = consistency boundary. JANGAN bikin aggregate terlalu besar (god aggregate) atau terlalu kecil (anemic).
3. **§2 VOs** — VO yang baik punya 1 invariant clear. Kalau lo ga bisa narasi "VO X guarantees Y", lo ga butuh VO itu.
4. **§3 Inbound ports** — 1 port per noun, 1 method per command. Kalau ada 2 method yang sama, kemungkinan lo butuh pecah aggregate.
5. **§3 Outbound ports** — declare interface, never concrete. Kalau lo nulis `*PostgresX`, lo udah salah.
6. **§4 Usecases** — pure orchestration, no I/O logic (delegate ke adapters via ports).
7. **§6 Deferred** — TULIS. Keputusan "ga dibangun" = keputusan. Dokumentasiin alasannya. Lo bakal lupa 3 bulan lagi.
