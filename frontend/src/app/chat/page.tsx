import { AppShell } from "@/components/app-shell";
import { AuthGuard } from "@/components/auth-guard";

export default function ChatPage() {
  return (
    <AuthGuard>
      <AppShell>
        <div className="rounded-lg border border-line bg-white p-6 shadow-sm">
          <h1 className="text-xl font-semibold text-ink">Chat</h1>
          <p className="mt-2 text-sm text-muted">
            Chat interface will be implemented after authentication.
          </p>
        </div>
      </AppShell>
    </AuthGuard>
  );
}
