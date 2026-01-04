import React from "react";

import { useAuth } from "../../lib/auth";
import { Button } from "../ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "../ui/card";
import { Input } from "../ui/input";

type AuthFormProps = {
  title: string;
  action: string;
  onSubmit: (email: string, password: string, username?: string) => Promise<void>;
  footer: React.ReactNode;
  showUsername?: boolean;
  emailLabel?: string;
  emailType?: "email" | "text";
};

export function AuthForm({
  title,
  action,
  onSubmit,
  footer,
  showUsername = false,
  emailLabel = "Email",
  emailType = "email",
}: AuthFormProps) {
  const [email, setEmail] = React.useState("");
  const [username, setUsername] = React.useState("");
  const [password, setPassword] = React.useState("");
  const [error, setError] = React.useState<string | null>(null);
  const auth = useAuth();

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    setError(null);
    try {
      await onSubmit(email, password, username);
    } catch (err) {
      setError((err as Error).message);
    }
  }

  return (
    <section className="mx-auto max-w-md">
      <Card className="rounded-3xl border-border/70 bg-card/90">
        <CardHeader>
          <CardTitle className="text-2xl">{title}</CardTitle>
          <CardDescription>Enter your credentials to continue.</CardDescription>
        </CardHeader>
        <CardContent>
          <form className="space-y-4" onSubmit={handleSubmit}>
            <label className="block text-sm text-muted-foreground">
              {emailLabel}
              <Input
                type={emailType}
                value={email}
                onChange={(event) => setEmail(event.target.value)}
                className="mt-2"
                required
              />
            </label>
            {showUsername && (
              <label className="block text-sm text-muted-foreground">
                Username
                <Input
                  type="text"
                  value={username}
                  onChange={(event) => setUsername(event.target.value)}
                  className="mt-2"
                  required
                />
              </label>
            )}
            <label className="block text-sm text-muted-foreground">
              Password
              <Input
                type="password"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
                className="mt-2"
                minLength={8}
                required
              />
            </label>
            <Button type="submit" disabled={auth.loading} className="w-full">
              {auth.loading ? "Working..." : action}
            </Button>
            {error && <p className="text-xs text-destructive">{error}</p>}
          </form>
        </CardContent>
        <CardFooter className="text-sm text-muted-foreground">
          {footer}
        </CardFooter>
      </Card>
    </section>
  );
}
