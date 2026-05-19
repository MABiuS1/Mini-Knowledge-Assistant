# Knowledge Assistant

Mini Knowledge Assistant for the Dev Trainee Assessment.

## Tech Stack

- Frontend: Next.js, TypeScript, Tailwind CSS
- Backend: Go Fiber
- Database: PostgreSQL with pgvector
- AI: OpenAI chat and embeddings
- Runtime: Docker Compose

## Setup & Run

```bash
cp .env.example .env
docker compose up --build
```

Open:

- Frontend: http://localhost:3000
- Backend health: http://localhost:8080/api/health

Default login:

- Username: `admin`
- Password: `admin123`

## Features Done

- [ ] Login + Protected Routes
- [ ] File Upload
- [ ] Chat with AI
- [ ] Chat with Uploaded File Context
- [ ] Token Usage Counter
- [ ] Markdown rendering
- [ ] Citations
- [ ] Streaming response
- [ ] RAG with Vector DB
- [ ] Conversation history
- [ ] Docker Compose + Healthcheck
- [ ] Unit tests

## Architecture

The application is split into a Next.js frontend and a Go Fiber backend. The backend owns authentication, file upload, document parsing, chunking, embeddings, retrieval, chat orchestration, and token usage tracking. PostgreSQL stores application data and pgvector stores document embeddings for RAG retrieval.

## Known Issues

- Initial scaffold only. Application features are not implemented yet.

