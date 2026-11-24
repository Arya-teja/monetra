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
  token?: string;
};

// Calls the Go API (server-side only) and unwraps the standard response envelope.
export async function apiRequest<T>(
  path: string,
  { method, body, token }: RequestOptions,
): Promise<ApiResult<T>> {
  try {
    const res = await fetch(`${process.env.API_URL}/api/v1${path}`, {
      method,
      headers: {
        "Content-Type": "application/json",
        ...(token && { Authorization: `Bearer ${token}` }),
      },
      body: body === undefined ? undefined : JSON.stringify(body),
      cache: "no-store",
    });
    if (res.status === 204) {
      return { ok: true, data: undefined as T };
    }
    const json = await res.json();
    return res.ok ? { ok: true, data: json.data } : { ok: false, error: json.error };
  } catch {
    return {
      ok: false,
      error: { code: "INTERNAL_ERROR", message: "Server tidak dapat dihubungi." },
    };
  }
}
