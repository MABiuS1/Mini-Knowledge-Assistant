import Link from "next/link";

export default function HomePage() {
  return (
    <main className="min-h-screen p-3 sm:p-4">
      <div className="glass-panel mx-auto grid min-h-[calc(100vh-2rem)] max-w-[1440px] rounded-lg lg:grid-cols-[250px_1fr]">
        <aside className="border-b border-line p-5 lg:border-b-0 lg:border-r">
          <div className="flex items-center gap-3 text-base font-semibold text-ink">
            <span className="glass-logo flex h-9 w-9 items-center justify-center rounded-full text-sm">
              ✦
            </span>
            Knowledge AI
          </div>
          <p className="mt-2 pl-12 text-xs text-muted">Digital Clarity</p>
          <div className="mt-8 space-y-3">
            <Link href="/login" className="glass-active block rounded-md px-4 py-3 text-sm">
              Login
            </Link>
            <Link href="/upload" className="block rounded-md px-4 py-3 text-sm text-muted hover:bg-white/10 hover:text-ink">
              Documents
            </Link>
            <Link href="/chat" className="block rounded-md px-4 py-3 text-sm text-muted hover:bg-white/10 hover:text-ink">
              Chat
            </Link>
          </div>
        </aside>

        <section className="relative flex min-h-[680px] flex-col items-center justify-center overflow-hidden px-6 py-12 text-center">
          <span className="cosmic-orb top-32 h-28 w-28 opacity-90" />
          <div className="mt-36">
            <h1 className="mx-auto max-w-xl text-4xl font-semibold leading-tight text-ink sm:text-5xl">
              Ready to Ark Something ?
            </h1>
            <p className="mx-auto mt-4 max-w-lg text-sm text-muted">
              Upload knowledge, ask focused questions, and review grounded citations in one glass workspace.
            </p>
          </div>

          <div className="glass-panel mt-20 w-full max-w-xl rounded-lg p-3">
            <div className="mb-3 flex flex-wrap justify-center gap-2">
              <Link href="/upload" className="ghost-button rounded-md px-3 py-2 text-sm">
                Upload Document
              </Link>
              <Link href="/chat" className="ghost-button rounded-md px-3 py-2 text-sm">
                Ask Anything
              </Link>
              <Link href="/login" className="ghost-button rounded-md px-3 py-2 text-sm">
                Sign In
              </Link>
            </div>
            <Link
              href="/login"
              className="glass-input flex min-h-[64px] items-center justify-between rounded-full px-5 py-4 text-left text-base text-muted"
            >
              Start with Knowledge AI
              <span className="glow-button flex h-12 w-12 items-center justify-center rounded-full text-2xl text-white">
                ›
              </span>
            </Link>
          </div>
        </section>
      </div>
    </main>
  );
}
