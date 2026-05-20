import Link from "next/link";

export default function HomePage() {
  return (
    <main className="min-h-screen bg-surface px-6 py-10">
      <div className="mx-auto flex max-w-5xl flex-col gap-8">
        <header className="flex flex-col gap-2">
          <p className="text-sm font-medium uppercase tracking-wide text-brand">
            Mini Knowledge Assistant
          </p>
          <h1 className="text-3xl font-semibold text-ink">
            Ask questions from your uploaded documents.
          </h1>
          <p className="max-w-2xl text-base text-muted">
            This scaffold will become the authenticated chat and document upload
            workspace for the assessment.
          </p>
        </header>

        <section className="grid gap-4 md:grid-cols-3">
          <Link
            href="/login"
            className="rounded-lg border border-line bg-white p-5 shadow-sm transition hover:border-brand"
          >
            <h2 className="text-lg font-semibold text-ink">Login</h2>
            <p className="mt-2 text-sm text-muted">
              Authenticate with the mock admin account.
            </p>
          </Link>
          <Link
            href="/upload"
            className="rounded-lg border border-line bg-white p-5 shadow-sm transition hover:border-brand"
          >
            <h2 className="text-lg font-semibold text-ink">Upload</h2>
            <p className="mt-2 text-sm text-muted">
              Add PDF or TXT files for document-aware answers.
            </p>
          </Link>
          <Link
            href="/chat"
            className="rounded-lg border border-line bg-white p-5 shadow-sm transition hover:border-brand"
          >
            <h2 className="text-lg font-semibold text-ink">Chat</h2>
            <p className="mt-2 text-sm text-muted">
              Start a conversation and track token usage.
            </p>
          </Link>
        </section>
      </div>
    </main>
  );
}

