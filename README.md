# Knowledge Assistant

Mini Knowledge Assistant for the Dev Trainee Assessment. The app lets an admin user upload PDF/TXT documents, parse them into chunks, embed those chunks, retrieve relevant context, and chat with an AI assistant using citations and token usage tracking.

## Demo

- Live demo: https://pleasant-solace-production.up.railway.app/login
- Username: `admin`
- Password: `admin123`

## Tech Stack

- Frontend: Next.js, TypeScript, Tailwind CSS
- Backend: Go Fiber
- Database: PostgreSQL with pgvector
- AI provider: Gemini by default, OpenAI-compatible chat client also supported
- Runtime: Docker Compose

## Features

- [x] Login and protected routes
- [x] HTTP-only cookie session authentication
- [x] PDF/TXT upload validation
- [x] PDF/TXT parsing and chunking
- [x] Gemini chat provider
- [x] Gemini embeddings
- [x] RAG retrieval with pgvector
- [x] Chat with selected document context
- [x] Token usage per message and per session
- [x] Markdown rendering
- [x] Citations with document/chunk/snippet metadata
- [x] Conversation history
- [x] SSE-style streaming response UI with non-streaming fallback
- [x] Docker Compose healthchecks
- [x] Backend unit and route tests

## Setup

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` and set real provider keys:

```env
AI_API_KEY=replace-with-your-key
EMBEDDING_API_KEY=replace-with-your-key
```

Default local values are included for ports, database credentials, upload size, and Gemini models. Do not commit `.env`.

## Run With Docker

```bash
docker compose up --build
```

Open:

- Frontend: http://localhost:3000
- Backend health: http://localhost:8080/api/health
- PostgreSQL host port: `15432`

Default login:

- Username: `admin`
- Password: `admin123`

To reset the database and uploaded runtime data managed by Docker volumes:

```bash
docker compose down -v
docker compose up --build
```

## Local Development

Start PostgreSQL only:

```bash
docker compose up postgres
```

Run backend locally from the repository root:

```bash
set -a
source .env
set +a
export DATABASE_URL="postgres://knowledge:knowledge@localhost:15432/knowledge_assistant?sslmode=disable"
export UPLOAD_DIR="../data/uploads"
cd backend
go run ./cmd/server
```

Run frontend locally:

```bash
cd frontend
npm install
npm run dev
```

For local frontend development, ensure these values are available to Next.js:

```env
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_MAX_UPLOAD_BYTES=10485760
```

## Testing

Backend:

```bash
cd backend
go test ./...
go test ./... -cover
```

Frontend:

```bash
cd frontend
npm run typecheck
```

Docker validation:

```bash
docker compose config --quiet
docker compose build
```

## API Summary

- `GET /api/health`: service health
- `POST /api/auth/login`: login with username/password
- `POST /api/auth/logout`: logout and clear session cookie
- `GET /api/me`: current authenticated user
- `GET /api/documents`: list uploaded documents
- `POST /api/documents/upload`: upload one PDF/TXT file as form field `file`
- `POST /api/chat`: non-streaming chat completion
- `POST /api/chat/stream`: SSE response with `delta` and `done` events
- `GET /api/chat/conversations`: list conversation history
- `GET /api/chat/conversations/:id`: load one conversation
- `POST /api/rag/retrieve`: debug/test retrieval endpoint

## Architecture

The frontend is a protected Next.js app with login, document upload, document list, chat, markdown rendering, citations, and conversation history. It talks to the backend through cookie-authenticated API calls.

The backend is a Go Fiber API. It owns authentication, session cookies, upload validation, document parsing, chunking, embedding, retrieval, chat orchestration, message persistence, usage tracking, and SSE output.

PostgreSQL stores users, sessions, documents, chunks, conversations, messages, and usage events. pgvector stores chunk embeddings in the same database and supports similarity search for RAG.

## Known Issues

- PDF extraction is best-effort. Text-based and tagged PDFs work better than scanned image-only PDFs.
- Streaming currently streams persisted response chunks from the backend service path. It is not provider-native token streaming yet.
- Citations are returned for retrieved chunks, but loaded conversation history does not currently reconstruct prior citation metadata.
- The app seeds one admin user for assessment use; there is no user management UI.
