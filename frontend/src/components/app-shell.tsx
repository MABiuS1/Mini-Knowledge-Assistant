"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useState } from "react";
import { useAuth } from "@/components/auth-provider";

const navItems = [
  { href: "/upload", label: "Documents", icon: "▤" },
];

type AppShellProps = {
  children: React.ReactNode;
  onNewChat?: () => void;
  sidebarContent?: React.ReactNode;
};

export function AppShell({
  children,
  onNewChat,
  sidebarContent,
}: AppShellProps) {
  const pathname = usePathname();
  const router = useRouter();
  const { signOut, user } = useAuth();
  const [isSigningOut, setIsSigningOut] = useState(false);
  const [isSidebarOpen, setIsSidebarOpen] = useState(false);

  async function handleLogout() {
    setIsSigningOut(true);
    try {
      await signOut();
      router.replace("/login");
    } finally {
      setIsSigningOut(false);
    }
  }

  return (
    <main className="min-h-screen p-2 sm:p-4">
      <button
        type="button"
        onClick={() => setIsSidebarOpen(true)}
        className="glass-active fixed left-3 top-3 z-30 flex h-10 w-10 items-center justify-center rounded-full text-lg lg:hidden"
        aria-label="Open navigation"
      >
        ☰
      </button>
      {isSidebarOpen ? (
        <button
          type="button"
          aria-label="Close navigation"
          onClick={() => setIsSidebarOpen(false)}
          className="fixed inset-0 z-30 bg-black/50 backdrop-blur-sm lg:hidden"
        />
      ) : null}
      <div className="mx-auto grid max-w-[1440px] gap-3 lg:grid-cols-[250px_1fr] lg:gap-4">
        <aside
          className={
            isSidebarOpen
              ? "glass-panel fixed bottom-2 left-2 top-2 z-40 flex w-[min(280px,calc(100vw-1rem))] flex-col overflow-y-auto rounded-lg p-3 transition-transform duration-200 sm:p-4 lg:static lg:min-h-[calc(100vh-2rem)] lg:w-auto lg:translate-x-0"
              : "glass-panel fixed bottom-2 left-2 top-2 z-40 flex w-[min(280px,calc(100vw-1rem))] -translate-x-[calc(100%+1rem)] flex-col overflow-y-auto rounded-lg p-3 transition-transform duration-200 sm:p-4 lg:static lg:min-h-[calc(100vh-2rem)] lg:w-auto lg:translate-x-0"
          }
        >
          <Link
            href="/chat"
            onClick={() => setIsSidebarOpen(false)}
            className="flex items-center gap-3 text-sm font-semibold text-ink sm:text-base"
          >
            <span className="glass-logo flex h-9 w-9 items-center justify-center rounded-full text-sm font-semibold">
              ✦
            </span>
            <span>
              Knowledge <span className="text-purple-300">AI</span>
            </span>
          </Link>


          {onNewChat ? (
            <button
              type="button"
              onClick={() => {
                onNewChat();
                setIsSidebarOpen(false);
              }}
              className="glass-active mt-5 flex items-center gap-3 rounded-md px-3 py-2.5 text-left text-sm font-semibold sm:mt-8 sm:px-4"
            >
              <span className="w-4 text-center text-base leading-none">✎</span>
              New Chat
            </button>
          ) : (
            <Link
              href="/chat"
              onClick={() => setIsSidebarOpen(false)}
              className="glass-active mt-5 flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-semibold sm:mt-8 sm:px-4"
            >
              <span className="w-4 text-center text-base leading-none">✎</span>
              New Chat
            </Link>
          )}

          <nav className="mt-5 space-y-2">
            {navItems.map((item) => {
              const active = pathname === item.href;
              return (
                <Link
                  key={`${item.href}-${item.label}`}
                  href={item.href}
                  onClick={() => setIsSidebarOpen(false)}
                  className={
                    active
                      ? "glass-active flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium sm:px-4"
                      : "flex items-center gap-3 rounded-md px-3 py-2.5 text-sm font-medium text-muted hover:bg-white/10 hover:text-ink sm:px-4"
                  }
                >
                  <span className="w-4 text-center text-base leading-none">
                    {item.icon}
                  </span>
                  {item.label}
                </Link>
              );
            })}
          </nav>

          {sidebarContent ? (
            <div className="mt-4 border-t border-line pt-4">
              {sidebarContent}
            </div>
          ) : null}

          <div className="mt-4 space-y-3 border-t border-line pt-4 lg:mt-auto lg:space-y-4">
            <div className="flex items-center gap-3 px-4">
              <span className="glass-logo flex h-8 w-8 items-center justify-center rounded-full text-xs font-semibold">
                {user?.username?.slice(0, 1).toUpperCase() ?? "A"}
              </span>
              <div className="min-w-0">
                <p className="truncate text-sm font-medium text-ink">
                  {user?.username ?? "Admin"}
                </p>
                <p className="truncate text-xs text-muted">knowledge-ai</p>
              </div>
            </div>

            <div>
              <button
                type="button"
                disabled={isSigningOut}
                onClick={() => {
                  setIsSidebarOpen(false);
                  void handleLogout();
                }}
                className="flex w-full items-center gap-3 rounded-md px-4 py-2.5 text-left text-sm text-muted hover:bg-white/10 hover:text-ink disabled:cursor-not-allowed disabled:opacity-60"
              >
                <span className="w-4 text-center">↪</span>
                {isSigningOut ? "Signing out..." : "Logout"}
              </button>
            </div>
          </div>
        </aside>

        <div className="min-w-0">{children}</div>
      </div>
    </main>
  );
}
