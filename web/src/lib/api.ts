import { API_URL as BASE } from "@/lib/config";

export interface Alias {
  id: string;
  user_id: string;
  address: string;
  label: string;
  enabled: boolean;
  emails_forwarded: number;
  emails_blocked: number;
  created_at: string;
}

export interface User {
  id: string;
  email: string;
  provider: string;
  created_at: string;
}

// Clerk owns the session. AuthGate wires these up before any page calls
// the API, so requests always carry a fresh short-lived session token.
let auth: {
  getToken: () => Promise<string | null>;
  signOut: (redirectUrl: string) => Promise<unknown>;
} = {
  getToken: async () => null,
  signOut: async (redirectUrl) => window.location.replace(redirectUrl),
};

export function configureAuth(next: typeof auth) {
  auth = next;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = await auth.getToken();
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });

  if (res.status === 401 || res.status === 403) {
    // 401: session is invalid. 403: signed in with an email relay doesn't accept
    const body = await res.json().catch(() => null);
    const redirect = body?.error === "email_not_allowed" ? "/?error=email_not_allowed" : "/";
    await auth.signOut(redirect);
    throw new Error(`${res.status}: ${body?.error ?? "unauthorized"}`);
  }
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`${res.status}: ${text}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  // Aliases
  listAliases: () => request<Alias[]>("/api/aliases"),
  createAlias: (label?: string) =>
    request<Alias>("/api/aliases", {
      method: "POST",
      body: JSON.stringify({ label: label ?? "" }),
    }),
  updateAlias: (id: string, patch: { label?: string; enabled?: boolean }) =>
    request<Alias>(`/api/aliases/${id}`, {
      method: "PATCH",
      body: JSON.stringify(patch),
    }),
  deleteAlias: (id: string) =>
    request<void>(`/api/aliases/${id}`, { method: "DELETE" }),

  // User
  getMe: () => request<User>("/api/users/me"),
  deleteMe: () => request<void>("/api/users/me", { method: "DELETE" }),
};
