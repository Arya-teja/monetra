type ApiError = {
  code: string;
  message: string;
  details?: Record<string, string>;
};

export type ApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; error: ApiError };

type RequestOptions = {
  method: "GET" | "POST" | "PATCH" | "DELETE";
  body?: unknown;
};

// Calls the Go API (server-side only) and unwraps the standard response envelope.
export async function apiRequest<T>(
  path: string,
  { method, body }: RequestOptions,
): Promise<ApiResult<T>> {
  try {
    const res = await fetch(`${process.env.API_URL}/api/v1${path}`, {
      method,
      headers: { "Content-Type": "application/json" },
      body: body === undefined ? undefined : JSON.stringify(body),
      cache: "no-store",
    });
    const json = await res.json();
    return res.ok ? { ok: true, data: json.data } : { ok: false, error: json.error };
  } catch {
    return {
      ok: false,
      error: { code: "INTERNAL_ERROR", message: "Server tidak dapat dihubungi." },
    };
  }
}
