# AGENTS.md

## Repo Shape

- This is a two-app Go monorepo: `backend/` is the JSON API, `frontend/` is the Go-rendered HTMX app.
- Deploy `backend/` and `frontend/` as separate Vercel projects; each has its own `go.mod` and `vercel.json`.
- Run Go commands from the relevant app directory, not the repo root.

## Commands

- Backend verification: `cd backend && go test ./...`
- Frontend verification: `cd frontend && go test ./...`
- Backend dev server: `cd backend && go run ./cmd/server`
- Frontend dev server: `cd frontend && go run ./cmd/server`
- Format changed Go files with `gofmt`; there is no repo-level task runner or CI config.

## Local Environment

- Local dev servers load `.env` from their own working directory.
- Backend needs `backend/.env` with `DATABASE_URL` and usually `ALLOWED_ORIGIN=http://localhost:3000`.
- Frontend needs `frontend/.env` with `API_BASE_URL=http://localhost:8080`.
- PostgreSQL migrations must be run manually and in order:
  - `psql "$DATABASE_URL" -f backend/migrations/001_init.sql`
  - `psql "$DATABASE_URL" -f backend/migrations/002_seed.sql`

## Backend Notes

- Vercel API entrypoints live under `backend/api/*/index.go`; local routing is wired manually in `backend/cmd/server/main.go`.
- Shared backend DB code is in `backend/internal/db/db.go`; it uses `DATABASE_URL` via `pgxpool`.
- The app is intentionally single-user for now. Do not add auth/user tables unless explicitly requested.

## Frontend Notes

- Frontend routes and HTMX action handlers are wired in `frontend/internal/app/app.go`.
- Templates and CSS are embedded from `frontend/assets/`; edit files under `frontend/assets/templates/` and `frontend/assets/static/`.
- Keep embedded asset paths compatible with `frontend/assets/assets.go`; Go embed cannot include files outside that package tree.
- Frontend tests use a fake backend, so `cd frontend && go test ./...` does not require PostgreSQL.

## Common Gotchas

- A page banner like `backend returned 500...` usually means the backend lacks `DATABASE_URL`, cannot reach Postgres, or migrations were not run.
- Adding a backend endpoint requires both a Vercel function under `backend/api/` and a local route in `backend/cmd/server/main.go`.
- Adding a frontend route requires updating `ServeHTTP` in `frontend/internal/app/app.go`; Vercel routes everything through `frontend/api/index.go`.
