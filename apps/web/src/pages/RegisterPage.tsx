import { Link, useNavigate } from "@tanstack/react-router";

import { useAuth } from "../lib/auth";
import { AuthForm } from "../components/auth/AuthForm";

export function RegisterPage() {
  const auth = useAuth();
  const navigate = useNavigate();
  return (
    <AuthForm
      title="Create your account"
      action="Sign up"
      showUsername
      externalError={auth.authError}
      onSubmit={async (email, password, username) => {
        await auth.register(email, username ?? "", password);
        await navigate({ to: "/" });
      }}
      footer={
        <>
          Already have an account?{" "}
          <Link to="/login" className="text-primary hover:underline">
            Log in
          </Link>
        </>
      }
    />
  );
}
