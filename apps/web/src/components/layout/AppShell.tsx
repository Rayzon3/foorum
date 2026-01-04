import { Link, Outlet } from "@tanstack/react-router";

import { useAuth } from "../../lib/auth";
import { Button } from "../ui/button";

export function AppShell() {
  const auth = useAuth();

  return (
    <div className="min-h-screen">
      <header className="border-b border-border/60 bg-background/70 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
          <div className="flex items-center gap-3">
            <div className="h-10 w-10 rounded-2xl bg-primary/20 text-center text-xl font-semibold leading-10 text-primary">
              J
            </div>
            <div>
              <p className="text-lg font-semibold">Jabber</p>
              <p className="text-xs text-muted-foreground">
                Connect the dots in realtime
              </p>
            </div>
          </div>
          <nav className="flex items-center gap-4 text-sm">
            <Button variant="ghost" asChild>
              <Link to="/">Home</Link>
            </Button>
            {!auth.user && (
              <>
                <Button variant="outline" asChild>
                  <Link to="/login">Log in</Link>
                </Button>
                <Button asChild>
                  <Link to="/register">Sign up</Link>
                </Button>
              </>
            )}
            {auth.user && (
              <Button variant="secondary" onClick={auth.logout}>
                Log out
              </Button>
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
