export class ApiError extends Error {
  constructor(public status: number, message: string) {
    super(message);
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const resp = await fetch(path, {
    credentials: "include",
    headers: { "Content-Type": "application/json", ...(init?.headers ?? {}) },
    ...init,
  });
  if (!resp.ok) {
    throw new ApiError(resp.status, await resp.text());
  }
  return resp.status === 204 ? (undefined as T) : ((await resp.json()) as T);
}

export const api = {
  login: (password: string) =>
    request<{ ok: boolean }>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ password }),
    }),
  logout: () => request<{ ok: boolean }>("/api/auth/logout", { method: "POST" }),

  listConversations: () =>
    request<Conversation[]>("/api/conv"),

  createConversation: (body: { title: string; provider: string; model: string }) =>
    request<Conversation>("/api/conv", {
      method: "POST",
      body: JSON.stringify(body),
    }),

  listMessages: (id: string) =>
    request<Message[]>(`/api/conv/${id}/messages`),

  renameConversation: (id: string, title: string) =>
    request<{ ok: boolean }>(`/api/conv/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ title }),
    }),

  deleteConversation: (id: string) =>
    request<{ ok: boolean }>(`/api/conv/${id}`, { method: "DELETE" }),

  getSettings: () => request<Settings>("/api/settings"),
};

export interface Conversation {
  id: string;
  title: string;
  provider: string;
  model: string;
  created_at: string;
  updated_at: string;
}

export interface Message {
  id: number;
  role: "user" | "assistant" | "tool";
  content: string;
  tool_name: string | null;
  tool_calls: string | null;
  partial: boolean;
  created_at: string;
}

export interface ProviderInfo {
  id: string;
  available: boolean;
  models: string[];
  tool_capable: string[];
}

export interface Settings {
  default_provider: string;
  default_model: string;
  providers: ProviderInfo[];
}
