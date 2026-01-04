export type User = {
  id: string;
  email: string;
  username: string;
  createdAt: string;
};

export type Post = {
  id: string;
  title: string;
  body: string;
  createdAt: string;
  author: {
    id: string;
    email: string;
    username: string;
  };
  score: number;
  myVote: number;
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

export async function fetchFeed(token?: string) {
  return request<Post[]>("/api/v1/posts", {
    headers: token ? { Authorization: `Bearer ${token}` } : {}
  });
}

export async function createPost(
  token: string,
  title: string,
  body: string
) {
  return request<Post>("/api/v1/posts", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`
    },
    body: JSON.stringify({ title, body })
  });
}

export async function votePost(token: string, postID: string, value: number) {
  return request<{ status: string }>(`/api/v1/posts/${postID}/vote`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`
    },
    body: JSON.stringify({ value })
  });
}
