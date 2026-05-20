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
