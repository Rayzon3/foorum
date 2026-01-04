# Jabber v3

Monorepo scaffold for a social networking web app.

## Structure

- `apps/api`: Go + Chi + Postgres API with JWT auth
- `apps/web`: TanStack Router + React Query + Tailwind UI

## Quick start

1. Install deps:
   - `pnpm run install:all`
2. Start Postgres:
   - `docker compose up -d`
3. Apply schema:
   - `export DATABASE_URL="postgres://postgres:postgres@localhost:5432/jabber?sslmode=disable"`
   - `pnpm run migrate:up`
4. Configure env:
   - `cp apps/api/.env.example apps/api/.env`
5. Run API:
   - `pnpm run dev:api`
6. Run web:
   - `pnpm run dev:web`

## Auth endpoints

- `POST /api/v1/auth/register`
- `POST /api/v1/auth/login`
- `GET /api/v1/me` (requires `Authorization: Bearer <token>`)

## Post endpoints

- `GET /api/v1/posts` (optional `Authorization: Bearer <token>`)
- `POST /api/v1/posts` (requires `Authorization: Bearer <token>`)
- `POST /api/v1/posts/:postID/vote` (requires `Authorization: Bearer <token>`, body `{ "value": 1 | -1 | 0 }`)
