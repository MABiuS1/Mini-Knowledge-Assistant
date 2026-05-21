# Architecture Decisions

## Go Fiber Backend

### Context

The assessment can be implemented with framework API routes, but this project uses Go Fiber as a dedicated backend. The backend has to handle authentication, file upload validation, document parsing, embeddings, retrieval, chat orchestration, streaming, and persistence.

### Decision

Use a standalone Go Fiber service behind `/api/*`, with Next.js kept as the frontend application.

### Why

Go Fiber gives the backend a clear boundary and keeps application rules out of React components. It also makes backend behavior easier to test with service tests and Fiber route tests. This structure is useful for the assessment because API responsibilities are explicit:

- Auth and session cookies live in the backend.
- Upload limits and file validation are enforced server-side.
- RAG retrieval and AI provider calls are not exposed directly to the browser.
- Token usage and conversation persistence are handled in one place.

### Alternatives Considered

- Next.js API routes only
- A single full-stack Next.js app
- Go Fiber as a separate service

### Trade-offs

The separate backend adds CORS, cookie, and Docker Compose coordination. It also means local development needs both frontend and backend processes. The trade-off is acceptable because the backend has enough responsibilities to justify a clear service boundary.

## Postgres + pgvector

### Context

The app needs relational data for users, sessions, documents, chunks, conversations, messages, and usage events. It also needs vector similarity search for RAG retrieval.

### Decision

Use PostgreSQL as the primary database and pgvector for document chunk embeddings.

### Why

Postgres keeps the core application data model simple and reliable. pgvector avoids introducing a separate vector database for a small assessment project while still supporting cosine similarity search over embeddings. Docker Compose can start one database service with migrations and the required extension.

### Alternatives Considered

- SQLite plus a vector service
- PostgreSQL plus Qdrant
- PostgreSQL plus Pinecone or another hosted vector database
- PostgreSQL plus pgvector

### Trade-offs

pgvector is simpler to operate but has fewer retrieval features than a dedicated vector database. For this scope, simple top-k chunk retrieval is enough. If the corpus became large, the next improvements would be better indexing strategy, filtering, reranking, and observability around retrieval quality.

## RAG And Citation Strategy

### Context

The assistant should answer questions from uploaded documents and show where the answer came from. Sending entire documents to the model is not reliable once files become large, and it also wastes tokens.

### Decision

Parse documents into overlapping text chunks, embed each chunk, retrieve the most relevant chunks for a user query, inject those chunks into the prompt, and return citation metadata for the retrieved chunks.

### Why

Chunk retrieval keeps prompts smaller and gives the model focused context. Citations make answers easier to verify and are useful for assessment review because the response can point back to document, chunk, and snippet metadata.

The current prompt strategy is conservative: when document context is provided, the model is instructed to answer from that context and to say when the context is insufficient instead of guessing.

### Alternatives Considered

- Send the full uploaded document to the model
- Keyword search only
- Embedding retrieval without citations
- RAG with citations

### Trade-offs

RAG quality depends on text extraction, chunk size, overlap, embedding model quality, and retrieval ranking. Current citations are based on retrieved chunks, not exact sentence-level attribution. This is sufficient for a practical first version, but future improvements could include reranking, citation IDs inserted into the answer text, and preserving citation metadata in conversation history.
