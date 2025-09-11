# Repository Guidelines

## Project Structure & Module Organization
- `api/` — Go backend (Gin); migrations in `cmd/migrate/migrations/`, business code in `internal/`.
- `mobile/` — React Native (Expo) app; routes in `app/`, UI in `components/`.
- `web/` — Next.js 15 site; app routes in `src/`, static assets in `public/`.
- `docs/` — Additional documentation. See `README.md` for architecture details.

## Build, Test, and Development Commands
- API
  - `cd api && make watch` — Run API with auto-reload.
  - `cd api && make workers` — Start background workers.
  - `cd api && make migrate-up` — Apply DB migrations.
  - `cd api && make test` — Integration tests in Docker; writes `coverage.out`.
- Mobile
  - `cd mobile && npm start` — Expo dev server.
  - `cd mobile && npm test` — Jest test runner.
  - `cd mobile && npm run lint` — ESLint; `lint-staged` + Prettier on commit.
- Web
  - `cd web && npm run dev` — Next.js dev server.
  - `cd web && npm run build && npm start` — Production build and start.

## Coding Style & Naming Conventions
- Go (api): `golangci-lint` configured via `api/.golangci.yml`; run `make lint`. Use `go fmt`, idiomatic naming (`CamelCase` for exported, `camelCase` for locals), small cohesive packages.
- TypeScript/JS (web, mobile): ESLint + Prettier.
  - Prettier: 2 spaces, single quotes, no semicolons, 100 char width.
  - Filenames: `PascalCase` for components, `camelCase` for hooks/utils, `kebab-case` for routes.

## Testing Guidelines
- API: Prefer table-driven tests under `api/internal/...` and `api/tests`. Coverage gates exist; keep unit ≥70% and integration ≥80% (see `make coverage-threshold`).
- Mobile: Place tests in `mobile/__tests__/` or `*.test.ts(x)`. Use Testing Library and `jest-setup.js`.
- Web: Linting required; add tests if introducing complex logic.
- Run locally before PR: `make test` (api), `npm test` (mobile).

## Commit & Pull Request Guidelines
- Commits: Follow Conventional Commit style where practical: `feat:`, `fix:`, `chore:`, with optional scope (`web:`, `api:`). Keep messages imperative and focused.
- PRs must:
  - Describe the change, rationale, and impact.
  - Link issues (e.g., `Closes #123`).
  - Include screenshots/GIFs for UI changes (mobile/web).
  - Pass CI (lint, tests, coverage). No secrets in VCS; use `.env`/`.env.example`.

## Security & Configuration Tips
- Never commit real API keys. Copy from `.env.example` files and set local `.env`/`.env.local`.
- For local infra, use `api/docker-compose.yml` and run migrations before starting services.
