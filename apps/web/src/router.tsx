import React from "react";
import {
  Link,
  Outlet,
  createRootRoute,
  createRoute,
  createRouter,
  useNavigate,
} from "@tanstack/react-router";

import { useAuth } from "./lib/auth";

function AppShell() {
  const auth = useAuth();

  return (
    <div className="min-h-screen">
      <header className="border-b border-slate-800/80 bg-slate-900/70 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-2xl bg-sky-400/20 text-center text-xl font-semibold leading-10 text-sky-200">
              J
            </div>
            <div>
              <p className="text-lg font-semibold">Jabber</p>
              <p className="text-xs text-slate-400">
                Connect the dots in realtime
              </p>
            </div>
          </div>
          <nav className="flex items-center gap-4 text-sm">
            <Link to="/" className="text-slate-200 hover:text-white">
              Home
            </Link>
            {!auth.user && (
              <>
                <Link to="/login" className="text-slate-200 hover:text-white">
                  Log in
                </Link>
                <Link
                  to="/register"
                  className="rounded-full bg-sky-400/20 px-4 py-2 text-sky-200 hover:bg-sky-400/30"
                >
                  Sign up
                </Link>
              </>
            )}
            {auth.user && (
              <button
                onClick={auth.logout}
                className="rounded-full border border-slate-700 px-4 py-2 text-slate-200 hover:border-slate-500"
              >
                Log out
              </button>
            )}
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-10">
        <Outlet />
      </main>
    </div>
  );
}

function HomePage() {
  const auth = useAuth();

  return (
    <section className="grid gap-8 md:grid-cols-[2fr_1fr]">
      <div className="space-y-6">
        <div className="rounded-3xl border border-slate-800 bg-slate-900/60 p-8 shadow-xl shadow-slate-950/60">
          <h1 className="text-3xl font-semibold text-white">
            Start new connections
          </h1>
          <p className="mt-3 text-slate-300">
            Jabber is a lightweight social layer. Build rooms, follow people,
            and keep the conversation flowing. This scaffold includes auth,
            routing, and a Go API.
          </p>
          {!auth.user && (
            <div className="mt-6 flex flex-wrap gap-4">
              <Link
                to="/register"
                className="rounded-full bg-sky-400 px-6 py-2 text-sm font-semibold text-slate-950"
              >
                Create account
              </Link>
              <Link
                to="/login"
                className="rounded-full border border-slate-700 px-6 py-2 text-sm font-semibold text-slate-200"
              >
                Log in
              </Link>
            </div>
          )}
        </div>
        <div className="grid gap-4 sm:grid-cols-2">
          <FeatureCard
            title="Fast onboarding"
            detail="Register, login, and hit /me in minutes."
          />
          <FeatureCard
            title="TanStack ready"
            detail="Router + React Query wired and waiting."
          />
          <FeatureCard
            title="Go + Chi"
            detail="Simple HTTP handlers with JWT auth."
          />
          <FeatureCard
            title="Postgres"
            detail="SQL-first schema with migrations."
          />
        </div>
      </div>
      <aside className="space-y-6">
        <div className="rounded-3xl border border-slate-800 bg-slate-900/70 p-6">
          <h2 className="text-lg font-semibold text-white">Session</h2>
          <div className="mt-3 text-sm text-slate-300">
            {auth.loading && "Checking session..."}
            {!auth.loading && auth.user && (
              <div className="space-y-2">
                <p className="text-slate-200">Signed in as</p>
                <p className="font-semibold text-white">{auth.user.email}</p>
                <p className="text-xs text-slate-400">
                  User ID: {auth.user.id}
                </p>
              </div>
            )}
            {!auth.loading && !auth.user && (
              <p>No active session. Try signing up.</p>
            )}
          </div>
          {auth.error && (
            <p className="mt-3 text-xs text-rose-300">
              Auth error: {auth.error}
            </p>
          )}
        </div>
        <div className="rounded-3xl border border-slate-800 bg-gradient-to-br from-slate-900/70 to-slate-950/80 p-6">
          <h3 className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-400">
            Next
          </h3>
          <ul className="mt-3 space-y-2 text-sm text-slate-300">
            <li>Spin up profile feeds and follow graphs.</li>
            <li>Swap JWT for refresh tokens as needed.</li>
            <li>Build the social graph in Postgres.</li>
          </ul>
        </div>
      </aside>
    </section>
  );
}

function FeatureCard({ title, detail }: { title: string; detail: string }) {
  return (
    <div className="rounded-2xl border border-slate-800 bg-slate-900/50 p-4">
      <p className="text-sm font-semibold text-white">{title}</p>
      <p className="mt-2 text-xs text-slate-400">{detail}</p>
    </div>
  );
}

function AuthForm({
  title,
  action,
  onSubmit,
  footer,
}: {
  title: string;
  action: string;
  onSubmit: (email: string, password: string) => Promise<void>;
  footer: React.ReactNode;
}) {
  const [email, setEmail] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [error, setError] = React.useState<string | null>(null);
  const auth = useAuth();

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setError(null);
    try {
      await onSubmit(email, password);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <section className="mx-auto max-w-md rounded-3xl border border-slate-800 bg-slate-900/70 p-8 shadow-xl shadow-slate-950/60">
      <h1 className="text-2xl font-semibold text-white">{title}</h1>
      <p className="mt-2 text-sm text-slate-400">
        Enter your credentials to continue.
      </p>
      <form className="mt-6 space-y-4" onSubmit={handleSubmit}>
        <label className="block text-sm text-slate-300">
          Email
          <input
            type="email"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
            className="mt-2 w-full rounded-2xl border border-slate-700 bg-slate-950/70 px-4 py-2 text-slate-100"
            required
          />
        </label>
        <label className="block text-sm text-slate-300">
          Password
          <input
            type="password"
            value={password}
            onChange={(event) => setPassword(event.target.value)}
            className="mt-2 w-full rounded-2xl border border-slate-700 bg-slate-950/70 px-4 py-2 text-slate-100"
            minLength={8}
            required
          />
        </label>
        <button
          type="submit"
          disabled={auth.loading}
          className="w-full rounded-full bg-sky-400 px-6 py-2 text-sm font-semibold text-slate-950 disabled:opacity-60"
        >
          {auth.loading ? "Working..." : action}
        </button>
        {error && <p className="text-xs text-rose-300">{error}</p>}
      </form>
      <div className="mt-6 text-sm text-slate-400">{footer}</div>
    </section>
  );
}

function LoginPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  return (
    <AuthForm
      title="Welcome back"
      action="Log in"
      onSubmit={async (email, password) => {
        await auth.login(email, password);
        await navigate({ to: "/" });
      }}
      footer={
        <>
          Need an account?{" "}
          <Link to="/register" className="text-sky-300 hover:text-sky-200">
            Create one
          </Link>
        </>
      }
    />
  );
}

function RegisterPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  return (
    <AuthForm
      title="Create your account"
      action="Sign up"
      onSubmit={async (email, password) => {
        await auth.register(email, password);
        await navigate({ to: "/" });
      }}
      footer={
        <>
          Already have an account?{" "}
          <Link to="/login" className="text-sky-300 hover:text-sky-200">
            Log in
          </Link>
        </>
      }
    />
  );
}

const rootRoute = createRootRoute({
  component: AppShell,
});

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: HomePage,
});

const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: LoginPage,
});

const registerRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/register",
  component: RegisterPage,
});

const routeTree = rootRoute.addChildren([
  indexRoute,
  loginRoute,
  registerRoute,
]);

export const router = createRouter({ routeTree });

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
