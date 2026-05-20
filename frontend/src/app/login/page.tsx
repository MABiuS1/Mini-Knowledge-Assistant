import { LoginForm } from "@/components/login-form";

export default function LoginPage() {
  return (
    <main className="flex min-h-screen items-center justify-center bg-surface px-6">
      <section className="w-full max-w-sm rounded-lg border border-line bg-white p-6 shadow-sm">
        <h1 className="text-xl font-semibold text-ink">Sign in</h1>
        <p className="mt-2 text-sm text-muted">
          Use the mock admin account to access the workspace.
        </p>
        <LoginForm />
      </section>
    </main>
  );
}
