"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuth } from "@/components/auth-provider";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const { status } = useAuth();

  useEffect(() => {
    if (status === "unauthenticated") {
      router.replace("/login");
    }
  }, [router, status]);

  if (status === "loading") {
    return (
      <main className="flex min-h-screen items-center justify-center bg-surface px-6">
        <div className="rounded-lg border border-line bg-white px-5 py-4 text-sm text-muted shadow-sm">
          Checking session...
        </div>
      </main>
    );
  }

  if (status === "unauthenticated") {
    return null;
  }

  return children;
}

