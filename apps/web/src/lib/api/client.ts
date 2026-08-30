const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "";

interface RequestOptions extends RequestInit {
  skipAuth?: boolean;
}

class ApiError extends Error {
  constructor(public response: Response, message?: string) {
    super(message ?? `API error ${response.status}`);
  }
}

function getAuthToken(): string | null {
  if (typeof window !== "undefined") {
    return (
      localStorage.getItem("access_token") ||
      localStorage.getItem("token") ||
      null
    );
  }
  return null;
}

async function apiRequest<T = unknown>(
  path: string,
  options: RequestOptions = {}
): Promise<T> {
  const url = `${API_BASE}${path}`;
  const headers = new Headers(options.headers);
  if (!headers.has("Content-Type") && options.body && typeof options.body === "string") {
    headers.set("Content-Type", "application/json");
  }

  if (!options.skipAuth) {
    const token = getAuthToken();
    if (token && !headers.has("Authorization")) {
      headers.set("Authorization", `Bearer ${token}`);
    }
  }

  const res = await fetch(url, {
    ...options,
    credentials: "include",
    headers,
  });

  if (!res.ok) {
    // Backend returns errors under either `error` or `message` keys.
    let message: string | undefined;
    try {
      const body = await res.json();
      message = body?.error || body?.message || body?.details;
    } catch {
      // ignore non-JSON error bodies
    }
    throw new ApiError(res, message);
  }

  if (res.status === 204) {
    return undefined as T;
  }

  return (await res.json()) as T;
}

export { ApiError, apiRequest, API_BASE };
export default apiRequest;
