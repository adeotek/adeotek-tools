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
  if (resp.status === 401 && path !== "/api/auth/login") {
    window.location.href = "/login";
    throw new ApiError(401, "unauthenticated");
  }
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
  getStats: () => request<Stats>("/api/stats"),
  triggerSync: () => request<{ ok: boolean }>("/api/stats/sync", { method: "POST" }),

  uploadFiles: async (files: File[]): Promise<UploadResponse> => {
    const form = new FormData();
    files.forEach((f) => form.append("files", f));
    const resp = await fetch("/api/ingest/upload", {
      method: "POST",
      credentials: "include",
      body: form,
    });
    if (resp.status === 401) {
      window.location.href = "/login";
      throw new ApiError(401, "unauthenticated");
    }
    if (!resp.ok) throw new ApiError(resp.status, await resp.text());
    return resp.json() as Promise<UploadResponse>;
  },
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

export interface RepoStats {
  name: string;
  url: string;
  branch: string;
  include_globs: string[];
  chunks: number;
  files: number;
}

export interface KbTableStats {
  table: string;
  rows: number;
}

export interface UploadLogEntry {
  id: number;
  filename: string;
  file_size: number;
  mime_type: string;
  uploaded_at: string;
  status: "ok" | "error";
  chunks_created: number;
  replaced: boolean;
  error_message: string | null;
}

export interface UploadFileResult {
  filename: string;
  status: "ok" | "error";
  chunks_created: number;
  replaced: boolean;
  error_message: string | null;
}

export interface UploadResponse {
  results: UploadFileResult[];
}

export interface Stats {
  last_sync_at: string | null;
  is_syncing: boolean;
  total_chunks: number;
  repos: RepoStats[];
  kb_tables: KbTableStats[];
  conversations: number;
  messages: number;
  uploaded_files: number;
  uploaded_chunks: number;
  upload_log: UploadLogEntry[];
}
