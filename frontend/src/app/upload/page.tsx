import { AppShell } from "@/components/app-shell";
import { AuthGuard } from "@/components/auth-guard";

export default function UploadPage() {
  return (
    <AuthGuard>
      <AppShell>
        <div className="rounded-lg border border-line bg-white p-6 shadow-sm">
          <h1 className="text-xl font-semibold text-ink">Upload documents</h1>
          <p className="mt-2 text-sm text-muted">
            File upload controls will be implemented after authentication.
          </p>
        </div>
      </AppShell>
    </AuthGuard>
  );
}
