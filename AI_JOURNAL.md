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
