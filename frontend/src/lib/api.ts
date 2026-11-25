import { redirect } from "next/navigation";

type ApiError = {
  code: string;
  message: string;
  details?: Record<string, string>;
};

type PageMeta = {
  page: number;
  pageSize: number;
  total: number;
};

export type ApiResult<T> =
  | { ok: true; data: T; meta?: PageMeta }
  | { ok: false; error: ApiError };

type RequestOptions = {
  method: "GET" | "POST" | "PATCH" | "DELETE";
  body?: unknown;
  token?: string;
};

// Route that clears a session whose Go token was rejected (see app/api/session-expired).
const SESSION_EXPIRED_PATH = "/api/session-expired";

// Calls the Go API (server-side only) and unwraps the standard response envelope.
export async function apiRequest<T>(
  path: string,
  { method, body, token }: RequestOptions,
): Promise<ApiResult<T>> {
  let res: Response;
  let json: { data?: T; meta?: PageMeta; error?: ApiError } = {};
  try {
    res = await fetch(`${process.env.API_URL}/api/v1${path}`, {
      method,
      headers: {
        "Content-Type": "application/json",
        ...(token && { Authorization: `Bearer ${token}` }),
      },
      body: body === undefined ? undefined : JSON.stringify(body),
      cache: "no-store",
    });
    if (res.status !== 204) {
      json = await res.json();
    }
  } catch {
    return {
      ok: false,
      error: { code: "INTERNAL_ERROR", message: "Server tidak dapat dihubungi." },
    };
  }

  // The user is logged in here, but Go rejected the token (logged out on another
  // device, expired, or the server secret changed): end the session instead of
  // showing errors. Outside the try block, because redirect() works by throwing.
  if (res.status === 401 && token) {
    redirect(SESSION_EXPIRED_PATH);
  }

  return res.ok
    ? { ok: true, data: json.data as T, meta: json.meta }
    : { ok: false, error: json.error as ApiError };
}
