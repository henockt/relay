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

function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("relay_token");
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const token = getToken();
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  });

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