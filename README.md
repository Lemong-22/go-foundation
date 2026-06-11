# go-foundation

> **Purpose:** Sandbox untuk belajar Go + Postgres + Hexagonal Architecture dengan reference read-only dari `entropy-course` (project mentor Stephen Antoni). Insight yang dapat di-apply ke `LE-REMINDER` (TS/Bun/Next.js) di phase berikutnya.

**Status:** Week 1 — Foundation. Lihat `plans/WEEK-1-2-PHASE-A.md` untuk plan eksekusi.

**Owner:** Lemong-22 (Yosia EH). Belajar mandiri, supervised oleh AI agent (VIN).

---

## ⚠️ Read-Only Vendor Notice

Folder `vendor/entropy-course/` adalah **read-only copy** dari project mentor. Tujuan:
- Referensi baca saat belajar (buka source code, baca docs, pelajari pattern).
- Bukti bahwa pattern yang gw pelajari teruji di production-grade project.

**DILARANG KERAS** di folder `reference/`:
- ❌ Edit file
- ❌ Commit perubahan
- ❌ Run migrations / write ke DB dari path ini
- ❌ `git push` ke `luxeave/entropy-course` (sudah di-set ke remote yg ga reachable)

Semua code project gw ada di `cmd/`, `internal/`, `migrations/`, `scripts/`. Folder `reference/` punya `chmod 555` (read-only) sebagai safeguard.

---

## Quick Reference

| Path | Isi |
|---|---|
| `plans/` | Plan eksekusi per minggu (week 1, 2, dst) |
| `journal/` | Progress log harian (apa yang dipelajari, blocker, decisions) |
| `templates/` | Template PRD, SPEC, ROADMAP, DESIGN yang bisa lo copy-paste untuk project lain |
| `reference/entropy-course/` | READ-ONLY reference code (renamed from `vendor/` to avoid Go's vendor mechanism confusion) |
| `cmd/` | Entry point binary (untuk sekarang: `cmd/hello`, `cmd/counter`) |
| `internal/` | Hexagonal core (domain, ports, usecase, adapter) — dikembangkan per minggu |
| `migrations/` | SQL migration files |
| `scripts/` | Helper scripts (setup DB, dll) |

---

## Tech Stack

| Concern | Choice | Versi |
|---|---|---|
| Language | Go | 1.22.2 toolchain, projects may declare `go 1.24.0` minimum |
| CLI framework | `cobra` + `viper` | v1.8.1 / v1.19.0 |
| Database | PostgreSQL | 16 (canister23 local) |
| DB driver | `pgx` v5 | v5.5.5 |
| Migrations | `golang-migrate` | (planned Week 2) |
| UUID | `google/uuid` | v1.6.0 |
| HTTP router | `net/http` stdlib | (planned Week 2) |

---

## Cara Baca Repo Ini (Untuk Lo di Masa Depan)

1. **Mau tau "gimana caranya X"** → buka `journal/2026-WEEK-XX.md` (lo paling sering lupa detail teknis)
2. **Mau tau "kenapa arsitektur-nya begitu"** → buka `vendor/entropy-course/docs/spec.md` + `vendor/entropy-course/docs/roadmap.md` (pembenaran keputusan ada di sana)
3. **Mau tau "phase sekarang lo lagi di mana"** → buka `plans/WEEK-X-Y-PHASE-Z.md` (current week + deliverable)
4. **Mau mulai project baru** → copy dari `templates/` (PRD-TEMPLATE, SPEC-TEMPLATE, dll)

---

## Setup Lokal

```bash
# 1. Install Go 1.22.2 (kalau belum)
sudo apt install -y golang-1.22
export PATH=/usr/lib/go-1.22/bin:$PATH
echo 'export PATH=/usr/lib/go-1.22/bin:$PATH' >> ~/.bashrc

# 2. Setup Postgres database untuk belajar
sudo -u postgres psql <<'SQL'
CREATE DATABASE go_foundation;
CREATE USER go_foundation WITH PASSWORD 'devpass';
GRANT ALL PRIVILEGES ON DATABASE go_foundation TO go_foundation;
\c go_foundation
GRANT ALL ON SCHEMA public TO go_foundation;
SQL

# 3. Setup env
export DATABASE_URL="postgres://go_foundation:devpass@localhost:5432/go_foundation?sslmode=disable"
echo 'export DATABASE_URL="postgres://go_foundation:devpass@localhost:5432/go_foundation?sslmode=disable"' >> ~/.bashrc

# 4. Verify
cd ~/vin-workspace/projects/go-foundation
go version
go run ./cmd/hello
```

---

## Progress Tracker

| Week | Phase | Status | Started | Completed | Notes |
|---|---|---|---|---|---|
| 1 | Foundation (Go + Postgres + minimal API) | 🟡 In progress | 2026-06-11 | — | see `journal/2026-WEEK-23.md` |

Lihat `journal/INDEX.md` untuk link ke weekly logs.
