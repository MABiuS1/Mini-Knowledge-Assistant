# Architecture Decisions

## Decision 1: Use Go Fiber as a separate backend

### Context

The assessment allows Next.js or Nuxt.js API routes, but the implementation requirement for this project is to use Go Fiber for the backend.

### Alternatives Considered

- Next.js API routes
- Nuxt server routes
- Go Fiber as a standalone API

### Why Go Fiber

Go Fiber gives the backend a clear boundary for authentication, uploads, document processing, RAG, and OpenAI orchestration. This makes the system easier to explain in review and keeps frontend code focused on user experience.

### Trade-offs

The app needs an extra service in Docker Compose and explicit CORS/cookie handling between frontend and backend.

## Decision 2: Use PostgreSQL with pgvector

### Context

The app needs relational data for users, documents, conversations, and usage, plus vector search for RAG.

### Alternatives Considered

- SQLite plus Qdrant
- PostgreSQL plus Qdrant
- PostgreSQL plus pgvector

### Why PostgreSQL + pgvector

PostgreSQL with pgvector keeps the data model in one database while still supporting vector similarity search. It reduces operational complexity for a take-home project and works well with Docker Compose.

### Trade-offs

Dedicated vector databases may offer more advanced retrieval features, but pgvector is enough for the project scope.

## Decision 3: Use RAG with citations

### Context

The assessment rewards answering questions from uploaded documents and showing where the answer came from.

### Alternatives Considered

- Send the whole document to the model
- Keyword search only
- Chunking, embeddings, retrieval, and citations

### Why RAG

Chunking and retrieval handles larger files better than sending the entire document to the model. Citations also make answers easier to verify and improve the grading signal.

### Trade-offs

RAG adds complexity around chunking, embeddings, and retrieval quality. The implementation will keep the retrieval pipeline simple and document known limitations.

