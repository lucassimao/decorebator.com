# Repository Guidelines

## Project Structure & Module Organization
- `api/`: Go 1.23 backend (Gin) with layers under `internal/*`; migrations in `cmd/migrate/migrations/`; workers in `cmd/workers/`; tests split between `internal/tests/unit` and `tests/integration`.
- `mobile/`: Expo app under `app/` with shared UI in `components/`, translations in `i18n/`, theming in `theme/`, media in `assets/`, and Jest mocks in `__mocks__/`.
- `web/`: Next.js App Router in `src/`; localized copy in `messages/`, marketing assets in `public/`.
- `docs/`, `todo/`: design notes and backlog context—update alongside feature work.

## Build, Test, and Development Commands
- **API backend**: `cd api && make setup` installs tools; `make test` runs the dockerized suite; `make lint`/`make format-check` match CI; use `docker compose -f docker-compose.yml up` for services, then `make run` (and `make workers` in another terminal).
  - **Note**: Always check `api/Makefile` for the most up-to-date targets (e.g., `lint-changed`, `lint-staged`, `format`, `test-unit`).
- **Mobile app**: `cd mobile && npm install`; `npm run start` launches Expo; run `npm run lint`, `npm run typecheck`, and `npm test` (Jest + Testing Library) before pushing.
- **Web app**: `cd web && npm install`; `npm run dev` serves Next (port 4000); `npm run build` creates production output and sitemap; `npm run lint` or `npm run format:check` guard style.
  - **Note**: Next 16 uses the new `proxy` convention (replaces `middleware`) and `next lint` is deprecated; use `npx eslint .` with the repo’s flat config if linting is needed.

## Coding Style & Naming Conventions
- **Go**: Run `make format` (gofmt + goimports) and `make lint` (golangci-lint); use PascalCase for exports, camelCase for locals, SCREAMING_SNAKE_CASE for env vars.
- **TypeScript/Tailwind**: Prettier (see `web/.prettierrc.json`) sets 2-space indent, single quotes, Tailwind class sorting; keep PascalCase components, camelCase utilities, and pull design tokens from `theme/` or `styles/` instead of hard-coded values.

## Testing Guidelines
- **API**: Keep unit specs in `internal/tests/unit`, integration specs in `tests/integration`, and enforce ≥70% coverage with `make coverage-threshold`.
- **Mobile**: Co-locate `.test.tsx` files with components, use Testing Library queries, and mock network or storage via `__mocks__/`.
- **Web**: No automated suite yet—run `npm run lint` and add Playwright or Vitest when UI logic expands.

## Commit & Pull Request Guidelines
- Write concise, imperative commit subjects and describe intent in the body.
- Ensure lint, format, and tests pass locally; Husky scripts assume staged files are clean.
- PRs should outline scope, affected modules, linked issues, and include UI screenshots or API contract notes when relevant; call out environment or schema changes.

## Configuration & Secrets
- API expects `.env` values for Postgres, Redis, MinIO, OpenAI, SendGrid, and Stripe; bootstrap from `.env.example` before running `make` targets.
- Mobile expects Expo public env vars from `mobile/.env.example` (API URL, RevenueCat keys, PostHog, Sentry, optional app domain).
- Web uses `NEXT_PUBLIC_API_URL`, `STATIC_AUTHENTICATION`, and `SITE_URL` for public-quiz fetching/sitemaps.
- Store secrets outside version control and use `docker-compose.override.yml` or Expo/EAS secrets for local overrides—never commit real keys.
