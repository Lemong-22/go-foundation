# PRD: [Project Name]

**Status:** Draft | Approved | Superseded  
**Author:** [Nama lo]  
**Date:** YYYY-MM-DD  
**Version:** 0.1.0

> **Template origin:** Di-extract dari `vendor/entropy-course/docs/course-cli-prd.md` v0.1.0 (Stephen Antoni, 2026-05-24). Sesuaikan dengan konteks lo.

---

## 1. Overview

[2-3 kalimat: apa produknya, untuk siapa, kenapa penting. JANGAN list fitur di sini — itu masuk Goals/Non-Goals.]

## 2. Goals

- [Goal 1 — measurable, spesifik]
- [Goal 2]
- [Goal 3]

## 3. Non-Goals

- [Yang JELAS bukan scope. Penting untuk nyegah scope creep. Misal: "No web UI in this phase" / "No multi-tenancy in v1"]

## 4. User Roles

[Opsional. Kalau ada permission model. Gunakan tabel seperti mentor:]

| Capability | Role A | Role B | Admin |
|---|---|---|---|
| [Action 1] | ✅ | ❌ | ✅ |
| [Action 2] | ✅ | ❌ | ✅ |

## 5. Data Model

### 5.1 [Entity A]

```sql
CREATE TABLE entity_a (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  field_1    TEXT NOT NULL,
  field_2    TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

**Fields:**
- `id` — [deskripsi]
- `field_1` — [deskripsi + kenapa required]
- `field_2` — [deskripsi + kenapa optional]

### 5.2 [Entity B]

[Sama format]

---

## 6. Command Reference

[Opsional — kalau project CLI-based, list commands. Tabel sangat membantu.]

Binary name: `[project-name]`
Command pattern: `[project-name] <noun> <verb> [args] [flags]`

### 6.1 [Entity A] Commands

```
project-name entity_a create   --field-1 <val> [--field-2 <val>]
project-name entity_a list     [--filter <val>] [--output json]
project-name entity_a get      <id>          [--output json]
project-name entity_a update   <id>          [--field-1 <val>]
project-name entity_a delete   <id>          [--force]
```

| Command | Description | Roles |
|---|---|---|
| `entity_a create` | Create new entity | All |
| `entity_a list` | List entities | All |

---

## 7. Global Flags

| Flag | Short | Description |
|---|---|---|
| `--output` | `-o` | Output format: `table` (default) or `json` |
| `--quiet` | `-q` | Print only ID (for piping) |
| `--force` | `-f` | Skip confirmation on destructive commands |
| `--verbose` | `-v` | Debug info, stack traces, SQL |
| `--config` | | Path to config file |

---

## 8. Output Formats

### Default (table)
```
ID        NAME              STATUS
────────────────────────────────────
abc123    Entity One        active
def456    Entity Two        draft
```

### JSON (`--output json`)
```json
[
  {"id": "abc123", "name": "Entity One", "status": "active"}
]
```

**Rule:** All data output → **stdout**. All errors/warnings → **stderr**.

---

## 9. Error Handling & Exit Codes

| Scenario | Exit Code | Example stderr |
|---|---|---|
| Success | `0` | — |
| Validation error | `1` | `Error: --title is required` |
| Not found | `2` | `Error: entity abc123 not found` |
| Permission denied | `3` | `Error: permission denied` |
| Internal / DB error | `5` | `Error: internal error (run with --verbose for details)` |

**Destructive command confirmation:**
```bash
$ project-name entity_a delete abc123
Delete entity "Entity One"? [y/N]:
```
Bypassed dengan `--force`.

---

## 10. [Concern] (Deferred / Future)

> ⚠️ **TODO: [concern] deferred to future iteration.**

[Jelaskan apa yang lo skip, kenapa, dan approach kalau nanti di-implement.]

---

## 11. Configuration

Config file location: `~/.config/<project>/config.yaml`

```yaml
# db_url: postgres://user:***@localhost:5432/dbname
# auth_token: <deferred>
```

Priority order (highest to lowest):
1. CLI flags
2. Environment variables
3. Config file
4. Defaults

---

## 12. Tech Stack

| Concern | Choice |
|---|---|
| Language | [Go / TypeScript / Python] |
| CLI framework | [cobra+viper / commander / click] |
| Database | [PostgreSQL / SQLite / MongoDB] |
| DB driver | [pgx / prisma / motor] |
| Migrations | [golang-migrate / drizzle / alembic] |
| [Concern lain] | [Choice] |

---

## 13. Project Structure (Proposed)

```
project-name/
├── cmd/
│   ├── root.go           ← global flags, config loading
│   ├── entity_a.go       ← entity_a subcommands
│   ├── entity_b.go       ← entity_b subcommands
│   └── migrate.go        ← migration subcommands
├── internal/
│   ├── db/               ← connection, query helpers
│   ├── domain/           ← Entities, VOs
│   ├── repository/       ← DB queries per entity
│   ├── service/          ← business logic
│   └── output/           ← formatters
├── migrations/           ← SQL files
├── main.go
└── go.mod
```

---

## 14. Success Metrics

- [ ] [Measurable outcome 1]
- [ ] [Measurable outcome 2]
- [ ] [Measurable outcome 3]
- [ ] `migrate up` works on fresh DB

---

## 15. Out of Scope / Future Iterations

- [Item 1]
- [Item 2]
- [Item 3]

---

## Catatan Pakai Template Ini

1. **Hapus section yang ga relevan** untuk project lo. Misal: project non-CLI, hapus §6 dan §9.
2. **JANGAN skip §3 Non-Goals** — ini yang sering lupa ditulis, dan ini yang bedain PRD bagus vs PRD asal.
3. **§5 Data Model SQL langsung** — biar lo & reviewer bisa langsung verify schema.
4. **§10 Deferred** — dokumentasi "yang ga dibangun" sama pentingnya dengan "yang dibangun". Bikin keputusan terlihat.
5. **§14 Success Metrics** — kalau ga ada metrics, lo ga bisa tau "done" kapan.
