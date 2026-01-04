import { Link, useNavigate } from "@tanstack/react-router";

import { useAuth } from "../lib/auth";
import { AuthForm } from "../components/auth/AuthForm";

export function LoginPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  return (
    <AuthForm
      title="Welcome back"
      action="Log in"
      emailLabel="Email or username"
      emailType="text"
      onSubmit={async (identifier, password) => {
        await auth.login(identifier, password);
        await navigate({ to: "/" });
      }}
      footer={
        <>
          Need an account?{" "}
          <Link to="/register" className="text-primary hover:underline">
            Create one
          </Link>
        </>
      }
    />
  );
}
