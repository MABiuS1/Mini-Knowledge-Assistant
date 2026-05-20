# AI Usage Journal

## Session 1: Planning the assessment implementation

**Prompt:** "Read the assessment document and create a detailed implementation plan. Backend must use Go Fiber."

**AI Response:** Summarized the assessment into required features, bonus priorities, architecture, documents to submit, and a five-day execution plan.

**My Adjustment:** Chose a split architecture with Next.js frontend, Go Fiber backend, PostgreSQL + pgvector, and OpenAI to fit the requirement while keeping deployment simple with Docker Compose.

## Session 2: Project scaffold

**Prompt:** "Implement the first scaffold step only so I can commit it myself."

**AI Response:** Created the initial project directories and skeleton files for Docker Compose, environment variables, README, AI journal, and architecture decisions.

**My Adjustment:** Kept the first commit intentionally small with no feature implementation, so later commits can map cleanly to the assessment plan.

## Session 3: Database migration design

**Prompt:** "Add the PostgreSQL + pgvector migration step for users, sessions, documents, chunks, conversations, messages, usage events, and a seeded admin user."

**AI Response:** Proposed one initialization migration that enables pgvector and pgcrypto, creates the core application tables, adds indexes, and seeds the mock admin account with a bcrypt password hash.

**My Adjustment:** Kept the schema minimal for the assessment scope and used one database service for both relational data and vector search to keep Docker Compose simple.

## Session 4: Backend authentication

**Prompt:** "Implement backend login/logout/me routes with protected route middleware and unit tests."

**AI Response:** Added a session-based auth service, PostgreSQL repository, Fiber handlers, auth middleware, and route tests using a fake auth service.

**My Adjustment:** Used httpOnly cookie sessions instead of exposing JWTs to the browser, while still supporting Bearer tokens for easier API testing.

## Session 5: Next.js frontend scaffold

**Prompt:** "Scaffold the Next.js frontend with TypeScript, Tailwind, a basic layout, API client wrapper, and Dockerfile."

**AI Response:** Created the frontend project configuration, placeholder app pages, global styles, typed API request helper, and Dockerfile for standalone Next.js builds.

**My Adjustment:** Kept login, upload, and chat pages as placeholders so the scaffold commit stays separate from the authentication UI feature.

## Session 6: Frontend login and route protection

**Prompt:** "Implement the frontend login page, protected routes, session check, and logout flow against the Go Fiber auth API."

**AI Response:** Added an auth API wrapper, React auth provider, login form, route guard, and authenticated app shell for the chat and upload pages.

**My Adjustment:** Used client-side session checks against `/api/me` because the backend owns the httpOnly session cookie, and kept upload/chat content as placeholders for later feature commits.

## Session 7: Backend upload validation

**Prompt:** "Implement the protected file upload endpoint with PDF/TXT validation, file size limits, safe file names, metadata persistence, and tests."

**AI Response:** Added a document upload service, PostgreSQL metadata store, Fiber upload handler, protected route registration, and tests for invalid file type, oversize files, unsafe names, and successful upload persistence.

**My Adjustment:** Kept parsing and chunking out of this commit so upload validation remains a focused change, and used randomized stored file names to avoid trusting user-provided paths.

## Session 8: Document parsing and chunking

**Prompt:** "Parse uploaded TXT/PDF files, split readable text into ordered chunks, and save chunks in PostgreSQL."

**AI Response:** Added TXT/PDF text extraction, whitespace normalization, overlapping word chunks, transactional document + chunk persistence, and tests for chunk ordering and empty content.

**My Adjustment:** Kept embeddings out of this step so parsing/chunking can be verified independently before RAG retrieval is added.

## Session 9: Basic AI chat and token usage

**Prompt:** "Implement a protected basic chat endpoint that calls OpenAI, saves messages, and returns token usage."

**AI Response:** Added a chat service, OpenAI Chat Completions HTTP client, PostgreSQL conversation/message/usage persistence, protected `/api/chat` route, and tests with fake AI and HTTP servers.

**My Adjustment:** Used non-streaming Chat Completions first because the response includes prompt/completion/total token usage directly; streaming is left for a later dedicated feature step.

## Session 10: Gemini provider support

**Prompt:** "Switch the AI provider to Gemini because I have quota there, while keeping OpenAI configurable."

**AI Response:** Added `AI_PROVIDER` config, a Gemini generateContent client, Gemini token usage mapping, provider selection in the server, and tests for Gemini success/error responses.

**My Adjustment:** Kept the chat service provider-agnostic so `/api/chat` response shape and database token tracking do not change when switching between OpenAI and Gemini.

## Session 11: Embeddings and retrieval

**Prompt:** "Add embeddings and document retrieval for RAG using Gemini embeddings with pgvector."

**AI Response:** Added a Gemini embedContent client, embedding configuration, RAG indexing/retrieval service, pgvector repository methods, upload-time chunk embedding, and a protected retrieval endpoint for testing.

**My Adjustment:** Set Gemini embeddings to 1536 dimensions to match the existing pgvector schema and kept context injection/citations for the next commit.

## Session 12: Document-context chat and citations

**Prompt:** "Use retrieved document chunks in /api/chat and return citations."

**AI Response:** Updated the chat service to retrieve chunks when document IDs are provided, inject document context into the AI prompt, and return citation metadata with snippets.

**My Adjustment:** Kept the RAG retrieval interface provider-agnostic and left streaming for a later step so citation behavior stays easy to test.
