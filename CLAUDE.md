# CLAUDE.md — Monetra

Personal & family finance tracker. Product scope, API contract, data model, design tokens, and build order live in **`PRD.md` — read it before any task**.

## Repo
- `frontend/` — Next.js 16 (App Router), TypeScript, Tailwind CSS, shadcn/ui, Zod, NextAuth v5 (session layer only)
- `backend/` — Go ≥ 1.22, `net/http` router, clean architecture (`internal/domain` → `service` → `repository` / `http`)

## Working rules
- Respond in Indonesian. Code, comments, and commit messages in English.
- Read the actual files before changing them. Never assume file, folder, DTO, or schema structure.
- Before editing, explain the approach briefly and wait for approval. Edit files or run commands only when Arya says so.
- If there is a simpler or better option than what was asked, say so with a short reason before implementing.
- Only install dependencies listed in `PRD.md` §6. Ask before adding anything else.
- Work in the stage order of `PRD.md` §10; one feature slice at a time (Go endpoint → Next.js UI).
- Deliver working code: no placeholders, no TODOs, no dead code, no unused imports.

## Code principles
- **KISS** — choose the simplest solution that works; no clever code.
- **YAGNI** — build only what the current stage needs; no speculative options, layers, or config.
- **SoC** — HTTP handling, business logic, and data access stay in separate layers; UI components don't fetch or hold business rules.
- **DRY** — extract shared logic once it is truly repeated (rule of three); don't abstract prematurely, KISS wins over DRY.
- **SOLID** — small single-purpose units; depend on interfaces defined in `domain`, injected from `cmd/api/main.go`.

## Domain rules (must never break)
- Money is `BIGINT` Rupiah (`int64` in Go, `number` in TS). Never use float.
- Any create/update/delete of a transaction adjusts the account balance inside one DB transaction.
- Every query is scoped to the authenticated `user_id`.
- Month boundaries use `Asia/Jakarta`.

## Conventions
- Go: errors wrapped with context (`fmt.Errorf("create transaction: %w", err)`); handlers return the standard response/error envelope from `PRD.md` §8; business-logic tests live next to `service/`.
- Web: mobile-first, accessible (labels, AA contrast, 44 px touch targets); format Rupiah with `Intl.NumberFormat('id-ID')`; use design tokens from `PRD.md` §9, never raw hex in components.