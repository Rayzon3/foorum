export type User = {
  id: string;
  email: string;
  username: string;
  createdAt: string;
};

const baseUrl = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const response = await fetch(`${baseUrl}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(options.headers ?? {})
    },
    ...options
  });

  if (!response.ok) {
    const body = await response.json().catch(() => ({}));
    const error = new Error(body.error ?? "request_failed");
    throw error;
  }

  return response.json();
}

export async function register(email: string, username: string, password: string) {
  return request<{ token: string; user: User }>("/api/v1/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, username, password })
  });
}

export async function login(identifier: string, password: string) {
  return request<{ token: string; user: User }>("/api/v1/auth/login", {
    method: "POST",
    body: JSON.stringify({ email: identifier, password })
  });
}

export async function fetchMe(token: string) {
  return request<User>("/api/v1/me", {
    headers: {
      Authorization: `Bearer ${token}`
    }
  });
}
