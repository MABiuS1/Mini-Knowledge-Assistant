import { LoginForm } from "@/components/login-form";

export default function LoginPage() {
  return (
    <main className="relative flex min-h-screen items-center justify-center overflow-hidden px-6 py-10">
      <span className="cosmic-orb left-[16%] top-[24%] h-24 w-24 opacity-80" />
      <span className="cosmic-orb-soft bottom-[14%] left-[8%] h-20 w-20 opacity-80" />
      <span className="cosmic-orb right-[-4rem] top-[-3rem] h-56 w-56 opacity-70" />
      <span className="cosmic-orb bottom-[9%] right-[12%] h-28 w-28 opacity-85" />

      <section className="glass-panel relative w-full max-w-sm rounded-lg px-9 py-10 shadow-purple-500/30">
        <div className="glass-logo mx-auto mb-5 flex h-12 w-12 items-center justify-center rounded-lg text-lg font-semibold">
          ✦
        </div>
        <p className="text-center text-sm font-semibold text-ink">
          Knowledge AI
        </p>
        <p className="mt-1 text-center text-sm text-muted">Digital Clarity</p>
        <h1 className="mt-8 text-center text-base font-medium text-ink">
          Welcome Back
        </h1>
        <p className="mt-1 text-center text-sm text-muted">
          Please enter your details
        </p>
        <LoginForm />
        <p className="mt-8 text-center text-xs uppercase text-muted">
          Protected by Knowledge Shield
        </p>
      </section>
    </main>
  );
}
