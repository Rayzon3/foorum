# Jabber v3

Jabber is social app currently in development. Build rooms create posts, follow people, and keep the conversation flowing.

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

## RTC (Spaces-style) signaling

- WebSocket: `GET /api/v1/rooms/:roomID/ws` (requires `Authorization: Bearer <token>` or `?token=...`)
  - Example client URL: `ws://localhost:8080/api/v1/rooms/lobby/ws?token=<jwt>`

Message protocol (JSON):

- Client → Server:
  - `{ "type": "join", "payload": { "role": "speaker" | "listener" } }`
  - `{ "type": "offer", "sdp": "<sdp>" }`
  - `{ "type": "candidate", "candidate": { "candidate": "<ice>", "sdpMid": "audio", "sdpMLineIndex": 0 } }`
- Server → Client:
  - `{ "type": "answer", "sdp": "<sdp>" }`
  - `{ "type": "candidate", "candidate": { "candidate": "<ice>", "sdpMid": "audio", "sdpMLineIndex": 0 } }`
