# Dimoo Tracker

A single-user Pop Mart figurine tracker with a Go backend, PostgreSQL database, and a Go-rendered HTMX frontend. The frontend and backend are designed to deploy as separate Vercel projects.

## Structure

```text
backend/
  api/                 Vercel Go serverless API functions
  cmd/server/          local backend server
  internal/            database, JSON helpers, models
  migrations/          PostgreSQL schema and seed data

frontend/
  api/                 Vercel Go serverless frontend function
  assets/              embedded HTML templates and CSS
  cmd/server/          local frontend server
  internal/            app router and backend client
```

## Features

- Shelf page for featured owned figurines.
- My Collection page for every owned figurine.
- Search page with HTMX filtering by keyword, series, and character.
- Wishlist page for wanted figurines.
- Add/remove collection, wishlist, and shelf items.
- Single-user data model now, with isolated data access ready for later user scoping.

## Database Setup

Create a PostgreSQL database, then run the migrations in order:

```sh
psql "$DATABASE_URL" -f backend/migrations/001_init.sql
psql "$DATABASE_URL" -f backend/migrations/002_seed.sql
```

The schema uses `pgcrypto` for UUID generation.

## Local Backend

```sh
cd backend
cp .env.example .env
go run ./cmd/server
```

The backend listens on `http://localhost:8080` by default.

## Local Frontend

```sh
cd frontend
cp .env.example .env
go run ./cmd/server
```

The frontend listens on `http://localhost:3000` by default.

## Backend API

- `GET /api/health`
- `GET /api/series`
- `GET /api/figurines?q=&series_id=&character=`
- `GET /api/collection`
- `POST /api/collection` with `{ "figurine_id": "..." }`
- `DELETE /api/collection?id=...`
- `GET /api/wishlist`
- `POST /api/wishlist` with `{ "figurine_id": "..." }`
- `DELETE /api/wishlist?id=...`
- `GET /api/shelf`
- `POST /api/shelf` with `{ "figurine_id": "..." }`
- `DELETE /api/shelf?id=...`

## Vercel Deployment

Deploy `backend/` as one Vercel project.

Set environment variables:

- `DATABASE_URL`
- `ALLOWED_ORIGIN`, set to the frontend deployment URL

Deploy `frontend/` as a separate Vercel project.

Set environment variables:

- `API_BASE_URL`, set to the backend deployment URL

After the backend is deployed, run migrations against the production PostgreSQL database before using the frontend.

## Notes

- Seed data is intentionally small and uses placeholder metadata rather than bundled copyrighted images.
- `image_url` is supported in the database for future real collection images.
- Auth is intentionally omitted for the first version. When multi-user support is added, collection, wishlist, and shelf rows can be scoped by `user_id`.
