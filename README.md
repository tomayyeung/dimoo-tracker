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
psql "$DATABASE_URL" -f backend/migrations/003_catalog_slugs.sql
```

The schema uses `pgcrypto` for UUID generation.

## Catalog Imports

Catalog metadata can be maintained in `backend/catalog.json` and imported into PostgreSQL. The importer upserts series by `slug` and figurines by `(series_id, slug)`, so it is safe to rerun after editing names, rarity, or image paths.

Run the slug/image-path migration before importing:

```sh
psql "$DATABASE_URL" -f backend/migrations/003_catalog_slugs.sql
```

Then import:

```sh
cd backend
go run ./cmd/import-catalog ./catalog.json
```

The catalog uses a shared Supabase Storage base URL, a series-level folder, and a figurine-level filename:

```json
{
  "storage_base_url": "https://YOUR_PROJECT_ID.supabase.co/storage/v1/object/public/figurines",
  "series": [
    {
      "name": "Dimoo Dream Travel",
      "slug": "dimoo-dream-travel",
      "image_path": "dimoo-dream-travel",
      "figurines": [
        {
          "name": "Cloud Boarding Pass",
          "slug": "cloud-boarding-pass",
          "image_path": "cloud-boarding-pass.webp"
        }
      ]
    }
  ]
}
```

The importer builds `figurines.image_url` as:

```text
{storage_base_url}/{series.image_path}/{figurine.image_path}
```

For Supabase Storage, upload files to the public `figurines` bucket like:

```text
dimoo-dream-travel/cloud-boarding-pass.webp
dimoo-dream-travel/moonlit-suitcase.webp
```

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
